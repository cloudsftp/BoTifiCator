package analyzer

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudsftp/botificator/pkg/db"
)

type DailyReport struct {
	Time       time.Time
	Ma111      float64
	Ma350x2    float64
	Difference float64
}

func Analyze(ctx context.Context, dataProvider *db.DataProvider) ([]DailyReport, error) {
	averages, err := dataProvider.GetMovingAverages(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get moving averages: %s", err)
	}
	_ = averages

	return nil, nil
}
