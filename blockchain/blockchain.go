package blockchain

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	huobiblockchain "github.com/KyberNetwork/reserve-data/exchange/huobi/blockchain"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
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

// Blockchain object for interact with blockchain
type Blockchain struct {
	*blockchain.BaseBlockchain
	wrapper      *blockchain.Contract
	pricing      *blockchain.Contract
	reserve      *blockchain.Contract
	tokenIndices map[string]tbindex
	// listed tokens is all listed tokens on
	// our pricing contract
	listedTokens []ethereum.Address
	mu           sync.RWMutex

	localSetRateNonce, localDepositNonce         uint64
	setRateNonceTimestamp, depositNonceTimestamp uint64

	contractAddress *common.ContractAddressConfiguration
	sr              storage.SettingReader
	l               *zap.SugaredLogger
}

// ListedTokens return listed tokens from pricing contract
func (bc *Blockchain) ListedTokens() []ethereum.Address {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	res := make([]ethereum.Address, 0, len(bc.listedTokens))
	res = append(res, bc.listedTokens...)
	return res
}

// CheckTokenIndices check if token is listed
func (bc *Blockchain) CheckTokenIndices(tokenAddr ethereum.Address) error {
	opts := bc.GetCallOpts(0)
	pricingAddr := bc.contractAddress.Pricing
	tokenAddrs := []ethereum.Address{}
	tokenAddrs = append(tokenAddrs, tokenAddr)
	_, _, err := bc.GeneratedGetTokenIndicies(
		opts,
		pricingAddr,
		tokenAddrs,
	)
	if err != nil {
		return err
	}
	return nil
}

// LoadAndSetTokenIndices load and set token indices
func (bc *Blockchain) LoadAndSetTokenIndices() error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.tokenIndices = map[string]tbindex{}
	opts := bc.GetCallOpts(0)
	pricingAddr := bc.contractAddress.Pricing

	tokenAddrs, err := bc.getListedTokens()
	if err != nil {
		return err
	}
	bc.listedTokens = tokenAddrs

	bulkIndices, indicesInBulk, err := bc.GeneratedGetTokenIndicies(
		opts,
		pricingAddr,
		tokenAddrs,
	)
	if err != nil {
		return err
	}
	for i, tok := range tokenAddrs {
		bc.tokenIndices[tok.Hex()] = newTBIndex(
			bulkIndices[i].Uint64(),
			indicesInBulk[i].Uint64(),
		)
	}
	bc.l.Infof("Token indices: %+v", bc.tokenIndices)
	return nil
}

// RegisterPricingOperator add address as reserve operator for set token rates
func (bc *Blockchain) RegisterPricingOperator(signer blockchain.Signer, nonceCorpus blockchain.NonceCorpus) {
	bc.l.Infow("reserve pricing address", "address", signer.GetAddress().Hex())
	bc.MustRegisterOperator(blockchain.PricingOP, blockchain.NewOperator(signer, nonceCorpus))
}

// RegisterDepositOperator add address as depostor for reserve
func (bc *Blockchain) RegisterDepositOperator(signer blockchain.Signer, nonceCorpus blockchain.NonceCorpus) {
	bc.l.Infow("reserve depositor address", "address", signer.GetAddress().Hex())
	bc.MustRegisterOperator(blockchain.DepositOP, blockchain.NewOperator(signer, nonceCorpus))
}

func readablePrint(data map[ethereum.Address]byte) string {
	result := ""
	for addr, b := range data {
		result = result + "|" + fmt.Sprintf("%s-%d", addr.Hex(), b)
	}
	return result
}

//====================== Write calls ===============================

// SetRates set token rate to reserve
// we got a bug when compact is not set to old compact
// or when one of buy/sell got overflowed, it discards
// the other's compact
// TODO: Need better test coverage
func (bc *Blockchain) SetRates(
	tokens []ethereum.Address,
	buys []*big.Int,
	sells []*big.Int,
	block *big.Int,
	nonce *big.Int,
	gasPrice *big.Int) (*types.Transaction, error) {
	pricingAddr := bc.contractAddress.Pricing
	block.Add(block, big.NewInt(1))
	copts := bc.GetCallOpts(0)
	baseBuys, baseSells, _, _, _, err := bc.GeneratedGetTokenRates(
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
	bbuys, bsells, indices := BuildCompactBulk(
		newCBuys,
		newCSells,
		bc.tokenIndices,
	)
	opts, err := bc.GetTxOpts(blockchain.PricingOP, nonce, gasPrice, nil)
	if err != nil {
		bc.l.Infow("Getting transaction opts failed", "err", err)
		return nil, err
	}
	var tx *types.Transaction
	if len(baseTokens) > 0 {
		// set base tx
		tx, err = bc.GeneratedSetBaseRate(
			opts, baseTokens, newBBuys, newBSells,
			bbuys, bsells, block, indices)
		if tx != nil {
			bc.l.Infof(
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
		tx, err = bc.GeneratedSetCompactData(
			opts, bbuys, bsells, block, indices)
		if tx != nil {
			bc.l.Infof(
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
	return bc.SignAndBroadcast(tx, blockchain.PricingOP)

}

// Send withdraw token from reserve to another address (here is cex)
func (bc *Blockchain) Send(
	asset commonv3.Asset,
	amount *big.Int,
	dest ethereum.Address,
	nonce *big.Int,
	gasPrice *big.Int) (*types.Transaction, error) {

	opts, err := bc.GetTxOpts(blockchain.DepositOP, nonce, gasPrice, nil)
	if err != nil {
		return nil, err
	}
	tx, err := bc.GeneratedWithdraw(
		opts,
		asset.Address,
		amount,
		dest)
	if err != nil {
		bc.l.Errorw("failed to create tx", "err", err,
			"from", bc.MustGetOperator(blockchain.DepositOP).Address.String(),
			"asset", asset.Address.String(),
			"amount", amount.String(),
			"dest", dest.String(),
		)
		return nil, err
	}
	return bc.SignAndBroadcast(tx, blockchain.DepositOP)
}

// TransferToSelf use to override nonce
func (bc *Blockchain) TransferToSelf(op string, gasPrice *big.Int, nonce *big.Int) (*types.Transaction, error) {
	opAcc := bc.MustGetOperator(op)
	tx, err := bc.BuildSendETHTx(blockchain.TxOpts{
		Operator: opAcc,
		Nonce:    nonce,
		Value:    big.NewInt(0),
		GasPrice: gasPrice,
		GasLimit: 0,
	}, opAcc.Address)
	if err != nil {
		bc.l.Errorw("failed to create tx", "err", err,
			"from", opAcc.Address.String(),
		)
		return nil, err
	}
	return bc.SignAndBroadcast(tx, op)
}

//====================== Readonly calls ============================

// FetchBalanceData return token balance on reserve
func (bc *Blockchain) FetchBalanceData(reserve ethereum.Address, atBlock uint64) (map[rtypes.AssetID]common.BalanceEntry, error) {
	result := map[rtypes.AssetID]common.BalanceEntry{}
	tokens := []ethereum.Address{}
	allAssets, err := bc.sr.GetAssets()
	if err != nil {
		return result, err
	}
	assets := commonv3.AssetsHaveAddress(allAssets)
	for _, tok := range assets {
		tokens = append(tokens, tok.Address)
	}
	timestamp := common.GetTimestamp()
	opts := bc.GetCallOpts(atBlock)
	balances, err := bc.GeneratedGetBalances(opts, reserve, tokens)
	returnTime := common.GetTimestamp()
	bc.l.Infow("Fetcher ------> balances", "balances", balances, "err", err)
	if err != nil {
		for _, token := range assets {
			result[token.ID] = common.BalanceEntry{
				Valid:      false,
				Error:      err.Error(),
				Timestamp:  timestamp,
				ReturnTime: returnTime,
			}
		}
	} else {
		for i, asset := range assets {
			if balances[i].Cmp(Big0) == 0 || balances[i].Cmp(BigMax) > 0 {
				bc.l.Infow("balances of token is invalid", "token_symbol", asset.Symbol, "balances",
					balances[i].String(), "asset_id", asset.ID)
				result[asset.ID] = common.BalanceEntry{
					Valid:      false,
					Error:      "Got strange balances from node. It equals to 0 or is bigger than 10^33",
					Timestamp:  timestamp,
					ReturnTime: returnTime,
					Balance:    common.RawBalance(*balances[i]),
				}
			} else {
				result[asset.ID] = common.BalanceEntry{
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

// FetchRates return token rates with eth
func (bc *Blockchain) FetchRates(atBlock uint64, currentBlock uint64) (common.AllRateEntry, error) {
	result := common.AllRateEntry{}
	tokenAddrs := []ethereum.Address{}
	validTokens := []commonv3.Asset{}
	allAssets, err := bc.sr.GetAssets()
	if err != nil {
		return result, err
	}
	assets := commonv3.AssetsHaveSetRate(allAssets)
	listTokens := bc.ListedTokens()
	ltMap := make(map[ethereum.Address]struct{}, len(listTokens))
	for _, v := range listTokens {
		ltMap[v] = struct{}{}
	}
	for _, s := range assets {
		_, isListed := ltMap[s.Address]
		if !isListed {
			bc.l.Errorw("token/asset not listed and will be ignore from get rate",
				"asset", s.Symbol, "addr", s.Address.Hex())
		}
		// TODO: add a isETH method
		if !s.IsNetworkAsset() && isListed {
			tokenAddrs = append(tokenAddrs, s.Address)
			validTokens = append(validTokens, s)
		}
	}
	timestamp := common.GetTimestamp()
	opts := bc.GetCallOpts(atBlock)
	pricingAddr := bc.contractAddress.Pricing
	baseBuys, baseSells, compactBuys, compactSells, blocks, err := bc.GeneratedGetTokenRates(
		opts, pricingAddr, tokenAddrs,
	)
	if err != nil {
		return result, err
	}
	returnTime := common.GetTimestamp()
	result.Timestamp = timestamp
	result.ReturnTime = returnTime
	result.BlockNumber = currentBlock

	result.Data = map[rtypes.AssetID]common.RateEntry{}
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

// GetPrice return token rates
func (bc *Blockchain) GetPrice(token ethereum.Address, block *big.Int, priceType string, qty *big.Int, atBlock uint64) (*big.Int, error) {
	opts := bc.GetCallOpts(atBlock)
	if priceType == "buy" {
		return bc.GeneratedGetRate(opts, token, block, true, qty)
	}
	return bc.GeneratedGetRate(opts, token, block, false, qty)
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
func (bc *Blockchain) GetMinedNonceWithOP(op string) (uint64, error) {
	const localNonceExpiration = time.Minute * 2
	var localNonce, localTimestamp *uint64
	// base on op value, we bind selected nonce and timestamp field to local var for easier use it with below main logic
	switch op {
	case blockchain.PricingOP:
		localNonce, localTimestamp = &bc.localSetRateNonce, &bc.setRateNonceTimestamp
	case blockchain.DepositOP:
		localNonce, localTimestamp = &bc.localDepositNonce, &bc.depositNonceTimestamp
	default:
		return 0, fmt.Errorf("get minedNonce for unexpected op [%s]", op)
	}
	nonceFromNode, err := bc.GetMinedNonce(op)
	if err != nil {
		return nonceFromNode, err
	}
	if nonceFromNode < *localNonce {
		bc.l.Infow("nonce returned from node is smaller than cached nonce", "op", op,
			"node_value", nonceFromNode, "local_value", *localNonce)
		if common.NowInMillis()-*localTimestamp > uint64(localNonceExpiration/time.Millisecond) {
			bc.l.Infow("cached nonce stalled, overwriting with nonce from node", "op", op,
				"local_value", *localNonce, "node_value", nonceFromNode)
			*localNonce = nonceFromNode
			*localTimestamp = common.NowInMillis()
			return nonceFromNode, nil
		}
		bc.l.Infow("using cached nonce instead of nonce from node", "op", op,
			"local_value", *localNonce, "node_value", nonceFromNode)
		return *localNonce, nil
	}

	bc.l.Infow("updating local cached nonce", "op", op,
		"local_value", *localNonce, "node_value", nonceFromNode)
	*localNonce = nonceFromNode
	*localTimestamp = common.NowInMillis()
	return nonceFromNode, nil
}

// NewBlockchain return new blockchain object to to interact with blockchain
func NewBlockchain(base *blockchain.BaseBlockchain,
	contractAddressConf *common.ContractAddressConfiguration,
	sr storage.SettingReader,
) (*Blockchain, error) {
	l := zap.S()
	l.Infow("wrapper address", "address", contractAddressConf.Wrapper.Hex())
	wrapper := blockchain.NewContract(
		contractAddressConf.Wrapper,
		wrapperABI,
	)

	l.Infow("reserve address", "address", contractAddressConf.Reserve.Hex())
	reserve := blockchain.NewContract(
		contractAddressConf.Reserve,
		reserveABI,
	)

	l.Infow("pricing address", "address", contractAddressConf.Pricing.Hex())
	pricing := blockchain.NewContract(
		contractAddressConf.Pricing,
		pricingABI,
	)

	return &Blockchain{
		BaseBlockchain:  base,
		wrapper:         wrapper,
		pricing:         pricing,
		reserve:         reserve,
		contractAddress: contractAddressConf,
		sr:              sr,
		l:               l,
	}, nil
}

// GetPricingOPAddress return pricing op address
func (bc *Blockchain) GetPricingOPAddress() ethereum.Address {
	return bc.MustGetOperator(blockchain.PricingOP).Address
}

// GetDepositOPAddress return deposit op address
func (bc *Blockchain) GetDepositOPAddress() ethereum.Address {
	return bc.MustGetOperator(blockchain.DepositOP).Address
}

// GetIntermediatorOPAddress return intermediator op address
func (bc *Blockchain) GetIntermediatorOPAddress() ethereum.Address {
	return bc.MustGetOperator(huobiblockchain.HuobiOP).Address
}

// GetWrapperAddress return wrapper address
func (bc *Blockchain) GetWrapperAddress() ethereum.Address {
	return bc.contractAddress.Wrapper
}

// GetProxyAddress return network address
func (bc *Blockchain) GetProxyAddress() ethereum.Address {
	return bc.contractAddress.Proxy
}

// GetReserveAddress return reserve address
func (bc *Blockchain) GetReserveAddress() ethereum.Address {
	return bc.contractAddress.Reserve
}

// GetRateQueryHelperAddress return rateQueryHelperAddress
func (bc *Blockchain) GetRateQueryHelperAddress() ethereum.Address {
	return bc.contractAddress.RateQueryHelper
}

func (bc *Blockchain) getListedTokens() ([]ethereum.Address, error) {
	opts := bc.GetCallOpts(0)
	return bc.GeneratedGetListedTokens(opts)
}
