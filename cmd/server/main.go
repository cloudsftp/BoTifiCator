package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"resty.dev/v3"

	"github.com/joho/godotenv"

	"github.com/cloudsftp/botificator/pkg/analyzer"
	"github.com/cloudsftp/botificator/pkg/db"
	"github.com/cloudsftp/botificator/pkg/load"
)

var startTime = time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not load environment: %w\n", err)
		os.Exit(1)
	}

	pool, err := db.SetupDatabase(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not setup database: %w\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	client := resty.New()
	defer client.Close()
	err = load.LoadDataIntoDatabase(ctx, client, pool, startTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not load data into database: %w\n", err)
		os.Exit(1)
	}

	averages, err := db.GetMovingAverages(ctx, pool)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get moving averages: %w\n", err)
		os.Exit(1)
	}

	err = analyzer.Analyze(averages)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not analyze averages: %w\n", err)
		os.Exit(1)
	}
}
