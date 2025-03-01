package parsing

import (
	"strings"
)

type ReplaceTypes interface {
	string | func(data string) string | func(data string)
}

type ReplaceDataOptions struct {
	Once             bool
	OnlyAllowOnStart bool
}

func ReplaceData[F ReplaceTypes](fileContent, startStr, endStr string, op F, opt ...ReplaceDataOptions) string {
	once := false
	startOnIndex0 := false
	if len(opt) > 0 {
		once = opt[0].Once
	}
	if len(opt) > 0 {
		startOnIndex0 = opt[0].OnlyAllowOnStart
	}

	var val F
	for {
		start, _, raw := FindData(fileContent, Pattern{Start: startStr, End: endStr})
		if start == -1 {
			return fileContent
		}
		if startOnIndex0 && start != 1 {
			return fileContent
		}
		switch any(val).(type) { // any != any
		case func(string) string:
			replaceWith := any(op).(func(string) string)(raw) //revive:disable:unchecked-type-assertion
			fileContent = strings.ReplaceAll(fileContent, startStr+raw+endStr, replaceWith)
		case func(string):
			any(op).(func(string))(raw)
			fileContent = strings.ReplaceAll(fileContent, startStr+raw+endStr, "")
		default:
			replaceWith := any(op).(string) //revive:disable:unchecked-type-assertion
			fileContent = strings.ReplaceAll(fileContent, startStr+raw+endStr, replaceWith)
		}
		if once {
			return fileContent
		}
	}
}

func ReplaceDataString(fileContent, startStr, endStr, replaceWith string) string {
	for {
		start, _, raw := FindData(fileContent, Pattern{Start: startStr, End: endStr})
		if start == -1 {
			return fileContent
		}

		fileContent = strings.ReplaceAll(fileContent, startStr+raw+endStr, replaceWith)
	}
}

func ReplaceDataFunc(fileContent, startStr, endStr string, process func(data string) string) string {
	for {
		start, _, raw := FindData(fileContent, Pattern{Start: startStr, End: endStr})
		if start == -1 {
			return fileContent
		}
		replaceWith := process(raw)
		fileContent = strings.ReplaceAll(fileContent, startStr+raw+endStr, replaceWith)
	}
}
