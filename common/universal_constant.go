package common

const (
	MiningStatusMined       = "mined"
	MiningStatusFailed      = "failed"
	MiningStatusSubmitted   = "submitted"
	MiningStatusLost        = "lost"
	MiningStatusPending     = "pending"
	MiningStatusNA          = "" // status when we can determine blockchain status
	ExchangeStatusDone      = "done"
	ExchangeStatusNA        = ""        // status when we can determine exchange status
	ExchangeStatusPending   = "pending" // status use in withdraw/order, when exchange confirm explicit status
	ExchangeStatusFailed    = "failed"
	ExchangeStatusSubmitted = "submitted"
	ExchangeStatusCancelled = "cancelled"
	ExchangeStatusLost      = "lost"
	ActionDeposit           = "deposit"
	ActionTrade             = "trade"
	ActionWithdraw          = "withdraw"
	ActionSetRate           = "set_rates"
	ActionCancelSetRate     = "cancel_set_rates"
)

const (
	// maxGasPrice this value only use when it can't receive value from network contract
	HighBoundGasPrice float64 = 200.0
)
