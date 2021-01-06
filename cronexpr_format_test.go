package cronexpr

import (
	"testing"
	"time"
)

func TestFormattedExpressionsParsing(t *testing.T) {
	tests := []struct {
		name    string
		format  CronFormat
		expr    string
		wantErr bool
	}{
		{
			name:   "parsing day of week with CronFormatStandard",
			expr:   "0 0 11 ? * 2 *",
			format: CronFormatStandard,
		},
		{
			name:   "parsing day of week with CronFormatQuartz",
			expr:   "0 0 11 ? * 2 *",
			format: CronFormatQuartz,
		},
		{
			name:    "invalid day of week in CronFormatQuartz",
			expr:    "0 0 11 ? * 0 *",
			format:  CronFormatQuartz,
			wantErr: true,
		},
		{
			name:   "parsing day of week span with CronFormatQuartz",
			expr:   "0 0 11 ? * 2-3 *",
			format: CronFormatQuartz,
		},
		{
			name:   "parsing any day of week with CronFormatQuartz",
			expr:   "0 0 11 ? * * *",
			format: CronFormatQuartz,
		},
		{
			name:    "parsing invalid day of week span with CronFormatQuartz",
			expr:    "0 0 11 ? * 1- *",
			format:  CronFormatQuartz,
			wantErr: true,
		},
		{
			name:   "parsing /X format",
			expr:   "0 /5 * ? * * *",
			format: CronFormatQuartz,
		},
		{
			name:   "parsing ** format",
			expr:   "0 */5 ** ? * * *",
			format: CronFormatQuartz,
		},
		{
			name:   "parsing ?? format",
			expr:   "0 */5 * ?? * * *",
			format: CronFormatQuartz,
		},
		{
			name:   "parsing * for day of week",
			expr:   "0 0 2 * 1-7 *",
			format: CronFormatQuartz,
		},
		{
			name:   "parsing * for day of week without specifying the day",
			expr:   "0 0 11 ? *",
			format: CronFormatQuartz,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseForFormat(tt.format, tt.expr)
			if err != nil && !tt.wantErr {
				t.Errorf(`Parse("%s") returned "%s"`, tt.expr, err.Error())
			} else if err == nil && tt.wantErr {
				t.Errorf(`Parse("%s") did not return error`, tt.expr)
			}
		})
	}
}

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
		{
			name:   "parsing day of week span with CronFormatQuartz",
			expr:   "0 0 11 ? * 3-6 *", // tuesday to friday
			format: CronFormatQuartz,
			times: []crontimes{
				{"2020-12-12 00:00:00", "2020-12-15 11:00:00"},
			},
		},
		{
			name:   "parsing any day of week with CronFormatQuartz",
			expr:   "0 0 11 ? * * *",
			format: CronFormatQuartz,
			times: []crontimes{
				{"2020-12-12 00:00:00", "2020-12-12 11:00:00"},
			},
		},
		{
			name:   "parsing sunday value with CronFormatQuartz",
			expr:   "0 0 11 ? * 1 *",
			format: CronFormatQuartz,
			times: []crontimes{
				{"2020-12-12 00:00:00", "2020-12-13 11:00:00"},
			},
		},
		{
			name:   "parsing SUN literal with CronFormatQuartz",
			expr:   "0 0 11 ? * SUN *",
			format: CronFormatQuartz,
			times: []crontimes{
				{"2020-12-12 00:00:00", "2020-12-13 11:00:00"},
			},
		},
		{
			name:   "parsing sunday literal with CronFormatQuartz",
			expr:   "0 0 11 ? * sunday *",
			format: CronFormatQuartz,
			times: []crontimes{
				{"2020-12-12 00:00:00", "2020-12-13 11:00:00"},
			},
		},
		{
			name:   "parsing TUE literal with CronFormatQuartz",
			expr:   "0 0 11 ? * TUE *",
			format: CronFormatQuartz,
			times: []crontimes{
				{"2020-12-14 00:00:00", "2020-12-15 11:00:00"},
			},
		},
		{
			name:   "parsing tuesday literal with CronFormatQuartz",
			expr:   "0 0 11 ? * tuesday *",
			format: CronFormatQuartz,
			times: []crontimes{
				{"2020-12-14 00:00:00", "2020-12-15 11:00:00"},
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

func TestInvalidFormatExpressionsParsing(t *testing.T) {
	tests := []struct {
		name    string
		format  CronFormat
		expr    string
		wantErr bool
	}{
		{
			name:    "parsing out of range for end of week span with CronFormatQuartz",
			expr:    "0 0 11 ? * 1-8 *",
			format:  CronFormatQuartz,
			wantErr: true,
		},
		{
			name:    "parsing negative value for day of week span with CronFormatQuartz",
			expr:    "0 0 11 ? * -3 *",
			format:  CronFormatQuartz,
			wantErr: true,
		},
		{
			name:    "parsing out of range for end of week span with CronFormatQuartz",
			expr:    "0 0 11 ? * 8 *",
			format:  CronFormatQuartz,
			wantErr: true,
		},
		{
			name:    "parsing out of range for start of week span with CronFormatQuartz",
			expr:    "0 0 11 ? * 0-7 *",
			format:  CronFormatQuartz,
			wantErr: true,
		},
		{
			name:    "provide negative value for end of the week span with CronFormatQuartz",
			expr:    "0 0 11 ? * -2-7 *",
			format:  CronFormatQuartz,
			wantErr: true,
		},
		{
			name:    "parsing out of range for start of week span with CronFormatQuartz",
			expr:    "0 0 11 ? * 1--6 *",
			format:  CronFormatQuartz,
			wantErr: true,
		},
		{
			name:    "parsing missing second minutes hour",
			expr:    "? * 0-7 *",
			format:  CronFormatQuartz,
			wantErr: true,
		},
		{
			name:    "parsing invalid seconds",
			expr:    "60 0 11 ? * 1-7 *",
			format:  CronFormatQuartz,
			wantErr: true,
		},
		{
			name:    "parsing invalid minutes",
			expr:    "0 60 11 ? * 1-8 *",
			format:  CronFormatQuartz,
			wantErr: true,
		},
		{
			name:    "parsing invalid hour",
			expr:    "0 0 24 ? * 1-7 *",
			format:  CronFormatQuartz,
			wantErr: true,
		},
		{
			name:    "parsing invalid day of month",
			expr:    "0 0 2 32 * 1-7 *",
			format:  CronFormatQuartz,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseForFormat(tt.format, tt.expr)
			if err != nil && !tt.wantErr {
				t.Errorf(`Parse("%s") returned "%s"`, tt.expr, err.Error())
			} else if err == nil && tt.wantErr {
				t.Errorf(`Parse("%s") did not return error`, tt.expr)
			}
		})
	}
}
