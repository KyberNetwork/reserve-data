package http

type Permission int

const (
	ReadOnlyPermission    Permission = iota // can only read data
	RebalancePermission                     // can do everything except configure setting
	ConfigurePermission                     // can read data and configure setting, cannot set rates, deposit, withdraw, trade, cancel activities
	ConfirmConfPermission                   // can read data and confirm configuration proposal
)
