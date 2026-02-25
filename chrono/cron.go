package chrono

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// predefinedSchedules maps cron macros to their 5-field equivalents.
var predefinedSchedules = map[string]string{
	"@yearly":   "0 0 1 1 *",
	"@annually": "0 0 1 1 *",
	"@monthly":  "0 0 1 * *",
	"@weekly":   "0 0 * * 0",
	"@daily":    "0 0 * * *",
	"@midnight": "0 0 * * *",
	"@hourly":   "0 * * * *",
}

// CronSchedule represents a cron-expression-based schedule.
// It supports standard 5-field cron expressions:
//
//	┌───────────── minute (0 - 59)
//	│ ┌───────────── hour (0 - 23)
//	│ │ ┌───────────── day of month (1 - 31)
//	│ │ │ ┌───────────── month (1 - 12)
//	│ │ │ │ ┌───────────── day of week (0 - 6, 0 = Sunday)
//	│ │ │ │ │
//	* * * * *
//
// Field syntax:
//   - * : all values
//   - */n : every nth value
//   - n : specific value
//   - n-m : range from n to m (inclusive)
//   - n-m/s : range with step
//   - n,m,o : comma-separated list
//
// Predefined macros: @yearly, @annually, @monthly, @weekly, @daily, @midnight, @hourly
type CronSchedule struct {
	minutes     []int
	hours       []int
	daysOfMonth []int
	months      []int
	daysOfWeek  []int
	expr        string
}

// NewCronSchedule creates a new CronSchedule from a cron expression string.
// Returns ErrInvalidCronExpr if the expression is malformed.
func NewCronSchedule(expr string) (*CronSchedule, error) {
	expr = strings.TrimSpace(expr)

	// Check for predefined macros
	if replacement, ok := predefinedSchedules[strings.ToLower(expr)]; ok {
		expr = replacement
	}

	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return nil, fmt.Errorf("%w: expected 5 fields, got %d", ErrInvalidCronExpr, len(fields))
	}

	cs := &CronSchedule{expr: expr}
	var err error

	cs.minutes, err = parseCronField(fields[0], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("%w: minute field: %v", ErrInvalidCronExpr, err)
	}

	cs.hours, err = parseCronField(fields[1], 0, 23)
	if err != nil {
		return nil, fmt.Errorf("%w: hour field: %v", ErrInvalidCronExpr, err)
	}

	cs.daysOfMonth, err = parseCronField(fields[2], 1, 31)
	if err != nil {
		return nil, fmt.Errorf("%w: day-of-month field: %v", ErrInvalidCronExpr, err)
	}

	cs.months, err = parseCronField(fields[3], 1, 12)
	if err != nil {
		return nil, fmt.Errorf("%w: month field: %v", ErrInvalidCronExpr, err)
	}

	cs.daysOfWeek, err = parseCronField(fields[4], 0, 6)
	if err != nil {
		return nil, fmt.Errorf("%w: day-of-week field: %v", ErrInvalidCronExpr, err)
	}

	return cs, nil
}

// Next returns the next activation time after the given time.
// It searches up to 4 years ahead to account for leap year edge cases.
// Returns the zero time if no next activation is found within the search window.
func (cs *CronSchedule) Next(from time.Time) time.Time {
	// Start from the next second, truncated to the minute
	t := from.Add(time.Minute - time.Duration(from.Second())*time.Second -
		time.Duration(from.Nanosecond())).Truncate(time.Minute)

	// Search up to 4 years ahead
	limit := t.Add(4 * 365 * 24 * time.Hour)

	for t.Before(limit) {
		// Check month
		if !intSliceContains(cs.months, int(t.Month())) {
			t = time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
			continue
		}

		// Check day of month
		if !intSliceContains(cs.daysOfMonth, t.Day()) {
			t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, t.Location())
			continue
		}

		// Check day of week
		if !intSliceContains(cs.daysOfWeek, int(t.Weekday())) {
			t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, t.Location())
			continue
		}

		// Check hour
		if !intSliceContains(cs.hours, t.Hour()) {
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+1, 0, 0, 0, t.Location())
			continue
		}

		// Check minute
		if !intSliceContains(cs.minutes, t.Minute()) {
			t = t.Add(time.Minute)
			continue
		}

		return t
	}

	return time.Time{}
}

// String returns the original cron expression.
func (cs *CronSchedule) String() string {
	return cs.expr
}

// parseCronField parses a single cron field and returns the matching values.
func parseCronField(field string, min, max int) ([]int, error) {
	if field == "*" {
		return makeRange(min, max, 1), nil
	}

	var values []int

	// Handle comma-separated list
	parts := strings.Split(field, ",")
	for _, part := range parts {
		partValues, err := parseCronPart(part, min, max)
		if err != nil {
			return nil, err
		}
		values = append(values, partValues...)
	}

	// Remove duplicates and sort
	values = uniqueInts(values)
	sort.Ints(values)

	if len(values) == 0 {
		return nil, fmt.Errorf("no values resolved for field: %s", field)
	}

	return values, nil
}

// parseCronPart parses a single part of a cron field (handles ranges and steps).
func parseCronPart(part string, min, max int) ([]int, error) {
	// Check for step value
	stepParts := strings.SplitN(part, "/", 2)

	step := 1
	if len(stepParts) == 2 {
		var err error
		step, err = strconv.Atoi(stepParts[1])
		if err != nil || step <= 0 {
			return nil, fmt.Errorf("invalid step value: %s", stepParts[1])
		}
	}

	rangeStr := stepParts[0]

	// Wildcard with step
	if rangeStr == "*" {
		return makeRange(min, max, step), nil
	}

	// Check for range
	rangeParts := strings.SplitN(rangeStr, "-", 2)
	if len(rangeParts) == 2 {
		rangeMin, err := strconv.Atoi(rangeParts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid range start: %s", rangeParts[0])
		}
		rangeMax, err := strconv.Atoi(rangeParts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid range end: %s", rangeParts[1])
		}
		if rangeMin < min || rangeMax > max || rangeMin > rangeMax {
			return nil, fmt.Errorf("range %d-%d out of bounds [%d, %d]", rangeMin, rangeMax, min, max)
		}
		return makeRange(rangeMin, rangeMax, step), nil
	}

	// Single value
	val, err := strconv.Atoi(rangeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid value: %s", rangeStr)
	}
	if val < min || val > max {
		return nil, fmt.Errorf("value %d out of bounds [%d, %d]", val, min, max)
	}

	return []int{val}, nil
}

// makeRange creates a slice of integers from start to end (inclusive) with the given step.
func makeRange(start, end, step int) []int {
	var result []int
	for i := start; i <= end; i += step {
		result = append(result, i)
	}
	return result
}

// intSliceContains checks if a sorted slice contains a value.
func intSliceContains(slice []int, val int) bool {
	idx := sort.SearchInts(slice, val)
	return idx < len(slice) && slice[idx] == val
}

// uniqueInts removes duplicate values from a slice.
func uniqueInts(slice []int) []int {
	seen := make(map[int]bool, len(slice))
	result := make([]int, 0, len(slice))
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}
