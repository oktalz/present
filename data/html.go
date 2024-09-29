package data

import (
	"bytes"
	"cmp"
	"strings"
	"text/template"

	configuration "github.com/oktalz/present/config"
	"github.com/oktalz/present/types"
	"github.com/oktalz/present/ui"
)

type TemplateData struct {
	Slides        []types.Slide
	CSS           string
	JS            string
	Menu          []types.Menu
	Title         string
	Port          int
	PageNext      []string
	PagePrevious  []string
	TerminalCast  []string
	TerminalClose []string
	MenuKey       []string
	Admin         bool
}

func regenerateHTML(presentation types.Presentation, config configuration.Config, adminPrivileges bool) ([]byte, error) {
	slides := presentation.Slides
	for i := range presentation.Slides {
		slides[i].IsAdmin = adminPrivileges
	}
	tmpl, err := template.New("web").Parse(string(ui.WebTemplate()))
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	pageNextStr := cmp.Or(config.Controls.NextPage, "ArrowRight,ArrowDown,PageDown,Space")
	pageNextStr = strings.ReplaceAll(pageNextStr, "Space", " ")
	pagePreviousStr := cmp.Or(config.Controls.PreviousPage, "ArrowLeft,ArrowUp,PageUp")
	pagePreviousStr = strings.ReplaceAll(pagePreviousStr, "Space", " ")
	terminalCastStr := cmp.Or(config.Controls.TerminalCast, "r")
	terminalCastStr = strings.ReplaceAll(terminalCastStr, "Space", " ")
	terminalCloseStr := cmp.Or(config.Controls.TerminalClose, "c")
	terminalCloseStr = strings.ReplaceAll(terminalCloseStr, "Space", " ")
	menuKeyStr := cmp.Or(config.Controls.Menu, "m")
	menuKeyStr = strings.ReplaceAll(menuKeyStr, "Space", " ")
	err = tmpl.Execute(&out, TemplateData{
		Admin:         adminPrivileges,
		Slides:        slides,
		CSS:           presentation.CSS,
		JS:            presentation.JS,
		Title:         presentation.Title,
		Menu:          presentation.Menu,
		PageNext:      strings.Split(pageNextStr, ","),
		PagePrevious:  strings.Split(pagePreviousStr, ","),
		TerminalCast:  strings.Split(terminalCastStr, ","),
		TerminalClose: strings.Split(terminalCloseStr, ","),
		MenuKey:       strings.Split(menuKeyStr, ","),
	})
	if err != nil {
		return nil, err
	}
	str := out.String()
	str = strings.ReplaceAll(str, " ", "")

	return []byte(str), nil
}
