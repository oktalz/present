package handlers

import (
	"net/http"
	"slices"
	"strings"

	"github.com/goccy/go-json"
	configuration "github.com/oktalz/present/config"
	"github.com/oktalz/present/data"
	"github.com/oktalz/present/hash"
)

func APIConnections(config configuration.Config, server data.Server) http.Handler {
	users = make(map[string]User)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("present-sec")
		var pass string
		if err == nil {
			// Cookie exists, you can access its value using cookie.Value
			// log.Println("Cookie value:", cookie.Value)
			pass = cookie.Value
		}

		adminOK := hash.Equal(pass, config.Security.AdminPwd)
		if !adminOK {
			LoginRedirect(w, r, "/")
			return
		}

		users := server.GetUsers()
		// remove all logged in users (they have - character in their name)
		users = slices.DeleteFunc[[]string](users, func(s string) bool {
			return strings.Contains(s, "-")
		})

		err = json.NewEncoder(w).Encode(map[string]int{"count": len(users)})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
