package coinbase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var sampleOrderbook = []Price{
	Price{
		Size:  10,
		Price: 200,
	},
	Price{
		Size:  20,
		Price: 210,
	},
	Price{
		Size:  30,
		Price: 220,
	},
	Price{
		Size:  10,
		Price: 230,
	},
}

type TestCase struct {
	amount  float64
	isBid   bool
	afp     float64
	isError bool
}

func TestCalculateAfp(t *testing.T) {
	tests := []TestCase{
		TestCase{
			amount:  1,
			isBid:   true,
			isError: false,
			afp:     230,
		},
		TestCase{
			amount:  1,
			isBid:   false,
			isError: false,
			afp:     200,
		},
		TestCase{
			amount:  4100,
			isBid:   false,
			isError: false,
			afp:     205,
		},
		TestCase{
			amount:  4500,
			isBid:   true,
			isError: false,
			afp:     225,
		},
		TestCase{
			amount:  20000,
			isBid:   true,
			isError: true,
			afp:     0,
		},
		TestCase{
			amount:  20000,
			isBid:   false,
			isError: true,
			afp:     0,
		},
	}
	for _, test := range tests {
		afp, err := getAfp(sampleOrderbook, test.amount, test.isBid)
		if test.isError {
			assert.Error(t, err)
		} else {
			assert.Equal(t, test.afp, afp)
		}
	}
}
