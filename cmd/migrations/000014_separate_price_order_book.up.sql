CREATE TABLE "order_book_data"
(
    id SERIAL,
    created TIMESTAMPTZ NOT NULL,
    data JSON NOT NULL,
    PRIMARY KEY (id,created)
) PARTITION BY RANGE (created);

CREATE TABLE order_book_data_default PARTITION OF order_book_data DEFAULT;
CREATE INDEX "order_book_data_created_index" ON "order_book_data" (created);
