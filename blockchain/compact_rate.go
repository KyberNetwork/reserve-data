package blockchain

import (
	"math/big"
	"strconv"

	ethereum "github.com/ethereum/go-ethereum/common"
)

type compactRate struct {
	Base    *big.Int
	Compact byte
}

// Convert rate to compact rate with preferred base
// if it is impossible to use preferred base because Compact doesnt fit
// 8bits, the base is changed to the rate, Compact is set to 0 and
// return overflow = true
func calculateCompactRate(rate *big.Int, base *big.Int) (compactrate *compactRate, overflow bool) {
	if base.Cmp(big.NewInt(0)) == 0 {
		if rate.Cmp(big.NewInt(0)) == 0 {
			return &compactRate{rate, 0}, false
		}
		return &compactRate{rate, 0}, true
	}
	// rate = base * (1 + compact/1000) => compact = (rate / base - 1) * 1000
	// compat in range [-128, 127] but it represent for -12.8 -> 12.7 percent so we divide compat/1000 instead of 100
	fRate := new(big.Float).SetInt(rate)
	fBase := new(big.Float).SetInt(base)
	div := new(big.Float).Quo(fRate, fBase)
	div = new(big.Float).Add(div, big.NewFloat(-1.0))
	compact := new(big.Float).Mul(div, big.NewFloat(1000.0))
	// using text to round float
	str := compact.Text('f', 0)
	intComp, _ := strconv.ParseInt(str, 10, 64)
	if -128 <= intComp && intComp <= 127 {
		// capable to change compact
		return &compactRate{base, byte(intComp)}, false
	}
	// incapable to change compact, need to change base
	return &compactRate{rate, 0}, true
}

type bulk [14]byte

func buildCompactBulk(compatBuys, compatSells map[ethereum.Address]byte, indices map[string]tbindex) ([]bulk, []bulk, []*big.Int) {
	var buyBulkResults []bulk
	var sellBulkResults []bulk
	var bulkIndices []*big.Int
	buyBulks := map[uint64]bulk{}
	sellBulks := map[uint64]bulk{}

	// collect compat into bulk, group by bulk index
	for addr, buyCompact := range compatBuys {
		compatLoc := indices[addr.Hex()]
		// buy bulk
		b := buyBulks[compatLoc.BulkIndex]
		b[compatLoc.IndexInBulk] = buyCompact
		buyBulks[compatLoc.BulkIndex] = b

		// sell bulk
		b = sellBulks[compatLoc.BulkIndex]
		b[compatLoc.IndexInBulk] = compatSells[addr]
		sellBulks[compatLoc.BulkIndex] = b
	}
	// collect bulk into slice, match with bulk index position
	for index, buy := range buyBulks {
		buyBulkResults = append(buyBulkResults, buy)
		sellBulkResults = append(sellBulkResults, sellBulks[index])
		bulkIndices = append(bulkIndices, big.NewInt(int64(index)))
	}
	return buyBulkResults, sellBulkResults, bulkIndices
}
