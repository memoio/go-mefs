// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package mapper

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

// MapperABI is the input ABI used to generate the binding from.
const MapperABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"add\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"get\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"Add\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"data\",\"type\":\"string\"}],\"name\":\"Error\",\"type\":\"event\"}]"

// MapperBin is the compiled bytecode used for deploying new contracts.
const MapperBin = `0x6080604052600080546001600160a01b0319163317905561041b806100256000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c80630a3b0a4f146100515780630ca05f9f1461008b5780636d4ce63c146100b1578063893d20e814610109575b600080fd5b6100776004803603602081101561006757600080fd5b50356001600160a01b031661012d565b604080519115158252519081900360200190f35b610077600480360360208110156100a157600080fd5b50356001600160a01b03166102a6565b6100b9610323565b60408051602080825283518183015283519192839290830191858101910280838360005b838110156100f55781810151838201526020016100dd565b505050509050019250505060405180910390f35b610111610385565b604080516001600160a01b039092168252519081900360200190f35b600080546001600160a01b03163314156102405761014a82610394565b156101b757604080516020808252600f908201527fe5b7b2e69c89e6ada4e59cb0e59d8000000000000000000000000000000000008183015290517f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa9181900360600190a150600061023b565b6001805480820182556000919091527fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf60180546001600160a01b0384166001600160a01b0319909116811790915560408051918252517f87dc5eecd6d6bdeae407c426da6bfba5b7190befc554ed5d4d62dd5cf939fbae9181900360200190a15060015b6102a1565b604080516020808252600e908201527fe4bda0e4b88de698af6f776e65720000000000000000000000000000000000008183015290517f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa9181900360600190a15b919050565b600080546001600160a01b031633141561024057600080546001600160a01b038481166001600160a01b0319831681179093556040805191909216808252602082019390935281517f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90929181900390910190a160019150506102a1565b6060600180548060200260200160405190810160405280929190818152602001828054801561037b57602002820191906000526020600020905b81546001600160a01b0316815260019091019060200180831161035d575b5050505050905090565b6000546001600160a01b031690565b6000805b6001548110156103e657826001600160a01b0316600182815481106103b957fe5b6000918252602090912001546001600160a01b031614156103de5760019150506102a1565b600101610398565b5060009291505056fea165627a7a723058209d9341ff2d768aeaf4eb29bb08bf056e9900807aea51b345a69923a22225440e0029`

// DeployMapper deploys a new Ethereum contract, binding an instance of Mapper to it.
func DeployMapper(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Mapper, error) {
	parsed, err := abi.JSON(strings.NewReader(MapperABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(MapperBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Mapper{MapperCaller: MapperCaller{contract: contract}, MapperTransactor: MapperTransactor{contract: contract}, MapperFilterer: MapperFilterer{contract: contract}}, nil
}

// Mapper is an auto generated Go binding around an Ethereum contract.
type Mapper struct {
	MapperCaller     // Read-only binding to the contract
	MapperTransactor // Write-only binding to the contract
	MapperFilterer   // Log filterer for contract events
}

// MapperCaller is an auto generated read-only Go binding around an Ethereum contract.
type MapperCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MapperTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MapperTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MapperFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MapperFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MapperSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MapperSession struct {
	Contract     *Mapper           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MapperCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MapperCallerSession struct {
	Contract *MapperCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// MapperTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MapperTransactorSession struct {
	Contract     *MapperTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MapperRaw is an auto generated low-level Go binding around an Ethereum contract.
type MapperRaw struct {
	Contract *Mapper // Generic contract binding to access the raw methods on
}

// MapperCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MapperCallerRaw struct {
	Contract *MapperCaller // Generic read-only contract binding to access the raw methods on
}

// MapperTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MapperTransactorRaw struct {
	Contract *MapperTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMapper creates a new instance of Mapper, bound to a specific deployed contract.
func NewMapper(address common.Address, backend bind.ContractBackend) (*Mapper, error) {
	contract, err := bindMapper(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Mapper{MapperCaller: MapperCaller{contract: contract}, MapperTransactor: MapperTransactor{contract: contract}, MapperFilterer: MapperFilterer{contract: contract}}, nil
}

// NewMapperCaller creates a new read-only instance of Mapper, bound to a specific deployed contract.
func NewMapperCaller(address common.Address, caller bind.ContractCaller) (*MapperCaller, error) {
	contract, err := bindMapper(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MapperCaller{contract: contract}, nil
}

// NewMapperTransactor creates a new write-only instance of Mapper, bound to a specific deployed contract.
func NewMapperTransactor(address common.Address, transactor bind.ContractTransactor) (*MapperTransactor, error) {
	contract, err := bindMapper(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MapperTransactor{contract: contract}, nil
}

// NewMapperFilterer creates a new log filterer instance of Mapper, bound to a specific deployed contract.
func NewMapperFilterer(address common.Address, filterer bind.ContractFilterer) (*MapperFilterer, error) {
	contract, err := bindMapper(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MapperFilterer{contract: contract}, nil
}

// bindMapper binds a generic wrapper to an already deployed contract.
func bindMapper(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MapperABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Mapper *MapperRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Mapper.Contract.MapperCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Mapper *MapperRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Mapper.Contract.MapperTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Mapper *MapperRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Mapper.Contract.MapperTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Mapper *MapperCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Mapper.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Mapper *MapperTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Mapper.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Mapper *MapperTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Mapper.Contract.contract.Transact(opts, method, params...)
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() constant returns(address[])
func (_Mapper *MapperCaller) Get(opts *bind.CallOpts) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _Mapper.contract.Call(opts, out, "get")
	return *ret0, err
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() constant returns(address[])
func (_Mapper *MapperSession) Get() ([]common.Address, error) {
	return _Mapper.Contract.Get(&_Mapper.CallOpts)
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() constant returns(address[])
func (_Mapper *MapperCallerSession) Get() ([]common.Address, error) {
	return _Mapper.Contract.Get(&_Mapper.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Mapper *MapperCaller) GetOwner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Mapper.contract.Call(opts, out, "getOwner")
	return *ret0, err
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Mapper *MapperSession) GetOwner() (common.Address, error) {
	return _Mapper.Contract.GetOwner(&_Mapper.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Mapper *MapperCallerSession) GetOwner() (common.Address, error) {
	return _Mapper.Contract.GetOwner(&_Mapper.CallOpts)
}

// Add is a paid mutator transaction binding the contract method 0x0a3b0a4f.
//
// Solidity: function add(address addr) returns(bool)
func (_Mapper *MapperTransactor) Add(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _Mapper.contract.Transact(opts, "add", addr)
}

// Add is a paid mutator transaction binding the contract method 0x0a3b0a4f.
//
// Solidity: function add(address addr) returns(bool)
func (_Mapper *MapperSession) Add(addr common.Address) (*types.Transaction, error) {
	return _Mapper.Contract.Add(&_Mapper.TransactOpts, addr)
}

// Add is a paid mutator transaction binding the contract method 0x0a3b0a4f.
//
// Solidity: function add(address addr) returns(bool)
func (_Mapper *MapperTransactorSession) Add(addr common.Address) (*types.Transaction, error) {
	return _Mapper.Contract.Add(&_Mapper.TransactOpts, addr)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Mapper *MapperTransactor) AlterOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Mapper.contract.Transact(opts, "alterOwner", newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Mapper *MapperSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Mapper.Contract.AlterOwner(&_Mapper.TransactOpts, newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Mapper *MapperTransactorSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Mapper.Contract.AlterOwner(&_Mapper.TransactOpts, newOwner)
}

// MapperAddIterator is returned from FilterAdd and is used to iterate over the raw logs and unpacked data for Add events raised by the Mapper contract.
type MapperAddIterator struct {
	Event *MapperAdd // Event containing the contract specifics and raw log

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
func (it *MapperAddIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MapperAdd)
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
		it.Event = new(MapperAdd)
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
func (it *MapperAddIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MapperAddIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MapperAdd represents a Add event raised by the Mapper contract.
type MapperAdd struct {
	Addr common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAdd is a free log retrieval operation binding the contract event 0x87dc5eecd6d6bdeae407c426da6bfba5b7190befc554ed5d4d62dd5cf939fbae.
//
// Solidity: event Add(address addr)
func (_Mapper *MapperFilterer) FilterAdd(opts *bind.FilterOpts) (*MapperAddIterator, error) {

	logs, sub, err := _Mapper.contract.FilterLogs(opts, "Add")
	if err != nil {
		return nil, err
	}
	return &MapperAddIterator{contract: _Mapper.contract, event: "Add", logs: logs, sub: sub}, nil
}

// WatchAdd is a free log subscription operation binding the contract event 0x87dc5eecd6d6bdeae407c426da6bfba5b7190befc554ed5d4d62dd5cf939fbae.
//
// Solidity: event Add(address addr)
func (_Mapper *MapperFilterer) WatchAdd(opts *bind.WatchOpts, sink chan<- *MapperAdd) (event.Subscription, error) {

	logs, sub, err := _Mapper.contract.WatchLogs(opts, "Add")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MapperAdd)
				if err := _Mapper.contract.UnpackLog(event, "Add", log); err != nil {
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

// MapperAlterOwnerIterator is returned from FilterAlterOwner and is used to iterate over the raw logs and unpacked data for AlterOwner events raised by the Mapper contract.
type MapperAlterOwnerIterator struct {
	Event *MapperAlterOwner // Event containing the contract specifics and raw log

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
func (it *MapperAlterOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MapperAlterOwner)
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
		it.Event = new(MapperAlterOwner)
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
func (it *MapperAlterOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MapperAlterOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MapperAlterOwner represents a AlterOwner event raised by the Mapper contract.
type MapperAlterOwner struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAlterOwner is a free log retrieval operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Mapper *MapperFilterer) FilterAlterOwner(opts *bind.FilterOpts) (*MapperAlterOwnerIterator, error) {

	logs, sub, err := _Mapper.contract.FilterLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return &MapperAlterOwnerIterator{contract: _Mapper.contract, event: "AlterOwner", logs: logs, sub: sub}, nil
}

// WatchAlterOwner is a free log subscription operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Mapper *MapperFilterer) WatchAlterOwner(opts *bind.WatchOpts, sink chan<- *MapperAlterOwner) (event.Subscription, error) {

	logs, sub, err := _Mapper.contract.WatchLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MapperAlterOwner)
				if err := _Mapper.contract.UnpackLog(event, "AlterOwner", log); err != nil {
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

// MapperErrorIterator is returned from FilterError and is used to iterate over the raw logs and unpacked data for Error events raised by the Mapper contract.
type MapperErrorIterator struct {
	Event *MapperError // Event containing the contract specifics and raw log

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
func (it *MapperErrorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MapperError)
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
		it.Event = new(MapperError)
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
func (it *MapperErrorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MapperErrorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MapperError represents a Error event raised by the Mapper contract.
type MapperError struct {
	Data string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterError is a free log retrieval operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Mapper *MapperFilterer) FilterError(opts *bind.FilterOpts) (*MapperErrorIterator, error) {

	logs, sub, err := _Mapper.contract.FilterLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return &MapperErrorIterator{contract: _Mapper.contract, event: "Error", logs: logs, sub: sub}, nil
}

// WatchError is a free log subscription operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Mapper *MapperFilterer) WatchError(opts *bind.WatchOpts, sink chan<- *MapperError) (event.Subscription, error) {

	logs, sub, err := _Mapper.contract.WatchLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MapperError)
				if err := _Mapper.contract.UnpackLog(event, "Error", log); err != nil {
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
