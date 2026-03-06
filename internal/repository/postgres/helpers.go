package postgres

import (
	"fmt"
	"strings"
	"time"
)

func isDuplicateKeyError(err error) bool {
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "23505")
}

// isCheckConstraintError checks if error is from a CHECK constraint violation
func isCheckConstraintError(err error) bool {
	return strings.Contains(err.Error(), "check constraint") ||
		strings.Contains(err.Error(), "23514")
}

// isConstraintViolationError checks for any constraint violation (check, numeric, etc.)
func isConstraintViolationError(err error) bool {
	return isCheckConstraintError(err) ||
		strings.Contains(err.Error(), "value too") ||
		strings.Contains(err.Error(), "numeric field overflow") ||
		strings.Contains(err.Error(), "23000")
}

// currentDayOfWeek returns the current day of week as int (0=Mon, 6=Sun),
// matching the WorkingHours.DayOfWeek convention.
func currentDayOfWeek() int {
	d := time.Now().Weekday() // Sunday=0, Monday=1, ...
	if d == time.Sunday {
		return 6
	}
	return int(d) - 1
}

// currentTimeHHMM returns the current time in "HH:MM" format.
func currentTimeHHMM() string {
	now := time.Now()
	return fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())
}
