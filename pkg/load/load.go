package load

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/cloudsftp/botificator/pkg/api"
	"github.com/cloudsftp/botificator/pkg/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"resty.dev/v3"
)

const (
	bistampRootUrl = "https://www.bitstamp.net/api/v2/"

	step = 5 * 60
)

var startTime = time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

func LoadDataIntoDatabase(ctx context.Context, client *resty.Client, pool *pgxpool.Pool) error {
	logrus.Debug("Updating database...")

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection from pool: %w", err)
	}
	defer conn.Release()

	startTimestamp, ok, err := db.GetLatestTimestamp(ctx, conn.Conn())
	if err != nil {
		return fmt.Errorf("failed to get the latest timestamp: %w", err)
	}
	startTimestamp += step
	if !ok {
		startTimestamp = startTime.Unix()
	}

	currentTimestamp := startTimestamp
	for {
		logrus.Tracef("Currently downloading data for timestamp %s", time.Unix(currentTimestamp, 0).Format(time.RFC3339))

		lastTimestamp, done, err := downloadData(ctx, client, conn.Conn(), currentTimestamp)
		if err != nil {
			return err
		}

		if done {
			logrus.Debug("Done updating database")
			return nil
		}

		currentTimestamp = lastTimestamp + step

		// Try not to get rate-limited
		time.Sleep(100 * time.Millisecond)
	}
}

func downloadData(ctx context.Context, client *resty.Client, conn *pgx.Conn, currentTimestamp int64) (int64, bool, error) {
	result, err := client.R().WithContext(ctx).SetQueryParams(map[string]string{
		"step":                   fmt.Sprint(step),
		"limit":                  "1000",
		"exclude_current_candle": "true",
		"start":                  fmt.Sprint(currentTimestamp),
	}).Get(bistampRootUrl + "ohlc/btcusd")

	if err != nil {
		return 0, false, fmt.Errorf("requesting data from bitstamp: %w", err)
	}

	if result.Err != nil {
		return 0, false, fmt.Errorf("result contains error: %w", result.Err)
	}

	var data api.HistoricalDataResponseWrapper
	err = json.NewDecoder(result.Body).Decode(&data)
	if err != nil {
		return 0, false, fmt.Errorf("could not decode body: %w", err)
	}

	if len(data.Inner.Data) == 0 {
		return 0, true, nil
	}

	elements := data.Inner.Data
	_, err = db.InsertDataPoints(ctx, conn, elements)
	if err != nil {
		return 0, false, err
	}

	lastTimestampString := elements[len(elements)-1].Timestamp
	lastTimestamp, err := strconv.ParseInt(lastTimestampString, 10, 64)
	if err != nil {
		return 0, false, err
	}

	return lastTimestamp, false, nil
}
