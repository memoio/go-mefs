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
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// MapperABI is the input ABI used to generate the binding from.
const MapperABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"Add\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"data\",\"type\":\"string\"}],\"name\":\"Error\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"add\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"get\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// MapperBin is the compiled bytecode used for deploying new contracts.
var MapperBin = "0x6080604052738026796fd7ce63eae824314aa5bacf55643e893d600260006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555034801561006557600080fd5b50336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506107ea806100b56000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c80630a3b0a4f146100515780630ca05f9f146100ab5780636d4ce63c14610105578063893d20e814610164575b600080fd5b6100936004803603602081101561006757600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610198565b60405180821515815260200191505060405180910390f35b6100ed600480360360208110156100c157600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506104bc565b60405180821515815260200191505060405180910390f35b61010d61065b565b6040518080602001828103825283818151815260200191508051906020019060200280838360005b83811015610150578082015181840152602081019050610135565b505050509050019250505060405180910390f35b61016c6106e9565b604051808273ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461025c576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b6000600260009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166333c767726040518163ffffffff1660e01b815260040160206040518083038186803b1580156102c657600080fd5b505afa1580156102da573d6000803e3d6000fd5b505050506040513d60208110156102f057600080fd5b81019080805190602001909291905050509050600161ffff168161ffff1610610381576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252600d8152602001807f6164642069732062616e6e65640000000000000000000000000000000000000081525060200191505060405180910390fd5b61038a83610712565b15610401577f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa60405180806020018281038252600b8152602001807f68617320616c726561647900000000000000000000000000000000000000000081525060200191505060405180910390a160009150506104b7565b6001839080600181540180825580915050600190039060005260206000200160009091909190916101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f87dc5eecd6d6bdeae407c426da6bfba5b7190befc554ed5d4d62dd5cf939fbae83604051808273ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390a160019150505b919050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610580576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050826000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e908184604051808373ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a16001915050919050565b606060018054806020026020016040519081016040528092919081815260200182805480156106df57602002820191906000526020600020905b8160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019060010190808311610695575b5050505050905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b600080600090505b6001805490508110156107a9578273ffffffffffffffffffffffffffffffffffffffff166001828154811061074b57fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16141561079c5760019150506107af565b808060010191505061071a565b50600090505b91905056fea2646970667358221220cc9d21e0bbf7a713e7bd34c955dd3f934d7eb81b6040f7866d0c588ad7c25d1b64736f6c63430007030033"

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
// Solidity: function get() view returns(address[])
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
// Solidity: function get() view returns(address[])
func (_Mapper *MapperSession) Get() ([]common.Address, error) {
	return _Mapper.Contract.Get(&_Mapper.CallOpts)
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() view returns(address[])
func (_Mapper *MapperCallerSession) Get() ([]common.Address, error) {
	return _Mapper.Contract.Get(&_Mapper.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
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
// Solidity: function getOwner() view returns(address)
func (_Mapper *MapperSession) GetOwner() (common.Address, error) {
	return _Mapper.Contract.GetOwner(&_Mapper.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
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

// ParseAdd is a log parse operation binding the contract event 0x87dc5eecd6d6bdeae407c426da6bfba5b7190befc554ed5d4d62dd5cf939fbae.
//
// Solidity: event Add(address addr)
func (_Mapper *MapperFilterer) ParseAdd(log types.Log) (*MapperAdd, error) {
	event := new(MapperAdd)
	if err := _Mapper.contract.UnpackLog(event, "Add", log); err != nil {
		return nil, err
	}
	return event, nil
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

// ParseAlterOwner is a log parse operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Mapper *MapperFilterer) ParseAlterOwner(log types.Log) (*MapperAlterOwner, error) {
	event := new(MapperAlterOwner)
	if err := _Mapper.contract.UnpackLog(event, "AlterOwner", log); err != nil {
		return nil, err
	}
	return event, nil
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

// ParseError is a log parse operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Mapper *MapperFilterer) ParseError(log types.Log) (*MapperError, error) {
	event := new(MapperError)
	if err := _Mapper.contract.UnpackLog(event, "Error", log); err != nil {
		return nil, err
	}
	return event, nil
}
