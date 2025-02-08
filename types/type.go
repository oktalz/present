package types

import configuration "github.com/oktalz/present/config"

type Code struct {
	Header  string
	Code    string
	Footer  string
	IsEmpty bool
}

type Asciinema struct {
	Cast     string `json:"cast"`
	URL      string `json:"url"`
	Loop     bool   `json:"loop"`
	AutoPlay bool   `json:"autoplay"`
}

type Cast struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type TerminalCommand struct {
	Dir      string
	DirFixed bool
	App      string
	Cmd      []string
	Code     Code
	Index    int
	FileName string
}

type Content struct {
	Markdown string
	HTML     string
}

type Page struct {
	Header Content
	Data   Content
	Footer Content
}

type Slide struct {
	Page                  Page
	Admin                 Content
	Notes                 string
	IsAdmin               bool
	Terminal              TerminalCommand
	Asciinema             *Asciinema
	Cast                  *Cast
	TerminalCommandBefore []TerminalCommand
	TerminalCommand       []TerminalCommand
	TerminalCommandAfter  []TerminalCommand

	SlideCmdBefore          []TerminalCommand
	SlideCmdAfter           []TerminalCommand
	JS                      string // javascript function to call on page enter
	Path                    string
	UseTmpFolder            bool
	CanEdit                 bool
	HasCast                 bool
	HasCastStreamed         bool
	HasRun                  bool
	HasTerminal             bool
	BackgroundImage         string
	BackgroundColor         string
	PageIndex               int
	PagePrint               int
	FontSize                string
	TerminalFontSize        string
	TerminalFontColor       string
	TerminalBackgroundColor string
	HideRunButton           bool
	Link                    string
	LinkNext                string
	LinkPrev                string
	PrintOnly               bool
	PrintDisable            bool
	HidePageNumber          bool
	Title                   string
	EnableOverflow          bool
}

type Menu struct {
	Link      int
	PagePrint int
	PageIndex int
	Title     string
}

type PresentationOptions struct {
	AspectRatioMin     configuration.AspectRatio
	AspectRatioMax     configuration.AspectRatio
	DisableAspectRatio bool
}

type Presentation struct {
	CSS       string
	Options   PresentationOptions
	JS        string
	HTML      string
	Slides    []Slide
	Menu      []Menu
	Title     string
	Author    string
	Replacers map[string]string
	Endpoints map[string]TerminalCommand
}

type TerminalOutputLine struct {
	Timestamp string
	Line      string
}

type ReadOptions struct {
	DefaultFontSize                string
	AspectRatioMin                 configuration.AspectRatio
	AspectRatioMax                 configuration.AspectRatio
	DisableAspectRatio             bool
	DefaultBackgroundColor         string
	EveryDashIsATransition         bool
	DefaultTerminalFontSize        string
	DefaultTerminalFontColor       string
	DefaultTerminalBackgroundColor string
	HideRunButton                  bool
	HidePageNumber                 bool
	KeepPagePrintOnCut             bool
}
