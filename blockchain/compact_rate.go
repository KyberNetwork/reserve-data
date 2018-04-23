package blockchain

import (
	"math/big"
	"strconv"

	ethereum "github.com/ethereum/go-ethereum/common"
)

type CompactRate struct {
	Base    *big.Int
	Compact byte
}

// Convert rate to compact rate with preferred base
// if it is impossible to use preferred base because Compact doesnt fit
// 8bits, the base is changed to the rate, Compact is set to 0 and
// return overflow = true
func BigIntToCompactRate(rate *big.Int, base *big.Int) (compactrate *CompactRate, overflow bool) {
	if base.Cmp(big.NewInt(0)) == 0 {
		if rate.Cmp(big.NewInt(0)) == 0 {
			return &CompactRate{
				rate, 0,
			}, false
		} else {
			return &CompactRate{
				rate, 0,
			}, true
		}
	}
	// rate = base * (1 + compact/1000)
	// compact = (rate / base - 1) * 1000
	fRate := big.NewFloat(0).SetInt(rate)
	fBase := big.NewFloat(0).SetInt(base)
	div := big.NewFloat(0).Quo(fRate, fBase)
	div = div.Add(div, big.NewFloat(-1.0))
	compact := big.NewFloat(0).Mul(div, big.NewFloat(1000.0))
	// using text to round float
	str := compact.Text('f', 0)
	intComp, _ := strconv.ParseInt(str, 10, 64)
	if -128 <= intComp && intComp <= 127 {
		// capable to change compact
		return &CompactRate{
			base, byte(intComp),
		}, false
	} else {
		// incapable to change compact, need to change base
		return &CompactRate{
			rate, 0,
		}, true
	}
}

type bulk struct {
	data [14]byte
}

func BuildCompactBulk(newBuys, newSells map[ethereum.Address]byte, indices map[string]tbindex) ([][14]byte, [][14]byte, []*big.Int) {
	buyResults := [][14]byte{}
	sellResults := [][14]byte{}
	indexResults := []*big.Int{}
	buyBulks := map[uint64]*bulk{}
	sellBulks := map[uint64]*bulk{}
	for addr, buyCompact := range newBuys {
		index := indices[addr.Hex()]
		_, exist := buyBulks[index.BulkIndex]
		if !exist {
			buyBulks[index.BulkIndex] = &bulk{}
			sellBulks[index.BulkIndex] = &bulk{}
		}
		b := buyBulks[index.BulkIndex]
		b.data[index.IndexInBulk] = buyCompact
		b = sellBulks[index.BulkIndex]
		b.data[index.IndexInBulk] = newSells[addr]
	}
	for index, buy := range buyBulks {
		buyResults = append(buyResults, buy.data)
		sellResults = append(sellResults, sellBulks[index].data)
		indexResults = append(indexResults, big.NewInt(int64(index)))
	}
	return buyResults, sellResults, indexResults
}
