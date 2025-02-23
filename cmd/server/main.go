package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"resty.dev/v3"

	"github.com/cloudsftp/botificator/pkg/db"
	"github.com/cloudsftp/botificator/pkg/load"
)

var startTime = time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

func main() {
	ctx := context.Background()

	pool, err := db.SetupDatabase(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not setup database: %s\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	client := resty.New()
	defer client.Close()
	err = load.LoadDataIntoDatabase(ctx, client, pool, startTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not load data into database: %s\n", err)
		os.Exit(1)
	}
}
