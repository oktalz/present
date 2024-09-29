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
