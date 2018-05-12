// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package blockchain

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// PricingABI is the input ABI used to generate the binding from.
const PricingABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"alerter\",\"type\":\"address\"}],\"name\":\"removeAlerter\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"reserve\",\"type\":\"address\"}],\"name\":\"setReserveAddress\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"}],\"name\":\"disableTokenTrade\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"validRateDurationInBlocks\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"tokens\",\"type\":\"address[]\"},{\"name\":\"baseBuy\",\"type\":\"uint256[]\"},{\"name\":\"baseSell\",\"type\":\"uint256[]\"},{\"name\":\"buy\",\"type\":\"bytes14[]\"},{\"name\":\"sell\",\"type\":\"bytes14[]\"},{\"name\":\"blockNumber\",\"type\":\"uint256\"},{\"name\":\"indices\",\"type\":\"uint256[]\"}],\"name\":\"setBaseRate\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"}],\"name\":\"enableTokenTrade\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"pendingAdmin\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOperators\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getListedTokens\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"sendTo\",\"type\":\"address\"}],\"name\":\"withdrawToken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newAlerter\",\"type\":\"address\"}],\"name\":\"addAlerter\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"numTokensInCurrentCompactData\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"command\",\"type\":\"uint256\"},{\"name\":\"param\",\"type\":\"uint256\"}],\"name\":\"getStepFunctionData\",\"outputs\":[{\"name\":\"\",\"type\":\"int256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"buy\",\"type\":\"bytes14[]\"},{\"name\":\"sell\",\"type\":\"bytes14[]\"},{\"name\":\"blockNumber\",\"type\":\"uint256\"},{\"name\":\"indices\",\"type\":\"uint256[]\"}],\"name\":\"setCompactData\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"duration\",\"type\":\"uint256\"}],\"name\":\"setValidRateDurationInBlocks\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"}],\"name\":\"getTokenBasicData\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"},{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"transferAdmin\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"claimAdmin\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getAlerters\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"}],\"name\":\"getRateUpdateBlock\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"xBuy\",\"type\":\"int256[]\"},{\"name\":\"yBuy\",\"type\":\"int256[]\"},{\"name\":\"xSell\",\"type\":\"int256[]\"},{\"name\":\"ySell\",\"type\":\"int256[]\"}],\"name\":\"setQtyStepFunction\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOperator\",\"type\":\"address\"}],\"name\":\"addOperator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"reserveContract\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"tokenImbalanceData\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"removeOperator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"currentBlockNumber\",\"type\":\"uint256\"},{\"name\":\"buy\",\"type\":\"bool\"},{\"name\":\"qty\",\"type\":\"uint256\"}],\"name\":\"getRate\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"xBuy\",\"type\":\"int256[]\"},{\"name\":\"yBuy\",\"type\":\"int256[]\"},{\"name\":\"xSell\",\"type\":\"int256[]\"},{\"name\":\"ySell\",\"type\":\"int256[]\"}],\"name\":\"setImbalanceStepFunction\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"minimalRecordResolution\",\"type\":\"uint256\"},{\"name\":\"maxPerBlockImbalance\",\"type\":\"uint256\"},{\"name\":\"maxTotalImbalance\",\"type\":\"uint256\"}],\"name\":\"setTokenControlInfo\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"buyAmount\",\"type\":\"int256\"},{\"name\":\"rateUpdateBlock\",\"type\":\"uint256\"},{\"name\":\"currentBlock\",\"type\":\"uint256\"}],\"name\":\"recordImbalance\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"sendTo\",\"type\":\"address\"}],\"name\":\"withdrawEther\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"buy\",\"type\":\"bool\"}],\"name\":\"getBasicRate\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"}],\"name\":\"addToken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"}],\"name\":\"getCompactData\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"bytes1\"},{\"name\":\"\",\"type\":\"bytes1\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"}],\"name\":\"getTokenControlInfo\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"admin\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_admin\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"sendTo\",\"type\":\"address\"}],\"name\":\"TokenWithdraw\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"sendTo\",\"type\":\"address\"}],\"name\":\"EtherWithdraw\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"pendingAdmin\",\"type\":\"address\"}],\"name\":\"TransferAdminPending\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"newAdmin\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"previousAdmin\",\"type\":\"address\"}],\"name\":\"AdminClaimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"newAlerter\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"isAdd\",\"type\":\"bool\"}],\"name\":\"AlerterAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"newOperator\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"isAdd\",\"type\":\"bool\"}],\"name\":\"OperatorAdded\",\"type\":\"event\"}]"

// Pricing is an auto generated Go binding around an Ethereum contract.
type Pricing struct {
	PricingCaller     // Read-only binding to the contract
	PricingTransactor // Write-only binding to the contract
}

// PricingCaller is an auto generated read-only Go binding around an Ethereum contract.
type PricingCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PricingTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PricingTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PricingSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PricingSession struct {
	Contract     *Pricing          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PricingCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PricingCallerSession struct {
	Contract *PricingCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// PricingTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PricingTransactorSession struct {
	Contract     *PricingTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// PricingRaw is an auto generated low-level Go binding around an Ethereum contract.
type PricingRaw struct {
	Contract *Pricing // Generic contract binding to access the raw methods on
}

// PricingCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PricingCallerRaw struct {
	Contract *PricingCaller // Generic read-only contract binding to access the raw methods on
}

// PricingTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PricingTransactorRaw struct {
	Contract *PricingTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPricing creates a new instance of Pricing, bound to a specific deployed contract.
func NewPricing(address common.Address, backend bind.ContractBackend) (*Pricing, error) {
	contract, err := bindPricing(address, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Pricing{PricingCaller: PricingCaller{contract: contract}, PricingTransactor: PricingTransactor{contract: contract}}, nil
}

// NewPricingCaller creates a new read-only instance of Pricing, bound to a specific deployed contract.
func NewPricingCaller(address common.Address, caller bind.ContractCaller) (*PricingCaller, error) {
	contract, err := bindPricing(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &PricingCaller{contract: contract}, nil
}

// NewPricingTransactor creates a new write-only instance of Pricing, bound to a specific deployed contract.
func NewPricingTransactor(address common.Address, transactor bind.ContractTransactor) (*PricingTransactor, error) {
	contract, err := bindPricing(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &PricingTransactor{contract: contract}, nil
}

// bindPricing binds a generic wrapper to an already deployed contract.
func bindPricing(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(PricingABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, nil), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Pricing *PricingRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Pricing.Contract.PricingCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Pricing *PricingRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Pricing.Contract.PricingTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Pricing *PricingRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Pricing.Contract.PricingTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Pricing *PricingCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Pricing.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Pricing *PricingTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Pricing.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Pricing *PricingTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Pricing.Contract.contract.Transact(opts, method, params...)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() constant returns(address)
func (_Pricing *PricingCaller) Admin(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Pricing.contract.Call(opts, out, "admin")
	return *ret0, err
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() constant returns(address)
func (_Pricing *PricingSession) Admin() (common.Address, error) {
	return _Pricing.Contract.Admin(&_Pricing.CallOpts)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() constant returns(address)
func (_Pricing *PricingCallerSession) Admin() (common.Address, error) {
	return _Pricing.Contract.Admin(&_Pricing.CallOpts)
}

// GetAlerters is a free data retrieval call binding the contract method 0x7c423f54.
//
// Solidity: function getAlerters() constant returns(address[])
func (_Pricing *PricingCaller) GetAlerters(opts *bind.CallOpts) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _Pricing.contract.Call(opts, out, "getAlerters")
	return *ret0, err
}

// GetAlerters is a free data retrieval call binding the contract method 0x7c423f54.
//
// Solidity: function getAlerters() constant returns(address[])
func (_Pricing *PricingSession) GetAlerters() ([]common.Address, error) {
	return _Pricing.Contract.GetAlerters(&_Pricing.CallOpts)
}

// GetAlerters is a free data retrieval call binding the contract method 0x7c423f54.
//
// Solidity: function getAlerters() constant returns(address[])
func (_Pricing *PricingCallerSession) GetAlerters() ([]common.Address, error) {
	return _Pricing.Contract.GetAlerters(&_Pricing.CallOpts)
}

// GetBasicRate is a free data retrieval call binding the contract method 0xcf8fee11.
//
// Solidity: function getBasicRate(token address, buy bool) constant returns(uint256)
func (_Pricing *PricingCaller) GetBasicRate(opts *bind.CallOpts, token common.Address, buy bool) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Pricing.contract.Call(opts, out, "getBasicRate", token, buy)
	return *ret0, err
}

// GetBasicRate is a free data retrieval call binding the contract method 0xcf8fee11.
//
// Solidity: function getBasicRate(token address, buy bool) constant returns(uint256)
func (_Pricing *PricingSession) GetBasicRate(token common.Address, buy bool) (*big.Int, error) {
	return _Pricing.Contract.GetBasicRate(&_Pricing.CallOpts, token, buy)
}

// GetBasicRate is a free data retrieval call binding the contract method 0xcf8fee11.
//
// Solidity: function getBasicRate(token address, buy bool) constant returns(uint256)
func (_Pricing *PricingCallerSession) GetBasicRate(token common.Address, buy bool) (*big.Int, error) {
	return _Pricing.Contract.GetBasicRate(&_Pricing.CallOpts, token, buy)
}

// GetCompactData is a free data retrieval call binding the contract method 0xe4a2ac62.
//
// Solidity: function getCompactData(token address) constant returns(uint256, uint256, bytes1, bytes1)
func (_Pricing *PricingCaller) GetCompactData(opts *bind.CallOpts, token common.Address) (*big.Int, *big.Int, [1]byte, [1]byte, error) {
	var (
		ret0 = new(*big.Int)
		ret1 = new(*big.Int)
		ret2 = new([1]byte)
		ret3 = new([1]byte)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
		ret3,
	}
	err := _Pricing.contract.Call(opts, out, "getCompactData", token)
	return *ret0, *ret1, *ret2, *ret3, err
}

// GetCompactData is a free data retrieval call binding the contract method 0xe4a2ac62.
//
// Solidity: function getCompactData(token address) constant returns(uint256, uint256, bytes1, bytes1)
func (_Pricing *PricingSession) GetCompactData(token common.Address) (*big.Int, *big.Int, [1]byte, [1]byte, error) {
	return _Pricing.Contract.GetCompactData(&_Pricing.CallOpts, token)
}

// GetCompactData is a free data retrieval call binding the contract method 0xe4a2ac62.
//
// Solidity: function getCompactData(token address) constant returns(uint256, uint256, bytes1, bytes1)
func (_Pricing *PricingCallerSession) GetCompactData(token common.Address) (*big.Int, *big.Int, [1]byte, [1]byte, error) {
	return _Pricing.Contract.GetCompactData(&_Pricing.CallOpts, token)
}

// GetListedTokens is a free data retrieval call binding the contract method 0x2ba996a5.
//
// Solidity: function getListedTokens() constant returns(address[])
func (_Pricing *PricingCaller) GetListedTokens(opts *bind.CallOpts) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _Pricing.contract.Call(opts, out, "getListedTokens")
	return *ret0, err
}

// GetListedTokens is a free data retrieval call binding the contract method 0x2ba996a5.
//
// Solidity: function getListedTokens() constant returns(address[])
func (_Pricing *PricingSession) GetListedTokens() ([]common.Address, error) {
	return _Pricing.Contract.GetListedTokens(&_Pricing.CallOpts)
}

// GetListedTokens is a free data retrieval call binding the contract method 0x2ba996a5.
//
// Solidity: function getListedTokens() constant returns(address[])
func (_Pricing *PricingCallerSession) GetListedTokens() ([]common.Address, error) {
	return _Pricing.Contract.GetListedTokens(&_Pricing.CallOpts)
}

// GetOperators is a free data retrieval call binding the contract method 0x27a099d8.
//
// Solidity: function getOperators() constant returns(address[])
func (_Pricing *PricingCaller) GetOperators(opts *bind.CallOpts) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _Pricing.contract.Call(opts, out, "getOperators")
	return *ret0, err
}

// GetOperators is a free data retrieval call binding the contract method 0x27a099d8.
//
// Solidity: function getOperators() constant returns(address[])
func (_Pricing *PricingSession) GetOperators() ([]common.Address, error) {
	return _Pricing.Contract.GetOperators(&_Pricing.CallOpts)
}

// GetOperators is a free data retrieval call binding the contract method 0x27a099d8.
//
// Solidity: function getOperators() constant returns(address[])
func (_Pricing *PricingCallerSession) GetOperators() ([]common.Address, error) {
	return _Pricing.Contract.GetOperators(&_Pricing.CallOpts)
}

// GetRate is a free data retrieval call binding the contract method 0xb8e9c22e.
//
// Solidity: function getRate(token address, currentBlockNumber uint256, buy bool, qty uint256) constant returns(uint256)
func (_Pricing *PricingCaller) GetRate(opts *bind.CallOpts, token common.Address, currentBlockNumber *big.Int, buy bool, qty *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Pricing.contract.Call(opts, out, "getRate", token, currentBlockNumber, buy, qty)
	return *ret0, err
}

// GetRate is a free data retrieval call binding the contract method 0xb8e9c22e.
//
// Solidity: function getRate(token address, currentBlockNumber uint256, buy bool, qty uint256) constant returns(uint256)
func (_Pricing *PricingSession) GetRate(token common.Address, currentBlockNumber *big.Int, buy bool, qty *big.Int) (*big.Int, error) {
	return _Pricing.Contract.GetRate(&_Pricing.CallOpts, token, currentBlockNumber, buy, qty)
}

// GetRate is a free data retrieval call binding the contract method 0xb8e9c22e.
//
// Solidity: function getRate(token address, currentBlockNumber uint256, buy bool, qty uint256) constant returns(uint256)
func (_Pricing *PricingCallerSession) GetRate(token common.Address, currentBlockNumber *big.Int, buy bool, qty *big.Int) (*big.Int, error) {
	return _Pricing.Contract.GetRate(&_Pricing.CallOpts, token, currentBlockNumber, buy, qty)
}

// GetRateUpdateBlock is a free data retrieval call binding the contract method 0x8036d757.
//
// Solidity: function getRateUpdateBlock(token address) constant returns(uint256)
func (_Pricing *PricingCaller) GetRateUpdateBlock(opts *bind.CallOpts, token common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Pricing.contract.Call(opts, out, "getRateUpdateBlock", token)
	return *ret0, err
}

// GetRateUpdateBlock is a free data retrieval call binding the contract method 0x8036d757.
//
// Solidity: function getRateUpdateBlock(token address) constant returns(uint256)
func (_Pricing *PricingSession) GetRateUpdateBlock(token common.Address) (*big.Int, error) {
	return _Pricing.Contract.GetRateUpdateBlock(&_Pricing.CallOpts, token)
}

// GetRateUpdateBlock is a free data retrieval call binding the contract method 0x8036d757.
//
// Solidity: function getRateUpdateBlock(token address) constant returns(uint256)
func (_Pricing *PricingCallerSession) GetRateUpdateBlock(token common.Address) (*big.Int, error) {
	return _Pricing.Contract.GetRateUpdateBlock(&_Pricing.CallOpts, token)
}

// GetStepFunctionData is a free data retrieval call binding the contract method 0x62674e93.
//
// Solidity: function getStepFunctionData(token address, command uint256, param uint256) constant returns(int256)
func (_Pricing *PricingCaller) GetStepFunctionData(opts *bind.CallOpts, token common.Address, command *big.Int, param *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Pricing.contract.Call(opts, out, "getStepFunctionData", token, command, param)
	return *ret0, err
}

// GetStepFunctionData is a free data retrieval call binding the contract method 0x62674e93.
//
// Solidity: function getStepFunctionData(token address, command uint256, param uint256) constant returns(int256)
func (_Pricing *PricingSession) GetStepFunctionData(token common.Address, command *big.Int, param *big.Int) (*big.Int, error) {
	return _Pricing.Contract.GetStepFunctionData(&_Pricing.CallOpts, token, command, param)
}

// GetStepFunctionData is a free data retrieval call binding the contract method 0x62674e93.
//
// Solidity: function getStepFunctionData(token address, command uint256, param uint256) constant returns(int256)
func (_Pricing *PricingCallerSession) GetStepFunctionData(token common.Address, command *big.Int, param *big.Int) (*big.Int, error) {
	return _Pricing.Contract.GetStepFunctionData(&_Pricing.CallOpts, token, command, param)
}

// GetTokenBasicData is a free data retrieval call binding the contract method 0x721bba59.
//
// Solidity: function getTokenBasicData(token address) constant returns(bool, bool)
func (_Pricing *PricingCaller) GetTokenBasicData(opts *bind.CallOpts, token common.Address) (bool, bool, error) {
	var (
		ret0 = new(bool)
		ret1 = new(bool)
	)
	out := &[]interface{}{
		ret0,
		ret1,
	}
	err := _Pricing.contract.Call(opts, out, "getTokenBasicData", token)
	return *ret0, *ret1, err
}

// GetTokenBasicData is a free data retrieval call binding the contract method 0x721bba59.
//
// Solidity: function getTokenBasicData(token address) constant returns(bool, bool)
func (_Pricing *PricingSession) GetTokenBasicData(token common.Address) (bool, bool, error) {
	return _Pricing.Contract.GetTokenBasicData(&_Pricing.CallOpts, token)
}

// GetTokenBasicData is a free data retrieval call binding the contract method 0x721bba59.
//
// Solidity: function getTokenBasicData(token address) constant returns(bool, bool)
func (_Pricing *PricingCallerSession) GetTokenBasicData(token common.Address) (bool, bool, error) {
	return _Pricing.Contract.GetTokenBasicData(&_Pricing.CallOpts, token)
}

// GetTokenControlInfo is a free data retrieval call binding the contract method 0xe7d4fd91.
//
// Solidity: function getTokenControlInfo(token address) constant returns(uint256, uint256, uint256)
func (_Pricing *PricingCaller) GetTokenControlInfo(opts *bind.CallOpts, token common.Address) (*big.Int, *big.Int, *big.Int, error) {
	var (
		ret0 = new(*big.Int)
		ret1 = new(*big.Int)
		ret2 = new(*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
	}
	err := _Pricing.contract.Call(opts, out, "getTokenControlInfo", token)
	return *ret0, *ret1, *ret2, err
}

// GetTokenControlInfo is a free data retrieval call binding the contract method 0xe7d4fd91.
//
// Solidity: function getTokenControlInfo(token address) constant returns(uint256, uint256, uint256)
func (_Pricing *PricingSession) GetTokenControlInfo(token common.Address) (*big.Int, *big.Int, *big.Int, error) {
	return _Pricing.Contract.GetTokenControlInfo(&_Pricing.CallOpts, token)
}

// GetTokenControlInfo is a free data retrieval call binding the contract method 0xe7d4fd91.
//
// Solidity: function getTokenControlInfo(token address) constant returns(uint256, uint256, uint256)
func (_Pricing *PricingCallerSession) GetTokenControlInfo(token common.Address) (*big.Int, *big.Int, *big.Int, error) {
	return _Pricing.Contract.GetTokenControlInfo(&_Pricing.CallOpts, token)
}

// NumTokensInCurrentCompactData is a free data retrieval call binding the contract method 0x5085c9f1.
//
// Solidity: function numTokensInCurrentCompactData() constant returns(uint256)
func (_Pricing *PricingCaller) NumTokensInCurrentCompactData(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Pricing.contract.Call(opts, out, "numTokensInCurrentCompactData")
	return *ret0, err
}

// NumTokensInCurrentCompactData is a free data retrieval call binding the contract method 0x5085c9f1.
//
// Solidity: function numTokensInCurrentCompactData() constant returns(uint256)
func (_Pricing *PricingSession) NumTokensInCurrentCompactData() (*big.Int, error) {
	return _Pricing.Contract.NumTokensInCurrentCompactData(&_Pricing.CallOpts)
}

// NumTokensInCurrentCompactData is a free data retrieval call binding the contract method 0x5085c9f1.
//
// Solidity: function numTokensInCurrentCompactData() constant returns(uint256)
func (_Pricing *PricingCallerSession) NumTokensInCurrentCompactData() (*big.Int, error) {
	return _Pricing.Contract.NumTokensInCurrentCompactData(&_Pricing.CallOpts)
}

// PendingAdmin is a free data retrieval call binding the contract method 0x26782247.
//
// Solidity: function pendingAdmin() constant returns(address)
func (_Pricing *PricingCaller) PendingAdmin(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Pricing.contract.Call(opts, out, "pendingAdmin")
	return *ret0, err
}

// PendingAdmin is a free data retrieval call binding the contract method 0x26782247.
//
// Solidity: function pendingAdmin() constant returns(address)
func (_Pricing *PricingSession) PendingAdmin() (common.Address, error) {
	return _Pricing.Contract.PendingAdmin(&_Pricing.CallOpts)
}

// PendingAdmin is a free data retrieval call binding the contract method 0x26782247.
//
// Solidity: function pendingAdmin() constant returns(address)
func (_Pricing *PricingCallerSession) PendingAdmin() (common.Address, error) {
	return _Pricing.Contract.PendingAdmin(&_Pricing.CallOpts)
}

// ReserveContract is a free data retrieval call binding the contract method 0xa7f43acd.
//
// Solidity: function reserveContract() constant returns(address)
func (_Pricing *PricingCaller) ReserveContract(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Pricing.contract.Call(opts, out, "reserveContract")
	return *ret0, err
}

// ReserveContract is a free data retrieval call binding the contract method 0xa7f43acd.
//
// Solidity: function reserveContract() constant returns(address)
func (_Pricing *PricingSession) ReserveContract() (common.Address, error) {
	return _Pricing.Contract.ReserveContract(&_Pricing.CallOpts)
}

// ReserveContract is a free data retrieval call binding the contract method 0xa7f43acd.
//
// Solidity: function reserveContract() constant returns(address)
func (_Pricing *PricingCallerSession) ReserveContract() (common.Address, error) {
	return _Pricing.Contract.ReserveContract(&_Pricing.CallOpts)
}

// TokenImbalanceData is a free data retrieval call binding the contract method 0xa80c609e.
//
// Solidity: function tokenImbalanceData( address,  uint256) constant returns(uint256)
func (_Pricing *PricingCaller) TokenImbalanceData(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Pricing.contract.Call(opts, out, "tokenImbalanceData", arg0, arg1)
	return *ret0, err
}

// TokenImbalanceData is a free data retrieval call binding the contract method 0xa80c609e.
//
// Solidity: function tokenImbalanceData( address,  uint256) constant returns(uint256)
func (_Pricing *PricingSession) TokenImbalanceData(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _Pricing.Contract.TokenImbalanceData(&_Pricing.CallOpts, arg0, arg1)
}

// TokenImbalanceData is a free data retrieval call binding the contract method 0xa80c609e.
//
// Solidity: function tokenImbalanceData( address,  uint256) constant returns(uint256)
func (_Pricing *PricingCallerSession) TokenImbalanceData(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _Pricing.Contract.TokenImbalanceData(&_Pricing.CallOpts, arg0, arg1)
}

// ValidRateDurationInBlocks is a free data retrieval call binding the contract method 0x16265694.
//
// Solidity: function validRateDurationInBlocks() constant returns(uint256)
func (_Pricing *PricingCaller) ValidRateDurationInBlocks(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Pricing.contract.Call(opts, out, "validRateDurationInBlocks")
	return *ret0, err
}

// ValidRateDurationInBlocks is a free data retrieval call binding the contract method 0x16265694.
//
// Solidity: function validRateDurationInBlocks() constant returns(uint256)
func (_Pricing *PricingSession) ValidRateDurationInBlocks() (*big.Int, error) {
	return _Pricing.Contract.ValidRateDurationInBlocks(&_Pricing.CallOpts)
}

// ValidRateDurationInBlocks is a free data retrieval call binding the contract method 0x16265694.
//
// Solidity: function validRateDurationInBlocks() constant returns(uint256)
func (_Pricing *PricingCallerSession) ValidRateDurationInBlocks() (*big.Int, error) {
	return _Pricing.Contract.ValidRateDurationInBlocks(&_Pricing.CallOpts)
}

// AddAlerter is a paid mutator transaction binding the contract method 0x408ee7fe.
//
// Solidity: function addAlerter(newAlerter address) returns()
func (_Pricing *PricingTransactor) AddAlerter(opts *bind.TransactOpts, newAlerter common.Address) (*types.Transaction, error) {
	return _Pricing.contract.Transact(opts, "addAlerter", newAlerter)
}

// AddAlerter is a paid mutator transaction binding the contract method 0x408ee7fe.
//
// Solidity: function addAlerter(newAlerter address) returns()
func (_Pricing *PricingSession) AddAlerter(newAlerter common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.AddAlerter(&_Pricing.TransactOpts, newAlerter)
}

// AddAlerter is a paid mutator transaction binding the contract method 0x408ee7fe.
//
// Solidity: function addAlerter(newAlerter address) returns()
func (_Pricing *PricingTransactorSession) AddAlerter(newAlerter common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.AddAlerter(&_Pricing.TransactOpts, newAlerter)
}

// AddOperator is a paid mutator transaction binding the contract method 0x9870d7fe.
//
// Solidity: function addOperator(newOperator address) returns()
func (_Pricing *PricingTransactor) AddOperator(opts *bind.TransactOpts, newOperator common.Address) (*types.Transaction, error) {
	return _Pricing.contract.Transact(opts, "addOperator", newOperator)
}

// AddOperator is a paid mutator transaction binding the contract method 0x9870d7fe.
//
// Solidity: function addOperator(newOperator address) returns()
func (_Pricing *PricingSession) AddOperator(newOperator common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.AddOperator(&_Pricing.TransactOpts, newOperator)
}

// AddOperator is a paid mutator transaction binding the contract method 0x9870d7fe.
//
// Solidity: function addOperator(newOperator address) returns()
func (_Pricing *PricingTransactorSession) AddOperator(newOperator common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.AddOperator(&_Pricing.TransactOpts, newOperator)
}

// AddToken is a paid mutator transaction binding the contract method 0xd48bfca7.
//
// Solidity: function addToken(token address) returns()
func (_Pricing *PricingTransactor) AddToken(opts *bind.TransactOpts, token common.Address) (*types.Transaction, error) {
	return _Pricing.contract.Transact(opts, "addToken", token)
}

// AddToken is a paid mutator transaction binding the contract method 0xd48bfca7.
//
// Solidity: function addToken(token address) returns()
func (_Pricing *PricingSession) AddToken(token common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.AddToken(&_Pricing.TransactOpts, token)
}

// AddToken is a paid mutator transaction binding the contract method 0xd48bfca7.
//
// Solidity: function addToken(token address) returns()
func (_Pricing *PricingTransactorSession) AddToken(token common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.AddToken(&_Pricing.TransactOpts, token)
}

// ClaimAdmin is a paid mutator transaction binding the contract method 0x77f50f97.
//
// Solidity: function claimAdmin() returns()
func (_Pricing *PricingTransactor) ClaimAdmin(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Pricing.contract.Transact(opts, "claimAdmin")
}

// ClaimAdmin is a paid mutator transaction binding the contract method 0x77f50f97.
//
// Solidity: function claimAdmin() returns()
func (_Pricing *PricingSession) ClaimAdmin() (*types.Transaction, error) {
	return _Pricing.Contract.ClaimAdmin(&_Pricing.TransactOpts)
}

// ClaimAdmin is a paid mutator transaction binding the contract method 0x77f50f97.
//
// Solidity: function claimAdmin() returns()
func (_Pricing *PricingTransactorSession) ClaimAdmin() (*types.Transaction, error) {
	return _Pricing.Contract.ClaimAdmin(&_Pricing.TransactOpts)
}

// DisableTokenTrade is a paid mutator transaction binding the contract method 0x158859f7.
//
// Solidity: function disableTokenTrade(token address) returns()
func (_Pricing *PricingTransactor) DisableTokenTrade(opts *bind.TransactOpts, token common.Address) (*types.Transaction, error) {
	return _Pricing.contract.Transact(opts, "disableTokenTrade", token)
}

// DisableTokenTrade is a paid mutator transaction binding the contract method 0x158859f7.
//
// Solidity: function disableTokenTrade(token address) returns()
func (_Pricing *PricingSession) DisableTokenTrade(token common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.DisableTokenTrade(&_Pricing.TransactOpts, token)
}

// DisableTokenTrade is a paid mutator transaction binding the contract method 0x158859f7.
//
// Solidity: function disableTokenTrade(token address) returns()
func (_Pricing *PricingTransactorSession) DisableTokenTrade(token common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.DisableTokenTrade(&_Pricing.TransactOpts, token)
}

// EnableTokenTrade is a paid mutator transaction binding the contract method 0x1d6a8bda.
//
// Solidity: function enableTokenTrade(token address) returns()
func (_Pricing *PricingTransactor) EnableTokenTrade(opts *bind.TransactOpts, token common.Address) (*types.Transaction, error) {
	return _Pricing.contract.Transact(opts, "enableTokenTrade", token)
}

// EnableTokenTrade is a paid mutator transaction binding the contract method 0x1d6a8bda.
//
// Solidity: function enableTokenTrade(token address) returns()
func (_Pricing *PricingSession) EnableTokenTrade(token common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.EnableTokenTrade(&_Pricing.TransactOpts, token)
}

// EnableTokenTrade is a paid mutator transaction binding the contract method 0x1d6a8bda.
//
// Solidity: function enableTokenTrade(token address) returns()
func (_Pricing *PricingTransactorSession) EnableTokenTrade(token common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.EnableTokenTrade(&_Pricing.TransactOpts, token)
}

// RecordImbalance is a paid mutator transaction binding the contract method 0xc6fd2103.
//
// Solidity: function recordImbalance(token address, buyAmount int256, rateUpdateBlock uint256, currentBlock uint256) returns()
func (_Pricing *PricingTransactor) RecordImbalance(opts *bind.TransactOpts, token common.Address, buyAmount *big.Int, rateUpdateBlock *big.Int, currentBlock *big.Int) (*types.Transaction, error) {
	return _Pricing.contract.Transact(opts, "recordImbalance", token, buyAmount, rateUpdateBlock, currentBlock)
}

// RecordImbalance is a paid mutator transaction binding the contract method 0xc6fd2103.
//
// Solidity: function recordImbalance(token address, buyAmount int256, rateUpdateBlock uint256, currentBlock uint256) returns()
func (_Pricing *PricingSession) RecordImbalance(token common.Address, buyAmount *big.Int, rateUpdateBlock *big.Int, currentBlock *big.Int) (*types.Transaction, error) {
	return _Pricing.Contract.RecordImbalance(&_Pricing.TransactOpts, token, buyAmount, rateUpdateBlock, currentBlock)
}

// RecordImbalance is a paid mutator transaction binding the contract method 0xc6fd2103.
//
// Solidity: function recordImbalance(token address, buyAmount int256, rateUpdateBlock uint256, currentBlock uint256) returns()
func (_Pricing *PricingTransactorSession) RecordImbalance(token common.Address, buyAmount *big.Int, rateUpdateBlock *big.Int, currentBlock *big.Int) (*types.Transaction, error) {
	return _Pricing.Contract.RecordImbalance(&_Pricing.TransactOpts, token, buyAmount, rateUpdateBlock, currentBlock)
}

// RemoveAlerter is a paid mutator transaction binding the contract method 0x01a12fd3.
//
// Solidity: function removeAlerter(alerter address) returns()
func (_Pricing *PricingTransactor) RemoveAlerter(opts *bind.TransactOpts, alerter common.Address) (*types.Transaction, error) {
	return _Pricing.contract.Transact(opts, "removeAlerter", alerter)
}

// RemoveAlerter is a paid mutator transaction binding the contract method 0x01a12fd3.
//
// Solidity: function removeAlerter(alerter address) returns()
func (_Pricing *PricingSession) RemoveAlerter(alerter common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.RemoveAlerter(&_Pricing.TransactOpts, alerter)
}

// RemoveAlerter is a paid mutator transaction binding the contract method 0x01a12fd3.
//
// Solidity: function removeAlerter(alerter address) returns()
func (_Pricing *PricingTransactorSession) RemoveAlerter(alerter common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.RemoveAlerter(&_Pricing.TransactOpts, alerter)
}

// RemoveOperator is a paid mutator transaction binding the contract method 0xac8a584a.
//
// Solidity: function removeOperator(operator address) returns()
func (_Pricing *PricingTransactor) RemoveOperator(opts *bind.TransactOpts, operator common.Address) (*types.Transaction, error) {
	return _Pricing.contract.Transact(opts, "removeOperator", operator)
}

// RemoveOperator is a paid mutator transaction binding the contract method 0xac8a584a.
//
// Solidity: function removeOperator(operator address) returns()
func (_Pricing *PricingSession) RemoveOperator(operator common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.RemoveOperator(&_Pricing.TransactOpts, operator)
}

// RemoveOperator is a paid mutator transaction binding the contract method 0xac8a584a.
//
// Solidity: function removeOperator(operator address) returns()
func (_Pricing *PricingTransactorSession) RemoveOperator(operator common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.RemoveOperator(&_Pricing.TransactOpts, operator)
}

// SetBaseRate is a paid mutator transaction binding the contract method 0x1a4813d7.
//
// Solidity: function setBaseRate(tokens address[], baseBuy uint256[], baseSell uint256[], buy bytes14[], sell bytes14[], blockNumber uint256, indices uint256[]) returns()
func (_Pricing *PricingTransactor) SetBaseRate(opts *bind.TransactOpts, tokens []common.Address, baseBuy []*big.Int, baseSell []*big.Int, buy [][14]byte, sell [][14]byte, blockNumber *big.Int, indices []*big.Int) (*types.Transaction, error) {
	return _Pricing.contract.Transact(opts, "setBaseRate", tokens, baseBuy, baseSell, buy, sell, blockNumber, indices)
}

// SetBaseRate is a paid mutator transaction binding the contract method 0x1a4813d7.
//
// Solidity: function setBaseRate(tokens address[], baseBuy uint256[], baseSell uint256[], buy bytes14[], sell bytes14[], blockNumber uint256, indices uint256[]) returns()
func (_Pricing *PricingSession) SetBaseRate(tokens []common.Address, baseBuy []*big.Int, baseSell []*big.Int, buy [][14]byte, sell [][14]byte, blockNumber *big.Int, indices []*big.Int) (*types.Transaction, error) {
	return _Pricing.Contract.SetBaseRate(&_Pricing.TransactOpts, tokens, baseBuy, baseSell, buy, sell, blockNumber, indices)
}

// SetBaseRate is a paid mutator transaction binding the contract method 0x1a4813d7.
//
// Solidity: function setBaseRate(tokens address[], baseBuy uint256[], baseSell uint256[], buy bytes14[], sell bytes14[], blockNumber uint256, indices uint256[]) returns()
func (_Pricing *PricingTransactorSession) SetBaseRate(tokens []common.Address, baseBuy []*big.Int, baseSell []*big.Int, buy [][14]byte, sell [][14]byte, blockNumber *big.Int, indices []*big.Int) (*types.Transaction, error) {
	return _Pricing.Contract.SetBaseRate(&_Pricing.TransactOpts, tokens, baseBuy, baseSell, buy, sell, blockNumber, indices)
}

// SetCompactData is a paid mutator transaction binding the contract method 0x64887334.
//
// Solidity: function setCompactData(buy bytes14[], sell bytes14[], blockNumber uint256, indices uint256[]) returns()
func (_Pricing *PricingTransactor) SetCompactData(opts *bind.TransactOpts, buy [][14]byte, sell [][14]byte, blockNumber *big.Int, indices []*big.Int) (*types.Transaction, error) {
	return _Pricing.contract.Transact(opts, "setCompactData", buy, sell, blockNumber, indices)
}

// SetCompactData is a paid mutator transaction binding the contract method 0x64887334.
//
// Solidity: function setCompactData(buy bytes14[], sell bytes14[], blockNumber uint256, indices uint256[]) returns()
func (_Pricing *PricingSession) SetCompactData(buy [][14]byte, sell [][14]byte, blockNumber *big.Int, indices []*big.Int) (*types.Transaction, error) {
	return _Pricing.Contract.SetCompactData(&_Pricing.TransactOpts, buy, sell, blockNumber, indices)
}

// SetCompactData is a paid mutator transaction binding the contract method 0x64887334.
//
// Solidity: function setCompactData(buy bytes14[], sell bytes14[], blockNumber uint256, indices uint256[]) returns()
func (_Pricing *PricingTransactorSession) SetCompactData(buy [][14]byte, sell [][14]byte, blockNumber *big.Int, indices []*big.Int) (*types.Transaction, error) {
	return _Pricing.Contract.SetCompactData(&_Pricing.TransactOpts, buy, sell, blockNumber, indices)
}

// SetImbalanceStepFunction is a paid mutator transaction binding the contract method 0xbc9cbcc8.
//
// Solidity: function setImbalanceStepFunction(token address, xBuy int256[], yBuy int256[], xSell int256[], ySell int256[]) returns()
func (_Pricing *PricingTransactor) SetImbalanceStepFunction(opts *bind.TransactOpts, token common.Address, xBuy []*big.Int, yBuy []*big.Int, xSell []*big.Int, ySell []*big.Int) (*types.Transaction, error) {
	return _Pricing.contract.Transact(opts, "setImbalanceStepFunction", token, xBuy, yBuy, xSell, ySell)
}

// SetImbalanceStepFunction is a paid mutator transaction binding the contract method 0xbc9cbcc8.
//
// Solidity: function setImbalanceStepFunction(token address, xBuy int256[], yBuy int256[], xSell int256[], ySell int256[]) returns()
func (_Pricing *PricingSession) SetImbalanceStepFunction(token common.Address, xBuy []*big.Int, yBuy []*big.Int, xSell []*big.Int, ySell []*big.Int) (*types.Transaction, error) {
	return _Pricing.Contract.SetImbalanceStepFunction(&_Pricing.TransactOpts, token, xBuy, yBuy, xSell, ySell)
}

// SetImbalanceStepFunction is a paid mutator transaction binding the contract method 0xbc9cbcc8.
//
// Solidity: function setImbalanceStepFunction(token address, xBuy int256[], yBuy int256[], xSell int256[], ySell int256[]) returns()
func (_Pricing *PricingTransactorSession) SetImbalanceStepFunction(token common.Address, xBuy []*big.Int, yBuy []*big.Int, xSell []*big.Int, ySell []*big.Int) (*types.Transaction, error) {
	return _Pricing.Contract.SetImbalanceStepFunction(&_Pricing.TransactOpts, token, xBuy, yBuy, xSell, ySell)
}

// SetQtyStepFunction is a paid mutator transaction binding the contract method 0x80d8b380.
//
// Solidity: function setQtyStepFunction(token address, xBuy int256[], yBuy int256[], xSell int256[], ySell int256[]) returns()
func (_Pricing *PricingTransactor) SetQtyStepFunction(opts *bind.TransactOpts, token common.Address, xBuy []*big.Int, yBuy []*big.Int, xSell []*big.Int, ySell []*big.Int) (*types.Transaction, error) {
	return _Pricing.contract.Transact(opts, "setQtyStepFunction", token, xBuy, yBuy, xSell, ySell)
}

// SetQtyStepFunction is a paid mutator transaction binding the contract method 0x80d8b380.
//
// Solidity: function setQtyStepFunction(token address, xBuy int256[], yBuy int256[], xSell int256[], ySell int256[]) returns()
func (_Pricing *PricingSession) SetQtyStepFunction(token common.Address, xBuy []*big.Int, yBuy []*big.Int, xSell []*big.Int, ySell []*big.Int) (*types.Transaction, error) {
	return _Pricing.Contract.SetQtyStepFunction(&_Pricing.TransactOpts, token, xBuy, yBuy, xSell, ySell)
}

// SetQtyStepFunction is a paid mutator transaction binding the contract method 0x80d8b380.
//
// Solidity: function setQtyStepFunction(token address, xBuy int256[], yBuy int256[], xSell int256[], ySell int256[]) returns()
func (_Pricing *PricingTransactorSession) SetQtyStepFunction(token common.Address, xBuy []*big.Int, yBuy []*big.Int, xSell []*big.Int, ySell []*big.Int) (*types.Transaction, error) {
	return _Pricing.Contract.SetQtyStepFunction(&_Pricing.TransactOpts, token, xBuy, yBuy, xSell, ySell)
}

// SetReserveAddress is a paid mutator transaction binding the contract method 0x14673d31.
//
// Solidity: function setReserveAddress(reserve address) returns()
func (_Pricing *PricingTransactor) SetReserveAddress(opts *bind.TransactOpts, reserve common.Address) (*types.Transaction, error) {
	return _Pricing.contract.Transact(opts, "setReserveAddress", reserve)
}

// SetReserveAddress is a paid mutator transaction binding the contract method 0x14673d31.
//
// Solidity: function setReserveAddress(reserve address) returns()
func (_Pricing *PricingSession) SetReserveAddress(reserve common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.SetReserveAddress(&_Pricing.TransactOpts, reserve)
}

// SetReserveAddress is a paid mutator transaction binding the contract method 0x14673d31.
//
// Solidity: function setReserveAddress(reserve address) returns()
func (_Pricing *PricingTransactorSession) SetReserveAddress(reserve common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.SetReserveAddress(&_Pricing.TransactOpts, reserve)
}

// SetTokenControlInfo is a paid mutator transaction binding the contract method 0xbfee3569.
//
// Solidity: function setTokenControlInfo(token address, minimalRecordResolution uint256, maxPerBlockImbalance uint256, maxTotalImbalance uint256) returns()
func (_Pricing *PricingTransactor) SetTokenControlInfo(opts *bind.TransactOpts, token common.Address, minimalRecordResolution *big.Int, maxPerBlockImbalance *big.Int, maxTotalImbalance *big.Int) (*types.Transaction, error) {
	return _Pricing.contract.Transact(opts, "setTokenControlInfo", token, minimalRecordResolution, maxPerBlockImbalance, maxTotalImbalance)
}

// SetTokenControlInfo is a paid mutator transaction binding the contract method 0xbfee3569.
//
// Solidity: function setTokenControlInfo(token address, minimalRecordResolution uint256, maxPerBlockImbalance uint256, maxTotalImbalance uint256) returns()
func (_Pricing *PricingSession) SetTokenControlInfo(token common.Address, minimalRecordResolution *big.Int, maxPerBlockImbalance *big.Int, maxTotalImbalance *big.Int) (*types.Transaction, error) {
	return _Pricing.Contract.SetTokenControlInfo(&_Pricing.TransactOpts, token, minimalRecordResolution, maxPerBlockImbalance, maxTotalImbalance)
}

// SetTokenControlInfo is a paid mutator transaction binding the contract method 0xbfee3569.
//
// Solidity: function setTokenControlInfo(token address, minimalRecordResolution uint256, maxPerBlockImbalance uint256, maxTotalImbalance uint256) returns()
func (_Pricing *PricingTransactorSession) SetTokenControlInfo(token common.Address, minimalRecordResolution *big.Int, maxPerBlockImbalance *big.Int, maxTotalImbalance *big.Int) (*types.Transaction, error) {
	return _Pricing.Contract.SetTokenControlInfo(&_Pricing.TransactOpts, token, minimalRecordResolution, maxPerBlockImbalance, maxTotalImbalance)
}

// SetValidRateDurationInBlocks is a paid mutator transaction binding the contract method 0x6c6295b8.
//
// Solidity: function setValidRateDurationInBlocks(duration uint256) returns()
func (_Pricing *PricingTransactor) SetValidRateDurationInBlocks(opts *bind.TransactOpts, duration *big.Int) (*types.Transaction, error) {
	return _Pricing.contract.Transact(opts, "setValidRateDurationInBlocks", duration)
}

// SetValidRateDurationInBlocks is a paid mutator transaction binding the contract method 0x6c6295b8.
//
// Solidity: function setValidRateDurationInBlocks(duration uint256) returns()
func (_Pricing *PricingSession) SetValidRateDurationInBlocks(duration *big.Int) (*types.Transaction, error) {
	return _Pricing.Contract.SetValidRateDurationInBlocks(&_Pricing.TransactOpts, duration)
}

// SetValidRateDurationInBlocks is a paid mutator transaction binding the contract method 0x6c6295b8.
//
// Solidity: function setValidRateDurationInBlocks(duration uint256) returns()
func (_Pricing *PricingTransactorSession) SetValidRateDurationInBlocks(duration *big.Int) (*types.Transaction, error) {
	return _Pricing.Contract.SetValidRateDurationInBlocks(&_Pricing.TransactOpts, duration)
}

// TransferAdmin is a paid mutator transaction binding the contract method 0x75829def.
//
// Solidity: function transferAdmin(newAdmin address) returns()
func (_Pricing *PricingTransactor) TransferAdmin(opts *bind.TransactOpts, newAdmin common.Address) (*types.Transaction, error) {
	return _Pricing.contract.Transact(opts, "transferAdmin", newAdmin)
}

// TransferAdmin is a paid mutator transaction binding the contract method 0x75829def.
//
// Solidity: function transferAdmin(newAdmin address) returns()
func (_Pricing *PricingSession) TransferAdmin(newAdmin common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.TransferAdmin(&_Pricing.TransactOpts, newAdmin)
}

// TransferAdmin is a paid mutator transaction binding the contract method 0x75829def.
//
// Solidity: function transferAdmin(newAdmin address) returns()
func (_Pricing *PricingTransactorSession) TransferAdmin(newAdmin common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.TransferAdmin(&_Pricing.TransactOpts, newAdmin)
}

// WithdrawEther is a paid mutator transaction binding the contract method 0xce56c454.
//
// Solidity: function withdrawEther(amount uint256, sendTo address) returns()
func (_Pricing *PricingTransactor) WithdrawEther(opts *bind.TransactOpts, amount *big.Int, sendTo common.Address) (*types.Transaction, error) {
	return _Pricing.contract.Transact(opts, "withdrawEther", amount, sendTo)
}

// WithdrawEther is a paid mutator transaction binding the contract method 0xce56c454.
//
// Solidity: function withdrawEther(amount uint256, sendTo address) returns()
func (_Pricing *PricingSession) WithdrawEther(amount *big.Int, sendTo common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.WithdrawEther(&_Pricing.TransactOpts, amount, sendTo)
}

// WithdrawEther is a paid mutator transaction binding the contract method 0xce56c454.
//
// Solidity: function withdrawEther(amount uint256, sendTo address) returns()
func (_Pricing *PricingTransactorSession) WithdrawEther(amount *big.Int, sendTo common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.WithdrawEther(&_Pricing.TransactOpts, amount, sendTo)
}

// WithdrawToken is a paid mutator transaction binding the contract method 0x3ccdbb28.
//
// Solidity: function withdrawToken(token address, amount uint256, sendTo address) returns()
func (_Pricing *PricingTransactor) WithdrawToken(opts *bind.TransactOpts, token common.Address, amount *big.Int, sendTo common.Address) (*types.Transaction, error) {
	return _Pricing.contract.Transact(opts, "withdrawToken", token, amount, sendTo)
}

// WithdrawToken is a paid mutator transaction binding the contract method 0x3ccdbb28.
//
// Solidity: function withdrawToken(token address, amount uint256, sendTo address) returns()
func (_Pricing *PricingSession) WithdrawToken(token common.Address, amount *big.Int, sendTo common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.WithdrawToken(&_Pricing.TransactOpts, token, amount, sendTo)
}

// WithdrawToken is a paid mutator transaction binding the contract method 0x3ccdbb28.
//
// Solidity: function withdrawToken(token address, amount uint256, sendTo address) returns()
func (_Pricing *PricingTransactorSession) WithdrawToken(token common.Address, amount *big.Int, sendTo common.Address) (*types.Transaction, error) {
	return _Pricing.Contract.WithdrawToken(&_Pricing.TransactOpts, token, amount, sendTo)
}
