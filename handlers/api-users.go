package handlers

import (
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/mileusna/useragent"
	configuration "github.com/oktalz/present/config"
	"github.com/oktalz/present/hash"
)

type User struct {
	LoginTime time.Time           `json:"login_time"`
	Username  string              `json:"username"`
	IP        string              `json:"ip"`
	Hash      string              `json:"-"`
	UA        useragent.UserAgent `json:"ua"`
	Admin     bool                `json:"admin"`
}

var (
	users   map[string]User
	muUsers = &sync.Mutex{}
)

func APIUsers(config configuration.Config) http.Handler {
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

		muUsers.Lock()
		defer muUsers.Unlock()
		usersSlice := []User{}
		for _, v := range users {
			usersSlice = append(usersSlice, v)
		}
		slices.SortFunc(usersSlice, func(a, b User) int {
			return a.LoginTime.Compare(b.LoginTime)
		})

		err = json.NewEncoder(w).Encode(usersSlice)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
