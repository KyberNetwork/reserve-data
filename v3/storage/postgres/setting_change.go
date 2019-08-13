package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/KyberNetwork/reserve-data/v3/common"
	"github.com/KyberNetwork/reserve-data/v3/storage"
)

// CreateSettingChange creates an object change in database and return id
func (s *Storage) CreateSettingChange(obj common.SettingChange) (uint64, error) {
	var id uint64
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to parse json data %+v", obj)
	}
	tx, err := s.db.Beginx()
	if err != nil {
		return 0, err
	}
	defer rollbackUnlessCommitted(tx)

	if err = tx.Stmtx(s.stmts.newSettingChange).Get(&id, jsonData); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	log.Printf("create obj change success with id=%d\n", id)
	return id, nil
}

type dbSettingChangeEntry struct {
	Type common.ChangeType `json:"type"`
	Data json.RawMessage   `json:"data"`
}

type changeListDB struct {
	ChangeList []dbSettingChangeEntry `json:"change_list"`
}
type settingChangeDB struct {
	ID      uint64    `db:"id"`
	Created time.Time `db:"created"`
	Data    []byte    `db:"data"`
}

func (objDB settingChangeDB) ToCommon() (common.SettingChangeResponse, error) {
	var dbResult changeListDB
	err := json.Unmarshal(objDB.Data, &dbResult)
	if err != nil {
		return common.SettingChangeResponse{}, err
	}
	res := make([]common.SettingChangeEntry, 0, len(dbResult.ChangeList))
	for _, o := range dbResult.ChangeList {
		i, err := common.SettingChangeFromType(o.Type)
		if err != nil {
			return common.SettingChangeResponse{}, err
		}
		if err = json.Unmarshal(o.Data, i); err != nil {
			return common.SettingChangeResponse{}, errors.Wrap(err, fmt.Sprintf("decode error for %+v", i))
		}
		res = append(res, common.SettingChangeEntry{
			Type: o.Type,
			Data: i,
		})
	}
	return common.SettingChangeResponse{
		ChangeList: res,
		ID:         objDB.ID,
		Created:    objDB.Created,
	}, nil
}

// GetSettingChange returns a object with a given id and type
func (s *Storage) GetSettingChange(id uint64) (common.SettingChangeResponse, error) {
	var dbResult settingChangeDB
	err := s.stmts.getSettingChange.Get(&dbResult, id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("object change not found in database id=%d\n", id)
			return common.SettingChangeResponse{}, common.ErrNotFound
		}
		return common.SettingChangeResponse{}, err
	}
	res, err := dbResult.ToCommon()
	if err != nil {
		return common.SettingChangeResponse{}, err
	}
	return res, nil
}

// GetSettingChanges return list object change.
func (s *Storage) GetSettingChanges() ([]common.SettingChangeResponse, error) {
	var dbResult []settingChangeDB
	err := s.stmts.getSettingChange.Select(&dbResult, nil)
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

// RejectSettingChange delete object change with a given id
func (s *Storage) RejectSettingChange(id uint64) error {
	var returnedID uint64
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer rollbackUnlessCommitted(tx)
	err = tx.Stmtx(s.stmts.deleteSettingChange).Get(&returnedID, id)
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
	log.Printf("reject object change success with id=%d\n", id)
	return nil
}
func (s *Storage) getSettingChange(tx *sqlx.Tx, id uint64) (common.SettingChangeResponse, error) {
	var dbResult settingChangeDB
	err := tx.Stmtx(s.stmts.getSettingChange).Get(&dbResult, id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("object change not found in database id=%d\n", id)
			return common.SettingChangeResponse{}, common.ErrNotFound
		}
		return common.SettingChangeResponse{}, err
	}
	res, err := dbResult.ToCommon()
	if err != nil {
		log.Printf("failed to unmarshal object, err=%v\n", err)
		return common.SettingChangeResponse{}, err
	}
	return res, nil
}
func (s *Storage) applyChange(tx *sqlx.Tx, i int, entry common.SettingChangeEntry) error {
	var err error
	switch entry.Type {
	case common.ChangeTypeChangeAssetAddr:
		u, ok := entry.Data.(*common.ChangeAssetAddressEntry)
		if !ok {
			msg := fmt.Sprintf("bad cast at %d to %s\n", i, common.ChangeTypeChangeAssetAddr)
			log.Println(msg)
			return errors.Wrap(err, msg)
		}
		err = s.changeAssetAddress(tx, u.ID, u.Address)
		if err != nil {
			msg := fmt.Sprintf("entry asset address %d, err=%v\n", i, err)
			log.Println(msg)
			return err
		}
	case common.ChangeTypeCreateAsset:
		a, ok := entry.Data.(*common.CreateAssetEntry)
		if !ok {
			msg := fmt.Sprintf("bad cast at %d to %s\n", i, common.ChangeTypeCreateAsset)
			log.Println(msg)
			return errors.Wrap(err, msg)
		}
		_, err = s.createAsset(tx, a.Symbol, a.Name, a.Address, a.Decimals, a.Transferable, a.SetRate, a.Rebalance,
			a.IsQuote, a.PWI, a.RebalanceQuadratic, a.Exchanges, a.Target)
		if err != nil {
			msg := fmt.Sprintf("create asset %d, err=%v\n", i, err)
			log.Println(msg)
			return err
		}
	case common.ChangeTypeCreateAssetExchange:
		a, ok := entry.Data.(*common.CreateAssetExchangeEntry)
		if !ok {
			msg := fmt.Sprintf("bad cast at %d to %s\n", i, common.ChangeTypeCreateAssetExchange)
			log.Println(msg)
			return errors.Wrap(err, msg)
		}
		_, err = s.createAssetExchange(tx, a.ExchangeID, a.AssetID, a.Symbol, a.DepositAddress, a.MinDeposit,
			a.WithdrawFee, a.TargetRecommended, a.TargetRatio)
		if err != nil {
			msg := fmt.Sprintf("create asset exchange %d, err=%v\n", i, err)
			log.Println(msg)
			return err
		}
	case common.ChangeTypeCreateTradingBy:
		a, ok := entry.Data.(*common.CreateTradingByEntry)
		if !ok {
			msg := fmt.Sprintf("bad cast at %d to %s\n", i, common.ChangeTypeCreateTradingBy)
			log.Println(msg)
			return errors.Wrap(err, msg)
		}
		_, err = s.createTradingBy(tx, a.AssetID, a.TradingPairID)
		if err != nil {
			msg := fmt.Sprintf("create trading by %d, err=%v\n", i, err)
			log.Println(msg)
			return err
		}
	case common.ChangeTypeCreateTradingPair:
		a, ok := entry.Data.(*common.CreateTradingPairEntry)
		if !ok {
			msg := fmt.Sprintf("bad cast at %d to %s\n", i, common.ChangeTypeCreateTradingPair)
			log.Println(msg)
			return errors.Wrap(err, msg)
		}
		_, err = s.createTradingPair(tx, a.ExchangeID, a.Base, a.Quote, a.PricePrecision, a.AmountPrecision, a.AmountLimitMin,
			a.AmountLimitMax, a.PriceLimitMin, a.PriceLimitMax, a.MinNotional)
		if err != nil {
			msg := fmt.Sprintf("create trading pair %d, err=%v\n", i, err)
			log.Println(msg)
			return err
		}
	case common.ChangeTypeUpdateAsset:
		a, ok := entry.Data.(*common.UpdateAssetEntry)
		if !ok {
			msg := fmt.Sprintf("bad cast at %d to %s\n", i, common.ChangeTypeUpdateAsset)
			log.Println(msg)
			return errors.Wrap(err, msg)
		}
		err = s.updateAsset(tx, a.AssetID, storage.UpdateAssetOpts{
			Symbol:             a.Symbol,
			Transferable:       a.Transferable,
			Address:            a.Address,
			IsQuote:            a.IsQuote,
			Rebalance:          a.Rebalance,
			SetRate:            a.SetRate,
			Decimals:           a.Decimals,
			Name:               a.Name,
			Target:             a.Target,
			PWI:                a.PWI,
			RebalanceQuadratic: a.RebalanceQuadratic,
		})
		if err != nil {
			msg := fmt.Sprintf("update asset %d, err=%v\n", i, err)
			log.Println(msg)
			return err
		}
	case common.ChangeTypeUpdateAssetExchange:
		a, ok := entry.Data.(*common.UpdateAssetExchangeEntry)
		if !ok {
			msg := fmt.Sprintf("bad cast at %d to %s\n", i, common.ChangeTypeUpdateAssetExchange)
			log.Println(msg)
			return errors.Wrap(err, msg)
		}
		err = s.updateAssetExchange(tx, a.ID, storage.UpdateAssetExchangeOpts{
			Symbol:            a.Symbol,
			DepositAddress:    a.DepositAddress,
			MinDeposit:        a.MinDeposit,
			WithdrawFee:       a.WithdrawFee,
			TargetRecommended: a.TargetRecommended,
			TargetRatio:       a.TargetRatio,
		})
		if err != nil {
			msg := fmt.Sprintf("bad cast at %d to %s\n", i, common.ChangeTypeUpdateAssetExchange)
			log.Println(msg)
			return err
		}
	case common.ChangeTypeUpdateExchange:
		a, ok := entry.Data.(*common.UpdateExchangeEntry)
		if !ok {
			msg := fmt.Sprintf("bad cast at %d to %s\n", i, common.ChangeTypeUpdateExchange)
			log.Println(msg)
			return errors.Wrap(err, msg)
		}
		err = s.updateExchange(tx, a.ExchangeID, storage.UpdateExchangeOpts{
			TradingFeeMaker: a.TradingFeeMaker,
			TradingFeeTaker: a.TradingFeeTaker,
			Disable:         a.Disable,
		})
		if err != nil {
			msg := fmt.Sprintf("update exchange %d, err=%v\n", i, err)
			log.Println(msg)
			return err
		}
	case common.ChangeTypeDeleteAssetExchange:
		// TODO: implement delete
	case common.ChangeTypeDeleteTradingBy:
	case common.ChangeTypeDeleteTradingPair:
	case common.ChangeTypeUnknown:
		return fmt.Errorf("change type not set at %d", i)
	}
	return nil
}

// ConfirmSettingChange apply object change with a given id
func (s *Storage) ConfirmSettingChange(id uint64, commit bool) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "create transaction error")
	}
	defer rollbackUnlessCommitted(tx)
	changeObj, err := s.getSettingChange(tx, id)
	if err != nil {
		return errors.Wrap(err, "get object change error")
	}

	for i, change := range changeObj.ChangeList {
		if err = s.applyChange(tx, i, change); err != nil {
			return err
		}
	}
	if commit {
		if err := tx.Commit(); err != nil {
			log.Printf("object change id=%d has been failed to confirm, err=%v\n", id, err)
			return err
		}
		log.Printf("object change has been confirmed successfully, id=%d\n", id)
		return nil
	}
	return nil
}
