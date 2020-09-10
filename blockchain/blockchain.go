package blockchain

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	huobiblockchain "github.com/KyberNetwork/reserve-data/exchange/huobi/blockchain"
	"github.com/KyberNetwork/reserve-data/settings"
)

var (
	// Big0 zero in big.Int
	Big0 = big.NewInt(0)

	// BigMax max big.Int
	BigMax = big.NewInt(10).Exp(big.NewInt(10), big.NewInt(33), nil)
)

// tbindex is where the token data stored in blockchain.
// In blockchain, data of a token (sell/buy rates) is stored in an array of 32 bytes values called (tokenRatesCompactData).
// Each data is stored in a byte.
// https://github.com/KyberNetwork/smart-contracts/blob/fed8e09dc6e4365e1597474d9b3f53634eb405d2/contracts/ConversionRates.sol#L48
type tbindex struct {
	// BulkIndex is the index of bytes32 value that store data of multiple tokens.
	BulkIndex uint64
	// IndexInBulk is the index in the above BulkIndex value where the sell/buy rates are stored following structure:
	// sell: IndexInBulk + 4
	// buy: IndexInBulk + 8
	IndexInBulk uint64
}

// newTBIndex creates new tbindex instance with given parameters.
func newTBIndex(bulkIndex, indexInBulk uint64) tbindex {
	return tbindex{BulkIndex: bulkIndex, IndexInBulk: indexInBulk}
}

// Blockchain object to interact with blockchain in reserve core
type Blockchain struct {
	*blockchain.BaseBlockchain
	wrapper      *blockchain.Contract
	pricing      *blockchain.Contract
	reserve      *blockchain.Contract
	tokenIndices map[string]tbindex
	// ListedTokens is for check fill zero as delisted token
	// should have zero rate
	listedTokens []ethereum.Address
	mu           sync.RWMutex

	localSetRateNonce, localDepositNonce         uint64
	setRateNonceTimestamp, depositNonceTimestamp uint64

	setting Setting
	l       *zap.SugaredLogger
}

//ListedTokens return listed tokens
func (b *Blockchain) ListedTokens() []ethereum.Address {
	return b.listedTokens
}

// CheckTokenIndices check token indices
func (b *Blockchain) CheckTokenIndices(tokenAddr ethereum.Address) error {
	opts := b.GetCallOpts(0)
	pricingAddr, err := b.setting.GetAddress(settings.Pricing)
	if err != nil {
		return err
	}
	tokenAddrs := []ethereum.Address{}
	tokenAddrs = append(tokenAddrs, tokenAddr)
	_, _, err = b.GeneratedGetTokenIndicies(
		opts,
		pricingAddr,
		tokenAddrs,
	)
	return err
}

// LoadAndSetTokenIndices load and set token indices
func (b *Blockchain) LoadAndSetTokenIndices() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tokenIndices = map[string]tbindex{}
	// this is not really needed. Just a safe guard. Use a very big indices so it would not exist.
	b.tokenIndices[ethereum.HexToAddress(b.setting.ETHToken().Address).Hex()] = tbindex{1000000, 1000000}
	opts := b.GetCallOpts(0)
	pricingAddr, err := b.setting.GetAddress(settings.Pricing)
	if err != nil {
		return err
	}

	// we used to load token indices for only internal token
	// in setting database, but as we need to set rate for delisted token
	// to 0 also, then we decided load all listed tokens
	tokenAddrs, err := b.getListedTokensFromPricingContract()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to load token list using contract %s", pricingAddr.String()))
	}
	b.listedTokens = tokenAddrs

	bulkIndices, indicesInBulk, err := b.GeneratedGetTokenIndicies(
		opts,
		pricingAddr,
		tokenAddrs,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to load token index using contract %s", pricingAddr.String()))
	}
	for i, tok := range tokenAddrs {
		b.tokenIndices[tok.Hex()] = newTBIndex(
			bulkIndices[i].Uint64(),
			indicesInBulk[i].Uint64(),
		)
	}
	b.l.Infof("Token indices: %+v", b.tokenIndices)
	return nil
}

// RegisterPricingOperator register pricing operator
func (b *Blockchain) RegisterPricingOperator(signer blockchain.Signer, nonceCorpus blockchain.NonceCorpus) {
	b.l.Infof("reserve pricing address: %s", signer.GetAddress().Hex())
	b.MustRegisterOperator(blockchain.PricingOP, blockchain.NewOperator(signer, nonceCorpus))
}

// RegisterDepositOperator register operator
func (b *Blockchain) RegisterDepositOperator(signer blockchain.Signer, nonceCorpus blockchain.NonceCorpus) {
	b.l.Infof("reserve depositor address: %s", signer.GetAddress().Hex())
	b.MustRegisterOperator(blockchain.DepositOP, blockchain.NewOperator(signer, nonceCorpus))
}

func readablePrint(data map[ethereum.Address]byte) string {
	result := ""
	for addr, b := range data {
		result = result + "|" + fmt.Sprintf("%s-%d", addr.Hex(), b)
	}
	return result
}

//====================== Write calls ===============================

// SetRates set token rates
// we got a bug when compact is not set to old compact
// or when one of buy/sell got overflowed, it discards
// the other's compact
// TODO: Need better test coverage
func (b *Blockchain) SetRates(
	tokens []ethereum.Address,
	buys []*big.Int,
	sells []*big.Int,
	block *big.Int,
	nonce *big.Int,
	gasPrice *big.Int) (*types.Transaction, error) {
	pricingAddr, err := b.setting.GetAddress(settings.Pricing)
	if err != nil {
		return nil, err
	}
	block.Add(block, big.NewInt(1))
	copts := b.GetCallOpts(0)
	baseBuys, baseSells, _, _, _, err := b.GeneratedGetTokenRates(
		copts, pricingAddr, tokens,
	)
	if err != nil {
		return nil, err
	}

	// This is commented out because we dont want to make too much of change. Don't remove
	// this check, it can be useful in the future.
	//
	// Don't submit any txs if it is just trying to set all tokens to 0 when they are already 0
	// if common.AllZero(buys, sells, baseBuys, baseSells) {
	// 	return nil, errors.New("Trying to set all rates to 0 but they are already 0. Skip the tx.")
	// }

	baseTokens := []ethereum.Address{}
	newBSells := []*big.Int{}
	newBBuys := []*big.Int{}
	newCSells := map[ethereum.Address]byte{}
	newCBuys := map[ethereum.Address]byte{}
	for i, token := range tokens {
		compactSell, overflow1 := BigIntToCompactRate(sells[i], baseSells[i])
		compactBuy, overflow2 := BigIntToCompactRate(buys[i], baseBuys[i])
		if overflow1 || overflow2 {
			baseTokens = append(baseTokens, token)
			newBSells = append(newBSells, sells[i])
			newBBuys = append(newBBuys, buys[i])
			newCSells[token] = 0
			newCBuys[token] = 0
		} else {
			newCSells[token] = compactSell.Compact
			newCBuys[token] = compactBuy.Compact
		}
	}
	bbuys, bsells, indices, err := BuildCompactBulk(
		newCBuys,
		newCSells,
		b.tokenIndices,
	)
	if err != nil {
		b.l.Warnw("failed to build compact bulk", "err", err)
		return nil, err
	}
	opts, err := b.GetTxOpts(blockchain.PricingOP, nonce, gasPrice, nil)
	if err != nil {
		b.l.Warnw("Getting transaction opts failed", "err", err)
		return nil, err
	}
	var tx *types.Transaction
	if len(baseTokens) > 0 {
		// set base tx
		tx, err = b.GeneratedSetBaseRate(
			opts, baseTokens, newBBuys, newBSells,
			bbuys, bsells, block, indices)
		if tx != nil {
			b.l.Infof(
				"broadcasting setbase tx %s, target buys(%s), target sells(%s), old base buy(%s) || old base sell(%s) || new base buy(%s) || new base sell(%s) || new compact buy(%s) || new compact sell(%s) || new buy bulk(%v) || new sell bulk(%v) || indices(%v)",
				tx.Hash().Hex(),
				buys, sells,
				baseBuys, baseSells,
				newBBuys, newBSells,
				readablePrint(newCBuys), readablePrint(newCSells),
				bbuys, bsells, indices,
			)
		}
	} else {
		// update compact tx
		tx, err = b.GeneratedSetCompactData(
			opts, bbuys, bsells, block, indices)
		if tx != nil {
			b.l.Infof(
				"broadcasting setcompact tx %s, target buys(%s), target sells(%s), old base buy(%s) || old base sell(%s) || new compact buy(%s) || new compact sell(%s) || new buy bulk(%v) || new sell bulk(%v) || indices(%v)",
				tx.Hash().Hex(),
				buys, sells,
				baseBuys, baseSells,
				readablePrint(newCBuys), readablePrint(newCSells),
				bbuys, bsells, indices,
			)
		}
	}
	if err != nil {
		return nil, err
	}
	return b.SignAndBroadcast(tx, blockchain.PricingOP)

}

// Send request to blockchain
func (b *Blockchain) Send(token common.Token, amount *big.Int, dest ethereum.Address, nonce *big.Int,
	gasPrice *big.Int) (*types.Transaction, error) {

	opts, err := b.GetTxOpts(blockchain.DepositOP, nonce, gasPrice, nil)
	if err != nil {
		return nil, err
	}
	tx, err := b.GeneratedWithdraw(
		opts,
		ethereum.HexToAddress(token.Address),
		amount, dest)
	if err != nil {
		return nil, err
	}
	return b.SignAndBroadcast(tx, blockchain.DepositOP)
}

//====================== Readonly calls ============================

// FetchBalanceData return reserve balance at a block
func (b *Blockchain) FetchBalanceData(reserve ethereum.Address, atBlock uint64) (map[string]common.BalanceEntry, error) {
	result := map[string]common.BalanceEntry{}
	tokens := []ethereum.Address{}
	tokensSetting, err := b.setting.GetInternalTokens()
	if err != nil {
		return result, err
	}
	for _, tok := range tokensSetting {
		tokens = append(tokens, ethereum.HexToAddress(tok.Address))
	}
	timestamp := common.GetTimestamp()
	opts := b.GetCallOpts(atBlock)
	balances, err := b.GeneratedGetBalances(opts, reserve, tokens)
	returnTime := common.GetTimestamp()
	b.l.Infof("Fetcher ------> balances: %v, err: %s", balances, common.ErrorToString(err))
	if err != nil {
		for _, token := range tokensSetting {
			result[token.ID] = common.BalanceEntry{
				Valid:      false,
				Error:      err.Error(),
				Timestamp:  timestamp,
				ReturnTime: returnTime,
			}
		}
	} else {
		for i, tok := range tokensSetting {
			if balances[i].Cmp(Big0) == 0 || balances[i].Cmp(BigMax) > 0 {
				b.l.Infof("Fetcher ------> balances of token %s is invalid", tok.ID)
				result[tok.ID] = common.BalanceEntry{
					Valid:      false,
					Error:      "Got strange balances from node. It equals to 0 or is bigger than 10^33",
					Timestamp:  timestamp,
					ReturnTime: returnTime,
					Balance:    common.RawBalance(*balances[i]),
				}
			} else {
				result[tok.ID] = common.BalanceEntry{
					Valid:      true,
					Timestamp:  timestamp,
					ReturnTime: returnTime,
					Balance:    common.RawBalance(*balances[i]),
				}
			}
		}
	}
	return result, nil
}

// FetchRates return all token rates
func (b *Blockchain) FetchRates(atBlock uint64, currentBlock uint64) (common.AllRateEntry, error) {
	result := common.AllRateEntry{}
	tokenAddrs := []ethereum.Address{}
	validTokens := []common.Token{}
	tokenSettings, err := b.setting.GetInternalTokens()
	if err != nil {
		return result, err
	}
	for _, s := range tokenSettings {
		if s.ID != "ETH" {
			tokenAddrs = append(tokenAddrs, ethereum.HexToAddress(s.Address))
			validTokens = append(validTokens, s)
		}
	}
	timestamp := common.GetTimestamp()
	opts := b.GetCallOpts(atBlock)
	pricingAddr, err := b.setting.GetAddress(settings.Pricing)
	if err != nil {
		return result, err
	}
	baseBuys, baseSells, compactBuys, compactSells, blocks, err := b.GeneratedGetTokenRates(
		opts, pricingAddr, tokenAddrs,
	)
	if err != nil {
		return result, err
	}
	returnTime := common.GetTimestamp()
	result.Timestamp = timestamp
	result.ReturnTime = returnTime
	result.BlockNumber = currentBlock

	result.Data = map[string]common.RateEntry{}
	for i, token := range validTokens {
		result.Data[token.ID] = common.NewRateEntry(
			baseBuys[i],
			compactBuys[i],
			baseSells[i],
			compactSells[i],
			blocks[i].Uint64(),
		)
	}
	return result, nil
}

// GetPrice return token rate
func (b *Blockchain) GetPrice(token ethereum.Address, block *big.Int, priceType string, qty *big.Int, atBlock uint64) (*big.Int, error) {
	opts := b.GetCallOpts(atBlock)
	if priceType == "buy" {
		return b.GeneratedGetRate(opts, token, block, true, qty)
	}
	return b.GeneratedGetRate(opts, token, block, false, qty)
}

// GetMinedNonceWithOP returns nonce of the pricing operator in confirmed
// state (not pending state).
//
// Getting mined nonce is not simple because there might be lag between
// node leading us to get outdated mined nonce from an unsynced node.
// To overcome this situation, we will keep a local nonce and require
// the nonce from node to be equal or greater than it.
// If the nonce from node is smaller than the local one, we will use
// the local one. However, if the local one stay with the same value
// for more than 15 mins, the local one is considered incorrect
// because the chain might be reorganized so we will invalidate it
// and assign it to the nonce from node.
func (b *Blockchain) GetMinedNonceWithOP(op string) (uint64, error) {
	const localNonceExpiration = time.Minute * 2
	var localNonce, localTimestamp *uint64
	// base on op value, we bind selected nonce and timestamp field to local var for easier use it with below main logic
	switch op {
	case blockchain.PricingOP:
		localNonce, localTimestamp = &b.localSetRateNonce, &b.setRateNonceTimestamp
	case blockchain.DepositOP:
		localNonce, localTimestamp = &b.localDepositNonce, &b.depositNonceTimestamp
	default:
		return 0, fmt.Errorf("get minedNonce for unexpected op [%s]", op)
	}
	nonceFromNode, err := b.GetMinedNonce(op)
	if err != nil {
		return nonceFromNode, err
	}
	if nonceFromNode < *localNonce {
		b.l.Infow("nonce returned from node is smaller than cached nonce", "op", op,
			"node_value", nonceFromNode, "local_value", *localNonce)
		if common.GetTimepoint()-*localTimestamp > uint64(localNonceExpiration/time.Millisecond) {
			b.l.Infow("cached nonce stalled, overwriting with nonce from node", "op", op,
				"local_value", *localNonce, "node_value", nonceFromNode)
			*localNonce = nonceFromNode
			*localTimestamp = common.GetTimepoint()
			return nonceFromNode, nil
		}
		b.l.Infow("using cached nonce instead of nonce from node", "op", op,
			"local_value", *localNonce, "node_value", nonceFromNode)
		return *localNonce, nil
	}

	b.l.Infow("updating local cached nonce", "op", op,
		"local_value", *localNonce, "node_value", nonceFromNode)
	*localNonce = nonceFromNode
	*localTimestamp = common.GetTimepoint()
	return nonceFromNode, nil
}

// NewBlockchain return new blockchain object
func NewBlockchain(base *blockchain.BaseBlockchain, setting Setting) (*Blockchain, error) {
	wrapperAddr, err := setting.GetAddress(settings.Wrapper)
	if err != nil {
		return nil, err
	}
	l := zap.S()
	l.Infof("wrapper address: %s", wrapperAddr.Hex())
	wrapper := blockchain.NewContract(
		wrapperAddr,
		blockchain.WrapperABI,
	)
	reserveAddr, err := setting.GetAddress(settings.Reserve)
	if err != nil {
		return nil, err
	}
	l.Infof("reserve address: %s", reserveAddr.Hex())
	reserve := blockchain.NewContract(
		reserveAddr,
		blockchain.ReserveABI,
	)
	pricingAddr, err := setting.GetAddress(settings.Pricing)
	if err != nil {
		return nil, err
	}
	l.Infof("pricing address: %s", pricingAddr.Hex())
	pricing := blockchain.NewContract(
		pricingAddr,
		blockchain.PricingABI,
	)

	return &Blockchain{
		BaseBlockchain: base,
		wrapper:        wrapper,
		pricing:        pricing,
		reserve:        reserve,
		setting:        setting,
		l:              l,
	}, nil
}

// GetPricingOPAddress return pricing operator address
func (b *Blockchain) GetPricingOPAddress() ethereum.Address {
	return b.MustGetOperator(blockchain.PricingOP).Address
}

// GetDepositOPAddress return deposit operator address
func (b *Blockchain) GetDepositOPAddress() ethereum.Address {
	return b.MustGetOperator(blockchain.DepositOP).Address
}

// GetIntermediatorOPAddress return intermediator operator address
func (b *Blockchain) GetIntermediatorOPAddress() ethereum.Address {
	return b.MustGetOperator(huobiblockchain.HuobiOP).Address
}

// getListedTokensFromPricingContract return listed token in reserve contract
func (b *Blockchain) getListedTokensFromPricingContract() ([]ethereum.Address, error) {
	opts := b.GetCallOpts(0)
	return b.GeneratedGetListedTokens(opts)
}
