package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	v3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

// Storage is an implementation of storage.Interface that use PostgreSQL as database system.
type Storage struct {
	db    *sqlx.DB
	l     *zap.SugaredLogger
	stmts *preparedStmts
}

func (s *Storage) initExchanges() error {
	const query = `INSERT INTO "exchanges" (id, name)
VALUES (unnest($1::INT[]),
        unnest($2::TEXT[])) ON CONFLICT(name) DO NOTHING;`

	var (
		idParams   []int
		nameParams []string
	)
	for name, ex := range common.ValidExchangeNames {
		nameParams = append(nameParams, name)
		idParams = append(idParams, int(ex))
	}

	_, err := s.db.Exec(query, pq.Array(idParams), pq.Array(nameParams))
	if err != nil {
		return err
	}

	return err
}

func (s *Storage) initAssets() error {
	var (
		defaultNormalUpdatePerPeriod float64 = 1
		defaultMaxImbalanceRatio     float64 = 2
		ethAddr                              = "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"
	)
	_, err := s.stmts.newAsset.Exec(&createAssetParams{
		Symbol:                "ETH",
		Name:                  "Ethereum",
		Address:               &ethAddr,
		Decimals:              18,
		Transferable:          true,
		SetRate:               v3.SetRateNotSet.String(),
		Rebalance:             false,
		IsQuote:               true,
		NormalUpdatePerPeriod: defaultNormalUpdatePerPeriod,
		MaxImbalanceRatio:     defaultMaxImbalanceRatio,
	})
	return err
}

// NewStorage creates a new Storage instance from given configuration.
func NewStorage(db *sqlx.DB) (*Storage, error) {
	l := zap.S()
	stmts, err := newPreparedStmts(db)
	if err != nil {
		return nil, fmt.Errorf("failed to preprare statements err=%s", err.Error())
	}

	s := &Storage{db: db, stmts: stmts, l: l}

	if err = s.initFeedData(); err != nil {
		return nil, fmt.Errorf("failed to init feed data, err=%s", err)
	}

	assets, err := s.GetAssets()
	if err != nil {
		return nil, fmt.Errorf("failed to get existing exchanges")
	}

	if err = s.initExchanges(); err != nil {
		return nil, fmt.Errorf("failed to initialize exchanges err=%s", err.Error())
	}

	if len(assets) == 0 {
		if err = s.initAssets(); err != nil {
			return nil, fmt.Errorf("failed to initialize assets err=%s", err.Error())
		}
	}
	return s, nil
}
