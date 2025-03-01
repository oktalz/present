package parsing

import (
	"math"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/oktalz/present/types"
)

type Pattern struct {
	Start    string
	End      string
	AltStart string
	AltEnd   string
}

func FindData(fileContent string, pattern Pattern) (start int, end int, result string) {
	start = strings.Index(fileContent, pattern.Start)
	if start == -1 {
		return -1, -1, ""
	}
	start += len(pattern.Start)
	content := fileContent[start:]
	index := strings.Index(content, pattern.End)
	if index == -1 {
		return -1, -1, ""
	}
	result = content[:index]
	// now we need to check if the startStr occurs before the endStr
	count := strings.Count(result, pattern.Start)
	if count > 0 || len(pattern.AltStart) > 0 || len(pattern.AltEnd) > 0 {
		altStart := pattern.AltStart
		if len(altStart) == 0 {
			altStart = pattern.Start
		}
		altEnd := pattern.AltEnd
		if len(altEnd) == 0 {
			altEnd = pattern.End
		}
		starts := findAllIndexes(content, altStart)
		ends := findAllIndexes(content, altEnd)
		// now I need to find last index, but I need to move as many times as I see startStr within the content
		last := 0
		for i := range starts {
			if last >= len(ends) {
				break
			}
			if starts[i] > ends[last] {
				break
			}
			last++
		}
		if last > 0 && last < len(ends) {
			result = content[:ends[last]]
		}
	}
	return start, start + len(result), result
}

func FindDataWithAlternative(fileContent string, pattern1, pattern2 Pattern) (start int, end int, result string) {
	if fileContent == "" {
		return -1, -1, ""
	}

	start, _, result = FindData(fileContent, pattern1)
	if start == -1 {
		start, _, result = FindData(fileContent, pattern2)
		if start == -1 {
			return -1, -1, ""
		}
	}

	return start, start + len(result), result
}

// FindDataWithCode searches for content delimited by startStr and endStr within
// fileContent. It also extracts any code block following the found content.
// Returns the start and end indexes of the found content, the content itself,
// the code block, and the code's language if specified. If the start or end
// strings are not found, or if the code block is not properly terminated,
// returns -1 for indexes and empty strings for content, code, and language.
//
//revive:disable:function-result-limit
func FindDataWithCode(fileContent, startStr, endStr string) (start, end int, result, codeBlock, language string) {
	start = strings.Index(fileContent, startStr)
	if start == -1 {
		return -1, -1, "", "", ""
	}
	start += len(startStr)
	content := fileContent[start:]
	index := strings.Index(content, endStr)
	if index == -1 {
		return -1, -1, "", "", ""
	}
	result = content[:index]
	// now we need to check if the startStr occures before the endStr
	count := strings.Count(result, startStr)
	if count > 0 {
		// we need to find the last startStr differently, we have nesting
		starts := findAllIndexes(content, startStr)
		ends := findAllIndexes(content, endStr)
		// now I need to find last index, but I need to move as many times as I see startStr within the content
		last := 0
		for i := range starts {
			if starts[i] > ends[last] {
				break
			}
			last++
		}
		result = content[:ends[last]]
	}

	// now we need to find the code.
	// code is after the result
	codeStart := strings.Index(fileContent[start+len(result)+1:], "```")
	if codeStart == -1 {
		return start, start + len(result), result, "", ""
	}
	code := fileContent[start+len(result)+1+codeStart+3:]
	codeEnd := strings.Index(code, "```")
	if codeEnd == -1 {
		return -1, -1, "", "", ""
	}
	codeStart = strings.Index(code, "\n")
	if codeStart == -1 {
		codeStart = 0
	}
	language = code[:codeStart]
	code = code[codeStart+1 : codeEnd]
	return start, start + len(result), result, code, language
}

func findAllIndexes(text, substring string) []int {
	var indexes []int
	for i := 0; i < len(text); {
		index := strings.Index(text[i:], substring)
		if index == -1 {
			break
		}
		indexes = append(indexes, i+index)
		i += index + len(substring)
	}
	return indexes
}

type ParseResult struct {
	IsStream           bool
	IsEdit             bool
	Before             []types.TerminalCommand
	Cmd                []types.TerminalCommand
	After              []types.TerminalCommand
	CodeBlockShowStart int
	CodeBlockShowEnd   int
	NewCode            string
	InjectCode         bool
	Path               string
	Lang               string
	ID                 string
	JS                 string
	Endpoint           string
}

func NewShortPattern(start, end string) Pattern {
	return Pattern{
		Start:    start,
		End:      end,
		AltStart: string(start[len(start)-1]),
		AltEnd:   string(end[0]),
	}
}

func ParseCast(cast string, code string) ParseResult { //revive:disable:function-length,cognitive-complexity,cyclomatic
	// .cast
	// + .stream
	// + .edit
	// + .before({folder}go mod init)
	// + .show(0:8)
	// + .save(main.go)
	// + .run(go run .)
	// + .after({folder}echo "done")
	// + .parallel()
	// + .parallel({folder})
	// + .path(/path/to/code)
	// + .lang(go)
	// + .id(my-id)
	// + .js(findPODid())
	// + .endpoint(my-endpoint)
	result := ParseResult{
		Before:             []types.TerminalCommand{},
		Cmd:                []types.TerminalCommand{},
		After:              []types.TerminalCommand{},
		CodeBlockShowStart: 0,
		CodeBlockShowEnd:   math.MaxInt,
	}
	result.IsStream = strings.Contains(cast, ".stream")
	result.IsEdit = strings.Contains(cast, ".edit")
	isBlock := strings.Contains(cast, ".save") || strings.Contains(cast, ".source")
	var start int
	var end int
	var data string
	var content string
	var lang string

	hasRun := -1

	data = cast
	for {
		start, end, content = FindDataWithAlternative(data,
			NewShortPattern(".endpoint(", ")"), NewShortPattern(".endpoint{", "}"))
		if start == -1 {
			break
		}
		if content != "" {
			result.Endpoint = content
		}
		data = data[end+1:]
	}

	data = cast
	for {
		start, end, content = FindDataWithAlternative(data, NewShortPattern(".id(", ")"), NewShortPattern(".id{", "}"))
		if start == -1 {
			break
		}
		if content != "" {
			result.ID = content
		}
		data = data[end+1:]
	}

	data = cast
	for {
		start, end, content = FindDataWithAlternative(data, NewShortPattern(".js(", ")"), NewShortPattern(".js{", "}"))
		if start == -1 {
			break
		}
		if content != "" {
			result.JS = content
		}
		data = data[end+1:]
	}

	data = cast
	for {
		start, end, content = FindDataWithAlternative(data,
			NewShortPattern(".path(", ")"), NewShortPattern(".path{", "}"))
		if start == -1 {
			break
		}
		if content != "" {
			result.Path = content
		}
		data = data[end+1:]
	}

	data = cast
	for {
		start, end, content = FindDataWithAlternative(data,
			NewShortPattern(".lang(", ")"), NewShortPattern(".lang{", "}"))
		if start == -1 {
			break
		}
		if content != "" {
			lang = content
		}
		data = data[end+1:]
	}

	data = cast
	for {
		start, end, content = FindDataWithAlternative(data,
			NewShortPattern(".source(", ")"), NewShortPattern(".source{", "}"))
		if start == -1 {
			break
		}
		start1, end1, folder := FindData(content, Pattern{Start: "{", End: "}"})
		if start1 != -1 {
			content = content[end1+1:]
		} else { // no special folder
			folder = result.Path
		}

		if lang == "" {
			if slices.Contains([]string{"go.mod", "go.sum"}, content) {
				// known exceptions
				lang = "go"
			} else {
				// extract extension from content
				lang = path.Ext(content)
				lang = strings.TrimPrefix(lang, ".")
			}
		}

		pth := path.Join(folder, content)
		file, err := os.ReadFile(pth)
		if err == nil {
			// just save the code, all else will be processed later again
			code = string(file)
		}
		// result.Path = content
		data = data[end+1:]
	}
	result.Lang = lang

	data = cast
	for {
		start, end, content = FindDataWithAlternative(data,
			NewShortPattern(".show(", ")"), NewShortPattern(".show{", "}"))
		if start == -1 {
			break
		}
		parts := strings.Split(content, ":")
		if len(parts) > 1 {
			result.CodeBlockShowStart, _ = strconv.Atoi(parts[0])
			if result.CodeBlockShowStart > 0 {
				result.CodeBlockShowStart-- // to have human readable indexes
			}
			result.CodeBlockShowEnd, _ = strconv.Atoi(parts[1])
		}
		data = data[end+1:]
	}

	data = cast
	for {
		start, end, content = FindDataWithAlternative(data,
			NewShortPattern(".before(", ")"), NewShortPattern(".before{", "}"))
		if start == -1 {
			break
		}
		result.Before = append(result.Before, ParseCommand(content))
		data = data[end+1:]
	}

	data = cast
	for {
		start, _, content = FindDataWithAlternative(data, NewShortPattern(".run(", ")"), NewShortPattern(".run{", "}"))
		if start == -1 {
			break
		}
		tc := ParseCommand(content)
		if isBlock {
			splitCode(code, &result, &tc)
		} else {
			tc.Code = types.Code{
				IsEmpty: true,
			}
		}

		result.Cmd = append(result.Cmd, tc)
		hasRun = len(result.Cmd) - 1
		// data = data[end+1:]
		// only one run is allowed per cast
		break //lint:ignore U1000,SA4004 // done intentionally
	}

	data = cast
	for {
		start, end, content = FindDataWithAlternative(data,
			NewShortPattern(".after(", ")"), NewShortPattern(".after{", "}"))
		if start == -1 {
			break
		}
		result.After = append(result.After, ParseCommand(content))
		data = data[end+1:]
	}

	data = cast
	for {
		start, end, content = FindDataWithAlternative(data,
			NewShortPattern(".parallel(", ")"), NewShortPattern(".parallel{", "}"))
		if start == -1 {
			break
		}
		tc := ParseCommand(content)
		result.Before = append(result.Before, tc)
		data = data[end+1:]
	}

	data = cast
	for {
		start, end, content = FindDataWithAlternative(data,
			NewShortPattern(".save(", ")"), NewShortPattern(".save{", "}"))
		if start == -1 {
			break
		}
		if hasRun == -1 {
			tc := types.TerminalCommand{
				FileName: content,
			}
			splitCode(code, &result, &tc)
			result.Cmd = append(result.Cmd, tc)
		} else {
			result.Cmd[hasRun].FileName = content
		}
		// result.FileName = content
		data = data[end+1:]
	}

	for {
		start, end, content = FindDataWithAlternative(data,
			NewShortPattern(".source(", ")"), NewShortPattern(".source{", "}"))
		if start == -1 {
			break
		}
		result.InjectCode = true
		start1, end1, _ := FindData(content, Pattern{Start: "{", End: "}"})
		if start1 != -1 {
			content = content[end1+1:]
		}
		if hasRun == -1 {
			tc := types.TerminalCommand{
				FileName: content,
			}
			splitCode(code, &result, &tc)
			result.Cmd = append(result.Cmd, tc)
		} else {
			result.Cmd[hasRun].FileName = content
		}
		if result.NewCode == "" {
			result.NewCode = code
		}
		data = data[end+1:]
	}

	result.NewCode = strings.ReplaceAll(result.NewCode, "\t", "    ")

	return result
}

func ParseCommand(command string) types.TerminalCommand {
	tc := types.TerminalCommand{
		Index: -1,
	}
	start, end, folder := FindData(command, Pattern{Start: "{", End: "}"})
	if start != -1 {
		tc.Dir = getOSPath(folder)
		tc.DirFixed = true
		command = command[end+1:]
	}
	//nolint:godox
	parts := strings.Split(command, " ") // TODO handle go run . "some param in quotes" 1 2 ...
	if len(parts) > 0 {
		tc.App = parts[0]
	}
	if len(parts) > 1 {
		tc.Cmd = parts[1:]
	}
	return tc
}

func splitCode(code string, result *ParseResult, tc *types.TerminalCommand) {
	var header string
	var footer string
	codeLines := strings.Split(code, "\n")
	code = ""
	if result.CodeBlockShowStart > len(codeLines) {
		result.CodeBlockShowStart = len(codeLines)
	}
	if result.CodeBlockShowEnd > len(codeLines) {
		result.CodeBlockShowEnd = len(codeLines)
	}
	for i := range result.CodeBlockShowStart {
		header += codeLines[i] + "\n"
	}
	until := min(result.CodeBlockShowEnd, len(codeLines))
	for i := result.CodeBlockShowStart; i < until; i++ {
		code += codeLines[i] + "\n"
	}
	for i := result.CodeBlockShowEnd; i < len(codeLines); i++ {
		footer += codeLines[i] + "\n"
	}
	// for {
	// 	if strings.HasSuffix(code, "\n") {
	// 		code = code[:len(code)-1]
	// 		continue
	// 	}
	// 	break
	// }
	if code == "\n" {
		code = ""
	}
	if result.CodeBlockShowStart != 0 || result.CodeBlockShowEnd != math.MaxInt {
		if tc.Code.Footer == "" && len(code) > 1 && code[len(code)-2] == '\n' {
			for len(code) > 1 && code[len(code)-2] == '\n' {
				code = code[:len(code)-1]
			}
			result.NewCode = code
		}
	}
	tc.Code = types.Code{
		Header: header,
		Code:   code,
		Footer: footer,
	}
	if tc.Code.Header == "" && tc.Code.Code == "" && tc.Code.Footer == "" {
		tc.Code.IsEmpty = true
	}
	if result.CodeBlockShowStart != 0 || result.CodeBlockShowEnd != math.MaxInt {
		result.NewCode = code
	}
}
