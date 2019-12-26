package postgres

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	// "github.com/lib/pq"
	// "github.com/pkg/errors"
	pgutil "github.com/KyberNetwork/reserve-data/common/postgres"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

func (s *Storage) setFeedConfiguration(tx *sqlx.Tx, feedConfiguration common.SetFeedConfigurationEntry) error {
	var feedName string
	err := s.stmts.setFeedConfiguration.Get(&feedName, feedConfiguration)
	if err != nil {
		if err == sql.ErrNoRows {
			return common.ErrExchangeNotExists
		}
		return fmt.Errorf("failed to set feed config, feed=%s err=%s,", feedName, err)
	}
	return nil
}

// GetFeedConfigurations return all feed configuration
func (s *Storage) GetFeedConfigurations() ([]common.FeedConfiguration, error) {
	var result []common.FeedConfiguration
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer pgutil.RollbackUnlessCommitted(tx)

	if err := tx.Stmtx(s.stmts.getAsset).Select(&result); err != nil {
		return nil, err
	}
	return result, nil
}
