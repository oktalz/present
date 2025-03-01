package handlers

import "net/http"

// AccessControlAllow is a middleware that sets CORS headers on the response.
// It allows requests from any origin specified in the request's "Origin" header
// and supports HTTP methods POST, GET, OPTIONS, PUT, and DELETE. It also sets
// the allowed headers to "Accept, Content-Type, Content-Length, Accept-Encoding,
// X-CSRF-Token, Authorization". The next handler in the chain is called after
// setting these headers.
func AccessControlAllow(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		next.ServeHTTP(w, r)
	})
}
