package main

import (
	"bytes"
	"log"
	"maps"
	"net/http"
	"strings"
)

type responseWriter struct {
	Body         bytes.Buffer
	CustomHeader http.Header
	StatusCode   int
}

func (crw *responseWriter) Header() http.Header {
	return crw.CustomHeader
}

func (crw *responseWriter) Write(b []byte) (int, error) {
	return crw.Body.Write(b)
}

func (crw *responseWriter) WriteHeader(statusCode int) {
	crw.StatusCode = statusCode
}

type fallbackFileServer struct {
	primary   http.Handler
	secondary http.Handler
	eTag      string
}

func (s *fallbackFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) { //nolint:varnamelen
	if r.Header.Get("If-None-Match") == s.eTag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	filePath := r.URL.Path
	if strings.Contains(filePath, ".env") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	origin := r.Header.Get("Origin")
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	rw := responseWriter{ //nolint:exhaustruct
		CustomHeader: make(http.Header),
		StatusCode:   http.StatusOK,
	}
	rw.Header().Set("ETag", s.eTag)
	s.primary.ServeHTTP(&rw, r)
	if rw.StatusCode == http.StatusNotFound {
		s.secondary.ServeHTTP(w, r)
		return
	}
	maps.Copy(w.Header(), rw.CustomHeader)
	w.WriteHeader(rw.StatusCode)
	_, err := w.Write(rw.Body.Bytes())
	if err != nil {
		log.Println(err)
	}
}
