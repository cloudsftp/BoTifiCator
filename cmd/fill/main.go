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
	_ = conn

	if err != nil {
		fmt.Fprintf(os.Stderr, "could not setup database: %s\n", err)
		os.Exit(1)
	}

	os.Exit(0)

	client := resty.New()
	err = getData(ctx, client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get data: %s\n", err)
		os.Exit(1)
	}
}

func setupDatabase(ctx context.Context) (*pgx.Conn, error) {
	connStr := "postgres://postgres:mysecretpassword@localhost:5432/postgres"

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	err = createTable(ctx, conn, "btc_5min")
	if err != nil {
		return nil, err
	}

	/*
		_, err =
			conn.Exec(ctx, "INSERT INTO btc (time, open, high, low, close, volume) VALUES (to_timestamp($1), $2, $3, $4, $5, $6)",
				recordTimestamp, record[1], record[2], record[3], record[4], record[5])

		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to insert record: %v\n", err)
		}
	*/

	return conn, nil
}

func createTable(ctx context.Context, conn *pgx.Conn, tableName string) error {
	_, err := conn.Exec(ctx, fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			time TIMESTAMPTZ NOT NULL,
			open DECIMAL NOT NULL,
			high DECIMAL NOT NULL,
			low DECIMAL NOT NULL,
			close DECIMAL NOT NULL,
			volume DECIMAL NOT NULL
		);
	`, tableName))
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, fmt.Sprintf(`
		SELECT create_hypertable('%s', 'time', if_not_exists => true);
	`, tableName))
	if err != nil {
		return err
	}

	return nil
}

type HistoricalDataPoint struct {
	Timestamp int64
	Open      int64
	High      int64
	Low       int64
	Close     int64
	Volume    float64
}

type HistoricalDataPointResponse struct {
	Timestamp string `json:"timestamp"`
	Open      string `json:"open"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Close     string `json:"close"`
	Volume    string `json:"volume"`
}

type HistoricalDataResponse struct {
	Pair string                        `json:"pair"`
	Data []HistoricalDataPointResponse `json:"ohlc"`
}

type HistoricalDataResponseWrapper struct {
	Inner HistoricalDataResponse `json:"data"`
}

func getData(ctx context.Context, client *resty.Client) error {
	result, err := client.R().WithContext(ctx).SetQueryParams(map[string]string{
		"step":                   "300",
		"limit":                  "100",
		"exclude_current_candle": "true",
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

	fmt.Println(data)

	return nil
}
