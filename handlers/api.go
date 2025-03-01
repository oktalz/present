package handlers

import (
	"net/http"

	"github.com/goccy/go-json"
	"github.com/oktalz/present/data"
)

var currentSlide = int64(-10)

func api(w http.ResponseWriter, _ *http.Request) {
	err := json.NewEncoder(w).Encode(data.Presentation())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func API() http.Handler { //revive:disable:confusing-naming
	return AccessControlAllow(http.HandlerFunc(api))
}
