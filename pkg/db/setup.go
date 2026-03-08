package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func (d *DataProvider) Close() {
	d.pool.Close()
}

func New(ctx context.Context) (*DataProvider, error) {
	pool, err := createConnectionPool(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not create a connection pool: %w", err)
	}

	err = createTables(ctx, pool)
	if err != nil {
		return nil, fmt.Errorf("could not create tables: %w", err)
	}

	return &DataProvider{pool}, nil
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
		return fmt.Errorf("could not create table %s: %w", ohclTable, err)
	}

	_, err = pool.Exec(ctx, fmt.Sprintf(`
		SELECT create_hypertable('%s', 'time', if_not_exists => true)
	`, ohclTable))
	if err != nil {
		return fmt.Errorf("could not create hypertable %s: %w", ohclTable, err)
	}

	// Daily Average
	_, err = pool.Exec(ctx, fmt.Sprintf(`
		CREATE MATERIALIZED VIEW IF NOT EXISTS %s
		WITH (timescaledb.continuous) AS
		SELECT
			time_bucket('1 day', time) AS day,
			AVG(low) as average
		FROM
			%s
		GROUP BY
			day
    `, dailyAverageView, ohclTable))
	if err != nil {
		return fmt.Errorf("could not create view %s : %w", dailyAverageView, err)
	}

	_, err = pool.Exec(ctx, fmt.Sprintf(`
		ALTER MATERIALIZED VIEW %s set (timescaledb.materialized_only = false);
    `, dailyAverageView))
	if err != nil {
		return fmt.Errorf("could not set %s to real time: %w", dailyAverageView, err)
	}

	// Weekly Average
	_, err = pool.Exec(ctx, fmt.Sprintf(`
		CREATE MATERIALIZED VIEW IF NOT EXISTS %s
		WITH (timescaledb.continuous) AS
		SELECT
			time_bucket('1 week', time) AS day,
			AVG(low) as average
		FROM
			%s
		GROUP BY
			day
    `, weeklyAverageView, ohclTable))
	if err != nil {
		return fmt.Errorf("could not create view %s : %w", weeklyAverageView, err)
	}

	_, err = pool.Exec(ctx, fmt.Sprintf(`
		ALTER MATERIALIZED VIEW %s set (timescaledb.materialized_only = false);
    `, weeklyAverageView))
	if err != nil {
		return fmt.Errorf("could not set %s to real time: %w", weeklyAverageView, err)
	}

	return nil
}
