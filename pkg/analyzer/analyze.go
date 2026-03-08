package analyzer

import (
	"context"
	"fmt"

	"github.com/cloudsftp/botificator/pkg/db"
)

type DailyReport struct {
	averages *db.MovingAverages
}

func NewDailyReport(a *db.MovingAverages) DailyReport {
	return DailyReport{averages: a}
}

func Analyze(ctx context.Context, dataProvider *db.DataProvider) ([]DailyReport, error) {
	movingAverages, err := dataProvider.GetMovingAverages(ctx, 3)
	if err != nil {
		return nil, fmt.Errorf("could not get moving averages: %w", err)
	}

	var reports []DailyReport
	for _, averages := range movingAverages {
		reports = append(reports, NewDailyReport(&averages))
	}

	return reports, nil
}
