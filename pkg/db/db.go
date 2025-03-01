package db

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	"github.com/cloudsftp/botificator/pkg/analyzer"
	"github.com/cloudsftp/botificator/pkg/api"
)

const (
	ohclTable        = "btc_5min"
	dailyAverageView = "btc_daily_avg"
)

// GetLatestTimestamp returns the timestamp of the latest row
func GetLatestTimestamp(ctx context.Context, conn *pgx.Conn) (int64, bool, error) {
	query := fmt.Sprintf(`
		SELECT EXTRACT(EPOCH FROM time) AS unix_seconds
		FROM %s
		ORDER BY time DESC
		LIMIT 1
	`, ohclTable)

	result, err := conn.Query(ctx, query)
	if err != nil {
		log.Errorf("could not execute query to get latest item in %s: %s", ohclTable, err)
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
	return fmt.Sprintf("(ORDER BY day ROWS BETWEEN %d PRECEDING AND CURRENT ROW)", numRows-1)
}

func GetMovingAverages(ctx context.Context, pool *pgxpool.Pool) ([]analyzer.MovingAverages, error) {
	query := fmt.Sprintf(`
		SELECT
			day,
			avg(average) over %s AS ma111,
			2 * avg(average) over %s AS ma350x2
		FROM %s
		ORDER BY day DESC
		LIMIT 14;
    `,
		movingAverageSqlRange(111),
		movingAverageSqlRange(350),
		dailyAverageView,
	)

	result, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("could not get moving averages: %w", err)
	}
	defer result.Close()

	var averages []analyzer.MovingAverages
	for result.Next() {
		var averagesRow analyzer.MovingAverages
		err = result.Scan(&averagesRow.Time, &averagesRow.Ma111, &averagesRow.Ma350x2)
		if err != nil {
			return nil, fmt.Errorf("could not scan row: %w", err)
		}

		averages = append(averages, averagesRow)
	}

	return averages, nil
}
