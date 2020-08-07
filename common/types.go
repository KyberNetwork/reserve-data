package common

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
)

type Version uint64
type Timestamp string

func (ts Timestamp) MustToUint64() uint64 {
	res, err := strconv.ParseUint(string(ts), 10, 64)
	//  this should never happen. Timestamp is never manually entered.
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

// TimeToTimepoint convert from timestamp to millisecond
func TimeToTimepoint(t time.Time) uint64 {
	timestamp := t.UnixNano() / int64(time.Millisecond)
	return uint64(timestamp)
}

// TimepointToTime convert from time millisecond to timestamp
func TimepointToTime(t uint64) time.Time {
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
type ExchangeInfo map[TokenPairID]ExchangePrecisionLimit

func NewExchangeInfo() ExchangeInfo {
	return ExchangeInfo(make(map[TokenPairID]ExchangePrecisionLimit))
}

func (ei ExchangeInfo) Get(pair TokenPairID) (ExchangePrecisionLimit, error) {
	info, exist := ei[pair]
	if !exist {
		return info, fmt.Errorf("token pair is not existed")
	}
	return info, nil

}

func (ei ExchangeInfo) GetData() map[TokenPairID]ExchangePrecisionLimit {
	data := map[TokenPairID]ExchangePrecisionLimit(ei)
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

type TokenPairID string

func NewTokenPairID(base, quote string) TokenPairID {
	return TokenPairID(fmt.Sprintf("%s-%s", base, quote))
}

type ExchangeID string

type ActivityID struct {
	Timepoint uint64
	EID       string
}

func (ai ActivityID) ToBytes() [64]byte {
	var b [64]byte
	temp := make([]byte, 64)
	binary.BigEndian.PutUint64(temp, ai.Timepoint)
	temp = append(temp, []byte(ai.EID)...)
	copy(b[0:], temp)
	return b
}

func (ai ActivityID) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%s|%s", strconv.FormatUint(ai.Timepoint, 10), ai.EID)), nil
}

func (ai *ActivityID) UnmarshalText(b []byte) error {
	id, err := StringToActivityID(string(b))
	if err != nil {
		return err
	}
	ai.Timepoint = id.Timepoint
	ai.EID = id.EID
	return nil
}

func (ai ActivityID) String() string {
	res, _ := ai.MarshalText()
	return string(res)
}

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

//New ActivityRecord return an activity record with params["token"] only as token.ID
func NewActivityRecord(action string, id ActivityID, destination string, params, result map[string]interface{}, exStatus, miStatus string, timestamp Timestamp) ActivityRecord {
	//if any params is a token, save it as tokenID
	for k, v := range params {
		if tok, ok := v.(Token); ok {
			params[k] = tok.ID
		}
	}
	tokens, ok := params["tokens"].([]Token)
	if ok {
		var tokenIDs []string
		for _, t := range tokens {
			tokenIDs = append(tokenIDs, t.ID)
		}
		params["tokens"] = tokenIDs
	}
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

func (ar ActivityRecord) IsBlockchainPending() bool {
	switch ar.Action {
	case ActionWithdraw, ActionDeposit, ActionSetRate:
		return (ar.MiningStatus == "" || ar.MiningStatus == MiningStatusSubmitted) && ar.ExchangeStatus != ExchangeStatusFailed
	}
	return true
}

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

// ActivityStatus for trade, deposit, withdraw
type ActivityStatus struct {
	ExchangeStatus string
	Tx             string
	BlockNumber    uint64
	MiningStatus   string
	Fee            float64
	Error          error
}

// NewActivityStatus creates a new ActivityStatus instance.
func NewActivityStatus(exchangeStatus, tx string, blockNumber uint64, miningStatus string, fee float64, err error) ActivityStatus {
	return ActivityStatus{
		ExchangeStatus: exchangeStatus,
		Tx:             tx,
		BlockNumber:    blockNumber,
		MiningStatus:   miningStatus,
		Fee:            fee,
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
	Symbol      string // I don't add json tag to keep it consistent with other existed component
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

type ExStatus struct {
	Timestamp uint64 `json:"timestamp"`
	Status    bool   `json:"status"`
}

type ExchangesStatus map[string]ExStatus

type ExchangeNotiContent struct {
	FromTime  uint64 `json:"fromTime"`
	ToTime    uint64 `json:"toTime"`
	IsWarning bool   `json:"isWarning"`
	Message   string `json:"msg"`
}

type ExchangeTokenNoti map[string]ExchangeNotiContent

type ExchangeActionNoti map[string]ExchangeTokenNoti

type ExchangeNotifications map[string]ExchangeActionNoti

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

type FeeSetRate struct {
	TimeStamp     uint64     `json:"timeStamp"`
	GasUsed       *big.Float `json:"gasUsed"`
	TotalGasSpent *big.Float `json:"totalGasSpent"`
}

type AddressesResponse struct {
	Addresses map[string]interface{} `json:"addresses"`
}

func NewAddressResponse(addrs map[string]interface{}) *AddressesResponse {
	return &AddressesResponse{
		Addresses: addrs,
	}
}

type TokenResponse struct {
	Tokens  []Token `json:"tokens"`
	Version uint64  `json:"version"`
}

func NewTokenResponse(tokens []Token, version uint64) *TokenResponse {
	return &TokenResponse{
		Tokens:  tokens,
		Version: version,
	}
}

type ExchangeResponse struct {
	Exchanges map[string]*ExchangeSetting `json:"exchanges"`
	Version   uint64                      `json:"version"`
}

func NewExchangeResponse(exs map[string]*ExchangeSetting, version uint64) *ExchangeResponse {
	return &ExchangeResponse{
		Exchanges: exs,
		Version:   version,
	}
}

type AllSettings struct {
	Addresses *AddressesResponse `json:"addresses"`
	Tokens    *TokenResponse     `json:"tokens"`
	Exchanges *ExchangeResponse  `json:"exchanges"`
}

func NewAllSettings(addrs *AddressesResponse, toks *TokenResponse, exs *ExchangeResponse) *AllSettings {
	return &AllSettings{
		Addresses: addrs,
		Tokens:    toks,
		Exchanges: exs,
	}
}

//QuantityStepFunction represent a quatity step function for a token
type QuantityStepFunction struct {
	XBuy  []*big.Int `json:"x_buy"`
	YBuy  []*big.Int `json:"y_buy"`
	XSell []*big.Int `json:"x_sell"`
	YSell []*big.Int `json:"y_sell"`
}

//ImbalanceStepFunction represent an imbalance step function for a token
type ImbalanceStepFunction struct {
	XBuy  []*big.Int `json:"x_buy"`
	YBuy  []*big.Int `json:"y_buy"`
	XSell []*big.Int `json:"x_sell"`
	YSell []*big.Int `json:"y_sell"`
}

//StepFunctionResponse api response about a token step function
type StepFunctionResponse struct {
	QuantityStepResponse  QuantityStepFunction  `json:"quantity_step_function"`
	ImbalanceStepResponse ImbalanceStepFunction `json:"imbalance_step_function"`
}

//FetcherConfiguration is configuration of fetcher
type FetcherConfiguration struct {
	BTC bool `json:"btc" binding:"exists"`
}

// AddressConfig type defines a list of address attribute available in core.
type AddressConfig struct {
	Reserve string `json:"reserve"`

	// not officially shown but you can find it here: https://github.com/KyberNetwork/smart-contracts/blob/master/contracts/mock/Wrapper.sol
	Wrapper string `json:"wrapper"`

	// Pricing is ConversionRates contract
	Pricing string `json:"pricing"`

	// as we used to miscalling this contract as network
	// we still keep its json name as network
	// then other component won't be affect
	Proxy string `json:"network"`
}

// RunnerConfig list configs for runner
type RunnerConfig struct {
	OrderBookFetchingInterval  time.Duration
	AuthDataFetchingInterval   time.Duration
	RateFetchingInterval       time.Duration
	BlockFetchingInterval      time.Duration
	GlobalDataFetchingInterval time.Duration
}

// FeedProviderResponse ...
type FeedProviderResponse struct {
	Valid bool
	Error string
	Bid   float64 `json:"bid,string"`
	Ask   float64 `json:"ask,string"`
}
