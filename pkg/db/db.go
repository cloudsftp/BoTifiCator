package db

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"github.com/cloudsftp/botificator/pkg/api"
)

const (
	ohlcTable = "btc_ohlc_5min"

	dailyAverageView    = "btc_avg_1day"
	weeklyAverageView   = "btc_avg_1week"
	movingWeeklyAvgView = "btc_moving_avg_1week"
)

type DataProvider struct {
	pool *pgxpool.Pool
}

// GetLatestTimestamp returns the timestamp of the latest row
func (d *DataProvider) GetLatestTimestamp(ctx context.Context) (int64, bool) {
	query := fmt.Sprintf(`
		SELECT
            EXTRACT(EPOCH FROM time) AS unix_seconds
		FROM %s
		ORDER BY time DESC
		LIMIT 1
	`, ohlcTable)

	row := d.pool.QueryRow(ctx, query)

	var latestTimestamp int64
	err := row.Scan(&latestTimestamp)
	if err != nil {
		return 0, false
	}

	return latestTimestamp, true
}

// InsertDataPoints efficiently inserts multiple data points using COPY.
// Returns true if any rows were inserted, false otherwise.
func (d *DataProvider) InsertDataPoints(ctx context.Context, elements []api.HistoricalDataPoint) (bool, error) {
	copyCount, err := d.pool.CopyFrom(
		ctx,
		pgx.Identifier{ohlcTable},
		[]string{"time", "open", "high", "low", "close", "volume"},
		pgx.CopyFromSlice(len(elements), func(i int) ([]any, error) {
			element := elements[i]

			unixSeconds, err := strconv.ParseInt(element.Timestamp, 10, 64)
			if err != nil {
				return nil, err
			}
			timeDate := time.Unix(unixSeconds, 0)

			return []any{
				timeDate,
				element.Open,
				element.High,
				element.Low,
				element.Close,
				element.Volume,
			}, nil
		}),
	)

	if err != nil {
		return false, fmt.Errorf("could not execute query to insert rows: %w", err)
	}

	if copyCount == 0 {
		logrus.Warn("no rows inserted")
		return false, nil
	}

	return true, nil
}

type ReportDataDaily struct {
	Day     time.Time `db:"day"`
	Average float64   `db:"average"`
}

func (d *DataProvider) getReportDataDaily(ctx context.Context) ([]ReportDataDaily, error) {
	query := fmt.Sprintf(`
		SELECT
			day,
			average
		FROM %s
		ORDER BY day DESC
		OFFSET 1
		LIMIT 2
	`, dailyAverageView)

	rows, err := d.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("could not get daily report data: %w", err)
	}

	values, err := pgx.CollectRows(rows, pgx.RowToStructByName[ReportDataDaily])
	if err != nil {
		return nil, fmt.Errorf("could not deserialize daily report data: %w", err)
	}

	return values, nil
}

type ReportDataWeekly struct {
	Week             time.Time `db:"week"`
	Average          float64   `db:"average"`
	MovingAverage200 float64   `db:"moving_average_200"`
	MovingAverage100 float64   `db:"moving_average_100"`
}

func (d *DataProvider) getReportDataWeekly(ctx context.Context) ([]ReportDataWeekly, error) {
	query := fmt.Sprintf(`
		SELECT
			w.week,
			w.average,
			m.moving_average_200,
			m.moving_average_100
		FROM %s w
		JOIN %s m
		ON
            w.week = m.week
		ORDER BY w.week DESC
		OFFSET 1
		LIMIT 2
	`, weeklyAverageView, movingWeeklyAvgView)

	rows, err := d.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("could not get weekly report data: %w", err)
	}

	values, err := pgx.CollectRows(rows, pgx.RowToStructByName[ReportDataWeekly])
	if err != nil {
		return nil, fmt.Errorf("could not deserialize weekly report data: %w", err)
	}

	return values, nil
}

type ReportData struct {
	Day    time.Time
	Daily  []ReportDataDaily
	Weekly []ReportDataWeekly
}

func (d *DataProvider) GetReportData(ctx context.Context) (*ReportData, error) {
	daily, err := d.getReportDataDaily(ctx)
	if err != nil {
		return nil, err
	}
	day := daily[0].Day

	weekly, err := d.getReportDataWeekly(ctx)
	if err != nil {
		return nil, err
	}

	return &ReportData{day, daily, weekly}, nil
}
