package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

const tableName = "btc_5min"

func setupDatabase(ctx context.Context) (*pgx.Conn, error) {
	connStr := "postgres://postgres:mysecretpassword@localhost:5432/postgres"

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}

	err = createTable(ctx, conn, tableName)
	if err != nil {
		return nil, err
	}

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

func getLatestTimestamp(ctx context.Context, conn *pgx.Conn, tableName string) (int64, bool, error) {
	query := fmt.Sprintf(`
		SELECT EXTRACT(EPOCH FROM time) AS unix_seconds
		FROM %s
		ORDER BY time DESC
		LIMIT 1;
	`, tableName)

	result, err := conn.Query(ctx, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not execute query to get latest item in %s: %s\n", tableName, err)
		return 0, false, err
	}

	if !result.Next() {
		return 0, false, nil
	}

	var latestTimestamp int64
	err = result.Scan(&latestTimestamp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not execute query to get values from result: %s\n", err)
		return 0, false, err
	}

	return latestTimestamp, true, nil
}

func insertDataPoint(ctx context.Context, conn *pgx.Conn, tableName string, element HistoricalDataPoint) error {
	query := fmt.Sprintf(`
		INSERT INTO %s  (time, open, high, low, close, volume)
		VALUES (to_timestamp($1), $2, $3, $4, $5, $6)
	`, tableName)

	result, err := conn.Exec(ctx, query, element.Timestamp, element.Open, element.High, element.Low, element.Close, element.Volume)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not execute query to insert row in %s: %s\n", tableName, err)
		return err
	}

	if result.RowsAffected() != 1 {
		fmt.Fprintf(os.Stderr, "could not execute query to insert row in %s\n", tableName)
		return fmt.Errorf("unexpected number of rows affected: %d", result.RowsAffected())
	}

	return nil
}
