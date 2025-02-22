# BoTifiCator

Project for monitoring some selected bitcoin indicators

## Database

Up

    docker run --name timescaledb-container -e POSTGRES_PASSWORD=mysecretpassword -p 5432:5432 -d timescale/timescaledb:latest-pg15
    
Interact

    docker exec -it timescaledb-container psql -U postgres -p 5432 -h localhost

Create table

    CREATE TABLE btc (
        time TIMESTAMPTZ NOT NULL,
        open DECIMAL NOT NULL,
        high DECIMAL NOT NULL,
        low DECIMAL NOT NULL,
        close DECIMAL NOT NULL,
        volume DECIMAL NOT NULL
    );
    
    SELECT create_hypertable('btc', 'time');


