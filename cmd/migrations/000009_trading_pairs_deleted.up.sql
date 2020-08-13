CREATE TABLE trading_pairs_deleted
(
    deleted_at      timestamptz NOT NULL,
    id               INT PRIMARY KEY,
    exchange_id      INT NOT NULL,
    base_id          INT NOT NULL,
    quote_id         INT NOT NULL,
    price_precision  INT   NOT NULL,
    amount_precision INT   NOT NULL,
    amount_limit_min FLOAT NOT NULL,
    amount_limit_max FLOAT NOT NULL,
    price_limit_min  FLOAT NOT NULL,
    price_limit_max  FLOAT NOT NULL,
    min_notional     FLOAT NOT NULL
);