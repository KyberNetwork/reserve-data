package blockchain

import (
	"math/big"
	"reflect"
	"testing"

	ethereum "github.com/ethereum/go-ethereum/common"
)

type BigIntToCompactRateTestCase struct {
	RateInput *big.Int
	BaseInput *big.Int

	ExpectedBase     uint64
	ExpectedCompact  byte
	ExpectedOverflow bool
}

func testOutcome(testcase BigIntToCompactRateTestCase, t *testing.T) {
	rate, overflow := calculateCompactRate(
		testcase.RateInput,
		testcase.BaseInput,
	)

	if testcase.ExpectedBase != rate.Base.Uint64() ||
		testcase.ExpectedCompact != rate.Compact ||
		testcase.ExpectedOverflow != overflow {
		t.Fatalf("Expect CompactRate(%d, %d) overflow(%t), got CompactRate(%s, %d) overflow(%t)",
			testcase.ExpectedBase, testcase.ExpectedCompact, testcase.ExpectedOverflow,
			rate.Base.Text(10), rate.Compact, overflow,
		)
	}
}

func TestBigIntToCompactRate(t *testing.T) {
	testOutcome(BigIntToCompactRateTestCase{
		RateInput:        big.NewInt(40),
		BaseInput:        big.NewInt(30),
		ExpectedBase:     40,
		ExpectedCompact:  0,
		ExpectedOverflow: true,
	}, t)
	// test not overflow with positive compact
	testOutcome(BigIntToCompactRateTestCase{
		RateInput:        big.NewInt(31),
		BaseInput:        big.NewInt(30),
		ExpectedBase:     30,
		ExpectedCompact:  33,
		ExpectedOverflow: false,
	}, t)
	// test not overflow with max positive compact
	testOutcome(BigIntToCompactRateTestCase{
		RateInput:        big.NewInt(3381),
		BaseInput:        big.NewInt(3000),
		ExpectedBase:     3000,
		ExpectedCompact:  127,
		ExpectedOverflow: false,
	}, t)
	// test min overflow rate
	testOutcome(BigIntToCompactRateTestCase{
		RateInput:        big.NewInt(3384),
		BaseInput:        big.NewInt(3000),
		ExpectedBase:     3384,
		ExpectedCompact:  0,
		ExpectedOverflow: true,
	}, t)
	// test not overflow with negative compact
	testOutcome(BigIntToCompactRateTestCase{
		RateInput:        big.NewInt(2700),
		BaseInput:        big.NewInt(3000),
		ExpectedBase:     3000,
		ExpectedCompact:  156,
		ExpectedOverflow: false,
	}, t)
	// test not overflow min negative compact
	testOutcome(BigIntToCompactRateTestCase{
		RateInput:        big.NewInt(2616),
		BaseInput:        big.NewInt(3000),
		ExpectedBase:     3000,
		ExpectedCompact:  128,
		ExpectedOverflow: false,
	}, t)
	// test not overflow max negative compact
	testOutcome(BigIntToCompactRateTestCase{
		RateInput:        big.NewInt(2997),
		BaseInput:        big.NewInt(3000),
		ExpectedBase:     3000,
		ExpectedCompact:  255,
		ExpectedOverflow: false,
	}, t)
	// test overflow with negative compact
	testOutcome(BigIntToCompactRateTestCase{
		RateInput:        big.NewInt(1000),
		BaseInput:        big.NewInt(3000),
		ExpectedBase:     1000,
		ExpectedCompact:  0,
		ExpectedOverflow: true,
	}, t)
}

func TestBuildCompactBulk(t *testing.T) {
	addr1 := "0x14535eE720e329f66071B86486763Da4637034aE"
	addr2 := "0x24535eE720e329F66071b86486763da4637034AE"
	addr3 := "0x34535ee720e329f66071B86486763Da4637034aE"

	buysInput := map[ethereum.Address]byte{
		ethereum.HexToAddress(addr1): 23,
		ethereum.HexToAddress(addr2): 24,
		ethereum.HexToAddress(addr3): 25,
	}
	sellsInput := map[ethereum.Address]byte{
		ethereum.HexToAddress(addr1): 26,
		ethereum.HexToAddress(addr2): 27,
		ethereum.HexToAddress(addr3): 28,
	}
	indicesInput := map[string]tbindex{
		addr1: newTBIndex(3, 9),
		addr2: newTBIndex(9, 5),
		addr3: newTBIndex(9, 6),
	}
	buyBulk, sellBulk, indices := buildCompactBulk(
		buysInput, sellsInput, indicesInput)
	expectedBuys := []bulk{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 23, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 24, 25, 0, 0, 0, 0, 0, 0, 0},
	}
	expectedSells := []bulk{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 26, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 27, 28, 0, 0, 0, 0, 0, 0, 0},
	}
	expectedIndices := []*big.Int{
		big.NewInt(3),
		big.NewInt(9),
	}

	expectedBuys1 := []bulk{
		{0, 0, 0, 0, 0, 24, 25, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 23, 0, 0, 0, 0},
	}
	expectedSells1 := []bulk{
		{0, 0, 0, 0, 0, 27, 28, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 26, 0, 0, 0, 0},
	}
	expectedIndices1 := []*big.Int{
		big.NewInt(9),
		big.NewInt(3),
	}
	if !reflect.DeepEqual(expectedBuys, buyBulk) ||
		!reflect.DeepEqual(expectedSells, sellBulk) ||
		!reflect.DeepEqual(expectedIndices, indices) {
		if !reflect.DeepEqual(expectedBuys1, buyBulk) ||
			!reflect.DeepEqual(expectedSells1, sellBulk) ||
			!reflect.DeepEqual(expectedIndices1, indices) {
			t.Fatalf("Expected buys(%+v), sells(%+v), indices(%+v), got buys(%+v), sells(%+v), indices(%+v)",
				expectedBuys, expectedSells, expectedIndices,
				buyBulk, sellBulk, indices,
			)
		}
	}
}
