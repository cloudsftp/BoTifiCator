CREATE MATERIALIZED VIEW IF NOT EXISTS
    btc_avg_1day
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', time) AS day,
    avg(open) as daily_average
FROM
    btc_ohlc_5min
GROUP BY
    day
;

ALTER MATERIALIZED VIEW
    btc_avg_1day
SET
    (timescaledb.materialized_only = false)
;
