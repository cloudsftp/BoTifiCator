package main

import (
	"fmt"
	"time"

	"github.com/cloudsftp/botificator/pkg/analyzer"
	"github.com/cloudsftp/botificator/pkg/db"
)

func showMessage(date time.Time) {
	averages := &db.MovingAverages{
		Day:               date,
		DailyAverage:      85432.50,
		MovingAverage200W: 48250.75,
	}

	report := analyzer.NewDailyReport(averages)
	fmt.Println(report.Markdown("Yesterday"))
}

func main() {
	showMessage(analyzer.Date(2026, 4, 1))
	fmt.Println("\n---")
	showMessage(analyzer.Date(2027, 4, 1))
	fmt.Println("\n---")
	showMessage(analyzer.Date(2028, 4, 2))
}
