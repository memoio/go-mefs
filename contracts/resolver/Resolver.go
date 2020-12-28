// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package resolver

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

// ResolverABI is the input ABI used to generate the binding from.
const ResolverABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"mapper\",\"type\":\"address\"}],\"name\":\"Add\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"oldMapper\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newMapper\",\"type\":\"address\"}],\"name\":\"AlterMapper\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"data\",\"type\":\"string\"}],\"name\":\"Error\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"mapper\",\"type\":\"address\"}],\"name\":\"add\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"mapper\",\"type\":\"address\"}],\"name\":\"alterMapper\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"get\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getBanned\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"param\",\"type\":\"bool\"}],\"name\":\"setBanned\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// ResolverBin is the compiled bytecode used for deploying new contracts.
var ResolverBin = "0x60806040526000600260006101000a81548160ff02191690831515021790555034801561002b57600080fd5b50336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550610ad68061007b6000396000f3fe608060405234801561001057600080fd5b506004361061007d5760003560e01c80636f88b5791161005b5780636f88b57914610166578063893d20e814610186578063ac5c505e146101ba578063c2bc2efc146102145761007d565b80630a3b0a4f146100825780630ca05f9f146100dc5780634b9c5d3b14610136575b600080fd5b6100c46004803603602081101561009857600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610282565b60405180821515815260200191505060405180910390f35b61011e600480360360208110156100f257600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506104fe565b60405180821515815260200191505060405180910390f35b6101646004803603602081101561014c57600080fd5b8101908080351515906020019092919050505061069d565b005b61016e61077b565b60405180821515815260200191505060405180910390f35b61018e610792565b604051808273ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6101fc600480360360208110156101d057600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506107bb565b60405180821515815260200191505060405180910390f35b6102566004803603602081101561022a57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610a37565b604051808273ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6000600260009054906101000a900460ff1615610307576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260098152602001807f69732062616e6e6564000000000000000000000000000000000000000000000081525060200191505060405180910390fd5b600073ffffffffffffffffffffffffffffffffffffffff16600160003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff161461040b577f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa60405180806020018281038252600b8152602001807f68617320616c726561647900000000000000000000000000000000000000000081525060200191505060405180910390a1600090506104f9565b81600160003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f473b736fe95295e8fbc851ca8acdc12a750976edad27a92f666b3d888eb895d33383604051808373ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a1600190505b919050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146105c2576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050826000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e908184604051808373ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a16001915050919050565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461075e576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b80600260006101000a81548160ff02191690831515021790555050565b6000600260009054906101000a900460ff16905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b60008073ffffffffffffffffffffffffffffffffffffffff16600160003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614156108c1577f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa6040518080602001828103825260088152602001807f6e6f74206861766500000000000000000000000000000000000000000000000081525060200191505060405180910390a160009050610a32565b6000600160003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905082600160003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507fa74fe41d06f59ab4da1dec9b736b2e9cc0b6f36b502d0c5276c5e52b2f2f8dd2338285604051808473ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff168152602001935050505060405180910390a160019150505b919050565b6000600160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905091905056fea26469706673582212201471bed20715e56ab644cc90dec1535ed24c7963f5b5a2b815dcdf801f06250a64736f6c63430007030033"

// DeployResolver deploys a new Ethereum contract, binding an instance of Resolver to it.
func DeployResolver(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Resolver, error) {
	parsed, err := abi.JSON(strings.NewReader(ResolverABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ResolverBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Resolver{ResolverCaller: ResolverCaller{contract: contract}, ResolverTransactor: ResolverTransactor{contract: contract}, ResolverFilterer: ResolverFilterer{contract: contract}}, nil
}

// Resolver is an auto generated Go binding around an Ethereum contract.
type Resolver struct {
	ResolverCaller     // Read-only binding to the contract
	ResolverTransactor // Write-only binding to the contract
	ResolverFilterer   // Log filterer for contract events
}

// ResolverCaller is an auto generated read-only Go binding around an Ethereum contract.
type ResolverCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ResolverTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ResolverTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ResolverFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ResolverFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ResolverSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ResolverSession struct {
	Contract     *Resolver         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ResolverCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ResolverCallerSession struct {
	Contract *ResolverCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// ResolverTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ResolverTransactorSession struct {
	Contract     *ResolverTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ResolverRaw is an auto generated low-level Go binding around an Ethereum contract.
type ResolverRaw struct {
	Contract *Resolver // Generic contract binding to access the raw methods on
}

// ResolverCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ResolverCallerRaw struct {
	Contract *ResolverCaller // Generic read-only contract binding to access the raw methods on
}

// ResolverTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ResolverTransactorRaw struct {
	Contract *ResolverTransactor // Generic write-only contract binding to access the raw methods on
}

// NewResolver creates a new instance of Resolver, bound to a specific deployed contract.
func NewResolver(address common.Address, backend bind.ContractBackend) (*Resolver, error) {
	contract, err := bindResolver(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Resolver{ResolverCaller: ResolverCaller{contract: contract}, ResolverTransactor: ResolverTransactor{contract: contract}, ResolverFilterer: ResolverFilterer{contract: contract}}, nil
}

// NewResolverCaller creates a new read-only instance of Resolver, bound to a specific deployed contract.
func NewResolverCaller(address common.Address, caller bind.ContractCaller) (*ResolverCaller, error) {
	contract, err := bindResolver(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ResolverCaller{contract: contract}, nil
}

// NewResolverTransactor creates a new write-only instance of Resolver, bound to a specific deployed contract.
func NewResolverTransactor(address common.Address, transactor bind.ContractTransactor) (*ResolverTransactor, error) {
	contract, err := bindResolver(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ResolverTransactor{contract: contract}, nil
}

// NewResolverFilterer creates a new log filterer instance of Resolver, bound to a specific deployed contract.
func NewResolverFilterer(address common.Address, filterer bind.ContractFilterer) (*ResolverFilterer, error) {
	contract, err := bindResolver(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ResolverFilterer{contract: contract}, nil
}

// bindResolver binds a generic wrapper to an already deployed contract.
func bindResolver(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ResolverABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Resolver *ResolverRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Resolver.Contract.ResolverCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Resolver *ResolverRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Resolver.Contract.ResolverTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Resolver *ResolverRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Resolver.Contract.ResolverTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Resolver *ResolverCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Resolver.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Resolver *ResolverTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Resolver.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Resolver *ResolverTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Resolver.Contract.contract.Transact(opts, method, params...)
}

// Get is a free data retrieval call binding the contract method 0xc2bc2efc.
//
// Solidity: function get(address owner) view returns(address)
func (_Resolver *ResolverCaller) Get(opts *bind.CallOpts, owner common.Address) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Resolver.contract.Call(opts, out, "get", owner)
	return *ret0, err
}

// Get is a free data retrieval call binding the contract method 0xc2bc2efc.
//
// Solidity: function get(address owner) view returns(address)
func (_Resolver *ResolverSession) Get(owner common.Address) (common.Address, error) {
	return _Resolver.Contract.Get(&_Resolver.CallOpts, owner)
}

// Get is a free data retrieval call binding the contract method 0xc2bc2efc.
//
// Solidity: function get(address owner) view returns(address)
func (_Resolver *ResolverCallerSession) Get(owner common.Address) (common.Address, error) {
	return _Resolver.Contract.Get(&_Resolver.CallOpts, owner)
}

// GetBanned is a free data retrieval call binding the contract method 0x6f88b579.
//
// Solidity: function getBanned() view returns(bool)
func (_Resolver *ResolverCaller) GetBanned(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Resolver.contract.Call(opts, out, "getBanned")
	return *ret0, err
}

// GetBanned is a free data retrieval call binding the contract method 0x6f88b579.
//
// Solidity: function getBanned() view returns(bool)
func (_Resolver *ResolverSession) GetBanned() (bool, error) {
	return _Resolver.Contract.GetBanned(&_Resolver.CallOpts)
}

// GetBanned is a free data retrieval call binding the contract method 0x6f88b579.
//
// Solidity: function getBanned() view returns(bool)
func (_Resolver *ResolverCallerSession) GetBanned() (bool, error) {
	return _Resolver.Contract.GetBanned(&_Resolver.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_Resolver *ResolverCaller) GetOwner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Resolver.contract.Call(opts, out, "getOwner")
	return *ret0, err
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_Resolver *ResolverSession) GetOwner() (common.Address, error) {
	return _Resolver.Contract.GetOwner(&_Resolver.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_Resolver *ResolverCallerSession) GetOwner() (common.Address, error) {
	return _Resolver.Contract.GetOwner(&_Resolver.CallOpts)
}

// Add is a paid mutator transaction binding the contract method 0x0a3b0a4f.
//
// Solidity: function add(address mapper) returns(bool)
func (_Resolver *ResolverTransactor) Add(opts *bind.TransactOpts, mapper common.Address) (*types.Transaction, error) {
	return _Resolver.contract.Transact(opts, "add", mapper)
}

// Add is a paid mutator transaction binding the contract method 0x0a3b0a4f.
//
// Solidity: function add(address mapper) returns(bool)
func (_Resolver *ResolverSession) Add(mapper common.Address) (*types.Transaction, error) {
	return _Resolver.Contract.Add(&_Resolver.TransactOpts, mapper)
}

// Add is a paid mutator transaction binding the contract method 0x0a3b0a4f.
//
// Solidity: function add(address mapper) returns(bool)
func (_Resolver *ResolverTransactorSession) Add(mapper common.Address) (*types.Transaction, error) {
	return _Resolver.Contract.Add(&_Resolver.TransactOpts, mapper)
}

// AlterMapper is a paid mutator transaction binding the contract method 0xac5c505e.
//
// Solidity: function alterMapper(address mapper) returns(bool)
func (_Resolver *ResolverTransactor) AlterMapper(opts *bind.TransactOpts, mapper common.Address) (*types.Transaction, error) {
	return _Resolver.contract.Transact(opts, "alterMapper", mapper)
}

// AlterMapper is a paid mutator transaction binding the contract method 0xac5c505e.
//
// Solidity: function alterMapper(address mapper) returns(bool)
func (_Resolver *ResolverSession) AlterMapper(mapper common.Address) (*types.Transaction, error) {
	return _Resolver.Contract.AlterMapper(&_Resolver.TransactOpts, mapper)
}

// AlterMapper is a paid mutator transaction binding the contract method 0xac5c505e.
//
// Solidity: function alterMapper(address mapper) returns(bool)
func (_Resolver *ResolverTransactorSession) AlterMapper(mapper common.Address) (*types.Transaction, error) {
	return _Resolver.Contract.AlterMapper(&_Resolver.TransactOpts, mapper)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Resolver *ResolverTransactor) AlterOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Resolver.contract.Transact(opts, "alterOwner", newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Resolver *ResolverSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Resolver.Contract.AlterOwner(&_Resolver.TransactOpts, newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Resolver *ResolverTransactorSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Resolver.Contract.AlterOwner(&_Resolver.TransactOpts, newOwner)
}

// SetBanned is a paid mutator transaction binding the contract method 0x4b9c5d3b.
//
// Solidity: function setBanned(bool param) returns()
func (_Resolver *ResolverTransactor) SetBanned(opts *bind.TransactOpts, param bool) (*types.Transaction, error) {
	return _Resolver.contract.Transact(opts, "setBanned", param)
}

// SetBanned is a paid mutator transaction binding the contract method 0x4b9c5d3b.
//
// Solidity: function setBanned(bool param) returns()
func (_Resolver *ResolverSession) SetBanned(param bool) (*types.Transaction, error) {
	return _Resolver.Contract.SetBanned(&_Resolver.TransactOpts, param)
}

// SetBanned is a paid mutator transaction binding the contract method 0x4b9c5d3b.
//
// Solidity: function setBanned(bool param) returns()
func (_Resolver *ResolverTransactorSession) SetBanned(param bool) (*types.Transaction, error) {
	return _Resolver.Contract.SetBanned(&_Resolver.TransactOpts, param)
}

// ResolverAddIterator is returned from FilterAdd and is used to iterate over the raw logs and unpacked data for Add events raised by the Resolver contract.
type ResolverAddIterator struct {
	Event *ResolverAdd // Event containing the contract specifics and raw log

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
func (it *ResolverAddIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ResolverAdd)
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
		it.Event = new(ResolverAdd)
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
func (it *ResolverAddIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ResolverAddIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ResolverAdd represents a Add event raised by the Resolver contract.
type ResolverAdd struct {
	Owner  common.Address
	Mapper common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterAdd is a free log retrieval operation binding the contract event 0x473b736fe95295e8fbc851ca8acdc12a750976edad27a92f666b3d888eb895d3.
//
// Solidity: event Add(address owner, address mapper)
func (_Resolver *ResolverFilterer) FilterAdd(opts *bind.FilterOpts) (*ResolverAddIterator, error) {

	logs, sub, err := _Resolver.contract.FilterLogs(opts, "Add")
	if err != nil {
		return nil, err
	}
	return &ResolverAddIterator{contract: _Resolver.contract, event: "Add", logs: logs, sub: sub}, nil
}

// WatchAdd is a free log subscription operation binding the contract event 0x473b736fe95295e8fbc851ca8acdc12a750976edad27a92f666b3d888eb895d3.
//
// Solidity: event Add(address owner, address mapper)
func (_Resolver *ResolverFilterer) WatchAdd(opts *bind.WatchOpts, sink chan<- *ResolverAdd) (event.Subscription, error) {

	logs, sub, err := _Resolver.contract.WatchLogs(opts, "Add")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ResolverAdd)
				if err := _Resolver.contract.UnpackLog(event, "Add", log); err != nil {
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

// ParseAdd is a log parse operation binding the contract event 0x473b736fe95295e8fbc851ca8acdc12a750976edad27a92f666b3d888eb895d3.
//
// Solidity: event Add(address owner, address mapper)
func (_Resolver *ResolverFilterer) ParseAdd(log types.Log) (*ResolverAdd, error) {
	event := new(ResolverAdd)
	if err := _Resolver.contract.UnpackLog(event, "Add", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ResolverAlterMapperIterator is returned from FilterAlterMapper and is used to iterate over the raw logs and unpacked data for AlterMapper events raised by the Resolver contract.
type ResolverAlterMapperIterator struct {
	Event *ResolverAlterMapper // Event containing the contract specifics and raw log

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
func (it *ResolverAlterMapperIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ResolverAlterMapper)
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
		it.Event = new(ResolverAlterMapper)
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
func (it *ResolverAlterMapperIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ResolverAlterMapperIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ResolverAlterMapper represents a AlterMapper event raised by the Resolver contract.
type ResolverAlterMapper struct {
	Owner     common.Address
	OldMapper common.Address
	NewMapper common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterAlterMapper is a free log retrieval operation binding the contract event 0xa74fe41d06f59ab4da1dec9b736b2e9cc0b6f36b502d0c5276c5e52b2f2f8dd2.
//
// Solidity: event AlterMapper(address owner, address oldMapper, address newMapper)
func (_Resolver *ResolverFilterer) FilterAlterMapper(opts *bind.FilterOpts) (*ResolverAlterMapperIterator, error) {

	logs, sub, err := _Resolver.contract.FilterLogs(opts, "AlterMapper")
	if err != nil {
		return nil, err
	}
	return &ResolverAlterMapperIterator{contract: _Resolver.contract, event: "AlterMapper", logs: logs, sub: sub}, nil
}

// WatchAlterMapper is a free log subscription operation binding the contract event 0xa74fe41d06f59ab4da1dec9b736b2e9cc0b6f36b502d0c5276c5e52b2f2f8dd2.
//
// Solidity: event AlterMapper(address owner, address oldMapper, address newMapper)
func (_Resolver *ResolverFilterer) WatchAlterMapper(opts *bind.WatchOpts, sink chan<- *ResolverAlterMapper) (event.Subscription, error) {

	logs, sub, err := _Resolver.contract.WatchLogs(opts, "AlterMapper")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ResolverAlterMapper)
				if err := _Resolver.contract.UnpackLog(event, "AlterMapper", log); err != nil {
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

// ParseAlterMapper is a log parse operation binding the contract event 0xa74fe41d06f59ab4da1dec9b736b2e9cc0b6f36b502d0c5276c5e52b2f2f8dd2.
//
// Solidity: event AlterMapper(address owner, address oldMapper, address newMapper)
func (_Resolver *ResolverFilterer) ParseAlterMapper(log types.Log) (*ResolverAlterMapper, error) {
	event := new(ResolverAlterMapper)
	if err := _Resolver.contract.UnpackLog(event, "AlterMapper", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ResolverAlterOwnerIterator is returned from FilterAlterOwner and is used to iterate over the raw logs and unpacked data for AlterOwner events raised by the Resolver contract.
type ResolverAlterOwnerIterator struct {
	Event *ResolverAlterOwner // Event containing the contract specifics and raw log

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
func (it *ResolverAlterOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ResolverAlterOwner)
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
		it.Event = new(ResolverAlterOwner)
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
func (it *ResolverAlterOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ResolverAlterOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ResolverAlterOwner represents a AlterOwner event raised by the Resolver contract.
type ResolverAlterOwner struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAlterOwner is a free log retrieval operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Resolver *ResolverFilterer) FilterAlterOwner(opts *bind.FilterOpts) (*ResolverAlterOwnerIterator, error) {

	logs, sub, err := _Resolver.contract.FilterLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return &ResolverAlterOwnerIterator{contract: _Resolver.contract, event: "AlterOwner", logs: logs, sub: sub}, nil
}

// WatchAlterOwner is a free log subscription operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Resolver *ResolverFilterer) WatchAlterOwner(opts *bind.WatchOpts, sink chan<- *ResolverAlterOwner) (event.Subscription, error) {

	logs, sub, err := _Resolver.contract.WatchLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ResolverAlterOwner)
				if err := _Resolver.contract.UnpackLog(event, "AlterOwner", log); err != nil {
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
func (_Resolver *ResolverFilterer) ParseAlterOwner(log types.Log) (*ResolverAlterOwner, error) {
	event := new(ResolverAlterOwner)
	if err := _Resolver.contract.UnpackLog(event, "AlterOwner", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ResolverErrorIterator is returned from FilterError and is used to iterate over the raw logs and unpacked data for Error events raised by the Resolver contract.
type ResolverErrorIterator struct {
	Event *ResolverError // Event containing the contract specifics and raw log

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
func (it *ResolverErrorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ResolverError)
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
		it.Event = new(ResolverError)
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
func (it *ResolverErrorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ResolverErrorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ResolverError represents a Error event raised by the Resolver contract.
type ResolverError struct {
	Data string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterError is a free log retrieval operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Resolver *ResolverFilterer) FilterError(opts *bind.FilterOpts) (*ResolverErrorIterator, error) {

	logs, sub, err := _Resolver.contract.FilterLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return &ResolverErrorIterator{contract: _Resolver.contract, event: "Error", logs: logs, sub: sub}, nil
}

// WatchError is a free log subscription operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Resolver *ResolverFilterer) WatchError(opts *bind.WatchOpts, sink chan<- *ResolverError) (event.Subscription, error) {

	logs, sub, err := _Resolver.contract.WatchLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ResolverError)
				if err := _Resolver.contract.UnpackLog(event, "Error", log); err != nil {
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
func (_Resolver *ResolverFilterer) ParseError(log types.Log) (*ResolverError, error) {
	event := new(ResolverError)
	if err := _Resolver.contract.UnpackLog(event, "Error", log); err != nil {
		return nil, err
	}
	return event, nil
}
