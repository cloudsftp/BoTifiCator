package analyzer

import (
	"fmt"
	"time"
)

func Date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

var (
	day                = time.Hour * 24
	mostRecentLowDate  = Date(2022, time.November, 22)
	mostRecentHighDate = Date(2025, time.October, 6)

	halvingDates = []time.Time{
		Date(2009, time.January, 3),
		Date(2012, time.November, 28),
		Date(2016, time.July, 9),
		Date(2020, time.May, 11),
		Date(2024, time.April, 20),
		Date(2028, time.April, 1), // estimate
	}
)

func dateString(day time.Time) string {
	return day.Format("2006-01-02")
}

func (d *DailyReport) DateString() string {
	return dateString(d.averages.Day)
}

func formatDays(days int, width int) string {
	years := days / 365
	remainingDays := days % 365

	s := ""
	if years > 0 {
		s = fmt.Sprintf("%1d years ", years)
	}
	s += fmt.Sprintf(" %3d days", remainingDays)

	return fmt.Sprintf("%*s", width, s)
}

func daysSince(now time.Time, prev time.Time) int {
	duration := now.Truncate(day).Sub(prev.Truncate(day))
	return int(duration.Hours() / 24)
}

func daysUntil(future time.Time, now time.Time) int {
	duration := future.Truncate(day).Sub(now.Truncate(day))
	return int(duration.Hours() / 24)
}

func daysSinceLastHalving(day time.Time) (int, time.Time) {
	lastHalving := halvingDates[len(halvingDates)-1]
	if day.Before(lastHalving) {
		lastHalving = halvingDates[len(halvingDates)-2]
	}

	return daysSince(day, lastHalving), lastHalving
}

func daysUntilNextHalving(day time.Time) (int, time.Time, bool) {
	nextHalving := halvingDates[len(halvingDates)-1]
	if day.After(nextHalving) {
		return 0, nextHalving, false
	}

	return daysUntil(nextHalving, day), nextHalving, true
}
