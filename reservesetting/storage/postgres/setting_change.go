package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	pgutil "github.com/KyberNetwork/reserve-data/common/postgres"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

const (
	settingChangeCatUnique = "setting_change_cat_key"
)

// CreateSettingChange creates an setting change in database and return id
func (s *Storage) CreateSettingChange(cat common.ChangeCatalog, obj common.SettingChange) (rtypes.SettingChangeID, error) {
	var id rtypes.SettingChangeID
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to parse json data %+v", obj)
	}
	tx, err := s.db.Beginx()
	if err != nil {
		return 0, err
	}
	defer pgutil.RollbackUnlessCommitted(tx)

	if err = tx.Stmtx(s.stmts.newSettingChange).Get(&id, cat.String(), jsonData); err != nil {
		pErr, ok := err.(*pq.Error)
		if !ok {
			return 0, fmt.Errorf("unknown returned err=%s", err.Error())
		}

		s.l.Infow("failed to create new setting change", "err", pErr.Message)
		if pErr.Code == errCodeUniqueViolation && pErr.Constraint == settingChangeCatUnique {
			return 0, common.ErrSettingChangeExists
		}
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	s.l.Infow("create setting change success", "id", id)
	return id, nil
}

type settingChangeDB struct {
	ID      rtypes.SettingChangeID `db:"id"`
	Created time.Time              `db:"created"`
	Data    []byte                 `db:"data"`
}

func (objDB settingChangeDB) ToCommon() (common.SettingChangeResponse, error) {
	var settingChange common.SettingChange
	err := json.Unmarshal(objDB.Data, &settingChange)
	if err != nil {
		return common.SettingChangeResponse{}, err
	}
	return common.SettingChangeResponse{
		ChangeList: settingChange.ChangeList,
		ID:         objDB.ID,
		Created:    objDB.Created,
	}, nil
}

// GetSettingChange returns a object with a given id
func (s *Storage) GetSettingChange(id rtypes.SettingChangeID) (common.SettingChangeResponse, error) {
	return s.getSettingChange(nil, id)
}

func (s *Storage) getSettingChange(tx *sqlx.Tx, id rtypes.SettingChangeID) (common.SettingChangeResponse, error) {
	var dbResult settingChangeDB
	sts := s.stmts.getSettingChange
	if tx != nil {
		sts = tx.Stmtx(sts)
	}
	err := sts.Get(&dbResult, id, nil, nil)
	if err != nil {
		if err == sql.ErrNoRows {
			return common.SettingChangeResponse{}, common.ErrNotFound
		}
		return common.SettingChangeResponse{}, err
	}
	res, err := dbResult.ToCommon()
	if err != nil {
		s.l.Errorw("failed to convert to common setting change", "err", err)
		return common.SettingChangeResponse{}, err
	}
	return res, nil
}

// GetSettingChanges return list setting change.
func (s *Storage) GetSettingChanges(cat common.ChangeCatalog, status common.ChangeStatus) ([]common.SettingChangeResponse, error) {
	s.l.Infow("get setting type", "catalog", cat)
	var dbResult []settingChangeDB
	err := s.stmts.getSettingChange.Select(&dbResult, nil, cat.String(), status.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	var result = make([]common.SettingChangeResponse, 0, 1) // although it's a slice, we expect only 1 for now.
	for _, p := range dbResult {
		rr, err := p.ToCommon()
		if err != nil {
			return nil, err
		}
		result = append(result, rr)
	}
	return result, nil
}

// RejectSettingChange delete setting change with a given id
func (s *Storage) RejectSettingChange(id rtypes.SettingChangeID) error {
	var returnedID uint64
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer pgutil.RollbackUnlessCommitted(tx)
	err = tx.Stmtx(s.stmts.updateSettingChangeStatus).Get(&returnedID, id, common.ChangeStatusRejected.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return common.ErrNotFound
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	s.l.Infow("reject setting change success", "id", id)
	return nil
}

func (s *Storage) applyChange(tx *sqlx.Tx, i int, entry common.SettingChangeEntry) error {
	var err error
	switch e := entry.Data.(type) {
	case *common.ChangeAssetAddressEntry:
		err = s.changeAssetAddress(tx, e.ID, e.Address)
		if err != nil {
			s.l.Infow("change asset address", "index", i, "err", err)
			return err
		}
	case *common.CreateAssetEntry:
		_, err = s.createAsset(tx, e.Symbol, e.Name, e.Address, e.Decimals, e.Transferable, e.SetRate, e.Rebalance,
			e.IsQuote, e.IsEnabled, e.PWI, e.RebalanceQuadratic, e.Exchanges, e.Target, e.StableParam, e.FeedWeight,
			e.NormalUpdatePerPeriod, e.MaxImbalanceRatio)
		if err != nil {
			s.l.Infow("create asset", "index", i, "err", err)
			return err
		}
	case *common.CreateAssetExchangeEntry:
		_, err = s.createAssetExchange(tx, e.ExchangeID, e.AssetID, e.Symbol, e.DepositAddress, e.MinDeposit,
			e.WithdrawFee, e.TargetRecommended, e.TargetRatio, e.TradingPairs)
		if err != nil {
			s.l.Infow("create asset exchange", "index", i, "err", err)
			return err
		}
	case *common.CreateTradingPairEntry:
		_, err = s.createTradingPair(tx, e.ExchangeID, e.Base, e.Quote, e.PricePrecision, e.AmountPrecision, e.AmountLimitMin,
			e.AmountLimitMax, e.PriceLimitMin, e.PriceLimitMax, e.MinNotional, e.AssetID)
		if err != nil {
			s.l.Infow("create trading pair", "index", i, "err", err)
			return err
		}
	case *common.UpdateAssetEntry:
		err = s.updateAsset(tx, e.AssetID, *e)
		if err != nil {
			s.l.Infow("update asset", "index", i, "err", err)
			return err
		}
	case *common.UpdateAssetExchangeEntry:
		err = s.updateAssetExchange(tx, e.ID, *e)
		if err != nil {
			s.l.Infow("update asset exchange", "index", i, "err", err)
			return err
		}
	case *common.UpdateExchangeEntry:
		err = s.updateExchange(tx, e.ExchangeID, *e)
		if err != nil {
			s.l.Infow("update exchange", "index", i, "err", err)
			return err
		}
	case *common.DeleteAssetExchangeEntry:
		err = s.deleteAssetExchange(tx, e.AssetExchangeID)
		if err != nil {
			s.l.Infow("delete asset exchange", "index", i, "err", err)
			return err
		}
	case *common.DeleteTradingPairEntry:
		err = s.deleteTradingPair(tx, e.TradingPairID)
		if err != nil {
			s.l.Infow("delete trading pair", "index", i, "err", err)
			return err
		}
	case *common.UpdateStableTokenParamsEntry:
		err = s.updateStableTokenParams(tx, e.Params)
		if err != nil {
			s.l.Infow("update stable token params", "index", i, "err", err)
			return err
		}
	case *common.SetFeedConfigurationEntry:
		err = s.setFeedConfiguration(tx, *e)
		if err != nil {
			s.l.Infow("set feed configuration", "index", i, "err", err)
			return err
		}
	default:
		return fmt.Errorf("unexpected change object %+v", e)
	}
	return nil
}

// ConfirmSettingChange apply setting change with a given id
func (s *Storage) ConfirmSettingChange(id rtypes.SettingChangeID, commit bool) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "create transaction error")
	}
	defer pgutil.RollbackUnlessCommitted(tx)
	changeObj, err := s.getSettingChange(tx, id)
	if err != nil {
		return errors.Wrap(err, "get setting change error")
	}

	for i, change := range changeObj.ChangeList {
		if err = s.applyChange(tx, i, change); err != nil {
			return err
		}
	}
	_, err = tx.Stmtx(s.stmts.updateSettingChangeStatus).Exec(id, common.ChangeStatusAccepted.String())
	if err != nil {
		return err
	}
	if commit {
		if err := tx.Commit(); err != nil {
			s.l.Infow("setting change has been failed to confirm", "id", id, "err", err)
			return err
		}
		s.l.Infow("setting change has been confirmed successfully", "id", id)
		return nil
	}
	s.l.Infow("setting change will be reverted due commit flag not set", "id", id)
	return nil
}
