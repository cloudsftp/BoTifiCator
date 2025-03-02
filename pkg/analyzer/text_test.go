package analyzer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFormatNumber(t *testing.T) {
	test_cases := []struct {
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

	for _, test_case := range test_cases {
		actual := formatNumber(test_case.number, test_case.width)
		assert.Equal(
			t, test_case.expected, actual,
			"in test case '%s'", test_case.name,
		)
	}
}
