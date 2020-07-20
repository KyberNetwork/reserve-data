		CREATE TABLE IF NOT EXISTS "binance_trade_history"(
		    id 				SERIAL PRIMARY KEY,
		    pair_id			BIGINT,
		    trade_id		TEXT UNIQUE NOT NULL,
		    price 			FLOAT NOT NULL,
		    qty 			FLOAT NOT NULL, 
		    type			TEXT NOT NULL,
		    time			BIGINT
		);
		ALTER TABLE binance_trade_history 
			ALTER COLUMN pair_id TYPE INTEGER;
		ALTER TABLE binance_trade_history 
			ADD CONSTRAINT binance_trade_history_pair_fkey FOREIGN KEY (pair_id) REFERENCES trading_pairs(id);