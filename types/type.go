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
	App      string
	FileName string
	Code     Code
	Cmd      []string
	Index    int
	DirFixed bool
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
	Asciinema               *Asciinema
	Cast                    *Cast
	Page                    Page
	Admin                   Content
	Notes                   string
	JS                      string // javascript function to call on page enter
	Path                    string
	BackgroundImage         string
	BackgroundColor         string
	FontSize                string
	TerminalFontSize        string
	TerminalFontColor       string
	TerminalBackgroundColor string
	Link                    string
	LinkNext                string
	LinkPrev                string
	Title                   string
	TerminalCommandBefore   []TerminalCommand
	TerminalCommand         []TerminalCommand
	TerminalCommandAfter    []TerminalCommand

	SlideCmdBefore  []TerminalCommand
	SlideCmdAfter   []TerminalCommand
	Terminal        TerminalCommand
	PageIndex       int
	PagePrint       int
	IsAdmin         bool
	UseTmpFolder    bool
	CanEdit         bool
	HasCast         bool
	HasCastStreamed bool
	HasRun          bool
	HasTerminal     bool
	HideRunButton   bool
	PrintOnly       bool
	PrintDisable    bool
	HidePageNumber  bool
	EnableOverflow  bool
}

type Menu struct {
	Title     string
	Link      int
	PagePrint int
	PageIndex int
}

type PresentationOptions struct {
	AspectRatioMin     configuration.AspectRatio
	AspectRatioMax     configuration.AspectRatio
	DisableAspectRatio bool
}

type Presentation struct {
	Replacers map[string]string
	Endpoints map[string]TerminalCommand
	CSS       string
	JS        string
	HTML      string
	Title     string
	Author    string
	Slides    []Slide
	Menu      []Menu
	Options   PresentationOptions
}

type TerminalOutputLine struct {
	Timestamp string
	Line      string
}

type ReadOptions struct {
	DefaultFontSize                string
	DefaultBackgroundColor         string
	DefaultTerminalFontSize        string
	DefaultTerminalFontColor       string
	DefaultTerminalBackgroundColor string
	AspectRatioMin                 configuration.AspectRatio
	AspectRatioMax                 configuration.AspectRatio
	DisableAspectRatio             bool
	EveryDashIsATransition         bool
	HideRunButton                  bool
	HidePageNumber                 bool
	KeepPagePrintOnCut             bool
}
