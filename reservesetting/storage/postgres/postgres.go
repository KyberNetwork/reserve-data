package postgres

import (
	"fmt"
	"time"

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
		defaultOrderDurationMillis   uint64  = 20000
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
		OrderDurationMillis:   defaultOrderDurationMillis,
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

func generateFetchDataMonthlyPartition(t time.Time) string {
	nextMonth := t.AddDate(0, 1, 0)
	tblName := fmt.Sprintf("fetch_data_%s", t.Format("2006_01"))
	query := fmt.Sprintf("CREATE TABLE %s PARTITION OF fetch_data FOR VALUES from('%s') TO ('%s')",
		tblName, t.Format("2006-01-02"), nextMonth.Format("2006-01-02"))
	return query
}

func firstDayOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func isTableExistsError(err error) bool {
	if err == nil {
		return false
	}
	const tableExistCode = "42P07"
	if pe, ok := err.(*pq.Error); ok && pe.Code == tableExistCode {
		return true
	}
	return false
}

// MakeFetchDataTablePartition will create partitions for current and next month
func (s *Storage) MakeFetchDataTablePartition() error {
	fom := firstDayOfMonth(time.Now())
	query := generateFetchDataMonthlyPartition(fom) // fom need to be first day of month because we use it in partition values
	l := zap.S()
	_, err := s.db.Exec(query)
	// it's fine if current month partition exists
	if err != nil {
		if !isTableExistsError(err) {
			return err
		}
	} else {
		l.Infow("success create partition", "month", fom.Format("2006_01"))
	}
	nextMonth := fom.AddDate(0, 1, 0)
	nextMonthPartQuery := generateFetchDataMonthlyPartition(nextMonth)
	_, err = s.db.Exec(nextMonthPartQuery)
	if err == nil {
		l.Infow("success create partition", "month", nextMonth.Format("2006_01"))
		return nil
	}
	if isTableExistsError(err) {
		return nil
	}
	return err
}
func generateOrderBookPartition(t time.Time) string {
	nextDay := t.AddDate(0, 0, 1)
	tblName := fmt.Sprintf("order_book_data_%s", t.Format("2006_01_02"))
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s PARTITION OF order_book_data FOR VALUES from('%s') TO ('%s')",
		tblName, t.Format("2006-01-02"), nextDay.Format("2006-01-02"))
	return query
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// PrepareOrderBookTablePartition will create partitions for current and next days
func (s *Storage) PrepareOrderBookTablePartition() error {
	l := zap.S()
	t := startOfDay(time.Now())
	for i := 0; i < 3; i++ {
		query := generateOrderBookPartition(t) // fom need to be first day of month because we use it in partition values
		if _, err := s.db.Exec(query); err != nil {
			l.Errorw("failed to create partition", "query", query)
			return err
		}
		l.Debugw("successful to init partition", query)
		t = t.Add(time.Hour * 24)
	}
	return nil
}
