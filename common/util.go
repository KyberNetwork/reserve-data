package common

import (
	"fmt"
	"math"
	"math/big"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ethereum/go-ethereum/common"
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

// CombineActivityStorageErrs return a combination of error between action error and storage error
func CombineActivityStorageErrs(err, sErr error) error {
	if err == nil && sErr == nil {
		return nil
	}
	if err != nil && sErr == nil {
		return err
	}
	if err == nil && sErr != nil {
		return sErr
	}
	return fmt.Errorf("action error: %v, storage error: %v", err, sErr)
}

// IsEthereumAddress returns true if the given address is ethereum.
func IsEthereumAddress(addr common.Address) bool {
	return addr == common.HexToAddress("0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee")
}

// GetCallerFunctionName return caller function name
func GetCallerFunctionName() string {
	fn := getFrame(2).Function
	lastSlash := strings.LastIndex(fn, "/")
	if lastSlash != -1 {
		return fn[lastSlash+1:]
	}
	return fn
}

func getFrame(skipFrames int) runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}

func CalculateNewPrice(pendingPrice float64, recommendPrice float64) float64 {
	switch {
	// if recommendPrice less than pending price, choose pending price as we can't override tx if use smaller price.
	case recommendPrice <= pendingPrice:
		return pendingPrice + (pendingPrice * 0.1) // increase 10% to prevent node to decline (I'm not sure on this)
	default:
		return math.Max(recommendPrice, pendingPrice*1.1) // recommendPrice should > pendingPrice at least 10% gwei
	}
}

// IsZeroAddress check if address is zero
func IsZeroAddress(a common.Address) bool {
	return a.Hash().Big().BitLen() == 0
}
