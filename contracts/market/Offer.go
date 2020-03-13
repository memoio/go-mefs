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

// OfferABI is the input ABI used to generate the binding from.
const OfferABI = "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"capacity\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"get\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// OfferBin is the compiled bytecode used for deploying new contracts.
var OfferBin = "0x608060405234801561001057600080fd5b506040516102803803806102808339818101604052606081101561003357600080fd5b508051602080830151604093840151600080546001600160a01b031916331790558451608081018652848152928301829052938201849052426060929092018290526001929092556002919091556003919091556004556101e7806100996000396000f3fe608060405234801561001057600080fd5b50600436106100415760003560e01c80630ca05f9f146100465780636d4ce63c14610080578063893d20e8146100ae575b600080fd5b61006c6004803603602081101561005c57600080fd5b50356001600160a01b03166100d2565b604080519115158252519081900360200190f35b610088610190565b604080519485526020850193909352838301919091526060830152519081900360800190f35b6100b66101a2565b604080516001600160a01b039092168252519081900360200190f35b600080546001600160a01b03163314610128576040805162461bcd60e51b81526020600482015260136024820152721bdb9b1e481bdddb995c8818d85b8818d85b1b606a1b604482015290519081900360640190fd5b600080546001600160a01b038481166001600160a01b0319831681179093556040805191909216808252602082019390935281517f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90929181900390910190a150600192915050565b60015460025460035460045490919293565b6000546001600160a01b03169056fea2646970667358221220b4334ad5b80e15671f467e9d256c53f39856e0a6efa7993217dd4fc33a5051be64736f6c63430006030033"

// DeployOffer deploys a new Ethereum contract, binding an instance of Offer to it.
func DeployOffer(auth *bind.TransactOpts, backend bind.ContractBackend, capacity *big.Int, duration *big.Int, price *big.Int) (common.Address, *types.Transaction, *Offer, error) {
	parsed, err := abi.JSON(strings.NewReader(OfferABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(OfferBin), backend, capacity, duration, price)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Offer{OfferCaller: OfferCaller{contract: contract}, OfferTransactor: OfferTransactor{contract: contract}, OfferFilterer: OfferFilterer{contract: contract}}, nil
}

// Offer is an auto generated Go binding around an Ethereum contract.
type Offer struct {
	OfferCaller     // Read-only binding to the contract
	OfferTransactor // Write-only binding to the contract
	OfferFilterer   // Log filterer for contract events
}

// OfferCaller is an auto generated read-only Go binding around an Ethereum contract.
type OfferCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OfferTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OfferTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OfferFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OfferFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OfferSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OfferSession struct {
	Contract     *Offer            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OfferCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OfferCallerSession struct {
	Contract *OfferCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// OfferTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OfferTransactorSession struct {
	Contract     *OfferTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OfferRaw is an auto generated low-level Go binding around an Ethereum contract.
type OfferRaw struct {
	Contract *Offer // Generic contract binding to access the raw methods on
}

// OfferCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OfferCallerRaw struct {
	Contract *OfferCaller // Generic read-only contract binding to access the raw methods on
}

// OfferTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OfferTransactorRaw struct {
	Contract *OfferTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOffer creates a new instance of Offer, bound to a specific deployed contract.
func NewOffer(address common.Address, backend bind.ContractBackend) (*Offer, error) {
	contract, err := bindOffer(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Offer{OfferCaller: OfferCaller{contract: contract}, OfferTransactor: OfferTransactor{contract: contract}, OfferFilterer: OfferFilterer{contract: contract}}, nil
}

// NewOfferCaller creates a new read-only instance of Offer, bound to a specific deployed contract.
func NewOfferCaller(address common.Address, caller bind.ContractCaller) (*OfferCaller, error) {
	contract, err := bindOffer(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OfferCaller{contract: contract}, nil
}

// NewOfferTransactor creates a new write-only instance of Offer, bound to a specific deployed contract.
func NewOfferTransactor(address common.Address, transactor bind.ContractTransactor) (*OfferTransactor, error) {
	contract, err := bindOffer(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OfferTransactor{contract: contract}, nil
}

// NewOfferFilterer creates a new log filterer instance of Offer, bound to a specific deployed contract.
func NewOfferFilterer(address common.Address, filterer bind.ContractFilterer) (*OfferFilterer, error) {
	contract, err := bindOffer(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OfferFilterer{contract: contract}, nil
}

// bindOffer binds a generic wrapper to an already deployed contract.
func bindOffer(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(OfferABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Offer *OfferRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Offer.Contract.OfferCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Offer *OfferRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Offer.Contract.OfferTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Offer *OfferRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Offer.Contract.OfferTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Offer *OfferCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Offer.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Offer *OfferTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Offer.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Offer *OfferTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Offer.Contract.contract.Transact(opts, method, params...)
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() constant returns(uint256, uint256, uint256, uint256)
func (_Offer *OfferCaller) Get(opts *bind.CallOpts) (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	var (
		ret0 = new(*big.Int)
		ret1 = new(*big.Int)
		ret2 = new(*big.Int)
		ret3 = new(*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
		ret3,
	}
	err := _Offer.contract.Call(opts, out, "get")
	return *ret0, *ret1, *ret2, *ret3, err
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() constant returns(uint256, uint256, uint256, uint256)
func (_Offer *OfferSession) Get() (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	return _Offer.Contract.Get(&_Offer.CallOpts)
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() constant returns(uint256, uint256, uint256, uint256)
func (_Offer *OfferCallerSession) Get() (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	return _Offer.Contract.Get(&_Offer.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Offer *OfferCaller) GetOwner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Offer.contract.Call(opts, out, "getOwner")
	return *ret0, err
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Offer *OfferSession) GetOwner() (common.Address, error) {
	return _Offer.Contract.GetOwner(&_Offer.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Offer *OfferCallerSession) GetOwner() (common.Address, error) {
	return _Offer.Contract.GetOwner(&_Offer.CallOpts)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Offer *OfferTransactor) AlterOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Offer.contract.Transact(opts, "alterOwner", newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Offer *OfferSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Offer.Contract.AlterOwner(&_Offer.TransactOpts, newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Offer *OfferTransactorSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Offer.Contract.AlterOwner(&_Offer.TransactOpts, newOwner)
}

// OfferAlterOwnerIterator is returned from FilterAlterOwner and is used to iterate over the raw logs and unpacked data for AlterOwner events raised by the Offer contract.
type OfferAlterOwnerIterator struct {
	Event *OfferAlterOwner // Event containing the contract specifics and raw log

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
func (it *OfferAlterOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OfferAlterOwner)
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
		it.Event = new(OfferAlterOwner)
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
func (it *OfferAlterOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OfferAlterOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OfferAlterOwner represents a AlterOwner event raised by the Offer contract.
type OfferAlterOwner struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAlterOwner is a free log retrieval operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Offer *OfferFilterer) FilterAlterOwner(opts *bind.FilterOpts) (*OfferAlterOwnerIterator, error) {

	logs, sub, err := _Offer.contract.FilterLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return &OfferAlterOwnerIterator{contract: _Offer.contract, event: "AlterOwner", logs: logs, sub: sub}, nil
}

// WatchAlterOwner is a free log subscription operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Offer *OfferFilterer) WatchAlterOwner(opts *bind.WatchOpts, sink chan<- *OfferAlterOwner) (event.Subscription, error) {

	logs, sub, err := _Offer.contract.WatchLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OfferAlterOwner)
				if err := _Offer.contract.UnpackLog(event, "AlterOwner", log); err != nil {
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
func (_Offer *OfferFilterer) ParseAlterOwner(log types.Log) (*OfferAlterOwner, error) {
	event := new(OfferAlterOwner)
	if err := _Offer.contract.UnpackLog(event, "AlterOwner", log); err != nil {
		return nil, err
	}
	return event, nil
}
