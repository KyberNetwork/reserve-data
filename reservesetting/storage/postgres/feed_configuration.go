package postgres

import (
	"database/sql"
	"fmt"

	pgutil "github.com/KyberNetwork/reserve-data/common/postgres"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/world"
	"github.com/jmoiron/sqlx"
)

type setFeedConfigurationParams struct {
	Name                 string   `db:"name"`
	Enabled              *bool    `db:"enabled"`
	BaseVolatilitySpread *float64 `db:"base_volatility_spread"`
	NormalSpread         *float64 `db:"normal_spread"`
}

func (s *Storage) initFeedData() error {
	// init all feed as enabled
	query := `INSERT INTO "feed_configurations" (name, enabled, base_volatility_spread, normal_spread) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING;`
	for _, feed := range world.AllFeeds() {
		if _, err := s.db.Exec(query, feed, true, 0, 0); err != nil {
			return err
		}
	}
	return nil
}

// UpdateFeedStatus update feed status
func (s *Storage) UpdateFeedStatus(name string, enabled bool) error {
	return s.setFeedConfiguration(nil, common.SetFeedConfigurationEntry{
		Name:    name,
		Enabled: common.BoolPointer(enabled),
	})
}

func (s *Storage) setFeedConfiguration(tx *sqlx.Tx, feedConfiguration common.SetFeedConfigurationEntry) error {
	var sts = s.stmts.setFeedConfiguration
	if tx != nil {
		sts = tx.NamedStmt(s.stmts.setFeedConfiguration)
	}
	var feedName string
	err := sts.Get(&feedName, setFeedConfigurationParams{
		Name:                 feedConfiguration.Name,
		Enabled:              feedConfiguration.Enabled,
		BaseVolatilitySpread: feedConfiguration.BaseVolatilitySpread,
		NormalSpread:         feedConfiguration.NormalSpread,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return common.ErrExchangeNotExists
		}
		return fmt.Errorf("failed to set feed config, err=%s,", err)
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

	if err := tx.Stmtx(s.stmts.getFeedConfigurations).Select(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetFeedConfiguration return feed configuration by name
func (s *Storage) GetFeedConfiguration(name string) (common.FeedConfiguration, error) {
	var result common.FeedConfiguration
	tx, err := s.db.Beginx()
	if err != nil {
		return result, err
	}
	defer pgutil.RollbackUnlessCommitted(tx)

	if err := tx.Stmtx(s.stmts.getFeedConfiguration).Get(&result, name); err != nil {
		if err == sql.ErrNoRows {
			return result, common.ErrNotFound
		}
		return result, err
	}
	return result, nil
}
