CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE IF NOT EXISTS
    btc_ohlc_5min
(
    time   TIMESTAMPTZ    NOT NULL UNIQUE,
    open   DECIMAL(30,5)  NOT NULL,
    high   DECIMAL(30,5)  NOT NULL,
    low    DECIMAL(30,5)  NOT NULL,
    close  DECIMAL(30,5)  NOT NULL,
    volume DECIMAL(40,20) NOT NULL
);

SELECT create_hypertable('btc_ohlc_5min', 'time', if_not_exists => true);

