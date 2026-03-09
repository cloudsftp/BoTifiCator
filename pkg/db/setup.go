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
