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
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

const (
	settingChangeCatUnique = "setting_change_cat_key"
)

// CreateSettingChange creates an setting change in database and return id
func (s *Storage) CreateSettingChange(cat common.ChangeCatalog, obj common.SettingChange, keyID string) (rtypes.SettingChangeID, error) {
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

	if err = tx.Stmtx(s.stmts.newSettingChange).Get(&id, cat.String(), jsonData, common.StringPointer(keyID)); err != nil {
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

type scheduleSettingChangeDB struct {
	ID rtypes.SettingChangeID `db:"id"`
}

// GetScheduleSettingChange update the setting change data
func (s *Storage) GetScheduleSettingChange() ([]rtypes.SettingChangeID, error) {
	var dbResult []scheduleSettingChangeDB
	err := s.stmts.getScheduleSettingChange.Select(&dbResult)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	var result []rtypes.SettingChangeID
	for _, r := range dbResult {
		result = append(result, r.ID)
	}
	return result, nil
}

type settingChangeDB struct {
	ID       rtypes.SettingChangeID `db:"id"`
	Created  time.Time              `db:"created"`
	Data     []byte                 `db:"data"`
	Proposer sql.NullString         `db:"proposer"`
	Rejector sql.NullString         `db:"rejector"`
}

func (objDB settingChangeDB) ToCommon() (common.SettingChangeResponse, error) {
	var settingChange common.SettingChange
	err := json.Unmarshal(objDB.Data, &settingChange)
	if err != nil {
		return common.SettingChangeResponse{}, err
	}
	return common.SettingChangeResponse{
		ChangeList:   settingChange.ChangeList,
		ID:           objDB.ID,
		Created:      objDB.Created,
		Proposer:     objDB.Proposer.String,
		Rejector:     objDB.Rejector.String,
		ScheduleTime: settingChange.ScheduleTime,
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
	listApproval, err := s.GetLisApprovalSettingChange(uint64(id))
	if err != nil {
		s.l.Errorw("failed to get approval info of setting change", "err", err)
		return common.SettingChangeResponse{}, err
	}
	res.ListApproval = listApproval
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
		listApproval, err := s.GetLisApprovalSettingChange(uint64(rr.ID))
		if err != nil {
			s.l.Errorw("failed to get approval info of setting change", "err", err, "setting change id", rr.ID)
			return nil, err
		}
		rr.ListApproval = listApproval
		result = append(result, rr)
	}
	return result, nil
}

// RejectSettingChange delete setting change with a given id
func (s *Storage) RejectSettingChange(id rtypes.SettingChangeID, keyID string) error {
	var returnedID uint64
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer pgutil.RollbackUnlessCommitted(tx)
	err = tx.Stmtx(s.stmts.updateSettingChangeStatus).Get(&returnedID, id, common.ChangeStatusRejected.String(), common.StringPointer(keyID))
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

func (s *Storage) applyChange(tx *sqlx.Tx, i int, entry common.SettingChangeEntry, adr *common.AdditionalDataReturn) error {
	var err error
	switch e := entry.Data.(type) {
	case *common.ChangeAssetAddressEntry:
		err = s.changeAssetAddress(tx, e.ID, e.Address)
		if err != nil {
			s.l.Errorw("change asset address", "index", i, "err", err)
			return err
		}
	case *common.CreateAssetEntry:
		_, tradingPairIDs, err := s.createAsset(tx, e.Symbol, e.Name, e.Address, e.Decimals, e.Transferable, e.SetRate, e.Rebalance,
			e.IsQuote, e.IsEnabled, e.PWI, e.RebalanceQuadratic, e.Exchanges, e.Target, e.StableParam, e.FeedWeight,
			e.NormalUpdatePerPeriod, e.MaxImbalanceRatio, e.OrderDurationMillis, e.PriceETHAmount, e.ExchangeETHAmount, e.SanityInfo)
		if err != nil {
			s.l.Errorw("create asset", "index", i, "err", err)
			return err
		}
		adr.AddedTradingPairs = append(adr.AddedTradingPairs, tradingPairIDs...)
	case *common.CreateAssetExchangeEntry:
		_, tradingPairIDs, err := s.createAssetExchange(tx, e.ExchangeID, e.AssetID, e.Symbol, e.DepositAddress, e.MinDeposit,
			e.WithdrawFee, e.TargetRecommended, e.TargetRatio, e.TradingPairs)
		if err != nil {
			s.l.Errorw("create asset exchange", "index", i, "err", err)
			return err
		}
		adr.AddedTradingPairs = append(adr.AddedTradingPairs, tradingPairIDs...)
	case *common.CreateTradingPairEntry:
		tpID, err := s.createTradingPair(tx, e.ExchangeID, e.Base, e.Quote, e.PricePrecision, e.AmountPrecision, e.AmountLimitMin,
			e.AmountLimitMax, e.PriceLimitMin, e.PriceLimitMax, e.MinNotional, e.StaleThreshold, e.AssetID)
		if err != nil {
			s.l.Errorw("create trading pair", "index", i, "err", err)
			return err
		}
		adr.AddedTradingPairs = append(adr.AddedTradingPairs, tpID)
	case *common.UpdateAssetEntry:
		err = s.updateAsset(tx, e.AssetID, *e)
		if err != nil {
			s.l.Errorw("update asset", "index", i, "err", err)
			return err
		}
	case *common.UpdateAssetExchangeEntry:
		err = s.updateAssetExchange(tx, e.ID, *e)
		if err != nil {
			s.l.Errorw("update asset exchange", "index", i, "err", err)
			return err
		}
	case *common.UpdateExchangeEntry:
		err = s.updateExchange(tx, e.ExchangeID, *e)
		if err != nil {
			s.l.Errorw("update exchange", "index", i, "err", err)
			return err
		}
	case *common.DeleteAssetExchangeEntry:
		err = s.deleteAssetExchange(tx, e.AssetExchangeID)
		if err != nil {
			s.l.Errorw("delete asset exchange", "index", i, "err", err)
			return err
		}
	case *common.DeleteTradingPairEntry:
		err = s.deleteTradingPair(tx, e.TradingPairID)
		if err != nil {
			s.l.Errorw("delete trading pair", "index", i, "err", err)
			return err
		}
	case *common.UpdateStableTokenParamsEntry:
		err = s.updateStableTokenParams(tx, e.Params)
		if err != nil {
			s.l.Errorw("update stable token params", "index", i, "err", err)
			return err
		}
	case *common.SetFeedConfigurationEntry:
		err = s.setFeedConfiguration(tx, *e)
		if err != nil {
			s.l.Errorw("set feed configuration", "index", i, "err", err)
			return err
		}
	case *common.UpdateTradingPairEntry:
		err = s.updateTradingPair(tx, e.TradingPairID, storage.UpdateTradingPairOpts{
			StaleThreshold: e.StaleThreshold,
		})
		if err != nil {
			s.l.Errorw("update trading pair", "index", i, "err", err)
			return err
		}
	default:
		return fmt.Errorf("unexpected change object %+v", e)
	}
	return nil
}

// ConfirmSettingChange apply setting change with a given id
func (s *Storage) ConfirmSettingChange(id rtypes.SettingChangeID, commit bool) (*common.AdditionalDataReturn, error) {
	adr := &common.AdditionalDataReturn{
		AddedTradingPairs: []rtypes.TradingPairID{},
	}
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, errors.Wrap(err, "create transaction error")
	}
	defer pgutil.RollbackUnlessCommitted(tx)
	changeObj, err := s.getSettingChange(tx, id)
	if err != nil {
		return nil, errors.Wrap(err, "get setting change error")
	}

	for i, change := range changeObj.ChangeList {
		if err = s.applyChange(tx, i, change, adr); err != nil {
			return nil, err
		}
	}
	_, err = tx.Stmtx(s.stmts.updateSettingChangeStatus).Exec(id, common.ChangeStatusAccepted.String(), nil)
	if err != nil {
		return nil, err
	}
	if commit {
		if err := tx.Commit(); err != nil {
			s.l.Infow("setting change has been failed to confirm", "id", id, "err", err)
			return nil, err
		}
		s.l.Infow("setting change has been confirmed successfully", "id", id)
		return adr, nil
	}
	s.l.Infow("setting change will be reverted due commit flag not set", "id", id)
	return nil, nil
}
