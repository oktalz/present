package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	configuration "github.com/oktalz/present/config"
	"github.com/oktalz/present/data"
	"github.com/oktalz/present/exec"
)

func SSE(server data.Server, config configuration.Config) http.Handler { //nolint:funlen,gocognit
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := cookieIDValue(w, r)
		isAdmin := (config.Security.AdminPwd == "") || cookieAdminAuth(config.Security.AdminPwd, r)

		if r.Method == http.MethodPost {
			if !isAdmin {
				http.Error(w, "not authorized", http.StatusUnauthorized)
				return
			}
			dec := json.NewDecoder(r.Body)
			var msg data.Message
			err := dec.Decode(&msg)
			if err != nil {
				log.Println(err)
				return
			}
			// fmt.Println(string(msg.Msg))
			// log.Println(msg.Slide)
			if msg.Pool != "" {
				// log.Println("user", userID, msg.Pool, msg.Value)
				msg.Author = userID
				server.Pool(msg)
				return
			}

			body := data.Message{
				Author: msg.Author,
				Slide:  msg.Slide,
				Reload: false,
			}

			server.Broadcast(body)
			go func() {
				presentation := data.Presentation()
				if body.Slide < len(presentation.Slides) && body.Slide >= 0 {
					slide := presentation.Slides[body.Slide]
					for _, cmd := range slide.SlideCmdBefore {
						if cmd.Dir == "" && slide.Path != "" {
							cmd.Dir = slide.Path
						}
						ch := make(chan string)
						go exec.CmdStreamWS(cmd, ch, 100*time.Second, false)
						<-ch
					}
				}
			}()
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		if isAdmin {
			// allow more than one tab to work properly only for admins
			userID = userID + "-" + r.RemoteAddr
		}
		// register with server
		serverEvent, err := server.Register(userID, isAdmin, atomic.LoadInt64(&CurrentSlide)) //nolint:varnamelen
		if err != nil {
			log.Println("register:", err)
			return
		}

		// Create a channel for client disconnection
		clientGone := r.Context().Done()

		rc := http.NewResponseController(w) //nolint:bodyclose
		err = rc.SetWriteDeadline(time.Time{})
		if err != nil {
			return
		}

		for {
			select {
			case <-clientGone:
				fmt.Println("Client " + userID + " disconnected")
				server.Unregister(userID)
				return
			case msg := <-serverEvent:
				if userID == msg.Author {
					continue
				}
				buf, _ := json.Marshal(msg)
				if msg.Reload {
					msg.Slide = int(CurrentSlide)
				}
				_, err := fmt.Fprintf(w, "data: %s\n\n", string(buf))
				if err != nil {
					return
				}
				err = rc.Flush()
				if err != nil {
					return
				}
			}
		}
	})
}
