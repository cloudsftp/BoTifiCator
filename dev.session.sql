create table IF NOT EXISTS test(
    value DECIMAL(20,2)
);

select 
    time,
    close
from btc_5min
;