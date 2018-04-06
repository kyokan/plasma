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

// PriorityQueueABI is the input ABI used to generate the binding from.
const PriorityQueueABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"priority\",\"type\":\"uint256\"}],\"name\":\"add\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"bubbleUp\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"remove\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"max\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"prune\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"getPriorities\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"pop\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"i\",\"type\":\"uint256\"}],\"name\":\"bubbleDown\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"priorities\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"item\",\"type\":\"uint256\"}],\"name\":\"DebugUint\",\"type\":\"event\"}]"

// PriorityQueue is an auto generated Go binding around an Ethereum contract.
type PriorityQueue struct {
	PriorityQueueCaller     // Read-only binding to the contract
	PriorityQueueTransactor // Write-only binding to the contract
	PriorityQueueFilterer   // Log filterer for contract events
}

// PriorityQueueCaller is an auto generated read-only Go binding around an Ethereum contract.
type PriorityQueueCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PriorityQueueTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PriorityQueueTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PriorityQueueFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PriorityQueueFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PriorityQueueSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PriorityQueueSession struct {
	Contract     *PriorityQueue    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PriorityQueueCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PriorityQueueCallerSession struct {
	Contract *PriorityQueueCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// PriorityQueueTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PriorityQueueTransactorSession struct {
	Contract     *PriorityQueueTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// PriorityQueueRaw is an auto generated low-level Go binding around an Ethereum contract.
type PriorityQueueRaw struct {
	Contract *PriorityQueue // Generic contract binding to access the raw methods on
}

// PriorityQueueCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PriorityQueueCallerRaw struct {
	Contract *PriorityQueueCaller // Generic read-only contract binding to access the raw methods on
}

// PriorityQueueTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PriorityQueueTransactorRaw struct {
	Contract *PriorityQueueTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPriorityQueue creates a new instance of PriorityQueue, bound to a specific deployed contract.
func NewPriorityQueue(address common.Address, backend bind.ContractBackend) (*PriorityQueue, error) {
	contract, err := bindPriorityQueue(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PriorityQueue{PriorityQueueCaller: PriorityQueueCaller{contract: contract}, PriorityQueueTransactor: PriorityQueueTransactor{contract: contract}, PriorityQueueFilterer: PriorityQueueFilterer{contract: contract}}, nil
}

// NewPriorityQueueCaller creates a new read-only instance of PriorityQueue, bound to a specific deployed contract.
func NewPriorityQueueCaller(address common.Address, caller bind.ContractCaller) (*PriorityQueueCaller, error) {
	contract, err := bindPriorityQueue(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PriorityQueueCaller{contract: contract}, nil
}

// NewPriorityQueueTransactor creates a new write-only instance of PriorityQueue, bound to a specific deployed contract.
func NewPriorityQueueTransactor(address common.Address, transactor bind.ContractTransactor) (*PriorityQueueTransactor, error) {
	contract, err := bindPriorityQueue(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PriorityQueueTransactor{contract: contract}, nil
}

// NewPriorityQueueFilterer creates a new log filterer instance of PriorityQueue, bound to a specific deployed contract.
func NewPriorityQueueFilterer(address common.Address, filterer bind.ContractFilterer) (*PriorityQueueFilterer, error) {
	contract, err := bindPriorityQueue(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PriorityQueueFilterer{contract: contract}, nil
}

// bindPriorityQueue binds a generic wrapper to an already deployed contract.
func bindPriorityQueue(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(PriorityQueueABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PriorityQueue *PriorityQueueRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PriorityQueue.Contract.PriorityQueueCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PriorityQueue *PriorityQueueRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PriorityQueue.Contract.PriorityQueueTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PriorityQueue *PriorityQueueRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PriorityQueue.Contract.PriorityQueueTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PriorityQueue *PriorityQueueCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _PriorityQueue.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PriorityQueue *PriorityQueueTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PriorityQueue.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PriorityQueue *PriorityQueueTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PriorityQueue.Contract.contract.Transact(opts, method, params...)
}

// Max is a free data retrieval call binding the contract method 0x6ac5db19.
//
// Solidity: function max() constant returns(uint256)
func (_PriorityQueue *PriorityQueueCaller) Max(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PriorityQueue.contract.Call(opts, out, "max")
	return *ret0, err
}

// Max is a free data retrieval call binding the contract method 0x6ac5db19.
//
// Solidity: function max() constant returns(uint256)
func (_PriorityQueue *PriorityQueueSession) Max() (*big.Int, error) {
	return _PriorityQueue.Contract.Max(&_PriorityQueue.CallOpts)
}

// Max is a free data retrieval call binding the contract method 0x6ac5db19.
//
// Solidity: function max() constant returns(uint256)
func (_PriorityQueue *PriorityQueueCallerSession) Max() (*big.Int, error) {
	return _PriorityQueue.Contract.Max(&_PriorityQueue.CallOpts)
}

// Priorities is a free data retrieval call binding the contract method 0xb36f923e.
//
// Solidity: function priorities( uint256) constant returns(uint256)
func (_PriorityQueue *PriorityQueueCaller) Priorities(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _PriorityQueue.contract.Call(opts, out, "priorities", arg0)
	return *ret0, err
}

// Priorities is a free data retrieval call binding the contract method 0xb36f923e.
//
// Solidity: function priorities( uint256) constant returns(uint256)
func (_PriorityQueue *PriorityQueueSession) Priorities(arg0 *big.Int) (*big.Int, error) {
	return _PriorityQueue.Contract.Priorities(&_PriorityQueue.CallOpts, arg0)
}

// Priorities is a free data retrieval call binding the contract method 0xb36f923e.
//
// Solidity: function priorities( uint256) constant returns(uint256)
func (_PriorityQueue *PriorityQueueCallerSession) Priorities(arg0 *big.Int) (*big.Int, error) {
	return _PriorityQueue.Contract.Priorities(&_PriorityQueue.CallOpts, arg0)
}

// Add is a paid mutator transaction binding the contract method 0x1003e2d2.
//
// Solidity: function add(priority uint256) returns()
func (_PriorityQueue *PriorityQueueTransactor) Add(opts *bind.TransactOpts, priority *big.Int) (*types.Transaction, error) {
	return _PriorityQueue.contract.Transact(opts, "add", priority)
}

// Add is a paid mutator transaction binding the contract method 0x1003e2d2.
//
// Solidity: function add(priority uint256) returns()
func (_PriorityQueue *PriorityQueueSession) Add(priority *big.Int) (*types.Transaction, error) {
	return _PriorityQueue.Contract.Add(&_PriorityQueue.TransactOpts, priority)
}

// Add is a paid mutator transaction binding the contract method 0x1003e2d2.
//
// Solidity: function add(priority uint256) returns()
func (_PriorityQueue *PriorityQueueTransactorSession) Add(priority *big.Int) (*types.Transaction, error) {
	return _PriorityQueue.Contract.Add(&_PriorityQueue.TransactOpts, priority)
}

// BubbleDown is a paid mutator transaction binding the contract method 0xb1b3c085.
//
// Solidity: function bubbleDown(i uint256) returns()
func (_PriorityQueue *PriorityQueueTransactor) BubbleDown(opts *bind.TransactOpts, i *big.Int) (*types.Transaction, error) {
	return _PriorityQueue.contract.Transact(opts, "bubbleDown", i)
}

// BubbleDown is a paid mutator transaction binding the contract method 0xb1b3c085.
//
// Solidity: function bubbleDown(i uint256) returns()
func (_PriorityQueue *PriorityQueueSession) BubbleDown(i *big.Int) (*types.Transaction, error) {
	return _PriorityQueue.Contract.BubbleDown(&_PriorityQueue.TransactOpts, i)
}

// BubbleDown is a paid mutator transaction binding the contract method 0xb1b3c085.
//
// Solidity: function bubbleDown(i uint256) returns()
func (_PriorityQueue *PriorityQueueTransactorSession) BubbleDown(i *big.Int) (*types.Transaction, error) {
	return _PriorityQueue.Contract.BubbleDown(&_PriorityQueue.TransactOpts, i)
}

// BubbleUp is a paid mutator transaction binding the contract method 0x4062def3.
//
// Solidity: function bubbleUp() returns()
func (_PriorityQueue *PriorityQueueTransactor) BubbleUp(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PriorityQueue.contract.Transact(opts, "bubbleUp")
}

// BubbleUp is a paid mutator transaction binding the contract method 0x4062def3.
//
// Solidity: function bubbleUp() returns()
func (_PriorityQueue *PriorityQueueSession) BubbleUp() (*types.Transaction, error) {
	return _PriorityQueue.Contract.BubbleUp(&_PriorityQueue.TransactOpts)
}

// BubbleUp is a paid mutator transaction binding the contract method 0x4062def3.
//
// Solidity: function bubbleUp() returns()
func (_PriorityQueue *PriorityQueueTransactorSession) BubbleUp() (*types.Transaction, error) {
	return _PriorityQueue.Contract.BubbleUp(&_PriorityQueue.TransactOpts)
}

// GetPriorities is a paid mutator transaction binding the contract method 0xa0daeed2.
//
// Solidity: function getPriorities() returns()
func (_PriorityQueue *PriorityQueueTransactor) GetPriorities(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PriorityQueue.contract.Transact(opts, "getPriorities")
}

// GetPriorities is a paid mutator transaction binding the contract method 0xa0daeed2.
//
// Solidity: function getPriorities() returns()
func (_PriorityQueue *PriorityQueueSession) GetPriorities() (*types.Transaction, error) {
	return _PriorityQueue.Contract.GetPriorities(&_PriorityQueue.TransactOpts)
}

// GetPriorities is a paid mutator transaction binding the contract method 0xa0daeed2.
//
// Solidity: function getPriorities() returns()
func (_PriorityQueue *PriorityQueueTransactorSession) GetPriorities() (*types.Transaction, error) {
	return _PriorityQueue.Contract.GetPriorities(&_PriorityQueue.TransactOpts)
}

// Pop is a paid mutator transaction binding the contract method 0xa4ece52c.
//
// Solidity: function pop() returns(uint256)
func (_PriorityQueue *PriorityQueueTransactor) Pop(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PriorityQueue.contract.Transact(opts, "pop")
}

// Pop is a paid mutator transaction binding the contract method 0xa4ece52c.
//
// Solidity: function pop() returns(uint256)
func (_PriorityQueue *PriorityQueueSession) Pop() (*types.Transaction, error) {
	return _PriorityQueue.Contract.Pop(&_PriorityQueue.TransactOpts)
}

// Pop is a paid mutator transaction binding the contract method 0xa4ece52c.
//
// Solidity: function pop() returns(uint256)
func (_PriorityQueue *PriorityQueueTransactorSession) Pop() (*types.Transaction, error) {
	return _PriorityQueue.Contract.Pop(&_PriorityQueue.TransactOpts)
}

// Prune is a paid mutator transaction binding the contract method 0x9f2b3154.
//
// Solidity: function prune() returns()
func (_PriorityQueue *PriorityQueueTransactor) Prune(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PriorityQueue.contract.Transact(opts, "prune")
}

// Prune is a paid mutator transaction binding the contract method 0x9f2b3154.
//
// Solidity: function prune() returns()
func (_PriorityQueue *PriorityQueueSession) Prune() (*types.Transaction, error) {
	return _PriorityQueue.Contract.Prune(&_PriorityQueue.TransactOpts)
}

// Prune is a paid mutator transaction binding the contract method 0x9f2b3154.
//
// Solidity: function prune() returns()
func (_PriorityQueue *PriorityQueueTransactorSession) Prune() (*types.Transaction, error) {
	return _PriorityQueue.Contract.Prune(&_PriorityQueue.TransactOpts)
}

// Remove is a paid mutator transaction binding the contract method 0x4cc82215.
//
// Solidity: function remove(id uint256) returns(bool)
func (_PriorityQueue *PriorityQueueTransactor) Remove(opts *bind.TransactOpts, id *big.Int) (*types.Transaction, error) {
	return _PriorityQueue.contract.Transact(opts, "remove", id)
}

// Remove is a paid mutator transaction binding the contract method 0x4cc82215.
//
// Solidity: function remove(id uint256) returns(bool)
func (_PriorityQueue *PriorityQueueSession) Remove(id *big.Int) (*types.Transaction, error) {
	return _PriorityQueue.Contract.Remove(&_PriorityQueue.TransactOpts, id)
}

// Remove is a paid mutator transaction binding the contract method 0x4cc82215.
//
// Solidity: function remove(id uint256) returns(bool)
func (_PriorityQueue *PriorityQueueTransactorSession) Remove(id *big.Int) (*types.Transaction, error) {
	return _PriorityQueue.Contract.Remove(&_PriorityQueue.TransactOpts, id)
}

// PriorityQueueDebugUintIterator is returned from FilterDebugUint and is used to iterate over the raw logs and unpacked data for DebugUint events raised by the PriorityQueue contract.
type PriorityQueueDebugUintIterator struct {
	Event *PriorityQueueDebugUint // Event containing the contract specifics and raw log

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
func (it *PriorityQueueDebugUintIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PriorityQueueDebugUint)
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
		it.Event = new(PriorityQueueDebugUint)
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
func (it *PriorityQueueDebugUintIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PriorityQueueDebugUintIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PriorityQueueDebugUint represents a DebugUint event raised by the PriorityQueue contract.
type PriorityQueueDebugUint struct {
	Sender common.Address
	Item   *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDebugUint is a free log retrieval operation binding the contract event 0xeb3fd3624cedb01be2e0078746766b8487dd5fb1923b6d1721838f1eebb7c792.
//
// Solidity: event DebugUint(sender address, item uint256)
func (_PriorityQueue *PriorityQueueFilterer) FilterDebugUint(opts *bind.FilterOpts) (*PriorityQueueDebugUintIterator, error) {

	logs, sub, err := _PriorityQueue.contract.FilterLogs(opts, "DebugUint")
	if err != nil {
		return nil, err
	}
	return &PriorityQueueDebugUintIterator{contract: _PriorityQueue.contract, event: "DebugUint", logs: logs, sub: sub}, nil
}

// WatchDebugUint is a free log subscription operation binding the contract event 0xeb3fd3624cedb01be2e0078746766b8487dd5fb1923b6d1721838f1eebb7c792.
//
// Solidity: event DebugUint(sender address, item uint256)
func (_PriorityQueue *PriorityQueueFilterer) WatchDebugUint(opts *bind.WatchOpts, sink chan<- *PriorityQueueDebugUint) (event.Subscription, error) {

	logs, sub, err := _PriorityQueue.contract.WatchLogs(opts, "DebugUint")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PriorityQueueDebugUint)
				if err := _PriorityQueue.contract.UnpackLog(event, "DebugUint", log); err != nil {
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
