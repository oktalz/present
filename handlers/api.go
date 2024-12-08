package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/oktalz/present/data"
)

var CurrentSlide = int64(-10)

func api(w http.ResponseWriter, _ *http.Request) {
	err := json.NewEncoder(w).Encode(data.Presentation())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func API() http.Handler {
	return AccessControlAllow(http.HandlerFunc(api))
}
