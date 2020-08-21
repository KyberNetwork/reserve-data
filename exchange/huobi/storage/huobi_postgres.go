package storage

import (
	"database/sql"
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/postgres"
	"github.com/KyberNetwork/reserve-data/exchange"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
)

// postgresStorage implements huobi storage in postgres
type postgresStorage struct {
	db    *sqlx.DB
	stmts preparedStmt
}

type preparedStmt struct {
	storeHistoryStmt     *sqlx.NamedStmt
	getHistoryStmt       *sqlx.Stmt
	getLastIDHistoryStmt *sqlx.Stmt

	storePendingTxStmt      *sqlx.Stmt
	getPendingTxStmt        *sqlx.Stmt
	storeIntermediateTxStmt *sqlx.Stmt
	getIntermediateTxStmt   *sqlx.Stmt
}

// NewPostgresStorage creates a new obj exchange.HuobiStorage by db engine=postgres
func NewPostgresStorage(db *sqlx.DB) (exchange.HuobiStorage, error) {
	storage := &postgresStorage{
		db: db,
	}
	err := storage.initStmts()
	return storage, err
}

func (s *postgresStorage) initStmts() error {
	var err error
	// history stmts
	s.stmts.storeHistoryStmt, err = s.db.PrepareNamed(`INSERT INTO "huobi_trade_history"
		(pair_id, trade_id, price, qty, type, time)
		VALUES(:pair_id, :trade_id, :price, :qty, :type, :time) ON CONFLICT (trade_id) DO UPDATE SET 
		                                                                                  price=excluded.price,
		                                                                                  qty=excluded.qty,
		                                                                                  time=excluded.time`)
	if err != nil {
		return err
	}
	s.stmts.getHistoryStmt, err = s.db.Preparex(`SELECT pair_id, trade_id, price, qty, type, time 
		FROM "huobi_trade_history"
		WHERE time >= $1 AND time <= $2`)
	if err != nil {
		return err
	}
	s.stmts.getLastIDHistoryStmt, err = s.db.Preparex(`SELECT pair_id, trade_id, price, qty, type, time FROM "huobi_trade_history"
											WHERE pair_id = $1 
											ORDER BY time DESC, trade_id DESC;`)
	if err != nil {
		return err
	}
	// pending stmts
	s.stmts.storePendingTxStmt, err = s.db.Preparex(`INSERT INTO "huobi_pending_intermediate_tx"
		(timepoint, eid, data)
		VALUES ($1, $2, $3);`)
	if err != nil {
		return err
	}
	s.stmts.getPendingTxStmt, err = s.db.Preparex(`SELECT timepoint, eid, data 
		FROM "huobi_pending_intermediate_tx";`)
	if err != nil {
		return err
	}
	s.stmts.storeIntermediateTxStmt, err = s.db.Preparex(`SELECT NULL FROM new_intermediate_tx($1, $2, $3)`)
	if err != nil {
		return err
	}

	s.stmts.getIntermediateTxStmt, err = s.db.Preparex(`SELECT data FROM "huobi_intermediate_tx"
				WHERE timepoint = $1 and eid =$2`)
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

type TxDB struct {
	Timepoint uint64 `db:"timepoint"`
	EID       string `db:"eid"`
	Data      []byte `db:"data"`
}

// StoreTradeHistory implements exchange.HuobiStorage and store trade history
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

// GetTradeHistory implements exchange.HuobiStorage and get trade history in a time period
func (s *postgresStorage) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	var result = make(common.ExchangeTradeHistory)
	var records []exchangeTradeHistoryDB
	err := s.stmts.getHistoryStmt.Select(&records, fromTime, toTime)
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

// GetLastIDTradeHistory implements exchange.HuobiStorage and returns the last ID with correspond pairID
func (s *postgresStorage) GetLastIDTradeHistory(pairID uint64) (string, error) {
	var record exchangeTradeHistoryDB
	err := s.stmts.getLastIDHistoryStmt.Get(&record, pairID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.Errorf("no history with pair_id=%v", pairID)
		}
		return "", err
	}
	return record.TradeID, nil

}

// StoreIntermediateTx implements exchange.HuobiStorage, delete pending tx and create a intermediate tx
func (s *postgresStorage) StoreIntermediateTx(id common.ActivityID, data common.TXEntry) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = s.stmts.storeIntermediateTxStmt.Exec(id.Timepoint, id.EID, dataJSON)
	return err
}

// StorePendingIntermediateTx implements exchange.HuobiStorage and store pending tx
func (s *postgresStorage) StorePendingIntermediateTx(id common.ActivityID, data common.TXEntry) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = s.stmts.storePendingTxStmt.Exec(id.Timepoint, id.EID, dataJSON)
	return err
}

// GetIntermedatorTx implements exchange.HuobiStorage and get intermediate tx
func (s *postgresStorage) GetIntermedatorTx(id common.ActivityID) (common.TXEntry, error) {
	var dataJSON []byte
	var result common.TXEntry
	err := s.stmts.getIntermediateTxStmt.Get(&dataJSON, id.Timepoint, id.EID)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(dataJSON, &result)
	return result, err
}

// GetPendingIntermediateTXs implements exchange.HuobiStorage and get pending tx
func (s *postgresStorage) GetPendingIntermediateTXs() (map[common.ActivityID]common.TXEntry, error) {
	var records []TxDB
	var result = make(map[common.ActivityID]common.TXEntry)
	err := s.stmts.getPendingTxStmt.Select(&records)
	if err != nil {
		return nil, err
	}
	for _, r := range records {
		activityID := common.ActivityID{
			EID:       r.EID,
			Timepoint: r.Timepoint,
		}
		var txEntry common.TXEntry
		err := json.Unmarshal(r.Data, &txEntry)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal tx entry: %v", string(r.Data))
		}
		result[activityID] = txEntry
	}
	return result, nil
}
