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

// RootABI is the input ABI used to generate the binding from.
const RootABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_query\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"int64\",\"name\":\"key\",\"type\":\"int64\"}],\"name\":\"AddRoot\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAllKey\",\"outputs\":[{\"internalType\":\"int64[]\",\"name\":\"\",\"type\":\"int64[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getLatest\",\"outputs\":[{\"internalType\":\"int64\",\"name\":\"\",\"type\":\"int64\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"int64\",\"name\":\"key\",\"type\":\"int64\"}],\"name\":\"getRoot\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"queryAddr\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"int64\",\"name\":\"key\",\"type\":\"int64\"},{\"internalType\":\"bytes32\",\"name\":\"value\",\"type\":\"bytes32\"}],\"name\":\"setRoot\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// RootBin is the compiled bytecode used for deploying new contracts.
var RootBin = "0x608060405234801561001057600080fd5b506040516105803803806105808339818101604052602081101561003357600080fd5b5051600080546001600160a01b03199081163317909155600380546001600160a01b039093169290911691909117905561050e806100726000396000f3fe608060405234801561001057600080fd5b506004361061007d5760003560e01c80637477f8501161005b5780637477f85014610138578063893d20e81461015e578063beac5b6314610166578063c36af460146101985761007d565b80630ca05f9f146100825780632a9a8d8d146100bc5780636b5b4335146100e0575b600080fd5b6100a86004803603602081101561009857600080fd5b50356001600160a01b03166101c3565b604080519115158252519081900360200190f35b6100c4610281565b604080516001600160a01b039092168252519081900360200190f35b6100e8610290565b60408051602080825283518183015283519192839290830191858101910280838360005b8381101561012457818101518382015260200161010c565b505050509050019250505060405180910390f35b6100a86004803603604081101561014e57600080fd5b50803560070b906020013561030e565b6100c4610439565b6101866004803603602081101561017c57600080fd5b503560070b610448565b60408051918252519081900360200190f35b6101a0610461565b604051808360070b60070b81526020018281526020019250505060405180910390f35b600080546001600160a01b03163314610219576040805162461bcd60e51b81526020600482015260136024820152721bdb9b1e481bdddb995c8818d85b8818d85b1b606a1b604482015290519081900360640190fd5b600080546001600160a01b038481166001600160a01b0319831681179093556040805191909216808252602082019390935281517f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90929181900390910190a150600192915050565b6003546001600160a01b031681565b6060600280548060200260200160405190810160405280929190818152602001828054801561030457602002820191906000526020600020906000905b82829054906101000a900460070b60070b815260200190600801906020826007010492830192600103820291508084116102cd5790505b5050505050905090565b600080546001600160a01b03163314610364576040805162461bcd60e51b81526020600482015260136024820152721bdb9b1e481bdddb995c8818d85b8818d85b1b606a1b604482015290519081900360640190fd5b600783810b900b6000908152600160205260409020546103e357600280546001810182556000919091527f405787fa12a823e0f2b7631cc41b3ba8828b3321ca811111fa75cd3aa3bb5ace6004820401805460039092166008026101000a67ffffffffffffffff81810219909316600787900b93909316029190911790555b600783810b900b600081815260016020908152604091829020859055815192835290517f352d2b8fd5bd4233af9478ba7ca7fe3da4d8a0438736005bc110bef2cab7443a9281900390910190a150600192915050565b6000546001600160a01b031690565b600790810b900b60009081526001602052604090205490565b60025460009081908061047b5750600091508190506104d4565b60006002600183038154811061048d57fe5b90600052602060002090600491828204019190066008029054906101000a900460070b905080600160008360070b60070b8152602001908152602001600020549350935050505b909156fea2646970667358221220092e90847d5a0e5774f619d383149322d2e56479573405cddb911274a588c3d064736f6c63430006030033"

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

// ParseAddRoot is a log parse operation binding the contract event 0x352d2b8fd5bd4233af9478ba7ca7fe3da4d8a0438736005bc110bef2cab7443a.
//
// Solidity: event AddRoot(int64 key)
func (_Root *RootFilterer) ParseAddRoot(log types.Log) (*RootAddRoot, error) {
	event := new(RootAddRoot)
	if err := _Root.contract.UnpackLog(event, "AddRoot", log); err != nil {
		return nil, err
	}
	return event, nil
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

// ParseAlterOwner is a log parse operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Root *RootFilterer) ParseAlterOwner(log types.Log) (*RootAlterOwner, error) {
	event := new(RootAlterOwner)
	if err := _Root.contract.UnpackLog(event, "AlterOwner", log); err != nil {
		return nil, err
	}
	return event, nil
}
