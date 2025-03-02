package analyzer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFormatNumber(t *testing.T) {
	testCases := []struct {
		name     string
		number   float64
		width    int
		expected string
	}{
		{
			"one",
			1., 0,
			"1.00",
		},
		{
			"a thousand",
			1_000.,
			0,
			"1,000.00",
		},
		{
			"a millie",
			1_000_000.,
			0,
			"1,000,000.00",
		},
		{
			"a millie with some change",
			1_234_567.89,
			0,
			"1,234,567.89",
		},
		{
			"ignore decimals after two places",
			1_234_567.891234,
			0,
			"1,234,567.89",
		},
		{
			"padding",
			1_234_567.891234,
			14,
			"  1,234,567.89",
		},
		{
			"padding for numbers with 3*n digits left",
			154_092.69,
			12,
			"  154,092.69",
		},
		{
			"padding for numbers with 3*n+2 digits left",
			54_092.69,
			12,
			"   54,092.69",
		},
	}

	for _, testCase := range testCases {
		actual := formatNumber(testCase.number, testCase.width)
		assert.Equal(
			t, testCase.expected, actual,
			"in test case '%s'", testCase.name,
		)
	}
}
