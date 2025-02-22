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

	latestTimestamp, ok, err := getLatestTimestamp(ctx, conn, tableName)
	if err != nil {
		os.Exit(1)
	}
	if !ok {
		latestTimestamp = startTime.Unix()
	}

	client := resty.New()
	err = getData(ctx, client, conn, latestTimestamp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get data: %s\n", err)
		os.Exit(1)
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

func getData(ctx context.Context, client *resty.Client, conn *pgx.Conn, latestTimestamp int64) error {
	result, err := client.R().WithContext(ctx).SetQueryParams(map[string]string{
		"step":                   "300",
		"limit":                  "100",
		"exclude_current_candle": "true",
		"start":                  fmt.Sprint(latestTimestamp),
	}).Get(bistampRootUrl + "ohlc/btcusd")

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to bitstamp: %v\n", err)
		return err
	}

	if result.Err != nil {
		fmt.Fprintf(os.Stderr, "Result contains error: %v\n", result.Err)
		return err
	}

	var data HistoricalDataResponseWrapper
	err = json.NewDecoder(result.Body).Decode(&data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not decode body: %v\n", err)
		return err
	}

	fmt.Println()

	for _, point := range data.Inner.Data {
		insertDataPoint(ctx, conn, tableName, point)
	}

	/*
		_, err =
			conn.Exec(ctx, "INSERT INTO btc (time, open, high, low, close, volume) VALUES (to_timestamp($1), $2, $3, $4, $5, $6)",
				recordTimestamp, record[1], record[2], record[3], record[4], record[5])

		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to insert record: %v\n", err)
		}
	*/

	return nil
}
