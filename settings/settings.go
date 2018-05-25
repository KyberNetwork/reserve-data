package settings

const (
	TOKEN_DB_FILE_PATH      string = "/go/src/github.com/KyberNetwork/reserve-data/cmd/token.db"
	TOKEN_DEFAULT_JSON_PATH string = "/go/src/github.com/KyberNetwork/reserve-data/cmd/token.json"
)

type Settings struct {
	Tokens    *TokenSetting
	Exchanges *ExchangeSetting
	Addresses *AddressSetting
}

var setting = InitSetting()

func InitToken() *TokenSetting {
	return &TokenSetting{}
}

func InitSetting() *Settings {
	return &Settings{}
}
