package cronexpr

type quartzExpression struct {
	*Expression
}

// dowFieldHandler overrides the default day of week parsing.
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
		// Remap directives
		directive.first = expr.remapDow(directive.first)
		directive.last = expr.remapDow(directive.last)

		// Populate from directives
		switch directive.kind {
		case none:
			// not supported
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

func (expr *quartzExpression) remapDow(x int) int {
	return ((x + 7) - 1) % 7
}
