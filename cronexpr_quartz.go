package cronexpr

import (
	"fmt"
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

	// Perform initial parse
	directives, err := genericFieldParse(s, dowDescriptor)
	if err != nil {
		return err
	}

	for _, directive := range directives {
		var ok bool
		sdirective := s[directive.sbeg:directive.send]
		switch directive.kind {
		case none:
			return fmt.Errorf("syntax error in day-of-week field: '%s'", sdirective)
		case one:
			directive.first, ok = expr.remapDow(directive.first)
			if !ok {
				return fmt.Errorf("syntax error in day-of-week field: '%s'", sdirective)
			}
			populateOne(expr.daysOfWeek, directive.first)
		case span:
			directive.first, ok = expr.remapDow(directive.first)
			if !ok {
				return fmt.Errorf("syntax error in day-of-week field: '%s'", sdirective)
			}
			directive.last, ok = expr.remapDow(directive.last)
			if !ok {
				return fmt.Errorf("syntax error in day-of-week field: '%s'", sdirective)
			}
			populateMany(expr.daysOfWeek, directive.first, directive.last, directive.step)
		case all:
			directive.first, ok = expr.remapDow(directive.first)
			if !ok {
				return fmt.Errorf("syntax error in day-of-week field: '%s'", sdirective)
			}
			directive.last, ok = expr.remapDow(directive.last)
			if !ok {
				return fmt.Errorf("syntax error in day-of-week field: '%s'", sdirective)
			}
			populateMany(expr.daysOfWeek, directive.first, directive.last, directive.step)
			expr.daysOfWeekRestricted = false
		}
	}

	return nil
}

func (expr *quartzExpression) remapDow(x int) (int, bool) {
	// only support 1-7
	if x >= 1 && x <= 7 {
		return ((x + 7) - 1) % 7, true
	}

	return 0, false
}
