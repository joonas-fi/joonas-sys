package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/apcera/termtables"
	"github.com/function61/gokit/os/osutil"
	"github.com/noamt/go-cldr/supplemental"
	"github.com/spf13/cobra"
	"golang.org/x/text/language"
)

// things that vary based on locale
type localeDefinition struct {
	firstDayOfWeek time.Weekday
}

/* Creates a calendar like this:

      February 2022
   Su Mo Tu We Th Fr Sa
          1  2  3  4  5
    6  7  8  9 10 11 12
   13 14 15 16 17 18 19
   20 21 22 23 24 25 26
   27 28
*/
func calendarEntrypoint() *cobra.Command {
	interactive := false

	cmd := &cobra.Command{
		Use:   "cal [month-number | year]",
		Short: "Displays a calendar",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(func() error {
				locale, err := getLocaleForTime()
				if err != nil {
					return err
				}

				switch {
				case len(args) == 1 && len(args[0]) == 4: // year
					year, err := strconv.Atoi(args[0])
					if err != nil {
						return err
					}

					return calendarPrintYear(year, *locale, os.Stdout)
				case len(args) == 1: // custom month (probably different than the running one)
					now, err := time.Parse("2006-01", args[0])
					if err != nil {
						return err
					}

					return calendarPrint(now, "January 2006", *locale, false, os.Stdout)
				default:
					now := time.Now()

					if interactive {
						return calendarNavigateInteractive(now, func(ts time.Time) error { // called many times
							// only do when for the *now* month
							highlightCurrentDay := ts.Equal(now)

							return calendarPrint(ts, "January 2006", *locale, highlightCurrentDay, os.Stdout)
						})
					}

					return calendarPrint(now, "January 2006", *locale, true, os.Stdout)
				}
			}())
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", interactive, "Navigate months with keyboard")

	return cmd
}

func calendarPrint(
	today time.Time,
	monthTitleFormat string,
	locale localeDefinition,
	highlightCurrentDay bool,
	output io.Writer,
) error {
	// which calendar month to print
	month := firstOfMonth(today).Month()

	cal := termtables.CreateTable()
	cal.Style = tableStyleWithoutBorder()
	cal.AddTitle(today.Format(monthTitleFormat))
	wkd := makeWeekdays(locale)
	cal.AddHeaders(wkd[0], wkd[1], wkd[2], wkd[3], wkd[4], wkd[5], wkd[6]) // not using wkd... b/c API is ...interface{}

	// represents:
	//   mo tu we th fr sa su
	weekdayCells := make([]string, 7)

	addWeekRow := func() {
		cal.AddRow(weekdayCells[0], weekdayCells[1], weekdayCells[2], weekdayCells[3], weekdayCells[4], weekdayCells[5], weekdayCells[6])
	}

	currentDate := firstOfMonth(today) // start from day 1 of the month

	// need padding of two empty cells if wednesday is first day:
	// mo tu we th fr sa su
	cellIdx := weekdaysSince(currentDate.Weekday(), locale.firstDayOfWeek)

	for currentDate.Month() == month {
		weekdayCells[cellIdx] = fmt.Sprintf("%2d", currentDate.Day())

		if cellIdx == (7 - 1) { // move to next week (= new row)
			addWeekRow()
			weekdayCells = make([]string, 7)
			cellIdx = 0
		} else {
			cellIdx++
		}

		currentDate = currentDate.AddDate(0, 0, 1) // day++
	}

	if cellIdx > 0 { // got unprinted cells in row?
		addWeekRow()
	}

	calRendered := func() string {
		if highlightCurrentDay {
			return highlightCurrentDayCell(cal.Render(), today)
		} else {
			return cal.Render()
		}
	}()

	_, err := fmt.Fprint(output, calRendered)
	return err
}

// prints each month of the year
func calendarPrintYear(year int, locale localeDefinition, output io.Writer) error {
	now := time.Now() // for hightlighting current day

	cal := termtables.CreateTable()
	cal.Style = tableStyleWithoutBorder()
	cal.AddTitle(time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"))

	renderMonth := func(month time.Month) string {
		month2, highlightCurrentDay := func() (time.Time, bool) {
			if month == now.Month() {
				return now, true // returning exact day so day hightlight works
			} else {
				return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC), false
			}
		}()

		output := bytes.Buffer{}
		if err := calendarPrint(
			month2,
			"January",
			locale,
			highlightCurrentDay,
			&output,
		); err != nil {
			panic(err)
		}
		return output.String()
	}

	threeMonthsRow := func(mon1, mon2, mon3 time.Month) {
		// termtables doesn't support multi-line cells so we've to break the multiline cells into
		// lines ourselves and insert them separately.
		mon1Lines := strings.Split(renderMonth(mon1), "\n")
		mon2Lines := strings.Split(renderMonth(mon2), "\n")
		mon3Lines := strings.Split(renderMonth(mon3), "\n")

		maxLines := int(math.Max(math.Max(float64(len(mon1Lines)), float64(len(mon2Lines))), float64(len(mon3Lines))))

		for i := 0; i < maxLines; i++ {
			cal.AddRow(sliceItemMaybe(mon1Lines, i), sliceItemMaybe(mon2Lines, i), sliceItemMaybe(mon3Lines, i))
		}
	}

	threeMonthsRow(1, 2, 3)    // Q1
	threeMonthsRow(4, 5, 6)    // Q2
	threeMonthsRow(7, 8, 9)    // Q3
	threeMonthsRow(10, 11, 12) // Q4

	_, err := fmt.Fprint(output, cal.Render())
	return err
}

// weekdaysSince(time.Tuesday, time.Monday) => 1
// weekdaysSince(time.Wednesday, time.Monday) => 2
// weekdaysSince(time.Monday, time.Monday) => 0
func weekdaysSince(weekday time.Weekday, since time.Weekday) int {
	n := int(weekday - since)
	if n < 0 {
		return 7 + n
	} else {
		return n
	}
}

// fn("2022-02-26") => "2022-02-01"
func firstOfMonth(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
}

// an ugly hack
func highlightCurrentDayCell(tableOutput string, currentDate time.Time) string {
	dayStr := fmt.Sprintf("%d", currentDate.Day())

	// ugly hack because we know that the number to search is in a cell (that is always surrounded by spaces)
	search := fmt.Sprintf(" %s ", dayStr)

	// reverse
	// https://stackoverflow.com/a/42449998
	// TODO: use library for producing the escape sequences?
	replace := " \033[7m" + dayStr + "\033[0m "

	return strings.Replace(tableOutput, search, replace, 1)
}

// sunday + 1 = monday
func addWeekdays(weekday time.Weekday, n int) time.Weekday {
	return time.Weekday((int(weekday) + n) % 7)
}

// => ["Mo", "Tu", ...]
func makeWeekdays(locale localeDefinition) []string {
	dayOfWeekAbbreviation := func(n int) string { // helper
		// Monday + 2 => "Wednesday"
		weekday := addWeekdays(locale.firstDayOfWeek, n).String()

		// "Monday" => "Mo"
		return weekday[0:2]
	}

	return []string{ // actual ordering depends on *firstDayOfWeek*
		dayOfWeekAbbreviation(0), // Mo
		dayOfWeekAbbreviation(1), // Tu
		dayOfWeekAbbreviation(2), // We
		dayOfWeekAbbreviation(3), // Th
		dayOfWeekAbbreviation(4), // Fr
		dayOfWeekAbbreviation(5), // Sa
		dayOfWeekAbbreviation(6), // Su
	}
}

func getLocaleForTime() (*localeDefinition, error) {
	// https://man.archlinux.org/man/locale.7 for variable names
	localeTime := os.Getenv("LC_TIME")
	switch localeTime {
	case "", "C.UTF-8": // "C" is for computer usage, i.e. parsable by other programs. https://stackoverflow.com/a/55693338
		// prefering GB due to sensible week number start :D
		// TODO: does CLDR define some sensible defaults?
		localeTime = "en_GB.UTF-8"
	}

	// "fi_FI.UTF-8" => "fi_FI"
	//
	// often these are like "fi_FI.UTF-8", i.e. <language>_<region>.<charEncoding>,
	// and Go's language.Parse() doesn't survive parsing charEncoding
	localeTimeWithoutCharEncoding := strings.TrimSuffix(localeTime, ".UTF-8")

	tag, err := language.Parse(localeTimeWithoutCharEncoding)
	if err != nil {
		return nil, err
	}

	region, _ := tag.Region()

	return &localeDefinition{
		firstDayOfWeek: supplemental.FirstDay.ByRegion(region),
	}, nil
}

func tableStyleWithoutBorder() *termtables.TableStyle {
	return &termtables.TableStyle{
		SkipBorder:   true,
		BorderX:      "",
		BorderY:      "",
		BorderI:      "",
		PaddingLeft:  1,
		PaddingRight: 1,
		Width:        80,
		Alignment:    termtables.AlignLeft,
	}
}

// gets slice item, only if it exists
func sliceItemMaybe(arr []string, i int) string {
	if len(arr) > i {
		return arr[i]
	} else {
		return ""
	}
}

func calendarNavigateInteractive(now time.Time, render func(time.Time) error) error {
	monthOffset := 0 // mutated with keyboard input

	keyReader := bufio.NewReader(os.Stdin)

	for {
		nowPlusMonthOffset := now.AddDate(0, monthOffset, 0)

		if err := render(nowPlusMonthOffset); err != nil {
			return err
		}

		// FIXME: really janky keyboard input implementation.
		//        improve: https://stackoverflow.com/questions/40159137/golang-reading-from-stdin-how-to-detect-special-keys-enter-backspace-etc

		key, err := keyReader.ReadString('\n')
		switch err {
		case io.EOF: // ctrl + d
			return nil
		case nil:
			// no-op
		default:
			return err
		}

		switch fmt.Sprintf("%x", key) {
		case "1b5b440a": // left
			monthOffset--
		case "1b5b430a": // right
			monthOffset++
		default:
			log.Printf("unknown sequence: %x", key)
		}
	}
}
