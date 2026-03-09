CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE MATERIALIZED VIEW IF NOT EXISTS btc_avg_1day
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', time) AS day,
    avg(open) as average
FROM
    btc_ohlc_5min
GROUP BY
    day
;

--CREATE MATERIALIZED VIEW IF NOT EXISTS btc_avg_1week
--WITH (timescaledb.continuous) AS
--SELECT
--    time_bucket('1 week', day) AS week,
--    avg(average) as average
--FROM
--    btc_avg_1day
--GROUP BY
--    week
--;
