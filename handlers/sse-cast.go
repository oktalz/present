package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	configuration "github.com/oktalz/present/config"
	"github.com/oktalz/present/data"
	"github.com/oktalz/present/exec"
	"github.com/oktalz/present/types"
)

func CastSSE(server data.Server, config configuration.Config) http.Handler { //nolint:funlen,gocognit,revive
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported!", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		userID := cookieIDValue(w, r)
		isAdmin := (config.Security.AdminPwd == "") || cookieAdminAuth(config.Security.AdminPwd, r)
		if !isAdmin {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}
		log.Println("cast", userID)

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "cannot read body", http.StatusInternalServerError)
			return
		}
		payload, err := parseJSONData(string(bodyBytes))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ch := make(chan string, 10)
		rc := http.NewResponseController(w)
		err = rc.SetWriteDeadline(time.Time{})
		if err != nil {
			return
		}

		// log.Println("recv: " + strconv.Itoa(payload.Slide))
		// log.Printf("recv: %s", payload.Code)

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
					go exec.CmdStream(tcBefore[i])
					time.Sleep(500 * time.Millisecond)
				} else {
					tcBefore[i].Dir = workingDir
					exec.CmdStream(tcBefore[i])
				}
			}
		}
		for i := len(terminalCommand) - 1; i >= 0; i-- {
			cmd := terminalCommand[i]
			cmd.Dir = workingDir
			if cmd.App == "" {
				continue
			}
			go exec.CmdStreamWS(cmd, ch, 1000*time.Second, false)
			if slide.HasCastStreamed {
				// this is for streaming
				first := true
				var data string
				for line := range ch {
					if first {
						first = false
						data = line + "\n\n"
					} else {
						data = "<br>" + line + "\n\n"
					}
					fmt.Fprint(w, data)
					flusher.Flush()
				}
			} else {
				lines := []string{}
				for line := range ch {
					lines = append(lines, line)
				}
				_, err := fmt.Fprintf(w, "%s\n\n", strings.Join(lines, "<br>"))
				if err != nil {
					return
				}
				flusher.Flush()
			}
			break
		}
		for i := range tcAfter {
			tcAfter[i].Dir = workingDir
			exec.CmdStream(tcAfter[i])
		}
		// alow flusher to send full data
		time.Sleep(100 * time.Millisecond)
	})
}
