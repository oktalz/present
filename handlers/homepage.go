package handlers

import (
	"log"
	"net/http"

	configuration "github.com/oktalz/present/config"
	"github.com/oktalz/present/data"
)

func Homepage(config configuration.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = cookieIDValue(w, r)
		userOK, adminPrivileges := cookieAuth(config.Security.UserPwd, config.Security.AdminPwd, r)
		if config.Security.UserPwd != "" {
			if !(userOK) {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
		}

		var response []byte
		var status int
		eTag := r.Header.Get("If-None-Match")
		if adminPrivileges {
			response, eTag, status = data.AdminHTML(eTag)
		} else {
			response, eTag, status = data.UserHTML(eTag)
		}

		if status == http.StatusNotModified {
			w.WriteHeader(status)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("ETag", eTag)
		w.WriteHeader(status)

		_, err := w.Write(response)
		if err != nil {
			log.Println(err)
			return
		}
	})
}
