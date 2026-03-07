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
	showMessage(time.Date(2027, 3, 6, 0, 0, 0, 0, time.UTC))
	fmt.Println("\n---")
	showMessage(time.Date(2028, 3, 6, 0, 0, 0, 0, time.UTC))
}
