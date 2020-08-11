package exchange

// HuobiDepth get its orderbook
type HuobiDepth struct {
	Status    string `json:"status"`
	Timestamp uint64 `json:"ts"`
	Tick      struct {
		Bids [][]float64 `json:"bids"`
		Asks [][]float64 `json:"asks"`
	} `json:"tick"`
	Reason string `json:"err-msg"`
}

// HuobiExchangeInfo exchange info on huobi
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

// HuobiInfo exchange info
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

// HuobiTrade trade on huobi
type HuobiTrade struct {
	Status  string `json:"status"`
	OrderID string `json:"data"`
	Reason  string `json:"err-msg"`
}

// HuobiCancel cancel order on huobi
type HuobiCancel struct {
	Status  string `json:"status"`
	OrderID string `json:"data"`
	Reason  string `json:"err-msg"`
}

// HuobiDeposits deposit on huobi
type HuobiDeposits struct {
	Status string         `json:"status"`
	Data   []HuobiDeposit `json:"data"`
	Reason string         `json:"err-msg"`
}

// HuobiDeposit deposit record on huobi
type HuobiDeposit struct {
	ID       uint64  `json:"id"`
	TxID     uint64  `json:"transaction-id"`
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
	State    string  `json:"state"`
	TxHash   string  `json:"tx-hash"`
	Address  string  `json:"address"`
}

// HuobiWithdraws record
type HuobiWithdraws struct {
	Status string                 `json:"status"`
	Data   []HuobiWithdrawHistory `json:"data"`
	Reason string                 `json:"err-msg"`
}

// HuobiWithdrawHistory from huobi
type HuobiWithdrawHistory struct {
	ID       uint64  `json:"id"`
	TxID     uint64  `json:"transaction-id"`
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
	State    string  `json:"state"`
	TxHash   string  `json:"tx-hash"`
	Address  string  `json:"address"`
	Fee      float64 `json:"fee"`
}

// HuobiWithdraw withdraw result from huobi
type HuobiWithdraw struct {
	Status  string `json:"status"`
	ErrCode string `json:"err-code"`
	Reason  string `json:"err-msg"`
	ID      uint64 `json:"data"`
}

// HuobiOrder orderon huobi
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

// HuobiOpenOrders is list of orders response from huobi
type HuobiOpenOrders struct {
	Status string `json:"status,omitempty"`
	Data   []struct {
		OrderID     uint64 `json:"id,omitempty"`
		Symbol      string `json:"symbol,omitempty"`
		AccountID   uint64 `json:"account-id"`
		OrigQty     string `json:"amount"`
		Price       string `json:"price"`
		Type        string `json:"type"`
		State       string `json:"state"`
		ExecutedQty string `json:"field-amount"`
	} `json:"data"`
	Reason string `json:"err-msg,omitempty"`
}

// HuobiDepositAddress response from huobi
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

// HuobiAccounts get huobi account info
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

// HuobiTradeHistory trade history record on huobi
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
