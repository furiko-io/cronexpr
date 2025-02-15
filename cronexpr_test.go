/*!
 * Copyright 2013 Raymond Hill
 *
 * Modifications 2020 - HashiCorp
 *
 * Project: github.com/gorhill/cronexpr
 * File: cronexpr_test.go
 * Version: 1.0
 * License: pick the one which suits you best:
 *   GPL v3 see <https://www.gnu.org/licenses/gpl.html>
 *   APL v2 see <http://www.apache.org/licenses/LICENSE-2.0>
 *
 */

package cronexpr

/******************************************************************************/

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/******************************************************************************/

type crontimes struct {
	from string
	next string
}

type crontest struct {
	expr   string
	layout string
	times  []crontimes
}

var crontests = []crontest{
	// Seconds
	{
		"* * * * * * *",
		"2006-01-02 15:04:05",
		[]crontimes{
			{"2013-01-01 00:00:00", "2013-01-01 00:00:01"},
			{"2013-01-01 00:00:59", "2013-01-01 00:01:00"},
			{"2013-01-01 00:59:59", "2013-01-01 01:00:00"},
			{"2013-01-01 23:59:59", "2013-01-02 00:00:00"},
			{"2013-02-28 23:59:59", "2013-03-01 00:00:00"},
			{"2016-02-28 23:59:59", "2016-02-29 00:00:00"},
			{"2012-12-31 23:59:59", "2013-01-01 00:00:00"},
		},
	},

	// every 5 Second
	{
		"*/5 * * * * * *",
		"2006-01-02 15:04:05",
		[]crontimes{
			{"2013-01-01 00:00:00", "2013-01-01 00:00:05"},
			{"2013-01-01 00:00:59", "2013-01-01 00:01:00"},
			{"2013-01-01 00:59:59", "2013-01-01 01:00:00"},
			{"2013-01-01 23:59:59", "2013-01-02 00:00:00"},
			{"2013-02-28 23:59:59", "2013-03-01 00:00:00"},
			{"2016-02-28 23:59:59", "2016-02-29 00:00:00"},
			{"2012-12-31 23:59:59", "2013-01-01 00:00:00"},
		},
	},

	// Minutes
	{
		"* * * * *",
		"2006-01-02 15:04:05",
		[]crontimes{
			{"2013-01-01 00:00:00", "2013-01-01 00:01:00"},
			{"2013-01-01 00:00:59", "2013-01-01 00:01:00"},
			{"2013-01-01 00:59:00", "2013-01-01 01:00:00"},
			{"2013-01-01 23:59:00", "2013-01-02 00:00:00"},
			{"2013-02-28 23:59:00", "2013-03-01 00:00:00"},
			{"2016-02-28 23:59:00", "2016-02-29 00:00:00"},
			{"2012-12-31 23:59:00", "2013-01-01 00:00:00"},
		},
	},

	// Minutes with interval
	{
		"17-43/5 * * * *",
		"2006-01-02 15:04:05",
		[]crontimes{
			{"2013-01-01 00:00:00", "2013-01-01 00:17:00"},
			{"2013-01-01 00:16:59", "2013-01-01 00:17:00"},
			{"2013-01-01 00:30:00", "2013-01-01 00:32:00"},
			{"2013-01-01 00:50:00", "2013-01-01 01:17:00"},
			{"2013-01-01 23:50:00", "2013-01-02 00:17:00"},
			{"2013-02-28 23:50:00", "2013-03-01 00:17:00"},
			{"2016-02-28 23:50:00", "2016-02-29 00:17:00"},
			{"2012-12-31 23:50:00", "2013-01-01 00:17:00"},
		},
	},

	// Minutes interval, list
	{
		"15-30/4,55 * * * *",
		"2006-01-02 15:04:05",
		[]crontimes{
			{"2013-01-01 00:00:00", "2013-01-01 00:15:00"},
			{"2013-01-01 00:16:00", "2013-01-01 00:19:00"},
			{"2013-01-01 00:30:00", "2013-01-01 00:55:00"},
			{"2013-01-01 00:55:00", "2013-01-01 01:15:00"},
			{"2013-01-01 23:55:00", "2013-01-02 00:15:00"},
			{"2013-02-28 23:55:00", "2013-03-01 00:15:00"},
			{"2016-02-28 23:55:00", "2016-02-29 00:15:00"},
			{"2012-12-31 23:54:00", "2012-12-31 23:55:00"},
			{"2012-12-31 23:55:00", "2013-01-01 00:15:00"},
		},
	},

	// Days of week
	{
		"0 0 * * MON",
		"Mon 2006-01-02 15:04",
		[]crontimes{
			{"2013-01-01 00:00:00", "Mon 2013-01-07 00:00"},
			{"2013-01-28 00:00:00", "Mon 2013-02-04 00:00"},
			{"2013-12-30 00:30:00", "Mon 2014-01-06 00:00"},
		},
	},
	{
		"0 0 * * friday",
		"Mon 2006-01-02 15:04",
		[]crontimes{
			{"2013-01-01 00:00:00", "Fri 2013-01-04 00:00"},
			{"2013-01-28 00:00:00", "Fri 2013-02-01 00:00"},
			{"2013-12-30 00:30:00", "Fri 2014-01-03 00:00"},
		},
	},
	{
		"0 0 * * 6,7",
		"Mon 2006-01-02 15:04",
		[]crontimes{
			{"2013-01-01 00:00:00", "Sat 2013-01-05 00:00"},
			{"2013-01-28 00:00:00", "Sat 2013-02-02 00:00"},
			{"2013-12-30 00:30:00", "Sat 2014-01-04 00:00"},
		},
	},
	{
		"0 0 * * 5-7",
		"Mon 2006-01-02 15:04",
		[]crontimes{
			{"2013-01-01 00:00:00", "Fri 2013-01-04 00:00"},
			{"2013-01-28 00:00:00", "Fri 2013-02-01 00:00"},
			{"2013-12-30 00:30:00", "Fri 2014-01-03 00:00"},
		},
	},

	// Specific days of week
	{
		"0 0 * * 6#5",
		"Mon 2006-01-02 15:04",
		[]crontimes{
			{"2013-09-02 00:00:00", "Sat 2013-11-30 00:00"},
		},
	},

	// Work day of month
	{
		"0 0 14W * *",
		"Mon 2006-01-02 15:04",
		[]crontimes{
			{"2013-03-31 00:00:00", "Mon 2013-04-15 00:00"},
			{"2013-08-31 00:00:00", "Fri 2013-09-13 00:00"},
		},
	},

	// Work day of month -- end of month
	{
		"0 0 30W * *",
		"Mon 2006-01-02 15:04",
		[]crontimes{
			{"2013-03-02 00:00:00", "Fri 2013-03-29 00:00"},
			{"2013-06-02 00:00:00", "Fri 2013-06-28 00:00"},
			{"2013-09-02 00:00:00", "Mon 2013-09-30 00:00"},
			{"2013-11-02 00:00:00", "Fri 2013-11-29 00:00"},
		},
	},

	// Last day of month
	{
		"0 0 L * *",
		"Mon 2006-01-02 15:04",
		[]crontimes{
			{"2013-09-02 00:00:00", "Mon 2013-09-30 00:00"},
			{"2014-01-01 00:00:00", "Fri 2014-01-31 00:00"},
			{"2014-02-01 00:00:00", "Fri 2014-02-28 00:00"},
			{"2016-02-15 00:00:00", "Mon 2016-02-29 00:00"},
		},
	},

	// Last work day of month
	{
		"0 0 LW * *",
		"Mon 2006-01-02 15:04",
		[]crontimes{
			{"2013-09-02 00:00:00", "Mon 2013-09-30 00:00"},
			{"2013-11-02 00:00:00", "Fri 2013-11-29 00:00"},
			{"2014-08-15 00:00:00", "Fri 2014-08-29 00:00"},
		},
	},

	// Zero padded months
	{
		"0 0 * 04 * *",
		"Mon 2006-01-02 15:04",
		[]crontimes{
			{"2013-09-02 00:00:00", "Tue 2014-04-01 00:00"},
			{"2014-04-03 03:00:00", "Fri 2014-04-04 00:00"},
			{"2014-08-15 00:00:00", "Wed 2015-04-01 00:00"},
		},
	},

	// TODO: more tests
}

func TestExpressions(t *testing.T) {
	for _, test := range crontests {
		for _, times := range test.times {
			from, _ := time.Parse("2006-01-02 15:04:05", times.from)
			expr, err := Parse(test.expr)
			if err != nil {
				t.Errorf(`Parse("%s") returned "%s"`, test.expr, err.Error())
			}
			next := expr.Next(from)
			nextstr := next.Format(test.layout)
			if nextstr != times.next {
				t.Errorf(`("%s").Next("%s") = "%s", got "%s"`, test.expr, times.from, times.next, nextstr)
			}
		}
	}
}

/******************************************************************************/

func TestZero(t *testing.T) {
	from, _ := time.Parse("2006-01-02", "2013-08-31")
	next := MustParse("* * * * * 1980").Next(from)
	if next.IsZero() == false {
		t.Error(`("* * * * * 1980").Next("2013-08-31").IsZero() returned 'false', expected 'true'`)
	}

	next = MustParse("* * * * * 2050").Next(from)
	if next.IsZero() == true {
		t.Error(`("* * * * * 2050").Next("2013-08-31").IsZero() returned 'true', expected 'false'`)
	}

	next = MustParse("* * * * * 2099").Next(time.Time{})
	if next.IsZero() == false {
		t.Error(`("* * * * * 2014").Next(time.Time{}).IsZero() returned 'true', expected 'false'`)
	}
}

/******************************************************************************/

type crontestWithError struct {
	crontest
	err error
}

func TestInvalidExpr(t *testing.T) {
	var tests = []crontestWithError{
		{
			crontest: crontest{
				expr: "20-10 * * * *",
			},
			err: errors.New("beginning of range (20) beyond end of range (10): 20-10"),
		},
		{
			crontest: crontest{
				expr: "1-60 * * * *",
			},
			err: errors.New("syntax error in minute field: '1-60'"),
		},
		{
			crontest: crontest{
				expr: "* 10-5 * * *",
			},
			err: errors.New("beginning of range (10) beyond end of range (5): 10-5"),
		},
		{
			crontest: crontest{
				expr: "* 10-24 * * *",
			},
			err: errors.New("syntax error in hour field: '10-24'"),
		},
		{
			crontest: crontest{
				expr: "* * 0-10 * *",
			},
			err: errors.New("syntax error in day-of-month field: '0-10'"),
		},
		{
			crontest: crontest{
				expr: "* * 31-10 * *",
			},
			err: errors.New("beginning of range (31) beyond end of range (10): 31-10"),
		},
		{
			crontest: crontest{
				expr: "* * * 0-11 *",
			},
			err: errors.New("syntax error in month field: '0-11'"),
		},
		{
			crontest: crontest{
				expr: "* * * 11-5 *",
			},
			err: errors.New("beginning of range (11) beyond end of range (5): 11-5"),
		},
		{
			crontest: crontest{
				expr: "* * * * 1-8",
			},
			err: errors.New("syntax error in day-of-week field: '1-8'"),
		},
		{
			crontest: crontest{
				expr: "* * * * 5-2",
			},
			err: errors.New("beginning of range (5) beyond end of range (2): 5-2"),
		},
	}

	for _, test := range tests {
		{
			_, err := Parse(test.expr)
			if assert.Error(t, err) {
				assert.Equal(t, test.err, err)
			}
		}
	}
}

/******************************************************************************/

func TestNextN(t *testing.T) {
	expected := []string{
		"Sat, 30 Nov 2013 00:00:00",
		"Sat, 29 Mar 2014 00:00:00",
		"Sat, 31 May 2014 00:00:00",
		"Sat, 30 Aug 2014 00:00:00",
		"Sat, 29 Nov 2014 00:00:00",
	}
	from, _ := time.Parse("2006-01-02 15:04:05", "2013-09-02 08:44:30")
	result := MustParse("0 0 * * 6#5").NextN(from, uint(len(expected)))
	if len(result) != len(expected) {
		t.Errorf(`MustParse("0 0 * * 6#5").NextN("2013-09-02 08:44:30", 5):\n"`)
		t.Errorf(`  Expected %d returned time values but got %d instead`, len(expected), len(result))
	}
	for i, next := range result {
		nextStr := next.Format("Mon, 2 Jan 2006 15:04:15")
		if nextStr != expected[i] {
			t.Errorf(`MustParse("0 0 * * 6#5").NextN("2013-09-02 08:44:30", 5):\n"`)
			t.Errorf(`  result[%d]: expected "%s" but got "%s"`, i, expected[i], nextStr)
		}
	}
}

func TestNextN_every5min(t *testing.T) {
	expected := []string{
		"Mon, 2 Sep 2013 08:45:00",
		"Mon, 2 Sep 2013 08:50:00",
		"Mon, 2 Sep 2013 08:55:00",
		"Mon, 2 Sep 2013 09:00:00",
		"Mon, 2 Sep 2013 09:05:00",
	}
	from, _ := time.Parse("2006-01-02 15:04:05", "2013-09-02 08:44:32")
	result := MustParse("*/5 * * * *").NextN(from, uint(len(expected)))
	if len(result) != len(expected) {
		t.Errorf(`MustParse("*/5 * * * *").NextN("2013-09-02 08:44:30", 5):\n"`)
		t.Errorf(`  Expected %d returned time values but got %d instead`, len(expected), len(result))
	}
	for i, next := range result {
		nextStr := next.Format("Mon, 2 Jan 2006 15:04:05")
		if nextStr != expected[i] {
			t.Errorf(`MustParse("*/5 * * * *").NextN("2013-09-02 08:44:30", 5):\n"`)
			t.Errorf(`  result[%d]: expected "%s" but got "%s"`, i, expected[i], nextStr)
		}
	}
}

func TestPeriodicConfig_DSTChange_Transitions(t *testing.T) {
	locName := "America/Los_Angeles"
	loc, err := time.LoadLocation(locName)
	require.NoError(t, err)

	cases := []struct {
		name     string
		pattern  string
		initTime time.Time
		expected []time.Time
	}{
		{
			"normal time",
			"0 2 * * * 2019",
			time.Date(2019, time.February, 7, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.February, 7, 2, 0, 0, 0, loc),
				time.Date(2019, time.February, 8, 2, 0, 0, 0, loc),
				time.Date(2019, time.February, 9, 2, 0, 0, 0, loc),
			},
		},
		{
			"Spring forward but not in switch time",
			"0 4 * * * 2019",
			time.Date(2019, time.March, 9, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.March, 9, 4, 0, 0, 0, loc),
				time.Date(2019, time.March, 10, 4, 0, 0, 0, loc),
				time.Date(2019, time.March, 11, 4, 0, 0, 0, loc),
			},
		},
		{
			"Spring forward at a skipped time odd",
			"2 2 * * * 2019",
			time.Date(2019, time.March, 9, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.March, 9, 2, 2, 0, 0, loc),
				// no time in March 10!
				time.Date(2019, time.March, 11, 2, 2, 0, 0, loc),
				time.Date(2019, time.March, 12, 2, 2, 0, 0, loc),
			},
		},
		{
			"Spring forward at a skipped time",
			"1 2 * * * 2019",
			time.Date(2019, time.March, 9, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.March, 9, 2, 1, 0, 0, loc),
				// no time in March 8!
				time.Date(2019, time.March, 11, 2, 1, 0, 0, loc),
				time.Date(2019, time.March, 12, 2, 1, 0, 0, loc),
			},
		},
		{
			"Spring forward at a skipped time boundary",
			"0 2 * * * 2019",
			time.Date(2019, time.March, 9, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.March, 9, 2, 0, 0, 0, loc),
				// no time in March 8!
				time.Date(2019, time.March, 11, 2, 0, 0, 0, loc),
				time.Date(2019, time.March, 12, 2, 0, 0, 0, loc),
			},
		},
		{
			"Spring forward at a boundary of repeating time",
			"0 1 * * * 2019",
			time.Date(2019, time.March, 9, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.March, 9, 1, 0, 0, 0, loc),
				time.Date(2019, time.March, 10, 0, 0, 0, 0, loc).Add(1 * time.Hour),
				time.Date(2019, time.March, 11, 1, 0, 0, 0, loc),
				time.Date(2019, time.March, 12, 1, 0, 0, 0, loc),
			},
		},
		{
			"Fall back: before transition",
			"30 0 * * * 2019",
			time.Date(2019, time.November, 3, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.November, 3, 0, 30, 0, 0, loc),
				time.Date(2019, time.November, 4, 0, 30, 0, 0, loc),
				time.Date(2019, time.November, 5, 0, 30, 0, 0, loc),
				time.Date(2019, time.November, 6, 0, 30, 0, 0, loc),
			},
		},
		{
			"Fall back: after transition",
			"30 3 * * * 2019",
			time.Date(2019, time.November, 3, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.November, 3, 3, 30, 0, 0, loc),
				time.Date(2019, time.November, 4, 3, 30, 0, 0, loc),
				time.Date(2019, time.November, 5, 3, 30, 0, 0, loc),
				time.Date(2019, time.November, 6, 3, 30, 0, 0, loc),
			},
		},
		{
			"Fall back: after transition starting in repeated span before",
			"30 3 * * * 2019",
			time.Date(2019, time.November, 3, 0, 10, 0, 0, loc).Add(1 * time.Hour),
			[]time.Time{
				time.Date(2019, time.November, 3, 3, 30, 0, 0, loc),
				time.Date(2019, time.November, 4, 3, 30, 0, 0, loc),
				time.Date(2019, time.November, 5, 3, 30, 0, 0, loc),
				time.Date(2019, time.November, 6, 3, 30, 0, 0, loc),
			},
		},
		{
			"Fall back: after transition starting in repeated span after",
			"30 3 * * * 2019",
			time.Date(2019, time.November, 3, 0, 10, 0, 0, loc).Add(2 * time.Hour),
			[]time.Time{
				time.Date(2019, time.November, 3, 3, 30, 0, 0, loc),
				time.Date(2019, time.November, 4, 3, 30, 0, 0, loc),
				time.Date(2019, time.November, 5, 3, 30, 0, 0, loc),
				time.Date(2019, time.November, 6, 3, 30, 0, 0, loc),
			},
		},
		{
			"Fall back: in repeated region",
			"30 1 * * * 2019",
			time.Date(2019, time.November, 3, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.November, 3, 0, 30, 0, 0, loc).Add(1 * time.Hour),
				time.Date(2019, time.November, 3, 0, 30, 0, 0, loc).Add(2 * time.Hour),
				time.Date(2019, time.November, 4, 1, 30, 0, 0, loc),
				time.Date(2019, time.November, 5, 1, 30, 0, 0, loc),
				time.Date(2019, time.November, 6, 1, 30, 0, 0, loc),
			},
		},
		{
			"Fall back: in repeated region boundary",
			"0 1 * * * 2019",
			time.Date(2019, time.November, 3, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.November, 3, 0, 0, 0, 0, loc).Add(1 * time.Hour),
				time.Date(2019, time.November, 3, 0, 0, 0, 0, loc).Add(2 * time.Hour),
				time.Date(2019, time.November, 4, 1, 0, 0, 0, loc),
				time.Date(2019, time.November, 5, 1, 0, 0, 0, loc),
				time.Date(2019, time.November, 6, 1, 0, 0, 0, loc),
			},
		},
		{
			"Fall back: in repeated region boundary 2",
			"0 2 * * * 2019",
			time.Date(2019, time.November, 3, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.November, 3, 0, 0, 0, 0, loc).Add(3 * time.Hour),
				time.Date(2019, time.November, 4, 2, 0, 0, 0, loc),
				time.Date(2019, time.November, 5, 2, 0, 0, 0, loc),
				time.Date(2019, time.November, 6, 2, 0, 0, 0, loc),
			},
		},
		{
			"Fall back: in repeated region, starting from within region",
			"30 1 * * * 2019",
			time.Date(2019, time.November, 3, 0, 40, 0, 0, loc).Add(1 * time.Hour),
			[]time.Time{
				time.Date(2019, time.November, 3, 0, 30, 0, 0, loc).Add(2 * time.Hour),
				time.Date(2019, time.November, 4, 1, 30, 0, 0, loc),
				time.Date(2019, time.November, 5, 1, 30, 0, 0, loc),
				time.Date(2019, time.November, 6, 1, 30, 0, 0, loc),
			},
		},
		{
			"Fall back: in repeated region, starting from within region 2",
			"30 1 * * * 2019",
			time.Date(2019, time.November, 3, 0, 40, 0, 0, loc).Add(2 * time.Hour),
			[]time.Time{
				time.Date(2019, time.November, 4, 1, 30, 0, 0, loc),
				time.Date(2019, time.November, 5, 1, 30, 0, 0, loc),
				time.Date(2019, time.November, 6, 1, 30, 0, 0, loc),
			},
		},
		{
			"Fall back: wildcard",
			"30 * * * * 2019",
			time.Date(2019, time.November, 3, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.November, 3, 0, 30, 0, 0, loc),
				time.Date(2019, time.November, 3, 0, 30, 0, 0, loc).Add(1 * time.Hour),
				time.Date(2019, time.November, 3, 0, 30, 0, 0, loc).Add(2 * time.Hour),
				time.Date(2019, time.November, 3, 2, 30, 0, 0, loc),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			expr := MustParse(c.pattern)

			starting := c.initTime
			for _, next := range c.expected {
				n := expr.Next(starting)
				if next != n {
					t.Fatalf("next(%v) = %v not %v", starting, next, n)
				}

				starting = next
			}
		})
	}
}

func TestPeriodicConfig_DSTChange_Transitions_LordHowe(t *testing.T) {
	locName := "Australia/Lord_Howe"
	loc, err := time.LoadLocation(locName)
	require.NoError(t, err)

	cases := []struct {
		name     string
		pattern  string
		initTime time.Time
		expected []time.Time
	}{
		{
			"normal time",
			"0 2 * * * 2019",
			time.Date(2019, time.February, 7, 1, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.February, 7, 2, 0, 0, 0, loc),
				time.Date(2019, time.February, 8, 2, 0, 0, 0, loc),
				time.Date(2019, time.February, 9, 2, 0, 0, 0, loc),
			},
		},
		{
			"backward: non repeated portion of the hour",
			"3 1 * * * 2019",
			time.Date(2019, time.April, 6, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.April, 6, 1, 3, 0, 0, loc),
				time.Date(2019, time.April, 7, 1, 3, 0, 0, loc),
				time.Date(2019, time.April, 8, 1, 3, 0, 0, loc),
				time.Date(2019, time.April, 9, 1, 3, 0, 0, loc),
			},
		},
		{
			"backward: repeated portion of the hour",
			"31 1 * * * 2019",
			time.Date(2019, time.April, 6, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.April, 6, 1, 31, 0, 0, loc),
				time.Date(2019, time.April, 7, 0, 31, 0, 0, loc).Add(60 * time.Minute),
				time.Date(2019, time.April, 7, 0, 31, 0, 0, loc).Add(90 * time.Minute),
				time.Date(2019, time.April, 8, 1, 31, 0, 0, loc),
				time.Date(2019, time.April, 9, 1, 31, 0, 0, loc),
			},
		},
		{
			"forward: skipped portion of the hour",
			"3 2 * * * 2019",
			time.Date(2019, time.October, 5, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.October, 5, 2, 3, 0, 0, loc),
				// no Oct 6
				time.Date(2019, time.October, 7, 2, 3, 0, 0, loc),
				time.Date(2019, time.October, 8, 2, 3, 0, 0, loc),
				time.Date(2019, time.October, 9, 2, 3, 0, 0, loc),
			},
		},
		{
			"forward: non-skipped portion of the hour",
			"31 2 * * * 2019",
			time.Date(2019, time.October, 5, 0, 0, 0, 0, loc),
			[]time.Time{
				time.Date(2019, time.October, 5, 2, 31, 0, 0, loc),
				// no Oct 6
				time.Date(2019, time.October, 7, 2, 31, 0, 0, loc),
				time.Date(2019, time.October, 8, 2, 31, 0, 0, loc),
				time.Date(2019, time.October, 9, 2, 31, 0, 0, loc),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			expr := MustParse(c.pattern)

			starting := c.initTime
			for _, next := range c.expected {
				n := expr.Next(starting)
				if next != n {
					t.Fatalf("next(%v) = %v not %v", starting, next, n)
				}

				starting = next
			}
		})
	}
}

func TestNext_DaylightSaving_Property(t *testing.T) {
	locName := "America/Los_Angeles"
	loc, err := time.LoadLocation(locName)
	if err != nil {
		t.Fatalf("failed to get location: %v", err)
	}

	cronExprs := []string{
		"* * * * *",
		"0 2 * * *",
		"* 1 * * *",
	}

	times := []time.Time{
		// spring forward
		time.Date(2019, time.March, 11, 0, 0, 0, 0, loc),
		time.Date(2019, time.March, 10, 0, 0, 0, 0, loc),
		time.Date(2019, time.March, 11, 0, 0, 0, 0, loc),

		// leap backwards
		time.Date(2019, time.November, 4, 0, 0, 0, 0, loc),
		time.Date(2019, time.November, 5, 0, 0, 0, 0, loc),
		time.Date(2019, time.November, 6, 0, 0, 0, 0, loc),
	}

	testSpan := 4 * time.Hour

	testCase := func(t *testing.T, cronExpr string, init time.Time) {
		cron := MustParse(cronExpr)

		prevNext := init
		for start := init; start.Before(init.Add(testSpan)); start = start.Add(1 * time.Minute) {
			next := cron.Next(start)
			if !next.After(start) {
				t.Fatalf("next(%v) = %v is not after start time", start, next)
			}

			if next.Before(prevNext) {
				t.Fatalf("next(%v) = %v reverted back in time from %v", start, next, prevNext)
			}

			if strings.HasPrefix(cronExpr, "* * ") {
				if next.Sub(start) != time.Minute {
					t.Fatalf("next(%v) = %v should be the next minute", start, next)
				}
			}

			prevNext = next
		}
	}

	for _, cron := range cronExprs {
		for _, startTime := range times {
			t.Run(fmt.Sprintf("%v: %v", cron, startTime), func(t *testing.T) {
				testCase(t, cron, startTime)
			})
		}
	}
}

func TestNext_DaylightSaving_Property_LordHowe(t *testing.T) {
	// Lord Howe, Australia is at GMT+1100 April-October and GMT+1030 otherwise.
	//
	// On April 7, 2019, at when clock approches 2am, the clock
	// transitions to 1.30am.
	//
	// On October 6, when the clock approaches 2am, the clock transitions
	// to 2.30am.
	locName := "Australia/Lord_Howe"
	loc, err := time.LoadLocation(locName)
	if err != nil {
		t.Fatalf("failed to get location: %v", err)
	}

	cronExprs := []string{
		"* * * * *",
		"0 2 * * *",
		"* 1 * * *",
		"35 1 * * *",
		"5 2 * * *",
	}

	times := []time.Time{
		// spring forward
		time.Date(2019, time.April, 6, 0, 0, 0, 0, loc),
		time.Date(2019, time.April, 7, 0, 0, 0, 0, loc),
		time.Date(2019, time.April, 8, 0, 0, 0, 0, loc),

		// leap backwards
		time.Date(2019, time.October, 5, 0, 0, 0, 0, loc),
		time.Date(2019, time.October, 6, 0, 0, 0, 0, loc),
		time.Date(2019, time.October, 7, 0, 0, 0, 0, loc),
	}

	testSpan := 4 * time.Hour

	testCase := func(t *testing.T, cronExpr string, init time.Time) {
		cron := MustParse(cronExpr)

		prevNext := init
		for start := init; start.Before(init.Add(testSpan)); start = start.Add(1 * time.Minute) {
			next := cron.Next(start)
			if !next.After(start) {
				t.Fatalf("next(%v) = %v is not after start time", start, next)
			}

			if next.Before(prevNext) {
				t.Fatalf("next(%v) = %v reverted back in time from %v", start, next, prevNext)
			}

			if strings.HasPrefix(cronExpr, "* * ") {
				if next.Sub(start) != time.Minute {
					t.Fatalf("next(%v) = %v should be the next minute", start, next)
				}
			}

			prevNext = next
		}
	}

	for _, cron := range cronExprs {
		for _, startTime := range times {
			t.Run(fmt.Sprintf("%v: %v", cron, startTime), func(t *testing.T) {
				testCase(t, cron, startTime)
			})
		}
	}
}

func TestNext_DaylightSaving_Property_Brazil(t *testing.T) {
	// Until 2018, Brazil/Sao Paulo and some South American countries used
	// to transition for daylight savings at midnight.
	//
	// When the clock approaches 2018-11-04 midnight, the clock transitions to 1am.
	locName := "America/Sao_Paulo"
	loc, err := time.LoadLocation(locName)
	if err != nil {
		t.Fatalf("failed to get location: %v", err)
	}

	cronExprs := []string{
		"* * * * *",
		"0 2 * * *",
		"* 1 * * *",
		"5 1 * * *",
		"5 23 * * *",
	}

	times := []time.Time{
		// spring forward
		time.Date(2018, time.February, 16, 22, 0, 0, 0, loc),
		time.Date(2018, time.February, 17, 22, 0, 0, 0, loc),
		time.Date(2018, time.February, 18, 22, 0, 0, 0, loc),

		// leap backwards
		time.Date(2018, time.November, 3, 23, 0, 0, 0, loc),
		time.Date(2018, time.November, 3, 23, 0, 0, 0, loc),
		time.Date(2018, time.November, 3, 23, 0, 0, 0, loc),
	}

	testSpan := 4 * time.Hour

	testCase := func(t *testing.T, cronExpr string, init time.Time) {
		cron := MustParse(cronExpr)

		prevNext := init
		for start := init; start.Before(init.Add(testSpan)); start = start.Add(1 * time.Minute) {
			next := cron.Next(start)
			if !next.After(start) {
				t.Fatalf("next(%v) = %v is not after start time", start, next)
			}

			if next.Before(prevNext) {
				t.Fatalf("next(%v) = %v reverted back in time from %v", start, next, prevNext)
			}

			if strings.HasPrefix(cronExpr, "* * ") {
				if next.Sub(start) != time.Minute {
					t.Fatalf("next(%v) = %v should be the next minute", start, next)
				}
			}

			prevNext = next
		}
	}

	for _, cron := range cronExprs {
		for _, startTime := range times {
			t.Run(fmt.Sprintf("%v: %v", cron, startTime), func(t *testing.T) {
				testCase(t, cron, startTime)
			})
		}
	}
}

func TestNext_DaylightSaving_AdditionalTests(t *testing.T) {
	var (
		locationSaoPaulo, _ = time.LoadLocation("America/Sao_Paulo")
		locationNewYork, _  = time.LoadLocation("America/New_York")
		locationSantiago, _ = time.LoadLocation("America/Santiago")
	)

	type crontest struct {
		name   string
		expr   string
		layout string
		times  []crontimes
		loc    *time.Location
	}
	crontests := []crontest{
		{
			name:   "Daylight Savings changeover for America/Sao Paulo, run hourly",
			expr:   "0 * ? * *",
			layout: time.RFC3339,
			loc:    locationSaoPaulo,
			times: []crontimes{
				// Move back 1 hour on Sunday, 18 February, 00:00 from UTC-2 to UTC-3
				{"2018-02-17T22:12:00-02:00", "2018-02-17T23:00:00-02:00"},
				{"2018-02-17T23:12:00-02:00", "2018-02-17T23:00:00-03:00"}, // should still run here
				{"2018-02-17T23:12:00-03:00", "2018-02-18T00:00:00-03:00"},
				{"2018-02-18T00:12:00-03:00", "2018-02-18T01:00:00-03:00"},
				// Move forward 1 hour on Sunday, 4 November, 00:00 from UTC-3 to UTC-2
				{"2018-11-03T22:12:00-03:00", "2018-11-03T23:00:00-03:00"},
				{"2018-11-03T23:12:00-03:00", "2018-11-04T01:00:00-02:00"},
				{"2018-11-04T01:12:00-02:00", "2018-11-04T02:00:00-02:00"},
			},
		},
		{
			name:   "Daylight Savings changeover for America/Sao Paulo, run hourly with minute/second offset",
			expr:   "50 10 * ? * * *",
			layout: time.RFC3339,
			loc:    locationSaoPaulo,
			times: []crontimes{
				// Move back 1 hour on Sunday, 18 February 2018, 00:00 from UTC-2 to UTC-3
				{"2018-02-17T22:12:00-02:00", "2018-02-17T23:10:50-02:00"},
				{"2018-02-17T23:12:00-02:00", "2018-02-17T23:10:50-03:00"}, // should still run here
				{"2018-02-17T23:59:59-02:00", "2018-02-17T23:10:50-03:00"},
				{"2018-02-17T23:10:50-03:00", "2018-02-18T00:10:50-03:00"},
				{"2018-02-17T23:12:00-03:00", "2018-02-18T00:10:50-03:00"},
				{"2018-02-18T00:12:00-03:00", "2018-02-18T01:10:50-03:00"},
				// Move forward 1 hour on Sunday, 4 November 2018, 00:00 from UTC-3 to UTC-2
				{"2018-11-03T22:12:00-03:00", "2018-11-03T23:10:50-03:00"},
				{"2018-11-03T23:12:00-03:00", "2018-11-04T01:10:50-02:00"},
				{"2018-11-03T23:59:59-03:00", "2018-11-04T01:10:50-02:00"},
				{"2018-11-04T01:10:50-02:00", "2018-11-04T02:10:50-02:00"},
				{"2018-11-04T01:12:00-02:00", "2018-11-04T02:10:50-02:00"},
			},
		},
		{
			name:   "Daylight Savings changeover for America/Sao Paulo, run every 15 mins with minute offset",
			expr:   "0 10/15 * ? * * *",
			layout: time.RFC3339,
			loc:    locationSaoPaulo,
			times: []crontimes{
				// Move back 1 hour on Sunday, 18 February 2018, 00:00 from UTC-2 to UTC-3
				{"2018-02-17T23:00:00-02:00", "2018-02-17T23:10:00-02:00"},
				{"2018-02-17T23:10:00-02:00", "2018-02-17T23:25:00-02:00"},
				{"2018-02-17T23:25:00-02:00", "2018-02-17T23:40:00-02:00"},
				{"2018-02-17T23:40:00-02:00", "2018-02-17T23:55:00-02:00"},
				{"2018-02-17T23:55:00-02:00", "2018-02-17T23:10:00-03:00"},
				{"2018-02-17T23:10:00-03:00", "2018-02-17T23:25:00-03:00"},
				{"2018-02-17T23:55:00-03:00", "2018-02-18T00:10:00-03:00"},
				{"2018-02-18T00:55:00-03:00", "2018-02-18T01:10:00-03:00"},
				// Move forward 1 hour on Sunday, 4 November 2018, 00:00 from UTC-3 to UTC-2
				{"2018-11-03T23:00:00-03:00", "2018-11-03T23:10:00-03:00"},
				{"2018-11-03T23:10:00-03:00", "2018-11-03T23:25:00-03:00"},
				{"2018-11-03T23:25:00-03:00", "2018-11-03T23:40:00-03:00"},
				{"2018-11-03T23:40:00-03:00", "2018-11-03T23:55:00-03:00"},
				{"2018-11-03T23:55:00-03:00", "2018-11-04T01:10:00-02:00"},
				{"2018-11-04T01:10:00-02:00", "2018-11-04T01:25:00-02:00"},
				{"2018-11-04T01:55:00-02:00", "2018-11-04T02:10:00-02:00"},
			},
		},
		{
			name:   "Daylight Savings changeover for America/Sao Paulo, run every 15 sec",
			expr:   "0/15 10 * ? * * *",
			layout: time.RFC3339,
			loc:    locationSaoPaulo,
			times: []crontimes{
				// Move back 1 hour on Sunday, 18 February 2018, 00:00 from UTC-2 to UTC-3
				{"2018-02-17T23:00:00-02:00", "2018-02-17T23:10:00-02:00"},
				{"2018-02-17T23:10:00-02:00", "2018-02-17T23:10:15-02:00"},
				{"2018-02-17T23:10:15-02:00", "2018-02-17T23:10:30-02:00"},
				{"2018-02-17T23:10:30-02:00", "2018-02-17T23:10:45-02:00"},
				{"2018-02-17T23:10:45-02:00", "2018-02-17T23:10:00-03:00"},
				{"2018-02-17T23:10:45-03:00", "2018-02-18T00:10:00-03:00"},
				{"2018-02-18T00:10:45-03:00", "2018-02-18T01:10:00-03:00"},
				// Move forward 1 hour on Sunday, 4 November 2018, 00:00 from UTC-3 to UTC-2
				{"2018-11-03T23:00:00-03:00", "2018-11-03T23:10:00-03:00"},
				{"2018-11-03T23:10:00-03:00", "2018-11-03T23:10:15-03:00"},
				{"2018-11-03T23:10:15-03:00", "2018-11-03T23:10:30-03:00"},
				{"2018-11-03T23:10:30-03:00", "2018-11-03T23:10:45-03:00"},
				{"2018-11-03T23:10:45-03:00", "2018-11-04T01:10:00-02:00"},
				{"2018-11-04T01:10:45-02:00", "2018-11-04T02:10:00-02:00"},
			},
		},
		{
			name:   "Daylight Savings changeover for America/Santiago, run daily",
			expr:   "0 0 ? * *",
			layout: time.RFC3339,
			loc:    locationSantiago,
			times: []crontimes{
				// Move forward 1 hour on Sunday, 5 September 2021, 00:00 from UTC-4 to UTC-3
				{"2021-09-03T00:00:00-04:00", "2021-09-04T00:00:00-04:00"},
				{"2021-09-04T00:00:00-04:00", "2021-09-06T00:00:00-03:00"}, // skips 1 day
				{"2021-09-05T01:00:00-03:00", "2021-09-06T00:00:00-03:00"},
				{"2021-09-06T00:00:00-03:00", "2021-09-07T00:00:00-03:00"},
				// Move backward 1 hour on Sunday, 3 April 2022, 00:00 from UTC-3 to UTC-4
				// Assumes information from https://www.timeanddate.com/time/change/chile/santiago?year=2022
				{"2022-03-31T00:00:00-03:00", "2022-04-01T00:00:00-03:00"},
				{"2022-04-01T00:00:00-03:00", "2022-04-02T00:00:00-03:00"},
				{"2022-04-02T00:00:00-03:00", "2022-04-03T00:00:00-04:00"}, // run later by 1h
				{"2022-04-02T23:59:59-03:00", "2022-04-03T00:00:00-04:00"},
				{"2022-04-02T23:00:00-04:00", "2022-04-03T00:00:00-04:00"},
				{"2022-04-02T23:59:59-04:00", "2022-04-03T00:00:00-04:00"},
				{"2022-04-03T00:00:00-04:00", "2022-04-04T00:00:00-04:00"},
			},
		},
		{
			name:   "Daylight Savings changeover for America/Santiago, run daily at 1am",
			expr:   "0 1 ? * *",
			layout: time.RFC3339,
			loc:    locationSantiago,
			times: []crontimes{
				// Move forward 1 hour on Sunday, 5 September 2021, 00:00 from UTC-4 to UTC-3
				{"2021-09-03T01:00:00-04:00", "2021-09-04T01:00:00-04:00"},
				{"2021-09-04T01:00:00-04:00", "2021-09-05T01:00:00-03:00"},
				{"2021-09-05T01:00:00-03:00", "2021-09-06T01:00:00-03:00"},
				{"2021-09-06T01:00:00-03:00", "2021-09-07T01:00:00-03:00"},
				// Move backward 1 hour on Sunday, 3 April 2022, 00:00 from UTC-3 to UTC-4
				{"2022-03-31T01:00:00-03:00", "2022-04-01T01:00:00-03:00"},
				{"2022-04-01T01:00:00-03:00", "2022-04-02T01:00:00-03:00"},
				{"2022-04-02T01:00:00-03:00", "2022-04-03T01:00:00-04:00"},
				{"2022-04-03T01:00:00-04:00", "2022-04-04T01:00:00-04:00"},
			},
		},
		{
			name:   "Daylight Savings changeover for America/Santiago, run on the 1st of each month at 12am",
			expr:   "0 0 1 * *",
			layout: time.RFC3339,
			loc:    locationSantiago,
			times: []crontimes{
				// Move forward 1 hour on Sunday, 5 September 2021, 00:00 from UTC-4 to UTC-3
				{"2021-08-01T00:00:00-04:00", "2021-09-01T00:00:00-04:00"},
				{"2021-09-01T00:00:00-04:00", "2021-10-01T00:00:00-03:00"}, // skipped 1h
				{"2021-10-01T00:00:00-03:00", "2021-11-01T00:00:00-03:00"},
				// Move backward 1 hour on Sunday, 3 April 2022, 00:00 from UTC-3 to UTC-4
				{"2022-03-01T00:00:00-03:00", "2022-04-01T00:00:00-03:00"},
				{"2022-04-01T00:00:00-03:00", "2022-05-01T00:00:00-04:00"}, // added 1h
				{"2022-05-01T00:00:00-04:00", "2022-06-01T00:00:00-04:00"},
			},
		},
		{
			name:   "Daylight Savings changeover for America/New_York, run hourly",
			expr:   "0 * ? * *",
			layout: time.RFC3339,
			loc:    locationNewYork,
			times: []crontimes{
				// Move forward 1 hour on 14 March 2021, 02:00 from UTC-5 to UTC-4
				// Daylight savings changes are reversed in the Northern Hemisphere.
				{"2021-03-13T23:00:00-05:00", "2021-03-14T00:00:00-05:00"},
				{"2021-03-14T00:00:00-05:00", "2021-03-14T01:00:00-05:00"},
				{"2021-03-14T01:00:00-05:00", "2021-03-14T03:00:00-04:00"},
				{"2021-03-14T01:59:59-05:00", "2021-03-14T03:00:00-04:00"},
				{"2021-03-14T03:00:00-04:00", "2021-03-14T04:00:00-04:00"},
				// Move back 1 hour on 7 Nov 2021, 02:00 from UTC-4 to UTC-5
				{"2021-11-06T23:00:00-04:00", "2021-11-07T00:00:00-04:00"},
				{"2021-11-07T00:00:00-04:00", "2021-11-07T01:00:00-04:00"},
				{"2021-11-07T01:00:00-04:00", "2021-11-07T01:00:00-05:00"},
				{"2021-11-07T01:59:59-04:00", "2021-11-07T01:00:00-05:00"},
				{"2021-11-07T01:00:00-05:00", "2021-11-07T02:00:00-05:00"},
				{"2021-11-07T02:00:00-05:00", "2021-11-07T03:00:00-05:00"},
				{"2021-11-07T03:00:00-05:00", "2021-11-07T04:00:00-05:00"},
			},
		},
		{
			name:   "Daylight Savings changeover for America/New_York, run every 2 hours",
			expr:   "0 0/2 ? * *",
			layout: time.RFC3339,
			loc:    locationNewYork,
			times: []crontimes{
				// Move forward 1 hour on 14 March 2021, 02:00 from UTC-5 to UTC-4
				{"2021-03-13T23:00:00-05:00", "2021-03-14T00:00:00-05:00"},
				{"2021-03-14T00:00:00-05:00", "2021-03-14T04:00:00-04:00"}, // 2am doesn't exist
				{"2021-03-14T01:00:00-05:00", "2021-03-14T04:00:00-04:00"},
				{"2021-03-14T03:00:00-04:00", "2021-03-14T04:00:00-04:00"},
				{"2021-03-14T04:00:00-04:00", "2021-03-14T06:00:00-04:00"},
				// Move back 1 hour on 7 Nov 2021, 02:00 from UTC-4 to UTC-5
				{"2021-11-06T23:00:00-04:00", "2021-11-07T00:00:00-04:00"},
				{"2021-11-07T00:00:00-04:00", "2021-11-07T02:00:00-05:00"},
				{"2021-11-07T01:00:00-04:00", "2021-11-07T02:00:00-05:00"},
				{"2021-11-07T01:00:00-05:00", "2021-11-07T02:00:00-05:00"},
				{"2021-11-07T02:00:00-05:00", "2021-11-07T04:00:00-05:00"},
				{"2021-11-07T03:00:00-05:00", "2021-11-07T04:00:00-05:00"},
			},
		},
		{
			name:   "Daylight Savings changeover for America/New_York, run every 2 hours on odd hours",
			expr:   "0 1/2 ? * *",
			layout: time.RFC3339,
			loc:    locationNewYork,
			times: []crontimes{
				// Move forward 1 hour on 14 March 2021, 02:00 from UTC-5 to UTC-4
				{"2021-03-13T23:00:00-05:00", "2021-03-14T01:00:00-05:00"},
				{"2021-03-14T00:00:00-05:00", "2021-03-14T01:00:00-05:00"},
				{"2021-03-14T01:00:00-05:00", "2021-03-14T03:00:00-04:00"}, // 2am doesn't exist
				{"2021-03-14T03:00:00-04:00", "2021-03-14T05:00:00-04:00"},
				// Move back 1 hour on 7 Nov 2021, 02:00 from UTC-4 to UTC-5
				{"2021-11-06T23:00:00-04:00", "2021-11-07T01:00:00-04:00"},
				{"2021-11-07T00:00:00-04:00", "2021-11-07T01:00:00-04:00"},
				{"2021-11-07T01:00:00-04:00", "2021-11-07T01:00:00-05:00"},
				{"2021-11-07T01:00:00-05:00", "2021-11-07T03:00:00-05:00"},
				{"2021-11-07T02:00:00-05:00", "2021-11-07T03:00:00-05:00"},
				{"2021-11-07T03:00:00-05:00", "2021-11-07T05:00:00-05:00"},
			},
		},
		{
			// In 2026, DST transition happens on November 1.
			name:   "Daylight Savings changeover for America/New_York in 2026, run on the 1st of each month at 2am",
			expr:   "0 2 1 * *",
			layout: time.RFC3339,
			loc:    locationNewYork,
			times: []crontimes{
				// Move forward 1 hour on March 8 2026, 02:00 from UTC-5 to UTC-4
				{"2026-02-01T02:00:00-05:00", "2026-03-01T02:00:00-05:00"},
				{"2026-03-01T02:00:00-05:00", "2026-04-01T02:00:00-04:00"}, // skipped 1h
				{"2026-04-01T02:00:00-04:00", "2026-05-01T02:00:00-04:00"},
				// Move back 1 hour on 7 Nov 2021, 02:00 from UTC-4 to UTC-5
				{"2026-10-01T02:00:00-04:00", "2026-11-01T02:00:00-05:00"}, // added 1h
				{"2026-11-01T00:00:00-04:00", "2026-11-01T02:00:00-05:00"},
			},
		},
	}
	for _, test := range crontests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			for _, times := range test.times {
				loc := time.UTC
				if test.loc != nil {
					loc = test.loc
				}
				from, err := time.ParseInLocation(time.RFC3339, times.from, loc)
				if err != nil {
					t.Errorf(`Cannot parse %v`, times.from)
					continue
				}

				expr, err := Parse(test.expr)
				if err != nil {
					t.Errorf(`Parse("%s") returned "%s"`, test.expr, err.Error())
					continue
				}
				next := expr.Next(from)
				nextstr := next.Format(test.layout)
				if nextstr != times.next {
					t.Errorf(`("%s").Next("%s") = "%s", got "%s"`, test.expr, times.from, times.next, nextstr)
				}
			}
		})
	}
}

// Issue: https://github.com/gorhill/cronexpr/issues/16
func TestInterval_Interval60Issue(t *testing.T) {
	_, err := Parse("*/60 * * * * *")
	if err == nil {
		t.Errorf("parsing with interval 60 should return err")
	}

	_, err = Parse("*/61 * * * * *")
	if err == nil {
		t.Errorf("parsing with interval 61 should return err")
	}

	_, err = Parse("2/60 * * * * *")
	if err == nil {
		t.Errorf("parsing with interval 60 should return err")
	}

	_, err = Parse("2-20/61 * * * * *")
	if err == nil {
		t.Errorf("parsing with interval 60 should return err")
	}
}

/******************************************************************************/

var benchmarkExpressions = []string{
	"* * * * *",
	"@hourly",
	"@weekly",
	"@yearly",
	"30 3 15W 3/3 *",
	"30 0 0 1-31/5 Oct-Dec * 2000,2006,2008,2013-2015",
	"0 0 0 * Feb-Nov/2 thu#3 2000-2050",
}
var benchmarkExpressionsLen = len(benchmarkExpressions)

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = MustParse(benchmarkExpressions[i%benchmarkExpressionsLen])
	}
}

func BenchmarkNext(b *testing.B) {
	exprs := make([]*Expression, benchmarkExpressionsLen)
	for i := 0; i < benchmarkExpressionsLen; i++ {
		exprs[i] = MustParse(benchmarkExpressions[i])
	}
	from := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expr := exprs[i%benchmarkExpressionsLen]
		next := expr.Next(from)
		next = expr.Next(next)
		next = expr.Next(next)
		next = expr.Next(next)
		next = expr.Next(next)
	}
}
