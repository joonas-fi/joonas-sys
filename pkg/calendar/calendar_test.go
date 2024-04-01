package calendar

import (
	"testing"
	"time"

	"github.com/function61/gokit/testing/assert"
)

func TestWeekdaysSinceMonday(t *testing.T) {
	for _, tc := range []struct {
		input  time.Weekday
		expect int
	}{
		{time.Monday, 0},
		{time.Tuesday, 1},
		{time.Wednesday, 2},
		{time.Thursday, 3},
		{time.Friday, 4},
		{time.Saturday, 5},
		{time.Sunday, 6},
	} {
		assert.Assert(t, weekdaysSince(tc.input, time.Monday) == tc.expect)
	}
}

func TestWeekdaysSinceSunday(t *testing.T) {
	for _, tc := range []struct {
		input  time.Weekday
		expect int
	}{
		{time.Monday, 1},
		{time.Tuesday, 2},
		{time.Wednesday, 3},
		{time.Thursday, 4},
		{time.Friday, 5},
		{time.Saturday, 6},
		{time.Sunday, 0},
	} {
		assert.Assert(t, weekdaysSince(tc.input, time.Sunday) == tc.expect)
	}
}

func TestWeekdaysSinceSaturday(t *testing.T) {
	for _, tc := range []struct {
		input  time.Weekday
		expect int
	}{
		{time.Monday, 2},
		{time.Tuesday, 3},
		{time.Wednesday, 4},
		{time.Thursday, 5},
		{time.Friday, 6},
		{time.Saturday, 0},
		{time.Sunday, 1},
	} {
		assert.Assert(t, weekdaysSince(tc.input, time.Saturday) == tc.expect)
	}
}

func TestWeekdays(t *testing.T) {
	// https://en.wikipedia.org/wiki/Week#Other_week_numbering_systems

	// civilized people
	assert.EqualJSON(t, makeWeekdays(localeDefinition{firstDayOfWeek: time.Monday}), `[
  "Mo",
  "Tu",
  "We",
  "Th",
  "Fr",
  "Sa",
  "Su"
]`)

	// murica & other weirdos
	assert.EqualJSON(t, makeWeekdays(localeDefinition{firstDayOfWeek: time.Sunday}), `[
  "Su",
  "Mo",
  "Tu",
  "We",
  "Th",
  "Fr",
  "Sa"
]`)

	// middle east
	assert.EqualJSON(t, makeWeekdays(localeDefinition{firstDayOfWeek: time.Saturday}), `[
  "Sa",
  "Su",
  "Mo",
  "Tu",
  "We",
  "Th",
  "Fr"
]`)
}
