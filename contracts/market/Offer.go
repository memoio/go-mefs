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
const OfferABI = "[{\"inputs\":[{\"name\":\"capacity\",\"type\":\"uint256\"},{\"name\":\"duration\",\"type\":\"uint256\"},{\"name\":\"price\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\",\"signature\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"data\",\"type\":\"string\"}],\"name\":\"Error\",\"type\":\"event\",\"signature\":\"0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"data\",\"type\":\"uint256\"}],\"name\":\"LogInt\",\"type\":\"event\",\"signature\":\"0xc8fa9a7021af252bc69defe2b981f7bd7858defe2a87641768fefdb8a03a07cd\"},{\"constant\":true,\"inputs\":[],\"name\":\"get\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\",\"signature\":\"0x6d4ce63c\"}]"

// OfferBin is the compiled bytecode used for deploying new contracts.
const OfferBin = `0x608060405234801561001057600080fd5b5060405160608061016b8339810180604052606081101561003057600080fd5b81019080805190602001909291908051906020019092919080519060200190929190505050606060405190810160405280848152602001838152602001828152506000808201518160000155602082015181600101556040820151816002015590505050505060c7806100a46000396000f3fe608060405260043610603f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680636d4ce63c146044575b600080fd5b348015604f57600080fd5b506056607a565b60405180848152602001838152602001828152602001935050505060405180910390f35b6000806000806000015460006001015460006002015492509250925090919256fea165627a7a723058206901c679355bcde89b60867d4d813e437395a1f68a2a51f5207c63ac60cd39ae0029`

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
// Solidity: function get() constant returns(uint256, uint256, uint256)
func (_Offer *OfferCaller) Get(opts *bind.CallOpts) (*big.Int, *big.Int, *big.Int, error) {
	var (
		ret0 = new(*big.Int)
		ret1 = new(*big.Int)
		ret2 = new(*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
	}
	err := _Offer.contract.Call(opts, out, "get")
	return *ret0, *ret1, *ret2, err
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() constant returns(uint256, uint256, uint256)
func (_Offer *OfferSession) Get() (*big.Int, *big.Int, *big.Int, error) {
	return _Offer.Contract.Get(&_Offer.CallOpts)
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() constant returns(uint256, uint256, uint256)
func (_Offer *OfferCallerSession) Get() (*big.Int, *big.Int, *big.Int, error) {
	return _Offer.Contract.Get(&_Offer.CallOpts)
}

// OfferErrorIterator is returned from FilterError and is used to iterate over the raw logs and unpacked data for Error events raised by the Offer contract.
type OfferErrorIterator struct {
	Event *OfferError // Event containing the contract specifics and raw log

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
func (it *OfferErrorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OfferError)
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
		it.Event = new(OfferError)
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
func (it *OfferErrorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OfferErrorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OfferError represents a Error event raised by the Offer contract.
type OfferError struct {
	Data string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterError is a free log retrieval operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Offer *OfferFilterer) FilterError(opts *bind.FilterOpts) (*OfferErrorIterator, error) {

	logs, sub, err := _Offer.contract.FilterLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return &OfferErrorIterator{contract: _Offer.contract, event: "Error", logs: logs, sub: sub}, nil
}

// WatchError is a free log subscription operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Offer *OfferFilterer) WatchError(opts *bind.WatchOpts, sink chan<- *OfferError) (event.Subscription, error) {

	logs, sub, err := _Offer.contract.WatchLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OfferError)
				if err := _Offer.contract.UnpackLog(event, "Error", log); err != nil {
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

// OfferLogIntIterator is returned from FilterLogInt and is used to iterate over the raw logs and unpacked data for LogInt events raised by the Offer contract.
type OfferLogIntIterator struct {
	Event *OfferLogInt // Event containing the contract specifics and raw log

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
func (it *OfferLogIntIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OfferLogInt)
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
		it.Event = new(OfferLogInt)
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
func (it *OfferLogIntIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OfferLogIntIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OfferLogInt represents a LogInt event raised by the Offer contract.
type OfferLogInt struct {
	Data *big.Int
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterLogInt is a free log retrieval operation binding the contract event 0xc8fa9a7021af252bc69defe2b981f7bd7858defe2a87641768fefdb8a03a07cd.
//
// Solidity: event LogInt(uint256 indexed data)
func (_Offer *OfferFilterer) FilterLogInt(opts *bind.FilterOpts, data []*big.Int) (*OfferLogIntIterator, error) {

	var dataRule []interface{}
	for _, dataItem := range data {
		dataRule = append(dataRule, dataItem)
	}

	logs, sub, err := _Offer.contract.FilterLogs(opts, "LogInt", dataRule)
	if err != nil {
		return nil, err
	}
	return &OfferLogIntIterator{contract: _Offer.contract, event: "LogInt", logs: logs, sub: sub}, nil
}

// WatchLogInt is a free log subscription operation binding the contract event 0xc8fa9a7021af252bc69defe2b981f7bd7858defe2a87641768fefdb8a03a07cd.
//
// Solidity: event LogInt(uint256 indexed data)
func (_Offer *OfferFilterer) WatchLogInt(opts *bind.WatchOpts, sink chan<- *OfferLogInt, data []*big.Int) (event.Subscription, error) {

	var dataRule []interface{}
	for _, dataItem := range data {
		dataRule = append(dataRule, dataItem)
	}

	logs, sub, err := _Offer.contract.WatchLogs(opts, "LogInt", dataRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OfferLogInt)
				if err := _Offer.contract.UnpackLog(event, "LogInt", log); err != nil {
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
