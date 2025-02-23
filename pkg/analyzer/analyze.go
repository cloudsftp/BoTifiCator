package analyzer

import (
	"time"
)

type MovingAverages struct {
	Time    time.Time
	Ma100   float64
	Ma350x2 float64
}

func Analyze(averages []MovingAverages) error {
	// TODO

	return nil
}
