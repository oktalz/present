package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/mileusna/useragent"
	configuration "github.com/oktalz/present/config"
	"github.com/oktalz/present/hash"
)

type User struct {
	Username  string              `json:"username"`
	IP        string              `json:"ip"`
	UA        useragent.UserAgent `json:"ua"`
	LoginTime time.Time           `json:"login_time"`
}

var (
	users   map[string]User
	muUsers = &sync.Mutex{}
)

func APIUsers(config configuration.Config) http.Handler {
	users = make(map[string]User)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("present")
		var pass string
		if err == nil {
			// Cookie exists, you can access its value using cookie.Value
			fmt.Println("Cookie value:", cookie.Value)
			pass = cookie.Value
		}

		passwordOK := hash.Equal(pass, config.Security.AdminPwd)
		log.Println("passwordOK", passwordOK)
		if !passwordOK {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		muUsers.Lock()
		defer muUsers.Unlock()
		usersSlice := []User{}
		for _, v := range users {
			usersSlice = append(usersSlice, v)
		}
		err = json.NewEncoder(w).Encode(usersSlice)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
