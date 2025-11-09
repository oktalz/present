package parsing

import (
	"strings"
)

type ReplaceTypes interface {
	string | func(data string) string | func(data string)
}

type ReplaceTypesPtr interface {
	string | func(data string) | func()
}

type ReplaceDataOptions struct {
	Once             bool
	OnlyAllowOnStart bool
}

// ReplaceDataOptionsPtr takes pointer to the fileContent.
//
// you either use Pattern or StartStr/EndStr
type ReplaceDataOptionsPtr[F ReplaceTypesPtr] struct {
	StartStr         *string
	EndStr           *string
	Pattern          *string
	Op               F
	Once             bool
	OnlyAllowOnStart bool
}

func ReplaceDataPtr[F ReplaceTypesPtr](fileContent *string, opt ReplaceDataOptionsPtr[F]) {
	once := opt.Once
	startOnIndex0 := opt.OnlyAllowOnStart
	op := opt.Op

	if opt.Pattern != nil {
		pattern := *opt.Pattern
		for {
			start := strings.Index(*fileContent, pattern)
			if start == -1 {
				return
			}
			if startOnIndex0 && start != 1 {
				return
			}
			switch any(op).(type) { // any != any
			case func(data string):
				// ok is not needed, but lets keep linters happy
				callFunc, ok := any(op).(func(string))
				if ok {
					callFunc(pattern)
				}
			case func():
				callFunc, ok := any(op).(func())
				if ok {
					callFunc()
				}
			default:
			}
			// this is different from ReplaceData, we do one replace
			*fileContent = strings.Replace(*fileContent, pattern, "", 1)
			if once {
				return
			}
		}
	}

	startStr := *opt.StartStr
	endStr := *opt.EndStr

	for {
		start, _, raw := FindData(*fileContent, Pattern{Start: startStr, End: endStr})
		if start == -1 {
			return
		}
		if startOnIndex0 && start != 1 {
			return
		}
		switch any(op).(type) { // any != any
		case func(data string):
			callFunc, ok := any(op).(func(string))
			if ok {
				callFunc(raw)
			}
		case func():
			callFunc, ok := any(op).(func())
			if ok {
				callFunc()
			}
		default:
		}
		*fileContent = strings.ReplaceAll(*fileContent, startStr+raw+endStr, "")
		if once {
			return
		}
	}
}
