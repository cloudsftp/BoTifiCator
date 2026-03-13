CREATE MATERIALIZED VIEW IF NOT EXISTS
    btc_avg_1week
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 week', time) AS week,
    avg(open) as average
FROM
    btc_ohlc_5min
GROUP BY
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
    time_bucket('1 week', week) as week,
    avg(average) OVER(ORDER BY week ROWS BETWEEN 199 PRECEDING AND CURRENT ROW) as moving_average_200,
    avg(average) OVER(ORDER BY week ROWS BETWEEN  99 PRECEDING AND CURRENT ROW) as moving_average_100
FROM
    btc_avg_1week
ORDER BY
    week
DESC
;
