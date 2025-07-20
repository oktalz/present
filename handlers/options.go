package handlers

import (
	"log"
	"net/http"

	configuration "github.com/oktalz/present/config"
)

func Options(optionsPage []byte, config configuration.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = cookieIDValue(w, r)
		userOK, adminPrivileges := cookieAuth(config.Security.UserPwd, config.Security.AdminPwd, r)
		if config.Security.AdminPwdDisable && !adminPrivileges {
			adminPrivileges = true
		}
		// If the user is not authenticated or does not have admin privileges,
		// check if cookie present-usr exists.
		if !adminPrivileges && !userOK {
			if config.Security.AllowAnyUser {
				cookie, err := r.Cookie("present-usr")
				if err != nil || cookie.Value == "" {
					LoginRedirect(w, r, "/options")
					return
				}
			} else {
				LoginRedirect(w, r, "/options")
				return
			}
		}
		_, err := w.Write(optionsPage)
		if err != nil {
			log.Println(err)
			return
		}
	})
}
