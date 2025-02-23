package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"resty.dev/v3"

	"github.com/joho/godotenv"

	"github.com/cloudsftp/botificator/pkg/db"
	"github.com/cloudsftp/botificator/pkg/load"
)

var startTime = time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not load environment: %s\n", err)
		os.Exit(1)
	}

	pool, err := db.SetupDatabase(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not setup database: %s\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	client := resty.New()
	defer client.Close()
	_ = client
	if false {
		err = load.LoadDataIntoDatabase(ctx, client, pool, startTime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not load data into database: %s\n", err)
			os.Exit(1)
		}
	}

	averages, err := db.GetMovingAverages(ctx, pool)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get moving averages: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(averages)
}
