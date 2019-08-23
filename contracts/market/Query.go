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
	_ = abi.U256
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// QueryABI is the input ABI used to generate the binding from.
const QueryABI = "[{\"inputs\":[{\"name\":\"capacity\",\"type\":\"uint256\"},{\"name\":\"duration\",\"type\":\"uint256\"},{\"name\":\"price\",\"type\":\"uint256\"},{\"name\":\"ks\",\"type\":\"uint256\"},{\"name\":\"ps\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\",\"signature\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"data\",\"type\":\"string\"}],\"name\":\"Error\",\"type\":\"event\",\"signature\":\"0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"data\",\"type\":\"uint256\"}],\"name\":\"LogInt\",\"type\":\"event\",\"signature\":\"0xc8fa9a7021af252bc69defe2b981f7bd7858defe2a87641768fefdb8a03a07cd\"},{\"constant\":true,\"inputs\":[],\"name\":\"get\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\",\"signature\":\"0x6d4ce63c\"},{\"constant\":false,\"inputs\":[],\"name\":\"setCompleted\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\",\"signature\":\"0x2e5295ee\"}]"

// QueryBin is the compiled bytecode used for deploying new contracts.
const QueryBin = `0x608060405234801561001057600080fd5b5060405160a0806102b0833981018060405260a081101561003057600080fd5b81019080805190602001909291908051906020019092919080519060200190929190805190602001909291908051906020019092919050505060c060405190810160405280868152602001858152602001848152602001602060405190810160405280858152508152602001602060405190810160405280848152508152602001600015158152506000808201518160000155602082015181600101556040820151816002015560608201518160030160008201518160000155505060808201518160040160008201518160000155505060a08201518160050160006101000a81548160ff0219169083151502179055509050505050505050610178806101386000396000f3fe60806040526004361061004c576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680632e5295ee146100515780636d4ce63c14610080575b600080fd5b34801561005d57600080fd5b506100666100d2565b604051808215151515815260200191505060405180910390f35b34801561008c57600080fd5b506100956100f9565b6040518087815260200186815260200185815260200184815260200183815260200182151515158152602001965050505050505060405180910390f35b60006001600060050160006101000a81548160ff0219169083151502179055506001905090565b6000806000806000806000800154600060010154600060020154600060030160000154600060040160000154600060050160009054906101000a900460ff1695509550955095509550955090919293949556fea165627a7a72305820915ec0079af3803174b557701e8782aa7859b50cda0fc6bab210a2d2f9be077f0029`

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
// Solidity: function get() constant returns(uint256, uint256, uint256, uint256, uint256, bool)
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
// Solidity: function get() constant returns(uint256, uint256, uint256, uint256, uint256, bool)
func (_Query *QuerySession) Get() (*big.Int, *big.Int, *big.Int, *big.Int, *big.Int, bool, error) {
	return _Query.Contract.Get(&_Query.CallOpts)
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() constant returns(uint256, uint256, uint256, uint256, uint256, bool)
func (_Query *QueryCallerSession) Get() (*big.Int, *big.Int, *big.Int, *big.Int, *big.Int, bool, error) {
	return _Query.Contract.Get(&_Query.CallOpts)
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

// QueryErrorIterator is returned from FilterError and is used to iterate over the raw logs and unpacked data for Error events raised by the Query contract.
type QueryErrorIterator struct {
	Event *QueryError // Event containing the contract specifics and raw log

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
func (it *QueryErrorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(QueryError)
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
		it.Event = new(QueryError)
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
func (it *QueryErrorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *QueryErrorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// QueryError represents a Error event raised by the Query contract.
type QueryError struct {
	Data string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterError is a free log retrieval operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Query *QueryFilterer) FilterError(opts *bind.FilterOpts) (*QueryErrorIterator, error) {

	logs, sub, err := _Query.contract.FilterLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return &QueryErrorIterator{contract: _Query.contract, event: "Error", logs: logs, sub: sub}, nil
}

// WatchError is a free log subscription operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Query *QueryFilterer) WatchError(opts *bind.WatchOpts, sink chan<- *QueryError) (event.Subscription, error) {

	logs, sub, err := _Query.contract.WatchLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(QueryError)
				if err := _Query.contract.UnpackLog(event, "Error", log); err != nil {
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

// QueryLogIntIterator is returned from FilterLogInt and is used to iterate over the raw logs and unpacked data for LogInt events raised by the Query contract.
type QueryLogIntIterator struct {
	Event *QueryLogInt // Event containing the contract specifics and raw log

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
func (it *QueryLogIntIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(QueryLogInt)
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
		it.Event = new(QueryLogInt)
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
func (it *QueryLogIntIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *QueryLogIntIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// QueryLogInt represents a LogInt event raised by the Query contract.
type QueryLogInt struct {
	Data *big.Int
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterLogInt is a free log retrieval operation binding the contract event 0xc8fa9a7021af252bc69defe2b981f7bd7858defe2a87641768fefdb8a03a07cd.
//
// Solidity: event LogInt(uint256 indexed data)
func (_Query *QueryFilterer) FilterLogInt(opts *bind.FilterOpts, data []*big.Int) (*QueryLogIntIterator, error) {

	var dataRule []interface{}
	for _, dataItem := range data {
		dataRule = append(dataRule, dataItem)
	}

	logs, sub, err := _Query.contract.FilterLogs(opts, "LogInt", dataRule)
	if err != nil {
		return nil, err
	}
	return &QueryLogIntIterator{contract: _Query.contract, event: "LogInt", logs: logs, sub: sub}, nil
}

// WatchLogInt is a free log subscription operation binding the contract event 0xc8fa9a7021af252bc69defe2b981f7bd7858defe2a87641768fefdb8a03a07cd.
//
// Solidity: event LogInt(uint256 indexed data)
func (_Query *QueryFilterer) WatchLogInt(opts *bind.WatchOpts, sink chan<- *QueryLogInt, data []*big.Int) (event.Subscription, error) {

	var dataRule []interface{}
	for _, dataItem := range data {
		dataRule = append(dataRule, dataItem)
	}

	logs, sub, err := _Query.contract.WatchLogs(opts, "LogInt", dataRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(QueryLogInt)
				if err := _Query.contract.UnpackLog(event, "LogInt", log); err != nil {
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
