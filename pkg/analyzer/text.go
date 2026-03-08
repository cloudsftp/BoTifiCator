package analyzer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
)

const (
	oneAndAHalfYears = 360 + 180
)

func (d *DailyReport) Markdown(title string) string {
	day := d.averages.Day

	heading := bot.EscapeMarkdown(fmt.Sprintf("%s (%s)", title, d.DateString()))

	daysUntilNextHalving, nextHalving, nextHalvingOk := daysUntilNextHalving(day)
	daysSinceLastHalving, lastHalving := daysSinceLastHalving(day)

	halvingWarning := ""
	switch {
	case !nextHalvingOk:
		halvingWarning = "Next halving is in the past. Please update the halving dates"
	case daysUntilNextHalving < oneAndAHalfYears:
		halvingWarning = "!!! The Halving is Near !!!"
	}
	if halvingWarning != "" {
		halvingWarning = fmt.Sprintf(
			"\n*%s*\n",
			bot.EscapeMarkdown(halvingWarning),
		)
	}

	numberWidth := 12

	nextHalvingLine := ""
	if nextHalvingOk {
		nextHalvingLine = fmt.Sprintf(`
%s until next halving (%s)
`,
			formatDays(daysUntilNextHalving, 12),
			dateString(nextHalving),
		)
	}

	content := bot.EscapeMarkdown(fmt.Sprintf(`
%s
%s since last halving (%s)

%s since last low
%s since last high

Average:   %s
200W MA:   %s
`,
		nextHalvingLine,
		formatDays(daysSinceLastHalving, 12),
		dateString(lastHalving),
		formatDays(daysSince(day, mostRecentLowDate), 12),
		formatDays(daysSince(day, mostRecentHighDate), 12),
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
