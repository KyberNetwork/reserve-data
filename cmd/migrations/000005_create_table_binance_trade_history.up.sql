		ALTER TABLE binance_trade_history 
			ALTER COLUMN pair_id TYPE INTEGER;
		ALTER TABLE binance_trade_history 
			ADD CONSTRAINT binance_trade_history_pair_fkey FOREIGN KEY (pair_id) REFERENCES trading_pairs(id);