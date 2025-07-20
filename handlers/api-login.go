package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/goccy/go-json"
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

		adminOK := hash.Equal(pass, adminPwd)
		passwordOK := hash.Equal(pass, userPwd) || adminOK
		if !passwordOK {
			if !config.Security.AllowAnyUser {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
		}
		ip := r.RemoteAddr
		userAgent := r.Header.Get("User-Agent")
		ua := useragent.Parse(userAgent)

		userType := "user"
		if adminOK {
			userType = "admin"
		}

		log.Println("/api/login", user, userType, "from", ip, ua.OS, ua.Name)

		muUsers.Lock()

		cookie, err := r.Cookie("present-loc")
		// find if the user has already made a login, use IP for matching
		if err == nil {
			// Cookie exists, you can access its value using cookie.Value
			for _, u := range users {
				if u.Hash == cookie.Value {
					// user already logged in, delete user from map
					delete(users, u.Username)
					log.Println("User", u.Username, "already logged in, removing from map")
				}
			}
		}

		hashUserLoc, err := hash.Hash(user + ip)
		if err != nil {
			log.Println(err)
			return
		}

		users[user] = User{
			Username:  user,
			Admin:     adminOK,
			IP:        ip,
			Hash:      hashUserLoc,
			LoginTime: time.Now(),
			UA:        ua,
		}
		defer muUsers.Unlock()

		cookieSet := http.Cookie{
			Name:  "present-loc",
			Value: hashUserLoc,
			Path:  "/",
		}
		http.SetCookie(w, &cookieSet)
		cookieSet = http.Cookie{
			Name:  "present-sec",
			Value: pass,
			Path:  "/",
		}
		http.SetCookie(w, &cookieSet)
		cookieSet = http.Cookie{
			Name:  "present-usr",
			Value: user,
			Path:  "/",
		}
		http.SetCookie(w, &cookieSet)
		// return json response with user and admin status
		response := map[string]bool{
			"admin": adminOK,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("User", user, adminOK, "logged in successfully")
	})
}
