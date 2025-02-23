package db

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cloudsftp/botificator/pkg/api"
)

const (
	ohclTable             = "btc_5min"
	dailyAverageView      = "btc_daily_avg"
	daily100MovingAverage = "btc_daily_100ma"
)

func SetupDatabase(ctx context.Context) (*pgxpool.Pool, error) {
	pool, err := createConnectionPool(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not create a connection pool: %w", err)
	}

	err = createTables(ctx, pool)
	if err != nil {
		return nil, fmt.Errorf("could not create tables: %w", err)
	}

	return pool, nil
}

func createConnectionPool(ctx context.Context) (*pgxpool.Pool, error) {
	pass := os.Getenv("DB_PASS")
	if pass == "" {
		return nil, fmt.Errorf("no environment variable DB_PASS")
	}

	host := os.Getenv("DB_HOST")
	if host == "" {
		return nil, fmt.Errorf("no environment variable DB_HOST")
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		return nil, fmt.Errorf("no environment variable DB_PORT")
	}

	connectionString := fmt.Sprintf("postgres://postgres:%s@%s:%s/postgres", pass, host, port)

	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 20
	config.MinConns = 2

	return pgxpool.NewWithConfig(ctx, config)
}

func createTables(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			time   TIMESTAMPTZ    NOT NULL UNIQUE,
			open   DECIMAL(30,5)  NOT NULL,
			high   DECIMAL(30,5)  NOT NULL,
			low    DECIMAL(30,5)  NOT NULL,
			close  DECIMAL(30,5)  NOT NULL,
			volume DECIMAL(40,20) NOT NULL
		)
	`, ohclTable))
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, fmt.Sprintf(`
		SELECT create_hypertable('%s', 'time', if_not_exists => true)
	`, ohclTable))
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, fmt.Sprintf(`
		CREATE MATERIALIZED VIEW %s
		WITH (timescaledb.continuous) AS
		SELECT
			time_bucket('1 day', time) AS bucket,
			AVG(close)
		FROM
			%s
		GROUP BY
			bucket
    `, dailyAverageView, ohclTable))

	_, err = pool.Exec(ctx, fmt.Sprintf(`
		ALTER MATERIALIZED VIEW %s set (timescaledb.materialized_only = false);
    `, dailyAverageView))

	return nil
}

func GetLatestTimestamp(ctx context.Context, conn *pgx.Conn) (int64, bool, error) {
	query := fmt.Sprintf(`
		SELECT EXTRACT(EPOCH FROM time) AS unix_seconds
		FROM %s
		ORDER BY time DESC
		LIMIT 1
	`, ohclTable)

	result, err := conn.Query(ctx, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not execute query to get latest item in %s: %s\n", ohclTable, err)
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

// InsertDataPoints efficiently inserts multiple data points using COPY.
// Returns true if any rows were inserted, false otherwise.
func InsertDataPoints(ctx context.Context, conn *pgx.Conn, elements []api.HistoricalDataPoint) (bool, error) {
	copyCount, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{ohclTable},
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
		return false, err
	}

	if copyCount == 0 {
		fmt.Fprint(os.Stderr, "no rows inserted")
		return false, nil
	}

	return true, nil
}
