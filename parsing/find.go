package parsing

import (
	"strings"
)

type PatternMiddle struct {
	Start       string
	StartAlt    string
	Middle      string
	MiddleStart string
	MiddleEnd   string
	End         string
	EndAlt      string
}

func PatternMiddleSimple(start, middle, end string) PatternMiddle {
	return PatternMiddle{
		Start:       start,
		StartAlt:    string(start[len(start)-1]),
		Middle:      middle,
		MiddleStart: string(middle[0]),
		MiddleEnd:   string(middle[len(middle)-1]),
		End:         end,
		EndAlt:      string(end[0]),
	}
}

func MatchMiddle(fileContent string, pattern PatternMiddle, process func(part1, part2 string) string) string {
	for {
		start, end, part1 := FindData(fileContent, Pattern{Start: pattern.Start, End: pattern.Middle, AltStart: pattern.StartAlt, AltEnd: pattern.MiddleStart})
		if start == -1 {
			return fileContent
		}
		middle, _, part2 := FindData(fileContent[end:], Pattern{Start: pattern.Middle, End: pattern.End, AltStart: pattern.MiddleEnd, AltEnd: pattern.EndAlt})
		if middle == -1 {
			return fileContent
		}
		res := process(part1, part2)
		what := pattern.Start + part1 + pattern.Middle + part2 + pattern.End
		fileContent = strings.ReplaceAll(fileContent, what, res)
	}
}
