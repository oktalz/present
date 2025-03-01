package reader

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	configuration "github.com/oktalz/present/config"
	"github.com/oktalz/present/parsing"
	"github.com/oktalz/present/types"
)

func listSlideFiles(directory string) ([]string, error) {
	var slideFiles []string

	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".slide") || strings.HasSuffix(file.Name(), ".present")) {
			slideFiles = append(slideFiles, filepath.Join(directory, file.Name()))
		}
	}

	return slideFiles, nil
}

func readSlideFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

//revive:disable:function-length,cognitive-complexity,cyclomatic
func processSlides(fileContent string, ro types.ReadOptions) types.Presentation {
	var title string
	var author string
	slides := []types.Slide{}

	var slide strings.Builder
	replacers := map[string]string{}
	replacersAfter := map[string]string{}
	currentSlideTitle := ""
	currentFontSize := ro.DefaultFontSize
	currentTerminalFontSize := ro.DefaultTerminalFontSize
	currentTerminalFontColor := ro.DefaultTerminalFontColor
	currentTerminalBackgroundColor := ro.DefaultTerminalBackgroundColor
	hideRunButton := ro.HideRunButton
	hidePageNumber := ro.HidePageNumber
	keepPagePrintOnCut := ro.KeepPagePrintOnCut
	enableOverflow := false

	currentBackgroundColor := ro.DefaultBackgroundColor
	defaultEveryDashIsACut := ro.EveryDashIsATransition
	_ = defaultEveryDashIsACut
	slideDashCut := ro.EveryDashIsATransition
	notes := ""
	adminPage := ""
	pageFooter := ""
	pageHeader := ""

	templates := []TemplateData{}

	fileContent = parsing.ReplaceData(fileContent, ".template", ".template.end", func(data string) string {
		templateData := TemplateData{
			Vars: []string{},
		}
		found := false
		data = parsing.ReplaceData(data, "{", "}", func(data string) {
			templateData.Name = data
			found = true
		}, parsing.ReplaceDataOptions{Once: true, OnlyAllowOnStart: true})
		if !found {
			data = parsing.ReplaceData(data, "(", ")", func(data string) {
				templateData.Name = data
				found = true
			}, parsing.ReplaceDataOptions{Once: true, OnlyAllowOnStart: true})
		}
		data = parsing.ReplaceData(data, "(", ")", func(data2 string) string {
			vars := strings.FieldsFunc(data2, func(r rune) bool {
				return r == ' ' || r == ','
			})
			for i := range vars {
				if !strings.HasPrefix(vars[i], ".") {
					vars[i] = "." + vars[i]
				}
			}
			templateData.Vars = vars
			return ""
		}, parsing.ReplaceDataOptions{Once: true, OnlyAllowOnStart: true})
		data = strings.TrimPrefix(data, "\n")
		templateData.Data = data
		templates = append(templates, templateData)
		return ""
	})

	for i := len(templates) - 1; i >= 0; i-- {
		fileContent = applyTemplate(fileContent, templates[i])
	}

	lines := strings.Split(fileContent, "\n")

	for index := 0; index < len(lines); index++ {
		line := lines[index]

		if strings.HasPrefix(line, ".replace.after") {
			found := false
			// we have a .replace.after{pattern}(result) line
			// parsing.MatchMiddle(line, ".replace.after{", "}(", ")", func(pattern, result string) {
			parsing.MatchMiddle(line, parsing.PatternMiddle{
				Start:       ".replace.after{",
				StartAlt:    "{",
				Middle:      "}(",
				MiddleStart: "}",
				MiddleEnd:   "(",
				End:         ")",
				EndAlt:      ")",
			}, func(pattern, result string) string {
				replacersAfter[pattern] = result
				found = true
				lines[index] = ""
				return ""
			})
			if !found {
				// we have a .replace.after(pattern){result} line
				// parsing.MatchMiddle(line, ".replace.after(", "){", "}", func(pattern, result string) {
				parsing.MatchMiddle(line, parsing.PatternMiddle{
					Start:       ".replace.after(",
					StartAlt:    "(",
					Middle:      "){",
					MiddleStart: ")",
					MiddleEnd:   "{",
					End:         "}",
					EndAlt:      "}",
				}, func(pattern, result string) string {
					replacersAfter[pattern] = result
					return ""
				})
				lines[index] = ""
			}
			continue
		}
		if strings.HasPrefix(line, ".replace") {
			found := false
			// we have a .replace{pattern}(result) line
			parsing.MatchMiddle(line, parsing.PatternMiddle{
				Start:       ".replace{",
				StartAlt:    "{",
				Middle:      "}(",
				MiddleStart: "}",
				MiddleEnd:   "(",
				End:         ")",
				EndAlt:      ")",
			}, func(pattern, result string) string {
				replacers[pattern] = result
				found = true
				lines[index] = ""
				return ""
			})
			if !found {
				// we have a .replace(pattern){result} line
				parsing.MatchMiddle(line, parsing.PatternMiddle{
					Start:       ".replace(",
					StartAlt:    "(",
					Middle:      "){",
					MiddleStart: ")",
					MiddleEnd:   "{",
					End:         "}",
					EndAlt:      "}",
				}, func(pattern, result string) string {
					replacers[pattern] = result
					return ""
				})
				lines[index] = ""
			}
			continue
		}
		if strings.HasPrefix(line, ".notes") {
			// we have notes
			index++
			var notesSB strings.Builder
			for {
				line = lines[index]
				lines[index] = ""
				if strings.HasPrefix(line, ".notes.end") {
					notes = notesSB.String()
					break
				}
				notesSB.WriteString(line)
				notesSB.WriteString("<br>")
				index++
			}

			continue
		}
		if strings.HasPrefix(line, ".admin") {
			// we have admin page
			index++
			var buff strings.Builder
			for {
				line = lines[index]
				lines[index] = ""
				if strings.HasPrefix(line, ".admin.end") {
					adminPage = buff.String()
					break
				}
				buff.WriteString(line)
				buff.WriteString("\n")
				index++
			}
			continue
		}
		if strings.HasPrefix(line, ".header") {
			index++
			var buff strings.Builder
			for {
				line = lines[index]
				lines[index] = ""
				if strings.HasPrefix(line, ".header.end") {
					pageHeader = buff.String()
					break
				}
				buff.WriteString(line)
				buff.WriteString("\n")
				index++
			}
			continue
		}
		if strings.HasPrefix(line, ".footer") {
			index++
			var buff strings.Builder
			for {
				line = lines[index]
				lines[index] = ""
				if strings.HasPrefix(line, ".footer.end") {
					pageFooter = buff.String()
					break
				}
				buff.WriteString(line)
				buff.WriteString("\n")
				index++
			}
			continue
		}
		if strings.HasPrefix(line, ".---") || strings.HasPrefix(line, ".===") {
			// we have reached delimiter, see if we have anything in buffer
			if slide.Len() > 0 {
				slide := slide.String()
				if len(strings.Trim(slide, " \n")) > 0 {
					slides = append(slides, types.Slide{
						Page: types.Page{
							Data: types.Content{
								Markdown: slide,
							},
							Header: types.Content{
								Markdown: pageHeader,
							},
							Footer: types.Content{
								Markdown: pageFooter,
							},
						},
						Notes: notes,
						Admin: types.Content{
							Markdown: adminPage,
						},
						FontSize:                currentFontSize,
						TerminalFontSize:        currentTerminalFontSize,
						TerminalFontColor:       currentTerminalFontColor,
						TerminalBackgroundColor: currentTerminalBackgroundColor,
						HideRunButton:           hideRunButton,
						HidePageNumber:          hidePageNumber,
						EnableOverflow:          enableOverflow,
						BackgroundColor:         currentBackgroundColor,
						Title:                   currentSlideTitle,
					})
					notes = ""
					adminPage = ""
					pageHeader = ""
					pageFooter = ""
				}
				currentFontSize = ro.DefaultFontSize
				currentTerminalFontSize = ro.DefaultTerminalFontSize
				currentTerminalFontColor = ro.DefaultTerminalFontColor
				currentTerminalBackgroundColor = ro.DefaultTerminalBackgroundColor
				currentBackgroundColor = ro.DefaultBackgroundColor
				hideRunButton = ro.HideRunButton
				hidePageNumber = ro.HidePageNumber
				keepPagePrintOnCut = ro.KeepPagePrintOnCut
				enableOverflow = false
				currentSlideTitle = ""
			}
			slide.Reset()
			slideDashCut = ro.EveryDashIsATransition
			continue
		}
		if strings.HasPrefix(line, ".global.font-size(") && strings.HasSuffix(line, ")") {
			currentFontSize = strings.TrimPrefix(line, ".global.font-size(")
			currentFontSize = strings.TrimSuffix(currentFontSize, ")")
			ro.DefaultFontSize = currentFontSize
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".global.aspect-ratio(") && strings.HasSuffix(line, ")") {
			aspect := strings.TrimPrefix(line, ".global.aspect-ratio(")
			aspect = strings.TrimSuffix(aspect, ")")
			aspectRatio := strings.Split(aspect, ":")
			if len(aspectRatio) == 1 {
				aspectRatio = strings.Split(aspect, "x")
			}
			if len(aspectRatio) == 2 {
				width, errW := strconv.Atoi(aspectRatio[0])
				height, errH := strconv.Atoi(aspectRatio[1])
				if errW == nil && errH == nil {
					ro.AspectRatioMin = configuration.AspectRatio{
						Width:  width,
						Height: height,
					}
					ro.AspectRatioMax = configuration.AspectRatio{
						Width:  width,
						Height: height,
					}
				}
			}
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".global.aspect-ratio-min(") && strings.HasSuffix(line, ")") {
			aspect := strings.TrimPrefix(line, ".global.aspect-ratio-min(")
			aspect = strings.TrimSuffix(aspect, ")")
			aspectRatio := strings.Split(aspect, ":")
			if len(aspectRatio) == 1 {
				aspectRatio = strings.Split(aspect, "x")
			}
			if len(aspectRatio) == 2 {
				width, errW := strconv.Atoi(aspectRatio[0])
				height, errH := strconv.Atoi(aspectRatio[1])
				if errW == nil && errH == nil {
					ro.AspectRatioMin = configuration.AspectRatio{
						Width:  width,
						Height: height,
					}
				}
			}
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".global.aspect-ratio-max(") && strings.HasSuffix(line, ")") {
			aspect := strings.TrimPrefix(line, ".global.aspect-ratio-max(")
			aspect = strings.TrimSuffix(aspect, ")")
			aspectRatio := strings.Split(aspect, ":")
			if len(aspectRatio) == 1 {
				aspectRatio = strings.Split(aspect, "x")
			}
			if len(aspectRatio) == 2 {
				width, errW := strconv.Atoi(aspectRatio[0])
				height, errH := strconv.Atoi(aspectRatio[1])
				if errW == nil && errH == nil {
					ro.AspectRatioMax = configuration.AspectRatio{
						Width:  width,
						Height: height,
					}
				}
			}
			lines[index] = ""
			continue
		}
		if line == ".global.disable.aspect-ratio" {
			ro.AspectRatioMin = configuration.AspectRatio{}
			ro.AspectRatioMax = configuration.AspectRatio{}
			ro.DisableAspectRatio = true
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".global.background-color(") && strings.HasSuffix(line, ")") {
			currentBackgroundColor = strings.TrimPrefix(line, ".global.background-color(")
			currentBackgroundColor = strings.TrimSuffix(currentBackgroundColor, ")")
			ro.DefaultBackgroundColor = currentBackgroundColor
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".global.dash.is.transition") {
			ro.EveryDashIsATransition = true
			slideDashCut = ro.EveryDashIsATransition
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".slide.dash.is.transition") {
			slideDashCut = true
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".slide.dash.disable.transition") {
			slideDashCut = false
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".slide.font-size(") && strings.HasSuffix(line, ")") {
			currentFontSize = strings.TrimPrefix(line, ".slide.font-size(")
			currentFontSize = strings.TrimSuffix(currentFontSize, ")")
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".slide.terminal.font-size(") && strings.HasSuffix(line, ")") {
			currentTerminalFontSize = strings.TrimPrefix(line, ".slide.terminal.font-size(")
			currentTerminalFontSize = strings.TrimSuffix(currentTerminalFontSize, ")")
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".global.terminal.font-size(") && strings.HasSuffix(line, ")") {
			ro.DefaultTerminalFontSize = strings.TrimPrefix(line, ".global.terminal.font-size(")
			ro.DefaultTerminalFontSize = strings.TrimSuffix(ro.DefaultTerminalFontSize, ")")
			currentTerminalFontSize = ro.DefaultTerminalFontSize
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".global.terminal.font-color(") && strings.HasSuffix(line, ")") {
			ro.DefaultTerminalFontColor = strings.TrimPrefix(line, ".global.terminal.font-color(")
			ro.DefaultTerminalFontColor = strings.TrimSuffix(ro.DefaultTerminalFontColor, ")")
			currentTerminalFontColor = ro.DefaultTerminalFontColor
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".slide.terminal.font-color(") && strings.HasSuffix(line, ")") {
			currentTerminalFontColor = strings.TrimPrefix(line, ".slide.terminal.font-color(")
			currentTerminalFontColor = strings.TrimSuffix(currentTerminalFontColor, ")")
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".global.terminal.background-color(") && strings.HasSuffix(line, ")") {
			ro.DefaultTerminalBackgroundColor = strings.TrimPrefix(line, ".global.terminal.background-color(")
			ro.DefaultTerminalBackgroundColor = strings.TrimSuffix(ro.DefaultTerminalBackgroundColor, ")")
			currentTerminalBackgroundColor = ro.DefaultTerminalBackgroundColor
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".slide.terminal.background-color(") && strings.HasSuffix(line, ")") {
			currentTerminalBackgroundColor = strings.TrimPrefix(line, ".slide.terminal.background-color(")
			currentTerminalBackgroundColor = strings.TrimSuffix(currentTerminalBackgroundColor, ")")
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".global.hide.run.button") {
			ro.HideRunButton = true
			hideRunButton = ro.HideRunButton
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".slide.hide.run.button") {
			hideRunButton = true
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".global.hide.page.number") {
			ro.HidePageNumber = true
			hidePageNumber = ro.HidePageNumber
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".slide.hide.page.number") {
			hidePageNumber = true
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".global.show.page.number") {
			ro.HidePageNumber = false
			hidePageNumber = ro.HidePageNumber
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".slide.show.page.number") {
			hidePageNumber = false
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".global.keep.page.print.on.transition") {
			ro.KeepPagePrintOnCut = true
			keepPagePrintOnCut = ro.KeepPagePrintOnCut
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".slide.keep.page.print.on.transition") {
			keepPagePrintOnCut = true
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".slide.enable.overflow") {
			enableOverflow = true
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".slide.title(") && strings.HasSuffix(line, ")") {
			currentSlideTitle = strings.TrimPrefix(line, ".slide.title(")
			currentSlideTitle = strings.TrimSuffix(currentSlideTitle, ")")
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".title(") && strings.HasSuffix(line, ")") {
			title = strings.TrimPrefix(line, ".title(")
			title = strings.TrimSuffix(title, ")")
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".author(") && strings.HasSuffix(line, ")") {
			author1 := strings.TrimPrefix(line, ".author(")
			author1 = strings.TrimSuffix(author1, ")")
			if author == "" {
				author = author1
			} else {
				author = author + ", " + author1
			}
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".slide.background-color(") && strings.HasSuffix(line, ")") {
			currentBackgroundColor = strings.TrimPrefix(line, ".slide.background-color(")
			currentBackgroundColor = strings.TrimSuffix(currentBackgroundColor, ")")
			lines[index] = ""
			continue
		}
		if strings.HasPrefix(line, ".transition.clean") {
			// we have reached cut.clean delimiter, see if we have anything in buffer and clean it
			var tmp string
			if slide.Len() > 0 {
				tmp = slide.String()
				slides = append(slides, types.Slide{
					Page: types.Page{
						Data: types.Content{
							Markdown: tmp,
						},
						Header: types.Content{
							Markdown: pageHeader,
						},
						Footer: types.Content{
							Markdown: pageFooter,
						},
					},
					Notes: notes,
					Admin: types.Content{
						Markdown: adminPage,
					},
					FontSize:                currentFontSize,
					TerminalFontSize:        currentTerminalFontSize,
					TerminalFontColor:       currentTerminalFontColor,
					TerminalBackgroundColor: currentTerminalBackgroundColor,
					HideRunButton:           hideRunButton,
					HidePageNumber:          hidePageNumber,
					EnableOverflow:          enableOverflow,
					BackgroundColor:         currentBackgroundColor,
					Title:                   currentSlideTitle,
					PrintDisable:            true,
				})
				notes = ""
				adminPage = ""
				pageHeader = ""
				pageFooter = ""
				// lastIndex++ index will remain the same
			}
			slide.Reset()
			// slide.WriteString(tmp)
			continue
		}
		isDashCut := slideDashCut && strings.HasPrefix(line, "-")
		if strings.HasPrefix(line, ".transition") || isDashCut {
			// we have reached cut delimiter, see if we have anything in buffer
			type Cut struct {
				Before string
				After  string
			}
			conv := []Cut{}
			// parsing.MatchMiddle(line, ".transition{", "}(", ")", func(pattern, result string) {
			parsing.MatchMiddle(line, parsing.PatternMiddle{
				Start:       ".transition{",
				StartAlt:    "{",
				Middle:      "}(",
				MiddleStart: "}",
				MiddleEnd:   "(",
				End:         ")",
				EndAlt:      ")",
			}, func(pattern, result string) string {
				conv = append(conv, Cut{Before: pattern, After: result})
				return ""
			})
			parsing.MatchMiddle(line, parsing.PatternMiddle{
				Start:       ".transition(",
				StartAlt:    "(",
				Middle:      "){",
				MiddleStart: ")",
				MiddleEnd:   "{",
				End:         "}",
				EndAlt:      "}",
			}, func(pattern, result string) string {
				conv = append(conv, Cut{Before: pattern, After: result})
				return ""
			})
			var tmp string
			printDisable := true
			if !isDashCut && keepPagePrintOnCut {
				printDisable = false
			}
			if slide.Len() > 0 {
				tmp = slide.String()
				slides = append(slides, types.Slide{
					Page: types.Page{
						Data: types.Content{
							Markdown: tmp,
						},
						Header: types.Content{
							Markdown: pageHeader,
						},
						Footer: types.Content{
							Markdown: pageFooter,
						},
					},
					Notes: notes,
					Admin: types.Content{
						Markdown: adminPage,
					},
					FontSize:                currentFontSize,
					TerminalFontSize:        currentTerminalFontSize,
					TerminalFontColor:       currentTerminalFontColor,
					TerminalBackgroundColor: currentTerminalBackgroundColor,
					HideRunButton:           hideRunButton,
					HidePageNumber:          hidePageNumber,
					EnableOverflow:          enableOverflow,
					BackgroundColor:         currentBackgroundColor,
					Title:                   currentSlideTitle,
					PrintDisable:            printDisable,
				})
				notes = ""
				adminPage = ""
				// pageHeader = "" this needs to be copied to next slide
				// pageFooter = "" this needs to be copied to next slide
				if !isDashCut {
					for _, replace := range conv {
						tmp = strings.ReplaceAll(tmp, replace.Before, replace.After)
					}
				}
				// lastIndex++ index will remain the same
			}
			slide.Reset()
			slide.WriteString(tmp)
			if isDashCut {
				slide.WriteString(line)
				slide.WriteString("\n")
			}
			continue
		}
		if strings.HasPrefix(line, ".//") {
			// we have reached comment, ignore it
			continue
		}
		slide.WriteString(line)
		slide.WriteString("\n")
	}
	if slide.Len() > 0 {
		slides = append(slides, types.Slide{
			Page: types.Page{
				Data: types.Content{
					Markdown: slide.String(),
				},
				Header: types.Content{
					Markdown: pageHeader,
				},
				Footer: types.Content{
					Markdown: pageFooter,
				},
			},
			Notes: notes,
			Admin: types.Content{
				Markdown: adminPage,
			},
			FontSize:                currentFontSize,
			TerminalFontSize:        currentTerminalFontSize,
			TerminalFontColor:       currentTerminalFontColor,
			TerminalBackgroundColor: currentTerminalBackgroundColor,
			HideRunButton:           hideRunButton,
			HidePageNumber:          hidePageNumber,
			EnableOverflow:          enableOverflow,
			BackgroundColor:         currentBackgroundColor,
			Title:                   currentSlideTitle,
		})
		// notes = ""
		// adminPage = ""
		// pageHeader = ""
		// pageFooter = ""
		// lastIndex++
	}
	for pattern, data := range replacers {
		for index := range slides {
			slides[index].Page.Header.Markdown = strings.ReplaceAll(slides[index].Page.Header.Markdown, pattern, data)
			slides[index].Page.Data.Markdown = strings.ReplaceAll(slides[index].Page.Data.Markdown, pattern, data)
			slides[index].Page.Footer.Markdown = strings.ReplaceAll(slides[index].Page.Footer.Markdown, pattern, data)
		}
	}
	return types.Presentation{
		Options: types.PresentationOptions{
			AspectRatioMin:     ro.AspectRatioMin,
			AspectRatioMax:     ro.AspectRatioMax,
			DisableAspectRatio: ro.DisableAspectRatio,
		},
		Slides:    slides,
		Title:     title,
		Author:    author,
		Replacers: replacersAfter,
	}
}
