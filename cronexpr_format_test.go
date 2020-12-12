package cronexpr

import (
	"testing"
	"time"
)

func TestFormattedExpressions(t *testing.T) {
	tests := []struct {
		name   string
		format CronFormat
		expr   string
		times  []crontimes
	}{
		{
			name:   "parsing day of week with CronFormatStandard",
			expr:   "0 0 11 ? * 2 *", // interprets as tuesday
			format: CronFormatStandard,
			times: []crontimes{
				{"2020-12-12 00:00:00", "2020-12-15 11:00:00"},
			},
		},
		{
			name:   "parsing day of week with CronFormatQuartz",
			expr:   "0 0 11 ? * 2 *", // interprets as monday
			format: CronFormatQuartz,
			times: []crontimes{
				{"2020-12-12 00:00:00", "2020-12-14 11:00:00"},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			for _, times := range tt.times {
				from, _ := time.Parse("2006-01-02 15:04:05", times.from)
				expr, err := ParseForFormat(tt.format, tt.expr)
				if err != nil {
					t.Errorf(`Parse("%s") returned "%s"`, tt.expr, err.Error())
					return
				}
				next := expr.Next(from)
				nextstr := next.Format("2006-01-02 15:04:05")
				if nextstr != times.next {
					t.Errorf(`("%s").Next("%s") = "%s", got "%s"`, tt.expr, times.from, times.next, nextstr)
				}
			}
		})
	}
}
