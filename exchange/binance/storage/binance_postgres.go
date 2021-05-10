package storage

import (
	"encoding/json"

	"github.com/jmoiron/sqlx"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/postgres"
	"github.com/KyberNetwork/reserve-data/exchange"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"

	_ "github.com/golang-migrate/migrate/v4/source/file" // driver for migration
)

// postgresStorage implements binance storage in postgres
type postgresStorage struct {
	db    *sqlx.DB
	stmts preparedStmt
}
type preparedStmt struct {
	storeHistoryStmt     *sqlx.NamedStmt
	getHistoryStmt       *sqlx.Stmt
	getLastIDHistoryStmt *sqlx.Stmt
}

// NewPostgresStorage creates new obj exchange.BinanceStorage with db engine = postgres
func NewPostgresStorage(db *sqlx.DB) (exchange.BinanceStorage, error) {
	storage := &postgresStorage{
		db: db,
	}
	return storage, storage.prepareStmts()
}

func (s *postgresStorage) prepareStmts() error {
	var err error
	s.stmts.storeHistoryStmt, err = s.db.PrepareNamed(`INSERT INTO "binance_trade_history"
		(pair_id, trade_id, price, qty, type, time)
		VALUES(:pair_id, :trade_id, :price, :qty, :type, :time) ON CONFLICT (trade_id) DO UPDATE SET
		price=excluded.price, qty=excluded.qty,time=excluded.time`)
	if err != nil {
		return err
	}
	s.stmts.getHistoryStmt, err = s.db.Preparex(`SELECT pair_id, trade_id, price, qty, type, time 
		FROM "binance_trade_history"
		JOIN trading_pairs ON binance_trade_history.pair_id = trading_pairs.id
		WHERE trading_pairs.exchange_id = $1 AND time >= $2 AND time <= $3`)
	if err != nil {
		return err
	}
	s.stmts.getLastIDHistoryStmt, err = s.db.Preparex(`SELECT pair_id, trade_id, price, qty, type, time FROM "binance_trade_history"
											WHERE pair_id = $1 
											ORDER BY time DESC, trade_id DESC;`)
	if err != nil {
		return err
	}
	return nil
}

type exchangeTradeHistoryDB struct {
	PairID  rtypes.TradingPairID `db:"pair_id"`
	TradeID string               `db:"trade_id"`
	Price   float64              `db:"price"`
	Qty     float64              `db:"qty"`
	Type    string               `db:"type"`
	Time    uint64               `db:"time"`
}

// StoreTradeHistory implements exchange.BinanceStorage and store trade history
func (s *postgresStorage) StoreTradeHistory(data common.ExchangeTradeHistory) error {
	// TODO: change this code when jmoiron/sqlx releases bulk request feature
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer postgres.RollbackUnlessCommitted(tx)

	for pairID, tradeHistory := range data {
		for _, history := range tradeHistory {
			_, err = tx.NamedStmt(s.stmts.storeHistoryStmt).Exec(exchangeTradeHistoryDB{
				PairID:  pairID,
				TradeID: history.ID,
				Price:   history.Price,
				Qty:     history.Qty,
				Type:    history.Type,
				Time:    history.Timestamp,
			})
			if err != nil {
				return err
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

// GetTradeHistory implements exchange.BinanceStorage and get trade history within a time period
func (s *postgresStorage) GetTradeHistory(exchangeID rtypes.ExchangeID, fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	var result = make(common.ExchangeTradeHistory)
	var records []exchangeTradeHistoryDB
	err := s.stmts.getHistoryStmt.Select(&records, exchangeID, fromTime, toTime)
	if err != nil {
		return result, err
	}
	for _, r := range records {
		tradeHistory := result[r.PairID]
		tradeHistory = append(tradeHistory, common.TradeHistory{
			ID:        r.TradeID,
			Price:     r.Price,
			Qty:       r.Qty,
			Type:      r.Type,
			Timestamp: r.Time,
		})
		result[r.PairID] = tradeHistory
	}
	return result, nil
}

// GetLastIDTradeHistory implements exchange.BinanceStorage and get the last ID with a correspond pairID
func (s *postgresStorage) GetLastIDTradeHistory(pairID rtypes.TradingPairID) (string, error) {
	var record exchangeTradeHistoryDB
	err := s.stmts.getLastIDHistoryStmt.Get(&record, pairID)
	if err != nil {
		// if err == sql.ErrorNoRow then last id trade history  equal 0
		return "", err
	}
	return record.TradeID, nil
}

// StoreIntermediateDeposit ...
func (s *postgresStorage) StoreIntermediateDeposit(id common.ActivityID, activity common.TXEntry) error {
	query := `INSERT INTO "binance_intermediate_tx" (timepoint, eid, data) 
	VALUES ($1, $2, $3) ON CONFLICT (timepoint, eid) DO UPDATE SET data= EXCLUDED.data;`
	timepoint := id.Timepoint
	dataJSON, err := json.Marshal(activity)
	if err != nil {
		return err
	}
	if _, err := s.db.Exec(query, timepoint, id.EID, dataJSON); err != nil {
		return err
	}
	return nil
}

type TxData struct {
	Data []byte `db:"data"`
}

// GetPendingIntermediateTx ...
func (s *postgresStorage) GetPendingIntermediateTx(id common.ActivityID) (common.TXEntry, error) {
	var (
		result common.TXEntry
		data   TxData
	)
	query := `SELECT data FROM "binance_intermediate_tx" WHERE eid = $1;`
	if err := s.db.Get(&data, query, id.EID); err != nil {
		return result, err
	}
	err := json.Unmarshal(data.Data, &result)
	return result, err
}
