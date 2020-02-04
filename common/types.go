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
type ExchangeInfo map[uint64]ExchangePrecisionLimit

func NewExchangeInfo() ExchangeInfo {
	return ExchangeInfo(make(map[uint64]ExchangePrecisionLimit))
}

func (ei ExchangeInfo) Get(pair uint64) (ExchangePrecisionLimit, error) {
	info, exist := ei[pair]
	if !exist {
		return info, fmt.Errorf("token pair is not existed")
	}
	return info, nil

}

func (ei ExchangeInfo) GetData() map[uint64]ExchangePrecisionLimit {
	data := map[uint64]ExchangePrecisionLimit(ei)
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

// ActivityRecord object
type ActivityRecord struct {
	Action         string         `json:"Action,omitempty"`
	ID             ActivityID     `json:"ID,omitempty"`
	Destination    string         `json:"Destination,omitempty"`
	Params         ActivityParams `json:"Params,omitempty"`
	Result         ActivityResult `json:"Result,omitempty"`
	ExchangeStatus string         `json:"ExchangeStatus,omitempty"`
	MiningStatus   string         `json:"MiningStatus,omitempty"`
	Timestamp      Timestamp      `json:"Timestamp,omitempty"`
}

// ActivityParams is params for activity
type ActivityParams struct {
	// deposit, withdraw params
	Exchange  ExchangeID `json:"exchange,omitempty"`
	Asset     uint64     `json:"asset,omitempty"`
	Amount    float64    `json:"amount,omitempty"`
	Timepoint uint64     `json:"timepoint,omitempty"`
	// SetRates params
	Assets []uint64   `json:"assets,omitempty"` // list of asset id
	Buys   []*big.Int `json:"buys,omitempty"`
	Sells  []*big.Int `json:"sells,omitempty"`
	Block  *big.Int   `json:"block,omitempty"`
	AFPMid []*big.Int `json:"afpMid,omitempty"`
	Msgs   []string   `json:"msgs,omitempty"`
	// Trade params
	Type  string  `json:"type,omitempty"`
	Base  string  `json:"base,omitempty"`
	Quote string  `json:"quote,omitempty"`
	Rate  float64 `json:"rate,omitempty"`
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
}

//NewActivityRecord return an activity record
func NewActivityRecord(action string, id ActivityID, destination string, params ActivityParams, result ActivityResult, exStatus, miStatus string, timestamp Timestamp) ActivityRecord {
	return ActivityRecord{
		Action:         action,
		ID:             id,
		Destination:    destination,
		Params:         params,
		Result:         result,
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
	case ActionSetRate:
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
	Error          error
}

// NewActivityStatus creates a new ActivityStatus instance.
func NewActivityStatus(exchangeStatus, tx string, blockNumber uint64, miningStatus string, err error) ActivityStatus {
	return ActivityStatus{
		ExchangeStatus: exchangeStatus,
		Tx:             tx,
		BlockNumber:    blockNumber,
		MiningStatus:   miningStatus,
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
	Data  map[uint64]OnePrice
}

type AllPriceResponse struct {
	Version    Version
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       map[uint64]OnePrice
	Block      uint64
}

type OnePriceResponse struct {
	Version    Version
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       OnePrice
	Block      uint64
}

type OnePrice map[ExchangeID]ExchangePrice

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
	Valid      bool
	Error      string
	Timestamp  Timestamp
	ReturnTime Timestamp
	Balance    RawBalance
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

type Order struct {
	ID          string // standard id across multiple exchanges
	Base        string
	Quote       string
	OrderID     string
	Price       float64
	OrigQty     float64 // original quantity
	ExecutedQty float64 // matched quantity
	TimeInForce string
	Type        string // market or limit
	Side        string // buy or sell
	StopPrice   string
	IcebergQty  string
	Time        uint64
}

type OrderEntry struct {
	Valid      bool
	Error      string
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       []Order
}

type AllOrderEntry map[ExchangeID]OrderEntry

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
	AvailableBalance map[string]float64
	LockedBalance    map[string]float64
	DepositBalance   map[string]float64
	Status           bool
}

type AllEBalanceResponse struct {
	Version    Version
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       map[ExchangeID]EBalanceEntry
}

// AuthDataSnapshot object
type AuthDataSnapshot struct {
	Valid             bool                         `json:"Valid,omitempty"`
	Error             string                       `json:"Error,omitempty"`
	Timestamp         Timestamp                    `json:"Timestamp,omitempty"`
	ReturnTime        Timestamp                    `json:"ReturnTime,omitempty"`
	ExchangeBalances  map[ExchangeID]EBalanceEntry `json:"ExchangeBalances,omitempty"`
	ReserveBalances   map[string]BalanceEntry      `json:"ReserveBalances,omitempty"`
	PendingActivities []ActivityRecord             `json:"PendingActivities,omitempty"`
	Block             uint64                       `json:"Block,omitempty"`
}

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
	ExchangeID uint64  `json:"exchange_id"`
	Available  float64 `json:"available"`
	Locked     float64 `json:"locked"`
	Name       string  `json:"name"`
	Error      string  `json:"error"`
}

// AuthdataBalance is balance for a token in reservesetting authata
type AuthdataBalance struct {
	Valid        bool              `json:"valid"`
	AssetID      uint64            `json:"asset_id"`
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
	AssetID        uint64
	MiningStatus   string
	ExchangeStatus string
	Amount         float64
	Timestamp      Timestamp
}

// NewTXEntry creates new instance of TXEntry.
func NewTXEntry(hash, exchange string, assetID uint64, miningStatus, exchangeStatus string, amount float64, timestamp Timestamp) TXEntry {
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
	Timestamp   Timestamp
	ReturnTime  Timestamp
	BaseBuy     float64
	CompactBuy  int8
	BaseSell    float64
	CompactSell int8
	Rate        float64
	Block       uint64
}

// AllRateEntry contains rates data of all tokens.
type AllRateEntry struct {
	Timestamp   Timestamp
	ReturnTime  Timestamp
	Data        map[uint64]RateEntry
	BlockNumber uint64
}

// AllRateResponse is the response to query all rates.
type AllRateResponse struct {
	Version       Version
	Timestamp     Timestamp
	ReturnTime    Timestamp
	Data          map[uint64]RateResponse
	BlockNumber   uint64
	ToBlockNumber uint64
}

type TradeHistory struct {
	ID        string
	Price     float64
	Qty       float64
	Type      string // buy or sell
	Timestamp uint64
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

type ExchangeTradeHistory map[uint64][]TradeHistory

type AllTradeHistory struct {
	Timestamp Timestamp
	Data      map[ExchangeID]ExchangeTradeHistory
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
	GoldData        SiteConfig `json:"gold_data"`
	OneForgeGoldETH SiteConfig `json:"one_forge_gold_eth"`
	OneForgeGoldUSD SiteConfig `json:"one_forge_gold_usd"`
	GDAXData        SiteConfig `json:"gdax_data"`
	KrakenData      SiteConfig `json:"kraken_data"`
	GeminiData      SiteConfig `json:"gemini_data"`
	GeminiBTC       SiteConfig `json:"gemini_btc"`

	CoinbaseBTC SiteConfig `json:"coinbase_btc"`
	BinanceBTC  SiteConfig `json:"binance_btc"`

	CoinbaseUSDC SiteConfig `json:"coinbase_usdc"`
	BinanceUSDC  SiteConfig `json:"binance_usdc"`
	CoinbaseUSD  SiteConfig `json:"coinbase_usd"`
	CoinbaseDAI  SiteConfig `json:"coinbase_dai"`
	HitDai       SiteConfig `json:"hit_dai"`

	BitFinexUSDT SiteConfig `json:"bit_finex_usdt"`
	BinanceUSDT  SiteConfig `json:"binance_usdt"`
	BinancePAX   SiteConfig `json:"binance_pax"`
	BinanceTUSD  SiteConfig `json:"binance_tusd"`
}

// ContractAddresses ...
type ContractAddresses struct {
	Proxy   ethereum.Address `json:"proxy"`
	Reserve ethereum.Address `json:"reserve"`
	Wrapper ethereum.Address `json:"wrapper"`
	Pricing ethereum.Address `json:"pricing"`
}

// ExchangeSiteConfig ...
type ExchangeSiteConfig struct {
	URL       string `json:"url"`
	AuthenURL string `json:"authen_url"`
}

// ExchangeEndpoints ...
type ExchangeEndpoints struct {
	Binance ExchangeSiteConfig `json:"binance"`
	Houbi   ExchangeSiteConfig `json:"houbi"`
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

// RawConfig include all configs read from files
type RawConfig struct {
	AWSConfig         archive.AWSConfig `json:"aws_config"`
	WorldEndpoints    WorldEndpoints    `json:"world_endpoints"`
	ContractAddresses ContractAddresses `json:"contract_addresses"`
	ExchangeEndpoints ExchangeEndpoints `json:"exchange_endpoints"`
	Nodes             Nodes             `json:"nodes"`
	FetcherDelay      FetcherDelay      `json:"fetcher_delay"`

	HTTPAPIAddr string `json:"http_api_addr"`

	PricingKeystore   string `json:"keystore_path"`
	PricingPassphrase string `json:"passphrase"`
	DepositKeystore   string `json:"keystore_deposit_path"`
	DepositPassphrase string `json:"passphrase_deposit"`

	BinanceKey     string `json:"binance_key"`
	BinanceSecret  string `json:"binance_secret"`
	Binance2Key    string `json:"binance_2_key"`
	Binance2Secret string `json:"binance_2_secret"`
	HoubiKey       string `json:"huobi_key"`
	HoubiSecret    string `json:"huobi_secret"`

	IntermediatorKeystore   string `json:"keystore_intermediator_path"`
	IntermediatorPassphrase string `json:"passphrase_intermediate_account"`
}
