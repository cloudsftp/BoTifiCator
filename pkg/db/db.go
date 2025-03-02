package db

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"github.com/cloudsftp/botificator/pkg/api"
)

const (
	ohclTable        = "btc_5min"
	dailyAverageView = "btc_daily_avg"
)

type DataProvider struct {
	pool *pgxpool.Pool
	lock *sync.RWMutex
}

// GetLatestTimestamp returns the timestamp of the latest row
func (d *DataProvider) GetLatestTimestamp(ctx context.Context) (int64, bool, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	query := fmt.Sprintf(`
		SELECT EXTRACT(EPOCH FROM time) AS unix_seconds
		FROM %s
		ORDER BY time DESC
		LIMIT 1
	`, ohclTable)

	result, err := d.pool.Query(ctx, query)
	if err != nil {
		logrus.Errorf("could not execute query to get latest item in %s: %s", ohclTable, err)
		return 0, false, err
	}
	defer result.Close()

	if !result.Next() {
		return 0, false, nil
	}

	var latestTimestamp int64
	err = result.Scan(&latestTimestamp)
	if err != nil {
		logrus.Errorf("could not execute query to get values from result: %s", err)
		return 0, false, err
	}

	return latestTimestamp, true, nil
}

// InsertDataPoints efficiently inserts multiple data points using COPY.
// Returns true if any rows were inserted, false otherwise.
func (d *DataProvider) InsertDataPoints(ctx context.Context, elements []api.HistoricalDataPoint) (bool, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	copyCount, err := d.pool.CopyFrom(
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
		logrus.Errorf("could not execute query to insert rows: %s", err)
		return false, err
	}

	if copyCount == 0 {
		logrus.Error("no rows inserted")
		return false, nil
	}

	return true, nil
}

func movingAverageSqlRange(numRows uint64) string {
	return fmt.Sprintf("(ORDER BY day DESC ROWS BETWEEN CURRENT ROW AND %d FOLLOWING)", numRows-1)
}

type MovingAverages struct {
	Time               time.Time
	DailyAverage       float64
	MovingAverage111   float64
	MovingAverage350x2 float64
}

func (d *DataProvider) GetMovingAverages(limit uint, ctx context.Context) ([]MovingAverages, error) {
	query := fmt.Sprintf(`
		SELECT
			day,
			average,
			avg(average) over %s AS ma111,
			2 * avg(average) over %s AS ma350x2
		FROM %s
		ORDER BY day DESC
		LIMIT %d;
    `,
		movingAverageSqlRange(111),
		movingAverageSqlRange(350),
		dailyAverageView,
		limit,
	)

	result, err := d.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("could not get moving averages: %w", err)
	}
	defer result.Close()

	var averages []MovingAverages
	for result.Next() {
		var row MovingAverages
		err = result.Scan(
			&row.Time,
			&row.DailyAverage,
			&row.MovingAverage111,
			&row.MovingAverage350x2,
		)
		if err != nil {
			return nil, fmt.Errorf("could not scan row: %w", err)
		}

		averages = append(averages, row)
	}

	return averages, nil
}
