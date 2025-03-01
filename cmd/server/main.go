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

	notificator, pool, client, err := setup(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error in setup: %s", err)
	}
	_ = notificator

	defer pool.Close()
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

func setup(ctx context.Context) (*notificator.Notificator, *pgxpool.Pool, *resty.Client, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not load environment: %w\n", err)
	}

	notificator, err := notificator.New()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not create notificator: %w\n", err)
	}

	pool, err := db.SetupDatabase(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not setup database: %w\n", err)
	}

	client := resty.New()

	err = notificator.SendMessageDeployed(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not send message: %w\n", err)
	}

	return notificator, pool, client, nil
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
