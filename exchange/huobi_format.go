package exchange

// HuobiDepth order book from huobi
type HuobiDepth struct {
	Status    string `json:"status"`
	Timestamp uint64 `json:"ts"`
	Tick      struct {
		Bids [][]float64 `json:"bids"`
		Asks [][]float64 `json:"asks"`
	} `json:"tick"`
	Reason string `json:"err-msg"`
}

// HuobiExchangeInfo exchange info
type HuobiExchangeInfo struct {
	Status string `json:"status"`
	Data   []struct {
		Base            string `json:"base-currency"`
		Quote           string `json:"quote-currency"`
		PricePrecision  int    `json:"price-precision"`
		AmountPrecision int    `json:"amount-precision"`
	} `json:"data"`
	Reason string `json:"err-msg"`
}

// HuobiInfo is exchange info
type HuobiInfo struct {
	Status string `json:"status"`
	Data   struct {
		ID    uint64 `json:"id"`
		Type  string `json:"type"`
		State string `json:"state"`
		List  []struct {
			Currency string `json:"currency"`
			Type     string `json:"type"`
			Balance  string `json:"balance"`
		} `json:"list"`
	} `json:"data"`
	Reason string `json:"err-msg"`
}

// HuobiTrade is a trade from huobi
type HuobiTrade struct {
	Status  string `json:"status"`
	OrderID string `json:"data"`
	Reason  string `json:"err-msg"`
}

// HuobiCancel a cancel request from huobi
type HuobiCancel struct {
	Status  string `json:"status"`
	OrderID string `json:"data"`
	Reason  string `json:"err-msg"`
}

// HuobiDeposits list deposit from huobi
type HuobiDeposits struct {
	Status string         `json:"status"`
	Data   []HuobiDeposit `json:"data"`
	Reason string         `json:"err-msg"`
}

// HuobiDeposit is deposit record from huobi
type HuobiDeposit struct {
	ID       uint64  `json:"id"`
	TxID     uint64  `json:"transaction-id"`
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
	State    string  `json:"state"`
	TxHash   string  `json:"tx-hash"`
	Address  string  `json:"address"`
}

// HuobiWithdraws is reponse from huobi api for withdrawals request
type HuobiWithdraws struct {
	Status string                 `json:"status"`
	Data   []HuobiWithdrawHistory `json:"data"`
	Reason string                 `json:"err-msg"`
}

// HuobiWithdrawHistory is withdraw history record from huobi
type HuobiWithdrawHistory struct {
	ID       uint64  `json:"id"`
	TxID     uint64  `json:"transaction-id"`
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
	State    string  `json:"state"`
	TxHash   string  `json:"tx-hash"`
	Address  string  `json:"address"`
}

// HuobiWithdraw represent a withdrawal from huobi
type HuobiWithdraw struct {
	Status  string `json:"status"`
	ErrCode string `json:"err-code"`
	Reason  string `json:"err-msg"`
	ID      uint64 `json:"data"`
}

// HuobiOrder represent an order on huobi
type HuobiOrder struct {
	Status string `json:"status"`
	Data   struct {
		OrderID     uint64 `json:"id"`
		Symbol      string `json:"symbol"`
		AccountID   uint64 `json:"account-id"`
		OrigQty     string `json:"amount"`
		Price       string `json:"price"`
		Type        string `json:"type"`
		State       string `json:"state"`
		ExecutedQty string `json:"field-amount"`
	} `json:"data"`
	Reason string `json:"err-msg"`
}

// HuobiDepositAddress return deposit address
type HuobiDepositAddress struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		Currency   string `json:"currency"`
		Address    string `json:"address"`
		AddressTag string `json:"addressTag"`
		Chain      string `json:"chain"`
	} `json:"data"`
}

// HuobiAccounts is accounts on huobi
type HuobiAccounts struct {
	Status string `json:"status"`
	Data   []struct {
		ID     uint64 `json:"id"`
		Type   string `json:"type"`
		State  string `json:"state"`
		UserID uint64 `json:"user-id"`
	} `json:"data"`
	Reason string `json:"err-msg"`
}

// HuobiTradeHistory is trade history from huobi
type HuobiTradeHistory struct {
	Status string `json:"status"`
	Data   []struct {
		ID         uint64 `json:"id"`
		Symbol     string `json:"symbol"`
		Amount     string `json:"amount"`
		Price      string `json:"price"`
		Timestamp  uint64 `json:"created-at"`
		Type       string `json:"type"`
		FinishedAt uint64 `json:"finished-at"`
	} `json:"data"`
}
