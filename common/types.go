package common

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	ether "github.com/ethereum/go-ethereum"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type Version uint64
type Timestamp string

func (self Timestamp) ToUint64() uint64 {
	res, err := strconv.ParseUint(string(self), 10, 64)
	if err != nil {
		panic(err)
	}
	return res
}

func GetTimestamp() Timestamp {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return Timestamp(strconv.Itoa(int(timestamp)))
}

func GetTimepoint() uint64 {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return uint64(timestamp)
}

func TimeToTimepoint(t time.Time) uint64 {
	timestamp := t.UnixNano() / int64(time.Millisecond)
	return uint64(timestamp)
}

func TimepointToTime(t uint64) time.Time {
	return time.Unix(0, int64(t)*int64(time.Millisecond))
}

type ExchangeAddresses struct {
	mu   sync.RWMutex
	data map[string]ethereum.Address
}

func NewExchangeAddresses() *ExchangeAddresses {
	return &ExchangeAddresses{
		mu:   sync.RWMutex{},
		data: map[string]ethereum.Address{},
	}
}

func (self *ExchangeAddresses) Update(tokenID string, address ethereum.Address) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.data[tokenID] = address
}

func (self *ExchangeAddresses) Get(tokenID string) (ethereum.Address, bool) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	address, supported := self.data[tokenID]
	return address, supported
}

func (self *ExchangeAddresses) GetData() map[string]ethereum.Address {
	self.mu.RLock()
	defer self.mu.RUnlock()
	dataCopy := map[string]ethereum.Address{}
	for k, v := range self.data {
		dataCopy[k] = v
	}
	return dataCopy
}

type ExchangePrecisionLimit struct {
	Precision   TokenPairPrecision
	AmountLimit TokenPairAmountLimit
	PriceLimit  TokenPairPriceLimit
	MinNotional float64
}

// ExchangeInfo is written and read concurrently
type ExchangeInfo struct {
	mu   sync.RWMutex
	data map[TokenPairID]ExchangePrecisionLimit
}

func NewExchangeInfo() *ExchangeInfo {
	return &ExchangeInfo{
		mu:   sync.RWMutex{},
		data: map[TokenPairID]ExchangePrecisionLimit{},
	}
}

func (self *ExchangeInfo) Update(pair TokenPairID, data ExchangePrecisionLimit) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.data[pair] = data
}

//Get return Exchange precision limit of on token pair of an exchange
func (self *ExchangeInfo) Get(pair TokenPairID) (ExchangePrecisionLimit, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	if info, exist := self.data[pair]; exist {
		return info, nil
	}
	return ExchangePrecisionLimit{}, errors.New("Token pair is not existed")
}

//GetData return precision and limit when trading all token pairs of an exchange
func (self *ExchangeInfo) GetData() map[TokenPairID]ExchangePrecisionLimit {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.data
}

//TokenPairPrecision represent precision when trading a token pair
type TokenPairPrecision struct {
	Amount int
	Price  int
}

//TokenPairAmountLimit represent amount min and max when trade a token pair
type TokenPairAmountLimit struct {
	Min float64
	Max float64
}

type TokenPairPriceLimit struct {
	Min float64
	Max float64
}

type TradingFee map[string]float64

type FundingFee struct {
	Withdraw map[string]float64
	Deposit  map[string]float64
}

func (self FundingFee) GetTokenFee(token string) float64 {
	withdrawFee := self.Withdraw
	return withdrawFee[token]
}

type ExchangesMinDeposit map[string]float64

type ExchangesMinDepositConfig struct {
	Exchanges map[string]ExchangesMinDeposit `json:"exchanges"`
}

type ExchangeFees struct {
	Trading TradingFee
	Funding FundingFee
}

type ExchangeFeesConfig struct {
	Exchanges map[string]ExchangeFees `json:"exchanges"`
}

func GetFeeFromFile(path string) (ExchangeFeesConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return ExchangeFeesConfig{}, err
	} else {
		result := ExchangeFeesConfig{}
		err := json.Unmarshal(data, &result)
		return result, err
	}
}

func GetMinDepositFromFile(path string) (ExchangesMinDepositConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return ExchangesMinDepositConfig{}, err
	} else {
		result := ExchangesMinDepositConfig{}
		err := json.Unmarshal(data, &result)
		log.Printf("min deposit: %+v", result)
		return result, err
	}
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

type TokenPairID string

func NewTokenPairID(base, quote string) TokenPairID {
	return TokenPairID(fmt.Sprintf("%s-%s", base, quote))
}

type ExchangeID string

type ActivityID struct {
	Timepoint uint64
	EID       string
}

func (self ActivityID) ToBytes() [64]byte {
	var b [64]byte
	temp := make([]byte, 64)
	binary.BigEndian.PutUint64(temp, self.Timepoint)
	temp = append(temp, []byte(self.EID)...)
	copy(b[0:], temp)
	return b
}

func (self ActivityID) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%s|%s", strconv.FormatUint(self.Timepoint, 10), self.EID)), nil
}

func (self *ActivityID) UnmarshalText(b []byte) error {
	id, err := StringToActivityID(string(b))
	if err != nil {
		return err
	} else {
		self.Timepoint = id.Timepoint
		self.EID = id.EID
		return nil
	}
}

func (self ActivityID) String() string {
	res, _ := self.MarshalText()
	return string(res)
}

func StringToActivityID(id string) (ActivityID, error) {
	result := ActivityID{}
	parts := strings.Split(id, "|")
	if len(parts) < 2 {
		return result, fmt.Errorf("Invalid activity id")
	} else {
		timeStr := parts[0]
		eid := strings.Join(parts[1:], "|")
		timepoint, err := strconv.ParseUint(timeStr, 10, 64)
		if err != nil {
			return result, err
		} else {
			result.Timepoint = timepoint
			result.EID = eid
			return result, nil
		}
	}
}

// NewActivityStatus creates new Activity ID.
func NewActivityID(timepoint uint64, eid string) ActivityID {
	return ActivityID{
		Timepoint: timepoint,
		EID:       eid,
	}
}

type ActivityRecord struct {
	Action         string
	ID             ActivityID
	Destination    string
	Params         map[string]interface{}
	Result         map[string]interface{}
	ExchangeStatus string
	MiningStatus   string
	Timestamp      Timestamp
}

func (self ActivityRecord) IsExchangePending() bool {
	switch self.Action {
	case "withdraw":
		return (self.ExchangeStatus == "" || self.ExchangeStatus == "submitted") &&
			self.MiningStatus != "failed"
	case "deposit":
		return (self.ExchangeStatus == "" || self.ExchangeStatus == "pending") &&
			self.MiningStatus != "failed"
	case "trade":
		return self.ExchangeStatus == "" || self.ExchangeStatus == "submitted"
	}
	return true
}

func (self ActivityRecord) IsBlockchainPending() bool {
	switch self.Action {
	case "withdraw", "deposit", "set_rates":
		return (self.MiningStatus == "" || self.MiningStatus == "submitted") && self.ExchangeStatus != "failed"
	}
	return true
}

func (self ActivityRecord) IsPending() bool {
	switch self.Action {
	case "withdraw":
		return (self.ExchangeStatus == "" || self.ExchangeStatus == "submitted" ||
			self.MiningStatus == "" || self.MiningStatus == "submitted") &&
			self.MiningStatus != "failed" && self.ExchangeStatus != "failed"
	case "deposit":
		return (self.ExchangeStatus == "" || self.ExchangeStatus == "pending" ||
			self.MiningStatus == "" || self.MiningStatus == "submitted") &&
			self.MiningStatus != "failed" && self.ExchangeStatus != "failed"
	case "trade":
		return (self.ExchangeStatus == "" || self.ExchangeStatus == "submitted") &&
			self.ExchangeStatus != "failed"
	case "set_rates":
		return (self.MiningStatus == "" || self.MiningStatus == "submitted") &&
			self.ExchangeStatus != "failed"
	}
	return true
}

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
	Quantity float64
	Rate     float64
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
	Data  map[TokenPairID]OnePrice
}

type AllPriceResponse struct {
	Version    Version
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       map[TokenPairID]OnePrice
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

func AddrToString(addr ethereum.Address) string {
	return strings.ToLower(addr.String())
}

type RawBalance big.Int

func (self *RawBalance) ToFloat(decimal int64) float64 {
	return BigToFloat((*big.Int)(self), decimal)
}

func (self RawBalance) MarshalJSON() ([]byte, error) {
	selfInt := (big.Int)(self)
	return selfInt.MarshalJSON()
}

func (self *RawBalance) UnmarshalJSON(text []byte) error {
	selfInt := (*big.Int)(self)
	return selfInt.UnmarshalJSON(text)
}

type BalanceEntry struct {
	Valid      bool
	Error      string
	Timestamp  Timestamp
	ReturnTime Timestamp
	Balance    RawBalance
}

func (self BalanceEntry) ToBalanceResponse(decimal int64) BalanceResponse {
	return BalanceResponse{
		Valid:      self.Valid,
		Error:      self.Error,
		Timestamp:  self.Timestamp,
		ReturnTime: self.ReturnTime,
		Balance:    self.Balance.ToFloat(decimal),
	}
}

type BalanceResponse struct {
	Valid      bool
	Error      string
	Timestamp  Timestamp
	ReturnTime Timestamp
	Balance    float64
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
	OrderId     string
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

type AuthDataSnapshot struct {
	Valid             bool
	Error             string
	Timestamp         Timestamp
	ReturnTime        Timestamp
	ExchangeBalances  map[ExchangeID]EBalanceEntry
	ReserveBalances   map[string]BalanceEntry
	PendingActivities []ActivityRecord
	Block             uint64
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

type AuthDataResponse struct {
	Version    Version
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       struct {
		Valid             bool
		Error             string
		Timestamp         Timestamp
		ReturnTime        Timestamp
		ExchangeBalances  map[ExchangeID]EBalanceEntry
		ReserveBalances   map[string]BalanceResponse
		PendingActivities []ActivityRecord
		Block             uint64
	}
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
	Token          string
	MiningStatus   string
	ExchangeStatus string
	Amount         float64
	Timestamp      Timestamp
}

// NewTXEntry creates new instance of TXEntry.
func NewTXEntry(hash, exchange, token, miningStatus, exchangeStatus string, amount float64, timestamp Timestamp) TXEntry {
	return TXEntry{
		Hash:           hash,
		Exchange:       exchange,
		Token:          token,
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
	Data        map[string]RateEntry
	BlockNumber uint64
}

// AllRateResponse is the response to query all rates.
type AllRateResponse struct {
	Version       Version
	Timestamp     Timestamp
	ReturnTime    Timestamp
	Data          map[string]RateResponse
	BlockNumber   uint64
	ToBlockNumber uint64
}

// KNLog is the common interface of some important logging events.
type KNLog interface {
	TxHash() ethereum.Hash
	BlockNo() uint64
	Type() string
}

type SetCatLog struct {
	Timestamp       uint64
	BlockNumber     uint64
	TransactionHash ethereum.Hash
	Index           uint

	Address  ethereum.Address
	Category string
}

func (self SetCatLog) BlockNo() uint64       { return self.BlockNumber }
func (self SetCatLog) Type() string          { return "SetCatLog" }
func (self SetCatLog) TxHash() ethereum.Hash { return self.TransactionHash }

type TradeLog struct {
	Timestamp       uint64
	BlockNumber     uint64
	TransactionHash ethereum.Hash
	Index           uint

	EtherReceivalSender ethereum.Address
	EtherReceivalAmount *big.Int

	UserAddress ethereum.Address
	SrcAddress  ethereum.Address
	DestAddress ethereum.Address
	SrcAmount   *big.Int
	DestAmount  *big.Int
	FiatAmount  float64

	ReserveAddress ethereum.Address
	WalletAddress  ethereum.Address
	WalletFee      *big.Int
	BurnFee        *big.Int
	IP             string
	Country        string
}

type ReserveRateEntry struct {
	BuyReserveRate  float64
	BuySanityRate   float64
	SellReserveRate float64
	SellSanityRate  float64
}

type ReserveTokenRateEntry map[string]ReserveRateEntry

type ReserveRates struct {
	Timestamp     uint64
	ReturnTime    uint64
	BlockNumber   uint64
	ToBlockNumber uint64
	Data          ReserveTokenRateEntry
}

func (self TradeLog) BlockNo() uint64       { return self.BlockNumber }
func (self TradeLog) Type() string          { return "TradeLog" }
func (self TradeLog) TxHash() ethereum.Hash { return self.TransactionHash }

type StatTicks map[uint64]interface{}

type TradeStats map[string]float64

type VolumeStats struct {
	ETHVolume float64 `json:"eth_amount"`
	USDAmount float64 `json:"usd_amount"`
	Volume    float64 `json:"volume"`
}

// NewVolumeStats creates a new instance of VolumeStats.
func NewVolumeStats(ethVolume, usdAmount, volume float64) VolumeStats {
	return VolumeStats{
		ETHVolume: ethVolume,
		USDAmount: usdAmount,
		Volume:    volume,
	}
}

type BurnFeeStats struct {
	TotalBurnFee float64
}

// NewBurnFeeStats creates a new instance of BurnFeeStats.
func NewBurnFeeStats(totalBurnFee float64) BurnFeeStats {
	return BurnFeeStats{TotalBurnFee: totalBurnFee}
}

type BurnFeeStatsTimeZone map[string]map[uint64]BurnFeeStats

type VolumeStatsTimeZone map[string]map[uint64]VolumeStats

type FeeStats map[int64]map[uint64]float64

type MetricStats struct {
	ETHVolume          float64 `json:"total_eth_volume"`
	USDVolume          float64 `json:"total_usd_amount"`
	BurnFee            float64 `json:"total_burn_fee"`
	TradeCount         int     `json:"total_trade"`
	UniqueAddr         int     `json:"unique_addresses"`
	KYCEd              int     `json:"kyced_addresses"`
	NewUniqueAddresses int     `json:"new_unique_addresses"`
	USDPerTrade        float64 `json:"usd_per_trade"`
	ETHPerTrade        float64 `json:"eth_per_trade"`
}

// NewMetricStats creates a new instance of MetricStats.
func NewMetricStats(ethVolume, usdVolume, burnFee float64,
	tradeCount, uniqueAddr, kycEd, newUniqueAddresses int,
	usdPerTrade, ethPerTrade float64) MetricStats {
	return MetricStats{
		ETHVolume:          ethVolume,
		USDVolume:          usdVolume,
		BurnFee:            burnFee,
		TradeCount:         tradeCount,
		UniqueAddr:         uniqueAddr,
		KYCEd:              kycEd,
		NewUniqueAddresses: newUniqueAddresses,
		USDPerTrade:        usdPerTrade,
		ETHPerTrade:        ethPerTrade,
	}
}

type MetricStatsTimeZone map[int64]map[uint64]MetricStats

type UserInfo struct {
	Addr      string  `json:"user_address"`
	Email     string  `json:"email"`
	ETHVolume float64 `json:"total_eth_volume"`
	USDVolume float64 `json:"total_usd_volume"`
}

type UserListResponse []UserInfo

func (h UserListResponse) Less(i, j int) bool {
	return h[i].ETHVolume < h[j].ETHVolume
}

func (h UserListResponse) Len() int      { return len(h) }
func (h UserListResponse) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

type UserInfoTimezone map[int64]map[uint64]UserInfo

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

type ExchangeTradeHistory map[TokenPairID][]TradeHistory

type AllTradeHistory struct {
	Timestamp Timestamp
	Data      map[ExchangeID]ExchangeTradeHistory
}

// NewAllTradeHistory creates a new AllTradeHistory instance.
func NewAllTradeHistory(timestamp Timestamp, data map[ExchangeID]ExchangeTradeHistory) AllTradeHistory {
	return AllTradeHistory{
		Timestamp: timestamp,
		Data:      data,
	}
}

type ExStatus struct {
	Timestamp uint64 `json:"timestamp"`
	Status    bool   `json:"status"`
}

type ExchangesStatus map[string]ExStatus

type TradeLogGeoInfoResp struct {
	Success bool `json:"success"`
	Data    struct {
		IP      string `json:"IP"`
		Country string `json:"Country"`
	} `json:"data"`
}

type HeatmapType struct {
	TotalETHValue        float64 `json:"total_eth_value"`
	TotalFiatValue       float64 `json:"total_fiat_value"`
	ToTalBurnFee         float64 `json:"total_burn_fee"`
	TotalTrade           int     `json:"total_trade"`
	TotalUniqueAddresses int     `json:"total_unique_addr"`
	TotalKYCUser         int     `json:"total_kyc_user"`
}

type Heatmap map[string]HeatmapType

type HeatmapObject struct {
	Country              string  `json:"country"`
	TotalETHValue        float64 `json:"total_eth_value"`
	TotalFiatValue       float64 `json:"total_fiat_value"`
	ToTalBurnFee         float64 `json:"total_burn_fee"`
	TotalTrade           int     `json:"total_trade"`
	TotalUniqueAddresses int     `json:"total_unique_addr"`
	TotalKYCUser         int     `json:"total_kyc_user"`
}

type HeatmapResponse []HeatmapObject

func (h HeatmapResponse) Less(i, j int) bool {
	return h[i].TotalETHValue < h[j].TotalETHValue
}

func (h HeatmapResponse) Len() int      { return len(h) }
func (h HeatmapResponse) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

type AnalyticPriceResponse struct {
	Timestamp uint64
	Data      map[string]interface{}
}

// NewAnalyticPriceResponse creates a new instance of AnalyticPriceResponse.
func NewAnalyticPriceResponse(timestamp uint64, data map[string]interface{}) AnalyticPriceResponse {
	return AnalyticPriceResponse{Timestamp: timestamp, Data: data}
}

type ExchangeNotiContent struct {
	FromTime  uint64 `json:"fromTime"`
	ToTime    uint64 `json:"toTime"`
	IsWarning bool   `json:"isWarning"`
	Message   string `json:"msg"`
}

type ExchangeTokenNoti map[string]ExchangeNotiContent

type ExchangeActionNoti map[string]ExchangeTokenNoti

type ExchangeNotifications map[string]ExchangeActionNoti

type CountryTokenHeatmap map[string]VolumeStats

type TokenHeatmap struct {
	Country   string  `json:"country"`
	Volume    float64 `json:"volume"`
	ETHVolume float64 `json:"total_eth_value"`
	USDVolume float64 `json:"total_fiat_value"`
}

type TokenHeatmapResponse []TokenHeatmap

func (t TokenHeatmapResponse) Less(i, j int) bool {
	return t[i].Volume < t[j].Volume
}

func (t TokenHeatmapResponse) Len() int      { return len(t) }
func (t TokenHeatmapResponse) Swap(i, j int) { t[i], t[j] = t[j], t[i] }

type UsersVolume map[string]StatTicks

// NewFilterQuery creates a new ether.FilterQuery instance.
func NewFilterQuery(fromBlock, toBlock *big.Int, addresses []ethereum.Address, topics [][]ethereum.Hash) ether.FilterQuery {
	return ether.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Addresses: addresses,
		Topics:    topics,
	}
}

type TransactionInfo struct {
	BlockNumber string `json:"blockNumber"`
	TimeStamp   string `json:"timeStamp"`
	Value       string `json:"value"`
	GasPrice    string `json:"gasPrice"`
	GasUsed     string `json:"gasUsed"`
}

type SetRateTxInfo struct {
	BlockNumber      string `json:"blockNumber"`
	TimeStamp        string `json:"timeStamp"`
	TransactionIndex string `json:"transactionIndex"`
	Input            string `json:"input"`
	GasPrice         string `json:"gasPrice"`
	GasUsed          string `json:"gasUsed"`
}

type StoreSetRateTx struct {
	TimeStamp uint64 `json:"timeStamp"`
	GasPrice  uint64 `json:"gasPrice"`
	GasUsed   uint64 `json:"gasUsed"`
}

func GetStoreTx(tx SetRateTxInfo) (StoreSetRateTx, error) {
	var storeTx StoreSetRateTx
	gasPriceUint, err := strconv.ParseUint(tx.GasPrice, 10, 64)
	if err != nil {
		log.Printf("Cant convert %s to uint64", tx.GasPrice)
		return storeTx, err
	}
	gasUsedUint, err := strconv.ParseUint(tx.GasUsed, 10, 64)
	if err != nil {
		log.Printf("Cant convert %s to uint64", tx.GasUsed)
		return storeTx, err
	}
	timeStampUint, err := strconv.ParseUint(tx.TimeStamp, 10, 64)
	if err != nil {
		log.Printf("Cant convert %s to uint64", tx.TimeStamp)
		return storeTx, err
	}
	storeTx = StoreSetRateTx{
		TimeStamp: timeStampUint,
		GasPrice:  gasPriceUint,
		GasUsed:   gasUsedUint,
	}
	return storeTx, nil
}

type FeeSetRate struct {
	TimeStamp     uint64     `json:"timeStamp"`
	GasUsed       *big.Float `json:"gasUsed"`
	TotalGasSpent *big.Float `json:"totalGasSpent"`
}
