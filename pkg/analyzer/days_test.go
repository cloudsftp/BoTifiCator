package analyzer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDaysSince(t *testing.T) {
	testCases := []struct {
		name     string
		now      time.Time
		prev     time.Time
		expected int
	}{
		{
			"yesterday",
			time.Now(),
			time.Now().Add(-24 * time.Hour),
			1,
		},
		{
			"a week ago",
			time.Now(),
			time.Now().Add(-24 * time.Hour * 7),
			7,
		},
	}

	for _, testCase := range testCases {
		actual := daysSince(testCase.now, testCase.prev)
		assert.Equal(
			t, testCase.expected, actual,
			"in test case '%s'", testCase.name,
		)
	}
}
