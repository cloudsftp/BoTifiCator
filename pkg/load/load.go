package load

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/cloudsftp/botificator/pkg/api"
	"github.com/cloudsftp/botificator/pkg/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"resty.dev/v3"
)

const (
	bistampRootUrl = "https://www.bitstamp.net/api/v2/"

	step = 5 * 60
)

func LoadDataIntoDatabase(ctx context.Context, client *resty.Client, pool *pgxpool.Pool, startTime time.Time) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	startTimestamp, ok, err := db.GetLatestTimestamp(ctx, conn.Conn())
	if err != nil {
		return err
	}
	startTimestamp += step
	if !ok {
		startTimestamp = startTime.Unix()
	}

	currentTimestamp := startTimestamp
	for {
		fmt.Printf("current timestamp: %s\n", time.Unix(currentTimestamp, 0).Format(time.RFC3339))

		lastTimestamp, done, err := downloadData(ctx, client, conn.Conn(), currentTimestamp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not get data: %s\n", err)
			return err
		}

		if done {
			fmt.Println("no more new data, exiting loop")
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
		fmt.Fprintf(os.Stderr, "Unable to connect to bitstamp: %v\n", err)
		return 0, false, err
	}

	if result.Err != nil {
		fmt.Fprintf(os.Stderr, "Result contains error: %v\n", result.Err)
		return 0, false, err
	}

	var data api.HistoricalDataResponseWrapper
	err = json.NewDecoder(result.Body).Decode(&data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not decode body: %v\n", err)
		return 0, false, err
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
