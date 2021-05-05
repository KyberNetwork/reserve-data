CREATE TABLE "binance_intermediate_tx" (
    timepoint 			BIGINT NOT NULL,
    eid					TEXT NOT NULL,
    data				JSON NOT NULL,
    isPending BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (timepoint, eid)
);
