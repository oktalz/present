package handlers

import (
	"log"
	"net/http"
)

func Login(loginPage []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write(loginPage)
		if err != nil {
			log.Println(err)
			return
		}
	})
}

func LoginRedirect(w http.ResponseWriter, r *http.Request, origin string) {
	http.Redirect(w, r, "/login?origin="+origin, http.StatusFound)
}
