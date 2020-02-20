// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package root

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
const DebugBin = `0x6080604052348015600f57600080fd5b50603e80601d6000396000f3fe6080604052600080fdfea265627a7a723058201ec0ce67cfd0b10b0d06b8325cd674162e4c83c698a380080731a174435c9fc464736f6c63430005090032`

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

// OwnedABI is the input ABI used to generate the binding from.
const OwnedABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"data\",\"type\":\"string\"}],\"name\":\"Error\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"}]"

// OwnedBin is the compiled bytecode used for deploying new contracts.
const OwnedBin = `0x608060405234801561001057600080fd5b50600080546001600160a01b031916331790556101b1806100326000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c80630ca05f9f1461003b578063893d20e814610075575b600080fd5b6100616004803603602081101561005157600080fd5b50356001600160a01b0316610099565b604080519115158252519081900360200190f35b61007d61016d565b604080516001600160a01b039092168252519081900360200190f35b600080546001600160a01b031633141561011657600080546001600160a01b038481166001600160a01b0319831681179093556040805191909216808252602082019390935281517f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90929181900390910190a16001915050610168565b604080516020808252600e908201526d725ed0725c46f34c57b7bbb732b960911b8183015290517f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa9181900360600190a15b919050565b6000546001600160a01b03169056fea265627a7a723058206eebcb963d8184f6e9e48542831c9df8599ca71e01328b3816081b8db9b61bf664736f6c63430005090032`

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

// OwnedAbstractABI is the input ABI used to generate the binding from.
const OwnedAbstractABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"}]"

// OwnedAbstractBin is the compiled bytecode used for deploying new contracts.
const OwnedAbstractBin = `0x`

// DeployOwnedAbstract deploys a new Ethereum contract, binding an instance of OwnedAbstract to it.
func DeployOwnedAbstract(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *OwnedAbstract, error) {
	parsed, err := abi.JSON(strings.NewReader(OwnedAbstractABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(OwnedAbstractBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &OwnedAbstract{OwnedAbstractCaller: OwnedAbstractCaller{contract: contract}, OwnedAbstractTransactor: OwnedAbstractTransactor{contract: contract}, OwnedAbstractFilterer: OwnedAbstractFilterer{contract: contract}}, nil
}

// OwnedAbstract is an auto generated Go binding around an Ethereum contract.
type OwnedAbstract struct {
	OwnedAbstractCaller     // Read-only binding to the contract
	OwnedAbstractTransactor // Write-only binding to the contract
	OwnedAbstractFilterer   // Log filterer for contract events
}

// OwnedAbstractCaller is an auto generated read-only Go binding around an Ethereum contract.
type OwnedAbstractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OwnedAbstractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OwnedAbstractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OwnedAbstractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OwnedAbstractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OwnedAbstractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OwnedAbstractSession struct {
	Contract     *OwnedAbstract    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OwnedAbstractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OwnedAbstractCallerSession struct {
	Contract *OwnedAbstractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// OwnedAbstractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OwnedAbstractTransactorSession struct {
	Contract     *OwnedAbstractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// OwnedAbstractRaw is an auto generated low-level Go binding around an Ethereum contract.
type OwnedAbstractRaw struct {
	Contract *OwnedAbstract // Generic contract binding to access the raw methods on
}

// OwnedAbstractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OwnedAbstractCallerRaw struct {
	Contract *OwnedAbstractCaller // Generic read-only contract binding to access the raw methods on
}

// OwnedAbstractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OwnedAbstractTransactorRaw struct {
	Contract *OwnedAbstractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOwnedAbstract creates a new instance of OwnedAbstract, bound to a specific deployed contract.
func NewOwnedAbstract(address common.Address, backend bind.ContractBackend) (*OwnedAbstract, error) {
	contract, err := bindOwnedAbstract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OwnedAbstract{OwnedAbstractCaller: OwnedAbstractCaller{contract: contract}, OwnedAbstractTransactor: OwnedAbstractTransactor{contract: contract}, OwnedAbstractFilterer: OwnedAbstractFilterer{contract: contract}}, nil
}

// NewOwnedAbstractCaller creates a new read-only instance of OwnedAbstract, bound to a specific deployed contract.
func NewOwnedAbstractCaller(address common.Address, caller bind.ContractCaller) (*OwnedAbstractCaller, error) {
	contract, err := bindOwnedAbstract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OwnedAbstractCaller{contract: contract}, nil
}

// NewOwnedAbstractTransactor creates a new write-only instance of OwnedAbstract, bound to a specific deployed contract.
func NewOwnedAbstractTransactor(address common.Address, transactor bind.ContractTransactor) (*OwnedAbstractTransactor, error) {
	contract, err := bindOwnedAbstract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OwnedAbstractTransactor{contract: contract}, nil
}

// NewOwnedAbstractFilterer creates a new log filterer instance of OwnedAbstract, bound to a specific deployed contract.
func NewOwnedAbstractFilterer(address common.Address, filterer bind.ContractFilterer) (*OwnedAbstractFilterer, error) {
	contract, err := bindOwnedAbstract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OwnedAbstractFilterer{contract: contract}, nil
}

// bindOwnedAbstract binds a generic wrapper to an already deployed contract.
func bindOwnedAbstract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(OwnedAbstractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OwnedAbstract *OwnedAbstractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _OwnedAbstract.Contract.OwnedAbstractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OwnedAbstract *OwnedAbstractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OwnedAbstract.Contract.OwnedAbstractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OwnedAbstract *OwnedAbstractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OwnedAbstract.Contract.OwnedAbstractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OwnedAbstract *OwnedAbstractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _OwnedAbstract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OwnedAbstract *OwnedAbstractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OwnedAbstract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OwnedAbstract *OwnedAbstractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OwnedAbstract.Contract.contract.Transact(opts, method, params...)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_OwnedAbstract *OwnedAbstractCaller) GetOwner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _OwnedAbstract.contract.Call(opts, out, "getOwner")
	return *ret0, err
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_OwnedAbstract *OwnedAbstractSession) GetOwner() (common.Address, error) {
	return _OwnedAbstract.Contract.GetOwner(&_OwnedAbstract.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_OwnedAbstract *OwnedAbstractCallerSession) GetOwner() (common.Address, error) {
	return _OwnedAbstract.Contract.GetOwner(&_OwnedAbstract.CallOpts)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_OwnedAbstract *OwnedAbstractTransactor) AlterOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _OwnedAbstract.contract.Transact(opts, "alterOwner", newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_OwnedAbstract *OwnedAbstractSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _OwnedAbstract.Contract.AlterOwner(&_OwnedAbstract.TransactOpts, newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_OwnedAbstract *OwnedAbstractTransactorSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _OwnedAbstract.Contract.AlterOwner(&_OwnedAbstract.TransactOpts, newOwner)
}

// OwnedAbstractAlterOwnerIterator is returned from FilterAlterOwner and is used to iterate over the raw logs and unpacked data for AlterOwner events raised by the OwnedAbstract contract.
type OwnedAbstractAlterOwnerIterator struct {
	Event *OwnedAbstractAlterOwner // Event containing the contract specifics and raw log

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
func (it *OwnedAbstractAlterOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OwnedAbstractAlterOwner)
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
		it.Event = new(OwnedAbstractAlterOwner)
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
func (it *OwnedAbstractAlterOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OwnedAbstractAlterOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OwnedAbstractAlterOwner represents a AlterOwner event raised by the OwnedAbstract contract.
type OwnedAbstractAlterOwner struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAlterOwner is a free log retrieval operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_OwnedAbstract *OwnedAbstractFilterer) FilterAlterOwner(opts *bind.FilterOpts) (*OwnedAbstractAlterOwnerIterator, error) {

	logs, sub, err := _OwnedAbstract.contract.FilterLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return &OwnedAbstractAlterOwnerIterator{contract: _OwnedAbstract.contract, event: "AlterOwner", logs: logs, sub: sub}, nil
}

// WatchAlterOwner is a free log subscription operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_OwnedAbstract *OwnedAbstractFilterer) WatchAlterOwner(opts *bind.WatchOpts, sink chan<- *OwnedAbstractAlterOwner) (event.Subscription, error) {

	logs, sub, err := _OwnedAbstract.contract.WatchLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OwnedAbstractAlterOwner)
				if err := _OwnedAbstract.contract.UnpackLog(event, "AlterOwner", log); err != nil {
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

// RootABI is the input ABI used to generate the binding from.
const RootABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"queryAddr\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getAllKey\",\"outputs\":[{\"name\":\"\",\"type\":\"int64[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"key\",\"type\":\"int64\"},{\"name\":\"value\",\"type\":\"bytes32\"}],\"name\":\"setRoot\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"key\",\"type\":\"int64\"}],\"name\":\"getRoot\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getLatest\",\"outputs\":[{\"name\":\"\",\"type\":\"int64\"},{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_query\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"key\",\"type\":\"int64\"}],\"name\":\"AddRoot\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"data\",\"type\":\"string\"}],\"name\":\"Error\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"}]"

// RootBin is the compiled bytecode used for deploying new contracts.
const RootBin = `0x608060405234801561001057600080fd5b506040516105aa3803806105aa8339818101604052602081101561003357600080fd5b5051600080546001600160a01b03199081163317909155600380546001600160a01b0390931692909116919091179055610538806100726000396000f3fe608060405234801561001057600080fd5b506004361061007d5760003560e01c80637477f8501161005b5780637477f85014610138578063893d20e81461015e578063beac5b6314610166578063c36af460146101985761007d565b80630ca05f9f146100825780632a9a8d8d146100bc5780636b5b4335146100e0575b600080fd5b6100a86004803603602081101561009857600080fd5b50356001600160a01b03166101c3565b604080519115158252519081900360200190f35b6100c4610297565b604080516001600160a01b039092168252519081900360200190f35b6100e86102a6565b60408051602080825283518183015283519192839290830191858101910280838360005b8381101561012457818101518382015260200161010c565b505050509050019250505060405180910390f35b6100a86004803603604081101561014e57600080fd5b50803560070b9060200135610324565b6100c4610464565b6101866004803603602081101561017c57600080fd5b503560070b610473565b60408051918252519081900360200190f35b6101a061048c565b604051808360070b60070b81526020018281526020019250505060405180910390f35b600080546001600160a01b031633141561024057600080546001600160a01b038481166001600160a01b0319831681179093556040805191909216808252602082019390935281517f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90929181900390910190a16001915050610292565b604080516020808252600e908201526d725ed0725c46f34c57b7bbb732b960911b8183015290517f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa9181900360600190a15b919050565b6003546001600160a01b031681565b6060600280548060200260200160405190810160405280929190818152602001828054801561031a57602002820191906000526020600020906000905b82829054906101000a900460070b60070b815260200190600801906020826007010492830192600103820291508084116102e35790505b5050505050905090565b600080546001600160a01b031633141561040c57600783810b900b6000908152600160205260409020546103b757600280546001810182556000919091527f405787fa12a823e0f2b7631cc41b3ba8828b3321ca811111fa75cd3aa3bb5ace6004820401805460039092166008026101000a67ffffffffffffffff81810219909316600787900b93909316029190911790555b600783810b900b600081815260016020908152604091829020859055815192835290517f352d2b8fd5bd4233af9478ba7ca7fe3da4d8a0438736005bc110bef2cab7443a9281900390910190a150600161045e565b604080516020808252600e908201526d725ed0725c46f34c57b7bbb732b960911b8183015290517f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa9181900360600190a15b92915050565b6000546001600160a01b031690565b600790810b900b60009081526001602052604090205490565b6002546000908190806104a65750600091508190506104ff565b6000600260018303815481106104b857fe5b90600052602060002090600491828204019190066008029054906101000a900460070b905080600160008360070b60070b8152602001908152602001600020549350935050505b909156fea265627a7a723058203929a87e7cf5d0ed983c50db0f89e3423f039c612b0a0119236470dfbbd5072b64736f6c63430005090032`

// DeployRoot deploys a new Ethereum contract, binding an instance of Root to it.
func DeployRoot(auth *bind.TransactOpts, backend bind.ContractBackend, _query common.Address) (common.Address, *types.Transaction, *Root, error) {
	parsed, err := abi.JSON(strings.NewReader(RootABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(RootBin), backend, _query)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Root{RootCaller: RootCaller{contract: contract}, RootTransactor: RootTransactor{contract: contract}, RootFilterer: RootFilterer{contract: contract}}, nil
}

// Root is an auto generated Go binding around an Ethereum contract.
type Root struct {
	RootCaller     // Read-only binding to the contract
	RootTransactor // Write-only binding to the contract
	RootFilterer   // Log filterer for contract events
}

// RootCaller is an auto generated read-only Go binding around an Ethereum contract.
type RootCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RootTransactor is an auto generated write-only Go binding around an Ethereum contract.
type RootTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RootFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type RootFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RootSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type RootSession struct {
	Contract     *Root             // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RootCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type RootCallerSession struct {
	Contract *RootCaller   // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// RootTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type RootTransactorSession struct {
	Contract     *RootTransactor   // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RootRaw is an auto generated low-level Go binding around an Ethereum contract.
type RootRaw struct {
	Contract *Root // Generic contract binding to access the raw methods on
}

// RootCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type RootCallerRaw struct {
	Contract *RootCaller // Generic read-only contract binding to access the raw methods on
}

// RootTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type RootTransactorRaw struct {
	Contract *RootTransactor // Generic write-only contract binding to access the raw methods on
}

// NewRoot creates a new instance of Root, bound to a specific deployed contract.
func NewRoot(address common.Address, backend bind.ContractBackend) (*Root, error) {
	contract, err := bindRoot(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Root{RootCaller: RootCaller{contract: contract}, RootTransactor: RootTransactor{contract: contract}, RootFilterer: RootFilterer{contract: contract}}, nil
}

// NewRootCaller creates a new read-only instance of Root, bound to a specific deployed contract.
func NewRootCaller(address common.Address, caller bind.ContractCaller) (*RootCaller, error) {
	contract, err := bindRoot(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &RootCaller{contract: contract}, nil
}

// NewRootTransactor creates a new write-only instance of Root, bound to a specific deployed contract.
func NewRootTransactor(address common.Address, transactor bind.ContractTransactor) (*RootTransactor, error) {
	contract, err := bindRoot(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &RootTransactor{contract: contract}, nil
}

// NewRootFilterer creates a new log filterer instance of Root, bound to a specific deployed contract.
func NewRootFilterer(address common.Address, filterer bind.ContractFilterer) (*RootFilterer, error) {
	contract, err := bindRoot(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &RootFilterer{contract: contract}, nil
}

// bindRoot binds a generic wrapper to an already deployed contract.
func bindRoot(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(RootABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Root *RootRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Root.Contract.RootCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Root *RootRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Root.Contract.RootTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Root *RootRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Root.Contract.RootTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Root *RootCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Root.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Root *RootTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Root.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Root *RootTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Root.Contract.contract.Transact(opts, method, params...)
}

// GetAllKey is a free data retrieval call binding the contract method 0x6b5b4335.
//
// Solidity: function getAllKey() constant returns(int64[])
func (_Root *RootCaller) GetAllKey(opts *bind.CallOpts) ([]int64, error) {
	var (
		ret0 = new([]int64)
	)
	out := ret0
	err := _Root.contract.Call(opts, out, "getAllKey")
	return *ret0, err
}

// GetAllKey is a free data retrieval call binding the contract method 0x6b5b4335.
//
// Solidity: function getAllKey() constant returns(int64[])
func (_Root *RootSession) GetAllKey() ([]int64, error) {
	return _Root.Contract.GetAllKey(&_Root.CallOpts)
}

// GetAllKey is a free data retrieval call binding the contract method 0x6b5b4335.
//
// Solidity: function getAllKey() constant returns(int64[])
func (_Root *RootCallerSession) GetAllKey() ([]int64, error) {
	return _Root.Contract.GetAllKey(&_Root.CallOpts)
}

// GetLatest is a free data retrieval call binding the contract method 0xc36af460.
//
// Solidity: function getLatest() constant returns(int64, bytes32)
func (_Root *RootCaller) GetLatest(opts *bind.CallOpts) (int64, [32]byte, error) {
	var (
		ret0 = new(int64)
		ret1 = new([32]byte)
	)
	out := &[]interface{}{
		ret0,
		ret1,
	}
	err := _Root.contract.Call(opts, out, "getLatest")
	return *ret0, *ret1, err
}

// GetLatest is a free data retrieval call binding the contract method 0xc36af460.
//
// Solidity: function getLatest() constant returns(int64, bytes32)
func (_Root *RootSession) GetLatest() (int64, [32]byte, error) {
	return _Root.Contract.GetLatest(&_Root.CallOpts)
}

// GetLatest is a free data retrieval call binding the contract method 0xc36af460.
//
// Solidity: function getLatest() constant returns(int64, bytes32)
func (_Root *RootCallerSession) GetLatest() (int64, [32]byte, error) {
	return _Root.Contract.GetLatest(&_Root.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Root *RootCaller) GetOwner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Root.contract.Call(opts, out, "getOwner")
	return *ret0, err
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Root *RootSession) GetOwner() (common.Address, error) {
	return _Root.Contract.GetOwner(&_Root.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Root *RootCallerSession) GetOwner() (common.Address, error) {
	return _Root.Contract.GetOwner(&_Root.CallOpts)
}

// GetRoot is a free data retrieval call binding the contract method 0xbeac5b63.
//
// Solidity: function getRoot(int64 key) constant returns(bytes32)
func (_Root *RootCaller) GetRoot(opts *bind.CallOpts, key int64) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _Root.contract.Call(opts, out, "getRoot", key)
	return *ret0, err
}

// GetRoot is a free data retrieval call binding the contract method 0xbeac5b63.
//
// Solidity: function getRoot(int64 key) constant returns(bytes32)
func (_Root *RootSession) GetRoot(key int64) ([32]byte, error) {
	return _Root.Contract.GetRoot(&_Root.CallOpts, key)
}

// GetRoot is a free data retrieval call binding the contract method 0xbeac5b63.
//
// Solidity: function getRoot(int64 key) constant returns(bytes32)
func (_Root *RootCallerSession) GetRoot(key int64) ([32]byte, error) {
	return _Root.Contract.GetRoot(&_Root.CallOpts, key)
}

// QueryAddr is a free data retrieval call binding the contract method 0x2a9a8d8d.
//
// Solidity: function queryAddr() constant returns(address)
func (_Root *RootCaller) QueryAddr(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Root.contract.Call(opts, out, "queryAddr")
	return *ret0, err
}

// QueryAddr is a free data retrieval call binding the contract method 0x2a9a8d8d.
//
// Solidity: function queryAddr() constant returns(address)
func (_Root *RootSession) QueryAddr() (common.Address, error) {
	return _Root.Contract.QueryAddr(&_Root.CallOpts)
}

// QueryAddr is a free data retrieval call binding the contract method 0x2a9a8d8d.
//
// Solidity: function queryAddr() constant returns(address)
func (_Root *RootCallerSession) QueryAddr() (common.Address, error) {
	return _Root.Contract.QueryAddr(&_Root.CallOpts)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Root *RootTransactor) AlterOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Root.contract.Transact(opts, "alterOwner", newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Root *RootSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Root.Contract.AlterOwner(&_Root.TransactOpts, newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Root *RootTransactorSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Root.Contract.AlterOwner(&_Root.TransactOpts, newOwner)
}

// SetRoot is a paid mutator transaction binding the contract method 0x7477f850.
//
// Solidity: function setRoot(int64 key, bytes32 value) returns(bool)
func (_Root *RootTransactor) SetRoot(opts *bind.TransactOpts, key int64, value [32]byte) (*types.Transaction, error) {
	return _Root.contract.Transact(opts, "setRoot", key, value)
}

// SetRoot is a paid mutator transaction binding the contract method 0x7477f850.
//
// Solidity: function setRoot(int64 key, bytes32 value) returns(bool)
func (_Root *RootSession) SetRoot(key int64, value [32]byte) (*types.Transaction, error) {
	return _Root.Contract.SetRoot(&_Root.TransactOpts, key, value)
}

// SetRoot is a paid mutator transaction binding the contract method 0x7477f850.
//
// Solidity: function setRoot(int64 key, bytes32 value) returns(bool)
func (_Root *RootTransactorSession) SetRoot(key int64, value [32]byte) (*types.Transaction, error) {
	return _Root.Contract.SetRoot(&_Root.TransactOpts, key, value)
}

// RootAddRootIterator is returned from FilterAddRoot and is used to iterate over the raw logs and unpacked data for AddRoot events raised by the Root contract.
type RootAddRootIterator struct {
	Event *RootAddRoot // Event containing the contract specifics and raw log

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
func (it *RootAddRootIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RootAddRoot)
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
		it.Event = new(RootAddRoot)
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
func (it *RootAddRootIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RootAddRootIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RootAddRoot represents a AddRoot event raised by the Root contract.
type RootAddRoot struct {
	Key int64
	Raw types.Log // Blockchain specific contextual infos
}

// FilterAddRoot is a free log retrieval operation binding the contract event 0x352d2b8fd5bd4233af9478ba7ca7fe3da4d8a0438736005bc110bef2cab7443a.
//
// Solidity: event AddRoot(int64 key)
func (_Root *RootFilterer) FilterAddRoot(opts *bind.FilterOpts) (*RootAddRootIterator, error) {

	logs, sub, err := _Root.contract.FilterLogs(opts, "AddRoot")
	if err != nil {
		return nil, err
	}
	return &RootAddRootIterator{contract: _Root.contract, event: "AddRoot", logs: logs, sub: sub}, nil
}

// WatchAddRoot is a free log subscription operation binding the contract event 0x352d2b8fd5bd4233af9478ba7ca7fe3da4d8a0438736005bc110bef2cab7443a.
//
// Solidity: event AddRoot(int64 key)
func (_Root *RootFilterer) WatchAddRoot(opts *bind.WatchOpts, sink chan<- *RootAddRoot) (event.Subscription, error) {

	logs, sub, err := _Root.contract.WatchLogs(opts, "AddRoot")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RootAddRoot)
				if err := _Root.contract.UnpackLog(event, "AddRoot", log); err != nil {
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

// RootAlterOwnerIterator is returned from FilterAlterOwner and is used to iterate over the raw logs and unpacked data for AlterOwner events raised by the Root contract.
type RootAlterOwnerIterator struct {
	Event *RootAlterOwner // Event containing the contract specifics and raw log

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
func (it *RootAlterOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RootAlterOwner)
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
		it.Event = new(RootAlterOwner)
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
func (it *RootAlterOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RootAlterOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RootAlterOwner represents a AlterOwner event raised by the Root contract.
type RootAlterOwner struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAlterOwner is a free log retrieval operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Root *RootFilterer) FilterAlterOwner(opts *bind.FilterOpts) (*RootAlterOwnerIterator, error) {

	logs, sub, err := _Root.contract.FilterLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return &RootAlterOwnerIterator{contract: _Root.contract, event: "AlterOwner", logs: logs, sub: sub}, nil
}

// WatchAlterOwner is a free log subscription operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Root *RootFilterer) WatchAlterOwner(opts *bind.WatchOpts, sink chan<- *RootAlterOwner) (event.Subscription, error) {

	logs, sub, err := _Root.contract.WatchLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RootAlterOwner)
				if err := _Root.contract.UnpackLog(event, "AlterOwner", log); err != nil {
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

// RootErrorIterator is returned from FilterError and is used to iterate over the raw logs and unpacked data for Error events raised by the Root contract.
type RootErrorIterator struct {
	Event *RootError // Event containing the contract specifics and raw log

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
func (it *RootErrorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RootError)
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
		it.Event = new(RootError)
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
func (it *RootErrorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RootErrorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RootError represents a Error event raised by the Root contract.
type RootError struct {
	Data string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterError is a free log retrieval operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Root *RootFilterer) FilterError(opts *bind.FilterOpts) (*RootErrorIterator, error) {

	logs, sub, err := _Root.contract.FilterLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return &RootErrorIterator{contract: _Root.contract, event: "Error", logs: logs, sub: sub}, nil
}

// WatchError is a free log subscription operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Root *RootFilterer) WatchError(opts *bind.WatchOpts, sink chan<- *RootError) (event.Subscription, error) {

	logs, sub, err := _Root.contract.WatchLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RootError)
				if err := _Root.contract.UnpackLog(event, "Error", log); err != nil {
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
