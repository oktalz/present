package handlers

import (
	"log"
	"net/http"

	configuration "github.com/oktalz/present/config"
)

func Stats(statsPage []byte, config configuration.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = cookieIDValue(w, r)
		userOK, adminPrivileges := cookieAuth(config.Security.UserPwd, config.Security.AdminPwd, r)
		if config.Security.AdminPwdDisable && !adminPrivileges {
			adminPrivileges = true
		}
		if !adminPrivileges {
			if !(userOK) {
				LoginRedirect(w, r, "/stats")
				return
			}
		}
		_, err := w.Write(statsPage)
		if err != nil {
			log.Println(err)
			return
		}
	})
}
