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
const PlasmaABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"exits\",\"outputs\":[{\"name\":\"owner\",\"type\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"exitsIndexes\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"txBytes\",\"type\":\"bytes\"}],\"name\":\"createSimpleMerkleRoot\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"currentChildBlock\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"txBytes\",\"type\":\"bytes\"}],\"name\":\"deposit\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"blocknum\",\"type\":\"uint256\"},{\"name\":\"txindex\",\"type\":\"uint256\"},{\"name\":\"oindex\",\"type\":\"uint256\"},{\"name\":\"txBytes\",\"type\":\"bytes\"},{\"name\":\"proof\",\"type\":\"bytes\"}],\"name\":\"startExit\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"root\",\"type\":\"bytes32\"}],\"name\":\"submitBlock\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"authority\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"childChain\",\"outputs\":[{\"name\":\"root\",\"type\":\"bytes32\"},{\"name\":\"created_at\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"root\",\"type\":\"bytes32\"}],\"name\":\"SubmitBlock\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"blocknum\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"txindex\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"oindex\",\"type\":\"uint256\"}],\"name\":\"ExitStarted\",\"type\":\"event\"}]"

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

// Authority is a free data retrieval call binding the contract method 0xbf7e214f.
//
// Solidity: function authority() constant returns(address)
func (_Plasma *PlasmaCaller) Authority(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Plasma.contract.Call(opts, out, "authority")
	return *ret0, err
}

// Authority is a free data retrieval call binding the contract method 0xbf7e214f.
//
// Solidity: function authority() constant returns(address)
func (_Plasma *PlasmaSession) Authority() (common.Address, error) {
	return _Plasma.Contract.Authority(&_Plasma.CallOpts)
}

// Authority is a free data retrieval call binding the contract method 0xbf7e214f.
//
// Solidity: function authority() constant returns(address)
func (_Plasma *PlasmaCallerSession) Authority() (common.Address, error) {
	return _Plasma.Contract.Authority(&_Plasma.CallOpts)
}

// ChildChain is a free data retrieval call binding the contract method 0xf95643b1.
//
// Solidity: function childChain( uint256) constant returns(root bytes32, created_at uint256)
func (_Plasma *PlasmaCaller) ChildChain(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Root      [32]byte
	CreatedAt *big.Int
}, error) {
	ret := new(struct {
		Root      [32]byte
		CreatedAt *big.Int
	})
	out := ret
	err := _Plasma.contract.Call(opts, out, "childChain", arg0)
	return *ret, err
}

// ChildChain is a free data retrieval call binding the contract method 0xf95643b1.
//
// Solidity: function childChain( uint256) constant returns(root bytes32, created_at uint256)
func (_Plasma *PlasmaSession) ChildChain(arg0 *big.Int) (struct {
	Root      [32]byte
	CreatedAt *big.Int
}, error) {
	return _Plasma.Contract.ChildChain(&_Plasma.CallOpts, arg0)
}

// ChildChain is a free data retrieval call binding the contract method 0xf95643b1.
//
// Solidity: function childChain( uint256) constant returns(root bytes32, created_at uint256)
func (_Plasma *PlasmaCallerSession) ChildChain(arg0 *big.Int) (struct {
	Root      [32]byte
	CreatedAt *big.Int
}, error) {
	return _Plasma.Contract.ChildChain(&_Plasma.CallOpts, arg0)
}

// CurrentChildBlock is a free data retrieval call binding the contract method 0x7a95f1e8.
//
// Solidity: function currentChildBlock() constant returns(uint256)
func (_Plasma *PlasmaCaller) CurrentChildBlock(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Plasma.contract.Call(opts, out, "currentChildBlock")
	return *ret0, err
}

// CurrentChildBlock is a free data retrieval call binding the contract method 0x7a95f1e8.
//
// Solidity: function currentChildBlock() constant returns(uint256)
func (_Plasma *PlasmaSession) CurrentChildBlock() (*big.Int, error) {
	return _Plasma.Contract.CurrentChildBlock(&_Plasma.CallOpts)
}

// CurrentChildBlock is a free data retrieval call binding the contract method 0x7a95f1e8.
//
// Solidity: function currentChildBlock() constant returns(uint256)
func (_Plasma *PlasmaCallerSession) CurrentChildBlock() (*big.Int, error) {
	return _Plasma.Contract.CurrentChildBlock(&_Plasma.CallOpts)
}

// Exits is a free data retrieval call binding the contract method 0x342de179.
//
// Solidity: function exits( uint256) constant returns(owner address, amount uint256)
func (_Plasma *PlasmaCaller) Exits(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Owner  common.Address
	Amount *big.Int
}, error) {
	ret := new(struct {
		Owner  common.Address
		Amount *big.Int
	})
	out := ret
	err := _Plasma.contract.Call(opts, out, "exits", arg0)
	return *ret, err
}

// Exits is a free data retrieval call binding the contract method 0x342de179.
//
// Solidity: function exits( uint256) constant returns(owner address, amount uint256)
func (_Plasma *PlasmaSession) Exits(arg0 *big.Int) (struct {
	Owner  common.Address
	Amount *big.Int
}, error) {
	return _Plasma.Contract.Exits(&_Plasma.CallOpts, arg0)
}

// Exits is a free data retrieval call binding the contract method 0x342de179.
//
// Solidity: function exits( uint256) constant returns(owner address, amount uint256)
func (_Plasma *PlasmaCallerSession) Exits(arg0 *big.Int) (struct {
	Owner  common.Address
	Amount *big.Int
}, error) {
	return _Plasma.Contract.Exits(&_Plasma.CallOpts, arg0)
}

// ExitsIndexes is a free data retrieval call binding the contract method 0x48da92ca.
//
// Solidity: function exitsIndexes( uint256) constant returns(uint256)
func (_Plasma *PlasmaCaller) ExitsIndexes(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Plasma.contract.Call(opts, out, "exitsIndexes", arg0)
	return *ret0, err
}

// ExitsIndexes is a free data retrieval call binding the contract method 0x48da92ca.
//
// Solidity: function exitsIndexes( uint256) constant returns(uint256)
func (_Plasma *PlasmaSession) ExitsIndexes(arg0 *big.Int) (*big.Int, error) {
	return _Plasma.Contract.ExitsIndexes(&_Plasma.CallOpts, arg0)
}

// ExitsIndexes is a free data retrieval call binding the contract method 0x48da92ca.
//
// Solidity: function exitsIndexes( uint256) constant returns(uint256)
func (_Plasma *PlasmaCallerSession) ExitsIndexes(arg0 *big.Int) (*big.Int, error) {
	return _Plasma.Contract.ExitsIndexes(&_Plasma.CallOpts, arg0)
}

// CreateSimpleMerkleRoot is a paid mutator transaction binding the contract method 0x751f4c0f.
//
// Solidity: function createSimpleMerkleRoot(txBytes bytes) returns(bytes32)
func (_Plasma *PlasmaTransactor) CreateSimpleMerkleRoot(opts *bind.TransactOpts, txBytes []byte) (*types.Transaction, error) {
	return _Plasma.contract.Transact(opts, "createSimpleMerkleRoot", txBytes)
}

// CreateSimpleMerkleRoot is a paid mutator transaction binding the contract method 0x751f4c0f.
//
// Solidity: function createSimpleMerkleRoot(txBytes bytes) returns(bytes32)
func (_Plasma *PlasmaSession) CreateSimpleMerkleRoot(txBytes []byte) (*types.Transaction, error) {
	return _Plasma.Contract.CreateSimpleMerkleRoot(&_Plasma.TransactOpts, txBytes)
}

// CreateSimpleMerkleRoot is a paid mutator transaction binding the contract method 0x751f4c0f.
//
// Solidity: function createSimpleMerkleRoot(txBytes bytes) returns(bytes32)
func (_Plasma *PlasmaTransactorSession) CreateSimpleMerkleRoot(txBytes []byte) (*types.Transaction, error) {
	return _Plasma.Contract.CreateSimpleMerkleRoot(&_Plasma.TransactOpts, txBytes)
}

// Deposit is a paid mutator transaction binding the contract method 0x98b1e06a.
//
// Solidity: function deposit(txBytes bytes) returns()
func (_Plasma *PlasmaTransactor) Deposit(opts *bind.TransactOpts, txBytes []byte) (*types.Transaction, error) {
	return _Plasma.contract.Transact(opts, "deposit", txBytes)
}

// Deposit is a paid mutator transaction binding the contract method 0x98b1e06a.
//
// Solidity: function deposit(txBytes bytes) returns()
func (_Plasma *PlasmaSession) Deposit(txBytes []byte) (*types.Transaction, error) {
	return _Plasma.Contract.Deposit(&_Plasma.TransactOpts, txBytes)
}

// Deposit is a paid mutator transaction binding the contract method 0x98b1e06a.
//
// Solidity: function deposit(txBytes bytes) returns()
func (_Plasma *PlasmaTransactorSession) Deposit(txBytes []byte) (*types.Transaction, error) {
	return _Plasma.Contract.Deposit(&_Plasma.TransactOpts, txBytes)
}

// StartExit is a paid mutator transaction binding the contract method 0xacef16c5.
//
// Solidity: function startExit(blocknum uint256, txindex uint256, oindex uint256, txBytes bytes, proof bytes) returns()
func (_Plasma *PlasmaTransactor) StartExit(opts *bind.TransactOpts, blocknum *big.Int, txindex *big.Int, oindex *big.Int, txBytes []byte, proof []byte) (*types.Transaction, error) {
	return _Plasma.contract.Transact(opts, "startExit", blocknum, txindex, oindex, txBytes, proof)
}

// StartExit is a paid mutator transaction binding the contract method 0xacef16c5.
//
// Solidity: function startExit(blocknum uint256, txindex uint256, oindex uint256, txBytes bytes, proof bytes) returns()
func (_Plasma *PlasmaSession) StartExit(blocknum *big.Int, txindex *big.Int, oindex *big.Int, txBytes []byte, proof []byte) (*types.Transaction, error) {
	return _Plasma.Contract.StartExit(&_Plasma.TransactOpts, blocknum, txindex, oindex, txBytes, proof)
}

// StartExit is a paid mutator transaction binding the contract method 0xacef16c5.
//
// Solidity: function startExit(blocknum uint256, txindex uint256, oindex uint256, txBytes bytes, proof bytes) returns()
func (_Plasma *PlasmaTransactorSession) StartExit(blocknum *big.Int, txindex *big.Int, oindex *big.Int, txBytes []byte, proof []byte) (*types.Transaction, error) {
	return _Plasma.Contract.StartExit(&_Plasma.TransactOpts, blocknum, txindex, oindex, txBytes, proof)
}

// SubmitBlock is a paid mutator transaction binding the contract method 0xbaa47694.
//
// Solidity: function submitBlock(root bytes32) returns()
func (_Plasma *PlasmaTransactor) SubmitBlock(opts *bind.TransactOpts, root [32]byte) (*types.Transaction, error) {
	return _Plasma.contract.Transact(opts, "submitBlock", root)
}

// SubmitBlock is a paid mutator transaction binding the contract method 0xbaa47694.
//
// Solidity: function submitBlock(root bytes32) returns()
func (_Plasma *PlasmaSession) SubmitBlock(root [32]byte) (*types.Transaction, error) {
	return _Plasma.Contract.SubmitBlock(&_Plasma.TransactOpts, root)
}

// SubmitBlock is a paid mutator transaction binding the contract method 0xbaa47694.
//
// Solidity: function submitBlock(root bytes32) returns()
func (_Plasma *PlasmaTransactorSession) SubmitBlock(root [32]byte) (*types.Transaction, error) {
	return _Plasma.Contract.SubmitBlock(&_Plasma.TransactOpts, root)
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
	Sender common.Address
	Value  *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(sender address, value uint256)
func (_Plasma *PlasmaFilterer) FilterDeposit(opts *bind.FilterOpts) (*PlasmaDepositIterator, error) {

	logs, sub, err := _Plasma.contract.FilterLogs(opts, "Deposit")
	if err != nil {
		return nil, err
	}
	return &PlasmaDepositIterator{contract: _Plasma.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(sender address, value uint256)
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

// PlasmaExitStartedIterator is returned from FilterExitStarted and is used to iterate over the raw logs and unpacked data for ExitStarted events raised by the Plasma contract.
type PlasmaExitStartedIterator struct {
	Event *PlasmaExitStarted // Event containing the contract specifics and raw log

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
func (it *PlasmaExitStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PlasmaExitStarted)
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
		it.Event = new(PlasmaExitStarted)
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
func (it *PlasmaExitStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PlasmaExitStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PlasmaExitStarted represents a ExitStarted event raised by the Plasma contract.
type PlasmaExitStarted struct {
	Sender   common.Address
	Blocknum *big.Int
	Txindex  *big.Int
	Oindex   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterExitStarted is a free log retrieval operation binding the contract event 0x41fb391842d4655dcf9e201494553582c224a4026944a79630a163459b2586da.
//
// Solidity: event ExitStarted(sender address, blocknum uint256, txindex uint256, oindex uint256)
func (_Plasma *PlasmaFilterer) FilterExitStarted(opts *bind.FilterOpts) (*PlasmaExitStartedIterator, error) {

	logs, sub, err := _Plasma.contract.FilterLogs(opts, "ExitStarted")
	if err != nil {
		return nil, err
	}
	return &PlasmaExitStartedIterator{contract: _Plasma.contract, event: "ExitStarted", logs: logs, sub: sub}, nil
}

// WatchExitStarted is a free log subscription operation binding the contract event 0x41fb391842d4655dcf9e201494553582c224a4026944a79630a163459b2586da.
//
// Solidity: event ExitStarted(sender address, blocknum uint256, txindex uint256, oindex uint256)
func (_Plasma *PlasmaFilterer) WatchExitStarted(opts *bind.WatchOpts, sink chan<- *PlasmaExitStarted) (event.Subscription, error) {

	logs, sub, err := _Plasma.contract.WatchLogs(opts, "ExitStarted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PlasmaExitStarted)
				if err := _Plasma.contract.UnpackLog(event, "ExitStarted", log); err != nil {
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

// PlasmaSubmitBlockIterator is returned from FilterSubmitBlock and is used to iterate over the raw logs and unpacked data for SubmitBlock events raised by the Plasma contract.
type PlasmaSubmitBlockIterator struct {
	Event *PlasmaSubmitBlock // Event containing the contract specifics and raw log

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
func (it *PlasmaSubmitBlockIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PlasmaSubmitBlock)
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
		it.Event = new(PlasmaSubmitBlock)
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
func (it *PlasmaSubmitBlockIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PlasmaSubmitBlockIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PlasmaSubmitBlock represents a SubmitBlock event raised by the Plasma contract.
type PlasmaSubmitBlock struct {
	Sender common.Address
	Root   [32]byte
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterSubmitBlock is a free log retrieval operation binding the contract event 0xc9c86cbe968b46caad5b3d440a9afa668cc1cb3ba8253cfc499ca7f3d0811e3c.
//
// Solidity: event SubmitBlock(sender address, root bytes32)
func (_Plasma *PlasmaFilterer) FilterSubmitBlock(opts *bind.FilterOpts) (*PlasmaSubmitBlockIterator, error) {

	logs, sub, err := _Plasma.contract.FilterLogs(opts, "SubmitBlock")
	if err != nil {
		return nil, err
	}
	return &PlasmaSubmitBlockIterator{contract: _Plasma.contract, event: "SubmitBlock", logs: logs, sub: sub}, nil
}

// WatchSubmitBlock is a free log subscription operation binding the contract event 0xc9c86cbe968b46caad5b3d440a9afa668cc1cb3ba8253cfc499ca7f3d0811e3c.
//
// Solidity: event SubmitBlock(sender address, root bytes32)
func (_Plasma *PlasmaFilterer) WatchSubmitBlock(opts *bind.WatchOpts, sink chan<- *PlasmaSubmitBlock) (event.Subscription, error) {

	logs, sub, err := _Plasma.contract.WatchLogs(opts, "SubmitBlock")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PlasmaSubmitBlock)
				if err := _Plasma.contract.UnpackLog(event, "SubmitBlock", log); err != nil {
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
