package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"resty.dev/v3"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"

	"github.com/cloudsftp/botificator/pkg/analyzer"
	"github.com/cloudsftp/botificator/pkg/db"
	"github.com/cloudsftp/botificator/pkg/load"
	"github.com/cloudsftp/botificator/pkg/notificator"
)

var startTime = time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not load environment: %s\n", err)
		os.Exit(1)
	}

	notificator, err := notificator.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create notificator: %s\n", err)
		os.Exit(1)
	}

	err = notificator.SendMessageDeployed(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not send message: %s\n", err)
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

	errors := make(chan error)
	var databaseLock *sync.RWMutex

	c := cron.New()
	c.AddFunc("*/15 * * * *", loadDataCron(ctx, client, pool, databaseLock, errors))

	c.AddFunc("0 5 * * *", sendUpdateCron(ctx, pool, databaseLock, errors))

	for {
		select {
		case err := <-errors:
			fmt.Fprintf(os.Stderr, "runtime error: %s", err)
		}
	}
}

func loadDataCron(
	ctx context.Context,
	client *resty.Client,
	pool *pgxpool.Pool,
	databaseLock *sync.RWMutex,
	errors chan<- error,
) func() {
	return func() {
		databaseLock.Lock()
		defer databaseLock.Unlock()

		err := load.LoadDataIntoDatabase(ctx, client, pool, startTime)
		if err != nil {
			errors <- fmt.Errorf("could not load data into database: %s\n", err)
		}
	}
}

func sendUpdateCron(
	ctx context.Context,
	pool *pgxpool.Pool,
	databaseLock *sync.RWMutex,
	errors chan<- error,
) func() {
	return func() {
		ok := databaseLock.TryRLock()
		if !ok {
			// TODO: wait for an hour, otherwise error out
			return
		}

		averages, err := db.GetMovingAverages(ctx, pool)
		if err != nil {
			errors <- fmt.Errorf("could not get moving averages: %s\n", err)
		}

		err = analyzer.Analyze(averages)
		if err != nil {
			errors <- fmt.Errorf("could not analyze averages: %s\n", err)
		}
	}
}
