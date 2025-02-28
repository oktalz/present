package main

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/oklog/ulid/v2"
	configuration "github.com/oktalz/present/config"
	"github.com/oktalz/present/data"
	"github.com/oktalz/present/handlers"
)

func configureServer(config configuration.Config) {
	wsServer := data.NewServer()
	data.Init(wsServer, &config)

	iframeHandler := handlers.IFrame(config)

	http.Handle("/{$}", handlers.Homepage(iframeHandler, config))
	http.Handle("POST /cast", handlers.CastSSE(config))
	http.Handle("/print", handlers.NoLayout(config))
	http.Handle("/iframe", iframeHandler)
	http.Handle("/login", handlers.Login(loginPage))
	http.Handle("/stats", handlers.Stats(statsPage, config))
	http.Handle("/events", handlers.SSE(wsServer, config))
	http.Handle("GET /api/login", handlers.APILogin(config))
	http.Handle("GET /api/users", handlers.APIUsers(config))
	http.Handle("GET /api/cmd/", handlers.APICmd(config))

	sub, err := fs.Sub(dist, "ui/static")
	if err != nil {
		panic(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	handler := &fallbackFileServer{
		primary:   http.FileServer(http.FS(sub)),
		secondary: http.FileServer(http.Dir(wd)),
		eTag:      ulid.Make().String(),
	}
	http.Handle("/", handler)
}

func startServer(config configuration.Config) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, os.Interrupt)

	configureServer(config)

	server := &http.Server{
		Addr:         config.Address + ":" + strconv.Itoa(config.Port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
	}

	go func() {
		log.Println("Listening on", server.Addr)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %s\n", err)
		}
		os.Exit(1) //revive:disable:deep-exit
	}()
	<-signalCh
	// Shutdown the server gracefully
	log.Println("Shutting down...")
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelShutdown()

	server.SetKeepAlivesEnabled(false)
	err := server.Shutdown(shutdownCtx)
	if err != nil {
		log.Printf("HTTP server shutdown error: %s\n", err)
	}
}
