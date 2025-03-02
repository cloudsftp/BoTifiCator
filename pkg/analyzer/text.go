package analyzer

import (
	"fmt"
	"strconv"

	"github.com/go-telegram/bot"
)

func (d *DailyReport) Markdown(title string) string {
	heading := bot.EscapeMarkdown(fmt.Sprintf("%s (%s)", title, d.DayString()))

	content := bot.EscapeMarkdown(fmt.Sprintf(`
  MA 350x2:  %s
  MA 111:    %s
  Gap:       %s

  Average:   %s
  Gap:       %s%%
`,
		formatNumber(d.averages.MovingAverage350x2),
		formatNumber(d.averages.MovingAverage111),
		formatNumber(d.PiCycleTopDifference()),
		formatNumber(d.averages.DailyAverage),
		formatNumber(d.PiCycleTopDifferencePercent()),
	))

	return fmt.Sprintf("*%s*\n\n```%s```", heading, content)
}

func formatNumber(f float64) string {
	placesAfterDot := 2

	s := strconv.FormatFloat(f, 'f', placesAfterDot, 64)
	b := []byte(s)
	length := len(b)

	if length <= 3+1+placesAfterDot {
		return s
	}

	result := b[length-(placesAfterDot+1) : length]

	j := 0
	for i := length - (1 + placesAfterDot + 1); i >= 0; i-- {
		if j%3 == 0 && j != 0 {
			result = append([]byte{','}, result...)
		}
		result = append([]byte{b[i]}, result...)
		j++
	}
	return string(result)
}
