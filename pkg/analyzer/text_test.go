package analyzer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFormatNumber(t *testing.T) {
	test_cases := []struct {
		name     string
		number   float64
		expected string
	}{
		{
			"one",
			1.,
			"1.00",
		},
		{
			"a thousand",
			1_000.,
			"1,000.00",
		},
		{
			"a millie",
			1_000_000.,
			"1,000,000.00",
		},
		{
			"a millie with some change",
			1_234_567.89,
			"1,234,567.89",
		},
		{
			"ignore decimals after two places",
			1_234_567.891234,
			"1,234,567.89",
		},
	}

	for _, test_case := range test_cases {
		actual := formatNumber(test_case.number)
		assert.Equal(
			t, test_case.expected, actual,
			"in test case '%s'", test_case.name,
		)
	}
}
