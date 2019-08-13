package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/lib/pq"

	"github.com/KyberNetwork/reserve-data/v3/common"
)

const (
	constraintUniqueAssetAddress = "addresses_address_key"
)

// ConfirmChangeAssetAddress confirm the pending change asset address.
func (s *Storage) ConfirmChangeAssetAddress(id uint64) error {
	var (
		createChangeAssetAddress common.CreateChangeAssetAddress
		pendingObj               common.PendingObject
		err                      error
	)
	pendingObj, err = s.GetPendingObject(id, common.PendingTypeChangeAssetAddr)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(pendingObj.Data, &createChangeAssetAddress); err != nil {
		return err
	}
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer rollbackUnlessCommitted(tx)

	for _, a := range createChangeAssetAddress.Assets {
		_, err = tx.Stmtx(s.stmts.changeAssetAddress).Exec(a.ID, a.Address)
		if err != nil {
			pErr, ok := err.(*pq.Error)
			if !ok {
				return fmt.Errorf("unknown returned err=%s", err.Error())
			}
			log.Printf("failed to create new change asset address err=%s", pErr.Message)
			switch pErr.Code {
			case errAssertFailure:
				if err == sql.ErrNoRows {
					return common.ErrNotFound
				}
			case errCodeUniqueViolation:
				if pErr.Constraint == constraintUniqueAssetAddress {
					return common.ErrAddressExists
				}
			default:
				return err
			}
		}
	}
	_, err = tx.Stmtx(s.stmts.deletePendingObject).Exec(id, common.PendingTypeChangeAssetAddr.String())
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	log.Printf("confirm ChangeAssetAddress success with id=%d\n", id)
	return nil
}
