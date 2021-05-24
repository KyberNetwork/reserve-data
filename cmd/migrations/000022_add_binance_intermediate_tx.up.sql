CREATE TABLE "binance_intermediate_tx" (
    id SERIAL PRIMARY KEY NOT NULL,
    timepoint BIGINT NOT NULL,
    eid		  TEXT NOT NULL,
    txhash    TEXT NOT NULL,
    nonce     INT NOT NULL,
    asset_id  INT NOT NULL,
    exchange_id INT NOT NULL,
    gas_price FLOAT NOT NULL,
    amount    FLOAT NOT NULL,
    status    TEXT NOT NULL DEFAULT '',
    created   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX timepoint_eid_idx ON binance_intermediate_tx(timepoint, eid);