package ui

import (
	_ "embed"
	"sync"
)

//go:embed web.go.tmpl
var webTemplate []byte

//go:embed web.css
var CSSFile []byte

//go:embed js.js
var JsFile []byte

//go:embed cast.js
var JsFileCast []byte

//go:embed socket.js
var JsFileSocket []byte

//go:embed dom-ready.js
var JsFileDOMReady []byte

var once sync.Once

func WebTemplate() []byte {
	once.Do(func() {
		webTemplate = append(webTemplate, []byte("\n"+`{{ define "css" }}`+"\n")...)
		webTemplate = append(webTemplate, CSSFile...)
		webTemplate = append(webTemplate, []byte(`{{ end }}`+"\n")...)
		webTemplate = append(webTemplate, []byte("\n"+`{{ define "js" }}`+"\n")...)
		webTemplate = append(webTemplate, JsFile...)
		webTemplate = append(webTemplate, JsFileCast...)
		webTemplate = append(webTemplate, JsFileSocket...)
		webTemplate = append(webTemplate, JsFileDOMReady...)
		webTemplate = append(webTemplate, []byte(`{{ end }}`+"\n")...)
	})
	return webTemplate
}
