package cronexpr

import (
	"fmt"
)

var (
	quartzDowTokens = map[string]int{
		`1`: 0, `sun`: 0, `sunday`: 0,
		`2`: 1, `mon`: 1, `monday`: 1,
		`3`: 2, `tue`: 2, `tuesday`: 2,
		`4`: 3, `wed`: 3, `wednesday`: 3,
		`5`: 4, `thu`: 4, `thursday`: 4,
		`6`: 5, `fri`: 5, `friday`: 5,
		`7`: 6, `sat`: 6, `saturday`: 6,
	}

	quartzDowDescriptor = fieldDescriptor{
		name:         "day-of-week",
		min:          0,
		max:          6,
		defaultList:  genericDefaultList[0:7],
		valuePattern: `0?[1-7]|sun|mon|tue|wed|thu|fri|sat|sunday|monday|tuesday|wednesday|thursday|friday|saturday`,
		atoi: func(s string) int {
			return quartzDowTokens[s]
		},
	}
)

// quartzExpression implements custom parsing for the Quartz scheduler format.
type quartzExpression struct {
	*Expression
}

// dowFieldHandler overrides the default day of week parsing.
// Day of week uses 1-7 for SUN-SAT, instead of 0-6 on standard implementations.
func (expr *quartzExpression) dowFieldHandler(s string) error {
	expr.daysOfWeekRestricted = true
	expr.daysOfWeek = make(map[int]bool)
	expr.lastWeekDaysOfWeek = make(map[int]bool)
	expr.specificWeekDaysOfWeek = make(map[int]bool)

	// Use custom descriptor
	directives, err := genericFieldParse(s, quartzDowDescriptor)
	if err != nil {
		return err
	}

	for _, directive := range directives {
		sdirective := s[directive.sbeg:directive.send]
		switch directive.kind {
		case none:
			// not implemented.
			return fmt.Errorf("syntax error in day-of-week field: '%s'", sdirective)
		case one:
			populateOne(expr.daysOfWeek, directive.first)
		case span:
			populateMany(expr.daysOfWeek, directive.first, directive.last, directive.step)
		case all:
			populateMany(expr.daysOfWeek, directive.first, directive.last, directive.step)
			expr.daysOfWeekRestricted = false
		}
	}

	return nil
}
