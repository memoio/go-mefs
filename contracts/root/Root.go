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
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// RootABI is the input ABI used to generate the binding from.
const RootABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_query\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"int64\",\"name\":\"key\",\"type\":\"int64\"}],\"name\":\"AddRoot\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAllKey\",\"outputs\":[{\"internalType\":\"int64[]\",\"name\":\"\",\"type\":\"int64[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getLatest\",\"outputs\":[{\"internalType\":\"int64\",\"name\":\"\",\"type\":\"int64\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"int64\",\"name\":\"key\",\"type\":\"int64\"}],\"name\":\"getRoot\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"queryAddr\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"int64\",\"name\":\"key\",\"type\":\"int64\"},{\"internalType\":\"bytes32\",\"name\":\"value\",\"type\":\"bytes32\"}],\"name\":\"setRoot\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// RootBin is the compiled bytecode used for deploying new contracts.
var RootBin = "0x6080604052738391984e2f1cc8f6b916f566c1d0a6bb8a15c73a600460006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555034801561006557600080fd5b506040516109933803806109938339818101604052602081101561008857600080fd5b8101908080519060200190929190505050336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506000600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16636778d3cb6040518163ffffffff1660e01b815260040160206040518083038186803b15801561014357600080fd5b505afa158015610157573d6000803e3d6000fd5b505050506040513d602081101561016d57600080fd5b8101908080519060200190929190505050905080156101f4576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260108152602001807f6465706c6f792069732062616e6e65640000000000000000000000000000000081525060200191505060405180910390fd5b81600360006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550505061074d806102466000396000f3fe608060405234801561001057600080fd5b506004361061007d5760003560e01c80637477f8501161005b5780637477f8501461016f578063893d20e8146101c0578063beac5b63146101f4578063c36af460146102395761007d565b80630ca05f9f146100825780632a9a8d8d146100dc5780636b5b433514610110575b600080fd5b6100c46004803603602081101561009857600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610261565b60405180821515815260200191505060405180910390f35b6100e4610400565b604051808273ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b610118610426565b6040518080602001828103825283818151815260200191508051906020019060200280838360005b8381101561015b578082015181840152602081019050610140565b505050509050019250505060405180910390f35b6101a86004803603604081101561018557600080fd5b81019080803560070b9060200190929190803590602001909291905050506104a4565b60405180821515815260200191505060405180910390f35b6101c8610648565b604051808273ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6102236004803603602081101561020a57600080fd5b81019080803560070b9060200190929190505050610671565b6040518082815260200191505060405180910390f35b610241610694565b604051808360070b81526020018281526020019250505060405180910390f35b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610325576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050826000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e908184604051808373ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a16001915050919050565b600360009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6060600280548060200260200160405190810160405280929190818152602001828054801561049a57602002820191906000526020600020906000905b82829054906101000a900460070b60070b815260200190600801906020826007010492830192600103820291508084116104635790505b5050505050905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610568576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b6000801b600160008560070b60070b81526020019081526020016000205414156105e65760028390806001815401808255809150506001900390600052602060002090600491828204019190066008029091909190916101000a81548167ffffffffffffffff021916908360070b67ffffffffffffffff1602179055505b81600160008560070b60070b8152602001908152602001600020819055507f352d2b8fd5bd4233af9478ba7ca7fe3da4d8a0438736005bc110bef2cab7443a83604051808260070b815260200191505060405180910390a16001905092915050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b6000600160008360070b60070b8152602001908152602001600020549050919050565b6000806000600280549050905060008114156106ba5760008060001b9250925050610713565b6000600260018303815481106106cc57fe5b90600052602060002090600491828204019190066008029054906101000a900460070b905080600160008360070b60070b8152602001908152602001600020549350935050505b909156fea2646970667358221220438a57fc165bfa15bcf269e4c850f3c9b7adc63b6def3b9b1dee08fe17b91ece64736f6c63430007030033"

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
// Solidity: function getAllKey() view returns(int64[])
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
// Solidity: function getAllKey() view returns(int64[])
func (_Root *RootSession) GetAllKey() ([]int64, error) {
	return _Root.Contract.GetAllKey(&_Root.CallOpts)
}

// GetAllKey is a free data retrieval call binding the contract method 0x6b5b4335.
//
// Solidity: function getAllKey() view returns(int64[])
func (_Root *RootCallerSession) GetAllKey() ([]int64, error) {
	return _Root.Contract.GetAllKey(&_Root.CallOpts)
}

// GetLatest is a free data retrieval call binding the contract method 0xc36af460.
//
// Solidity: function getLatest() view returns(int64, bytes32)
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
// Solidity: function getLatest() view returns(int64, bytes32)
func (_Root *RootSession) GetLatest() (int64, [32]byte, error) {
	return _Root.Contract.GetLatest(&_Root.CallOpts)
}

// GetLatest is a free data retrieval call binding the contract method 0xc36af460.
//
// Solidity: function getLatest() view returns(int64, bytes32)
func (_Root *RootCallerSession) GetLatest() (int64, [32]byte, error) {
	return _Root.Contract.GetLatest(&_Root.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
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
// Solidity: function getOwner() view returns(address)
func (_Root *RootSession) GetOwner() (common.Address, error) {
	return _Root.Contract.GetOwner(&_Root.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_Root *RootCallerSession) GetOwner() (common.Address, error) {
	return _Root.Contract.GetOwner(&_Root.CallOpts)
}

// GetRoot is a free data retrieval call binding the contract method 0xbeac5b63.
//
// Solidity: function getRoot(int64 key) view returns(bytes32)
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
// Solidity: function getRoot(int64 key) view returns(bytes32)
func (_Root *RootSession) GetRoot(key int64) ([32]byte, error) {
	return _Root.Contract.GetRoot(&_Root.CallOpts, key)
}

// GetRoot is a free data retrieval call binding the contract method 0xbeac5b63.
//
// Solidity: function getRoot(int64 key) view returns(bytes32)
func (_Root *RootCallerSession) GetRoot(key int64) ([32]byte, error) {
	return _Root.Contract.GetRoot(&_Root.CallOpts, key)
}

// QueryAddr is a free data retrieval call binding the contract method 0x2a9a8d8d.
//
// Solidity: function queryAddr() view returns(address)
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
// Solidity: function queryAddr() view returns(address)
func (_Root *RootSession) QueryAddr() (common.Address, error) {
	return _Root.Contract.QueryAddr(&_Root.CallOpts)
}

// QueryAddr is a free data retrieval call binding the contract method 0x2a9a8d8d.
//
// Solidity: function queryAddr() view returns(address)
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
