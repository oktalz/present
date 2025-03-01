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
	CSS           string
	JS            string
	HTML          string
	Title         string
	Author        string
	Slides        []types.Slide
	Menu          []types.Menu
	PageNext      []string
	PagePrevious  []string
	TerminalCast  []string
	TerminalClose []string
	MenuKey       []string
	Port          int
	Admin         bool
}

//revive:disable:line-length-limit
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
		HTML:          presentation.HTML,
		Title:         presentation.Title,
		Author:        presentation.Author,
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
