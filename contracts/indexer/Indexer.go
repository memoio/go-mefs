// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package indexer

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

// IndexerABI is the input ABI used to generate the binding from.
const IndexerABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"key\",\"type\":\"string\"},{\"name\":\"resolver\",\"type\":\"address\"}],\"name\":\"add\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"key\",\"type\":\"string\"}],\"name\":\"get\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"key\",\"type\":\"string\"},{\"name\":\"resolver\",\"type\":\"address\"}],\"name\":\"alterResolver\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"key\",\"type\":\"string\"},{\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"key\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"resolver\",\"type\":\"address\"}],\"name\":\"Add\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"key\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"form\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterResolver\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"key\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"form\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"data\",\"type\":\"string\"}],\"name\":\"Error\",\"type\":\"event\"}]"

// IndexerBin is the compiled bytecode used for deploying new contracts.
const IndexerBin = `0x608060405234801561001057600080fd5b50610ba8806100206000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c80632bffc7ed14610051578063693ec85e14610114578063cbdc3fe1146101de578063f60b53e21461028d575b600080fd5b6101006004803603604081101561006757600080fd5b810190602081018135600160201b81111561008157600080fd5b82018360208201111561009357600080fd5b803590602001918460018302840111600160201b831117156100b457600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250929550505090356001600160a01b0316915061033c9050565b604080519115158252519081900360200190f35b6101b86004803603602081101561012a57600080fd5b810190602081018135600160201b81111561014457600080fd5b82018360208201111561015657600080fd5b803590602001918460018302840111600160201b8311171561017757600080fd5b91908080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152509295506105d1945050505050565b604080516001600160a01b03938416815291909216602082015281519081900390910190f35b610100600480360360408110156101f457600080fd5b810190602081018135600160201b81111561020e57600080fd5b82018360208201111561022057600080fd5b803590602001918460018302840111600160201b8311171561024157600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250929550505090356001600160a01b031691506106b79050565b610100600480360360408110156102a357600080fd5b810190602081018135600160201b8111156102bd57600080fd5b8201836020820111156102cf57600080fd5b803590602001918460018302840111600160201b831117156102f057600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250929550505090356001600160a01b031691506109419050565b6000806001600160a01b03166000846040518082805190602001908083835b6020831061037a5780518252601f19909201916020918201910161035b565b51815160209384036101000a60001901801990921691161790529201948552506040519384900301909220546001600160a01b0316929092149150610421905057604080516020808252808201527fe5b7b2e69c89e6ada4e5908de7a7b0e5afb9e5ba94e79a847265736f6c7665728183015290517f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa9181900360600190a15060006105cb565b336000846040518082805190602001908083835b602083106104545780518252601f199092019160209182019101610435565b51815160209384036101000a6000190180199092169116179052920194855250604051938490038101842080546001600160a01b0319166001600160a01b039690961695909517909455505084518492600092879290918291908401908083835b602083106104d45780518252601f1990920191602091820191016104b5565b51815160209384036101000a600019018019909216911617905292019485525060408051948590038201852060010180546001600160a01b0319166001600160a01b039788161790553385830181905295881690850152606080855288519085015287517fec689a3871c35587e4800f14216f987ee744b924aff21741edc2e167e2dd43e8958995909450889350918291608083019187019080838360005b8381101561058b578181015183820152602001610573565b50505050905090810190601f1680156105b85780820380516001836020036101000a031916815260200191505b5094505050505060405180910390a15060015b92915050565b6000806000836040518082805190602001908083835b602083106106065780518252601f1990920191602091820191016105e7565b51815160209384036101000a600019018019909216911617905292019485525060405193849003810184205487516001600160a01b039091169460009450889350918291908401908083835b602083106106715780518252601f199092019160209182019101610652565b51815160209384036101000a600019018019909216911617905292019485525060405193849003019092206001015492966001600160a01b039093169550919350505050565b6000336001600160a01b03166000846040518082805190602001908083835b602083106106f55780518252601f1990920191602091820191016106d6565b51815160209384036101000a60001901801990921691161790529201948552506040519384900301909220546001600160a01b031692909214915061079e9050576040805160208082526018908201527fe4bda0e4b88de698afe6ada46b6579e79a84e4b8bbe4baba00000000000000008183015290517f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa9181900360600190a15060006105cb565b600080846040518082805190602001908083835b602083106107d15780518252601f1990920191602091820191016107b2565b51815160209384036101000a600019018019909216911617905292019485525060405193849003810184206001015488516001600160a01b039091169550879460009450899350918291908401908083835b602083106108425780518252601f199092019160209182019101610823565b51815160209384036101000a600019018019909216911617905292019485525060408051948590038201852060010180546001600160a01b0319166001600160a01b0397881617905586861685830152948816948401949094525050606080825286519082015285517f0a7047ba8be4d874e67aebc953a70ff6db03a81782549290ac646e0738ddfc04928792859288928291608083019187019080838360005b838110156108fb5781810151838201526020016108e3565b50505050905090810190601f1680156109285780820380516001836020036101000a031916815260200191505b5094505050505060405180910390a15060019392505050565b6000336001600160a01b03166000846040518082805190602001908083835b6020831061097f5780518252601f199092019160209182019101610960565b51815160209384036101000a60001901801990921691161790529201948552506040519384900301909220546001600160a01b0316929092149150610a289050576040805160208082526018908201527fe4bda0e4b88de698afe6ada46b6579e79a84e4b8bbe4baba00000000000000008183015290517f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa9181900360600190a15060006105cb565b600080846040518082805190602001908083835b60208310610a5b5780518252601f199092019160209182019101610a3c565b51815160209384036101000a600019018019909216911617905292019485525060405193849003810184205488516001600160a01b039091169550879460009450899350918291908401908083835b60208310610ac95780518252601f199092019160209182019101610aaa565b51815160209384036101000a600019018019909216911617905292019485525060408051948590038201852080546001600160a01b0319166001600160a01b0397881617905586861685830152948816948401949094525050606080825286519082015285517f46bd035a76a8302bb74520f9226b59925d8186784298f88ad636a4ea46b85b219287928592889282916080830191870190808383600083156108fb5781810151838201526020016108e356fea165627a7a7230582014758ba178ccbaad7bc92bfd70ac7c3aa2abe53266c4389f6f8a6c394c77d7280029`

// DeployIndexer deploys a new Ethereum contract, binding an instance of Indexer to it.
func DeployIndexer(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Indexer, error) {
	parsed, err := abi.JSON(strings.NewReader(IndexerABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(IndexerBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Indexer{IndexerCaller: IndexerCaller{contract: contract}, IndexerTransactor: IndexerTransactor{contract: contract}, IndexerFilterer: IndexerFilterer{contract: contract}}, nil
}

// Indexer is an auto generated Go binding around an Ethereum contract.
type Indexer struct {
	IndexerCaller     // Read-only binding to the contract
	IndexerTransactor // Write-only binding to the contract
	IndexerFilterer   // Log filterer for contract events
}

// IndexerCaller is an auto generated read-only Go binding around an Ethereum contract.
type IndexerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IndexerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IndexerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IndexerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IndexerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IndexerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IndexerSession struct {
	Contract     *Indexer          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IndexerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IndexerCallerSession struct {
	Contract *IndexerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// IndexerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IndexerTransactorSession struct {
	Contract     *IndexerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// IndexerRaw is an auto generated low-level Go binding around an Ethereum contract.
type IndexerRaw struct {
	Contract *Indexer // Generic contract binding to access the raw methods on
}

// IndexerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IndexerCallerRaw struct {
	Contract *IndexerCaller // Generic read-only contract binding to access the raw methods on
}

// IndexerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IndexerTransactorRaw struct {
	Contract *IndexerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIndexer creates a new instance of Indexer, bound to a specific deployed contract.
func NewIndexer(address common.Address, backend bind.ContractBackend) (*Indexer, error) {
	contract, err := bindIndexer(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Indexer{IndexerCaller: IndexerCaller{contract: contract}, IndexerTransactor: IndexerTransactor{contract: contract}, IndexerFilterer: IndexerFilterer{contract: contract}}, nil
}

// NewIndexerCaller creates a new read-only instance of Indexer, bound to a specific deployed contract.
func NewIndexerCaller(address common.Address, caller bind.ContractCaller) (*IndexerCaller, error) {
	contract, err := bindIndexer(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IndexerCaller{contract: contract}, nil
}

// NewIndexerTransactor creates a new write-only instance of Indexer, bound to a specific deployed contract.
func NewIndexerTransactor(address common.Address, transactor bind.ContractTransactor) (*IndexerTransactor, error) {
	contract, err := bindIndexer(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IndexerTransactor{contract: contract}, nil
}

// NewIndexerFilterer creates a new log filterer instance of Indexer, bound to a specific deployed contract.
func NewIndexerFilterer(address common.Address, filterer bind.ContractFilterer) (*IndexerFilterer, error) {
	contract, err := bindIndexer(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IndexerFilterer{contract: contract}, nil
}

// bindIndexer binds a generic wrapper to an already deployed contract.
func bindIndexer(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IndexerABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Indexer *IndexerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Indexer.Contract.IndexerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Indexer *IndexerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Indexer.Contract.IndexerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Indexer *IndexerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Indexer.Contract.IndexerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Indexer *IndexerCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Indexer.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Indexer *IndexerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Indexer.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Indexer *IndexerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Indexer.Contract.contract.Transact(opts, method, params...)
}

// Get is a free data retrieval call binding the contract method 0x693ec85e.
//
// Solidity: function get(string key) constant returns(address, address)
func (_Indexer *IndexerCaller) Get(opts *bind.CallOpts, key string) (common.Address, common.Address, error) {
	var (
		ret0 = new(common.Address)
		ret1 = new(common.Address)
	)
	out := &[]interface{}{
		ret0,
		ret1,
	}
	err := _Indexer.contract.Call(opts, out, "get", key)
	return *ret0, *ret1, err
}

// Get is a free data retrieval call binding the contract method 0x693ec85e.
//
// Solidity: function get(string key) constant returns(address, address)
func (_Indexer *IndexerSession) Get(key string) (common.Address, common.Address, error) {
	return _Indexer.Contract.Get(&_Indexer.CallOpts, key)
}

// Get is a free data retrieval call binding the contract method 0x693ec85e.
//
// Solidity: function get(string key) constant returns(address, address)
func (_Indexer *IndexerCallerSession) Get(key string) (common.Address, common.Address, error) {
	return _Indexer.Contract.Get(&_Indexer.CallOpts, key)
}

// Add is a paid mutator transaction binding the contract method 0x2bffc7ed.
//
// Solidity: function add(string key, address resolver) returns(bool)
func (_Indexer *IndexerTransactor) Add(opts *bind.TransactOpts, key string, resolver common.Address) (*types.Transaction, error) {
	return _Indexer.contract.Transact(opts, "add", key, resolver)
}

// Add is a paid mutator transaction binding the contract method 0x2bffc7ed.
//
// Solidity: function add(string key, address resolver) returns(bool)
func (_Indexer *IndexerSession) Add(key string, resolver common.Address) (*types.Transaction, error) {
	return _Indexer.Contract.Add(&_Indexer.TransactOpts, key, resolver)
}

// Add is a paid mutator transaction binding the contract method 0x2bffc7ed.
//
// Solidity: function add(string key, address resolver) returns(bool)
func (_Indexer *IndexerTransactorSession) Add(key string, resolver common.Address) (*types.Transaction, error) {
	return _Indexer.Contract.Add(&_Indexer.TransactOpts, key, resolver)
}

// AlterOwner is a paid mutator transaction binding the contract method 0xf60b53e2.
//
// Solidity: function alterOwner(string key, address owner) returns(bool)
func (_Indexer *IndexerTransactor) AlterOwner(opts *bind.TransactOpts, key string, owner common.Address) (*types.Transaction, error) {
	return _Indexer.contract.Transact(opts, "alterOwner", key, owner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0xf60b53e2.
//
// Solidity: function alterOwner(string key, address owner) returns(bool)
func (_Indexer *IndexerSession) AlterOwner(key string, owner common.Address) (*types.Transaction, error) {
	return _Indexer.Contract.AlterOwner(&_Indexer.TransactOpts, key, owner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0xf60b53e2.
//
// Solidity: function alterOwner(string key, address owner) returns(bool)
func (_Indexer *IndexerTransactorSession) AlterOwner(key string, owner common.Address) (*types.Transaction, error) {
	return _Indexer.Contract.AlterOwner(&_Indexer.TransactOpts, key, owner)
}

// AlterResolver is a paid mutator transaction binding the contract method 0xcbdc3fe1.
//
// Solidity: function alterResolver(string key, address resolver) returns(bool)
func (_Indexer *IndexerTransactor) AlterResolver(opts *bind.TransactOpts, key string, resolver common.Address) (*types.Transaction, error) {
	return _Indexer.contract.Transact(opts, "alterResolver", key, resolver)
}

// AlterResolver is a paid mutator transaction binding the contract method 0xcbdc3fe1.
//
// Solidity: function alterResolver(string key, address resolver) returns(bool)
func (_Indexer *IndexerSession) AlterResolver(key string, resolver common.Address) (*types.Transaction, error) {
	return _Indexer.Contract.AlterResolver(&_Indexer.TransactOpts, key, resolver)
}

// AlterResolver is a paid mutator transaction binding the contract method 0xcbdc3fe1.
//
// Solidity: function alterResolver(string key, address resolver) returns(bool)
func (_Indexer *IndexerTransactorSession) AlterResolver(key string, resolver common.Address) (*types.Transaction, error) {
	return _Indexer.Contract.AlterResolver(&_Indexer.TransactOpts, key, resolver)
}

// IndexerAddIterator is returned from FilterAdd and is used to iterate over the raw logs and unpacked data for Add events raised by the Indexer contract.
type IndexerAddIterator struct {
	Event *IndexerAdd // Event containing the contract specifics and raw log

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
func (it *IndexerAddIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IndexerAdd)
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
		it.Event = new(IndexerAdd)
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
func (it *IndexerAddIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IndexerAddIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IndexerAdd represents a Add event raised by the Indexer contract.
type IndexerAdd struct {
	Key      string
	Owner    common.Address
	Resolver common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterAdd is a free log retrieval operation binding the contract event 0xec689a3871c35587e4800f14216f987ee744b924aff21741edc2e167e2dd43e8.
//
// Solidity: event Add(string key, address owner, address resolver)
func (_Indexer *IndexerFilterer) FilterAdd(opts *bind.FilterOpts) (*IndexerAddIterator, error) {

	logs, sub, err := _Indexer.contract.FilterLogs(opts, "Add")
	if err != nil {
		return nil, err
	}
	return &IndexerAddIterator{contract: _Indexer.contract, event: "Add", logs: logs, sub: sub}, nil
}

// WatchAdd is a free log subscription operation binding the contract event 0xec689a3871c35587e4800f14216f987ee744b924aff21741edc2e167e2dd43e8.
//
// Solidity: event Add(string key, address owner, address resolver)
func (_Indexer *IndexerFilterer) WatchAdd(opts *bind.WatchOpts, sink chan<- *IndexerAdd) (event.Subscription, error) {

	logs, sub, err := _Indexer.contract.WatchLogs(opts, "Add")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IndexerAdd)
				if err := _Indexer.contract.UnpackLog(event, "Add", log); err != nil {
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

// IndexerAlterOwnerIterator is returned from FilterAlterOwner and is used to iterate over the raw logs and unpacked data for AlterOwner events raised by the Indexer contract.
type IndexerAlterOwnerIterator struct {
	Event *IndexerAlterOwner // Event containing the contract specifics and raw log

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
func (it *IndexerAlterOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IndexerAlterOwner)
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
		it.Event = new(IndexerAlterOwner)
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
func (it *IndexerAlterOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IndexerAlterOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IndexerAlterOwner represents a AlterOwner event raised by the Indexer contract.
type IndexerAlterOwner struct {
	Key  string
	Form common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAlterOwner is a free log retrieval operation binding the contract event 0x46bd035a76a8302bb74520f9226b59925d8186784298f88ad636a4ea46b85b21.
//
// Solidity: event AlterOwner(string key, address form, address to)
func (_Indexer *IndexerFilterer) FilterAlterOwner(opts *bind.FilterOpts) (*IndexerAlterOwnerIterator, error) {

	logs, sub, err := _Indexer.contract.FilterLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return &IndexerAlterOwnerIterator{contract: _Indexer.contract, event: "AlterOwner", logs: logs, sub: sub}, nil
}

// WatchAlterOwner is a free log subscription operation binding the contract event 0x46bd035a76a8302bb74520f9226b59925d8186784298f88ad636a4ea46b85b21.
//
// Solidity: event AlterOwner(string key, address form, address to)
func (_Indexer *IndexerFilterer) WatchAlterOwner(opts *bind.WatchOpts, sink chan<- *IndexerAlterOwner) (event.Subscription, error) {

	logs, sub, err := _Indexer.contract.WatchLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IndexerAlterOwner)
				if err := _Indexer.contract.UnpackLog(event, "AlterOwner", log); err != nil {
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

// IndexerAlterResolverIterator is returned from FilterAlterResolver and is used to iterate over the raw logs and unpacked data for AlterResolver events raised by the Indexer contract.
type IndexerAlterResolverIterator struct {
	Event *IndexerAlterResolver // Event containing the contract specifics and raw log

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
func (it *IndexerAlterResolverIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IndexerAlterResolver)
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
		it.Event = new(IndexerAlterResolver)
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
func (it *IndexerAlterResolverIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IndexerAlterResolverIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IndexerAlterResolver represents a AlterResolver event raised by the Indexer contract.
type IndexerAlterResolver struct {
	Key  string
	Form common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAlterResolver is a free log retrieval operation binding the contract event 0x0a7047ba8be4d874e67aebc953a70ff6db03a81782549290ac646e0738ddfc04.
//
// Solidity: event AlterResolver(string key, address form, address to)
func (_Indexer *IndexerFilterer) FilterAlterResolver(opts *bind.FilterOpts) (*IndexerAlterResolverIterator, error) {

	logs, sub, err := _Indexer.contract.FilterLogs(opts, "AlterResolver")
	if err != nil {
		return nil, err
	}
	return &IndexerAlterResolverIterator{contract: _Indexer.contract, event: "AlterResolver", logs: logs, sub: sub}, nil
}

// WatchAlterResolver is a free log subscription operation binding the contract event 0x0a7047ba8be4d874e67aebc953a70ff6db03a81782549290ac646e0738ddfc04.
//
// Solidity: event AlterResolver(string key, address form, address to)
func (_Indexer *IndexerFilterer) WatchAlterResolver(opts *bind.WatchOpts, sink chan<- *IndexerAlterResolver) (event.Subscription, error) {

	logs, sub, err := _Indexer.contract.WatchLogs(opts, "AlterResolver")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IndexerAlterResolver)
				if err := _Indexer.contract.UnpackLog(event, "AlterResolver", log); err != nil {
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

// IndexerErrorIterator is returned from FilterError and is used to iterate over the raw logs and unpacked data for Error events raised by the Indexer contract.
type IndexerErrorIterator struct {
	Event *IndexerError // Event containing the contract specifics and raw log

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
func (it *IndexerErrorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IndexerError)
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
		it.Event = new(IndexerError)
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
func (it *IndexerErrorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IndexerErrorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IndexerError represents a Error event raised by the Indexer contract.
type IndexerError struct {
	Data string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterError is a free log retrieval operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Indexer *IndexerFilterer) FilterError(opts *bind.FilterOpts) (*IndexerErrorIterator, error) {

	logs, sub, err := _Indexer.contract.FilterLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return &IndexerErrorIterator{contract: _Indexer.contract, event: "Error", logs: logs, sub: sub}, nil
}

// WatchError is a free log subscription operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Indexer *IndexerFilterer) WatchError(opts *bind.WatchOpts, sink chan<- *IndexerError) (event.Subscription, error) {

	logs, sub, err := _Indexer.contract.WatchLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IndexerError)
				if err := _Indexer.contract.UnpackLog(event, "Error", log); err != nil {
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
