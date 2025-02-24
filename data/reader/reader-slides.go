package reader

import (
	"bytes"
	"log"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/oklog/ulid/v2"
	configuration "github.com/oktalz/present/config"
	"github.com/oktalz/present/markdown"
	"github.com/oktalz/present/parsing"
	"github.com/oktalz/present/types"
)

func ReadFiles() types.Presentation { //nolint:funlen,gocognit,gocyclo,cyclop,maintidx
	ro := types.ReadOptions{
		DefaultFontSize:                "5svh",
		EveryDashIsATransition:         false,
		DefaultTerminalFontSize:        "6svh",
		DefaultBackgroundColor:         "white",
		DefaultTerminalFontColor:       "black",
		DefaultTerminalBackgroundColor: "rgb(253, 246, 227)",
		HidePageNumber:                 true,
		KeepPagePrintOnCut:             false,
		AspectRatioMin:                 configuration.AspectRatio{},
		AspectRatioMax:                 configuration.AspectRatio{},
		DisableAspectRatio:             false,
	}
	endpoints := map[string]types.TerminalCommand{}

	slides, err := listSlideFiles(".")
	if err != nil {
		panic(err)
	}

	var presentationFiles types.Presentation

	cssBytes, err := os.ReadFile("present.css")
	if err == nil {
		presentationFiles.CSS = string(cssBytes)
	}
	scriptBytes, err := os.ReadFile("present.js")
	if err == nil {
		presentationFiles.JS = string(scriptBytes)
	}
	htmlBytes, err := os.ReadFile("present.html")
	if err == nil {
		presentationFiles.HTML = string(htmlBytes)
	}

	var buff strings.Builder
	for _, slide := range slides {
		content, err := readSlideFile(slide)
		if err != nil {
			panic(err)
		}
		buff.WriteString(".==========================================\n")
		buff.WriteString(content)
		if buff.Len() > 0 {
			buff.WriteString("\n")
		}
	}

	presentationFile := processSlides(buff.String(), ro)
	if presentationFile.Options.AspectRatioMin.String() != "" {
		presentationFiles.Options.AspectRatioMin = presentationFile.Options.AspectRatioMin
	}
	if presentationFile.Options.AspectRatioMax.String() != "" {
		presentationFiles.Options.AspectRatioMax = presentationFile.Options.AspectRatioMax
	}
	if presentationFile.Options.DisableAspectRatio {
		presentationFiles.Options.DisableAspectRatio = presentationFile.Options.DisableAspectRatio
	}
	presentationFiles.Slides = append(presentationFiles.Slides, presentationFile.Slides...)
	if presentationFile.Title != "" {
		presentationFiles.Title = presentationFile.Title
	}
	if presentationFile.Author != "" {
		presentationFiles.Author = presentationFile.Author
	}
	if presentationFiles.Replacers == nil {
		presentationFiles.Replacers = make(map[string]string)
	}
	maps.Copy(presentationFiles.Replacers, presentationFile.Replacers)

	presentations := make([]types.Slide, 0)
	defaultBackend := ""
	for _, slide := range presentationFiles.Slides {
		if defaultBackend != "" {
			slide.BackgroundImage = defaultBackend
		}
		hasDefaultBackground := strings.Contains(slide.Page.Data.Markdown, ".global.background(")
		if hasDefaultBackground {
			lines := strings.SplitSeq(slide.Page.Data.Markdown, "\n")
			for line := range lines {
				if strings.HasPrefix(line, ".global.background(") {
					slide.Page.Data.Markdown = strings.Replace(slide.Page.Data.Markdown, line, "", 1)
					p := strings.TrimPrefix(line, ".global.background(")
					p = strings.TrimSuffix(p, ")")
					slide.BackgroundImage = p
					defaultBackend = p
					break
				}
			}
		}
		hasBackground := strings.Contains(slide.Page.Data.Markdown, ".slide.background(")
		if hasBackground {
			lines := strings.SplitSeq(slide.Page.Data.Markdown, "\n")
			for line := range lines {
				if strings.HasPrefix(line, ".slide.background(") {
					slide.Page.Data.Markdown = strings.Replace(slide.Page.Data.Markdown, line, "", 1)
					p := strings.TrimPrefix(line, ".slide.background(")
					p = strings.TrimSuffix(p, ")")
					slide.BackgroundImage = p
					break
				}
			}
		}

		markdownData := slide.Page.Data.Markdown
		start, _, data, _, _ := parsing.FindDataWithCode(markdownData, ".api.endpoint", "\n")
		for start != -1 {
			pc := parsing.ParseCast(".endpoint"+data, "")
			if pc.Cmd[0].Dir == "" {
				pc.Cmd[0].Dir = pc.Path
				pc.Cmd[0].DirFixed = true
			}
			endpoints[pc.Endpoint] = pc.Cmd[0]
			markdownData = strings.Replace(markdownData, ".api.endpoint"+data, "", 1)
			slide.Page.Data.Markdown = strings.Replace(slide.Page.Data.Markdown, ".api.endpoint"+data, "", 1)
			start, _, data, _, _ = parsing.FindDataWithCode(markdownData, ".api.endpoint", "\n")
		}

		markdownData = slide.Page.Data.Markdown
		start, _, data, _, _ = parsing.FindDataWithCode(markdownData, ".slide.actions", "\n")
		for start != -1 {
			pc := parsing.ParseCast(data, "")
			slide.SlideCmdBefore = append(slide.SlideCmdBefore, pc.Before...)
			slide.SlideCmdAfter = append(slide.SlideCmdAfter, pc.After...)
			if slide.JS == "" {
				slide.JS = pc.JS
			} else {
				slide.JS += ";" + pc.JS
			}

			markdownData = strings.Replace(markdownData, ".slide.actions"+data, "", 1)
			slide.Page.Data.Markdown = strings.Replace(slide.Page.Data.Markdown, ".slide.actions"+data, "", 1)
			start, _, data, _, _ = parsing.FindDataWithCode(markdownData, ".slide.actions", "\n")
		}

		markdownData = slide.Page.Data.Markdown
		start, _, data, _, _ = parsing.FindDataWithCode(markdownData, ".block", "\n")
		for start != -1 {
			// .block.source{filename.ext}.show{0:8}.path{/path/to/code}.lang{go}
			pc := parsing.ParseCast(data, "")
			for {
				if strings.HasSuffix(pc.NewCode, "\n") {
					pc.NewCode = pc.NewCode[:len(pc.NewCode)-1]
					continue
				}
				break
			}
			extraClass := "code-block "
			if pc.IsEdit {
				extraClass += "code-edit "
			}
			replaceWith := "```" + pc.Lang + "\n" + pc.NewCode + "\n```\n"
			md := markdown.GetMD()
			var buf bytes.Buffer
			if err := md.Convert([]byte(replaceWith), &buf); err != nil {
				log.Println(err)
			}
			replaceWith = buf.String()
			id := pc.ID
			if id != "" {
				id = `id="` + id + `" `
			}
			replaceWith = strings.Replace(replaceWith, `<code class="`, `<code `+id+`class="`+extraClass, 1)
			replaceWith = markdown.CreateCleanRAW(replaceWith).String()

			markdownData = strings.Replace(markdownData, ".block"+data, replaceWith, 1)
			slide.Page.Data.Markdown = strings.Replace(slide.Page.Data.Markdown, ".block"+data, replaceWith, 1)
			pc.NewCode = ""

			start, _, data, _, _ = parsing.FindDataWithCode(markdownData, ".block", "\n")
		}

		markdownData = slide.Page.Data.Markdown
		start, end, data, code, language := parsing.FindDataWithCode(markdownData, ".cast", "\n")
		for start != -1 {
			// .cast.stream.edit.before{go mod init}.show{0:8}.file{main.go}.run{go run .}.after{echo "done"}.path{/path/to/code}.lang{go}
			pc := parsing.ParseCast(data, code)
			// fmt.Println(pc)
			slide.TerminalCommandBefore = append(slide.TerminalCommandBefore, pc.Before...)
			slide.TerminalCommand = append(slide.TerminalCommand, pc.Cmd...)
			slide.TerminalCommandAfter = append(slide.TerminalCommandAfter, pc.After...)
			// tc.DirFixed = true
			var c types.Cast
			if slide.Cast != nil {
				c = *slide.Cast
			}
			slide.Cast = &c
			slide.HasRun = true
			slide.HasCast = true
			if pc.Path != "" {
				slide.Path = pc.Path
			}
			slide.UseTmpFolder = slide.Path == ""
			slide.CanEdit = slide.CanEdit || pc.IsEdit
			slide.HasCastStreamed = slide.HasCastStreamed || pc.IsStream
			codeLen := len(code)
			if pc.InjectCode {
				codeLen = 0
			}
			markdownData = markdownData[end+1+codeLen:] // slightly incorrect due to ```format but acceptable
			extraClass := "code-cast "
			if pc.IsEdit {
				extraClass += "code-edit "
			}
			if pc.InjectCode {
				for {
					if strings.HasSuffix(pc.NewCode, "\n") {
						pc.NewCode = pc.NewCode[:len(pc.NewCode)-1]
						continue
					}
					break
				}
				md := markdown.GetMD()
				var buf bytes.Buffer
				if err := md.Convert([]byte("```"+pc.Lang+"\n"+pc.NewCode+"\n```\n"), &buf); err != nil {
					log.Println(err)
				}
				res := buf.String()
				id := pc.ID
				if id != "" {
					id = `id="` + id + `" `
				}
				res = strings.Replace(res, `<code class="`, `<code `+id+`class="`+extraClass, 1)
				markdownData = strings.Replace(markdownData, ".cast"+data, res, 1)
				slide.Page.Data.Markdown = strings.Replace(slide.Page.Data.Markdown, ".cast"+data, res, 1)
				pc.NewCode = ""
			} else {
				markdownData = strings.Replace(markdownData, ".cast"+data, "", 1)
				slide.Page.Data.Markdown = strings.Replace(slide.Page.Data.Markdown, ".cast"+data, "", 1)
				if pc.NewCode == "" {
					pc.NewCode = code
				}
			}
			if pc.NewCode != "" {
				toReplace := "```" + language + "\n" + code + "\n```\n"
				toReplace2 := "```" + language + "\n" + code + "```\n"
				replaceWith := "```" + language + "\n" + pc.NewCode + "\n```\n"
				md := markdown.GetMD()
				var buf bytes.Buffer
				if err := md.Convert([]byte(replaceWith), &buf); err != nil {
					log.Println(err)
				}
				replaceWith = buf.String()
				id := pc.ID
				if id != "" {
					id = `id="` + id + `" `
				}
				replaceWith = strings.Replace(replaceWith, `<code class="`, `<code `+id+`class="`+extraClass, 1)
				for strings.Contains(replaceWith, "\n\n</code>") {
					replaceWith = strings.Replace(replaceWith, "\n\n</code>", "\n</code>", 1)
				}
				slide.Page.Data.Markdown = strings.Replace(slide.Page.Data.Markdown, toReplace, replaceWith, 1)
				slide.Page.Data.Markdown = strings.Replace(slide.Page.Data.Markdown, toReplace2, replaceWith, 1)
				markdownData = strings.Replace(markdownData, code, pc.NewCode, 1)
			}

			start, end, data, code, language = parsing.FindDataWithCode(markdownData, ".cast", "\n")
		}

		hasLink := strings.Contains(slide.Page.Data.Markdown, ".slide.print.disable")
		if hasLink {
			lines := strings.Split(slide.Page.Data.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".slide.print.disable") {
					slide.PrintDisable = true
					lines = slices.Delete(lines, index, index+1)
					slide.Page.Data.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}
		hasLink = strings.Contains(slide.Page.Data.Markdown, ".slide.print.only")
		if hasLink {
			lines := strings.Split(slide.Page.Data.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".slide.print.only") {
					slide.PrintOnly = true
					lines = slices.Delete(lines, index, index+1)
					slide.Page.Data.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}

		hasLink = strings.Contains(slide.Page.Data.Markdown, ".slide.link.next(")
		if hasLink {
			lines := strings.Split(slide.Page.Data.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".slide.link.next(") {
					link := strings.TrimPrefix(line, ".slide.link.next(")
					link = strings.TrimSuffix(link, ")")
					slide.LinkNext = link
					lines = slices.Delete(lines, index, index+1)
					slide.Page.Data.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}

		hasLink = strings.Contains(slide.Page.Data.Markdown, ".slide.link.previous(")
		if hasLink {
			lines := strings.Split(slide.Page.Data.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".slide.link.previous(") {
					link := strings.TrimPrefix(line, ".slide.link.previous(")
					link = strings.TrimSuffix(link, ")")
					slide.LinkPrev = link
					lines = slices.Delete(lines, index, index+1)
					slide.Page.Data.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}

		hasLink = strings.Contains(slide.Page.Data.Markdown, ".slide.link(")
		if hasLink {
			lines := strings.Split(slide.Page.Data.Markdown, "\n")
			for index, line := range lines {
				if strings.HasPrefix(line, ".slide.link(") {
					link := strings.TrimPrefix(line, ".slide.link(")
					link = strings.TrimSuffix(link, ")")
					slide.Link = link
					lines = slices.Delete(lines, index, index+1)
					slide.Page.Data.Markdown = strings.Join(lines, "\n")
					break
				}
			}
		}
		// .link{#link1#}(:cat:) .link{#link2#}(:dog:)
		slide.Page.Data.Markdown = parsing.MatchMiddle(slide.Page.Data.Markdown, parsing.PatternMiddleSimple(".link{", "}(", ")"), func(page, data string) string {
			id := markdown.CreateCleanMD(data)
			page = `#link#` + page + `#link#`
			return `<span onclick="setPageWithUpdate(` + page + `)" style="cursor: pointer;">` + id.String() + `</span>`
		})

		data = strings.ReplaceAll(slide.Page.Data.Markdown, "\n", "")
		if !(slide.Page.Header.Markdown == "" && data == "" && slide.Page.Footer.Markdown == "") {
			presentations = append(presentations, slide)
		}
	}

	// we need to determine what page are for print only
	// or presentation only and align page numbers
	// also if we have print only slides, we need to link slide before and after,
	// the ones that are not print only
	shiftPage := 0
	for index := range presentations {
		presentations[index].PageIndex = index
		if presentations[index].PrintDisable {
			shiftPage++
		}
		presentations[index].PagePrint = index + 1 - shiftPage
		if presentations[index].PrintOnly && index > 0 {
			// find first before, first after and set links (if not set already)
			indexBefore := index - 1
			for indexBefore > 0 {
				if presentations[indexBefore].PrintOnly {
					indexBefore--
				} else {
					break
				}
			}
			indexAfter := index + 1
			for indexAfter < len(presentations)-1 {
				if presentations[indexAfter].PrintOnly {
					indexAfter++
				} else {
					break
				}
			}
			if presentations[indexBefore].Link == "" {
				presentations[indexBefore].Link = ulid.Make().String()
			}
			if presentations[indexAfter].Link == "" {
				presentations[indexAfter].Link = ulid.Make().String()
			}
			presentations[indexBefore].LinkNext = presentations[indexAfter].Link
			presentations[indexAfter].LinkPrev = presentations[indexBefore].Link
		}
	}

	// ok now setup the menu
	menu := make([]types.Menu, 0)
	for i, p := range presentations {
		title := ""
		data := p.Page.Header.Markdown
		if data == "" {
			data = p.Page.Data.Markdown
		}
		lines := strings.SplitSeq(data, "\n")
		for line := range lines {
			ldata := line
			ldata = strings.ReplaceAll(ldata, "&#41;", ")")
			ldata = strings.ReplaceAll(ldata, "&#40;", "(")
			ldata = strings.ReplaceAll(ldata, "&#123;", "{")
			ldata = strings.ReplaceAll(ldata, "&#125;", "}")
			ldata = strings.ReplaceAll(ldata, "&#46;", ".")
			ldata = strings.ReplaceAll(ldata, "&#95;", "_")
			ldata = strings.ReplaceAll(ldata, "&#45;", "-")
			ldata = strings.ReplaceAll(ldata, "&#34;", `"`)
			index := strings.LastIndex(ldata, "#")
			if index > -1 {
				title = ldata[index+1:]
				index := strings.LastIndex(title, `"`)
				if index > -1 {
					title = title[index+1:]
				}
				title = strings.Trim(title, ` #*()`)
				break
			}
		}
		if p.Title != "" {
			title = p.Title
		}
		if len(menu) > 0 {
			if menu[len(menu)-1].Title != title {
				menu = append(menu, types.Menu{
					Link:      i,
					PageIndex: p.PageIndex,
					PagePrint: p.PagePrint,
					Title:     title,
				})
			}
		} else {
			menu = append(menu, types.Menu{
				Link:      i,
				PageIndex: p.PageIndex,
				PagePrint: p.PagePrint,
				Title:     title,
			})
		}
	}

	maps.Copy(endpoints, presentationFiles.Endpoints)

	return types.Presentation{
		Options:   presentationFiles.Options,
		Slides:    presentations,
		CSS:       presentationFiles.CSS,
		JS:        presentationFiles.JS,
		HTML:      presentationFiles.HTML,
		Menu:      menu,
		Title:     presentationFiles.Title,
		Author:    presentationFiles.Author,
		Replacers: presentationFiles.Replacers,
		Endpoints: endpoints,
	}
}
