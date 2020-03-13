// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package channel

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

// ChannelABI is the input ABI used to generate the binding from.
const ChannelABI = "[{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"timeout\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"closeChannel\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"ChannelTimeout\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"sign\",\"type\":\"bytes\"}],\"name\":\"CloseChannel\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getInfo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]"

// ChannelBin is the compiled bytecode used for deploying new contracts.
var ChannelBin = "0x608060405273e0f6a00fb23458731a5c73a02a36f1df2305090b600460006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506040516108b53803806108b58339818101604052604081101561007b57600080fd5b81019080805190602001909291908051906020019092919050505081600160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055504260028190555080600381905550505061077f806101366000396000f3fe6080604052600436106100385760003560e01c80632b7fa6be1461004457806339658245146101135780635a9b0b891461012a5761003f565b3661003f57005b600080fd5b6101116004803603606081101561005a57600080fd5b8101908080359060200190929190803590602001909291908035906020019064010000000081111561008b57600080fd5b82018360208201111561009d57600080fd5b803590602001918460018302840111640100000000831117156100bf57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f8201169050808301925050505050505091929192905050506101c2565b005b34801561011f57600080fd5b5061012861062e565b005b34801561013657600080fd5b5061013f6106e4565b604051808581526020018481526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200194505050505060405180910390f35b600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610285576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252600e8152602001807f696c6c6567616c2063616c6c657200000000000000000000000000000000000081525060200191505060405180910390fd5b60003083604051602001808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1660601b81526014018281526020019250505060405160208183030381529060405280519060200120905083811461035c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252600c8152602001807f696c6c6567616c2068617368000000000000000000000000000000000000000081525060200191505060405180910390fd5b6000600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166319045a2586856040518363ffffffff1660e01b81526004018083815260200180602001828103825283818151815260200191508051906020019080838360005b838110156103f35780820151818401526020810190506103d8565b50505050905090810190601f1680156104205780820380516001836020036101000a031916815260200191505b50935050505060206040518083038186803b15801561043e57600080fd5b505afa158015610452573d6000803e3d6000fd5b505050506040513d602081101561046857600080fd5b810190808051906020019092919050505090506000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161461053d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252600b8152602001807f696c6c6567616c2073696700000000000000000000000000000000000000000081525060200191505060405180910390fd5b600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc859081150290604051600060405180830381858888f193505050501580156105a5573d6000803e3d6000fd5b503373ffffffffffffffffffffffffffffffffffffffff167f01d42a9c1bb0e1a3464994bd2306368ef80e0dcf460c6123b5f7cbbcbf169fbb856040518082815260200191505060405180910390a26000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16ff5b426003546002540111156106aa576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252600e8152602001807f54696d65206973206e6f7420757000000000000000000000000000000000000081525060200191505060405180910390fd5b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16ff5b6000806000806002546003546000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff16600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681915080905093509350935093509091929356fea264697066735822122078a5eaa3c22714861337796bd1bc23bd3c8edfa4eae6b6b1825076a9f6403a9564736f6c63430006030033"

// DeployChannel deploys a new Ethereum contract, binding an instance of Channel to it.
func DeployChannel(auth *bind.TransactOpts, backend bind.ContractBackend, to common.Address, timeout *big.Int) (common.Address, *types.Transaction, *Channel, error) {
	parsed, err := abi.JSON(strings.NewReader(ChannelABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ChannelBin), backend, to, timeout)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Channel{ChannelCaller: ChannelCaller{contract: contract}, ChannelTransactor: ChannelTransactor{contract: contract}, ChannelFilterer: ChannelFilterer{contract: contract}}, nil
}

// Channel is an auto generated Go binding around an Ethereum contract.
type Channel struct {
	ChannelCaller     // Read-only binding to the contract
	ChannelTransactor // Write-only binding to the contract
	ChannelFilterer   // Log filterer for contract events
}

// ChannelCaller is an auto generated read-only Go binding around an Ethereum contract.
type ChannelCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ChannelTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ChannelTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ChannelFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ChannelFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ChannelSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ChannelSession struct {
	Contract     *Channel          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ChannelCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ChannelCallerSession struct {
	Contract *ChannelCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// ChannelTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ChannelTransactorSession struct {
	Contract     *ChannelTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// ChannelRaw is an auto generated low-level Go binding around an Ethereum contract.
type ChannelRaw struct {
	Contract *Channel // Generic contract binding to access the raw methods on
}

// ChannelCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ChannelCallerRaw struct {
	Contract *ChannelCaller // Generic read-only contract binding to access the raw methods on
}

// ChannelTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ChannelTransactorRaw struct {
	Contract *ChannelTransactor // Generic write-only contract binding to access the raw methods on
}

// NewChannel creates a new instance of Channel, bound to a specific deployed contract.
func NewChannel(address common.Address, backend bind.ContractBackend) (*Channel, error) {
	contract, err := bindChannel(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Channel{ChannelCaller: ChannelCaller{contract: contract}, ChannelTransactor: ChannelTransactor{contract: contract}, ChannelFilterer: ChannelFilterer{contract: contract}}, nil
}

// NewChannelCaller creates a new read-only instance of Channel, bound to a specific deployed contract.
func NewChannelCaller(address common.Address, caller bind.ContractCaller) (*ChannelCaller, error) {
	contract, err := bindChannel(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ChannelCaller{contract: contract}, nil
}

// NewChannelTransactor creates a new write-only instance of Channel, bound to a specific deployed contract.
func NewChannelTransactor(address common.Address, transactor bind.ContractTransactor) (*ChannelTransactor, error) {
	contract, err := bindChannel(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ChannelTransactor{contract: contract}, nil
}

// NewChannelFilterer creates a new log filterer instance of Channel, bound to a specific deployed contract.
func NewChannelFilterer(address common.Address, filterer bind.ContractFilterer) (*ChannelFilterer, error) {
	contract, err := bindChannel(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ChannelFilterer{contract: contract}, nil
}

// bindChannel binds a generic wrapper to an already deployed contract.
func bindChannel(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ChannelABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Channel *ChannelRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Channel.Contract.ChannelCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Channel *ChannelRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Channel.Contract.ChannelTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Channel *ChannelRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Channel.Contract.ChannelTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Channel *ChannelCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Channel.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Channel *ChannelTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Channel.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Channel *ChannelTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Channel.Contract.contract.Transact(opts, method, params...)
}

// GetInfo is a free data retrieval call binding the contract method 0x5a9b0b89.
//
// Solidity: function getInfo() constant returns(uint256, uint256, address, address)
func (_Channel *ChannelCaller) GetInfo(opts *bind.CallOpts) (*big.Int, *big.Int, common.Address, common.Address, error) {
	var (
		ret0 = new(*big.Int)
		ret1 = new(*big.Int)
		ret2 = new(common.Address)
		ret3 = new(common.Address)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
		ret3,
	}
	err := _Channel.contract.Call(opts, out, "getInfo")
	return *ret0, *ret1, *ret2, *ret3, err
}

// GetInfo is a free data retrieval call binding the contract method 0x5a9b0b89.
//
// Solidity: function getInfo() constant returns(uint256, uint256, address, address)
func (_Channel *ChannelSession) GetInfo() (*big.Int, *big.Int, common.Address, common.Address, error) {
	return _Channel.Contract.GetInfo(&_Channel.CallOpts)
}

// GetInfo is a free data retrieval call binding the contract method 0x5a9b0b89.
//
// Solidity: function getInfo() constant returns(uint256, uint256, address, address)
func (_Channel *ChannelCallerSession) GetInfo() (*big.Int, *big.Int, common.Address, common.Address, error) {
	return _Channel.Contract.GetInfo(&_Channel.CallOpts)
}

// ChannelTimeout is a paid mutator transaction binding the contract method 0x39658245.
//
// Solidity: function ChannelTimeout() returns()
func (_Channel *ChannelTransactor) ChannelTimeout(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Channel.contract.Transact(opts, "ChannelTimeout")
}

// ChannelTimeout is a paid mutator transaction binding the contract method 0x39658245.
//
// Solidity: function ChannelTimeout() returns()
func (_Channel *ChannelSession) ChannelTimeout() (*types.Transaction, error) {
	return _Channel.Contract.ChannelTimeout(&_Channel.TransactOpts)
}

// ChannelTimeout is a paid mutator transaction binding the contract method 0x39658245.
//
// Solidity: function ChannelTimeout() returns()
func (_Channel *ChannelTransactorSession) ChannelTimeout() (*types.Transaction, error) {
	return _Channel.Contract.ChannelTimeout(&_Channel.TransactOpts)
}

// CloseChannel is a paid mutator transaction binding the contract method 0x2b7fa6be.
//
// Solidity: function CloseChannel(bytes32 hash, uint256 value, bytes sign) returns()
func (_Channel *ChannelTransactor) CloseChannel(opts *bind.TransactOpts, hash [32]byte, value *big.Int, sign []byte) (*types.Transaction, error) {
	return _Channel.contract.Transact(opts, "CloseChannel", hash, value, sign)
}

// CloseChannel is a paid mutator transaction binding the contract method 0x2b7fa6be.
//
// Solidity: function CloseChannel(bytes32 hash, uint256 value, bytes sign) returns()
func (_Channel *ChannelSession) CloseChannel(hash [32]byte, value *big.Int, sign []byte) (*types.Transaction, error) {
	return _Channel.Contract.CloseChannel(&_Channel.TransactOpts, hash, value, sign)
}

// CloseChannel is a paid mutator transaction binding the contract method 0x2b7fa6be.
//
// Solidity: function CloseChannel(bytes32 hash, uint256 value, bytes sign) returns()
func (_Channel *ChannelTransactorSession) CloseChannel(hash [32]byte, value *big.Int, sign []byte) (*types.Transaction, error) {
	return _Channel.Contract.CloseChannel(&_Channel.TransactOpts, hash, value, sign)
}

// ChannelCloseChannelIterator is returned from FilterCloseChannel and is used to iterate over the raw logs and unpacked data for CloseChannel events raised by the Channel contract.
type ChannelCloseChannelIterator struct {
	Event *ChannelCloseChannel // Event containing the contract specifics and raw log

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
func (it *ChannelCloseChannelIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelCloseChannel)
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
		it.Event = new(ChannelCloseChannel)
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
func (it *ChannelCloseChannelIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelCloseChannelIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelCloseChannel represents a CloseChannel event raised by the Channel contract.
type ChannelCloseChannel struct {
	From  common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterCloseChannel is a free log retrieval operation binding the contract event 0x01d42a9c1bb0e1a3464994bd2306368ef80e0dcf460c6123b5f7cbbcbf169fbb.
//
// Solidity: event closeChannel(address indexed from, uint256 value)
func (_Channel *ChannelFilterer) FilterCloseChannel(opts *bind.FilterOpts, from []common.Address) (*ChannelCloseChannelIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _Channel.contract.FilterLogs(opts, "closeChannel", fromRule)
	if err != nil {
		return nil, err
	}
	return &ChannelCloseChannelIterator{contract: _Channel.contract, event: "closeChannel", logs: logs, sub: sub}, nil
}

// WatchCloseChannel is a free log subscription operation binding the contract event 0x01d42a9c1bb0e1a3464994bd2306368ef80e0dcf460c6123b5f7cbbcbf169fbb.
//
// Solidity: event closeChannel(address indexed from, uint256 value)
func (_Channel *ChannelFilterer) WatchCloseChannel(opts *bind.WatchOpts, sink chan<- *ChannelCloseChannel, from []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _Channel.contract.WatchLogs(opts, "closeChannel", fromRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelCloseChannel)
				if err := _Channel.contract.UnpackLog(event, "closeChannel", log); err != nil {
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

// ParseCloseChannel is a log parse operation binding the contract event 0x01d42a9c1bb0e1a3464994bd2306368ef80e0dcf460c6123b5f7cbbcbf169fbb.
//
// Solidity: event closeChannel(address indexed from, uint256 value)
func (_Channel *ChannelFilterer) ParseCloseChannel(log types.Log) (*ChannelCloseChannel, error) {
	event := new(ChannelCloseChannel)
	if err := _Channel.contract.UnpackLog(event, "closeChannel", log); err != nil {
		return nil, err
	}
	return event, nil
}
