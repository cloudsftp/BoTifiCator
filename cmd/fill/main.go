package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"resty.dev/v3"
)

const (
	bistampRootUrl = "https://www.bitstamp.net/api/v2/"

	step = 5 * 60
)

// var startTime = time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
var startTime = time.Date(2025, 2, 21, 0, 0, 0, 0, time.UTC)

func main() {
	ctx := context.Background()

	conn, err := setupDatabase(ctx)
	defer conn.Close(ctx)

	if err != nil {
		fmt.Fprintf(os.Stderr, "could not setup database: %s\n", err)
		os.Exit(1)
	}

	client := resty.New()
	for {
		startTimestamp, ok, err := getLatestTimestamp(ctx, conn, tableName)
		if err != nil {
			os.Exit(1)
		}
		startTimestamp += step
		if !ok {
			startTimestamp = startTime.Unix()
		}

		fmt.Printf("current timestamp: %s\n", time.Unix(startTimestamp, 0).Format(time.RFC3339))

		done, err := downloadData(ctx, client, conn, startTimestamp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not get data: %s\n", err)
			os.Exit(1)
		}

		if done {
			fmt.Fprint(os.Stderr, "no more new data, exiting")
			os.Exit(0)
		}

		// Try not to get rate-limited
		time.Sleep(100 * time.Millisecond)
	}
}

type HistoricalDataPoint struct {
	Timestamp string `json:"timestamp"`
	Open      string `json:"open"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Close     string `json:"close"`
	Volume    string `json:"volume"`
}

type HistoricalDataResponse struct {
	Pair string                `json:"pair"`
	Data []HistoricalDataPoint `json:"ohlc"`
}

type HistoricalDataResponseWrapper struct {
	Inner HistoricalDataResponse `json:"data"`
}

func downloadData(ctx context.Context, client *resty.Client, conn *pgx.Conn, currentTimestamp int64) (bool, error) {
	result, err := client.R().WithContext(ctx).SetQueryParams(map[string]string{
		"step":                   fmt.Sprint(step),
		"limit":                  "1000",
		"exclude_current_candle": "true",
		"start":                  fmt.Sprint(currentTimestamp),
	}).Get(bistampRootUrl + "ohlc/btcusd")

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to bitstamp: %v\n", err)
		return false, err
	}

	if result.Err != nil {
		fmt.Fprintf(os.Stderr, "Result contains error: %v\n", result.Err)
		return false, err
	}

	var data HistoricalDataResponseWrapper
	err = json.NewDecoder(result.Body).Decode(&data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not decode body: %v\n", err)
		return false, err
	}

	if len(data.Inner.Data) == 0 {
		return true, nil
	}

	for _, point := range data.Inner.Data {
		err = insertDataPoint(ctx, conn, tableName, point)
		if err != nil {
			return false, err
		}
	}

	return false, nil
}
