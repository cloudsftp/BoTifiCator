package main

import (
	"fmt"
	"time"

	"github.com/cloudsftp/botificator/pkg/analyzer"
	"github.com/cloudsftp/botificator/pkg/db"
)

func lastMonday(t time.Time) time.Time {
	daysBack := int(t.Weekday() - time.Monday)
	if daysBack < 0 {
		daysBack += 7
	}
	if daysBack == 0 {
		daysBack = 7
	}
	return t.AddDate(0, 0, -daysBack)
}

func showMessage(day time.Time) {
	week := lastMonday(day)

	data := &db.ReportData{
		Day: day,
		Daily: []db.ReportDataDaily{
			{
				Day:     day,
				Average: 50_000.,
			},
			{
				Day:     day.AddDate(0, 0, -1),
				Average: 49_000.,
			},
		},
		Weekly: []db.ReportDataWeekly{
			{
				Week:             week,
				Average:          60_000.,
				MovingAverage200: 30_000.,
				MovingAverage100: 40_000.,
			},
			{
				Week:             week.AddDate(0, 0, -7),
				Average:          59_000.,
				MovingAverage200: 29_000.,
				MovingAverage100: 39_000.,
			},
		},
	}

	report := analyzer.NewDailyReport(data)
	fmt.Println(report.Markdown("Yesterday"))
}

func main() {
	showMessage(analyzer.Date(2026, 4, 1))
	fmt.Println("\n---")
	showMessage(analyzer.Date(2027, 4, 1))
	fmt.Println("\n---")
	showMessage(analyzer.Date(2028, 4, 2))
}
