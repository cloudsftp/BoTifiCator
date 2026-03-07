package analyzer

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudsftp/botificator/pkg/db"
)

var (
	nextHalvingDate = time.Date(2028, time.April, 1, 0, 0, 0, 0, time.UTC)
)

type DailyReport struct {
	averages *db.MovingAverages
}

func (d *DailyReport) DayString() string {
	return d.averages.Day.Format("2006-01-02")
}

func (d *DailyReport) DaysUntilHalving() int {
	return daysUntil(nextHalvingDate, d.averages.Day)
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

func daysUntil(future time.Time, now time.Time) int {
	duration := future.Truncate(day).Sub(now.Truncate(day))
	return int(duration.Hours() / 24)
}
