package postgres

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/pkg/errors"

	"github.com/KyberNetwork/reserve-data/v3/common"
)

// CreatePendingObject creates a pending obj in database and return id
func (s *Storage) CreatePendingObject(obj interface{}, pendingObjectType common.PendingObjectType) (uint64, error) {
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

	if err = tx.Stmtx(s.stmts.newPendingObject).Get(&id, jsonData, pendingObjectType.String()); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	log.Printf("create pending obj success with id=%d and type:%v\n", id, pendingObjectType.String())
	return id, nil
}

type pendingObjectDB struct {
	ID      uint64    `db:"id"`
	Created time.Time `db:"created"`
	Data    []byte    `db:"data"`
}

func (objDB pendingObjectDB) ToCommon() common.PendingObject {
	return common.PendingObject{
		Data:    objDB.Data,
		ID:      objDB.ID,
		Created: objDB.Created,
	}
}

func (s *Storage) GetPendingObject(id uint64, pendingObjectType common.PendingObjectType) (common.PendingObject, error) {
	var pendingObjs pendingObjectDB
	err := s.stmts.getPendingObject.Get(&pendingObjs, id, pendingObjectType.String())
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("%v not found in database id=%d\n", pendingObjectType.String(), id)
			return common.PendingObject{}, common.ErrNotFound
		}
		return common.PendingObject{}, err
	}
	return pendingObjs.ToCommon(), nil
}

func (s *Storage) GetPendingObjects(pendingObjectType common.PendingObjectType) ([]common.PendingObject, error) {
	var pendingObjs []pendingObjectDB
	err := s.stmts.getPendingObject.Select(&pendingObjs, nil, pendingObjectType.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	var result = make([]common.PendingObject, 0, 1) // although it's a slice, we expect only 1 for now.
	for _, p := range pendingObjs {
		result = append(result, p.ToCommon())
	}
	return result, nil
}

func (s *Storage) RejectPendingObject(id uint64, pendingObjectType common.PendingObjectType) error {
	var returnedID uint64
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer rollbackUnlessCommitted(tx)
	err = tx.Stmtx(s.stmts.deletePendingObject).Get(&returnedID, id, pendingObjectType.String())
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
	log.Printf("reject pending obj success with id=%d and type:%v\n", id, pendingObjectType.String())
	return nil
}
