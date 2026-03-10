CREATE MATERIALIZED VIEW IF NOT EXISTS
    btc_avg_1week
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 week', day) AS week,
    avg(average) as average
FROM
    btc_avg_1day
GROUP BY
    week
;

ALTER MATERIALIZED VIEW
    btc_avg_1week
SET
    (timescaledb.materialized_only = false)
;
