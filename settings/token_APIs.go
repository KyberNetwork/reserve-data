package settings

import (
	"log"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

func (self *TokenSetting) GetAllTokens() ([]common.Token, error) {
	return self.Storage.GetAllTokens()
}

func (self *TokenSetting) GetActiveTokens() ([]common.Token, error) {
	return self.Storage.GetActiveTokens()
}

func (self *TokenSetting) GetInternalTokens() ([]common.Token, error) {
	return self.Storage.GetInternalTokens()
}

func (self *TokenSetting) GetExternalTokens() ([]common.Token, error) {
	return self.Storage.GetExternalTokens()
}

func (self *TokenSetting) GetTokenByID(id string) (common.Token, error) {
	return self.Storage.GetTokenByID(id)
}

func (self *TokenSetting) GetActiveTokenByID(id string) (common.Token, error) {
	return self.Storage.GetActiveTokenByID(id)
}

func (self *TokenSetting) GetInternalTokenByID(id string) (common.Token, error) {
	return self.Storage.GetInternalTokenByID(id)
}

func (self *TokenSetting) GetExternalTokenByID(id string) (common.Token, error) {
	return self.Storage.GetExternalTokenByID(id)
}

func (self *TokenSetting) GetTokenByAddress(addr ethereum.Address) (common.Token, error) {
	return self.Storage.GetTokenByAddress(addr)
}

func (self *TokenSetting) GetActiveTokenByAddress(addr ethereum.Address) (common.Token, error) {
	return self.Storage.GetActiveTokenByAddress(addr)
}

func (self *TokenSetting) GetInternalTokenByAddress(addr ethereum.Address) (common.Token, error) {
	return self.Storage.GetInternalTokenByAddress(addr)
}

func (self *TokenSetting) GetExternalTokenByAddress(addr ethereum.Address) (common.Token, error) {
	return self.Storage.GetExternalTokenByAddress(addr)
}

func (self *TokenSetting) ETHToken() common.Token {
	eth, err := self.Storage.GetInternalTokenByID("ETH")
	if err != nil {
		log.Panicf("There is no ETH token in token DB, this should not happen (%s)", err)
	}
	return eth
}
