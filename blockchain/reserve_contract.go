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

// ReserveContractABI is the input ABI used to generate the binding from.
const ReserveContractABI = "[{\"constant\":false,\"inputs\":[],\"name\":\"enableTrade\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"alerter\",\"type\":\"address\"}],\"name\":\"removeAlerter\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"pendingAdmin\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOperators\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"sendTo\",\"type\":\"address\"}],\"name\":\"withdrawToken\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newAlerter\",\"type\":\"address\"}],\"name\":\"addAlerter\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"sanityRatesContract\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"addr\",\"type\":\"address\"},{\"name\":\"approve\",\"type\":\"bool\"}],\"name\":\"approveWithdrawAddress\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"destination\",\"type\":\"address\"}],\"name\":\"withdraw\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"disableTrade\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"srcToken\",\"type\":\"address\"},{\"name\":\"srcAmount\",\"type\":\"uint256\"},{\"name\":\"destToken\",\"type\":\"address\"},{\"name\":\"destAddress\",\"type\":\"address\"},{\"name\":\"conversionRate\",\"type\":\"uint256\"},{\"name\":\"validate\",\"type\":\"bool\"}],\"name\":\"trade\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"transferAdmin\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"claimAdmin\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getAlerters\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"src\",\"type\":\"address\"},{\"name\":\"dest\",\"type\":\"address\"},{\"name\":\"srcQty\",\"type\":\"uint256\"},{\"name\":\"blockNumber\",\"type\":\"uint256\"}],\"name\":\"getConversionRate\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOperator\",\"type\":\"address\"}],\"name\":\"addOperator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"src\",\"type\":\"address\"},{\"name\":\"dest\",\"type\":\"address\"},{\"name\":\"dstQty\",\"type\":\"uint256\"},{\"name\":\"rate\",\"type\":\"uint256\"}],\"name\":\"getSrcQty\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"removeOperator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_kyberNetwork\",\"type\":\"address\"},{\"name\":\"_conversionRates\",\"type\":\"address\"},{\"name\":\"_sanityRates\",\"type\":\"address\"}],\"name\":\"setContracts\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"kyberNetwork\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"sendTo\",\"type\":\"address\"}],\"name\":\"withdrawEther\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"}],\"name\":\"getDecimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"conversionRatesContract\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"tradeEnabled\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"approvedWithdrawAddresses\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"admin\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"token\",\"type\":\"address\"}],\"name\":\"getBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"src\",\"type\":\"address\"},{\"name\":\"dest\",\"type\":\"address\"},{\"name\":\"srcQty\",\"type\":\"uint256\"},{\"name\":\"rate\",\"type\":\"uint256\"}],\"name\":\"getDestQty\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_kyberNetwork\",\"type\":\"address\"},{\"name\":\"_ratesContract\",\"type\":\"address\"},{\"name\":\"_admin\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DepositToken\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"origin\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"src\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"srcAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"destToken\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"destAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"destAddress\",\"type\":\"address\"}],\"name\":\"TradeExecute\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"enable\",\"type\":\"bool\"}],\"name\":\"TradeEnabled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"approve\",\"type\":\"bool\"}],\"name\":\"WithdrawAddressApproved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"destination\",\"type\":\"address\"}],\"name\":\"WithdrawFunds\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"network\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"rate\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"sanity\",\"type\":\"address\"}],\"name\":\"SetContractAddresses\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"sendTo\",\"type\":\"address\"}],\"name\":\"TokenWithdraw\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"sendTo\",\"type\":\"address\"}],\"name\":\"EtherWithdraw\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"pendingAdmin\",\"type\":\"address\"}],\"name\":\"TransferAdminPending\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"newAdmin\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"previousAdmin\",\"type\":\"address\"}],\"name\":\"AdminClaimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"newAlerter\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"isAdd\",\"type\":\"bool\"}],\"name\":\"AlerterAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"newOperator\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"isAdd\",\"type\":\"bool\"}],\"name\":\"OperatorAdded\",\"type\":\"event\"}]"

// ReserveContract is an auto generated Go binding around an Ethereum contract.
type ReserveContract struct {
	ReserveContractCaller     // Read-only binding to the contract
	ReserveContractTransactor // Write-only binding to the contract
}

// ReserveContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type ReserveContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReserveContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ReserveContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReserveContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ReserveContractSession struct {
	Contract     *ReserveContract  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ReserveContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ReserveContractCallerSession struct {
	Contract *ReserveContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// ReserveContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ReserveContractTransactorSession struct {
	Contract     *ReserveContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// ReserveContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type ReserveContractRaw struct {
	Contract *ReserveContract // Generic contract binding to access the raw methods on
}

// ReserveContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ReserveContractCallerRaw struct {
	Contract *ReserveContractCaller // Generic read-only contract binding to access the raw methods on
}

// ReserveContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ReserveContractTransactorRaw struct {
	Contract *ReserveContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewReserveContract creates a new instance of ReserveContract, bound to a specific deployed contract.
func NewReserveContract(address common.Address, backend bind.ContractBackend) (*ReserveContract, error) {
	contract, err := bindReserveContract(address, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ReserveContract{ReserveContractCaller: ReserveContractCaller{contract: contract}, ReserveContractTransactor: ReserveContractTransactor{contract: contract}}, nil
}

// NewReserveContractCaller creates a new read-only instance of ReserveContract, bound to a specific deployed contract.
func NewReserveContractCaller(address common.Address, caller bind.ContractCaller) (*ReserveContractCaller, error) {
	contract, err := bindReserveContract(address, caller, nil)
	if err != nil {
		return nil, err
	}
	return &ReserveContractCaller{contract: contract}, nil
}

// NewReserveContractTransactor creates a new write-only instance of ReserveContract, bound to a specific deployed contract.
func NewReserveContractTransactor(address common.Address, transactor bind.ContractTransactor) (*ReserveContractTransactor, error) {
	contract, err := bindReserveContract(address, nil, transactor)
	if err != nil {
		return nil, err
	}
	return &ReserveContractTransactor{contract: contract}, nil
}

// bindReserveContract binds a generic wrapper to an already deployed contract.
func bindReserveContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ReserveContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, nil), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ReserveContract *ReserveContractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ReserveContract.Contract.ReserveContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ReserveContract *ReserveContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ReserveContract.Contract.ReserveContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ReserveContract *ReserveContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ReserveContract.Contract.ReserveContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ReserveContract *ReserveContractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ReserveContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ReserveContract *ReserveContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ReserveContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ReserveContract *ReserveContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ReserveContract.Contract.contract.Transact(opts, method, params...)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() constant returns(address)
func (_ReserveContract *ReserveContractCaller) Admin(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _ReserveContract.contract.Call(opts, out, "admin")
	return *ret0, err
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() constant returns(address)
func (_ReserveContract *ReserveContractSession) Admin() (common.Address, error) {
	return _ReserveContract.Contract.Admin(&_ReserveContract.CallOpts)
}

// Admin is a free data retrieval call binding the contract method 0xf851a440.
//
// Solidity: function admin() constant returns(address)
func (_ReserveContract *ReserveContractCallerSession) Admin() (common.Address, error) {
	return _ReserveContract.Contract.Admin(&_ReserveContract.CallOpts)
}

// ApprovedWithdrawAddresses is a free data retrieval call binding the contract method 0xd7b7024d.
//
// Solidity: function approvedWithdrawAddresses( bytes32) constant returns(bool)
func (_ReserveContract *ReserveContractCaller) ApprovedWithdrawAddresses(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _ReserveContract.contract.Call(opts, out, "approvedWithdrawAddresses", arg0)
	return *ret0, err
}

// ApprovedWithdrawAddresses is a free data retrieval call binding the contract method 0xd7b7024d.
//
// Solidity: function approvedWithdrawAddresses( bytes32) constant returns(bool)
func (_ReserveContract *ReserveContractSession) ApprovedWithdrawAddresses(arg0 [32]byte) (bool, error) {
	return _ReserveContract.Contract.ApprovedWithdrawAddresses(&_ReserveContract.CallOpts, arg0)
}

// ApprovedWithdrawAddresses is a free data retrieval call binding the contract method 0xd7b7024d.
//
// Solidity: function approvedWithdrawAddresses( bytes32) constant returns(bool)
func (_ReserveContract *ReserveContractCallerSession) ApprovedWithdrawAddresses(arg0 [32]byte) (bool, error) {
	return _ReserveContract.Contract.ApprovedWithdrawAddresses(&_ReserveContract.CallOpts, arg0)
}

// ConversionRatesContract is a free data retrieval call binding the contract method 0xd5847d33.
//
// Solidity: function conversionRatesContract() constant returns(address)
func (_ReserveContract *ReserveContractCaller) ConversionRatesContract(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _ReserveContract.contract.Call(opts, out, "conversionRatesContract")
	return *ret0, err
}

// ConversionRatesContract is a free data retrieval call binding the contract method 0xd5847d33.
//
// Solidity: function conversionRatesContract() constant returns(address)
func (_ReserveContract *ReserveContractSession) ConversionRatesContract() (common.Address, error) {
	return _ReserveContract.Contract.ConversionRatesContract(&_ReserveContract.CallOpts)
}

// ConversionRatesContract is a free data retrieval call binding the contract method 0xd5847d33.
//
// Solidity: function conversionRatesContract() constant returns(address)
func (_ReserveContract *ReserveContractCallerSession) ConversionRatesContract() (common.Address, error) {
	return _ReserveContract.Contract.ConversionRatesContract(&_ReserveContract.CallOpts)
}

// GetAlerters is a free data retrieval call binding the contract method 0x7c423f54.
//
// Solidity: function getAlerters() constant returns(address[])
func (_ReserveContract *ReserveContractCaller) GetAlerters(opts *bind.CallOpts) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _ReserveContract.contract.Call(opts, out, "getAlerters")
	return *ret0, err
}

// GetAlerters is a free data retrieval call binding the contract method 0x7c423f54.
//
// Solidity: function getAlerters() constant returns(address[])
func (_ReserveContract *ReserveContractSession) GetAlerters() ([]common.Address, error) {
	return _ReserveContract.Contract.GetAlerters(&_ReserveContract.CallOpts)
}

// GetAlerters is a free data retrieval call binding the contract method 0x7c423f54.
//
// Solidity: function getAlerters() constant returns(address[])
func (_ReserveContract *ReserveContractCallerSession) GetAlerters() ([]common.Address, error) {
	return _ReserveContract.Contract.GetAlerters(&_ReserveContract.CallOpts)
}

// GetBalance is a free data retrieval call binding the contract method 0xf8b2cb4f.
//
// Solidity: function getBalance(token address) constant returns(uint256)
func (_ReserveContract *ReserveContractCaller) GetBalance(opts *bind.CallOpts, token common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ReserveContract.contract.Call(opts, out, "getBalance", token)
	return *ret0, err
}

// GetBalance is a free data retrieval call binding the contract method 0xf8b2cb4f.
//
// Solidity: function getBalance(token address) constant returns(uint256)
func (_ReserveContract *ReserveContractSession) GetBalance(token common.Address) (*big.Int, error) {
	return _ReserveContract.Contract.GetBalance(&_ReserveContract.CallOpts, token)
}

// GetBalance is a free data retrieval call binding the contract method 0xf8b2cb4f.
//
// Solidity: function getBalance(token address) constant returns(uint256)
func (_ReserveContract *ReserveContractCallerSession) GetBalance(token common.Address) (*big.Int, error) {
	return _ReserveContract.Contract.GetBalance(&_ReserveContract.CallOpts, token)
}

// GetConversionRate is a free data retrieval call binding the contract method 0x7cd44272.
//
// Solidity: function getConversionRate(src address, dest address, srcQty uint256, blockNumber uint256) constant returns(uint256)
func (_ReserveContract *ReserveContractCaller) GetConversionRate(opts *bind.CallOpts, src common.Address, dest common.Address, srcQty *big.Int, blockNumber *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ReserveContract.contract.Call(opts, out, "getConversionRate", src, dest, srcQty, blockNumber)
	return *ret0, err
}

// GetConversionRate is a free data retrieval call binding the contract method 0x7cd44272.
//
// Solidity: function getConversionRate(src address, dest address, srcQty uint256, blockNumber uint256) constant returns(uint256)
func (_ReserveContract *ReserveContractSession) GetConversionRate(src common.Address, dest common.Address, srcQty *big.Int, blockNumber *big.Int) (*big.Int, error) {
	return _ReserveContract.Contract.GetConversionRate(&_ReserveContract.CallOpts, src, dest, srcQty, blockNumber)
}

// GetConversionRate is a free data retrieval call binding the contract method 0x7cd44272.
//
// Solidity: function getConversionRate(src address, dest address, srcQty uint256, blockNumber uint256) constant returns(uint256)
func (_ReserveContract *ReserveContractCallerSession) GetConversionRate(src common.Address, dest common.Address, srcQty *big.Int, blockNumber *big.Int) (*big.Int, error) {
	return _ReserveContract.Contract.GetConversionRate(&_ReserveContract.CallOpts, src, dest, srcQty, blockNumber)
}

// GetDecimals is a free data retrieval call binding the contract method 0xcf54aaa0.
//
// Solidity: function getDecimals(token address) constant returns(uint256)
func (_ReserveContract *ReserveContractCaller) GetDecimals(opts *bind.CallOpts, token common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ReserveContract.contract.Call(opts, out, "getDecimals", token)
	return *ret0, err
}

// GetDecimals is a free data retrieval call binding the contract method 0xcf54aaa0.
//
// Solidity: function getDecimals(token address) constant returns(uint256)
func (_ReserveContract *ReserveContractSession) GetDecimals(token common.Address) (*big.Int, error) {
	return _ReserveContract.Contract.GetDecimals(&_ReserveContract.CallOpts, token)
}

// GetDecimals is a free data retrieval call binding the contract method 0xcf54aaa0.
//
// Solidity: function getDecimals(token address) constant returns(uint256)
func (_ReserveContract *ReserveContractCallerSession) GetDecimals(token common.Address) (*big.Int, error) {
	return _ReserveContract.Contract.GetDecimals(&_ReserveContract.CallOpts, token)
}

// GetDestQty is a free data retrieval call binding the contract method 0xfa64dffa.
//
// Solidity: function getDestQty(src address, dest address, srcQty uint256, rate uint256) constant returns(uint256)
func (_ReserveContract *ReserveContractCaller) GetDestQty(opts *bind.CallOpts, src common.Address, dest common.Address, srcQty *big.Int, rate *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ReserveContract.contract.Call(opts, out, "getDestQty", src, dest, srcQty, rate)
	return *ret0, err
}

// GetDestQty is a free data retrieval call binding the contract method 0xfa64dffa.
//
// Solidity: function getDestQty(src address, dest address, srcQty uint256, rate uint256) constant returns(uint256)
func (_ReserveContract *ReserveContractSession) GetDestQty(src common.Address, dest common.Address, srcQty *big.Int, rate *big.Int) (*big.Int, error) {
	return _ReserveContract.Contract.GetDestQty(&_ReserveContract.CallOpts, src, dest, srcQty, rate)
}

// GetDestQty is a free data retrieval call binding the contract method 0xfa64dffa.
//
// Solidity: function getDestQty(src address, dest address, srcQty uint256, rate uint256) constant returns(uint256)
func (_ReserveContract *ReserveContractCallerSession) GetDestQty(src common.Address, dest common.Address, srcQty *big.Int, rate *big.Int) (*big.Int, error) {
	return _ReserveContract.Contract.GetDestQty(&_ReserveContract.CallOpts, src, dest, srcQty, rate)
}

// GetOperators is a free data retrieval call binding the contract method 0x27a099d8.
//
// Solidity: function getOperators() constant returns(address[])
func (_ReserveContract *ReserveContractCaller) GetOperators(opts *bind.CallOpts) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _ReserveContract.contract.Call(opts, out, "getOperators")
	return *ret0, err
}

// GetOperators is a free data retrieval call binding the contract method 0x27a099d8.
//
// Solidity: function getOperators() constant returns(address[])
func (_ReserveContract *ReserveContractSession) GetOperators() ([]common.Address, error) {
	return _ReserveContract.Contract.GetOperators(&_ReserveContract.CallOpts)
}

// GetOperators is a free data retrieval call binding the contract method 0x27a099d8.
//
// Solidity: function getOperators() constant returns(address[])
func (_ReserveContract *ReserveContractCallerSession) GetOperators() ([]common.Address, error) {
	return _ReserveContract.Contract.GetOperators(&_ReserveContract.CallOpts)
}

// GetSrcQty is a free data retrieval call binding the contract method 0xa7fca953.
//
// Solidity: function getSrcQty(src address, dest address, dstQty uint256, rate uint256) constant returns(uint256)
func (_ReserveContract *ReserveContractCaller) GetSrcQty(opts *bind.CallOpts, src common.Address, dest common.Address, dstQty *big.Int, rate *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ReserveContract.contract.Call(opts, out, "getSrcQty", src, dest, dstQty, rate)
	return *ret0, err
}

// GetSrcQty is a free data retrieval call binding the contract method 0xa7fca953.
//
// Solidity: function getSrcQty(src address, dest address, dstQty uint256, rate uint256) constant returns(uint256)
func (_ReserveContract *ReserveContractSession) GetSrcQty(src common.Address, dest common.Address, dstQty *big.Int, rate *big.Int) (*big.Int, error) {
	return _ReserveContract.Contract.GetSrcQty(&_ReserveContract.CallOpts, src, dest, dstQty, rate)
}

// GetSrcQty is a free data retrieval call binding the contract method 0xa7fca953.
//
// Solidity: function getSrcQty(src address, dest address, dstQty uint256, rate uint256) constant returns(uint256)
func (_ReserveContract *ReserveContractCallerSession) GetSrcQty(src common.Address, dest common.Address, dstQty *big.Int, rate *big.Int) (*big.Int, error) {
	return _ReserveContract.Contract.GetSrcQty(&_ReserveContract.CallOpts, src, dest, dstQty, rate)
}

// KyberNetwork is a free data retrieval call binding the contract method 0xb78b842d.
//
// Solidity: function kyberNetwork() constant returns(address)
func (_ReserveContract *ReserveContractCaller) KyberNetwork(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _ReserveContract.contract.Call(opts, out, "kyberNetwork")
	return *ret0, err
}

// KyberNetwork is a free data retrieval call binding the contract method 0xb78b842d.
//
// Solidity: function kyberNetwork() constant returns(address)
func (_ReserveContract *ReserveContractSession) KyberNetwork() (common.Address, error) {
	return _ReserveContract.Contract.KyberNetwork(&_ReserveContract.CallOpts)
}

// KyberNetwork is a free data retrieval call binding the contract method 0xb78b842d.
//
// Solidity: function kyberNetwork() constant returns(address)
func (_ReserveContract *ReserveContractCallerSession) KyberNetwork() (common.Address, error) {
	return _ReserveContract.Contract.KyberNetwork(&_ReserveContract.CallOpts)
}

// PendingAdmin is a free data retrieval call binding the contract method 0x26782247.
//
// Solidity: function pendingAdmin() constant returns(address)
func (_ReserveContract *ReserveContractCaller) PendingAdmin(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _ReserveContract.contract.Call(opts, out, "pendingAdmin")
	return *ret0, err
}

// PendingAdmin is a free data retrieval call binding the contract method 0x26782247.
//
// Solidity: function pendingAdmin() constant returns(address)
func (_ReserveContract *ReserveContractSession) PendingAdmin() (common.Address, error) {
	return _ReserveContract.Contract.PendingAdmin(&_ReserveContract.CallOpts)
}

// PendingAdmin is a free data retrieval call binding the contract method 0x26782247.
//
// Solidity: function pendingAdmin() constant returns(address)
func (_ReserveContract *ReserveContractCallerSession) PendingAdmin() (common.Address, error) {
	return _ReserveContract.Contract.PendingAdmin(&_ReserveContract.CallOpts)
}

// SanityRatesContract is a free data retrieval call binding the contract method 0x47e6924f.
//
// Solidity: function sanityRatesContract() constant returns(address)
func (_ReserveContract *ReserveContractCaller) SanityRatesContract(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _ReserveContract.contract.Call(opts, out, "sanityRatesContract")
	return *ret0, err
}

// SanityRatesContract is a free data retrieval call binding the contract method 0x47e6924f.
//
// Solidity: function sanityRatesContract() constant returns(address)
func (_ReserveContract *ReserveContractSession) SanityRatesContract() (common.Address, error) {
	return _ReserveContract.Contract.SanityRatesContract(&_ReserveContract.CallOpts)
}

// SanityRatesContract is a free data retrieval call binding the contract method 0x47e6924f.
//
// Solidity: function sanityRatesContract() constant returns(address)
func (_ReserveContract *ReserveContractCallerSession) SanityRatesContract() (common.Address, error) {
	return _ReserveContract.Contract.SanityRatesContract(&_ReserveContract.CallOpts)
}

// TradeEnabled is a free data retrieval call binding the contract method 0xd621e813.
//
// Solidity: function tradeEnabled() constant returns(bool)
func (_ReserveContract *ReserveContractCaller) TradeEnabled(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _ReserveContract.contract.Call(opts, out, "tradeEnabled")
	return *ret0, err
}

// TradeEnabled is a free data retrieval call binding the contract method 0xd621e813.
//
// Solidity: function tradeEnabled() constant returns(bool)
func (_ReserveContract *ReserveContractSession) TradeEnabled() (bool, error) {
	return _ReserveContract.Contract.TradeEnabled(&_ReserveContract.CallOpts)
}

// TradeEnabled is a free data retrieval call binding the contract method 0xd621e813.
//
// Solidity: function tradeEnabled() constant returns(bool)
func (_ReserveContract *ReserveContractCallerSession) TradeEnabled() (bool, error) {
	return _ReserveContract.Contract.TradeEnabled(&_ReserveContract.CallOpts)
}

// AddAlerter is a paid mutator transaction binding the contract method 0x408ee7fe.
//
// Solidity: function addAlerter(newAlerter address) returns()
func (_ReserveContract *ReserveContractTransactor) AddAlerter(opts *bind.TransactOpts, newAlerter common.Address) (*types.Transaction, error) {
	return _ReserveContract.contract.Transact(opts, "addAlerter", newAlerter)
}

// AddAlerter is a paid mutator transaction binding the contract method 0x408ee7fe.
//
// Solidity: function addAlerter(newAlerter address) returns()
func (_ReserveContract *ReserveContractSession) AddAlerter(newAlerter common.Address) (*types.Transaction, error) {
	return _ReserveContract.Contract.AddAlerter(&_ReserveContract.TransactOpts, newAlerter)
}

// AddAlerter is a paid mutator transaction binding the contract method 0x408ee7fe.
//
// Solidity: function addAlerter(newAlerter address) returns()
func (_ReserveContract *ReserveContractTransactorSession) AddAlerter(newAlerter common.Address) (*types.Transaction, error) {
	return _ReserveContract.Contract.AddAlerter(&_ReserveContract.TransactOpts, newAlerter)
}

// AddOperator is a paid mutator transaction binding the contract method 0x9870d7fe.
//
// Solidity: function addOperator(newOperator address) returns()
func (_ReserveContract *ReserveContractTransactor) AddOperator(opts *bind.TransactOpts, newOperator common.Address) (*types.Transaction, error) {
	return _ReserveContract.contract.Transact(opts, "addOperator", newOperator)
}

// AddOperator is a paid mutator transaction binding the contract method 0x9870d7fe.
//
// Solidity: function addOperator(newOperator address) returns()
func (_ReserveContract *ReserveContractSession) AddOperator(newOperator common.Address) (*types.Transaction, error) {
	return _ReserveContract.Contract.AddOperator(&_ReserveContract.TransactOpts, newOperator)
}

// AddOperator is a paid mutator transaction binding the contract method 0x9870d7fe.
//
// Solidity: function addOperator(newOperator address) returns()
func (_ReserveContract *ReserveContractTransactorSession) AddOperator(newOperator common.Address) (*types.Transaction, error) {
	return _ReserveContract.Contract.AddOperator(&_ReserveContract.TransactOpts, newOperator)
}

// ApproveWithdrawAddress is a paid mutator transaction binding the contract method 0x546dc71c.
//
// Solidity: function approveWithdrawAddress(token address, addr address, approve bool) returns()
func (_ReserveContract *ReserveContractTransactor) ApproveWithdrawAddress(opts *bind.TransactOpts, token common.Address, addr common.Address, approve bool) (*types.Transaction, error) {
	return _ReserveContract.contract.Transact(opts, "approveWithdrawAddress", token, addr, approve)
}

// ApproveWithdrawAddress is a paid mutator transaction binding the contract method 0x546dc71c.
//
// Solidity: function approveWithdrawAddress(token address, addr address, approve bool) returns()
func (_ReserveContract *ReserveContractSession) ApproveWithdrawAddress(token common.Address, addr common.Address, approve bool) (*types.Transaction, error) {
	return _ReserveContract.Contract.ApproveWithdrawAddress(&_ReserveContract.TransactOpts, token, addr, approve)
}

// ApproveWithdrawAddress is a paid mutator transaction binding the contract method 0x546dc71c.
//
// Solidity: function approveWithdrawAddress(token address, addr address, approve bool) returns()
func (_ReserveContract *ReserveContractTransactorSession) ApproveWithdrawAddress(token common.Address, addr common.Address, approve bool) (*types.Transaction, error) {
	return _ReserveContract.Contract.ApproveWithdrawAddress(&_ReserveContract.TransactOpts, token, addr, approve)
}

// ClaimAdmin is a paid mutator transaction binding the contract method 0x77f50f97.
//
// Solidity: function claimAdmin() returns()
func (_ReserveContract *ReserveContractTransactor) ClaimAdmin(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ReserveContract.contract.Transact(opts, "claimAdmin")
}

// ClaimAdmin is a paid mutator transaction binding the contract method 0x77f50f97.
//
// Solidity: function claimAdmin() returns()
func (_ReserveContract *ReserveContractSession) ClaimAdmin() (*types.Transaction, error) {
	return _ReserveContract.Contract.ClaimAdmin(&_ReserveContract.TransactOpts)
}

// ClaimAdmin is a paid mutator transaction binding the contract method 0x77f50f97.
//
// Solidity: function claimAdmin() returns()
func (_ReserveContract *ReserveContractTransactorSession) ClaimAdmin() (*types.Transaction, error) {
	return _ReserveContract.Contract.ClaimAdmin(&_ReserveContract.TransactOpts)
}

// DisableTrade is a paid mutator transaction binding the contract method 0x6940030f.
//
// Solidity: function disableTrade() returns(bool)
func (_ReserveContract *ReserveContractTransactor) DisableTrade(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ReserveContract.contract.Transact(opts, "disableTrade")
}

// DisableTrade is a paid mutator transaction binding the contract method 0x6940030f.
//
// Solidity: function disableTrade() returns(bool)
func (_ReserveContract *ReserveContractSession) DisableTrade() (*types.Transaction, error) {
	return _ReserveContract.Contract.DisableTrade(&_ReserveContract.TransactOpts)
}

// DisableTrade is a paid mutator transaction binding the contract method 0x6940030f.
//
// Solidity: function disableTrade() returns(bool)
func (_ReserveContract *ReserveContractTransactorSession) DisableTrade() (*types.Transaction, error) {
	return _ReserveContract.Contract.DisableTrade(&_ReserveContract.TransactOpts)
}

// EnableTrade is a paid mutator transaction binding the contract method 0x0099d386.
//
// Solidity: function enableTrade() returns(bool)
func (_ReserveContract *ReserveContractTransactor) EnableTrade(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ReserveContract.contract.Transact(opts, "enableTrade")
}

// EnableTrade is a paid mutator transaction binding the contract method 0x0099d386.
//
// Solidity: function enableTrade() returns(bool)
func (_ReserveContract *ReserveContractSession) EnableTrade() (*types.Transaction, error) {
	return _ReserveContract.Contract.EnableTrade(&_ReserveContract.TransactOpts)
}

// EnableTrade is a paid mutator transaction binding the contract method 0x0099d386.
//
// Solidity: function enableTrade() returns(bool)
func (_ReserveContract *ReserveContractTransactorSession) EnableTrade() (*types.Transaction, error) {
	return _ReserveContract.Contract.EnableTrade(&_ReserveContract.TransactOpts)
}

// RemoveAlerter is a paid mutator transaction binding the contract method 0x01a12fd3.
//
// Solidity: function removeAlerter(alerter address) returns()
func (_ReserveContract *ReserveContractTransactor) RemoveAlerter(opts *bind.TransactOpts, alerter common.Address) (*types.Transaction, error) {
	return _ReserveContract.contract.Transact(opts, "removeAlerter", alerter)
}

// RemoveAlerter is a paid mutator transaction binding the contract method 0x01a12fd3.
//
// Solidity: function removeAlerter(alerter address) returns()
func (_ReserveContract *ReserveContractSession) RemoveAlerter(alerter common.Address) (*types.Transaction, error) {
	return _ReserveContract.Contract.RemoveAlerter(&_ReserveContract.TransactOpts, alerter)
}

// RemoveAlerter is a paid mutator transaction binding the contract method 0x01a12fd3.
//
// Solidity: function removeAlerter(alerter address) returns()
func (_ReserveContract *ReserveContractTransactorSession) RemoveAlerter(alerter common.Address) (*types.Transaction, error) {
	return _ReserveContract.Contract.RemoveAlerter(&_ReserveContract.TransactOpts, alerter)
}

// RemoveOperator is a paid mutator transaction binding the contract method 0xac8a584a.
//
// Solidity: function removeOperator(operator address) returns()
func (_ReserveContract *ReserveContractTransactor) RemoveOperator(opts *bind.TransactOpts, operator common.Address) (*types.Transaction, error) {
	return _ReserveContract.contract.Transact(opts, "removeOperator", operator)
}

// RemoveOperator is a paid mutator transaction binding the contract method 0xac8a584a.
//
// Solidity: function removeOperator(operator address) returns()
func (_ReserveContract *ReserveContractSession) RemoveOperator(operator common.Address) (*types.Transaction, error) {
	return _ReserveContract.Contract.RemoveOperator(&_ReserveContract.TransactOpts, operator)
}

// RemoveOperator is a paid mutator transaction binding the contract method 0xac8a584a.
//
// Solidity: function removeOperator(operator address) returns()
func (_ReserveContract *ReserveContractTransactorSession) RemoveOperator(operator common.Address) (*types.Transaction, error) {
	return _ReserveContract.Contract.RemoveOperator(&_ReserveContract.TransactOpts, operator)
}

// SetContracts is a paid mutator transaction binding the contract method 0xb3066d49.
//
// Solidity: function setContracts(_kyberNetwork address, _conversionRates address, _sanityRates address) returns()
func (_ReserveContract *ReserveContractTransactor) SetContracts(opts *bind.TransactOpts, _kyberNetwork common.Address, _conversionRates common.Address, _sanityRates common.Address) (*types.Transaction, error) {
	return _ReserveContract.contract.Transact(opts, "setContracts", _kyberNetwork, _conversionRates, _sanityRates)
}

// SetContracts is a paid mutator transaction binding the contract method 0xb3066d49.
//
// Solidity: function setContracts(_kyberNetwork address, _conversionRates address, _sanityRates address) returns()
func (_ReserveContract *ReserveContractSession) SetContracts(_kyberNetwork common.Address, _conversionRates common.Address, _sanityRates common.Address) (*types.Transaction, error) {
	return _ReserveContract.Contract.SetContracts(&_ReserveContract.TransactOpts, _kyberNetwork, _conversionRates, _sanityRates)
}

// SetContracts is a paid mutator transaction binding the contract method 0xb3066d49.
//
// Solidity: function setContracts(_kyberNetwork address, _conversionRates address, _sanityRates address) returns()
func (_ReserveContract *ReserveContractTransactorSession) SetContracts(_kyberNetwork common.Address, _conversionRates common.Address, _sanityRates common.Address) (*types.Transaction, error) {
	return _ReserveContract.Contract.SetContracts(&_ReserveContract.TransactOpts, _kyberNetwork, _conversionRates, _sanityRates)
}

// Trade is a paid mutator transaction binding the contract method 0x6cf69811.
//
// Solidity: function trade(srcToken address, srcAmount uint256, destToken address, destAddress address, conversionRate uint256, validate bool) returns(bool)
func (_ReserveContract *ReserveContractTransactor) Trade(opts *bind.TransactOpts, srcToken common.Address, srcAmount *big.Int, destToken common.Address, destAddress common.Address, conversionRate *big.Int, validate bool) (*types.Transaction, error) {
	return _ReserveContract.contract.Transact(opts, "trade", srcToken, srcAmount, destToken, destAddress, conversionRate, validate)
}

// Trade is a paid mutator transaction binding the contract method 0x6cf69811.
//
// Solidity: function trade(srcToken address, srcAmount uint256, destToken address, destAddress address, conversionRate uint256, validate bool) returns(bool)
func (_ReserveContract *ReserveContractSession) Trade(srcToken common.Address, srcAmount *big.Int, destToken common.Address, destAddress common.Address, conversionRate *big.Int, validate bool) (*types.Transaction, error) {
	return _ReserveContract.Contract.Trade(&_ReserveContract.TransactOpts, srcToken, srcAmount, destToken, destAddress, conversionRate, validate)
}

// Trade is a paid mutator transaction binding the contract method 0x6cf69811.
//
// Solidity: function trade(srcToken address, srcAmount uint256, destToken address, destAddress address, conversionRate uint256, validate bool) returns(bool)
func (_ReserveContract *ReserveContractTransactorSession) Trade(srcToken common.Address, srcAmount *big.Int, destToken common.Address, destAddress common.Address, conversionRate *big.Int, validate bool) (*types.Transaction, error) {
	return _ReserveContract.Contract.Trade(&_ReserveContract.TransactOpts, srcToken, srcAmount, destToken, destAddress, conversionRate, validate)
}

// TransferAdmin is a paid mutator transaction binding the contract method 0x75829def.
//
// Solidity: function transferAdmin(newAdmin address) returns()
func (_ReserveContract *ReserveContractTransactor) TransferAdmin(opts *bind.TransactOpts, newAdmin common.Address) (*types.Transaction, error) {
	return _ReserveContract.contract.Transact(opts, "transferAdmin", newAdmin)
}

// TransferAdmin is a paid mutator transaction binding the contract method 0x75829def.
//
// Solidity: function transferAdmin(newAdmin address) returns()
func (_ReserveContract *ReserveContractSession) TransferAdmin(newAdmin common.Address) (*types.Transaction, error) {
	return _ReserveContract.Contract.TransferAdmin(&_ReserveContract.TransactOpts, newAdmin)
}

// TransferAdmin is a paid mutator transaction binding the contract method 0x75829def.
//
// Solidity: function transferAdmin(newAdmin address) returns()
func (_ReserveContract *ReserveContractTransactorSession) TransferAdmin(newAdmin common.Address) (*types.Transaction, error) {
	return _ReserveContract.Contract.TransferAdmin(&_ReserveContract.TransactOpts, newAdmin)
}

// Withdraw is a paid mutator transaction binding the contract method 0x69328dec.
//
// Solidity: function withdraw(token address, amount uint256, destination address) returns(bool)
func (_ReserveContract *ReserveContractTransactor) Withdraw(opts *bind.TransactOpts, token common.Address, amount *big.Int, destination common.Address) (*types.Transaction, error) {
	return _ReserveContract.contract.Transact(opts, "withdraw", token, amount, destination)
}

// Withdraw is a paid mutator transaction binding the contract method 0x69328dec.
//
// Solidity: function withdraw(token address, amount uint256, destination address) returns(bool)
func (_ReserveContract *ReserveContractSession) Withdraw(token common.Address, amount *big.Int, destination common.Address) (*types.Transaction, error) {
	return _ReserveContract.Contract.Withdraw(&_ReserveContract.TransactOpts, token, amount, destination)
}

// Withdraw is a paid mutator transaction binding the contract method 0x69328dec.
//
// Solidity: function withdraw(token address, amount uint256, destination address) returns(bool)
func (_ReserveContract *ReserveContractTransactorSession) Withdraw(token common.Address, amount *big.Int, destination common.Address) (*types.Transaction, error) {
	return _ReserveContract.Contract.Withdraw(&_ReserveContract.TransactOpts, token, amount, destination)
}

// WithdrawEther is a paid mutator transaction binding the contract method 0xce56c454.
//
// Solidity: function withdrawEther(amount uint256, sendTo address) returns()
func (_ReserveContract *ReserveContractTransactor) WithdrawEther(opts *bind.TransactOpts, amount *big.Int, sendTo common.Address) (*types.Transaction, error) {
	return _ReserveContract.contract.Transact(opts, "withdrawEther", amount, sendTo)
}

// WithdrawEther is a paid mutator transaction binding the contract method 0xce56c454.
//
// Solidity: function withdrawEther(amount uint256, sendTo address) returns()
func (_ReserveContract *ReserveContractSession) WithdrawEther(amount *big.Int, sendTo common.Address) (*types.Transaction, error) {
	return _ReserveContract.Contract.WithdrawEther(&_ReserveContract.TransactOpts, amount, sendTo)
}

// WithdrawEther is a paid mutator transaction binding the contract method 0xce56c454.
//
// Solidity: function withdrawEther(amount uint256, sendTo address) returns()
func (_ReserveContract *ReserveContractTransactorSession) WithdrawEther(amount *big.Int, sendTo common.Address) (*types.Transaction, error) {
	return _ReserveContract.Contract.WithdrawEther(&_ReserveContract.TransactOpts, amount, sendTo)
}

// WithdrawToken is a paid mutator transaction binding the contract method 0x3ccdbb28.
//
// Solidity: function withdrawToken(token address, amount uint256, sendTo address) returns()
func (_ReserveContract *ReserveContractTransactor) WithdrawToken(opts *bind.TransactOpts, token common.Address, amount *big.Int, sendTo common.Address) (*types.Transaction, error) {
	return _ReserveContract.contract.Transact(opts, "withdrawToken", token, amount, sendTo)
}

// WithdrawToken is a paid mutator transaction binding the contract method 0x3ccdbb28.
//
// Solidity: function withdrawToken(token address, amount uint256, sendTo address) returns()
func (_ReserveContract *ReserveContractSession) WithdrawToken(token common.Address, amount *big.Int, sendTo common.Address) (*types.Transaction, error) {
	return _ReserveContract.Contract.WithdrawToken(&_ReserveContract.TransactOpts, token, amount, sendTo)
}

// WithdrawToken is a paid mutator transaction binding the contract method 0x3ccdbb28.
//
// Solidity: function withdrawToken(token address, amount uint256, sendTo address) returns()
func (_ReserveContract *ReserveContractTransactorSession) WithdrawToken(token common.Address, amount *big.Int, sendTo common.Address) (*types.Transaction, error) {
	return _ReserveContract.Contract.WithdrawToken(&_ReserveContract.TransactOpts, token, amount, sendTo)
}
