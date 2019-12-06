package configuration

import (
	"github.com/KyberNetwork/reserve-data/common/blockchain"
)

func PricingSignerFromConfigFile(keystorePath, passphrase string) *blockchain.EthereumSigner {
	return blockchain.NewEthereumSigner(keystorePath, passphrase)
}

func DepositSignerFromConfigFile(keystore, passphrase string) *blockchain.EthereumSigner {
	return blockchain.NewEthereumSigner(keystore, passphrase)
}

func HuobiIntermediatorSignerFromFile(keystore, passphrase string) *blockchain.EthereumSigner {
	return blockchain.NewEthereumSigner(keystore, passphrase)
}
