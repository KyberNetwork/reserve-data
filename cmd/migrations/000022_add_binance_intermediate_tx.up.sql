CREATE TABLE "binance_intermediate_tx" (
    timepoint 			BIGINT NOT NULL,
    eid					TEXT NOT NULL,
    data				JSON NOT NULL,
    PRIMARY KEY (timepoint, eid)
);
