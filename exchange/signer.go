package exchange

type Signer interface {
	GetLiquiKey() string
	GetBittrexKey() string
	LiquiSign(msg string) string
	BittrexSign(msg string) string
}
