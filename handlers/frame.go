package handlers

import (
	_ "embed"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	configuration "github.com/oktalz/present/config"
)

func IFrame(config configuration.Config) http.Handler { //nolint:funlen
	eTagFrame := ""
	var mu sync.RWMutex
	pageResult := page
	aspectRatioChange := func() {
		mu.Lock()
		defer mu.Unlock()
		pageResult = page
		fmt.Println("Aspect ratio changed", config.AspectRatio.Min.Width, config.AspectRatio.Min.Height, config.AspectRatio.Max.Width, config.AspectRatio.Max.Height)
		pageResult = strings.Replace(pageResult, "widthRatioMin = 16", "widthRatioMin = "+strconv.Itoa(config.AspectRatio.Min.Width)+" // custom", 1)
		pageResult = strings.Replace(pageResult, "widthRatioMax = 16", "widthRatioMax = "+strconv.Itoa(config.AspectRatio.Max.Width)+" // custom", 1)
		pageResult = strings.Replace(pageResult, "heightRatioMin = 9", "heightRatioMin = "+strconv.Itoa(config.AspectRatio.Min.Height)+" // custom", 1)
		pageResult = strings.Replace(pageResult, "heightRatioMax = 9", "heightRatioMax = "+strconv.Itoa(config.AspectRatio.Max.Height)+" // custom", 1)
		hasher := fnv.New64a()
		hasher.Write([]byte(pageResult))
		eTagFrame = strconv.FormatUint(hasher.Sum64(), 16)
	}

	aspectRatioChange()
	go func() {
		for {
			aspectRatio := <-config.AspectRatio.Min.ValueChanged
			mu.Lock()
			config.AspectRatio.Min.Height = aspectRatio.Height
			config.AspectRatio.Min.Width = aspectRatio.Width
			mu.Unlock()
			aspectRatioChange()
		}
	}()
	go func() {
		for {
			aspectRatio := <-config.AspectRatio.Max.ValueChanged
			mu.Lock()
			config.AspectRatio.Max.Height = aspectRatio.Height
			config.AspectRatio.Max.Width = aspectRatio.Width
			mu.Unlock()
			aspectRatioChange()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = cookieIDValue(w, r)
		userOK, adminPrivileges := cookieAuth(config.Security.UserPwd, config.Security.AdminPwd, r)
		if config.Security.AdminPwdDisable && !adminPrivileges {
			adminPrivileges = true
		}
		if config.Security.UserPwd != "" && !adminPrivileges {
			if !(userOK) {
				LoginRedirect(w, r, "/")
				return
			}
		}

		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		if r.Method == http.MethodOptions {
			return
		}

		var response []byte
		_ = response
		eTag := r.Header.Get("If-None-Match")
		mu.RLock()
		defer mu.RUnlock()
		if eTagFrame == eTag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("ETag", eTagFrame)
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(pageResult))
		if err != nil {
			log.Println(err)
			return
		}
	})
}

//go:embed frame.html
var page string
