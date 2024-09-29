package reader

import (
	"bytes"
	"log"
	"strings"
	"text/template"

	"github.com/oktalz/present/parsing"
)

type TemplateData struct {
	Name string
	Data string
	Vars []string
}

func applyTemplate(fileContent string, templateData TemplateData) string { //nolint:funlen
	startStr := "." + templateData.Name
	for {
		start := strings.Index(fileContent, startStr)
		if start == -1 {
			break
		}
		start += len(startStr)
		content := fileContent[start:]
		end := strings.Index(content, "\n")
		if end == -1 {
			break
		}
		toReplace := content[:end]
		var data any
		if strings.HasPrefix(toReplace, "{") { //nolint:gocritic
			_, end, item := parsing.FindData(toReplace, parsing.NewShortPattern("{", "}"))
			if end+1 == len(toReplace) {
				data = item
			} else {
				contentVars := toReplace
				dataMap := map[string]string{}
				found := true
				for found {
					found = false
					contentVars = parsing.MatchMiddle(contentVars, parsing.PatternMiddleSimple("{", "}(", ")"), func(part1, part2 string) string {
						dataMap[part1] = part2
						found = true
						return ""
					})
				}
				data = dataMap
			}
		} else if strings.HasPrefix(toReplace, "(") { //nolint:revive
		} else {
			vars := strings.TrimPrefix(toReplace, " ")
			if len(templateData.Vars) == 0 {
				data = vars
			} else {
				varsData := strings.Split(vars, " ")
				dataMap := map[string]string{}
				for index := range varsData {
					if len(templateData.Vars) > index {
						key := strings.TrimPrefix(templateData.Vars[index], ".")
						dataMap[key] = varsData[index]
					}
				}
				data = dataMap
			}
		}
		tmpl, err := template.New("test").Parse(templateData.Data)
		if err != nil {
			log.Println(err)
			return fileContent
		}

		var tpl bytes.Buffer
		err = tmpl.Execute(&tpl, data)
		if err != nil {
			log.Println(err)
			return fileContent
		}
		result := tpl.String()
		hasNovalue := strings.Contains(result, "<no value>")
		if hasNovalue {
			_ = result + " "
		}
		fileContent = strings.Replace(fileContent, startStr+toReplace, result, 1)
	}
	return fileContent
}
