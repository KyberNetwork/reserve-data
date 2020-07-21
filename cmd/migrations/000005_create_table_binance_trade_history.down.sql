		ALTER TABLE binance_trade_history 
			DROP CONSTRAINT IF EXISTS binance_trade_history_pair_fkey;
        ALTER TABLE binance_trade_history 
			ALTER COLUMN pair_id TYPE BIGINT;