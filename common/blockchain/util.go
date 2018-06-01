package blockchain

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"os"

	"github.com/KyberNetwork/reserve-data/common"
	ether "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	return hexutil.EncodeBig(number)
}

func toFilterArg(q ether.FilterQuery) interface{} {
	arg := map[string]interface{}{
		"fromBlock": toBlockNumArg(q.FromBlock),
		"toBlock":   toBlockNumArg(q.ToBlock),
		"address":   q.Addresses,
		"topics":    q.Topics,
	}
	if q.FromBlock == nil {
		arg["fromBlock"] = "0x0"
	}
	return arg
}

// ensureContext is a helper method to ensure a context is not nil, even if the
// user specified it as such.
func ensureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.TODO()
	}
	return ctx
}

// generateDevelopmentKeystore generates a random keystore for use in development and simulation.
// Don't use it for production purpose.
func generateDevelopmentKeystore(path, passphrase string) ([]byte, error) {
	const (
		// veryLightScryptN is the N parameter of Scrypt algorithm, use for keystore encryption
		veryLightScryptN = 2
		// veryLightScryptP is the p parameter of Scrypt algorithm, use for keystore encryption
		veryLightScryptP = 1
	)
	tmpDir, err := ioutil.TempDir("", "development-keystore")
	if err != nil {
		return nil, err
	}
	ks := keystore.NewKeyStore(tmpDir, veryLightScryptN, veryLightScryptP)
	if ks == nil {
		return nil, errors.New("failed to create keystore")
	}
	acc, err := ks.NewAccount(passphrase)
	if err != nil {
		return nil, err
	}
	keyJSON, err := ks.Export(acc, passphrase, passphrase)
	if err != nil {
		return nil, err
	}
	if err = ioutil.WriteFile(path, keyJSON, 0600); err != nil {
		return nil, err
	}
	return keyJSON, nil
}

// GenerateDevelopmentKeystoreIfNotExists generated a new keystore if the error in
// parameter indicates the file not exists and the running mode is simulation or
// development. In other cases, it returns the given error.
func GenerateDevelopmentKeystoreIfNotExists(openErr error, path, passphrase string) (io.Reader, error) {
	modes := map[string]struct{}{
		common.DEV_MODE:          {},
		common.SIMULATION_MODE:   {},
		common.ANALYTIC_DEV_MODE: {},
	}

	if _, ok := modes[common.RunningMode()]; !ok {
		return nil, openErr
	}

	if !os.IsNotExist(openErr) {
		return nil, openErr
	}

	keyJSON, err := generateDevelopmentKeystore(path, passphrase)
	if err != nil {
		return nil, err
	}
	log.Printf("development keystore generated: %s", path)
	return bytes.NewReader(keyJSON), nil
}
