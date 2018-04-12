package blockchain

type EthUSDRate interface {
	GetUSDRate(timepoint uint64) float64
}
