package analyzer

import (
	"context"
	"fmt"

	"github.com/cloudsftp/botificator/pkg/db"
)

type DailyReport struct {
	Data *db.ReportData
}

func NewDailyReport(d *db.ReportData) DailyReport {
	return DailyReport{Data: d}
}

func Analyze(
	ctx context.Context,
	dataProvider *db.DataProvider,
) (*DailyReport, error) {
	data, err := dataProvider.GetReportData(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get report data: %w", err)
	}

	return &DailyReport{data}, nil
}
