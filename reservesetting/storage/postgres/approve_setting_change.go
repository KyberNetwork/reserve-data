package postgres

import (
	"database/sql"
	"fmt"

	pgutil "github.com/KyberNetwork/reserve-data/common/postgres"
	v3 "github.com/KyberNetwork/reserve-data/reservesetting/common"

	"github.com/pkg/errors"
)

// ApproveSettingChange ...
func (s *Storage) ApproveSettingChange(keyID string, settingChangeID uint64) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "create transaction error")
	}
	defer pgutil.RollbackUnlessCommitted(tx)
	var id int
	if err = tx.Stmtx(s.stmts.approveSettingChange).Get(&id, keyID, settingChangeID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		s.l.Infow("cannot approve for setting change", "settingChangeID", settingChangeID, "keyID", keyID, "err", err)
		return err
	}
	return nil
}

// GetLisApprovalSettingChange ...
func (s *Storage) GetLisApprovalSettingChange(settingChangeID uint64) ([]v3.ApprovalSettingChangeInfo, error) {
	var (
		result []v3.ApprovalSettingChangeInfo
	)
	err := s.stmts.getApprovalSettingChange.Select(&result, settingChangeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, v3.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get data, err=%s,", err)
	}
	return result, nil
}

// DispproveSettingChange ...
func (s *Storage) DispproveSettingChange(keyID string, settingChangeID uint64) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "create transaction error")
	}
	defer pgutil.RollbackUnlessCommitted(tx)
	_, err = tx.Stmtx(s.stmts.deleteApprovalSettingChange).Exec(&keyID, settingChangeID)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		s.l.Infow("cannot disapprove for setting change", "settingChangeID", settingChangeID, "keyID", keyID, "err", err)
		return err
	}
	return nil
}
