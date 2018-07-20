package common

import (
	"math"
	"math/big"
	"path"
	"path/filepath"
	"runtime"
)

const truncLength int = 256

func TruncStr(src []byte) []byte {
	if len(src) > truncLength {
		result := string(src[0:truncLength]) + "..."
		return []byte(result)
	}
	return src
}

// CurrentDir returns current directory of the caller.
func CurrentDir() string {
	_, current, _, _ := runtime.Caller(1)
	return filepath.Join(path.Dir(current))
}

// CmdDirLocation returns the absolute location of cmd directory where
// public settings will be read.
func CmdDirLocation() string {
	_, fileName, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filepath.Dir(fileName)), "cmd")
}

// ErrorToString returns error as string and an empty string if the error is nil
func ErrorToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// FloatToBigInt converts a float to a big int with specific decimal
// Example:
// - FloatToBigInt(1, 4) = 10000
// - FloatToBigInt(1.234, 4) = 12340
func FloatToBigInt(amount float64, decimal int64) *big.Int {
	// 6 is our smallest precision
	if decimal < 6 {
		return big.NewInt(int64(amount * math.Pow10(int(decimal))))
	}
	result := big.NewInt(int64(amount * math.Pow10(6)))
	return result.Mul(result, big.NewInt(0).Exp(big.NewInt(10), big.NewInt(decimal-6), nil))
}

// BigToFloat converts a big int to float according to its number of decimal digits
// Example:
// - BigToFloat(1100, 3) = 1.1
// - BigToFloat(1100, 2) = 11
// - BigToFloat(1100, 5) = 0.11
func BigToFloat(b *big.Int, decimal int64) float64 {
	f := new(big.Float).SetInt(b)
	power := new(big.Float).SetInt(new(big.Int).Exp(
		big.NewInt(10), big.NewInt(decimal), nil,
	))
	res := new(big.Float).Quo(f, power)
	result, _ := res.Float64()
	return result
}

// GweiToWei converts Gwei as a float to Wei as a big int
func GweiToWei(n float64) *big.Int {
	return FloatToBigInt(n, 9)
}

// EthToWei converts Gwei as a float to Wei as a big int
func EthToWei(n float64) *big.Int {
	return FloatToBigInt(n, 18)
}

// AllZero returns true if all params are zero, false otherwise
// func AllZero(sets ...[]*big.Int) bool {
// 	big0 := big.NewInt(0)
// 	for _, set := range sets {
// 		for _, i := range set {
// 			if i.Cmp(big0) != 0 {
// 				return false
// 			}
// 		}
// 	}
// 	return true
// }
