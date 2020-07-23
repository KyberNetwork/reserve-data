package common

import "github.com/pkg/errors"

var (
	// ErrNotFound is the error to return when no record is found in database.
	ErrNotFound = errors.New("not found")
	// ErrAddressMissing is returned when the operation require address, but not provided or zero.
	ErrAddressMissing = errors.New("address is zero or missing")
	// ErrSymbolExists is returned when creating new asset with duplicated symbol.
	ErrSymbolExists = errors.New("symbol already exists")
	// ErrAddressExists is returned when the address to create is already exists.
	ErrAddressExists = errors.New("address already exists")
	// ErrExchangeFeeMissing is the error to return when user try to enable exchange, but fees are not set.
	ErrExchangeFeeMissing = errors.New("missing exchange fee configuration")
	// ErrPWIMissing is returned when PWI configuration is missing when set rate strategy is defined
	ErrPWIMissing = errors.New("missing PWI configuration")
	// ErrRebalanceQuadraticMissing is returned when rebalance quadratic configuration is missing when
	// rebalance is set to true.
	ErrRebalanceQuadraticMissing = errors.New("missing rebalance quadratic configuration")
	// ErrAssetExchangeMissing is returned when asset exchange configuration is missing for asset with
	// rebalance set to true.
	ErrAssetExchangeMissing = errors.New("missing asset exchange configuration")
	// ErrAssetTargetMissing is returned then asset target configuration is missing for asset with
	// rebalance set to true.
	ErrAssetTargetMissing = errors.New("missing asset target configuration")
	// ErrBadTradingPairConfiguration is returned when bad trading pair configuration is given.
	ErrBadTradingPairConfiguration = errors.New("bad trading pair configuration")
	// ErrDepositAddressMissing is returned when asset is transferable but no deposit address is
	// provided.
	ErrDepositAddressMissing = errors.New("missing deposit address for transferable asset")
	// ErrAssetExchangeAlreadyExist is returned when create a asset_exchange existing in db
	ErrAssetExchangeAlreadyExist = errors.New("asset already on exchange")
	// ErrQuoteAssetInvalid is returned when quote asset doesn't contain exchange id or
	// contains field is_quote=false in the  create/update trading  request
	ErrQuoteAssetInvalid = errors.New("quote asset is invalid")
	// ErrBaseAssetInvalid is returned when base asset doesn't contain exchange id in the  create/update trading  request
	ErrBaseAssetInvalid = errors.New("base asset is invalid")
	// ErrTradingByAlreadyExists is returned when trading by already exist for asset<->tradingPair
	ErrTradingByAlreadyExists = errors.New("trading by already exists")
	// ErrExchangeNotExists is returned when exchange is not exist.
	ErrExchangeNotExists = errors.New("exchange does not exist")
	// ErrFeedConfiguration is returned when exchange is not exist.
	ErrFeedConfiguration = errors.New("feed configuration does not exist")
	// ErrAssetNotExists is returned when foreign key is not exist.
	ErrAssetNotExists = errors.New("asset does not exist")
	// ErrTradingByAssetIDInvalid is returned when trading by asset id is invalid
	ErrTradingByAssetIDInvalid = errors.New("asset_id is either base or quote of trading_pair")
	// ErrAssetExchangeDeleteViolation is returned when there are trading pair relating to deleted asset exchange
	ErrAssetExchangeDeleteViolation = errors.New("asset exchange can be deleted only when no trading pair use the correspond asset")
	// ErrSettingChangeExists is return if SettingChange in same catalog already exists
	ErrSettingChangeExists = errors.New("setting change already exists, confirm/reject it first")
	// ErrAssetAddressIsNotIndexInContract is return if address is not index in contract
	ErrAssetAddressIsNotIndexInContract = errors.New("asset address is not index in address please check again")
	// ErrBlockchainHaveNotInitiated is return if blockchain have not initiated yet
	ErrBlockchainHaveNotInitiated = errors.New("blockchain instance have not initiated")
	// ErrNormalUpdaterPerPeriodNotPositive is return if normal update per period <= 0
	ErrNormalUpdaterPerPeriodNotPositive = errors.New("normal update per period is not positive")
	// ErrMaxImbalanceRatioNotPositive is return if max imbalance ratio <= 0
	ErrMaxImbalanceRatioNotPositive = errors.New("max imbalance ratio is not positive")
)
