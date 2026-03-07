package analyzer

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
)

var (
	day                = time.Hour * 24
	mostRecentLowDate  = time.Date(2022, time.November, 22, 0, 0, 0, 0, time.UTC)
	mostRecentHighDate = time.Date(2025, time.October, 6, 0, 0, 0, 0, time.UTC)
)

func (d *DailyReport) Markdown(title string) string {
	heading := bot.EscapeMarkdown(fmt.Sprintf("%s (%s)", title, d.DayString()))

	daysUntilHalving := daysUntil(nextHalvingDate, d.averages.Day)

	halvingWarning := ""
	if daysUntilHalving < 360 {
		halvingWarning = fmt.Sprintf("\n*%s*\n", bot.EscapeMarkdown(
			"!!! The Halving is Near !!!",
		))

	}

	numberWidth := 12

	content := bot.EscapeMarkdown(fmt.Sprintf(`
%d days until next halving

%d days since last low
%d days since last high

Average:   %s
200W MA:   %s
`,
		daysUntilHalving,
		daysSince(d.averages.Day, mostRecentLowDate),
		daysSince(d.averages.Day, mostRecentHighDate),
		formatNumber(d.averages.DailyAverage, numberWidth),
		formatNumber(d.averages.MovingAverage200W, numberWidth),
	))

	return fmt.Sprintf(
		"*%s*\n%s\n```%s```",
		heading,
		halvingWarning,
		content,
	)
}

func formatNumber(f float64, width int) string {
	numDigitsRight := 2

	fString := strconv.FormatFloat(f, 'f', numDigitsRight, 64)
	numDigitsLeft := len(fString) - numDigitsRight - 1
	numCommas := (numDigitsLeft - 1) / 3

	var result strings.Builder

	numSpaces := max(0, width-(len(fString)+numCommas))
	for range numSpaces {
		result.WriteRune(' ')
	}

	j := 0
	for i := numDigitsLeft; i > 0; i-- {
		if i%3 == 0 && i != numDigitsLeft {
			result.WriteRune(',')
		}
		result.WriteByte(fString[j])
		j += 1
	}

	result.WriteString(fString[j:])

	return result.String()
}

func daysSince(now time.Time, prev time.Time) int {
	duration := now.Truncate(day).Sub(prev.Truncate(day))
	return int(duration.Hours() / 24)
}
