package postgres

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	// "github.com/lib/pq"
	// "github.com/pkg/errors"
	pgutil "github.com/KyberNetwork/reserve-data/common/postgres"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/world"
)

type setFeedConfigurationParams struct {
	Name                 string   `db:"name"`
	Enabled              *bool    `db:"enabled"`
	BaseVolatilitySpread *float64 `db:"base_volatility_spread"`
	NormalSpread         *float64 `db:"normal_spread"`
}

type feedConfigurationDB struct {
	Name                 sql.NullString  `db:"name"`
	Enabled              sql.NullBool    `db:"enabled"`
	BaseVolatilitySpread sql.NullFloat64 `db:"base_volatility_spread"`
	NormalSpread         sql.NullFloat64 `db:"normal_spread"`
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
	var resultDB feedConfigurationDB
	tx, err := s.db.Beginx()
	if err != nil {
		return common.FeedConfiguration{}, err
	}
	defer pgutil.RollbackUnlessCommitted(tx)

	if err := tx.Stmtx(s.stmts.getFeedConfiguration).Get(&resultDB, name); err != nil {
		if err == sql.ErrNoRows {
			return common.FeedConfiguration{}, common.ErrNotFound
		}
		return common.FeedConfiguration{}, err
	}
	return common.FeedConfiguration{
		Name:                 resultDB.Name.String,
		Enabled:              resultDB.Enabled.Bool,
		BaseVolatilitySpread: resultDB.BaseVolatilitySpread.Float64,
		NormalSpread:         resultDB.NormalSpread.Float64,
	}, nil
}
