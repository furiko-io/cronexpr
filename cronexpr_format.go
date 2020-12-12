package cronexpr

import (
	"errors"
)

// CronFormat is an enum for different cron expression formats.
type CronFormat string

const (
	// The standard Cron format, see https://en.wikipedia.org/wiki/Cron#CRON_expression.
	// Uses the default implementation from https://github.com/gorhill/cronexpr.
	CronFormatStandard CronFormat = "standard"

	// Slight alteration to CronFormatStandard.
	// Day of week uses 1-7 for SUN-SAT, instead of 0-6 on standard implementations.
	// See http://www.quartz-scheduler.org/documentation/quartz-2.3.0/tutorials/crontrigger.html#format.
	CronFormatQuartz CronFormat = "quartz"
)

var ErrUnknownFormat = errors.New("unknown CronFormat")

// formattedExpression wraps an Expression acts as a middleware injector for different CronFormats.
type formattedExpression struct {
	*Expression
	handler expressionHandler
}

// expressionHandler supports delegation of field parsing to custom handlers.
// *Expression should always implement this interface, which provides the default implementation.
type expressionHandler interface {
	dowFieldHandler(s string) error
}

func newFormattedExpression(format CronFormat) (*formattedExpression, error) {
	e := &formattedExpression{
		Expression: &Expression{},
	}

	switch format {
	case CronFormatStandard:
		e.handler = e.Expression
	case CronFormatQuartz:
		e.handler = &quartzExpression{Expression: e.Expression}
	default:
		return nil, ErrUnknownFormat
	}

	return e, nil
}

func (e *formattedExpression) dowFieldHandler(s string) error {
	return e.handler.dowFieldHandler(s)
}
