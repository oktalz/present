package handlers

import (
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
		if config.AspectRatio.AspectRatio != "" {
			aspectRatio := strings.Split(config.AspectRatio.AspectRatio, ":")
			if len(aspectRatio) == 1 {
				aspectRatio = strings.Split(config.AspectRatio.AspectRatio, "x")
			}
			if len(aspectRatio) == 2 {
				width, _ := strconv.Atoi(aspectRatio[0])
				height, _ := strconv.Atoi(aspectRatio[1])
				pageResult = strings.Replace(pageResult, "widthRatio = 16", "widthRatio = "+strconv.Itoa(width), 1)
				pageResult = strings.Replace(pageResult, "heightRatio = 9", "heightRatio = "+strconv.Itoa(height), 1)
			}
		}
		hasher := fnv.New64a()
		hasher.Write([]byte(pageResult))
		eTagFrame = strconv.FormatUint(hasher.Sum64(), 16)
	}

	aspectRatioChange()
	go func() {
		for {
			aspectRatio := <-config.AspectRatio.ValueChanged
			config.AspectRatio.AspectRatio = aspectRatio
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

var page = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        .iframe-container {
            max-height: 100vh;
			max-width: 100vw;
            position: relative;
			margin: auto;
        }

        .iframe-container iframe {
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            border: none;
        }

		body {
			margin: 0;
			padding: 0;
			width: 100vw;
			height: 100vh;
			overflow: hidden;
		}
    </style>
</head>
<body>
    <div class="iframe-container">
        <iframe src="/print" id="present-iframe" name="present-iframe" sandbox="allow-scripts allow-same-origin allow-storage-access-by-user-activation"></iframe>
    </div>
	<script>
	  widthRatio = 16
	  heightRatio = 9
	  console.log(location.origin)
	  const urlParams = new URLSearchParams(window.location.search);
	  const newSrc = location.origin + "/print?" + urlParams.toString();
	  document.getElementById("present-iframe").src = newSrc;
	  calcAspectRatioFit = function() {
		expectedRatio = widthRatio / heightRatio
		currentRatio = window.innerWidth / window.innerHeight
		if (widthRatio >= heightRatio) {
			if (currentRatio < expectedRatio) {
				document.querySelector('.iframe-container').style.width = '100svw';
				document.querySelector('.iframe-container').style.height = (window.innerWidth * heightRatio / widthRatio) + 'px';
			} else {
				document.querySelector('.iframe-container').style.width = (window.innerHeight * widthRatio / heightRatio) + 'px';
				document.querySelector('.iframe-container').style.height = '100svh';
			}
		} else {
			if (currentRatio < expectedRatio) {
				document.querySelector('.iframe-container').style.width = '100svw';
				document.querySelector('.iframe-container').style.height = (window.innerWidth * heightRatio / widthRatio) + 'px';
			} else {
				document.querySelector('.iframe-container').style.width = (window.innerHeight * widthRatio / heightRatio) + 'px';
				document.querySelector('.iframe-container').style.height = '100svh';
			}
		}
	  }
	  calcAspectRatioFit()
	  window.addEventListener("resize", calcAspectRatioFit);
	</script>
</body>
</html>

`
