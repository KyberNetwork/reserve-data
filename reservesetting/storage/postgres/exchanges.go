package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

type exchangeDB struct {
	ID              rtypes.ExchangeID `db:"id"`
	Name            string            `db:"name"`
	TradingFeeMaker sql.NullFloat64   `db:"trading_fee_maker"`
	TradingFeeTaker sql.NullFloat64   `db:"trading_fee_taker"`
	Disable         bool              `db:"disable"`
}

func (s *Storage) GetExchanges() ([]common.Exchange, error) {
	var (
		qResults []exchangeDB
		results  []common.Exchange
	)

	if err := s.stmts.getExchanges.Select(&qResults); err != nil {
		return nil, fmt.Errorf("failed to query from database err=%s", err.Error())
	}

	for _, qResult := range qResults {
		result := common.Exchange{
			ID:      qResult.ID,
			Name:    qResult.Name,
			Disable: qResult.Disable,
		}

		if qResult.TradingFeeMaker.Valid {
			result.TradingFeeMaker = qResult.TradingFeeMaker.Float64
		}
		if qResult.TradingFeeTaker.Valid {
			result.TradingFeeTaker = qResult.TradingFeeTaker.Float64
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *Storage) GetExchange(id rtypes.ExchangeID) (common.Exchange, error) {
	var (
		qResult = exchangeDB{}
		result  common.Exchange
	)

	s.l.Infow("querying exchange from database", "id", id)
	if err := s.stmts.getExchange.Get(&qResult, id); err != nil {
		if err == sql.ErrNoRows {
			return common.Exchange{}, common.ErrNotFound
		}
		return common.Exchange{}, err
	}
	result = common.Exchange{
		ID:      qResult.ID,
		Name:    qResult.Name,
		Disable: qResult.Disable,
	}
	if qResult.TradingFeeMaker.Valid {
		result.TradingFeeMaker = qResult.TradingFeeMaker.Float64
	}
	if qResult.TradingFeeTaker.Valid {
		result.TradingFeeTaker = qResult.TradingFeeTaker.Float64
	}
	return result, nil
}

// GetExchangeByName return exchange by its name
func (s *Storage) GetExchangeByName(name string) (common.Exchange, error) {
	var (
		qResult = exchangeDB{}
		result  common.Exchange
	)
	s.l.Infow("querying exchange from database", "name", name)
	if err := s.stmts.getExchangeByName.Get(&qResult, name); err != nil {
		if err == sql.ErrNoRows {
			return result, common.ErrNotFound
		}
	}
	result = common.Exchange{
		ID:      qResult.ID,
		Name:    qResult.Name,
		Disable: qResult.Disable,
	}
	if qResult.TradingFeeMaker.Valid {
		result.TradingFeeMaker = qResult.TradingFeeMaker.Float64
	}
	if qResult.TradingFeeTaker.Valid {
		result.TradingFeeTaker = qResult.TradingFeeTaker.Float64
	}
	return result, nil
}

func (s *Storage) UpdateExchange(id rtypes.ExchangeID, updateOpts storage.UpdateExchangeOpts) error {
	return s.updateExchange(nil, id, updateOpts)
}

func (s *Storage) updateExchange(tx *sqlx.Tx, id rtypes.ExchangeID, updateOpts storage.UpdateExchangeOpts) error {

	var updateMsgs []string
	if updateOpts.TradingFeeMaker != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("trading_fee_maker=%f", *updateOpts.TradingFeeMaker))
	}
	if updateOpts.TradingFeeTaker != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("trading_fee_taker=%f", *updateOpts.TradingFeeTaker))
	}
	if updateOpts.Disable != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("disable=%t", *updateOpts.Disable))
	}

	s.l.Infow("updating exchange", "id", id, "fields", strings.Join(updateMsgs, " "))
	sts := s.stmts.updateExchange
	if tx != nil {
		sts = tx.NamedStmt(s.stmts.updateExchange)
	}
	var updatedID uint64
	err := sts.Get(&updatedID,
		struct {
			ID              rtypes.ExchangeID `db:"id"`
			TradingFeeMaker *float64          `db:"trading_fee_maker"`
			TradingFeeTaker *float64          `db:"trading_fee_taker"`
			Disable         *bool             `db:"disable"`
		}{
			ID:              id,
			TradingFeeMaker: updateOpts.TradingFeeMaker,
			TradingFeeTaker: updateOpts.TradingFeeTaker,
			Disable:         updateOpts.Disable,
		},
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return common.ErrExchangeNotExists
		}
		pErr, ok := err.(*pq.Error)
		if !ok {
			return fmt.Errorf("unknown error returned err=%s", err.Error())
		}

		// check_violation
		if pErr.Code == errCodeCheckViolation && pErr.Constraint == "disable_check" {
			s.l.Infow("required setting is missing, could not enable exchange")
			return common.ErrExchangeFeeMissing
		}

		return fmt.Errorf("failed to update exchange err=%s", pErr.Message)
	}
	return nil
}
