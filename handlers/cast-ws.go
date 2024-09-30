package handlers

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/coder/websocket"
	configuration "github.com/oktalz/present/config"
	"github.com/oktalz/present/data"
	"github.com/oktalz/present/exec"
	"github.com/oktalz/present/types"
)

func CastWS(server data.Server, config configuration.Config) http.Handler { //nolint:funlen,gocognit,revive
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "bye")
		log.Println("connected")

		ch := make(chan string, 10)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		mt, bodyBytes, err := conn.Read(ctx) //nolint:contextcheck // ??? linter is weird
		if err != nil {
			err = conn.Write(context.Background(), mt, []byte(err.Error())) //nolint:contextcheck
			if err != nil {
				log.Println("write:", err)
			}
			return
		}
		defer r.Body.Close()
		payload, err := parseJSONData(string(bodyBytes))
		if err != nil {
			err = conn.Write(context.Background(), mt, []byte(err.Error())) //nolint:contextcheck
			if err != nil {
				log.Println("write:", err)
			}
			return
		}
		// userID := cookieIDValue(w, r)
		adminPwd := config.Security.AdminPwd
		adminPrivileges := cookieAdminAuth(adminPwd, r)
		if adminPwd != "" && !adminPrivileges {
			err = conn.Write(context.Background(), mt, []byte("presenter<br>option only<br>🤷 💥 💔<br>")) //nolint:contextcheck
			if err != nil {
				log.Println("write:", err)
			}
			return
		}

		log.Println("recv: " + strconv.Itoa(payload.Slide))
		log.Printf("recv: %s", payload.Code)

		slideIndex := int64(payload.Slide)
		slides := data.Presentation().Slides
		if slideIndex < 0 || slideIndex >= int64(len(slides)) {
			http.Error(w, "Invalid slide number", http.StatusBadRequest)
		}
		slide := slides[slideIndex]
		terminalCommand := slide.TerminalCommand
		tcBefore := slide.TerminalCommandBefore
		tcAfter := slide.TerminalCommandAfter
		if slide.CanEdit {
			index := 0
			for i := range terminalCommand {
				if terminalCommand[i].Code.IsEmpty {
					continue
				}
				if terminalCommand[i].Index == -1 {
					terminalCommand[i].Index = 0
				}
				if len(payload.Code) > index && payload.Code[index] != "" {
					terminalCommand[i].Code.Code = payload.Code[index]
				}
				index++
			}
		}
		if payload.Block != nil && *payload.Block > -1 && *payload.Block < len(terminalCommand) {
			terminalCommand = []types.TerminalCommand{terminalCommand[*payload.Block]}
		}
		workingDir := slide.Path
		if workingDir == "" {
			workingDir = os.TempDir() + "/present-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			err = os.MkdirAll(workingDir, 0o755)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// defer os.RemoveAll(workingDir)
		}

		for _, cmd := range terminalCommand {
			if !cmd.DirFixed && slide.Path != "" {
				cmd.Dir = slide.Path
			}
			if cmd.Dir == "" { //nolint:nestif
				cmd.Dir = workingDir
			}
			if !cmd.Code.IsEmpty {
				err = os.WriteFile(filepath.Join(workingDir, cmd.FileName), []byte(cmd.Code.Header+cmd.Code.Code+cmd.Code.Footer), 0o600)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}
		}

		if len(tcBefore) > 0 {
			for i := range tcBefore {
				if tcBefore[i].DirFixed {
					go exec.CmdStream(tcBefore[i]) //nolint:contextcheck
					time.Sleep(500 * time.Millisecond)
				} else {
					tcBefore[i].Dir = workingDir
					exec.CmdStream(tcBefore[i]) //nolint:contextcheck
				}
			}
		}
		// for _, cmd := range terminalCommand {
		for i := len(terminalCommand) - 1; i >= 0; i-- {
			cmd := terminalCommand[i]
			cmd.Dir = workingDir
			if cmd.App == "" {
				continue
			}
			go exec.CmdStreamWS(cmd, ch, 100*time.Second, false) //nolint:contextcheck ///todo 100s
			if slide.HasCastStreamed {
				// this is for streaming
				for line := range ch {
					err = conn.Write(context.Background(), mt, []byte(line)) //nolint:contextcheck
					if err != nil {
						log.Println("write:", err)
						return
					}
				}
			} else {
				lines := []string{}
				for line := range ch {
					lines = append(lines, line)
				}
				err = conn.Write(context.Background(), mt, []byte(strings.Join(lines, "<br>"))) //nolint:contextcheck
				if err != nil {
					log.Println("write:", err)
					return
				}
			}
			break
		}
		for i := range tcAfter {
			tcAfter[i].Dir = workingDir
			exec.CmdStream(tcAfter[i]) //nolint:contextcheck
		}
	})
}
