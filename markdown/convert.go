package markdown

import (
	"bytes"
	"log"
	"strings"
	"sync"

	"github.com/oklog/ulid/v2"
	d2 "github.com/oktalz/goldmark-d2"
	"github.com/oktalz/present/parsing"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/mermaid"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
)

type blockData struct {
	ID   ulid.ULID
	Data string
}

var (
	blocks    []blockData
	mdPrivate goldmark.Markdown
	onceMD    sync.Once
)

func GetMD() goldmark.Markdown {
	onceMD.Do(func() {
		mdPrivate = goldmark.New(
			goldmark.WithExtensions(extension.GFM),
			goldmark.WithExtensions(&mermaid.Extender{
				NoScript: true,
			}),
			goldmark.WithExtensions(&d2.Extender{
				// Defaults when omitted
				Layout: d2dagrelayout.DefaultLayout,
				// ThemeID: d2themescatalog.,
			}),
			goldmark.WithExtensions(
				emoji.Emoji,
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				html.WithHardWraps(),
				// html.WithXHTML(),
				html.WithUnsafe(),
			),
		)
	})
	return mdPrivate
}

func ResetBlocks() {
	blocks = nil
}

func Convert(source string) (string, error) {
	if source == "" {
		return "", nil
	}
	md := GetMD()
	var buf bytes.Buffer
	if err := md.Convert([]byte(prepare(md, source)), &buf); err != nil {
		return "", err
	}
	res := buf.String()
	res = strings.TrimSuffix(res, "\n")
	res = strings.TrimPrefix(res, "\n")
	if strings.Index(res, "<p>") == 0 {
		res = strings.TrimPrefix(res, "<p>")
		res = strings.Replace(res, "</p>", "", 1)
	}
	if strings.HasSuffix(res, "</p>") {
		res = strings.TrimSuffix(res, "</p>")
		endIndex := strings.LastIndex(res, "<p>")
		if endIndex != -1 {
			res = res[:endIndex] + res[endIndex+3:]
		}
	}

	for index := len(blocks) - 1; index >= 0; index-- {
		res = strings.ReplaceAll(res, blocks[index].ID.String(), blocks[index].Data)
	}
	return res, nil
}

//revive:disable:function-length,cognitive-complexity,cyclomatic
func prepare(md goldmark.Markdown, fileContent string) string {
	fileContent = processReplace(fileContent, ".raw", ".raw.end", func(data string) string {
		return data
	})
	fileContent = processReplace(fileContent, ".raw{", "}", func(data string) string {
		return data
	})

	fileContent = processReplace(fileContent, ".tabs", ".tabs.end", func(data string) string {
		data = strings.TrimPrefix(data, "\n")
		tabs := strings.Split(data, ".tab")
		header := `<div class="tab">`
		var buff strings.Builder
		for _, tab := range tabs {
			if tab == "" {
				continue
			}
			tabID := ulid.Make().String()
			tabActive := ""
			class := " hidden-tab"
			if strings.HasPrefix(tab, "{active}") || strings.HasPrefix(tab, "{\"active\"}") {
				tabActive = " active"
				class = ""
				tab = strings.TrimPrefix(tab, "{\"active\"}")
				tab = strings.TrimPrefix(tab, "{active}")
			}
			tab = strings.TrimPrefix(tab, "{}")
			firstNewLine := strings.Index(tab, "\n")
			if firstNewLine == -1 {
				firstNewLine = len(tab)
			}
			tabName := strings.Trim(tab[:firstNewLine], " ")
			tabName = strings.Trim(tabName, "() ")
			tab = tab[firstNewLine+1:]
			contentID := CreateCleanMD(prepare(md, tab))
			header += `<button class="tablinks` + tabActive +
				`" onclick="tabChangeGlobal('` + tabID + `')" id='tab-` + tabID + `'>` + tabName + `</button>`
			tabContent := `<div class="tabcontent` + class + `" id="` + tabID + `">` + contentID.String() + `</div>`
			buff.WriteString(tabContent)
		}
		header += `</div>`
		_ = tabs
		return header + buff.String()
	})
	fileContent = ProcessReplaceMiddle(fileContent, parsing.PatternMiddle{
		Start:       ".api.pool.",
		Middle:      "{",
		End:         "}",
		StartAlt:    "",
		MiddleStart: "",
		MiddleEnd:   "",
		EndAlt:      "",
	}, func(data, display string) string {
		// .api.pool.1{option 2}
		parts := strings.SplitN(data, `.`, 2) //nolint:mnd
		if len(parts) != 2 {
			log.Println("ERROR PARSING", parts)
			return ``
		}
		// log.Println(".api.", parts)
		return `<span onclick="triggerPool('` + parts[0] + `', '` + parts[1] + `')" style="cursor: pointer;">` +
			CreateCleanMD(prepare(md, display)).String() + `</span>`
	})
	// fileContent = ProcessReplaceMiddle(fileContent, ".run", "{", "}", func(block, display string) string {
	fileContent = ProcessReplaceMiddle(fileContent, parsing.PatternMiddleSimple(".run{", "}(", ")"),
		func(block, display string) string {
			// .run.block.1{display}
			return `<span onclick="triggerBlockRun('` + block + `')" style="cursor: pointer;">` +
				CreateCleanMD(prepare(md, display)).String() + `</span>`
		})
	fileContent = processReplace(fileContent, ".center", ".center.end", func(data string) string {
		return `<div style="text-align:center">` + CreateCleanMD(prepare(md, data)).String() + `</div>`
	})
	// fileContent = ProcessReplaceMiddle(fileContent, ".link{", "}(", ")", func(page, data string) string {
	fileContent = ProcessReplaceMiddle(fileContent, parsing.PatternMiddleSimple(".link{", "}(", ")"),
		func(page, data string) string {
			return `<span onclick="setPageWithUpdate(#link#` + page + `#link#)" style="cursor: pointer;">` +
				CreateCleanMD(prepare(md, data)).String() + `</span>`
		})
	fileContent = processReplace(fileContent, ".image(", ")", func(data string) string {
		parts := strings.SplitN(data, ` `, 2) //nolint:mnd
		html := `<img src="` + parts[0] + `" `
		width := `auto`
		height := `auto`
		if len(parts) > 1 {
			wh := strings.SplitN(parts[1], `:`, 2)
			if len(wh) == 2 { //nolint:mnd
				if wh[0] != "" {
					width = wh[0]
				}
				if wh[1] != "" {
					height = wh[1]
				}
			}
		}
		html += `style="object-fit: contain; width: ` + width + `; height: ` + height + `;">`
		return html
	})
	fileContent = ProcessReplaceMiddle(fileContent, parsing.PatternMiddleSimple(".{", "}(", ")"),
		func(style, content string) string {
			var classes string
			var elemID string
			style = strings.Trim(style, `"'`)
			style = processReplace(style, ".class(", ")", func(data string) string {
				classes = data
				return ""
			})
			if classes != "" {
				classes = ` class="` + classes + `" `
			}
			style = processReplace(style, ".id(", ")", func(data string) string {
				elemID = data
				return ""
			})
			if elemID != "" {
				elemID = ` id="` + elemID + `" `
			}
			id := CreateCleanMD(prepare(md, content))
			html := `<span ` + elemID + classes + `style="` + style + `">` + id.String() + `</span>`
			return html
		})
	// fileContent = ProcessReplaceMiddle(fileContent, ".div{", "}(", ")", func(style, content string) string {
	fileContent = ProcessReplaceMiddle(fileContent, parsing.PatternMiddleSimple(".div{", "}(", ")"),
		func(style, content string) string {
			var classes string
			var elemID string
			style = strings.Trim(style, `"'`)
			style = processReplace(style, ".class(", ")", func(data string) string {
				classes = data
				return ""
			})
			if classes != "" {
				classes = ` class="` + classes + `" `
			}
			style = processReplace(style, ".id(", ")", func(data string) string {
				elemID = data
				return ""
			})
			if elemID != "" {
				elemID = ` id="` + elemID + `" `
			}
			id := CreateCleanMD(prepare(md, content))
			html := `<div ` + elemID + classes + `style="` + style + `">` + id.String() + `</div>`
			return html
		})
	// fileContent = ProcessReplaceMiddle(fileContent, ".css{", "}", ".css.end", func(style, content string) string {
	fileContent = ProcessReplaceMiddle(fileContent, parsing.PatternMiddleSimple(".css{", "}", ".css.end"),
		func(style, content string) string {
			var classes string
			var elemID string
			style = strings.Trim(style, `"'`)
			style = processReplace(style, ".class(", ")", func(data string) string {
				classes = data
				return ""
			})
			if classes != "" {
				classes = ` class="` + classes + `" `
			}
			style = processReplace(style, ".id(", ")", func(data string) string {
				elemID = data
				return ""
			})
			if elemID != "" {
				elemID = ` id="` + elemID + `" `
			}
			id := CreateCleanMD(prepare(md, content))
			html := `<div ` + elemID + classes + `style="` + style + `">` + id.String() + `</div>`
			return html
		})

	fileContent = processReplace(fileContent, ".bx{", "}", func(data string) string {
		return `<i class='bx ` + data + `'></i>`
	})

	fileContent = processReplace(fileContent, ".table", ".table.end", func(data string) string {
		html := `<table>`
		var currLine int
		trStarted := false
		tdData := ""
		split := func(data string, separators []string) []string {
			result := []string{data}
			// for example split("data.tr.td.\n", []string{".tr", ".td", "\n"})
			// should return []string{"data",".tr",".td"}
			for _, sep := range separators {
				oldResult := result
				result = []string{}
				for i := range oldResult {
					result2 := []string{}
					r := strings.Split(oldResult[i], sep)
					for index, v := range r {
						result2 = append(result2, v)
						if index < len(r)-1 && sep != "\n" {
							result2 = append(result2, sep)
						}
					}
					result = append(result, result2...)
				}
			}
			return result
		}

		// lines := strings.Split(data, "\n")
		lines := split(data, []string{"\n", ".tr", ".td"})
		// fmt.Println(lines2)
		i := 0
		for currLine = i + 1; currLine < len(lines); currLine++ {
			if lines[currLine] == ".tr" {
				if trStarted {
					if tdData != "" {
						id := CreateCleanMD(prepare(md, tdData))
						tdData = ""
						html += id.String() + `</td></tr>`
					} else {
						html += `</tr>`
					}
				} else {
					trStarted = true
				}
				if len(lines) > currLine+1 && strings.HasPrefix(lines[currLine+1], "{") {
					// find first }
					end := strings.Index(lines[currLine+1], "}")
					if end != -1 {
						css := lines[currLine+1][1:end]
						html += `<tr style="` + css + `">`
						lines[currLine+1] = strings.Replace(lines[currLine+1], "{"+css+"}", "", 1)
					} else {
						html += `<tr>`
					}
				} else {
					html += `<tr>`
				}
				continue
			}
			if strings.HasPrefix(lines[currLine], ".td") {
				line := lines[currLine]
				if tdData != "" {
					id := CreateCleanMD(prepare(md, tdData))
					html += id.String() + `</td>`
				}
				// check if next line starts with {
				if len(lines) > currLine+1 && strings.HasPrefix(lines[currLine+1], "{") {
					// find first }
					end := strings.Index(lines[currLine+1], "}")
					if end != -1 {
						css := lines[currLine+1][1:end]
						html += `</td><td style="` + css + `">`
						lines[currLine+1] = strings.Replace(lines[currLine+1], "{"+css+"}", "", 1)
					} else {
						html += `<td>`
					}
				} else {
					html += `<td>`
				}
				parts := strings.Split(line, " ")
				if len(parts) > 1 && strings.Join(parts[1:], " ") != "" {
					id := CreateCleanMD(strings.Join(parts[1:], " "))
					// solution := prepare(md, strings.Join(parts[1:], " "))
					html += id.String() + `</td>`
				}
				tdData = ""
				continue
			}
			tdData += lines[currLine] + "\n"
		}
		id := CreateCleanMD(prepare(md, tdData))
		// solution := prepare(md, tdData)
		if trStarted {
			html += id.String() + `</td></tr></table>`
		} else {
			html += `</table>`
		}

		return html
	})

	lines := strings.Split(fileContent, "\n")
	for i := range lines {
		if i >= len(lines) {
			break
		}
		index := strings.Index(lines[i], ".style ")
		isStyleBlock := false
		for index != -1 {
			lines[i], isStyleBlock = convertStyle(lines[i])
			index = strings.Index(lines[i], ".style ")
			if index != -1 {
				log.Println(lines[i])
			}
		}
		if isStyleBlock {
			var endLine int
			for endLine = i + 1; endLine < len(lines); endLine++ {
				if lines[endLine] == ".style.end" {
					break
				}
			}
			if endLine > len(lines) {
				endLine = len(lines) - 1
			}
			var centerLines []string
			for j := i + 1; j < endLine; j++ {
				centerLines = append(centerLines, lines[j])
			}
			var buf bytes.Buffer
			for index, line := range centerLines {
				if index > 0 {
					buf.WriteString("\n")
				}
				buf.WriteString(line)
			}
			solution := prepare(md, buf.String())
			id := CreateCleanMD(solution)
			lines[i] = lines[i] + "\n" + id.String() + `</div>`
			if i+1 < len(lines) && endLine+1 < len(lines) {
				lines = append(lines[:i+1], lines[endLine+1:]...)
			}
		}
		if strings.HasPrefix(lines[i], ".graph.pool.bar{") || strings.HasPrefix(lines[i], ".graph.pool.pie{") {
			// .graph.pool.bar{1}(60svh)
			id := ulid.Make().String()
			var graphType string
			var content string
			if strings.HasPrefix(lines[i], ".graph.pool.bar{") {
				graphType = "bar"
				content = strings.TrimPrefix(lines[i], ".graph.pool.bar{")
			} else {
				graphType = "pie"
				content = strings.TrimPrefix(lines[i], ".graph.pool.pie{")
			}
			data := strings.Split(content, "}")
			height := "60svh"
			if len(data) > 1 && strings.HasPrefix(data[1], "(") && strings.HasSuffix(data[1], ")") {
				height = strings.TrimSuffix(strings.TrimPrefix(data[1], "("), ")")
			}
			scales := ``
			displayLegend := `true`
			dataLabels := `,
      datalabels: {
        color: 'white',
        font: {
          size: 64
        }
      }`
			if graphType == "bar" {
				displayLegend = `false`
				dataLabels = ``
				scales = `,
      scales: {
        x: {
            ticks: {
                font: {
                    size: 15
                }
            }
        },
        y: {
          ticks: {
            stepSize: 1,
            min: 0
          },
          beginAtZero: true
        }
      }`
			}
			if len(data) > 0 {
				lines[i] = `
<div style="height: ` + height + `;">
  <canvas class="chart-` + data[0] + `" id="dynamic-chart-id` + id + `"></canvas>
</div>
<script>
  const ctx` + id + ` = document.getElementById('dynamic-chart-id` + id + `');
  const chartjs` + id + ` = new Chart(ctx` + id + `, {
    type: '` + graphType + `',
    data: {
      labels: [],
      datasets: [{
        label: '',
        data: [],
        borderWidth: 1
      }]
    },
    options: {
      plugins: {
        legend: {
          display: ` + displayLegend + `,
		  position: 'bottom',
          labels: {
            font: {
              size: 14
            }
          }
        }` + dataLabels + `
      }` + scales + `
    }
  });
  if (!charts.hasOwnProperty('ch` + data[0] + `')) {
    charts.ch` + data[0] + ` = {};
  }
  charts.ch` + data[0] + `.id` + id + ` = chartjs` + id + `;
</script>`
			} else {
				log.Println("missing graph type", lines[i])
			}
		}
	}
	fileContent = strings.Join(lines, "\n")
	return fileContent
}

func CreateCleanMD(data string) ulid.ULID {
	md := GetMD()
	var buf bytes.Buffer
	id := ulid.Make()
	if err := md.Convert([]byte(data), &buf); err != nil {
		blocks = append(blocks, blockData{ID: id, Data: ""})
		return id
	}
	solution := strings.TrimPrefix(buf.String(), "<p>")
	solution = strings.TrimSuffix(solution, "\n")
	solution = strings.TrimSuffix(solution, "</p>")

	blocks = append(blocks, blockData{ID: id, Data: solution})
	return id
}

func CreateCleanRAW(data string) ulid.ULID {
	id := ulid.Make()
	blocks = append(blocks, blockData{ID: id, Data: data})
	return id
}

func convertStyle(line string) (result string, isBlock bool) { //nolint:nonamedreturns
	index := strings.Index(line, ".style ")
	partBefore := line[:index] //nolint:gocritic
	partStyle := line[index:]
	partAfter := ""
	index = strings.Index(partStyle[1:], ".style ")
	if index != -1 {
		partAfter = partStyle[:index+1]
		partStyle = partStyle[index+1:]
	}
	parts := strings.SplitN(partStyle, `"`, 3)
	// log.Println(parts)
	_ = parts
	if len(parts) == 1 || len(parts) == 3 && parts[2] == "" {
		parts := strings.SplitN(partStyle, ` `, 2)
		parts[1] = strings.TrimPrefix(parts[1], `"`)
		parts[1] = strings.TrimSuffix(parts[1], `"`)
		line = partBefore + `<div style='` + parts[1] + `'>` + partAfter
		return line, true
	}
	if len(parts) == 3 {
		id := CreateCleanMD(parts[2])
		line = partBefore + `<div style='` + parts[1] + `'>` + id.String() + `</div>` + partAfter
	}
	if len(parts) == 2 {
		line = partBefore + `<div style='` + parts[1] + `'>` + partAfter
	}
	return line, false
}
