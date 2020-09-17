package common

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"

	"github.com/KyberNetwork/reserve-data/common/archive"
	rtypes "github.com/KyberNetwork/reserve-data/lib/rtypes"
)

// Version indicate fetched data version
type Version uint64

// Timestamp is data fetched timestamp
type Timestamp string

// Millis return uint64 of timestamp
func (ts Timestamp) Millis() uint64 {
	res, err := strconv.ParseUint(string(ts), 10, 64)
	//  this should never happen. Timestamp is never manually entered.
	if err != nil {
		panic(err)
	}
	return res
}

// GetTimestamp return timestamp
func GetTimestamp() Timestamp {
	return Timestamp(strconv.FormatUint(NowInMillis(), 10))
}

// NowInMillis return time by millisecond
func NowInMillis() uint64 {
	return TimeToMillis(time.Now())
}

// TimeToMillis convert timestamp from time.Time to milliseoncd (timepoint)
func TimeToMillis(t time.Time) uint64 {
	timestamp := t.UnixNano() / int64(time.Millisecond)
	return uint64(timestamp)
}

// MillisToTime convert timepoint (millisecond) to time.Time
func MillisToTime(t uint64) time.Time {
	return time.Unix(0, int64(t)*int64(time.Millisecond))
}

// ExchangeAddresses type store a map[tokenID]exchangeDepositAddress
type ExchangeAddresses map[string]ethereum.Address

func NewExchangeAddresses() *ExchangeAddresses {
	exAddr := make(ExchangeAddresses)
	return &exAddr
}

func (ea ExchangeAddresses) Remove(tokenID string) {
	delete(ea, tokenID)
}

func (ea ExchangeAddresses) Update(tokenID string, address ethereum.Address) {
	ea[tokenID] = address
}

func (ea ExchangeAddresses) Get(tokenID string) (ethereum.Address, bool) {
	address, supported := ea[tokenID]
	return address, supported
}

func (ea ExchangeAddresses) GetData() map[string]ethereum.Address {
	dataCopy := map[string]ethereum.Address{}
	for k, v := range ea {
		dataCopy[k] = v
	}
	return dataCopy
}

// ExchangePrecisionLimit store the precision and limit of a certain token pair on an exchange
// it is int the struct of [[int int], [float64 float64], [float64 float64], float64]
type ExchangePrecisionLimit struct {
	Precision   TokenPairPrecision   `json:"precision"`
	AmountLimit TokenPairAmountLimit `json:"amount_limit"`
	PriceLimit  TokenPairPriceLimit  `json:"price_limit"`
	MinNotional float64              `json:"min_notional"`
}

// ExchangeInfo is written and read concurrently
type ExchangeInfo map[rtypes.TradingPairID]ExchangePrecisionLimit

func NewExchangeInfo() ExchangeInfo {
	return ExchangeInfo(make(map[rtypes.TradingPairID]ExchangePrecisionLimit))
}

func (ei ExchangeInfo) Get(pair rtypes.TradingPairID) (ExchangePrecisionLimit, error) {
	info, exist := ei[pair]
	if !exist {
		return info, fmt.Errorf("token pair is not existed")
	}
	return info, nil

}

func (ei ExchangeInfo) GetData() map[rtypes.TradingPairID]ExchangePrecisionLimit {
	data := map[rtypes.TradingPairID]ExchangePrecisionLimit(ei)
	return data
}

//TokenPairPrecision represent precision when trading a token pair
type TokenPairPrecision struct {
	Amount int `json:"amount"`
	Price  int `json:"price"`
}

//TokenPairAmountLimit represent amount min and max when trade a token pair
type TokenPairAmountLimit struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type TokenPairPriceLimit struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type TradingFee map[string]float64

type FundingFee struct {
	Withdraw map[string]float64
	Deposit  map[string]float64
}

func (ff FundingFee) GetTokenFee(token string) float64 {
	withdrawFee := ff.Withdraw
	return withdrawFee[token]
}

type ExchangesMinDeposit map[string]float64

//ExchangeFees contains the fee for an exchanges
//It follow the struct of {trading: map[tokenID]float64, funding: {Withdraw: map[tokenID]float64, Deposit: map[tokenID]float64}}
type ExchangeFees struct {
	Trading TradingFee
	Funding FundingFee
}

func NewExchangeFee(tradingFee TradingFee, fundingFee FundingFee) ExchangeFees {
	return ExchangeFees{
		Trading: tradingFee,
		Funding: fundingFee,
	}
}

// NewFundingFee creates a new instance of FundingFee instance.
func NewFundingFee(widthraw, deposit map[string]float64) FundingFee {
	return FundingFee{
		Withdraw: widthraw,
		Deposit:  deposit,
	}
}

// ActivityID specify an activity
type ActivityID struct {
	Timepoint uint64
	EID       string
}

// ToBytes convert activity id into bytes
func (ai ActivityID) ToBytes() [64]byte {
	var b [64]byte
	temp := make([]byte, 64)
	binary.BigEndian.PutUint64(temp, ai.Timepoint)
	temp = append(temp, []byte(ai.EID)...)
	copy(b[0:], temp)
	return b
}

// MarshalText return activty into byte
func (ai ActivityID) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%s|%s", strconv.FormatUint(ai.Timepoint, 10), ai.EID)), nil
}

// UnmarshalText unmarshal activity from byte to activity id
func (ai *ActivityID) UnmarshalText(b []byte) error {
	id, err := StringToActivityID(string(b))
	if err != nil {
		return err
	}
	ai.Timepoint = id.Timepoint
	ai.EID = id.EID
	return nil
}

// String convert from activity id to string
func (ai ActivityID) String() string {
	res, _ := ai.MarshalText()
	return string(res)
}

// StringToActivityID convert from string to activity id
func StringToActivityID(id string) (ActivityID, error) {
	result := ActivityID{}
	parts := strings.Split(id, "|")
	if len(parts) < 2 {
		return result, fmt.Errorf("invalid activity id")
	}
	timeStr := parts[0]
	eid := strings.Join(parts[1:], "|")
	timepoint, err := strconv.ParseUint(timeStr, 10, 64)
	if err != nil {
		return result, err
	}
	result.Timepoint = timepoint
	result.EID = eid
	return result, nil
}

// NewActivityID creates new Activity ID.
func NewActivityID(timepoint uint64, eid string) ActivityID {
	return ActivityID{
		Timepoint: timepoint,
		EID:       eid,
	}
}

// CancelOrderResult is response when calling cancel an order
type CancelOrderResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// ActivityRecord object
type ActivityRecord struct {
	Action         string          `json:"action,omitempty"`
	ID             ActivityID      `json:"id,omitempty"`
	EID            string          `json:"eid"`
	Destination    string          `json:"destination,omitempty"`
	Params         *ActivityParams `json:"params,omitempty"`
	Result         *ActivityResult `json:"result,omitempty"`
	ExchangeStatus string          `json:"exchange_status,omitempty"`
	MiningStatus   string          `json:"mining_status,omitempty"`
	Timestamp      Timestamp       `json:"timestamp,omitempty"`
}

// AssetRateTrigger keep result of calculate token rate trigger
type AssetRateTrigger struct {
	AssetID rtypes.AssetID `json:"asset_id"`
	Count   int            `json:"count"`
}

// ActivityParams is params for activity
type ActivityParams struct {
	// deposit, withdraw params
	Exchange  rtypes.ExchangeID `json:"exchange,omitempty"`
	Asset     rtypes.AssetID    `json:"asset,omitempty"`
	Amount    float64           `json:"amount,omitempty"`
	Timepoint uint64            `json:"timepoint,omitempty"`
	// SetRates params
	Assets []rtypes.AssetID `json:"assets,omitempty"` // list of asset id
	Buys   []*big.Int       `json:"buys,omitempty"`
	Sells  []*big.Int       `json:"sells,omitempty"`
	Block  *big.Int         `json:"block,omitempty"`
	AFPMid []*big.Int       `json:"afpMid,omitempty"`
	Msgs   []string         `json:"msgs,omitempty"`
	// Trade params
	Type     string  `json:"type,omitempty"`
	Base     string  `json:"base,omitempty"`
	Quote    string  `json:"quote,omitempty"`
	Rate     float64 `json:"rate,omitempty"`
	Triggers []bool  `json:"triggers,omitempty"`
}

// ActivityResult is result of an activity
type ActivityResult struct {
	Tx       string `json:"tx,omitempty"`
	Nonce    uint64 `json:"nonce,omitempty"`
	GasPrice string `json:"gasPrice,omitempty"`
	Error    string `json:"error,omitempty"`
	// ID of withdraw
	ID string `json:"id,omitempty"`
	// params of trade
	Done      float64 `json:"done,omitempty"`
	Remaining float64 `json:"remaining,omitempty"`
	Finished  bool    `json:"finished,omitempty"`
	//
	StatusError string `json:"status_error,omitempty"`
	BlockNumber uint64 `json:"blockNumber,omitempty"`
	//
	WithdrawFee float64 `json:"withdraw_fee,omitempty"`
}

//NewActivityRecord return an activity record
func NewActivityRecord(action string, id ActivityID, destination string, params ActivityParams, result ActivityResult, exStatus, miStatus string, timestamp Timestamp) ActivityRecord {
	return ActivityRecord{
		Action:         action,
		ID:             id,
		Destination:    destination,
		Params:         &params,
		Result:         &result,
		ExchangeStatus: exStatus,
		MiningStatus:   miStatus,
		Timestamp:      timestamp,
	}
}

// IsExchangePending return true if activity is pending on centralized exchange
func (ar ActivityRecord) IsExchangePending() bool {
	switch ar.Action {
	case ActionWithdraw:
		return (ar.ExchangeStatus == "" || ar.ExchangeStatus == ExchangeStatusSubmitted) &&
			ar.MiningStatus != MiningStatusFailed
	case ActionDeposit:
		return (ar.ExchangeStatus == "" || ar.ExchangeStatus == ExchangeStatusPending) &&
			ar.MiningStatus != MiningStatusFailed
	case ActionTrade:
		return ar.ExchangeStatus == "" || ar.ExchangeStatus == ExchangeStatusSubmitted
	}
	return true
}

// IsBlockchainPending return true if activity is pending on blockchain
func (ar ActivityRecord) IsBlockchainPending() bool {
	switch ar.Action {
	case ActionWithdraw, ActionDeposit, ActionSetRate:
		return (ar.MiningStatus == "" || ar.MiningStatus == MiningStatusSubmitted) && ar.ExchangeStatus != ExchangeStatusFailed
	}
	return true
}

// IsPending return true if activity is pending
func (ar ActivityRecord) IsPending() bool {
	switch ar.Action {
	case ActionWithdraw:
		return (ar.ExchangeStatus == "" || ar.ExchangeStatus == ExchangeStatusSubmitted ||
			ar.MiningStatus == "" || ar.MiningStatus == MiningStatusSubmitted) &&
			ar.MiningStatus != MiningStatusFailed && ar.ExchangeStatus != ExchangeStatusFailed
	case ActionDeposit:
		return (ar.ExchangeStatus == "" || ar.ExchangeStatus == ExchangeStatusPending ||
			ar.MiningStatus == "" || ar.MiningStatus == MiningStatusSubmitted) &&
			ar.MiningStatus != MiningStatusFailed && ar.ExchangeStatus != ExchangeStatusFailed
	case ActionTrade:
		return (ar.ExchangeStatus == "" || ar.ExchangeStatus == ExchangeStatusSubmitted) &&
			ar.ExchangeStatus != ExchangeStatusFailed
	case ActionSetRate, ActionCancelSetRate:
		return (ar.MiningStatus == "" || ar.MiningStatus == MiningStatusSubmitted) &&
			ar.ExchangeStatus != ExchangeStatusFailed
	}
	return true
}

// ActivityStatus is status of an activity
type ActivityStatus struct {
	ExchangeStatus string
	Tx             string
	BlockNumber    uint64
	MiningStatus   string
	WithdrawFee    float64
	Error          error
}

// NewActivityStatus creates a new ActivityStatus instance.
func NewActivityStatus(exchangeStatus, tx string, blockNumber uint64, miningStatus string, withdrawFee float64, err error) ActivityStatus {
	return ActivityStatus{
		ExchangeStatus: exchangeStatus,
		Tx:             tx,
		BlockNumber:    blockNumber,
		MiningStatus:   miningStatus,
		WithdrawFee:    withdrawFee,
		Error:          err,
	}
}

type PriceEntry struct {
	Quantity float64 `json:"quantity"`
	Rate     float64 `json:"rate"`
}

// NewPriceEntry creates new instance of PriceEntry.
func NewPriceEntry(quantity, rate float64) PriceEntry {
	return PriceEntry{
		Quantity: quantity,
		Rate:     rate,
	}
}

type AllPriceEntry struct {
	Block uint64
	Data  map[rtypes.TradingPairID]OnePrice
}

type AllPriceResponse struct {
	Version    Version
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       map[rtypes.TradingPairID]OnePrice
	Block      uint64
}

type OnePriceResponse struct {
	Version    Version
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       OnePrice
	Block      uint64
}

type OnePrice map[rtypes.ExchangeID]ExchangePrice

type ExchangePrice struct {
	Valid      bool
	Error      string
	Timestamp  Timestamp
	Bids       []PriceEntry
	Asks       []PriceEntry
	ReturnTime Timestamp
}

type RawBalance big.Int

func (rb *RawBalance) ToFloat(decimal int64) float64 {
	return BigToFloat((*big.Int)(rb), decimal)
}

func (rb RawBalance) MarshalJSON() ([]byte, error) {
	selfInt := (big.Int)(rb)
	return selfInt.MarshalJSON()
}

func (rb *RawBalance) UnmarshalJSON(text []byte) error {
	selfInt := (*big.Int)(rb)
	return selfInt.UnmarshalJSON(text)
}

type BalanceEntry struct {
	Valid      bool       `json:"valid"`
	Error      string     `json:"error"`
	Timestamp  Timestamp  `json:"timestamp"`
	ReturnTime Timestamp  `json:"return_time"`
	Balance    RawBalance `json:"balance"`
}

func (be BalanceEntry) ToBalanceResponse(decimal int64) BalanceResponse {
	return BalanceResponse{
		Valid:      be.Valid,
		Error:      be.Error,
		Timestamp:  be.Timestamp,
		ReturnTime: be.ReturnTime,
		Balance:    be.Balance.ToFloat(decimal),
	}
}

// BalanceResponse represent response authdata for reserve balance
type BalanceResponse struct {
	Valid      bool      `json:"valid"`
	Error      string    `json:"error"`
	Timestamp  Timestamp `json:"timestamp"`
	ReturnTime Timestamp `json:"return_time"`
	Balance    float64   `json:"balance"`
}

type AllBalanceResponse struct {
	Version    Version
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       map[string]BalanceResponse
}

type RequestOrder struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
}

// Order accross multiple exchanges
type Order struct {
	ID            string               `json:"id,omitempty"` // standard id across multiple exchanges
	Symbol        string               `json:"symbol,omitempty"`
	Base          string               `json:"base"`
	Quote         string               `json:"quote"`
	OrderID       string               `json:"order_id"`
	Price         float64              `json:"price"`
	OrigQty       float64              `json:"orig_qty"`     // original quantity
	ExecutedQty   float64              `json:"executed_qty"` // matched quantity
	TimeInForce   string               `json:"time_in_force,omitempty"`
	Type          string               `json:"type"` // market or limit
	Side          string               `json:"side"` // buy or sell
	StopPrice     string               `json:"stop_price,omitempty"`
	IcebergQty    string               `json:"iceberg_qty,omitempty"`
	Time          uint64               `json:"time,omitempty"`
	TradingPairID rtypes.TradingPairID `json:"trading_pair_id,omitempty"`
}

type OrderEntry struct {
	Valid      bool
	Error      string
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       []Order
}

type AllOrderEntry map[rtypes.ExchangeID]OrderEntry

type AllOrderResponse struct {
	Version    Version
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       AllOrderEntry
}
type EBalanceEntry struct {
	Valid            bool
	Error            string
	Timestamp        Timestamp
	ReturnTime       Timestamp
	AvailableBalance map[rtypes.AssetID]float64
	LockedBalance    map[rtypes.AssetID]float64
	DepositBalance   map[rtypes.AssetID]float64
	Status           bool
}

type AllEBalanceResponse struct {
	Version    Version
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       map[rtypes.ExchangeID]EBalanceEntry
}

// AuthDataSnapshot object
type AuthDataSnapshot struct {
	Valid             bool                                `json:"valid,omitempty"`
	Error             string                              `json:"error,omitempty"`
	Timestamp         Timestamp                           `json:"timestamp,omitempty"`
	ReturnTime        Timestamp                           `json:"return_time,omitempty"`
	ExchangeBalances  map[rtypes.ExchangeID]EBalanceEntry `json:"exchange_balances,omitempty"`
	ReserveBalances   map[rtypes.AssetID]BalanceEntry     `json:"reserve_balances,omitempty"`
	PendingActivities []ActivityRecord                    `json:"pending_activities,omitempty"`
	Block             uint64                              `json:"block,omitempty"`
}

// deprecated
type AuthDataRecord struct {
	Timestamp Timestamp
	Data      AuthDataSnapshot
}

// NewAuthDataRecord creates a new AuthDataRecord instance.
func NewAuthDataRecord(timestamp Timestamp, data AuthDataSnapshot) AuthDataRecord {
	return AuthDataRecord{
		Timestamp: timestamp,
		Data:      data,
	}
}

// ExchangeBalance is balance of a token of an exchange
type ExchangeBalance struct {
	ExchangeID rtypes.ExchangeID `json:"exchange_id"`
	Available  float64           `json:"available"`
	Locked     float64           `json:"locked"`
	Name       string            `json:"name"`
	Error      string            `json:"error"`
}

// AuthdataBalance is balance for a token in reservesetting authata
type AuthdataBalance struct {
	Valid        bool              `json:"valid"`
	AssetID      rtypes.AssetID    `json:"asset_id"`
	Exchanges    []ExchangeBalance `json:"exchanges"`
	Reserve      float64           `json:"reserve"`
	ReserveError string            `json:"reserve_error"`
	Symbol       string            `json:"symbol"`
}

//PendingActivities is pending activities for authdata
type PendingActivities struct {
	SetRates []ActivityRecord `json:"set_rates"`
	Withdraw []ActivityRecord `json:"withdraw"`
	Deposit  []ActivityRecord `json:"deposit"`
	Trades   []ActivityRecord `json:"trades"`
}

// AuthDataResponseV3 is auth data format for reservesetting
type AuthDataResponseV3 struct {
	Balances          []AuthdataBalance `json:"balances"`
	PendingActivities PendingActivities `json:"pending_activities"`
	Version           Version           `json:"version"`
}

// RateEntry contains the buy/sell rates of a token and their compact forms.
type RateEntry struct {
	BaseBuy     *big.Int
	CompactBuy  int8
	BaseSell    *big.Int
	CompactSell int8
	Block       uint64
}

// NewRateEntry creates a new RateEntry instance.
func NewRateEntry(baseBuy *big.Int, compactBuy int8, baseSell *big.Int, compactSell int8, block uint64) RateEntry {
	return RateEntry{
		BaseBuy:     baseBuy,
		CompactBuy:  compactBuy,
		BaseSell:    baseSell,
		CompactSell: compactSell,
		Block:       block,
	}
}

type TXEntry struct {
	Hash           string
	Exchange       string
	AssetID        rtypes.AssetID
	MiningStatus   string
	ExchangeStatus string
	Amount         float64
	Timestamp      Timestamp
}

// NewTXEntry creates new instance of TXEntry.
func NewTXEntry(hash, exchange string, assetID rtypes.AssetID, miningStatus, exchangeStatus string, amount float64, timestamp Timestamp) TXEntry {
	return TXEntry{
		Hash:           hash,
		Exchange:       exchange,
		AssetID:        assetID,
		MiningStatus:   miningStatus,
		ExchangeStatus: exchangeStatus,
		Amount:         amount,
		Timestamp:      timestamp,
	}
}

// RateResponse is the human friendly format of a rate entry to returns in HTTP APIs.
type RateResponse struct {
	Timestamp   Timestamp `json:"timestamp"`
	ReturnTime  Timestamp `json:"return_time"`
	BaseBuy     float64   `json:"base_buy"`
	CompactBuy  int8      `json:"compact_buy"`
	BaseSell    float64   `json:"base_sell"`
	CompactSell int8      `json:"compact_sell"`
	Rate        float64   `json:"rate"`
	Block       uint64    `json:"block"`
}

// AllRateEntry contains rates data of all tokens.
type AllRateEntry struct {
	Timestamp   Timestamp
	ReturnTime  Timestamp
	Data        map[rtypes.AssetID]RateEntry
	BlockNumber uint64
}

// AllRateResponse is the response to query all rates.
type AllRateResponse struct {
	Version       Version                         `json:"version"`
	Timestamp     Timestamp                       `json:"timestamp"`
	ReturnTime    Timestamp                       `json:"return_time"`
	Data          map[rtypes.AssetID]RateResponse `json:"data"`
	BlockNumber   uint64                          `json:"block_number"`
	ToBlockNumber uint64                          `json:"to_block_number"`
}

type TradeHistory struct {
	ID        string  `json:"id"`
	Price     float64 `json:"price"`
	Qty       float64 `json:"qty"`
	Type      string  `json:"type"` // buy or sell
	Timestamp uint64  `json:"timestamp"`
}

// NewTradeHistory creates a new TradeHistory instance.
// typ: "buy" or "sell"
func NewTradeHistory(id string, price, qty float64, typ string, timestamp uint64) TradeHistory {
	return TradeHistory{
		ID:        id,
		Price:     price,
		Qty:       qty,
		Type:      typ,
		Timestamp: timestamp,
	}
}

// type TradingPairID uint64  TODO: update trading pair
type ExchangeTradeHistory map[rtypes.TradingPairID][]TradeHistory

type AllTradeHistory struct {
	Timestamp Timestamp                                  `json:"timestamp"`
	Data      map[rtypes.ExchangeID]ExchangeTradeHistory `json:"data"`
}

type ExStatus struct {
	Timestamp uint64 `json:"timestamp"`
	Status    bool   `json:"status"`
}

type AddressesResponse struct {
	Addresses map[string]ethereum.Address `json:"addresses"`
}

// NewAddressesResponse return new addresses response
func NewAddressesResponse(addrs map[string]ethereum.Address) *AddressesResponse {
	return &AddressesResponse{
		Addresses: addrs,
	}
}

// SiteConfig contain config for a remote api access
type SiteConfig struct {
	URL string `json:"url"`
}

// WorldEndpoints hold detail information to fetch feed(url,header, api key...)
type WorldEndpoints struct {
	OneForgeGoldETH SiteConfig `json:"one_forge_gold_eth"`
	OneForgeGoldUSD SiteConfig `json:"one_forge_gold_usd"`
	GDAXData        SiteConfig `json:"gdax_data"`
	KrakenData      SiteConfig `json:"kraken_data"`
	GeminiData      SiteConfig `json:"gemini_data"`

	CoinbaseETHBTC3 SiteConfig `json:"coinbase_eth_btc_3"`
	BinanceETHBTC3  SiteConfig `json:"binance_eth_btc_3"`

	CoinbaseETHUSDDAI5000 SiteConfig `json:"coinbase_eth_usd_dai_5000"`
}

// ContractAddresses ...
type ContractAddresses struct {
	Proxy   ethereum.Address `json:"proxy"`
	Reserve ethereum.Address `json:"reserve"`
	Wrapper ethereum.Address `json:"wrapper"`
	Pricing ethereum.Address `json:"pricing"`
}

// ExchangeEndpoints ...
type ExchangeEndpoints struct {
	Binance  SiteConfig `json:"binance"`
	Houbi    SiteConfig `json:"houbi"`
	Coinbase SiteConfig `json:"coinbase"`
}

// Nodes ...
type Nodes struct {
	Main   string   `json:"main"`
	Backup []string `json:"backup"`
}

// HumanDuration ...
type HumanDuration time.Duration

// UnmarshalJSON ...
func (d *HumanDuration) UnmarshalJSON(text []byte) error {
	if len(text) < 2 || (text[0] != '"' || text[len(text)-1] != '"') {
		return fmt.Errorf("expect value in double quote")
	}
	v, err := time.ParseDuration(string(text[1 : len(text)-1]))
	if err != nil {
		return err
	}
	*d = HumanDuration(v)
	return nil
}

// FetcherDelay ...
type FetcherDelay struct {
	OrderBook     HumanDuration `json:"order_book"`
	AuthData      HumanDuration `json:"auth_data"`
	RateFetching  HumanDuration `json:"rate_fetching"`
	BlockFetching HumanDuration `json:"block_fetching"`
	GlobalData    HumanDuration `json:"global_data"`
	TradeHistory  HumanDuration `json:"trade_history"`
}

// GasConfig ...
type GasConfig struct {
	PreferUseGasStation     bool   `json:"prefer_use_gas_station"`
	FetchMaxGasCacheSeconds int64  `json:"fetch_max_gas_cache_seconds"`
	GasStationAPIKey        string `json:"gas_station_api_key"`
}

// RawConfig include all configs read from files
type RawConfig struct {
	AWSConfig         archive.AWSConfig `json:"aws_config"`
	WorldEndpoints    WorldEndpoints    `json:"world_endpoints"`
	ContractAddresses ContractAddresses `json:"contract_addresses"`
	ExchangeEndpoints ExchangeEndpoints `json:"exchange_endpoints"`
	Nodes             Nodes             `json:"nodes"`
	FetcherDelay      FetcherDelay      `json:"fetcher_delay"`
	GasConfig         GasConfig         `json:"gas_config"`

	HTTPAPIAddr string `json:"http_api_addr"`

	PricingKeystore   string `json:"keystore_path"`
	PricingPassphrase string `json:"passphrase"`
	DepositKeystore   string `json:"keystore_deposit_path"`
	DepositPassphrase string `json:"passphrase_deposit"`

	BinanceAccountID  string `json:"binance_account_id"`
	BinanceKey        string `json:"binance_key"`
	BinanceSecret     string `json:"binance_secret"`
	BinanceAccount2ID string `json:"binance_account_2_id"`
	Binance2Key       string `json:"binance_2_key"`
	Binance2Secret    string `json:"binance_2_secret"`
	HoubiKey          string `json:"huobi_key"`
	HoubiSecret       string `json:"huobi_secret"`

	IntermediatorKeystore   string `json:"keystore_intermediator_path"`
	IntermediatorPassphrase string `json:"passphrase_intermediate_account"`

	MigrationPath      string `json:"migration_folder_path"`
	MarketDataBaseURL  string `json:"market_data_base_url"`
	AccountDataBaseURL string `json:"account_data_base_url"`
}

// FeedProviderResponse ...
type FeedProviderResponse struct {
	Valid bool
	Error string
	Bid   float64 `json:"bid,string"`
	Ask   float64 `json:"ask,string"`
}
