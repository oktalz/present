package markdown

import (
	"strings"

	"github.com/oktalz/present/parsing"
)

func ProcessReplace(fileContent, startStr, endStr string, process func(data string) string) string {
	for {
		var raw string
		var start int
		hasNewLine := true
		start, _, raw = parsing.FindData(fileContent, parsing.Pattern{Start: "\n" + startStr, End: endStr + "\n"})
		if start == -1 {
			start, _, raw = parsing.FindData(fileContent, parsing.Pattern{Start: startStr, End: endStr})
			if start == -1 {
				return fileContent
			}
			hasNewLine = false
		}

		result := process(raw)
		if result != "" {
			result = CreateCleanRAW(result).String()
		}
		ss := startStr
		es := endStr
		if hasNewLine {
			ss = "\n" + startStr
			es = endStr + "\n"
		}
		fileContent = strings.ReplaceAll(fileContent, ss+raw+es, result)
	}
}

func processReplace(fileContent, startStr, endStr string, process func(data string) string) string {
	for {
		var raw string
		var start int
		hasNewLine := true
		start, _, raw = parsing.FindData(fileContent, parsing.Pattern{Start: "\n" + startStr, End: endStr})
		if start == -1 {
			start, _, raw = parsing.FindData(fileContent, parsing.Pattern{Start: startStr, End: endStr})
			if start == -1 {
				return fileContent
			}
			hasNewLine = false
		}

		result := process(raw)
		if result != "" {
			result = CreateCleanRAW(result).String()
		}
		ss := startStr
		es := endStr
		if hasNewLine {
			ss = "\n" + startStr
			// es = endStr + "\n"
		}
		fileContent = strings.ReplaceAll(fileContent, ss+raw+es, result)
	}
}

func ProcessReplaceMiddle(fileContent string, pattern parsing.PatternMiddle, process func(part1, part2 string) string) string {
	for {
		start, end, part1 := parsing.FindData(fileContent, parsing.Pattern{Start: pattern.Start, End: pattern.Middle, AltStart: pattern.StartAlt, AltEnd: pattern.MiddleStart})
		if start == -1 {
			return fileContent
		}
		tmp := fileContent[end:]
		middle, _, part2 := parsing.FindData(tmp, parsing.Pattern{Start: pattern.Middle, End: pattern.End, AltStart: pattern.MiddleEnd, AltEnd: pattern.EndAlt})
		if middle == -1 {
			return fileContent
		}
		result := process(part1, part2)
		result = CreateCleanRAW(result).String()
		what := pattern.Start + part1 + pattern.Middle + part2 + pattern.End
		// fileContent = strings.ReplaceAll(fileContent, "\n"+what+"\n", result)
		fileContent = strings.ReplaceAll(fileContent, "\n"+what, result)
		fileContent = strings.ReplaceAll(fileContent, what, result)
	}
}

func ProcessReplaceData(fileContent, str, result string) string {
	start := strings.Index(fileContent, str)
	if start == -1 {
		return fileContent
	}
	result = CreateCleanRAW(result).String()
	fileContent = strings.ReplaceAll(fileContent, str, result)
	return fileContent
}

func ProcessHideData(fileContent, str string) string {
	start := strings.Index(fileContent, str)
	if start == -1 {
		return fileContent
	}
	result := CreateCleanRAW(str).String()
	fileContent = strings.ReplaceAll(fileContent, str, result)
	return fileContent
}
