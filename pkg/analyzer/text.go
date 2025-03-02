package analyzer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
)

func (d *DailyReport) Markdown(title string) string {
	heading := bot.EscapeMarkdown(fmt.Sprintf("%s (%s)", title, d.DayString()))

	numberWidth := 12

	content := bot.EscapeMarkdown(fmt.Sprintf(`
MA 350x2:  %s
MA 111:    %s
Gap:       %s

Average:   %s
Gap:       %s%%
`,
		formatNumber(d.averages.MovingAverage350x2, numberWidth),
		formatNumber(d.averages.MovingAverage111, numberWidth),
		formatNumber(d.PiCycleTopDifference(), numberWidth),
		formatNumber(d.averages.DailyAverage, numberWidth),
		formatNumber(d.PiCycleTopDifferencePercent(), numberWidth),
	))

	return fmt.Sprintf("*%s*\n\n```%s```", heading, content)
}

func formatNumber(f float64, width int) string {
	numDigitsRight := 2

	fString := strconv.FormatFloat(f, 'f', numDigitsRight, 64)
	numDigitsLeft := len(fString) - numDigitsRight - 1
	numCommas := (numDigitsLeft - 1) / 3

	var result strings.Builder

	numSpaces := max(0, width-(len(fString)+numCommas))
	for i := 0; i < numSpaces; i++ {
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
