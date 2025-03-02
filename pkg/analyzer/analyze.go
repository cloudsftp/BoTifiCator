package analyzer

import (
	"context"
	"fmt"

	"github.com/cloudsftp/botificator/pkg/db"
)

type DailyReport struct {
	averages *db.MovingAverages
}

func (d *DailyReport) DayString() string {
	return d.averages.Day.Format("2006-01-02")
}

func (d *DailyReport) PiCycleTopDifference() float64 {
	return d.averages.MovingAverage350x2 - d.averages.MovingAverage111
}

func (d *DailyReport) PiCycleTopDifferencePercent() float64 {
	return d.PiCycleTopDifference() / d.averages.DailyAverage * 100.
}

func Analyze(ctx context.Context, dataProvider *db.DataProvider) ([]DailyReport, error) {
	movingAverages, err := dataProvider.GetMovingAverages(ctx, 3)
	if err != nil {
		return nil, fmt.Errorf("could not get moving averages: %w", err)
	}

	var reports []DailyReport
	for _, averages := range movingAverages {
		reports = append(reports, DailyReport{&averages})
	}

	return reports, nil
}
