package data

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/oklog/ulid/v2"
	configuration "github.com/oktalz/present/config"
	"github.com/oktalz/present/data/reader"
	"github.com/oktalz/present/fsnotify"
	"github.com/oktalz/present/markdown"
	"github.com/oktalz/present/types"
)

var (
	muPresentation sync.RWMutex
	presentation   types.Presentation
	hTMLNormal     []byte
	hTMLAdmin      []byte
	eTagHTMLNormal string
	eTagHTMLAdmin  string
)

func UserHTML(eTag string) ([]byte, string, int) {
	muPresentation.RLock()
	defer muPresentation.RUnlock()
	if eTag == eTagHTMLNormal {
		return nil, "", http.StatusNotModified
	}
	return hTMLNormal, eTagHTMLNormal, http.StatusOK
}

func AdminHTML(eTag string) ([]byte, string, int) {
	muPresentation.RLock()
	defer muPresentation.RUnlock()
	if eTag == eTagHTMLAdmin {
		return nil, "", http.StatusNotModified
	}

	return hTMLAdmin, eTagHTMLAdmin, http.StatusOK
}

func Presentation() types.Presentation {
	muPresentation.RLock()
	defer muPresentation.RUnlock()
	slides := make([]types.Slide, len(presentation.Slides))
	copy(slides, presentation.Slides)
	menu := make([]types.Menu, len(presentation.Menu))
	copy(menu, presentation.Menu)
	result := types.Presentation{
		Slides:    slides,
		CSS:       presentation.CSS,
		JS:        presentation.JS,
		HTML:      presentation.HTML,
		Menu:      menu,
		Title:     presentation.Title,
		Author:    presentation.Author,
		Endpoints: presentation.Endpoints,
	}
	return result
}

type Message struct {
	Data   any
	ID     string
	Author string
	Pool   string
	Value  string
	Msg    []byte
	Slide  int
	Admin  bool
	Reload bool
}

//revive:disable:function-length,cognitive-complexity,cyclomatic
func Init(server Server, config *configuration.Config) {
	filesModified := fsnotify.FileWatcher()

	// initial read
	go func() {
		filesModified <- struct{}{}
	}()

	firstRun := true
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for range filesModified {
			muPresentation.Lock()
			presentation = reader.ReadFiles()
			if presentation.Options.AspectRatioMin.String() != "" {
				go func() {
					config.AspectRatio.Min.ValueChanged <- presentation.Options.AspectRatioMin
				}()
			}
			if presentation.Options.AspectRatioMax.String() != "" {
				go func() {
					config.AspectRatio.Max.ValueChanged <- presentation.Options.AspectRatioMax
				}()
			}
			if presentation.Options.DisableAspectRatio {
				config.AspectRatio.Disable = true
				go func() {
					config.AspectRatio.Min.ValueChanged <- presentation.Options.AspectRatioMin
				}()
			}
			var err error
			for i := range presentation.Slides {
				var adminHTML string
				if presentation.Slides[i].Admin.Markdown != "" {
					adminHTML, err = markdown.Convert(presentation.Slides[i].Admin.Markdown)
					if err != nil {
						log.Println(err)
					}
				}

				res, err := markdown.Convert(presentation.Slides[i].Page.Data.Markdown)
				if err != nil {
					log.Println(err)
				}
				resHeader, err := markdown.Convert(presentation.Slides[i].Page.Header.Markdown)
				if err != nil {
					log.Println(err)
				}
				resFooter, err := markdown.Convert(presentation.Slides[i].Page.Footer.Markdown)
				if err != nil {
					log.Println(err)
				}
				presentation.Replacers["<p></p>"] = ""
				presentation.Replacers["<p> </p>"] = ""
				presentation.Replacers["<p> </p>"] = ""
				presentation.Replacers["<p>\u00a0</p>"] = ""
				for old, new := range presentation.Replacers {
					res = strings.ReplaceAll(res, old, new)
					if resHeader != "" {
						resHeader = strings.ReplaceAll(resHeader, old, new)
					}
					if resFooter != "" {
						resFooter = strings.ReplaceAll(resFooter, old, new)
					}
					if adminHTML != "" {
						adminHTML = strings.ReplaceAll(adminHTML, old, new)
					}
				}
				presentation.Slides[i].Page.Data.HTML = res
				presentation.Slides[i].Page.Header.HTML = resHeader
				presentation.Slides[i].Page.Footer.HTML = resFooter
				presentation.Slides[i].Admin.HTML = adminHTML
			}

			links := make(map[string]int, 0)
			for index, p := range presentation.Slides {
				if p.Link != "" {
					links[p.Link] = index
				}
			}
			for link, page := range links {
				for index := range len(presentation.Slides) {
					p := presentation.Slides[index]
					linkToReplace := `#link#` + link + `#link#`
					presentation.Slides[index].Page.Header.HTML = strings.ReplaceAll(
						p.Page.Header.HTML, linkToReplace, strconv.Itoa(page))
					presentation.Slides[index].Page.Data.HTML = strings.ReplaceAll(
						p.Page.Data.HTML, linkToReplace, strconv.Itoa(page))
					presentation.Slides[index].Page.Footer.HTML = strings.ReplaceAll(
						p.Page.Footer.HTML, linkToReplace, strconv.Itoa(page))
					if p.LinkNext == link {
						presentation.Slides[index].LinkNext = strconv.Itoa(page)
					}
					if p.LinkPrev == link {
						presentation.Slides[index].LinkPrev = strconv.Itoa(page)
					}
				}
			}

			markdown.ResetBlocks()

			hTMLNormal, err = regenerateHTML(presentation, *config, false)
			if err != nil {
				log.Println(err)
			} else {
				eTagHTMLNormal = ulid.Make().String()
			}

			hTMLAdmin, err = regenerateHTML(presentation, *config, true)
			if err != nil {
				log.Println(err)
			} else {
				eTagHTMLAdmin = ulid.Make().String()
			}

			server.Broadcast(Message{
				Reload: true,
			})
			muPresentation.Unlock()
			if firstRun {
				firstRun = false
				wg.Done()
			}
		}
	}()

	wg.Wait()
}
