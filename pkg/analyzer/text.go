package analyzer

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/go-telegram/bot"
)

const (
	oneAndAHalfYears = 360 + 180
	numberWidth      = 12
)

//go:embed message.tmpl
var messageTemplateString string

var messageTemplate *template.Template

func init() {
	messageTemplate = template.Must(template.New("message").Funcs(template.FuncMap{
		"daysSince":  daysSince,
		"formatDays": formatDays,
		"dateString": dateString,
		"formatNumber": func(number float64) string {
			return formatNumber(number, numberWidth)
		},
		"escape": bot.EscapeMarkdown,
	}).Parse(messageTemplateString))
}

func (d *DailyReport) Markdown(title string) (string, error) {
	day := d.Data.Day

	daysUntilNextHalving, nextHalving, nextHalvingOk := daysUntilNextHalving(day)
	daysSinceLastHalving, lastHalving := daysSinceLastHalving(day)

	var result strings.Builder
	err := messageTemplate.Execute(&result, map[string]any{
		"d":                    d,
		"day":                  day,
		"daysSinceLastHalving": daysSinceLastHalving,
		"lastHalving":          lastHalving,
		"daysUntilNextHalving": daysUntilNextHalving,
		"nextHalving":          nextHalving,
		"nextHalvingOk":        nextHalvingOk,
		"oneAndAHalfYears":     oneAndAHalfYears,
		"mostRecentLowDate":    mostRecentLowDate,
		"mostRecentHighDate":   mostRecentHighDate,
	})
	if err != nil {
		return "", fmt.Errorf("could not execute the template: %w", err)
	}

	return result.String(), nil
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
