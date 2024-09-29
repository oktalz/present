package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/oktalz/present/data"
	"github.com/oktalz/present/exec"
)

var CurrentSlide = int64(-10)

func WS(server data.Server, adminPwd string) http.Handler { //nolint:funlen,gocognit
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			log.Println("accept:", err)
			return
		}
		defer conn.Close(websocket.StatusAbnormalClosure, "error? ")

		userID := cookieIDValue(w, r)
		isAdmin := (adminPwd == "") || cookieAdminAuth(adminPwd, r)
		// register with server
		serverEvent, err := server.Register(userID, isAdmin, atomic.LoadInt64(&CurrentSlide)) //nolint:varnamelen
		if err != nil {
			log.Println("register:", err)
			return
		}
		defer server.Unregister(userID)
		browserEvent := make(chan data.Message)
		ctx := context.Background()
		go func(ctx context.Context) { //nolint:contextcheck
			defer ctx.Done()
			for {
				_, message, err := conn.Read(context.Background()) //nolint:contextcheck
				if err != nil {
					log.Println("read:", userID, err)
					return
				}
				var msg data.Message
				err = json.Unmarshal(message, &msg)
				if err != nil {
					log.Println(err)
					continue
				}
				browserEvent <- data.Message{
					Author: userID,
					Msg:    message,
					Slide:  msg.Slide,
					Pool:   msg.Pool,
					Value:  msg.Value,
				}
				if msg.Pool == "" && msg.Data == nil {
					atomic.StoreInt64(&CurrentSlide, int64(msg.Slide))
				}
			}
		}(ctx)

		for {
			select {
			case msg := <-serverEvent:
				if userID == msg.Author {
					continue
				}
				buf, _ := json.Marshal(msg)
				// log.Println(id, string(buf))
				if msg.Reload {
					msg.Slide = int(CurrentSlide)
				}
				err = conn.Write(context.Background(), websocket.MessageText, buf) //nolint:contextcheck
				if err != nil {
					log.Println("write:", err)
					return
				}
			case msg := <-browserEvent:
				if msg.Pool != "" {
					// log.Println("user", userID, msg.Pool, msg.Value)
					msg.Author = userID
					server.Pool(msg)
					continue
				}

				body := data.Message{
					Author: msg.Author,
					Slide:  msg.Slide,
					Reload: false,
				}

				if isAdmin {
					server.Broadcast(body)
					go func() { //nolint:contextcheck
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
				}
			case <-ctx.Done():
				return
			}
		}
	})
}
