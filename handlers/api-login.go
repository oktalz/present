package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/mileusna/useragent"
	configuration "github.com/oktalz/present/config"
	"github.com/oktalz/present/hash"
)

func APILogin(config configuration.Config) http.Handler {
	userPwd := config.Security.UserPwd
	adminPwd := config.Security.AdminPwd
	users = make(map[string]User)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()
		var err error

		pass, err = hash.Hash(pass)
		if err != nil {
			log.Println(err)
			return
		}

		passwordOK := hash.Equal(pass, userPwd) || hash.Equal(pass, adminPwd)
		if !passwordOK {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		ip := r.RemoteAddr
		userAgent := r.Header.Get("User-Agent")
		ua := useragent.Parse(userAgent)

		log.Println("/api/login", user, "OK", "from", ip, ua.OS, ua.Name)

		muUsers.Lock()
		users[user] = User{
			Username:  user,
			IP:        ip,
			LoginTime: time.Now(),
		}
		defer muUsers.Unlock()

		cookieSet := http.Cookie{
			Name:  "present",
			Value: pass,
			Path:  "/",
		}
		http.SetCookie(w, &cookieSet)
	})
}
