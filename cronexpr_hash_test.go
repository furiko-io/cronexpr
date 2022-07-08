package cronexpr

import (
	"fmt"
	"testing"
	"time"
)

func ExampleHashString() {
	for _, str := range []string{
		"myid1",
		"myid2",
		"myid3",
	} {
		fmt.Printf("%d\n", HashString(str))
	}

	// Output:
	// 316181436714908099
	// 8964299977724969587
	// 12738036773875955645
}

func Test_hash_GetValue(t *testing.T) {
	tests := []struct {
		name  string
		value uint64
		min   int
		max   int
		want  int
	}{
		{name: "zero value", value: 0, min: 0, max: 9, want: 0},
		{name: "offset", value: 0, min: 5, max: 9, want: 5},
		{name: "value", value: 3, min: 0, max: 9, want: 3},
		{name: "modulo", value: 15, min: 0, max: 9, want: 5},
		{name: "modulo boundary", value: 10, min: 0, max: 9, want: 0},
		{name: "value with offset", value: 3, min: 5, max: 9, want: 8},
		{name: "modulo with offset", value: 13, min: 5, max: 9, want: 8},
		{name: "modulo boundary with offset", value: 10, min: 5, max: 9, want: 5},
		// int(12738036773875955645) = -5708707299833595971, ensure that we don't get a negative value out of range
		{name: "negative integer overflow mod 16", value: 12738036773875955645, min: 0, max: 15, want: 13},
		{name: "negative integer overflow mod 15", value: 12738036773875955645, min: 0, max: 14, want: 4},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			h := hash{
				value:  tt.value,
				hashed: true,
			}
			if got := h.GetValue(tt.min, tt.max); got != tt.want {
				t.Errorf("GetValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHashExpressions(t *testing.T) {
	tests := []struct {
		name  string
		expr  string
		fmt   CronFormat
		opts  []ParseOption
		times map[string][]crontimes
	}{
		{
			name: "parsing single H in minute",
			expr: "0 H * ? * * *",
			fmt:  CronFormatStandard,
			times: map[string][]crontimes{
				"myid1": { // hash mod 60 = 59
					{"2021-09-01 00:00:00", "2021-09-01 00:59:00"},
				},
				"myid2": { // hash mod 60 = 7
					{"2021-09-01 00:00:00", "2021-09-01 00:07:00"},
				},
			},
		},
		{
			name: "parsing single H in hour",
			expr: "0 0 H ? * * *",
			fmt:  CronFormatStandard,
			times: map[string][]crontimes{
				"myid1": { // hash mod 24 = 11
					{"2021-09-01 00:00:00", "2021-09-01 11:00:00"},
				},
				"myid2": { // hash mod 24 = 19
					{"2021-09-01 00:00:00", "2021-09-01 19:00:00"},
				},
			},
		},
		{
			name: "parsing single H in day of month",
			expr: "0 0 0 H * * *",
			fmt:  CronFormatStandard,
			times: map[string][]crontimes{
				"myid1": { // hash mod 28 = 27
					{"2021-09-01 00:00:00", "2021-09-28 00:00:00"},
				},
				"myid2": { // hash mod 28 = 3
					{"2021-09-01 00:00:00", "2021-09-04 00:00:00"},
				},
			},
		},
		{
			name: "parsing single H in month",
			expr: "0 0 0 ? H * *",
			fmt:  CronFormatStandard,
			times: map[string][]crontimes{
				"myid1": { // hash mod 12 = 11
					{"2021-01-01 00:00:00", "2021-12-01 00:00:00"},
				},
				"myid2": { // hash mod 12 = 7
					{"2021-01-01 00:00:00", "2021-08-01 00:00:00"},
				},
			},
		},
		{
			name: "parsing single H in day of week",
			expr: "0 0 0 ? * H *",
			fmt:  CronFormatStandard,
			times: map[string][]crontimes{
				"myid1": { // hash mod 7 = 6
					{"2021-09-01 00:00:00", "2021-09-04 00:00:00"},
				},
				"myid2": { // hash mod 7 = 3
					{"2021-09-01 00:00:00", "2021-09-08 00:00:00"},
				},
			},
		},
		{
			name: "parsing single H in day of week with quartz",
			expr: "0 0 0 ? * H *",
			fmt:  CronFormatQuartz,
			times: map[string][]crontimes{
				"myid1": { // hash mod 7 = 6
					{"2021-09-01 00:00:00", "2021-09-04 00:00:00"},
				},
				"myid2": { // hash mod 7 = 3
					{"2021-09-01 00:00:00", "2021-09-08 00:00:00"},
				},
			},
		},
		{
			name: "parsing single H in year",
			expr: "0 0 0 ? * * H",
			fmt:  CronFormatStandard,
			times: map[string][]crontimes{
				"myid1": { // hash mod 130 = 49
					{"2015-01-01 00:00:00", "2019-01-01 00:00:00"},
					{"2021-01-01 00:00:00", "0001-01-01 00:00:00"}, // out of range, more than 2099
				},
				"myid2": { // hash mod 130 = 47
					{"2015-01-01 00:00:00", "2017-01-01 00:00:00"},
					{"2021-01-01 00:00:00", "0001-01-01 00:00:00"}, // out of range, more than 2099
				},
			},
		},
		{
			name: "parsing H in minute and hour, schedule once a day",
			expr: "0 H H ? * * *",
			fmt:  CronFormatStandard,
			times: map[string][]crontimes{
				"myid1": {
					{"2021-09-01 00:00:00", "2021-09-01 11:59:00"},
					{"2021-09-01 11:59:00", "2021-09-02 11:59:00"},
				},
				"myid2": {
					{"2021-09-01 00:00:00", "2021-09-01 19:07:00"},
					{"2021-09-01 19:07:00", "2021-09-02 19:07:00"},
				},
			},
		},
		{
			// Should be identical to above
			name: "parsing H in minute and hour, schedule once a day, with implicit 0",
			expr: "H H ? * * *",
			fmt:  CronFormatStandard,
			times: map[string][]crontimes{
				"myid1": {
					{"2021-09-01 00:00:00", "2021-09-01 11:59:00"},
					{"2021-09-01 11:59:00", "2021-09-02 11:59:00"},
				},
				"myid2": {
					{"2021-09-01 00:00:00", "2021-09-01 19:07:00"},
					{"2021-09-01 19:07:00", "2021-09-02 19:07:00"},
				},
			},
		},
		{
			name: "parsing H in seconds, minute and hour, schedule once a day",
			expr: "H H H ? * * *",
			fmt:  CronFormatStandard,
			times: map[string][]crontimes{
				"myid1": {
					{"2021-09-01 00:00:00", "2021-09-01 11:59:59"},
					{"2021-09-01 11:59:59", "2021-09-02 11:59:59"},
				},
				"myid2": {
					{"2021-09-01 00:00:00", "2021-09-01 19:07:07"},
					{"2021-09-01 19:07:07", "2021-09-02 19:07:07"},
				},
			},
		},
		{
			// Should be identical to above
			name: "parsing H in minute and hour, schedule once a day, using WithHashEmptySeconds",
			expr: "H H ? * * *",
			fmt:  CronFormatStandard,
			opts: []ParseOption{WithHashEmptySeconds()},
			times: map[string][]crontimes{
				"myid1": {
					{"2021-09-01 00:00:00", "2021-09-01 11:59:59"},
					{"2021-09-01 11:59:59", "2021-09-02 11:59:59"},
				},
				"myid2": {
					{"2021-09-01 00:00:00", "2021-09-01 19:07:07"},
					{"2021-09-01 19:07:07", "2021-09-02 19:07:07"},
				},
			},
		},
		{
			name: "parsing H in minute and hour, schedule once a day, using WithHashFields",
			expr: "H H ? * * *",
			fmt:  CronFormatStandard,
			opts: []ParseOption{WithHashEmptySeconds(), WithHashFields()},
			times: map[string][]crontimes{
				"myid1": {
					{"2021-09-01 00:00:00", "2021-09-01 07:36:44"},
					{"2021-09-01 07:36:44", "2021-09-02 07:36:44"},
				},
				"myid2": {
					{"2021-09-01 00:00:00", "2021-09-01 04:43:56"},
					{"2021-09-01 04:43:56", "2021-09-02 04:43:56"},
				},
			},
		},
		{
			name: "jenkins example",
			expr: "H H(0-7) * * *",
			fmt:  CronFormatStandard,
			times: map[string][]crontimes{
				"myid1": {
					{"2021-09-01 00:00:00", "2021-09-01 03:59:00"},
					{"2021-09-01 03:59:00", "2021-09-02 03:59:00"},
				},
				"myid2": {
					{"2021-09-01 00:00:00", "2021-09-01 03:07:00"},
					{"2021-09-01 03:07:00", "2021-09-02 03:07:00"},
				},
				"myid3": {
					{"2021-09-01 00:00:00", "2021-09-01 05:49:00"},
					{"2021-09-01 05:49:00", "2021-09-02 05:49:00"},
				},
			},
		},
		{
			name: "parsing H/5 in minute",
			expr: "0 H/5 * ? * * *",
			fmt:  CronFormatStandard,
			times: map[string][]crontimes{
				"myid1": { // hash mod 5 = 4
					{"2021-09-01 00:00:00", "2021-09-01 00:04:00"},
					{"2021-09-01 00:04:00", "2021-09-01 00:09:00"},
					{"2021-09-01 00:09:00", "2021-09-01 00:14:00"},
					{"2021-09-01 00:59:00", "2021-09-01 01:04:00"},
				},
				"myid2": { // hash mod 5 = 2
					{"2021-09-01 00:00:00", "2021-09-01 00:02:00"},
					{"2021-09-01 00:02:00", "2021-09-01 00:07:00"},
					{"2021-09-01 00:59:00", "2021-09-01 01:02:00"},
				},
			},
		},
		{
			name: "parsing H/7 in minute",
			expr: "0 H/7 * ? * * *",
			fmt:  CronFormatStandard,
			times: map[string][]crontimes{
				"myid1": { // hash mod 7 = 6
					{"2021-09-01 00:00:00", "2021-09-01 00:06:00"},
					{"2021-09-01 00:06:00", "2021-09-01 00:13:00"},
					{"2021-09-01 00:13:00", "2021-09-01 00:20:00"},
					{"2021-09-01 00:55:00", "2021-09-01 01:06:00"},
				},
				"myid2": { // hash mod 7 = 3
					{"2021-09-01 00:00:00", "2021-09-01 00:03:00"},
					{"2021-09-01 00:03:00", "2021-09-01 00:10:00"},
					{"2021-09-01 00:52:00", "2021-09-01 00:59:00"},
					{"2021-09-01 00:59:00", "2021-09-01 01:03:00"},
				},
			},
		},
		{
			name: "parsing H/5 in minute and H in seconds",
			expr: "H H/5 * ? * * *",
			fmt:  CronFormatStandard,
			times: map[string][]crontimes{
				"myid1": { // hash mod 5 = 4, hash mod 60 = 59
					{"2021-09-01 00:00:00", "2021-09-01 00:04:59"},
					{"2021-09-01 00:04:59", "2021-09-01 00:09:59"},
					{"2021-09-01 00:09:59", "2021-09-01 00:14:59"},
					{"2021-09-01 00:59:59", "2021-09-01 01:04:59"},
				},
				"myid2": { // hash mod 5 = 2, hash mod 60 = 7
					{"2021-09-01 00:00:00", "2021-09-01 00:02:07"},
					{"2021-09-01 00:02:07", "2021-09-01 00:07:07"},
					{"2021-09-01 00:59:07", "2021-09-01 01:02:07"},
				},
			},
		},
		{
			// Should be identical to above
			name: "parsing H/5 in minute and WithHashEmptySeconds",
			expr: "H/5 * ? * * *",
			fmt:  CronFormatStandard,
			opts: []ParseOption{WithHashEmptySeconds()},
			times: map[string][]crontimes{
				"myid1": { // hash mod 5 = 4, hash mod 60 = 59
					{"2021-09-01 00:00:00", "2021-09-01 00:04:59"},
					{"2021-09-01 00:04:59", "2021-09-01 00:09:59"},
					{"2021-09-01 00:09:59", "2021-09-01 00:14:59"},
					{"2021-09-01 00:59:59", "2021-09-01 01:04:59"},
				},
				"myid2": { // hash mod 5 = 2, hash mod 60 = 7
					{"2021-09-01 00:00:00", "2021-09-01 00:02:07"},
					{"2021-09-01 00:02:07", "2021-09-01 00:07:07"},
					{"2021-09-01 00:59:07", "2021-09-01 01:02:07"},
				},
			},
		},
		{
			name: "parsing H/5 in minute and WithHashFields",
			expr: "H/5 * ? * * *",
			fmt:  CronFormatStandard,
			opts: []ParseOption{WithHashEmptySeconds(), WithHashFields()},
			times: map[string][]crontimes{
				"myid1": {
					{"2021-09-01 00:00:00", "2021-09-01 00:01:44"},
					{"2021-09-01 00:01:44", "2021-09-01 00:06:44"},
					{"2021-09-01 00:06:44", "2021-09-01 00:11:44"},
					{"2021-09-01 00:56:44", "2021-09-01 01:01:44"},
				},
				"myid2": {
					{"2021-09-01 00:00:00", "2021-09-01 00:03:56"},
					{"2021-09-01 00:03:56", "2021-09-01 00:08:56"},
					{"2021-09-01 00:58:56", "2021-09-01 01:03:56"},
				},
			},
		},
		{
			name: "parsing H(5-20) in minute", // once per hour within 5-20 min range
			expr: "0 H(5-20) * ? * * *",
			fmt:  CronFormatStandard,
			times: map[string][]crontimes{
				"myid1": { // hash mod 16 = 3
					{"2021-09-01 00:00:00", "2021-09-01 00:08:00"},
					{"2021-09-01 00:08:00", "2021-09-01 01:08:00"},
				},
				"myid2": { // hash mod 16 = 3
					{"2021-09-01 00:00:00", "2021-09-01 00:08:00"},
					{"2021-09-01 00:08:00", "2021-09-01 01:08:00"},
				},
				"myid3": {
					// Negative integer overflow test: int(hash("myid3")) mod 16 = -3
					// This test ensures that we don't get 00:02:00 (from 00:05:00 - 00:03:00).
					{"2021-09-01 00:00:00", "2021-09-01 00:18:00"},
					{"2021-09-01 00:18:00", "2021-09-01 01:18:00"},
				},
			},
		},
		{
			name: "parsing H(5-20)/5 in minute",
			expr: "0 H(5-20)/5 * ? * * *",
			fmt:  CronFormatStandard,
			times: map[string][]crontimes{
				"myid1": { // hash mod 5 = 4
					{"2021-09-01 00:00:00", "2021-09-01 00:09:00"},
					{"2021-09-01 00:09:00", "2021-09-01 00:14:00"},
					{"2021-09-01 00:14:00", "2021-09-01 00:19:00"},
					{"2021-09-01 00:19:00", "2021-09-01 01:09:00"},
				},
				"myid2": { // hash mod 5 = 2
					{"2021-09-01 00:00:00", "2021-09-01 00:07:00"},
					{"2021-09-01 00:07:00", "2021-09-01 00:12:00"},
					{"2021-09-01 00:12:00", "2021-09-01 00:17:00"},
					{"2021-09-01 00:17:00", "2021-09-01 01:07:00"},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			for hashID, ctimes := range tt.times {
				for _, times := range ctimes {
					from, _ := time.Parse("2006-01-02 15:04:05", times.from)
					opts := []ParseOption{WithHash(hashID)}
					opts = append(opts, tt.opts...)
					expr, err := ParseForFormat(tt.fmt, tt.expr, opts...)
					if err != nil {
						t.Errorf(`ParseForFormat("%s", "%s", WithHash("%s")) returned "%s"`,
							tt.fmt, tt.expr, hashID, err.Error())
						return
					}
					next := expr.Next(from)
					nextstr := next.Format("2006-01-02 15:04:05")
					if nextstr != times.next {
						t.Errorf(`("%s").Next("%s") = "%s", got "%s"`, tt.expr, times.from, times.next, nextstr)
					}
				}
			}
		})
	}
}
