package db

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/cloudsftp/botificator/pkg/api"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = "btc_5min"

func SetupDatabase(ctx context.Context) (*pgxpool.Pool, error) {
	connectionString := "postgres://postgres:mysecretpassword@localhost:5432/postgres"

	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 20
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	err = createTable(ctx, pool, tableName)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func createTable(ctx context.Context, pool *pgxpool.Pool, tableName string) error {
	_, err := pool.Exec(ctx, fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			time   TIMESTAMPTZ    NOT NULL UNIQUE,
			open   DECIMAL(30,5)  NOT NULL,
			high   DECIMAL(30,5)  NOT NULL,
			low    DECIMAL(30,5)  NOT NULL,
			close  DECIMAL(30,5)  NOT NULL,
			volume DECIMAL(40,20) NOT NULL
		)
	`, tableName))
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, fmt.Sprintf(`
		SELECT create_hypertable('%s', 'time', if_not_exists => true)
	`, tableName))
	if err != nil {
		return err
	}

	return nil
}

func GetLatestTimestamp(ctx context.Context, conn *pgx.Conn) (int64, bool, error) {
	query := fmt.Sprintf(`
		SELECT EXTRACT(EPOCH FROM time) AS unix_seconds
		FROM %s
		ORDER BY time DESC
		LIMIT 1
	`, tableName)

	result, err := conn.Query(ctx, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not execute query to get latest item in %s: %s\n", tableName, err)
		return 0, false, err
	}
	defer result.Close()

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

func InsertDataPoints(ctx context.Context, conn *pgx.Conn, elements []api.HistoricalDataPoint) error {
	copyCount, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{tableName},
		[]string{"time", "open", "high", "low", "close", "volume"},
		pgx.CopyFromSlice(len(elements), func(i int) ([]any, error) {
			element := elements[i]

			unixSeconds, err := strconv.ParseInt(element.Timestamp, 10, 64)
			if err != nil {
				return nil, err
			}
			timeDate := time.Unix(unixSeconds, 0)

			return []any{timeDate, element.Open, element.High, element.Low, element.Close, element.Volume}, nil
		}),
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "could not execute query to insert rows: %s\n", err)
		return err
	}

	if copyCount == 0 {
		fmt.Fprint(os.Stderr, "row not inserted, time already exists\n")
	}

	return nil
}
