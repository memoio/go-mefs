// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package market

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

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// QueryABI is the input ABI used to generate the binding from.
const QueryABI = "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"capacity\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"ks\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"ps\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"get\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"setCompleted\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// QueryBin is the compiled bytecode used for deploying new contracts.
var QueryBin = "0x6080604052738026796fd7ce63eae824314aa5bacf55643e893d600760006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555034801561006557600080fd5b5060405161076f38038061076f833981810160405260a081101561008857600080fd5b810190808051906020019092919080519060200190929190805190602001909291908051906020019092919080519060200190929190505050336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506000600760009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663f49ded5a6040518163ffffffff1660e01b815260040160206040518083038186803b15801561016b57600080fd5b505afa15801561017f573d6000803e3d6000fd5b505050506040513d602081101561019557600080fd5b81019080805190602001909291905050509050600161ffff168161ffff1610610226576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260198152602001807f6465706c6f79696e672071756572792069732062616e6e65640000000000000081525060200191505060405180910390fd5b6040518060c001604052808781526020018681526020018581526020016040518060200160405280868152508152602001604051806020016040528085815250815260200160001515815250600160008201518160000155602082015181600101556040820151816002015560608201518160030160008201518160000155505060808201518160040160008201518160000155505060a08201518160050160006101000a81548160ff02191690831515021790555090505050505050505061047b806102f46000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c80630ca05f9f146100515780632e5295ee146100ab5780636d4ce63c146100cb578063893d20e81461010e575b600080fd5b6100936004803603602081101561006757600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610142565b60405180821515815260200191505060405180910390f35b6100b36102e1565b60405180821515815260200191505060405180910390f35b6100d36103c9565b604051808781526020018681526020018581526020018481526020018381526020018215158152602001965050505050505060405180910390f35b61011661041c565b604051808273ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610206576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050826000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e908184604051808373ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a16001915050919050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146103a5576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b60018060050160006101000a81548160ff0219169083151502179055506001905090565b6000806000806000806001600001546001800154600160020154600160030160000154600160040160000154600160050160009054906101000a900460ff16955095509550955095509550909192939495565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1690509056fea2646970667358221220e71a1ba99ba8cd4f8c85c01fb8662c747718a852de9072d9dbc5c4b9c6d786f164736f6c63430007030033"

// DeployQuery deploys a new Ethereum contract, binding an instance of Query to it.
func DeployQuery(auth *bind.TransactOpts, backend bind.ContractBackend, capacity *big.Int, duration *big.Int, price *big.Int, ks *big.Int, ps *big.Int) (common.Address, *types.Transaction, *Query, error) {
	parsed, err := abi.JSON(strings.NewReader(QueryABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(QueryBin), backend, capacity, duration, price, ks, ps)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Query{QueryCaller: QueryCaller{contract: contract}, QueryTransactor: QueryTransactor{contract: contract}, QueryFilterer: QueryFilterer{contract: contract}}, nil
}

// Query is an auto generated Go binding around an Ethereum contract.
type Query struct {
	QueryCaller     // Read-only binding to the contract
	QueryTransactor // Write-only binding to the contract
	QueryFilterer   // Log filterer for contract events
}

// QueryCaller is an auto generated read-only Go binding around an Ethereum contract.
type QueryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// QueryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type QueryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// QueryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type QueryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// QuerySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type QuerySession struct {
	Contract     *Query            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// QueryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type QueryCallerSession struct {
	Contract *QueryCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// QueryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type QueryTransactorSession struct {
	Contract     *QueryTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// QueryRaw is an auto generated low-level Go binding around an Ethereum contract.
type QueryRaw struct {
	Contract *Query // Generic contract binding to access the raw methods on
}

// QueryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type QueryCallerRaw struct {
	Contract *QueryCaller // Generic read-only contract binding to access the raw methods on
}

// QueryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type QueryTransactorRaw struct {
	Contract *QueryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewQuery creates a new instance of Query, bound to a specific deployed contract.
func NewQuery(address common.Address, backend bind.ContractBackend) (*Query, error) {
	contract, err := bindQuery(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Query{QueryCaller: QueryCaller{contract: contract}, QueryTransactor: QueryTransactor{contract: contract}, QueryFilterer: QueryFilterer{contract: contract}}, nil
}

// NewQueryCaller creates a new read-only instance of Query, bound to a specific deployed contract.
func NewQueryCaller(address common.Address, caller bind.ContractCaller) (*QueryCaller, error) {
	contract, err := bindQuery(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &QueryCaller{contract: contract}, nil
}

// NewQueryTransactor creates a new write-only instance of Query, bound to a specific deployed contract.
func NewQueryTransactor(address common.Address, transactor bind.ContractTransactor) (*QueryTransactor, error) {
	contract, err := bindQuery(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &QueryTransactor{contract: contract}, nil
}

// NewQueryFilterer creates a new log filterer instance of Query, bound to a specific deployed contract.
func NewQueryFilterer(address common.Address, filterer bind.ContractFilterer) (*QueryFilterer, error) {
	contract, err := bindQuery(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &QueryFilterer{contract: contract}, nil
}

// bindQuery binds a generic wrapper to an already deployed contract.
func bindQuery(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(QueryABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Query *QueryRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Query.Contract.QueryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Query *QueryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Query.Contract.QueryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Query *QueryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Query.Contract.QueryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Query *QueryCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Query.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Query *QueryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Query.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Query *QueryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Query.Contract.contract.Transact(opts, method, params...)
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() view returns(uint256, uint256, uint256, uint256, uint256, bool)
func (_Query *QueryCaller) Get(opts *bind.CallOpts) (*big.Int, *big.Int, *big.Int, *big.Int, *big.Int, bool, error) {
	var (
		ret0 = new(*big.Int)
		ret1 = new(*big.Int)
		ret2 = new(*big.Int)
		ret3 = new(*big.Int)
		ret4 = new(*big.Int)
		ret5 = new(bool)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
		ret3,
		ret4,
		ret5,
	}
	err := _Query.contract.Call(opts, out, "get")
	return *ret0, *ret1, *ret2, *ret3, *ret4, *ret5, err
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() view returns(uint256, uint256, uint256, uint256, uint256, bool)
func (_Query *QuerySession) Get() (*big.Int, *big.Int, *big.Int, *big.Int, *big.Int, bool, error) {
	return _Query.Contract.Get(&_Query.CallOpts)
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() view returns(uint256, uint256, uint256, uint256, uint256, bool)
func (_Query *QueryCallerSession) Get() (*big.Int, *big.Int, *big.Int, *big.Int, *big.Int, bool, error) {
	return _Query.Contract.Get(&_Query.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_Query *QueryCaller) GetOwner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Query.contract.Call(opts, out, "getOwner")
	return *ret0, err
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_Query *QuerySession) GetOwner() (common.Address, error) {
	return _Query.Contract.GetOwner(&_Query.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_Query *QueryCallerSession) GetOwner() (common.Address, error) {
	return _Query.Contract.GetOwner(&_Query.CallOpts)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Query *QueryTransactor) AlterOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Query.contract.Transact(opts, "alterOwner", newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Query *QuerySession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Query.Contract.AlterOwner(&_Query.TransactOpts, newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Query *QueryTransactorSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Query.Contract.AlterOwner(&_Query.TransactOpts, newOwner)
}

// SetCompleted is a paid mutator transaction binding the contract method 0x2e5295ee.
//
// Solidity: function setCompleted() returns(bool)
func (_Query *QueryTransactor) SetCompleted(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Query.contract.Transact(opts, "setCompleted")
}

// SetCompleted is a paid mutator transaction binding the contract method 0x2e5295ee.
//
// Solidity: function setCompleted() returns(bool)
func (_Query *QuerySession) SetCompleted() (*types.Transaction, error) {
	return _Query.Contract.SetCompleted(&_Query.TransactOpts)
}

// SetCompleted is a paid mutator transaction binding the contract method 0x2e5295ee.
//
// Solidity: function setCompleted() returns(bool)
func (_Query *QueryTransactorSession) SetCompleted() (*types.Transaction, error) {
	return _Query.Contract.SetCompleted(&_Query.TransactOpts)
}

// QueryAlterOwnerIterator is returned from FilterAlterOwner and is used to iterate over the raw logs and unpacked data for AlterOwner events raised by the Query contract.
type QueryAlterOwnerIterator struct {
	Event *QueryAlterOwner // Event containing the contract specifics and raw log

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
func (it *QueryAlterOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(QueryAlterOwner)
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
		it.Event = new(QueryAlterOwner)
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
func (it *QueryAlterOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *QueryAlterOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// QueryAlterOwner represents a AlterOwner event raised by the Query contract.
type QueryAlterOwner struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAlterOwner is a free log retrieval operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Query *QueryFilterer) FilterAlterOwner(opts *bind.FilterOpts) (*QueryAlterOwnerIterator, error) {

	logs, sub, err := _Query.contract.FilterLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return &QueryAlterOwnerIterator{contract: _Query.contract, event: "AlterOwner", logs: logs, sub: sub}, nil
}

// WatchAlterOwner is a free log subscription operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Query *QueryFilterer) WatchAlterOwner(opts *bind.WatchOpts, sink chan<- *QueryAlterOwner) (event.Subscription, error) {

	logs, sub, err := _Query.contract.WatchLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(QueryAlterOwner)
				if err := _Query.contract.UnpackLog(event, "AlterOwner", log); err != nil {
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

// ParseAlterOwner is a log parse operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Query *QueryFilterer) ParseAlterOwner(log types.Log) (*QueryAlterOwner, error) {
	event := new(QueryAlterOwner)
	if err := _Query.contract.UnpackLog(event, "AlterOwner", log); err != nil {
		return nil, err
	}
	return event, nil
}
