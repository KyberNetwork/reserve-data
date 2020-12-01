package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/DataDog/zstd"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/lib/caller"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

const (
	fetchDataTable = "fetch_data" // data fetch from exchange and blockchain
	activityTable  = "activity"
	// data type constant

	authDataExpiredDuration uint64 = 10 * 86400000 // 10day in milisec
)

//go:generate enumer -type=fetchDataType -linecomment -json=true -sql=true
type fetchDataType int

const (
	_            fetchDataType = iota // price
	rateDataType                      // rate
	authDataType                      // auth_data
	goldDataType                      // gold
	btcDataType                       // btc
	usdDataType                       // usd
)

var (
	// ErrorNotFound record not found error
	ErrorNotFound = errors.New("record not found")
)

// PostgresStorage struct
type PostgresStorage struct {
	db *sqlx.DB
	l  *zap.SugaredLogger
}

// NewPostgresStorage return new db instance
func NewPostgresStorage(db *sqlx.DB) (*PostgresStorage, error) {
	s := &PostgresStorage{
		db: db,
		l:  zap.S(),
	}

	return s, nil
}

func getDataType(data interface{}) fetchDataType {
	switch data.(type) {
	case common.AuthDataSnapshot, *common.AuthDataSnapshot:
		return authDataType
	case common.BTCData, *common.BTCData:
		return btcDataType
	case common.GoldData, *common.GoldData:
		return goldDataType
	case common.USDData, *common.USDData:
		return usdDataType
	case common.AllRateEntry, *common.AllRateEntry:
		return rateDataType
	}
	panic(fmt.Sprintf("unexpected data type %+v", data))
}

func (ps *PostgresStorage) storeFetchData(data interface{}, timepoint uint64) error {
	query := fmt.Sprintf(`INSERT INTO "%s" (created, data, type) VALUES ($1, $2, $3)`, fetchDataTable)
	timestamp := common.MillisToTime(timepoint)
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}
	dataType := getDataType(data)
	if _, err := ps.db.Exec(query, timestamp, dataJSON, dataType); err != nil {
		return err
	}
	return nil
}

func (ps *PostgresStorage) storeOrderBookData(data common.AllPriceEntry, timepoint uint64) error {
	query := `INSERT INTO order_book_data (created, data) VALUES ($1, $2)`
	timestamp := common.MillisToTime(timepoint)
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if _, err := ps.db.Exec(query, timestamp, dataJSON); err != nil {
		return err
	}
	return nil
}
func (ps *PostgresStorage) currentOrderBookVersion(timepoint uint64) (common.Version, error) {
	var (
		v  common.Version
		ts time.Time
	)
	timestamp := common.MillisToTime(timepoint)
	query := `SELECT created
FROM (
	SELECT
		id,
		created
	FROM
		order_book_data
	WHERE
		created <= $1
	ORDER BY
		created DESC
	LIMIT 1) a1
WHERE
	$1 - created <= interval '150' second` // max duration block is 10 ~ 150 second
	if err := ps.db.Get(&ts, query, timestamp); err != nil {
		if err == sql.ErrNoRows {
			return v, fmt.Errorf("there is no version at timestamp: %d", timepoint)
		}
		return v, err
	}
	v = common.Version(common.TimeToMillis(ts))
	return v, nil
}

func (ps *PostgresStorage) currentVersion(dataType fetchDataType, timepoint uint64) (common.Version, error) {
	var (
		v  common.Version
		ts time.Time
	)
	timestamp := common.MillisToTime(timepoint)
	query := fmt.Sprintf(`SELECT created
FROM (
	SELECT
		id,
		created
	FROM
		"%s"
	WHERE
		created <= $1
		AND TYPE = $2
	ORDER BY
		created DESC
	LIMIT 1) a1
WHERE
	$1 - created <= interval '150' second`,
		fetchDataTable) // max duration block is 10 ~ 150 second
	if err := ps.db.Get(&ts, query, timestamp, dataType); err != nil {
		if err == sql.ErrNoRows {
			return v, fmt.Errorf("there is no version at timestamp: %d", timepoint)
		}
		return v, err
	}
	v = common.Version(common.TimeToMillis(ts))
	return v, nil
}

func (ps *PostgresStorage) getData(o interface{}, v common.Version) error {
	var (
		data   []byte
		logger = ps.l.With("func", caller.GetCurrentFunctionName())
	)
	ts := common.MillisToTime(uint64(v))
	dataType := getDataType(o)
	query := fmt.Sprintf(`SELECT data FROM "%s" WHERE created = $1 AND type = $2`, fetchDataTable)
	if err := ps.db.Get(&data, query, ts, dataType); err != nil {
		logger.Errorw("failed to get data from fetchData table", "error", err)
		return err
	}
	var err error
	if isZSTDPayload(data) {
		data, err = zstd.Decompress(nil, data)
		if err != nil {
			logger.Errorw("decompress error", "err", err)
			return err
		}
	}
	return json.Unmarshal(data, o)
}

func (ps *PostgresStorage) getOrderBookData(v common.Version) (common.AllPriceEntry, error) {
	var (
		data   []byte
		logger = ps.l.With("func", caller.GetCurrentFunctionName())
	)
	ts := common.MillisToTime(uint64(v))
	query := `SELECT data FROM order_book_data WHERE created = $1`
	if err := ps.db.Get(&data, query, ts); err != nil {
		logger.Errorw("failed to get data from order_book_data table", "error", err)
		return common.AllPriceEntry{}, err
	}
	var allPrices common.AllPriceEntry
	err := json.Unmarshal(data, &allPrices)
	if err != nil {
		logger.Errorw("unmarshal order book data failed", "err", err, "data", string(data))
	}
	return allPrices, err
}

// StorePrice store price
func (ps *PostgresStorage) StorePrice(priceEntry common.AllPriceEntry, timepoint uint64) error {
	return ps.storeOrderBookData(priceEntry, timepoint)
}

// CurrentPriceVersion return current price version
func (ps *PostgresStorage) CurrentPriceVersion(timepoint uint64) (common.Version, error) {
	return ps.currentOrderBookVersion(timepoint)
}

// GetAllPrices return all prices currently save in db
func (ps *PostgresStorage) GetAllPrices(v common.Version) (common.AllPriceEntry, error) {
	allPrices, err := ps.getOrderBookData(v)
	return allPrices, err
}

// GetOnePrice return one price
func (ps *PostgresStorage) GetOnePrice(pairID rtypes.TradingPairID, v common.Version) (common.OnePrice, error) {
	allPrices, err := ps.GetAllPrices(v)
	if err != nil {
		return common.OnePrice{}, err
	}
	onePrice, exist := allPrices.Data[pairID]
	if exist {
		return onePrice, nil
	}
	return common.OnePrice{}, errors.New("pair id does not exist")
}

// StoreAuthSnapshot store authdata
func (ps *PostgresStorage) StoreAuthSnapshot(authData *common.AuthDataSnapshot, timepoint uint64) error {
	return ps.storeFetchData(authData, timepoint)
}

// CurrentAuthDataVersion return current auth data version
func (ps *PostgresStorage) CurrentAuthDataVersion(timepoint uint64) (common.Version, error) {
	return ps.currentVersion(authDataType, timepoint)
}

// GetAuthData return auth data
func (ps *PostgresStorage) GetAuthData(v common.Version) (common.AuthDataSnapshot, error) {
	var (
		authData common.AuthDataSnapshot
	)
	err := ps.getData(&authData, v)
	return authData, err
}

// ExportExpiredAuthData export data to store on s3 storage
func (ps *PostgresStorage) ExportExpiredAuthData(timepoint uint64, filePath string) (uint64, error) {

	// create export file
	outFile, err := os.Create(filePath)
	if err != nil {
		return 0, err
	}
	defer func() {
		if cErr := outFile.Close(); cErr != nil {
			ps.l.Errorf("Close file error: %s", cErr.Error())
		}
	}()

	// Get expire data
	timepointExpireData := timepoint - authDataExpiredDuration
	timestampExpire := common.MillisToTime(timepointExpireData)
	query := fmt.Sprintf(`SELECT data FROM "%s" WHERE type = $1 AND created < $2`, fetchDataTable)
	rows, err := ps.db.Query(query, authDataType, timestampExpire)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	defer func() {
		if cle := rows.Close(); cle != nil {
			ps.l.Errorf("close result error %v", cle)
		}
	}()

	var count uint64
	for rows.Next() {
		var data []byte
		err = rows.Scan(&data)
		if err != nil {
			return 0, err
		}
		_, _ = outFile.Write(data)
		_, err = outFile.Write([]byte{'\n'})
		if err != nil {
			return 0, err
		}
		count++
	}
	return count, nil
}

// PruneExpiredAuthData remove expire auth data from database
func (ps *PostgresStorage) PruneExpiredAuthData(timepoint uint64) (uint64, error) {
	var (
		count uint64
	)
	// Get expire data
	timepointExpireData := timepoint - authDataExpiredDuration
	timestampExpire := common.MillisToTime(timepointExpireData)
	query := fmt.Sprintf(`WITH deleted AS 
	(DELETE FROM "%s" WHERE type = $1 AND created < $2 RETURNING *) SELECT count(*) FROM deleted`, fetchDataTable)
	if err := ps.db.Get(&count, query, authDataType, timestampExpire); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

// StoreRate store rate
func (ps *PostgresStorage) StoreRate(allRateEntry common.AllRateEntry, timepoint uint64) error {
	return ps.storeFetchData(allRateEntry, timepoint)
}

// CurrentRateVersion return current rate version
func (ps *PostgresStorage) CurrentRateVersion(timepoint uint64) (common.Version, error) {
	return ps.currentVersion(rateDataType, timepoint)
}

// GetRate return rate at a specific version
func (ps *PostgresStorage) GetRate(v common.Version) (common.AllRateEntry, error) {
	var (
		rate common.AllRateEntry
	)
	err := ps.getData(&rate, v)
	return rate, err
}
func isZSTDPayload(data []byte) bool {
	// zstd magic prefix 0xFD2FB528
	if len(data) > 4 && data[0] == 0x28 && data[1] == 0xB5 && data[2] == 0x2F && data[3] == 0xFD {
		return true
	}
	return false
}

// GetRates return rate from time to time
func (ps *PostgresStorage) GetRates(fromTime, toTime uint64) ([]common.AllRateEntry, error) {
	var (
		rates   []common.AllRateEntry
		allData [][]byte
		logger  = ps.l.With("func", caller.GetCallerFunctionName())
	)
	query := fmt.Sprintf(`SELECT data FROM "%s" WHERE type = $1 AND created >= $2 AND created <= $3`, fetchDataTable)
	from := common.MillisToTime(fromTime)
	to := common.MillisToTime(toTime)

	if err := ps.db.Select(&allData, query, rateDataType, from, to); err != nil {
		logger.Errorw("failed to get rates from database", "error", err)
		return nil, err
	}
	var err error
	var buff []byte
	for _, data := range allData {
		var rate common.AllRateEntry
		if isZSTDPayload(data) {
			data, err = zstd.Decompress(buff, data)
			if err != nil {
				logger.Errorw("decompress error", "err", err)
				return nil, err
			}
			buff = data // keep buff refer to recent decompressed buffer as it can reuse for next time
		}
		if err := json.Unmarshal(data, &rate); err != nil {
			logger.Errorw("failed to unmarshal rates", "error", err)
			return nil, err
		}
		rates = append(rates, rate)
	}
	return rates, nil
}

var (
	allowedActions = map[string]struct{}{
		common.ActionTrade:         {},
		common.ActionSetRate:       {},
		common.ActionCancelSetRate: {},
		common.ActionDeposit:       {},
		common.ActionWithdraw:      {},
	}
)

// GetAllRecords return all activities records from database
func (ps *PostgresStorage) GetAllRecords(fromTime, toTime uint64, actions []string) ([]common.ActivityRecord, error) {
	var (
		activities = []common.ActivityRecord{}
		data       [][]byte
	)

	query := fmt.Sprintf(`SELECT data FROM "%s" WHERE data->>'action' IN (?) AND created >= ? AND created <= ?`, activityTable)
	if len(actions) == 0 {
		actions = []string{common.ActionWithdraw, common.ActionDeposit, common.ActionTrade} // adjust default behavior to list all activity exclude set_rate
	}
	if len(actions) > len(allowedActions) {
		actions = actions[:len(allowedActions)]
	}
	from := common.MillisToTime(fromTime)
	to := common.MillisToTime(toTime)
	query, args, err := sqlx.In(query, actions, from, to)
	if err != nil {
		return activities, err
	}
	query = ps.db.Rebind(query)

	if err := ps.db.Select(&data, query, args...); err != nil {
		return activities, err
	}
	for _, dataByte := range data {
		var activity common.ActivityRecord
		if err := json.Unmarshal(dataByte, &activity); err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}
	return activities, nil
}

type assetRateTriggerDB struct {
	AssetID rtypes.AssetID `db:"asset_id"`
	Count   int            `db:"count"`
}

// GetAssetRateTriggers return a list of AssetRateTrigger, which calculate number of time an asset was record for setRate with trigger=true
func (ps *PostgresStorage) GetAssetRateTriggers(fromTime uint64, toTime uint64) ([]common.AssetRateTrigger, error) {
	from := common.MillisToTime(fromTime)
	to := common.MillisToTime(toTime)
	var res []assetRateTriggerDB
	query := `WITH sr AS(
	SELECT * FROM activity WHERE created BETWEEN $1 AND $2 AND data->>'action'='set_rates' AND data->'result'->>'tx' != '0x0000000000000000000000000000000000000000000000000000000000000000' 
)
SELECT id AS asset_id, (SELECT count(*) FROM (
	SELECT (data->'params'->'triggers'->(
			SELECT array_position(TRANSLATE((data->'params'->'assets')::text,'[]','{}')::integer[], assets.id)-1 as idx 
			FROM sr WHERE ID = p1.ID
		)
	)::BOOLEAN AS trigg FROM sr p1
) p2 WHERE trigg IS TRUE) AS count FROM assets`
	err := ps.db.Select(&res, query, from, to)
	if err != nil {
		return nil, err
	}
	var cres = make([]common.AssetRateTrigger, 0, len(res))
	for _, r := range res {
		cres = append(cres, common.AssetRateTrigger{
			AssetID: r.AssetID,
			Count:   r.Count,
		})
	}
	return cres, nil
}

// GetPendingActivities return all pending activities
func (ps *PostgresStorage) GetPendingActivities() ([]common.ActivityRecord, error) {
	var (
		pendingActivities []common.ActivityRecord
		data              [][]byte
	)
	query := fmt.Sprintf(`SELECT data FROM "%s" WHERE is_pending IS TRUE`, activityTable)
	if err := ps.db.Select(&data, query); err != nil {
		return []common.ActivityRecord{}, err
	}
	for _, dataByte := range data {
		var activity common.ActivityRecord
		if err := json.Unmarshal(dataByte, &activity); err != nil {
			return []common.ActivityRecord{}, err
		}
		pendingActivities = append(pendingActivities, activity)
	}
	return pendingActivities, nil
}

// UpdateActivity update activity to finished if it is finished
func (ps *PostgresStorage) UpdateActivity(id common.ActivityID, act common.ActivityRecord) error {
	var (
		data []byte
	)
	// get activity from db
	getQuery := fmt.Sprintf(`SELECT data FROM "%s" WHERE timepoint = $1 AND eid = $2`, activityTable)
	if err := ps.db.Get(&data, getQuery, id.Timepoint, id.EID); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	// check if activity is not pending anymore update it
	updateQuery := fmt.Sprintf(`UPDATE "%s" SET is_pending = $1, data = $2 WHERE timepoint = $3 AND eid = $4`, activityTable)
	dataBytes, err := json.Marshal(act)
	if err != nil {
		return err
	}
	isPending := act.IsPending()
	if _, err := ps.db.Exec(updateQuery, isPending, dataBytes, id.Timepoint, id.EID); err != nil {
		return err
	}
	return nil
}

// GetActivity return activity record by id
func (ps *PostgresStorage) GetActivity(exchangeID rtypes.ExchangeID, id string) (common.ActivityRecord, error) {
	var (
		activityRecord common.ActivityRecord
		data           []byte
	)
	query := fmt.Sprintf(`SELECT data FROM "%s" WHERE data->>'destination' = $1 AND eid = $2`, activityTable)
	if err := ps.db.Get(&data, query, exchangeID.String(), id); err != nil {
		if err == sql.ErrNoRows {
			return common.ActivityRecord{}, ErrorNotFound
		}
		return common.ActivityRecord{}, err
	}
	if err := json.Unmarshal(data, &activityRecord); err != nil {
		return common.ActivityRecord{}, err
	}
	return activityRecord, nil
}
func getFirstAndCountPendingAction(
	l *zap.SugaredLogger,
	pendings []common.ActivityRecord,
	minedNonce uint64, activityType string) (*common.ActivityRecord, uint64, error) {
	var minNonce uint64 = math.MaxUint64
	var minPrice uint64 = math.MaxUint64
	var result *common.ActivityRecord
	var count uint64
	for i, act := range pendings {
		if act.Action == activityType {
			l.Infof("looking for pending (%s): %+v", activityType, act)
			avtNonce := act.Result.Nonce
			if avtNonce < minedNonce {
				l.Infof("NONCE_ISSUE: stalled pending %s transaction, pending: %d, mined: %d",
					activityType, avtNonce, minedNonce)
				continue
			} else if avtNonce-minedNonce > 1 {
				l.Infof("NONCE_ISSUE: pending %s transaction for inconsecutive nonce, mined nonce: %d, request nonce: %d",
					activityType, minedNonce, avtNonce)
			}
			// find smallest nonce that > minedNonce, count number of activity has this same nonce.
			gasPrice, err := strconv.ParseUint(act.Result.GasPrice, 10, 64)
			if err != nil {
				return nil, 0, err
			}
			if avtNonce == minNonce {
				if gasPrice < minPrice {
					minNonce = avtNonce
					result = &pendings[i]
					minPrice = gasPrice
				}
				count++
			} else if avtNonce < minNonce {
				minNonce = avtNonce
				result = &pendings[i]
				minPrice = gasPrice
				count = 1
			}
		}
	}

	if result == nil {
		l.Infof("NONCE_ISSUE: found no pending %s transaction with nonce newer than equal to mined nonce: %d",
			activityType, minedNonce)
	} else {
		l.Infof("NONCE_ISSUE: un-mined pending %s, nonce: %d, count: %d, mined nonce: %d",
			activityType, result.Result.Nonce, count, minedNonce)
	}

	return result, count, nil
}

// PendingActivityForAction return pending set rate activity
func (ps *PostgresStorage) PendingActivityForAction(minedNonce uint64, activityType string) (*common.ActivityRecord, uint64, error) {
	pendings, err := ps.GetPendingActivities()
	if err != nil {
		return nil, 0, err
	}
	return getFirstAndCountPendingAction(ps.l, pendings, minedNonce, activityType)
}

// HasPendingDeposit return true if there is any pending deposit for a token
func (ps *PostgresStorage) HasPendingDeposit(token commonv3.Asset, exchange common.Exchange) (bool, error) {
	var (
		pendingActivity common.ActivityRecord
		data            [][]byte
	)
	query := fmt.Sprintf(`SELECT data FROM "%s" WHERE is_pending IS TRUE AND data->>'action' = $1 AND data ->> 'destination' = $2`, activityTable)
	if err := ps.db.Select(&data, query, common.ActionDeposit, exchange.ID().String()); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	for _, activity := range data {
		if err := json.Unmarshal(activity, &pendingActivity); err != nil {
			return false, err
		}
		if pendingActivity.Params.Asset == token.ID {
			return true, nil
		}
	}
	return false, nil
}

// MaxPendingNonce return biggest nonce in pending activity for an action
func (ps *PostgresStorage) MaxPendingNonce(action string) (int64, error) {
	var v int64
	if err := ps.db.Get(&v, "SELECT MAX(data->'result'->>'nonce') FROM activity WHERE is_pending IS TRUE AND data->>'action' = $1", action); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return v, nil
}

// Record save activity
func (ps *PostgresStorage) Record(action string, id common.ActivityID, destination string,
	params common.ActivityParams, result common.ActivityResult,
	estatus string, mstatus string, timepoint uint64, isPending bool) error {
	record := common.NewActivityRecord(
		action,
		id,
		destination,
		params,
		result,
		estatus,
		mstatus,
		common.Timestamp(strconv.FormatUint(timepoint, 10)),
	)
	query := fmt.Sprintf(`INSERT INTO "%s" (created, data, is_pending, timepoint, eid) VALUES($1, $2, $3, $4, $5)`, activityTable)
	timestamp := common.MillisToTime(timepoint)
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	if _, err := ps.db.Exec(query, timestamp, data, isPending, id.Timepoint, id.EID); err != nil {
		return err
	}
	return nil
}

// StoreGoldInfo store gold info into database
func (ps *PostgresStorage) StoreGoldInfo(goldData common.GoldData) error {
	timepoint := goldData.Timestamp
	return ps.storeFetchData(goldData, timepoint)
}

// StoreBTCInfo store btc info into database
func (ps *PostgresStorage) StoreBTCInfo(btcData common.BTCData) error {
	timepoint := btcData.Timestamp
	return ps.storeFetchData(btcData, timepoint)
}

// StoreUSDInfo store btc info into database
func (ps *PostgresStorage) StoreUSDInfo(usdData common.USDData) error {
	timepoint := usdData.Timestamp
	return ps.storeFetchData(usdData, timepoint)
}

// GetGoldInfo return gold info
func (ps *PostgresStorage) GetGoldInfo(v common.Version) (common.GoldData, error) {
	var (
		goldData common.GoldData
	)
	err := ps.getData(&goldData, v)
	return goldData, err
}

// GetBTCInfo return BTC info
func (ps *PostgresStorage) GetBTCInfo(v common.Version) (common.BTCData, error) {
	var (
		btcData common.BTCData
	)
	err := ps.getData(&btcData, v)
	return btcData, err
}

// GetUSDInfo return USD info
func (ps *PostgresStorage) GetUSDInfo(v common.Version) (common.USDData, error) {
	var (
		usdData common.USDData
	)
	err := ps.getData(&usdData, v)
	return usdData, err
}

// CurrentGoldInfoVersion return btc info version
func (ps *PostgresStorage) CurrentGoldInfoVersion(timepoint uint64) (common.Version, error) {
	return ps.currentVersion(goldDataType, timepoint)
}

// CurrentBTCInfoVersion return current btc info version
func (ps *PostgresStorage) CurrentBTCInfoVersion(timepoint uint64) (common.Version, error) {
	return ps.currentVersion(btcDataType, timepoint)
}

// CurrentUSDInfoVersion return current btc info version
func (ps *PostgresStorage) CurrentUSDInfoVersion(timepoint uint64) (common.Version, error) {
	return ps.currentVersion(usdDataType, timepoint)
}
