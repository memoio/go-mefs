// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package upKeeping

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

// DebugABI is the input ABI used to generate the binding from.
const DebugABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"data\",\"type\":\"string\"}],\"name\":\"Error\",\"type\":\"event\"}]"

// DebugBin is the compiled bytecode used for deploying new contracts.
var DebugBin = "0x6080604052348015600f57600080fd5b50603e80601d6000396000f3fe6080604052600080fdfea265627a7a723058206b2be8d2a80c38af3527dd3354921be3b835ebdb0a14a328af7bd2db0a513b1064736f6c63430005090032"

// DeployDebug deploys a new Ethereum contract, binding an instance of Debug to it.
func DeployDebug(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Debug, error) {
	parsed, err := abi.JSON(strings.NewReader(DebugABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(DebugBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Debug{DebugCaller: DebugCaller{contract: contract}, DebugTransactor: DebugTransactor{contract: contract}, DebugFilterer: DebugFilterer{contract: contract}}, nil
}

// Debug is an auto generated Go binding around an Ethereum contract.
type Debug struct {
	DebugCaller     // Read-only binding to the contract
	DebugTransactor // Write-only binding to the contract
	DebugFilterer   // Log filterer for contract events
}

// DebugCaller is an auto generated read-only Go binding around an Ethereum contract.
type DebugCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DebugTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DebugTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DebugFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DebugFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DebugSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DebugSession struct {
	Contract     *Debug            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DebugCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DebugCallerSession struct {
	Contract *DebugCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// DebugTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DebugTransactorSession struct {
	Contract     *DebugTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DebugRaw is an auto generated low-level Go binding around an Ethereum contract.
type DebugRaw struct {
	Contract *Debug // Generic contract binding to access the raw methods on
}

// DebugCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DebugCallerRaw struct {
	Contract *DebugCaller // Generic read-only contract binding to access the raw methods on
}

// DebugTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DebugTransactorRaw struct {
	Contract *DebugTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDebug creates a new instance of Debug, bound to a specific deployed contract.
func NewDebug(address common.Address, backend bind.ContractBackend) (*Debug, error) {
	contract, err := bindDebug(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Debug{DebugCaller: DebugCaller{contract: contract}, DebugTransactor: DebugTransactor{contract: contract}, DebugFilterer: DebugFilterer{contract: contract}}, nil
}

// NewDebugCaller creates a new read-only instance of Debug, bound to a specific deployed contract.
func NewDebugCaller(address common.Address, caller bind.ContractCaller) (*DebugCaller, error) {
	contract, err := bindDebug(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DebugCaller{contract: contract}, nil
}

// NewDebugTransactor creates a new write-only instance of Debug, bound to a specific deployed contract.
func NewDebugTransactor(address common.Address, transactor bind.ContractTransactor) (*DebugTransactor, error) {
	contract, err := bindDebug(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DebugTransactor{contract: contract}, nil
}

// NewDebugFilterer creates a new log filterer instance of Debug, bound to a specific deployed contract.
func NewDebugFilterer(address common.Address, filterer bind.ContractFilterer) (*DebugFilterer, error) {
	contract, err := bindDebug(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DebugFilterer{contract: contract}, nil
}

// bindDebug binds a generic wrapper to an already deployed contract.
func bindDebug(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(DebugABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Debug *DebugRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Debug.Contract.DebugCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Debug *DebugRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Debug.Contract.DebugTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Debug *DebugRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Debug.Contract.DebugTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Debug *DebugCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Debug.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Debug *DebugTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Debug.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Debug *DebugTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Debug.Contract.contract.Transact(opts, method, params...)
}

// DebugErrorIterator is returned from FilterError and is used to iterate over the raw logs and unpacked data for Error events raised by the Debug contract.
type DebugErrorIterator struct {
	Event *DebugError // Event containing the contract specifics and raw log

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
func (it *DebugErrorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DebugError)
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
		it.Event = new(DebugError)
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
func (it *DebugErrorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DebugErrorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DebugError represents a Error event raised by the Debug contract.
type DebugError struct {
	Data string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterError is a free log retrieval operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Debug *DebugFilterer) FilterError(opts *bind.FilterOpts) (*DebugErrorIterator, error) {

	logs, sub, err := _Debug.contract.FilterLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return &DebugErrorIterator{contract: _Debug.contract, event: "Error", logs: logs, sub: sub}, nil
}

// WatchError is a free log subscription operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Debug *DebugFilterer) WatchError(opts *bind.WatchOpts, sink chan<- *DebugError) (event.Subscription, error) {

	logs, sub, err := _Debug.contract.WatchLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DebugError)
				if err := _Debug.contract.UnpackLog(event, "Error", log); err != nil {
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
func (_Debug *DebugFilterer) ParseError(log types.Log) (*DebugError, error) {
	event := new(DebugError)
	if err := _Debug.contract.UnpackLog(event, "Error", log); err != nil {
		return nil, err
	}
	return event, nil
}

// OwnedABI is the input ABI used to generate the binding from.
const OwnedABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"data\",\"type\":\"string\"}],\"name\":\"Error\",\"type\":\"event\"}]"

// OwnedFuncSigs maps the 4-byte function signature to its string representation.
var OwnedFuncSigs = map[string]string{
	"0ca05f9f": "alterOwner(address)",
	"893d20e8": "getOwner()",
}

// OwnedBin is the compiled bytecode used for deploying new contracts.
var OwnedBin = "0x608060405234801561001057600080fd5b50600080546001600160a01b031916331790556101b1806100326000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c80630ca05f9f1461003b578063893d20e814610075575b600080fd5b6100616004803603602081101561005157600080fd5b50356001600160a01b0316610099565b604080519115158252519081900360200190f35b61007d61016d565b604080516001600160a01b039092168252519081900360200190f35b600080546001600160a01b031633141561011657600080546001600160a01b038481166001600160a01b0319831681179093556040805191909216808252602082019390935281517f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90929181900390910190a16001915050610168565b604080516020808252600e908201526d725ed0725c46f34c57b7bbb732b960911b8183015290517f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa9181900360600190a15b919050565b6000546001600160a01b03169056fea265627a7a723058209d5d2d0af266adf900dcd69280cd6c67137fd502f4fefd6f6b90f8836e69012c64736f6c63430005090032"

// DeployOwned deploys a new Ethereum contract, binding an instance of Owned to it.
func DeployOwned(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Owned, error) {
	parsed, err := abi.JSON(strings.NewReader(OwnedABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(OwnedBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Owned{OwnedCaller: OwnedCaller{contract: contract}, OwnedTransactor: OwnedTransactor{contract: contract}, OwnedFilterer: OwnedFilterer{contract: contract}}, nil
}

// Owned is an auto generated Go binding around an Ethereum contract.
type Owned struct {
	OwnedCaller     // Read-only binding to the contract
	OwnedTransactor // Write-only binding to the contract
	OwnedFilterer   // Log filterer for contract events
}

// OwnedCaller is an auto generated read-only Go binding around an Ethereum contract.
type OwnedCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OwnedTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OwnedTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OwnedFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OwnedFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OwnedSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OwnedSession struct {
	Contract     *Owned            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OwnedCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OwnedCallerSession struct {
	Contract *OwnedCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// OwnedTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OwnedTransactorSession struct {
	Contract     *OwnedTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OwnedRaw is an auto generated low-level Go binding around an Ethereum contract.
type OwnedRaw struct {
	Contract *Owned // Generic contract binding to access the raw methods on
}

// OwnedCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OwnedCallerRaw struct {
	Contract *OwnedCaller // Generic read-only contract binding to access the raw methods on
}

// OwnedTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OwnedTransactorRaw struct {
	Contract *OwnedTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOwned creates a new instance of Owned, bound to a specific deployed contract.
func NewOwned(address common.Address, backend bind.ContractBackend) (*Owned, error) {
	contract, err := bindOwned(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Owned{OwnedCaller: OwnedCaller{contract: contract}, OwnedTransactor: OwnedTransactor{contract: contract}, OwnedFilterer: OwnedFilterer{contract: contract}}, nil
}

// NewOwnedCaller creates a new read-only instance of Owned, bound to a specific deployed contract.
func NewOwnedCaller(address common.Address, caller bind.ContractCaller) (*OwnedCaller, error) {
	contract, err := bindOwned(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OwnedCaller{contract: contract}, nil
}

// NewOwnedTransactor creates a new write-only instance of Owned, bound to a specific deployed contract.
func NewOwnedTransactor(address common.Address, transactor bind.ContractTransactor) (*OwnedTransactor, error) {
	contract, err := bindOwned(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OwnedTransactor{contract: contract}, nil
}

// NewOwnedFilterer creates a new log filterer instance of Owned, bound to a specific deployed contract.
func NewOwnedFilterer(address common.Address, filterer bind.ContractFilterer) (*OwnedFilterer, error) {
	contract, err := bindOwned(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OwnedFilterer{contract: contract}, nil
}

// bindOwned binds a generic wrapper to an already deployed contract.
func bindOwned(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(OwnedABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Owned *OwnedRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Owned.Contract.OwnedCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Owned *OwnedRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Owned.Contract.OwnedTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Owned *OwnedRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Owned.Contract.OwnedTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Owned *OwnedCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Owned.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Owned *OwnedTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Owned.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Owned *OwnedTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Owned.Contract.contract.Transact(opts, method, params...)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Owned *OwnedCaller) GetOwner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Owned.contract.Call(opts, out, "getOwner")
	return *ret0, err
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Owned *OwnedSession) GetOwner() (common.Address, error) {
	return _Owned.Contract.GetOwner(&_Owned.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Owned *OwnedCallerSession) GetOwner() (common.Address, error) {
	return _Owned.Contract.GetOwner(&_Owned.CallOpts)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Owned *OwnedTransactor) AlterOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Owned.contract.Transact(opts, "alterOwner", newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Owned *OwnedSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Owned.Contract.AlterOwner(&_Owned.TransactOpts, newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Owned *OwnedTransactorSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Owned.Contract.AlterOwner(&_Owned.TransactOpts, newOwner)
}

// OwnedAlterOwnerIterator is returned from FilterAlterOwner and is used to iterate over the raw logs and unpacked data for AlterOwner events raised by the Owned contract.
type OwnedAlterOwnerIterator struct {
	Event *OwnedAlterOwner // Event containing the contract specifics and raw log

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
func (it *OwnedAlterOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OwnedAlterOwner)
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
		it.Event = new(OwnedAlterOwner)
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
func (it *OwnedAlterOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OwnedAlterOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OwnedAlterOwner represents a AlterOwner event raised by the Owned contract.
type OwnedAlterOwner struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAlterOwner is a free log retrieval operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Owned *OwnedFilterer) FilterAlterOwner(opts *bind.FilterOpts) (*OwnedAlterOwnerIterator, error) {

	logs, sub, err := _Owned.contract.FilterLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return &OwnedAlterOwnerIterator{contract: _Owned.contract, event: "AlterOwner", logs: logs, sub: sub}, nil
}

// WatchAlterOwner is a free log subscription operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Owned *OwnedFilterer) WatchAlterOwner(opts *bind.WatchOpts, sink chan<- *OwnedAlterOwner) (event.Subscription, error) {

	logs, sub, err := _Owned.contract.WatchLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OwnedAlterOwner)
				if err := _Owned.contract.UnpackLog(event, "AlterOwner", log); err != nil {
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
func (_Owned *OwnedFilterer) ParseAlterOwner(log types.Log) (*OwnedAlterOwner, error) {
	event := new(OwnedAlterOwner)
	if err := _Owned.contract.UnpackLog(event, "AlterOwner", log); err != nil {
		return nil, err
	}
	return event, nil
}

// OwnedErrorIterator is returned from FilterError and is used to iterate over the raw logs and unpacked data for Error events raised by the Owned contract.
type OwnedErrorIterator struct {
	Event *OwnedError // Event containing the contract specifics and raw log

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
func (it *OwnedErrorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OwnedError)
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
		it.Event = new(OwnedError)
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
func (it *OwnedErrorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OwnedErrorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OwnedError represents a Error event raised by the Owned contract.
type OwnedError struct {
	Data string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterError is a free log retrieval operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Owned *OwnedFilterer) FilterError(opts *bind.FilterOpts) (*OwnedErrorIterator, error) {

	logs, sub, err := _Owned.contract.FilterLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return &OwnedErrorIterator{contract: _Owned.contract, event: "Error", logs: logs, sub: sub}, nil
}

// WatchError is a free log subscription operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Owned *OwnedFilterer) WatchError(opts *bind.WatchOpts, sink chan<- *OwnedError) (event.Subscription, error) {

	logs, sub, err := _Owned.contract.WatchLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OwnedError)
				if err := _Owned.contract.UnpackLog(event, "Error", log); err != nil {
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
func (_Owned *OwnedFilterer) ParseError(log types.Log) (*OwnedError, error) {
	event := new(OwnedError)
	if err := _Owned.contract.UnpackLog(event, "Error", log); err != nil {
		return nil, err
	}
	return event, nil
}

// UpKeepingABI is the input ABI used to generate the binding from.
const UpKeepingABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"provider\",\"type\":\"address\"},{\"name\":\"money\",\"type\":\"uint256\"}],\"name\":\"spaceTimePay\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOrder\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"address[]\"},{\"name\":\"\",\"type\":\"address[]\"},{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"user\",\"type\":\"address\"},{\"name\":\"keeper\",\"type\":\"address[]\"},{\"name\":\"provider\",\"type\":\"address[]\"},{\"name\":\"time\",\"type\":\"uint256\"},{\"name\":\"size\",\"type\":\"uint256\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"AddOrder\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"PayKeeper\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"PayProvider\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"data\",\"type\":\"string\"}],\"name\":\"Error\",\"type\":\"event\"}]"

// UpKeepingFuncSigs maps the 4-byte function signature to its string representation.
var UpKeepingFuncSigs = map[string]string{
	"0ca05f9f": "alterOwner(address)",
	"d36dedd2": "getOrder()",
	"893d20e8": "getOwner()",
	"1e042234": "spaceTimePay(address,uint256)",
}

// UpKeepingBin is the compiled bytecode used for deploying new contracts.
var UpKeepingBin = "0x6080604052604051610933380380610933833981810160405260a081101561002657600080fd5b81516020830180519193928301929164010000000081111561004757600080fd5b8201602081018481111561005a57600080fd5b815185602082028301116401000000008211171561007757600080fd5b5050929190602001805164010000000081111561009357600080fd5b820160208101848111156100a657600080fd5b81518560208202830111640100000000821117156100c357600080fd5b5050602080830151604093840151600080546001600160a01b03199081163317909155855160c0810187526001600160a01b038b168082528186018b905296810186905260608101849052608081018390523460a0820152600180549092169096178155885194975091955093929091610142916002918901906101b2565b506040820151805161015e9160028401916020909101906101b2565b50606082015160038201556080820151600482015560a0909101516005909101556040517f0905316f7faca135c292b6e6f8d91c19128d372722215fe029e74e75ef84c08790600090a1505050505061023e565b828054828255906000526020600020908101928215610207579160200282015b8281111561020757825182546001600160a01b0319166001600160a01b039091161782556020909201916001909101906101d2565b50610213929150610217565b5090565b61023b91905b808211156102135780546001600160a01b031916815560010161021d565b90565b6106e68061024d6000396000f3fe60806040526004361061003f5760003560e01c80630ca05f9f146100415780631e04223414610088578063893d20e8146100c1578063d36dedd2146100f2575b005b34801561004d57600080fd5b506100746004803603602081101561006457600080fd5b50356001600160a01b03166101ce565b604080519115158252519081900360200190f35b34801561009457600080fd5b50610074600480360360408110156100ab57600080fd5b506001600160a01b038135169060200135610290565b3480156100cd57600080fd5b506100d661053e565b604080516001600160a01b039092168252519081900360200190f35b3480156100fe57600080fd5b5061010761054d565b60405180876001600160a01b03166001600160a01b031681526020018060200180602001868152602001858152602001848152602001838103835288818151815260200191508051906020019060200280838360005b8381101561017557818101518382015260200161015d565b50505050905001838103825287818151815260200191508051906020019060200280838360005b838110156101b457818101518382015260200161019c565b505050509050019850505050505050505060405180910390f35b600080546001600160a01b031633141561024b57600080546001600160a01b038481166001600160a01b0319831681179093556040805191909216808252602082019390935281517f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90929181900390910190a1600191505061028b565b604080516020808252600e908201526d725ed0725c46f34c57b7bbb732b960911b8183015290516000805160206106928339815191529181900360600190a15b919050565b600080805b6002548110156102d65760028054829081106102ad57fe5b6000918252602090912001546001600160a01b03163314156102ce57600191505b600101610295565b5080156104f057303183111561033d576040805160208082526018908201527fe59088e7baa6e4b8ade79a84e4bd99e9a29de4b88de8b6b300000000000000008183015290516000805160206106928339815191529181900360600190a1600091506104eb565b61034684610644565b61039f57604080516020808252808201527fe8a681e8bdace8b4a6e79a84e59cb0e59d80e4b88de698af70726f76696465728183015290516000805160206106928339815191529181900360600190a1600091506104eb565b604051600a8404906001600160a01b038616906009830280156108fc02916000818181858888f193505050501580156103dc573d6000803e3d6000fd5b5060405160098202906001600160a01b0387169033907f1569130f5bdbde161a213db1c477e4f2670f09e2a9c1c08ca9bafe749b80cb4190600090a460025460005b818110156104e357600280548290811061043457fe5b6000918252602090912001546001600160a01b03166108fc83858161045557fe5b049081150290604051600060405180830381858888f19350505050158015610481573d6000803e3d6000fd5b5081838161048b57fe5b0460018001828154811061049b57fe5b60009182526020822001546040516001600160a01b039091169133917faa4c66f6ddfadc835acfabab55148a78bc3e6867ed1cdb36461a10685af4c0c39190a460010161041e565b506001935050505b610536565b604080516020808252600f908201526e725ed0725c46f34c57b5b2b2b832b960891b8183015290516000805160206106928339815191529181900360600190a150610538565b505b92915050565b6000546001600160a01b031690565b600154600454600554600654600280546040805160208084028201810190925282815260009760609788978a97889788976001600160a01b0390951696909560039590918791908301828280156105cd57602002820191906000526020600020905b81546001600160a01b031681526001909101906020018083116105af575b505050505094508380548060200260200160405190810160405280929190818152602001828054801561062957602002820191906000526020600020905b81546001600160a01b0316815260019091019060200180831161060b575b50505050509350955095509550955095509550909192939495565b600080805b60035481101561053657600380548290811061066157fe5b6000918252602090912001546001600160a01b03858116911614156106895760019150610536565b60010161064956fe08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aaa265627a7a723058206609ba43bfd077104068a9fb22f8c06708723c7c2f28fd1e12dafb11a80386ac64736f6c63430005090032"

// DeployUpKeeping deploys a new Ethereum contract, binding an instance of UpKeeping to it.
func DeployUpKeeping(auth *bind.TransactOpts, backend bind.ContractBackend, user common.Address, keeper []common.Address, provider []common.Address, time *big.Int, size *big.Int) (common.Address, *types.Transaction, *UpKeeping, error) {
	parsed, err := abi.JSON(strings.NewReader(UpKeepingABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(UpKeepingBin), backend, user, keeper, provider, time, size)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &UpKeeping{UpKeepingCaller: UpKeepingCaller{contract: contract}, UpKeepingTransactor: UpKeepingTransactor{contract: contract}, UpKeepingFilterer: UpKeepingFilterer{contract: contract}}, nil
}

// UpKeeping is an auto generated Go binding around an Ethereum contract.
type UpKeeping struct {
	UpKeepingCaller     // Read-only binding to the contract
	UpKeepingTransactor // Write-only binding to the contract
	UpKeepingFilterer   // Log filterer for contract events
}

// UpKeepingCaller is an auto generated read-only Go binding around an Ethereum contract.
type UpKeepingCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UpKeepingTransactor is an auto generated write-only Go binding around an Ethereum contract.
type UpKeepingTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UpKeepingFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type UpKeepingFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UpKeepingSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type UpKeepingSession struct {
	Contract     *UpKeeping        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// UpKeepingCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type UpKeepingCallerSession struct {
	Contract *UpKeepingCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// UpKeepingTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type UpKeepingTransactorSession struct {
	Contract     *UpKeepingTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// UpKeepingRaw is an auto generated low-level Go binding around an Ethereum contract.
type UpKeepingRaw struct {
	Contract *UpKeeping // Generic contract binding to access the raw methods on
}

// UpKeepingCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type UpKeepingCallerRaw struct {
	Contract *UpKeepingCaller // Generic read-only contract binding to access the raw methods on
}

// UpKeepingTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type UpKeepingTransactorRaw struct {
	Contract *UpKeepingTransactor // Generic write-only contract binding to access the raw methods on
}

// NewUpKeeping creates a new instance of UpKeeping, bound to a specific deployed contract.
func NewUpKeeping(address common.Address, backend bind.ContractBackend) (*UpKeeping, error) {
	contract, err := bindUpKeeping(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &UpKeeping{UpKeepingCaller: UpKeepingCaller{contract: contract}, UpKeepingTransactor: UpKeepingTransactor{contract: contract}, UpKeepingFilterer: UpKeepingFilterer{contract: contract}}, nil
}

// NewUpKeepingCaller creates a new read-only instance of UpKeeping, bound to a specific deployed contract.
func NewUpKeepingCaller(address common.Address, caller bind.ContractCaller) (*UpKeepingCaller, error) {
	contract, err := bindUpKeeping(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &UpKeepingCaller{contract: contract}, nil
}

// NewUpKeepingTransactor creates a new write-only instance of UpKeeping, bound to a specific deployed contract.
func NewUpKeepingTransactor(address common.Address, transactor bind.ContractTransactor) (*UpKeepingTransactor, error) {
	contract, err := bindUpKeeping(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &UpKeepingTransactor{contract: contract}, nil
}

// NewUpKeepingFilterer creates a new log filterer instance of UpKeeping, bound to a specific deployed contract.
func NewUpKeepingFilterer(address common.Address, filterer bind.ContractFilterer) (*UpKeepingFilterer, error) {
	contract, err := bindUpKeeping(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &UpKeepingFilterer{contract: contract}, nil
}

// bindUpKeeping binds a generic wrapper to an already deployed contract.
func bindUpKeeping(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(UpKeepingABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_UpKeeping *UpKeepingRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _UpKeeping.Contract.UpKeepingCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_UpKeeping *UpKeepingRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UpKeeping.Contract.UpKeepingTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_UpKeeping *UpKeepingRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _UpKeeping.Contract.UpKeepingTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_UpKeeping *UpKeepingCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _UpKeeping.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_UpKeeping *UpKeepingTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UpKeeping.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_UpKeeping *UpKeepingTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _UpKeeping.Contract.contract.Transact(opts, method, params...)
}

// GetOrder is a free data retrieval call binding the contract method 0xd36dedd2.
//
// Solidity: function getOrder() constant returns(address, address[], address[], uint256, uint256, uint256)
func (_UpKeeping *UpKeepingCaller) GetOrder(opts *bind.CallOpts) (common.Address, []common.Address, []common.Address, *big.Int, *big.Int, *big.Int, error) {
	var (
		ret0 = new(common.Address)
		ret1 = new([]common.Address)
		ret2 = new([]common.Address)
		ret3 = new(*big.Int)
		ret4 = new(*big.Int)
		ret5 = new(*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
		ret3,
		ret4,
		ret5,
	}
	err := _UpKeeping.contract.Call(opts, out, "getOrder")
	return *ret0, *ret1, *ret2, *ret3, *ret4, *ret5, err
}

// GetOrder is a free data retrieval call binding the contract method 0xd36dedd2.
//
// Solidity: function getOrder() constant returns(address, address[], address[], uint256, uint256, uint256)
func (_UpKeeping *UpKeepingSession) GetOrder() (common.Address, []common.Address, []common.Address, *big.Int, *big.Int, *big.Int, error) {
	return _UpKeeping.Contract.GetOrder(&_UpKeeping.CallOpts)
}

// GetOrder is a free data retrieval call binding the contract method 0xd36dedd2.
//
// Solidity: function getOrder() constant returns(address, address[], address[], uint256, uint256, uint256)
func (_UpKeeping *UpKeepingCallerSession) GetOrder() (common.Address, []common.Address, []common.Address, *big.Int, *big.Int, *big.Int, error) {
	return _UpKeeping.Contract.GetOrder(&_UpKeeping.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_UpKeeping *UpKeepingCaller) GetOwner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _UpKeeping.contract.Call(opts, out, "getOwner")
	return *ret0, err
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_UpKeeping *UpKeepingSession) GetOwner() (common.Address, error) {
	return _UpKeeping.Contract.GetOwner(&_UpKeeping.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_UpKeeping *UpKeepingCallerSession) GetOwner() (common.Address, error) {
	return _UpKeeping.Contract.GetOwner(&_UpKeeping.CallOpts)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_UpKeeping *UpKeepingTransactor) AlterOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _UpKeeping.contract.Transact(opts, "alterOwner", newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_UpKeeping *UpKeepingSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _UpKeeping.Contract.AlterOwner(&_UpKeeping.TransactOpts, newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_UpKeeping *UpKeepingTransactorSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _UpKeeping.Contract.AlterOwner(&_UpKeeping.TransactOpts, newOwner)
}

// SpaceTimePay is a paid mutator transaction binding the contract method 0x1e042234.
//
// Solidity: function spaceTimePay(address provider, uint256 money) returns(bool)
func (_UpKeeping *UpKeepingTransactor) SpaceTimePay(opts *bind.TransactOpts, provider common.Address, money *big.Int) (*types.Transaction, error) {
	return _UpKeeping.contract.Transact(opts, "spaceTimePay", provider, money)
}

// SpaceTimePay is a paid mutator transaction binding the contract method 0x1e042234.
//
// Solidity: function spaceTimePay(address provider, uint256 money) returns(bool)
func (_UpKeeping *UpKeepingSession) SpaceTimePay(provider common.Address, money *big.Int) (*types.Transaction, error) {
	return _UpKeeping.Contract.SpaceTimePay(&_UpKeeping.TransactOpts, provider, money)
}

// SpaceTimePay is a paid mutator transaction binding the contract method 0x1e042234.
//
// Solidity: function spaceTimePay(address provider, uint256 money) returns(bool)
func (_UpKeeping *UpKeepingTransactorSession) SpaceTimePay(provider common.Address, money *big.Int) (*types.Transaction, error) {
	return _UpKeeping.Contract.SpaceTimePay(&_UpKeeping.TransactOpts, provider, money)
}

// UpKeepingAddOrderIterator is returned from FilterAddOrder and is used to iterate over the raw logs and unpacked data for AddOrder events raised by the UpKeeping contract.
type UpKeepingAddOrderIterator struct {
	Event *UpKeepingAddOrder // Event containing the contract specifics and raw log

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
func (it *UpKeepingAddOrderIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UpKeepingAddOrder)
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
		it.Event = new(UpKeepingAddOrder)
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
func (it *UpKeepingAddOrderIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UpKeepingAddOrderIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UpKeepingAddOrder represents a AddOrder event raised by the UpKeeping contract.
type UpKeepingAddOrder struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterAddOrder is a free log retrieval operation binding the contract event 0x0905316f7faca135c292b6e6f8d91c19128d372722215fe029e74e75ef84c087.
//
// Solidity: event AddOrder()
func (_UpKeeping *UpKeepingFilterer) FilterAddOrder(opts *bind.FilterOpts) (*UpKeepingAddOrderIterator, error) {

	logs, sub, err := _UpKeeping.contract.FilterLogs(opts, "AddOrder")
	if err != nil {
		return nil, err
	}
	return &UpKeepingAddOrderIterator{contract: _UpKeeping.contract, event: "AddOrder", logs: logs, sub: sub}, nil
}

// WatchAddOrder is a free log subscription operation binding the contract event 0x0905316f7faca135c292b6e6f8d91c19128d372722215fe029e74e75ef84c087.
//
// Solidity: event AddOrder()
func (_UpKeeping *UpKeepingFilterer) WatchAddOrder(opts *bind.WatchOpts, sink chan<- *UpKeepingAddOrder) (event.Subscription, error) {

	logs, sub, err := _UpKeeping.contract.WatchLogs(opts, "AddOrder")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UpKeepingAddOrder)
				if err := _UpKeeping.contract.UnpackLog(event, "AddOrder", log); err != nil {
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

// ParseAddOrder is a log parse operation binding the contract event 0x0905316f7faca135c292b6e6f8d91c19128d372722215fe029e74e75ef84c087.
//
// Solidity: event AddOrder()
func (_UpKeeping *UpKeepingFilterer) ParseAddOrder(log types.Log) (*UpKeepingAddOrder, error) {
	event := new(UpKeepingAddOrder)
	if err := _UpKeeping.contract.UnpackLog(event, "AddOrder", log); err != nil {
		return nil, err
	}
	return event, nil
}

// UpKeepingAlterOwnerIterator is returned from FilterAlterOwner and is used to iterate over the raw logs and unpacked data for AlterOwner events raised by the UpKeeping contract.
type UpKeepingAlterOwnerIterator struct {
	Event *UpKeepingAlterOwner // Event containing the contract specifics and raw log

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
func (it *UpKeepingAlterOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UpKeepingAlterOwner)
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
		it.Event = new(UpKeepingAlterOwner)
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
func (it *UpKeepingAlterOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UpKeepingAlterOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UpKeepingAlterOwner represents a AlterOwner event raised by the UpKeeping contract.
type UpKeepingAlterOwner struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAlterOwner is a free log retrieval operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_UpKeeping *UpKeepingFilterer) FilterAlterOwner(opts *bind.FilterOpts) (*UpKeepingAlterOwnerIterator, error) {

	logs, sub, err := _UpKeeping.contract.FilterLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return &UpKeepingAlterOwnerIterator{contract: _UpKeeping.contract, event: "AlterOwner", logs: logs, sub: sub}, nil
}

// WatchAlterOwner is a free log subscription operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_UpKeeping *UpKeepingFilterer) WatchAlterOwner(opts *bind.WatchOpts, sink chan<- *UpKeepingAlterOwner) (event.Subscription, error) {

	logs, sub, err := _UpKeeping.contract.WatchLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UpKeepingAlterOwner)
				if err := _UpKeeping.contract.UnpackLog(event, "AlterOwner", log); err != nil {
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
func (_UpKeeping *UpKeepingFilterer) ParseAlterOwner(log types.Log) (*UpKeepingAlterOwner, error) {
	event := new(UpKeepingAlterOwner)
	if err := _UpKeeping.contract.UnpackLog(event, "AlterOwner", log); err != nil {
		return nil, err
	}
	return event, nil
}

// UpKeepingErrorIterator is returned from FilterError and is used to iterate over the raw logs and unpacked data for Error events raised by the UpKeeping contract.
type UpKeepingErrorIterator struct {
	Event *UpKeepingError // Event containing the contract specifics and raw log

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
func (it *UpKeepingErrorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UpKeepingError)
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
		it.Event = new(UpKeepingError)
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
func (it *UpKeepingErrorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UpKeepingErrorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UpKeepingError represents a Error event raised by the UpKeeping contract.
type UpKeepingError struct {
	Data string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterError is a free log retrieval operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_UpKeeping *UpKeepingFilterer) FilterError(opts *bind.FilterOpts) (*UpKeepingErrorIterator, error) {

	logs, sub, err := _UpKeeping.contract.FilterLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return &UpKeepingErrorIterator{contract: _UpKeeping.contract, event: "Error", logs: logs, sub: sub}, nil
}

// WatchError is a free log subscription operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_UpKeeping *UpKeepingFilterer) WatchError(opts *bind.WatchOpts, sink chan<- *UpKeepingError) (event.Subscription, error) {

	logs, sub, err := _UpKeeping.contract.WatchLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UpKeepingError)
				if err := _UpKeeping.contract.UnpackLog(event, "Error", log); err != nil {
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
func (_UpKeeping *UpKeepingFilterer) ParseError(log types.Log) (*UpKeepingError, error) {
	event := new(UpKeepingError)
	if err := _UpKeeping.contract.UnpackLog(event, "Error", log); err != nil {
		return nil, err
	}
	return event, nil
}

// UpKeepingPayKeeperIterator is returned from FilterPayKeeper and is used to iterate over the raw logs and unpacked data for PayKeeper events raised by the UpKeeping contract.
type UpKeepingPayKeeperIterator struct {
	Event *UpKeepingPayKeeper // Event containing the contract specifics and raw log

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
func (it *UpKeepingPayKeeperIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UpKeepingPayKeeper)
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
		it.Event = new(UpKeepingPayKeeper)
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
func (it *UpKeepingPayKeeperIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UpKeepingPayKeeperIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UpKeepingPayKeeper represents a PayKeeper event raised by the UpKeeping contract.
type UpKeepingPayKeeper struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterPayKeeper is a free log retrieval operation binding the contract event 0xaa4c66f6ddfadc835acfabab55148a78bc3e6867ed1cdb36461a10685af4c0c3.
//
// Solidity: event PayKeeper(address indexed from, address indexed to, uint256 indexed value)
func (_UpKeeping *UpKeepingFilterer) FilterPayKeeper(opts *bind.FilterOpts, from []common.Address, to []common.Address, value []*big.Int) (*UpKeepingPayKeeperIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var valueRule []interface{}
	for _, valueItem := range value {
		valueRule = append(valueRule, valueItem)
	}

	logs, sub, err := _UpKeeping.contract.FilterLogs(opts, "PayKeeper", fromRule, toRule, valueRule)
	if err != nil {
		return nil, err
	}
	return &UpKeepingPayKeeperIterator{contract: _UpKeeping.contract, event: "PayKeeper", logs: logs, sub: sub}, nil
}

// WatchPayKeeper is a free log subscription operation binding the contract event 0xaa4c66f6ddfadc835acfabab55148a78bc3e6867ed1cdb36461a10685af4c0c3.
//
// Solidity: event PayKeeper(address indexed from, address indexed to, uint256 indexed value)
func (_UpKeeping *UpKeepingFilterer) WatchPayKeeper(opts *bind.WatchOpts, sink chan<- *UpKeepingPayKeeper, from []common.Address, to []common.Address, value []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var valueRule []interface{}
	for _, valueItem := range value {
		valueRule = append(valueRule, valueItem)
	}

	logs, sub, err := _UpKeeping.contract.WatchLogs(opts, "PayKeeper", fromRule, toRule, valueRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UpKeepingPayKeeper)
				if err := _UpKeeping.contract.UnpackLog(event, "PayKeeper", log); err != nil {
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

// ParsePayKeeper is a log parse operation binding the contract event 0xaa4c66f6ddfadc835acfabab55148a78bc3e6867ed1cdb36461a10685af4c0c3.
//
// Solidity: event PayKeeper(address indexed from, address indexed to, uint256 indexed value)
func (_UpKeeping *UpKeepingFilterer) ParsePayKeeper(log types.Log) (*UpKeepingPayKeeper, error) {
	event := new(UpKeepingPayKeeper)
	if err := _UpKeeping.contract.UnpackLog(event, "PayKeeper", log); err != nil {
		return nil, err
	}
	return event, nil
}

// UpKeepingPayProviderIterator is returned from FilterPayProvider and is used to iterate over the raw logs and unpacked data for PayProvider events raised by the UpKeeping contract.
type UpKeepingPayProviderIterator struct {
	Event *UpKeepingPayProvider // Event containing the contract specifics and raw log

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
func (it *UpKeepingPayProviderIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UpKeepingPayProvider)
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
		it.Event = new(UpKeepingPayProvider)
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
func (it *UpKeepingPayProviderIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UpKeepingPayProviderIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UpKeepingPayProvider represents a PayProvider event raised by the UpKeeping contract.
type UpKeepingPayProvider struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterPayProvider is a free log retrieval operation binding the contract event 0x1569130f5bdbde161a213db1c477e4f2670f09e2a9c1c08ca9bafe749b80cb41.
//
// Solidity: event PayProvider(address indexed from, address indexed to, uint256 indexed value)
func (_UpKeeping *UpKeepingFilterer) FilterPayProvider(opts *bind.FilterOpts, from []common.Address, to []common.Address, value []*big.Int) (*UpKeepingPayProviderIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var valueRule []interface{}
	for _, valueItem := range value {
		valueRule = append(valueRule, valueItem)
	}

	logs, sub, err := _UpKeeping.contract.FilterLogs(opts, "PayProvider", fromRule, toRule, valueRule)
	if err != nil {
		return nil, err
	}
	return &UpKeepingPayProviderIterator{contract: _UpKeeping.contract, event: "PayProvider", logs: logs, sub: sub}, nil
}

// WatchPayProvider is a free log subscription operation binding the contract event 0x1569130f5bdbde161a213db1c477e4f2670f09e2a9c1c08ca9bafe749b80cb41.
//
// Solidity: event PayProvider(address indexed from, address indexed to, uint256 indexed value)
func (_UpKeeping *UpKeepingFilterer) WatchPayProvider(opts *bind.WatchOpts, sink chan<- *UpKeepingPayProvider, from []common.Address, to []common.Address, value []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var valueRule []interface{}
	for _, valueItem := range value {
		valueRule = append(valueRule, valueItem)
	}

	logs, sub, err := _UpKeeping.contract.WatchLogs(opts, "PayProvider", fromRule, toRule, valueRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UpKeepingPayProvider)
				if err := _UpKeeping.contract.UnpackLog(event, "PayProvider", log); err != nil {
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

// ParsePayProvider is a log parse operation binding the contract event 0x1569130f5bdbde161a213db1c477e4f2670f09e2a9c1c08ca9bafe749b80cb41.
//
// Solidity: event PayProvider(address indexed from, address indexed to, uint256 indexed value)
func (_UpKeeping *UpKeepingFilterer) ParsePayProvider(log types.Log) (*UpKeepingPayProvider, error) {
	event := new(UpKeepingPayProvider)
	if err := _UpKeeping.contract.UnpackLog(event, "PayProvider", log); err != nil {
		return nil, err
	}
	return event, nil
}
