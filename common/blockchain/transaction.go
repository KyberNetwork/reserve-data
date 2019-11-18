package blockchain

import (
	"encoding/json"
	"errors"
	"log"
	"math/big"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

type RPCTransaction struct {
	tx *types.Transaction
	txExtraInfo
}

type txExtraInfo struct {
	BlockNumber *string
	BlockHash   *ethereum.Hash
	From        ethereum.Address
}

func (tx *RPCTransaction) UnmarshalJSON(msg []byte) error {
	if err := json.Unmarshal(msg, &tx.tx); err != nil {
		return err
	}
	return json.Unmarshal(msg, &tx.txExtraInfo)
}

func (tx *RPCTransaction) BlockNumber() *big.Int {
	if tx.txExtraInfo.BlockNumber == nil {
		return big.NewInt(0)
	}
	blockno, err := hexutil.DecodeBig(*tx.txExtraInfo.BlockNumber)
	if err != nil {
		log.Printf("Error decoding block number: %v", err)
		return big.NewInt(0)
	}
	return blockno
}

// senderFromServer is a types.Signer that remembers the sender address returned by the RPC
// server. It is stored in the transaction's sender address cache to avoid an additional
// request in TransactionSender.
type senderFromServer struct {
	addr      ethereum.Address
	blockhash ethereum.Hash
}

var errNotCached = errors.New("sender not cached")

func setSenderFromServer(tx *types.Transaction, addr ethereum.Address, block ethereum.Hash) {
	// Use types.Sender for side-effect to store our signer into the cache.
	if _, err := types.Sender(&senderFromServer{addr, block}, tx); err != nil {
		log.Printf("Type sender error: %s", err.Error())
	}
}

func (s *senderFromServer) Equal(other types.Signer) bool {
	os, ok := other.(*senderFromServer)
	return ok && os.blockhash == s.blockhash
}

func (s *senderFromServer) Sender(tx *types.Transaction) (ethereum.Address, error) {
	if s.blockhash == (ethereum.Hash{}) {
		return ethereum.Address{}, errNotCached
	}
	return s.addr, nil
}

// Hash implementation is only for satisfy the interface definition. senderFromServer shouldn't be signing anything.
func (s *senderFromServer) Hash(tx *types.Transaction) ethereum.Hash {
	// This should never happen
	panic("can't sign with senderFromServer")
}

// SignatureValues implementation is only for satisfy the interface definition. senderFromServer shouldn't be signing anything.
func (s *senderFromServer) SignatureValues(tx *types.Transaction, sig []byte) (_, _, _ *big.Int, err error) {
	// This should never happen
	panic("can't sign with senderFromServer")
}
