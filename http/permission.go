package http

import (
	"github.com/KyberNetwork/reserve-stats/lib/httputil/middleware"
)

const (
	readOnlyPermission    middleware.Permission = iota // can only read data
	rebalancePermission                                // can do everything except configure setting
	configurePermission                                // can read data and configure setting, cannot set rates, deposit, withdraw, trade, cancel activities
	confirmConfPermission                              // can read data and confirm configuration proposal
)

var canReadPermissions = []middleware.Permission{
	readOnlyPermission,
	rebalancePermission,
	configurePermission,
	confirmConfPermission,
}

var confirmPermissions = []middleware.Permission{confirmConfPermission}

var configurePermissions = []middleware.Permission{configurePermission}

var rebalancePermissions = []middleware.Permission{rebalancePermission}
