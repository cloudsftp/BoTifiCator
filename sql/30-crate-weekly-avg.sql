CREATE MATERIALIZED VIEW IF NOT EXISTS
    btc_avg_1week
WITH (timescaledb.continuous) AS
SELECT
    day,
    time_bucket('1 week', day) AS week,
    avg(daily_average) as weekly_average
FROM
    btc_avg_1day
GROUP BY
    day,
    week
;

ALTER MATERIALIZED VIEW
    btc_avg_1week
SET
    (timescaledb.materialized_only = false)
;

CREATE OR REPLACE VIEW
    btc_moving_avg_1week
AS SELECT
    day,
    time_bucket('1 week', week) as week,
    avg(weekly_average) OVER(ORDER BY week ROWS BETWEEN 199 PRECEDING AND CURRENT ROW) as weekly_moving_average_200,
    avg(weekly_average) OVER(ORDER BY week ROWS BETWEEN  99 PRECEDING AND CURRENT ROW) as weekly_moving_average_100
FROM
    btc_avg_1week
ORDER BY
    day,
    week
DESC
