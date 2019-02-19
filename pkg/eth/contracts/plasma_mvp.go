// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// PlasmaABI is the input ABI used to generate the binding from.
const PlasmaABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"balances\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lastCommittedBlock\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"plasmaChain\",\"outputs\":[{\"name\":\"header\",\"type\":\"bytes32\"},{\"name\":\"numTxns\",\"type\":\"uint256\"},{\"name\":\"feeAmount\",\"type\":\"uint256\"},{\"name\":\"createdAt\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"operator\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"depositExitQueue\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"txExits\",\"outputs\":[{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"committedFee\",\"type\":\"uint256\"},{\"name\":\"createdAt\",\"type\":\"uint256\"},{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"state\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"txExitQueue\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"deposits\",\"outputs\":[{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"createdAt\",\"type\":\"uint256\"},{\"name\":\"ethBlockNum\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalWithdrawBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"depositExits\",\"outputs\":[{\"name\":\"amount\",\"type\":\"uint256\"},{\"name\":\"committedFee\",\"type\":\"uint256\"},{\"name\":\"createdAt\",\"type\":\"uint256\"},{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"state\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"minExitBond\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"depositNonce\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"oldOperator\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"newOperator\",\"type\":\"address\"}],\"name\":\"ChangedOperator\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"AddedToBalances\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"header\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"blockNumber\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"numTxns\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"feeAmount\",\"type\":\"uint256\"}],\"name\":\"BlockSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"depositor\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"depositNonce\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"ethBlockNum\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"position\",\"type\":\"uint256[3]\"},{\"indexed\":false,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"confirmSignatures\",\"type\":\"bytes\"},{\"indexed\":false,\"name\":\"committedFee\",\"type\":\"uint256\"}],\"name\":\"StartedTransactionExit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"nonce\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"committedFee\",\"type\":\"uint256\"}],\"name\":\"StartedDepositExit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"position\",\"type\":\"uint256[4]\"},{\"indexed\":false,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ChallengedExit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"position\",\"type\":\"uint256[4]\"},{\"indexed\":false,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"FinalizedExit\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOperator\",\"type\":\"address\"}],\"name\":\"changeOperator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"headers\",\"type\":\"bytes32[]\"},{\"name\":\"txnsPerBlock\",\"type\":\"uint256[]\"},{\"name\":\"feePerBlock\",\"type\":\"uint256[]\"},{\"name\":\"blockNum\",\"type\":\"uint256\"}],\"name\":\"submitBlock\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"deposit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"nonce\",\"type\":\"uint256\"},{\"name\":\"committedFee\",\"type\":\"uint256\"}],\"name\":\"startDepositExit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"txPos\",\"type\":\"uint256[3]\"},{\"name\":\"txBytes\",\"type\":\"bytes\"},{\"name\":\"proof\",\"type\":\"bytes\"},{\"name\":\"confirmSignatures\",\"type\":\"bytes\"},{\"name\":\"committedFee\",\"type\":\"uint256\"}],\"name\":\"startTransactionExit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"blockNumber\",\"type\":\"uint256\"},{\"name\":\"committedFee\",\"type\":\"uint256\"}],\"name\":\"startFeeExit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"exitingTxPos\",\"type\":\"uint256[4]\"},{\"name\":\"challengingTxPos\",\"type\":\"uint256[2]\"},{\"name\":\"txBytes\",\"type\":\"bytes\"},{\"name\":\"proof\",\"type\":\"bytes\"},{\"name\":\"confirmSignature\",\"type\":\"bytes\"}],\"name\":\"challengeExit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"finalizeDepositExits\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"finalizeTransactionExits\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"withdraw\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"plasmaChainBalance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_address\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// Plasma is an auto generated Go binding around an Ethereum contract.
type Plasma struct {
	PlasmaCaller     // Read-only binding to the contract
	PlasmaTransactor // Write-only binding to the contract
	PlasmaFilterer   // Log filterer for contract events
}

// PlasmaCaller is an auto generated read-only Go binding around an Ethereum contract.
type PlasmaCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PlasmaTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PlasmaTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PlasmaFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PlasmaFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PlasmaSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PlasmaSession struct {
	Contract     *Plasma           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PlasmaCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PlasmaCallerSession struct {
	Contract *PlasmaCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// PlasmaTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PlasmaTransactorSession struct {
	Contract     *PlasmaTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PlasmaRaw is an auto generated low-level Go binding around an Ethereum contract.
type PlasmaRaw struct {
	Contract *Plasma // Generic contract binding to access the raw methods on
}

// PlasmaCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PlasmaCallerRaw struct {
	Contract *PlasmaCaller // Generic read-only contract binding to access the raw methods on
}

// PlasmaTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PlasmaTransactorRaw struct {
	Contract *PlasmaTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPlasma creates a new instance of Plasma, bound to a specific deployed contract.
func NewPlasma(address common.Address, backend bind.ContractBackend) (*Plasma, error) {
	contract, err := bindPlasma(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Plasma{PlasmaCaller: PlasmaCaller{contract: contract}, PlasmaTransactor: PlasmaTransactor{contract: contract}, PlasmaFilterer: PlasmaFilterer{contract: contract}}, nil
}

// NewPlasmaCaller creates a new read-only instance of Plasma, bound to a specific deployed contract.
func NewPlasmaCaller(address common.Address, caller bind.ContractCaller) (*PlasmaCaller, error) {
	contract, err := bindPlasma(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PlasmaCaller{contract: contract}, nil
}

// NewPlasmaTransactor creates a new write-only instance of Plasma, bound to a specific deployed contract.
func NewPlasmaTransactor(address common.Address, transactor bind.ContractTransactor) (*PlasmaTransactor, error) {
	contract, err := bindPlasma(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PlasmaTransactor{contract: contract}, nil
}

// NewPlasmaFilterer creates a new log filterer instance of Plasma, bound to a specific deployed contract.
func NewPlasmaFilterer(address common.Address, filterer bind.ContractFilterer) (*PlasmaFilterer, error) {
	contract, err := bindPlasma(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PlasmaFilterer{contract: contract}, nil
}

// bindPlasma binds a generic wrapper to an already deployed contract.
func bindPlasma(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(PlasmaABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Plasma *PlasmaRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Plasma.Contract.PlasmaCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Plasma *PlasmaRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Plasma.Contract.PlasmaTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Plasma *PlasmaRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Plasma.Contract.PlasmaTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Plasma *PlasmaCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Plasma.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Plasma *PlasmaTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Plasma.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Plasma *PlasmaTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Plasma.Contract.contract.Transact(opts, method, params...)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_address address) constant returns(uint256)
func (_Plasma *PlasmaCaller) BalanceOf(opts *bind.CallOpts, _address common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Plasma.contract.Call(opts, out, "balanceOf", _address)
	return *ret0, err
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_address address) constant returns(uint256)
func (_Plasma *PlasmaSession) BalanceOf(_address common.Address) (*big.Int, error) {
	return _Plasma.Contract.BalanceOf(&_Plasma.CallOpts, _address)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(_address address) constant returns(uint256)
func (_Plasma *PlasmaCallerSession) BalanceOf(_address common.Address) (*big.Int, error) {
	return _Plasma.Contract.BalanceOf(&_Plasma.CallOpts, _address)
}

// Balances is a free data retrieval call binding the contract method 0x27e235e3.
//
// Solidity: function balances( address) constant returns(uint256)
func (_Plasma *PlasmaCaller) Balances(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Plasma.contract.Call(opts, out, "balances", arg0)
	return *ret0, err
}

// Balances is a free data retrieval call binding the contract method 0x27e235e3.
//
// Solidity: function balances( address) constant returns(uint256)
func (_Plasma *PlasmaSession) Balances(arg0 common.Address) (*big.Int, error) {
	return _Plasma.Contract.Balances(&_Plasma.CallOpts, arg0)
}

// Balances is a free data retrieval call binding the contract method 0x27e235e3.
//
// Solidity: function balances( address) constant returns(uint256)
func (_Plasma *PlasmaCallerSession) Balances(arg0 common.Address) (*big.Int, error) {
	return _Plasma.Contract.Balances(&_Plasma.CallOpts, arg0)
}

// DepositExitQueue is a free data retrieval call binding the contract method 0x5b3081d7.
//
// Solidity: function depositExitQueue( uint256) constant returns(uint256)
func (_Plasma *PlasmaCaller) DepositExitQueue(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Plasma.contract.Call(opts, out, "depositExitQueue", arg0)
	return *ret0, err
}

// DepositExitQueue is a free data retrieval call binding the contract method 0x5b3081d7.
//
// Solidity: function depositExitQueue( uint256) constant returns(uint256)
func (_Plasma *PlasmaSession) DepositExitQueue(arg0 *big.Int) (*big.Int, error) {
	return _Plasma.Contract.DepositExitQueue(&_Plasma.CallOpts, arg0)
}

// DepositExitQueue is a free data retrieval call binding the contract method 0x5b3081d7.
//
// Solidity: function depositExitQueue( uint256) constant returns(uint256)
func (_Plasma *PlasmaCallerSession) DepositExitQueue(arg0 *big.Int) (*big.Int, error) {
	return _Plasma.Contract.DepositExitQueue(&_Plasma.CallOpts, arg0)
}

// DepositExits is a free data retrieval call binding the contract method 0xce84f906.
//
// Solidity: function depositExits( uint256) constant returns(amount uint256, committedFee uint256, createdAt uint256, owner address, state uint8)
func (_Plasma *PlasmaCaller) DepositExits(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Amount       *big.Int
	CommittedFee *big.Int
	CreatedAt    *big.Int
	Owner        common.Address
	State        uint8
}, error) {
	ret := new(struct {
		Amount       *big.Int
		CommittedFee *big.Int
		CreatedAt    *big.Int
		Owner        common.Address
		State        uint8
	})
	out := ret
	err := _Plasma.contract.Call(opts, out, "depositExits", arg0)
	return *ret, err
}

// DepositExits is a free data retrieval call binding the contract method 0xce84f906.
//
// Solidity: function depositExits( uint256) constant returns(amount uint256, committedFee uint256, createdAt uint256, owner address, state uint8)
func (_Plasma *PlasmaSession) DepositExits(arg0 *big.Int) (struct {
	Amount       *big.Int
	CommittedFee *big.Int
	CreatedAt    *big.Int
	Owner        common.Address
	State        uint8
}, error) {
	return _Plasma.Contract.DepositExits(&_Plasma.CallOpts, arg0)
}

// DepositExits is a free data retrieval call binding the contract method 0xce84f906.
//
// Solidity: function depositExits( uint256) constant returns(amount uint256, committedFee uint256, createdAt uint256, owner address, state uint8)
func (_Plasma *PlasmaCallerSession) DepositExits(arg0 *big.Int) (struct {
	Amount       *big.Int
	CommittedFee *big.Int
	CreatedAt    *big.Int
	Owner        common.Address
	State        uint8
}, error) {
	return _Plasma.Contract.DepositExits(&_Plasma.CallOpts, arg0)
}

// DepositNonce is a free data retrieval call binding the contract method 0xde35f5cb.
//
// Solidity: function depositNonce() constant returns(uint256)
func (_Plasma *PlasmaCaller) DepositNonce(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Plasma.contract.Call(opts, out, "depositNonce")
	return *ret0, err
}

// DepositNonce is a free data retrieval call binding the contract method 0xde35f5cb.
//
// Solidity: function depositNonce() constant returns(uint256)
func (_Plasma *PlasmaSession) DepositNonce() (*big.Int, error) {
	return _Plasma.Contract.DepositNonce(&_Plasma.CallOpts)
}

// DepositNonce is a free data retrieval call binding the contract method 0xde35f5cb.
//
// Solidity: function depositNonce() constant returns(uint256)
func (_Plasma *PlasmaCallerSession) DepositNonce() (*big.Int, error) {
	return _Plasma.Contract.DepositNonce(&_Plasma.CallOpts)
}

// Deposits is a free data retrieval call binding the contract method 0xb02c43d0.
//
// Solidity: function deposits( uint256) constant returns(owner address, amount uint256, createdAt uint256, ethBlockNum uint256)
func (_Plasma *PlasmaCaller) Deposits(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Owner       common.Address
	Amount      *big.Int
	CreatedAt   *big.Int
	EthBlockNum *big.Int
}, error) {
	ret := new(struct {
		Owner       common.Address
		Amount      *big.Int
		CreatedAt   *big.Int
		EthBlockNum *big.Int
	})
	out := ret
	err := _Plasma.contract.Call(opts, out, "deposits", arg0)
	return *ret, err
}

// Deposits is a free data retrieval call binding the contract method 0xb02c43d0.
//
// Solidity: function deposits( uint256) constant returns(owner address, amount uint256, createdAt uint256, ethBlockNum uint256)
func (_Plasma *PlasmaSession) Deposits(arg0 *big.Int) (struct {
	Owner       common.Address
	Amount      *big.Int
	CreatedAt   *big.Int
	EthBlockNum *big.Int
}, error) {
	return _Plasma.Contract.Deposits(&_Plasma.CallOpts, arg0)
}

// Deposits is a free data retrieval call binding the contract method 0xb02c43d0.
//
// Solidity: function deposits( uint256) constant returns(owner address, amount uint256, createdAt uint256, ethBlockNum uint256)
func (_Plasma *PlasmaCallerSession) Deposits(arg0 *big.Int) (struct {
	Owner       common.Address
	Amount      *big.Int
	CreatedAt   *big.Int
	EthBlockNum *big.Int
}, error) {
	return _Plasma.Contract.Deposits(&_Plasma.CallOpts, arg0)
}

// LastCommittedBlock is a free data retrieval call binding the contract method 0x3acb097a.
//
// Solidity: function lastCommittedBlock() constant returns(uint256)
func (_Plasma *PlasmaCaller) LastCommittedBlock(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Plasma.contract.Call(opts, out, "lastCommittedBlock")
	return *ret0, err
}

// LastCommittedBlock is a free data retrieval call binding the contract method 0x3acb097a.
//
// Solidity: function lastCommittedBlock() constant returns(uint256)
func (_Plasma *PlasmaSession) LastCommittedBlock() (*big.Int, error) {
	return _Plasma.Contract.LastCommittedBlock(&_Plasma.CallOpts)
}

// LastCommittedBlock is a free data retrieval call binding the contract method 0x3acb097a.
//
// Solidity: function lastCommittedBlock() constant returns(uint256)
func (_Plasma *PlasmaCallerSession) LastCommittedBlock() (*big.Int, error) {
	return _Plasma.Contract.LastCommittedBlock(&_Plasma.CallOpts)
}

// MinExitBond is a free data retrieval call binding the contract method 0xd68545a3.
//
// Solidity: function minExitBond() constant returns(uint256)
func (_Plasma *PlasmaCaller) MinExitBond(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Plasma.contract.Call(opts, out, "minExitBond")
	return *ret0, err
}

// MinExitBond is a free data retrieval call binding the contract method 0xd68545a3.
//
// Solidity: function minExitBond() constant returns(uint256)
func (_Plasma *PlasmaSession) MinExitBond() (*big.Int, error) {
	return _Plasma.Contract.MinExitBond(&_Plasma.CallOpts)
}

// MinExitBond is a free data retrieval call binding the contract method 0xd68545a3.
//
// Solidity: function minExitBond() constant returns(uint256)
func (_Plasma *PlasmaCallerSession) MinExitBond() (*big.Int, error) {
	return _Plasma.Contract.MinExitBond(&_Plasma.CallOpts)
}

// Operator is a free data retrieval call binding the contract method 0x570ca735.
//
// Solidity: function operator() constant returns(address)
func (_Plasma *PlasmaCaller) Operator(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Plasma.contract.Call(opts, out, "operator")
	return *ret0, err
}

// Operator is a free data retrieval call binding the contract method 0x570ca735.
//
// Solidity: function operator() constant returns(address)
func (_Plasma *PlasmaSession) Operator() (common.Address, error) {
	return _Plasma.Contract.Operator(&_Plasma.CallOpts)
}

// Operator is a free data retrieval call binding the contract method 0x570ca735.
//
// Solidity: function operator() constant returns(address)
func (_Plasma *PlasmaCallerSession) Operator() (common.Address, error) {
	return _Plasma.Contract.Operator(&_Plasma.CallOpts)
}

// PlasmaChain is a free data retrieval call binding the contract method 0x3cdac42c.
//
// Solidity: function plasmaChain( uint256) constant returns(header bytes32, numTxns uint256, feeAmount uint256, createdAt uint256)
func (_Plasma *PlasmaCaller) PlasmaChain(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Header    [32]byte
	NumTxns   *big.Int
	FeeAmount *big.Int
	CreatedAt *big.Int
}, error) {
	ret := new(struct {
		Header    [32]byte
		NumTxns   *big.Int
		FeeAmount *big.Int
		CreatedAt *big.Int
	})
	out := ret
	err := _Plasma.contract.Call(opts, out, "plasmaChain", arg0)
	return *ret, err
}

// PlasmaChain is a free data retrieval call binding the contract method 0x3cdac42c.
//
// Solidity: function plasmaChain( uint256) constant returns(header bytes32, numTxns uint256, feeAmount uint256, createdAt uint256)
func (_Plasma *PlasmaSession) PlasmaChain(arg0 *big.Int) (struct {
	Header    [32]byte
	NumTxns   *big.Int
	FeeAmount *big.Int
	CreatedAt *big.Int
}, error) {
	return _Plasma.Contract.PlasmaChain(&_Plasma.CallOpts, arg0)
}

// PlasmaChain is a free data retrieval call binding the contract method 0x3cdac42c.
//
// Solidity: function plasmaChain( uint256) constant returns(header bytes32, numTxns uint256, feeAmount uint256, createdAt uint256)
func (_Plasma *PlasmaCallerSession) PlasmaChain(arg0 *big.Int) (struct {
	Header    [32]byte
	NumTxns   *big.Int
	FeeAmount *big.Int
	CreatedAt *big.Int
}, error) {
	return _Plasma.Contract.PlasmaChain(&_Plasma.CallOpts, arg0)
}

// PlasmaChainBalance is a free data retrieval call binding the contract method 0x45cbefa2.
//
// Solidity: function plasmaChainBalance() constant returns(uint256)
func (_Plasma *PlasmaCaller) PlasmaChainBalance(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Plasma.contract.Call(opts, out, "plasmaChainBalance")
	return *ret0, err
}

// PlasmaChainBalance is a free data retrieval call binding the contract method 0x45cbefa2.
//
// Solidity: function plasmaChainBalance() constant returns(uint256)
func (_Plasma *PlasmaSession) PlasmaChainBalance() (*big.Int, error) {
	return _Plasma.Contract.PlasmaChainBalance(&_Plasma.CallOpts)
}

// PlasmaChainBalance is a free data retrieval call binding the contract method 0x45cbefa2.
//
// Solidity: function plasmaChainBalance() constant returns(uint256)
func (_Plasma *PlasmaCallerSession) PlasmaChainBalance() (*big.Int, error) {
	return _Plasma.Contract.PlasmaChainBalance(&_Plasma.CallOpts)
}

// TotalWithdrawBalance is a free data retrieval call binding the contract method 0xc430c438.
//
// Solidity: function totalWithdrawBalance() constant returns(uint256)
func (_Plasma *PlasmaCaller) TotalWithdrawBalance(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Plasma.contract.Call(opts, out, "totalWithdrawBalance")
	return *ret0, err
}

// TotalWithdrawBalance is a free data retrieval call binding the contract method 0xc430c438.
//
// Solidity: function totalWithdrawBalance() constant returns(uint256)
func (_Plasma *PlasmaSession) TotalWithdrawBalance() (*big.Int, error) {
	return _Plasma.Contract.TotalWithdrawBalance(&_Plasma.CallOpts)
}

// TotalWithdrawBalance is a free data retrieval call binding the contract method 0xc430c438.
//
// Solidity: function totalWithdrawBalance() constant returns(uint256)
func (_Plasma *PlasmaCallerSession) TotalWithdrawBalance() (*big.Int, error) {
	return _Plasma.Contract.TotalWithdrawBalance(&_Plasma.CallOpts)
}

// TxExitQueue is a free data retrieval call binding the contract method 0x875b8ea0.
//
// Solidity: function txExitQueue( uint256) constant returns(uint256)
func (_Plasma *PlasmaCaller) TxExitQueue(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Plasma.contract.Call(opts, out, "txExitQueue", arg0)
	return *ret0, err
}

// TxExitQueue is a free data retrieval call binding the contract method 0x875b8ea0.
//
// Solidity: function txExitQueue( uint256) constant returns(uint256)
func (_Plasma *PlasmaSession) TxExitQueue(arg0 *big.Int) (*big.Int, error) {
	return _Plasma.Contract.TxExitQueue(&_Plasma.CallOpts, arg0)
}

// TxExitQueue is a free data retrieval call binding the contract method 0x875b8ea0.
//
// Solidity: function txExitQueue( uint256) constant returns(uint256)
func (_Plasma *PlasmaCallerSession) TxExitQueue(arg0 *big.Int) (*big.Int, error) {
	return _Plasma.Contract.TxExitQueue(&_Plasma.CallOpts, arg0)
}

// TxExits is a free data retrieval call binding the contract method 0x6d3d8b1a.
//
// Solidity: function txExits( uint256) constant returns(amount uint256, committedFee uint256, createdAt uint256, owner address, state uint8)
func (_Plasma *PlasmaCaller) TxExits(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Amount       *big.Int
	CommittedFee *big.Int
	CreatedAt    *big.Int
	Owner        common.Address
	State        uint8
}, error) {
	ret := new(struct {
		Amount       *big.Int
		CommittedFee *big.Int
		CreatedAt    *big.Int
		Owner        common.Address
		State        uint8
	})
	out := ret
	err := _Plasma.contract.Call(opts, out, "txExits", arg0)
	return *ret, err
}

// TxExits is a free data retrieval call binding the contract method 0x6d3d8b1a.
//
// Solidity: function txExits( uint256) constant returns(amount uint256, committedFee uint256, createdAt uint256, owner address, state uint8)
func (_Plasma *PlasmaSession) TxExits(arg0 *big.Int) (struct {
	Amount       *big.Int
	CommittedFee *big.Int
	CreatedAt    *big.Int
	Owner        common.Address
	State        uint8
}, error) {
	return _Plasma.Contract.TxExits(&_Plasma.CallOpts, arg0)
}

// TxExits is a free data retrieval call binding the contract method 0x6d3d8b1a.
//
// Solidity: function txExits( uint256) constant returns(amount uint256, committedFee uint256, createdAt uint256, owner address, state uint8)
func (_Plasma *PlasmaCallerSession) TxExits(arg0 *big.Int) (struct {
	Amount       *big.Int
	CommittedFee *big.Int
	CreatedAt    *big.Int
	Owner        common.Address
	State        uint8
}, error) {
	return _Plasma.Contract.TxExits(&_Plasma.CallOpts, arg0)
}

// ChallengeExit is a paid mutator transaction binding the contract method 0xd344e8e4.
//
// Solidity: function challengeExit(exitingTxPos uint256[4], challengingTxPos uint256[2], txBytes bytes, proof bytes, confirmSignature bytes) returns()
func (_Plasma *PlasmaTransactor) ChallengeExit(opts *bind.TransactOpts, exitingTxPos [4]*big.Int, challengingTxPos [2]*big.Int, txBytes []byte, proof []byte, confirmSignature []byte) (*types.Transaction, error) {
	return _Plasma.contract.Transact(opts, "challengeExit", exitingTxPos, challengingTxPos, txBytes, proof, confirmSignature)
}

// ChallengeExit is a paid mutator transaction binding the contract method 0xd344e8e4.
//
// Solidity: function challengeExit(exitingTxPos uint256[4], challengingTxPos uint256[2], txBytes bytes, proof bytes, confirmSignature bytes) returns()
func (_Plasma *PlasmaSession) ChallengeExit(exitingTxPos [4]*big.Int, challengingTxPos [2]*big.Int, txBytes []byte, proof []byte, confirmSignature []byte) (*types.Transaction, error) {
	return _Plasma.Contract.ChallengeExit(&_Plasma.TransactOpts, exitingTxPos, challengingTxPos, txBytes, proof, confirmSignature)
}

// ChallengeExit is a paid mutator transaction binding the contract method 0xd344e8e4.
//
// Solidity: function challengeExit(exitingTxPos uint256[4], challengingTxPos uint256[2], txBytes bytes, proof bytes, confirmSignature bytes) returns()
func (_Plasma *PlasmaTransactorSession) ChallengeExit(exitingTxPos [4]*big.Int, challengingTxPos [2]*big.Int, txBytes []byte, proof []byte, confirmSignature []byte) (*types.Transaction, error) {
	return _Plasma.Contract.ChallengeExit(&_Plasma.TransactOpts, exitingTxPos, challengingTxPos, txBytes, proof, confirmSignature)
}

// ChangeOperator is a paid mutator transaction binding the contract method 0x06394c9b.
//
// Solidity: function changeOperator(newOperator address) returns()
func (_Plasma *PlasmaTransactor) ChangeOperator(opts *bind.TransactOpts, newOperator common.Address) (*types.Transaction, error) {
	return _Plasma.contract.Transact(opts, "changeOperator", newOperator)
}

// ChangeOperator is a paid mutator transaction binding the contract method 0x06394c9b.
//
// Solidity: function changeOperator(newOperator address) returns()
func (_Plasma *PlasmaSession) ChangeOperator(newOperator common.Address) (*types.Transaction, error) {
	return _Plasma.Contract.ChangeOperator(&_Plasma.TransactOpts, newOperator)
}

// ChangeOperator is a paid mutator transaction binding the contract method 0x06394c9b.
//
// Solidity: function changeOperator(newOperator address) returns()
func (_Plasma *PlasmaTransactorSession) ChangeOperator(newOperator common.Address) (*types.Transaction, error) {
	return _Plasma.Contract.ChangeOperator(&_Plasma.TransactOpts, newOperator)
}

// Deposit is a paid mutator transaction binding the contract method 0xf340fa01.
//
// Solidity: function deposit(owner address) returns()
func (_Plasma *PlasmaTransactor) Deposit(opts *bind.TransactOpts, owner common.Address) (*types.Transaction, error) {
	return _Plasma.contract.Transact(opts, "deposit", owner)
}

// Deposit is a paid mutator transaction binding the contract method 0xf340fa01.
//
// Solidity: function deposit(owner address) returns()
func (_Plasma *PlasmaSession) Deposit(owner common.Address) (*types.Transaction, error) {
	return _Plasma.Contract.Deposit(&_Plasma.TransactOpts, owner)
}

// Deposit is a paid mutator transaction binding the contract method 0xf340fa01.
//
// Solidity: function deposit(owner address) returns()
func (_Plasma *PlasmaTransactorSession) Deposit(owner common.Address) (*types.Transaction, error) {
	return _Plasma.Contract.Deposit(&_Plasma.TransactOpts, owner)
}

// FinalizeDepositExits is a paid mutator transaction binding the contract method 0xfcf5f9eb.
//
// Solidity: function finalizeDepositExits() returns()
func (_Plasma *PlasmaTransactor) FinalizeDepositExits(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Plasma.contract.Transact(opts, "finalizeDepositExits")
}

// FinalizeDepositExits is a paid mutator transaction binding the contract method 0xfcf5f9eb.
//
// Solidity: function finalizeDepositExits() returns()
func (_Plasma *PlasmaSession) FinalizeDepositExits() (*types.Transaction, error) {
	return _Plasma.Contract.FinalizeDepositExits(&_Plasma.TransactOpts)
}

// FinalizeDepositExits is a paid mutator transaction binding the contract method 0xfcf5f9eb.
//
// Solidity: function finalizeDepositExits() returns()
func (_Plasma *PlasmaTransactorSession) FinalizeDepositExits() (*types.Transaction, error) {
	return _Plasma.Contract.FinalizeDepositExits(&_Plasma.TransactOpts)
}

// FinalizeTransactionExits is a paid mutator transaction binding the contract method 0x884fc7d6.
//
// Solidity: function finalizeTransactionExits() returns()
func (_Plasma *PlasmaTransactor) FinalizeTransactionExits(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Plasma.contract.Transact(opts, "finalizeTransactionExits")
}

// FinalizeTransactionExits is a paid mutator transaction binding the contract method 0x884fc7d6.
//
// Solidity: function finalizeTransactionExits() returns()
func (_Plasma *PlasmaSession) FinalizeTransactionExits() (*types.Transaction, error) {
	return _Plasma.Contract.FinalizeTransactionExits(&_Plasma.TransactOpts)
}

// FinalizeTransactionExits is a paid mutator transaction binding the contract method 0x884fc7d6.
//
// Solidity: function finalizeTransactionExits() returns()
func (_Plasma *PlasmaTransactorSession) FinalizeTransactionExits() (*types.Transaction, error) {
	return _Plasma.Contract.FinalizeTransactionExits(&_Plasma.TransactOpts)
}

// StartDepositExit is a paid mutator transaction binding the contract method 0x70e4abf6.
//
// Solidity: function startDepositExit(nonce uint256, committedFee uint256) returns()
func (_Plasma *PlasmaTransactor) StartDepositExit(opts *bind.TransactOpts, nonce *big.Int, committedFee *big.Int) (*types.Transaction, error) {
	return _Plasma.contract.Transact(opts, "startDepositExit", nonce, committedFee)
}

// StartDepositExit is a paid mutator transaction binding the contract method 0x70e4abf6.
//
// Solidity: function startDepositExit(nonce uint256, committedFee uint256) returns()
func (_Plasma *PlasmaSession) StartDepositExit(nonce *big.Int, committedFee *big.Int) (*types.Transaction, error) {
	return _Plasma.Contract.StartDepositExit(&_Plasma.TransactOpts, nonce, committedFee)
}

// StartDepositExit is a paid mutator transaction binding the contract method 0x70e4abf6.
//
// Solidity: function startDepositExit(nonce uint256, committedFee uint256) returns()
func (_Plasma *PlasmaTransactorSession) StartDepositExit(nonce *big.Int, committedFee *big.Int) (*types.Transaction, error) {
	return _Plasma.Contract.StartDepositExit(&_Plasma.TransactOpts, nonce, committedFee)
}

// StartFeeExit is a paid mutator transaction binding the contract method 0xed1695df.
//
// Solidity: function startFeeExit(blockNumber uint256, committedFee uint256) returns()
func (_Plasma *PlasmaTransactor) StartFeeExit(opts *bind.TransactOpts, blockNumber *big.Int, committedFee *big.Int) (*types.Transaction, error) {
	return _Plasma.contract.Transact(opts, "startFeeExit", blockNumber, committedFee)
}

// StartFeeExit is a paid mutator transaction binding the contract method 0xed1695df.
//
// Solidity: function startFeeExit(blockNumber uint256, committedFee uint256) returns()
func (_Plasma *PlasmaSession) StartFeeExit(blockNumber *big.Int, committedFee *big.Int) (*types.Transaction, error) {
	return _Plasma.Contract.StartFeeExit(&_Plasma.TransactOpts, blockNumber, committedFee)
}

// StartFeeExit is a paid mutator transaction binding the contract method 0xed1695df.
//
// Solidity: function startFeeExit(blockNumber uint256, committedFee uint256) returns()
func (_Plasma *PlasmaTransactorSession) StartFeeExit(blockNumber *big.Int, committedFee *big.Int) (*types.Transaction, error) {
	return _Plasma.Contract.StartFeeExit(&_Plasma.TransactOpts, blockNumber, committedFee)
}

// StartTransactionExit is a paid mutator transaction binding the contract method 0xcf024ea6.
//
// Solidity: function startTransactionExit(txPos uint256[3], txBytes bytes, proof bytes, confirmSignatures bytes, committedFee uint256) returns()
func (_Plasma *PlasmaTransactor) StartTransactionExit(opts *bind.TransactOpts, txPos [3]*big.Int, txBytes []byte, proof []byte, confirmSignatures []byte, committedFee *big.Int) (*types.Transaction, error) {
	return _Plasma.contract.Transact(opts, "startTransactionExit", txPos, txBytes, proof, confirmSignatures, committedFee)
}

// StartTransactionExit is a paid mutator transaction binding the contract method 0xcf024ea6.
//
// Solidity: function startTransactionExit(txPos uint256[3], txBytes bytes, proof bytes, confirmSignatures bytes, committedFee uint256) returns()
func (_Plasma *PlasmaSession) StartTransactionExit(txPos [3]*big.Int, txBytes []byte, proof []byte, confirmSignatures []byte, committedFee *big.Int) (*types.Transaction, error) {
	return _Plasma.Contract.StartTransactionExit(&_Plasma.TransactOpts, txPos, txBytes, proof, confirmSignatures, committedFee)
}

// StartTransactionExit is a paid mutator transaction binding the contract method 0xcf024ea6.
//
// Solidity: function startTransactionExit(txPos uint256[3], txBytes bytes, proof bytes, confirmSignatures bytes, committedFee uint256) returns()
func (_Plasma *PlasmaTransactorSession) StartTransactionExit(txPos [3]*big.Int, txBytes []byte, proof []byte, confirmSignatures []byte, committedFee *big.Int) (*types.Transaction, error) {
	return _Plasma.Contract.StartTransactionExit(&_Plasma.TransactOpts, txPos, txBytes, proof, confirmSignatures, committedFee)
}

// SubmitBlock is a paid mutator transaction binding the contract method 0xd84ba62f.
//
// Solidity: function submitBlock(headers bytes32[], txnsPerBlock uint256[], feePerBlock uint256[], blockNum uint256) returns()
func (_Plasma *PlasmaTransactor) SubmitBlock(opts *bind.TransactOpts, headers [][32]byte, txnsPerBlock []*big.Int, feePerBlock []*big.Int, blockNum *big.Int) (*types.Transaction, error) {
	return _Plasma.contract.Transact(opts, "submitBlock", headers, txnsPerBlock, feePerBlock, blockNum)
}

// SubmitBlock is a paid mutator transaction binding the contract method 0xd84ba62f.
//
// Solidity: function submitBlock(headers bytes32[], txnsPerBlock uint256[], feePerBlock uint256[], blockNum uint256) returns()
func (_Plasma *PlasmaSession) SubmitBlock(headers [][32]byte, txnsPerBlock []*big.Int, feePerBlock []*big.Int, blockNum *big.Int) (*types.Transaction, error) {
	return _Plasma.Contract.SubmitBlock(&_Plasma.TransactOpts, headers, txnsPerBlock, feePerBlock, blockNum)
}

// SubmitBlock is a paid mutator transaction binding the contract method 0xd84ba62f.
//
// Solidity: function submitBlock(headers bytes32[], txnsPerBlock uint256[], feePerBlock uint256[], blockNum uint256) returns()
func (_Plasma *PlasmaTransactorSession) SubmitBlock(headers [][32]byte, txnsPerBlock []*big.Int, feePerBlock []*big.Int, blockNum *big.Int) (*types.Transaction, error) {
	return _Plasma.Contract.SubmitBlock(&_Plasma.TransactOpts, headers, txnsPerBlock, feePerBlock, blockNum)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns(uint256)
func (_Plasma *PlasmaTransactor) Withdraw(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Plasma.contract.Transact(opts, "withdraw")
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns(uint256)
func (_Plasma *PlasmaSession) Withdraw() (*types.Transaction, error) {
	return _Plasma.Contract.Withdraw(&_Plasma.TransactOpts)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns(uint256)
func (_Plasma *PlasmaTransactorSession) Withdraw() (*types.Transaction, error) {
	return _Plasma.Contract.Withdraw(&_Plasma.TransactOpts)
}

// PlasmaAddedToBalancesIterator is returned from FilterAddedToBalances and is used to iterate over the raw logs and unpacked data for AddedToBalances events raised by the Plasma contract.
type PlasmaAddedToBalancesIterator struct {
	Event *PlasmaAddedToBalances // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PlasmaAddedToBalancesIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PlasmaAddedToBalances)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PlasmaAddedToBalances)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PlasmaAddedToBalancesIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PlasmaAddedToBalancesIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PlasmaAddedToBalances represents a AddedToBalances event raised by the Plasma contract.
type PlasmaAddedToBalances struct {
	Owner  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterAddedToBalances is a free log retrieval operation binding the contract event 0xf8552a24c7d58fd05114f6fc9db7b3a354db64d5fc758184af1696ccd8f158f3.
//
// Solidity: e AddedToBalances(owner address, amount uint256)
func (_Plasma *PlasmaFilterer) FilterAddedToBalances(opts *bind.FilterOpts) (*PlasmaAddedToBalancesIterator, error) {

	logs, sub, err := _Plasma.contract.FilterLogs(opts, "AddedToBalances")
	if err != nil {
		return nil, err
	}
	return &PlasmaAddedToBalancesIterator{contract: _Plasma.contract, event: "AddedToBalances", logs: logs, sub: sub}, nil
}

// WatchAddedToBalances is a free log subscription operation binding the contract event 0xf8552a24c7d58fd05114f6fc9db7b3a354db64d5fc758184af1696ccd8f158f3.
//
// Solidity: e AddedToBalances(owner address, amount uint256)
func (_Plasma *PlasmaFilterer) WatchAddedToBalances(opts *bind.WatchOpts, sink chan<- *PlasmaAddedToBalances) (event.Subscription, error) {

	logs, sub, err := _Plasma.contract.WatchLogs(opts, "AddedToBalances")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PlasmaAddedToBalances)
				if err := _Plasma.contract.UnpackLog(event, "AddedToBalances", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// PlasmaBlockSubmittedIterator is returned from FilterBlockSubmitted and is used to iterate over the raw logs and unpacked data for BlockSubmitted events raised by the Plasma contract.
type PlasmaBlockSubmittedIterator struct {
	Event *PlasmaBlockSubmitted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PlasmaBlockSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PlasmaBlockSubmitted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PlasmaBlockSubmitted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PlasmaBlockSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PlasmaBlockSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PlasmaBlockSubmitted represents a BlockSubmitted event raised by the Plasma contract.
type PlasmaBlockSubmitted struct {
	Header      [32]byte
	BlockNumber *big.Int
	NumTxns     *big.Int
	FeeAmount   *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterBlockSubmitted is a free log retrieval operation binding the contract event 0x044ff3798f9b3ad55d1155cea9a40508c71b4c64335f5dae87e8e11551515a06.
//
// Solidity: e BlockSubmitted(header bytes32, blockNumber uint256, numTxns uint256, feeAmount uint256)
func (_Plasma *PlasmaFilterer) FilterBlockSubmitted(opts *bind.FilterOpts) (*PlasmaBlockSubmittedIterator, error) {

	logs, sub, err := _Plasma.contract.FilterLogs(opts, "BlockSubmitted")
	if err != nil {
		return nil, err
	}
	return &PlasmaBlockSubmittedIterator{contract: _Plasma.contract, event: "BlockSubmitted", logs: logs, sub: sub}, nil
}

// WatchBlockSubmitted is a free log subscription operation binding the contract event 0x044ff3798f9b3ad55d1155cea9a40508c71b4c64335f5dae87e8e11551515a06.
//
// Solidity: e BlockSubmitted(header bytes32, blockNumber uint256, numTxns uint256, feeAmount uint256)
func (_Plasma *PlasmaFilterer) WatchBlockSubmitted(opts *bind.WatchOpts, sink chan<- *PlasmaBlockSubmitted) (event.Subscription, error) {

	logs, sub, err := _Plasma.contract.WatchLogs(opts, "BlockSubmitted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PlasmaBlockSubmitted)
				if err := _Plasma.contract.UnpackLog(event, "BlockSubmitted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// PlasmaChallengedExitIterator is returned from FilterChallengedExit and is used to iterate over the raw logs and unpacked data for ChallengedExit events raised by the Plasma contract.
type PlasmaChallengedExitIterator struct {
	Event *PlasmaChallengedExit // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PlasmaChallengedExitIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PlasmaChallengedExit)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PlasmaChallengedExit)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PlasmaChallengedExitIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PlasmaChallengedExitIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PlasmaChallengedExit represents a ChallengedExit event raised by the Plasma contract.
type PlasmaChallengedExit struct {
	Position [4]*big.Int
	Owner    common.Address
	Amount   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterChallengedExit is a free log retrieval operation binding the contract event 0xe1289dafb1083e540206bcd7d95a9705ba2590d6a9229c35a1c4c4c5efbda901.
//
// Solidity: e ChallengedExit(position uint256[4], owner address, amount uint256)
func (_Plasma *PlasmaFilterer) FilterChallengedExit(opts *bind.FilterOpts) (*PlasmaChallengedExitIterator, error) {

	logs, sub, err := _Plasma.contract.FilterLogs(opts, "ChallengedExit")
	if err != nil {
		return nil, err
	}
	return &PlasmaChallengedExitIterator{contract: _Plasma.contract, event: "ChallengedExit", logs: logs, sub: sub}, nil
}

// WatchChallengedExit is a free log subscription operation binding the contract event 0xe1289dafb1083e540206bcd7d95a9705ba2590d6a9229c35a1c4c4c5efbda901.
//
// Solidity: e ChallengedExit(position uint256[4], owner address, amount uint256)
func (_Plasma *PlasmaFilterer) WatchChallengedExit(opts *bind.WatchOpts, sink chan<- *PlasmaChallengedExit) (event.Subscription, error) {

	logs, sub, err := _Plasma.contract.WatchLogs(opts, "ChallengedExit")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PlasmaChallengedExit)
				if err := _Plasma.contract.UnpackLog(event, "ChallengedExit", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// PlasmaChangedOperatorIterator is returned from FilterChangedOperator and is used to iterate over the raw logs and unpacked data for ChangedOperator events raised by the Plasma contract.
type PlasmaChangedOperatorIterator struct {
	Event *PlasmaChangedOperator // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PlasmaChangedOperatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PlasmaChangedOperator)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PlasmaChangedOperator)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PlasmaChangedOperatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PlasmaChangedOperatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PlasmaChangedOperator represents a ChangedOperator event raised by the Plasma contract.
type PlasmaChangedOperator struct {
	OldOperator common.Address
	NewOperator common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterChangedOperator is a free log retrieval operation binding the contract event 0x3aff32e6289f2f2a2463481071051456b2768bb391c64ae3c91f9033e208cda1.
//
// Solidity: e ChangedOperator(oldOperator address, newOperator address)
func (_Plasma *PlasmaFilterer) FilterChangedOperator(opts *bind.FilterOpts) (*PlasmaChangedOperatorIterator, error) {

	logs, sub, err := _Plasma.contract.FilterLogs(opts, "ChangedOperator")
	if err != nil {
		return nil, err
	}
	return &PlasmaChangedOperatorIterator{contract: _Plasma.contract, event: "ChangedOperator", logs: logs, sub: sub}, nil
}

// WatchChangedOperator is a free log subscription operation binding the contract event 0x3aff32e6289f2f2a2463481071051456b2768bb391c64ae3c91f9033e208cda1.
//
// Solidity: e ChangedOperator(oldOperator address, newOperator address)
func (_Plasma *PlasmaFilterer) WatchChangedOperator(opts *bind.WatchOpts, sink chan<- *PlasmaChangedOperator) (event.Subscription, error) {

	logs, sub, err := _Plasma.contract.WatchLogs(opts, "ChangedOperator")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PlasmaChangedOperator)
				if err := _Plasma.contract.UnpackLog(event, "ChangedOperator", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// PlasmaDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the Plasma contract.
type PlasmaDepositIterator struct {
	Event *PlasmaDeposit // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PlasmaDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PlasmaDeposit)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PlasmaDeposit)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PlasmaDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PlasmaDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PlasmaDeposit represents a Deposit event raised by the Plasma contract.
type PlasmaDeposit struct {
	Depositor    common.Address
	Amount       *big.Int
	DepositNonce *big.Int
	EthBlockNum  *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0x36af321ec8d3c75236829c5317affd40ddb308863a1236d2d277a4025cccee1e.
//
// Solidity: e Deposit(depositor address, amount uint256, depositNonce uint256, ethBlockNum uint256)
func (_Plasma *PlasmaFilterer) FilterDeposit(opts *bind.FilterOpts) (*PlasmaDepositIterator, error) {

	logs, sub, err := _Plasma.contract.FilterLogs(opts, "Deposit")
	if err != nil {
		return nil, err
	}
	return &PlasmaDepositIterator{contract: _Plasma.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0x36af321ec8d3c75236829c5317affd40ddb308863a1236d2d277a4025cccee1e.
//
// Solidity: e Deposit(depositor address, amount uint256, depositNonce uint256, ethBlockNum uint256)
func (_Plasma *PlasmaFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *PlasmaDeposit) (event.Subscription, error) {

	logs, sub, err := _Plasma.contract.WatchLogs(opts, "Deposit")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PlasmaDeposit)
				if err := _Plasma.contract.UnpackLog(event, "Deposit", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// PlasmaFinalizedExitIterator is returned from FilterFinalizedExit and is used to iterate over the raw logs and unpacked data for FinalizedExit events raised by the Plasma contract.
type PlasmaFinalizedExitIterator struct {
	Event *PlasmaFinalizedExit // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PlasmaFinalizedExitIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PlasmaFinalizedExit)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PlasmaFinalizedExit)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PlasmaFinalizedExitIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PlasmaFinalizedExitIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PlasmaFinalizedExit represents a FinalizedExit event raised by the Plasma contract.
type PlasmaFinalizedExit struct {
	Position [4]*big.Int
	Owner    common.Address
	Amount   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterFinalizedExit is a free log retrieval operation binding the contract event 0xb5083a27a38f8a9aa999efb3306b7be96dc3f42010a968dd86627880ba7fdbe2.
//
// Solidity: e FinalizedExit(position uint256[4], owner address, amount uint256)
func (_Plasma *PlasmaFilterer) FilterFinalizedExit(opts *bind.FilterOpts) (*PlasmaFinalizedExitIterator, error) {

	logs, sub, err := _Plasma.contract.FilterLogs(opts, "FinalizedExit")
	if err != nil {
		return nil, err
	}
	return &PlasmaFinalizedExitIterator{contract: _Plasma.contract, event: "FinalizedExit", logs: logs, sub: sub}, nil
}

// WatchFinalizedExit is a free log subscription operation binding the contract event 0xb5083a27a38f8a9aa999efb3306b7be96dc3f42010a968dd86627880ba7fdbe2.
//
// Solidity: e FinalizedExit(position uint256[4], owner address, amount uint256)
func (_Plasma *PlasmaFilterer) WatchFinalizedExit(opts *bind.WatchOpts, sink chan<- *PlasmaFinalizedExit) (event.Subscription, error) {

	logs, sub, err := _Plasma.contract.WatchLogs(opts, "FinalizedExit")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PlasmaFinalizedExit)
				if err := _Plasma.contract.UnpackLog(event, "FinalizedExit", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// PlasmaStartedDepositExitIterator is returned from FilterStartedDepositExit and is used to iterate over the raw logs and unpacked data for StartedDepositExit events raised by the Plasma contract.
type PlasmaStartedDepositExitIterator struct {
	Event *PlasmaStartedDepositExit // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PlasmaStartedDepositExitIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PlasmaStartedDepositExit)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PlasmaStartedDepositExit)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PlasmaStartedDepositExitIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PlasmaStartedDepositExitIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PlasmaStartedDepositExit represents a StartedDepositExit event raised by the Plasma contract.
type PlasmaStartedDepositExit struct {
	Nonce        *big.Int
	Owner        common.Address
	Amount       *big.Int
	CommittedFee *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterStartedDepositExit is a free log retrieval operation binding the contract event 0xe90dc7204eb622c87d2c8d67d8e27afdfd34042584591e7b3d35014873cf9cfd.
//
// Solidity: e StartedDepositExit(nonce uint256, owner address, amount uint256, committedFee uint256)
func (_Plasma *PlasmaFilterer) FilterStartedDepositExit(opts *bind.FilterOpts) (*PlasmaStartedDepositExitIterator, error) {

	logs, sub, err := _Plasma.contract.FilterLogs(opts, "StartedDepositExit")
	if err != nil {
		return nil, err
	}
	return &PlasmaStartedDepositExitIterator{contract: _Plasma.contract, event: "StartedDepositExit", logs: logs, sub: sub}, nil
}

// WatchStartedDepositExit is a free log subscription operation binding the contract event 0xe90dc7204eb622c87d2c8d67d8e27afdfd34042584591e7b3d35014873cf9cfd.
//
// Solidity: e StartedDepositExit(nonce uint256, owner address, amount uint256, committedFee uint256)
func (_Plasma *PlasmaFilterer) WatchStartedDepositExit(opts *bind.WatchOpts, sink chan<- *PlasmaStartedDepositExit) (event.Subscription, error) {

	logs, sub, err := _Plasma.contract.WatchLogs(opts, "StartedDepositExit")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PlasmaStartedDepositExit)
				if err := _Plasma.contract.UnpackLog(event, "StartedDepositExit", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// PlasmaStartedTransactionExitIterator is returned from FilterStartedTransactionExit and is used to iterate over the raw logs and unpacked data for StartedTransactionExit events raised by the Plasma contract.
type PlasmaStartedTransactionExitIterator struct {
	Event *PlasmaStartedTransactionExit // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PlasmaStartedTransactionExitIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PlasmaStartedTransactionExit)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PlasmaStartedTransactionExit)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PlasmaStartedTransactionExitIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PlasmaStartedTransactionExitIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PlasmaStartedTransactionExit represents a StartedTransactionExit event raised by the Plasma contract.
type PlasmaStartedTransactionExit struct {
	Position          [3]*big.Int
	Owner             common.Address
	Amount            *big.Int
	ConfirmSignatures []byte
	CommittedFee      *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterStartedTransactionExit is a free log retrieval operation binding the contract event 0x20d695720ae96d3511520c6f51d6ab23aa19a3796da77024ad027b344bb72530.
//
// Solidity: e StartedTransactionExit(position uint256[3], owner address, amount uint256, confirmSignatures bytes, committedFee uint256)
func (_Plasma *PlasmaFilterer) FilterStartedTransactionExit(opts *bind.FilterOpts) (*PlasmaStartedTransactionExitIterator, error) {

	logs, sub, err := _Plasma.contract.FilterLogs(opts, "StartedTransactionExit")
	if err != nil {
		return nil, err
	}
	return &PlasmaStartedTransactionExitIterator{contract: _Plasma.contract, event: "StartedTransactionExit", logs: logs, sub: sub}, nil
}

// WatchStartedTransactionExit is a free log subscription operation binding the contract event 0x20d695720ae96d3511520c6f51d6ab23aa19a3796da77024ad027b344bb72530.
//
// Solidity: e StartedTransactionExit(position uint256[3], owner address, amount uint256, confirmSignatures bytes, committedFee uint256)
func (_Plasma *PlasmaFilterer) WatchStartedTransactionExit(opts *bind.WatchOpts, sink chan<- *PlasmaStartedTransactionExit) (event.Subscription, error) {

	logs, sub, err := _Plasma.contract.WatchLogs(opts, "StartedTransactionExit")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PlasmaStartedTransactionExit)
				if err := _Plasma.contract.UnpackLog(event, "StartedTransactionExit", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}
