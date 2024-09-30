package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	configuration "github.com/oktalz/present/config"
	"github.com/oktalz/present/data"
	"github.com/oktalz/present/exec"
)

func APICmd(config configuration.Config) http.Handler {
	users = make(map[string]User)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adminPwd := config.Security.AdminPwd
		adminPrivileges := cookieAdminAuth(adminPwd, r)
		if adminPwd != "" && !adminPrivileges {
			http.Error(w, "not authorized", http.StatusUnauthorized)
		}
		presentation := data.Presentation()
		path := r.URL.Path
		path = strings.TrimPrefix(path, "/api/cmd/")
		cmd, ok := presentation.Endpoints[path]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		ch := make(chan string)
		go exec.CmdStreamWS(cmd, ch, 10*time.Second, true) //nolint:contextcheck
		lines := []string{}
		for line := range ch {
			lines = append(lines, line)
		}
		result := strings.Join(lines, "\n")
		_, err := w.Write([]byte(result))
		if err != nil {
			log.Println(err)
		}
	})
}
