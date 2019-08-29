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
const ChannelABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"channelRecipient\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"channelSender\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"startDate\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"},{\"name\":\"value\",\"type\":\"uint256\"},{\"name\":\"sign\",\"type\":\"bytes\"}],\"name\":\"CloseChannel\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"ChannelTimeout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"timeOut\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getStartDate\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"timeout\",\"type\":\"uint256\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"proof\",\"type\":\"bytes32\"}],\"name\":\"CheckHash\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"sig\",\"type\":\"bytes\"}],\"name\":\"CheckSig\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"data\",\"type\":\"string\"}],\"name\":\"Error\",\"type\":\"event\"}]"

// ChannelFuncSigs maps the 4-byte function signature to its string representation.
var ChannelFuncSigs = map[string]string{
	"39658245": "ChannelTimeout()",
	"2b7fa6be": "CloseChannel(bytes32,uint256,bytes)",
	"04758e79": "channelRecipient()",
	"075aa0c4": "channelSender()",
	"78f305c6": "getStartDate()",
	"0b97bc86": "startDate()",
	"614d85e1": "timeOut()",
}

// ChannelBin is the compiled bytecode used for deploying new contracts.
var ChannelBin = "0x60806040526040516105823803806105828339818101604052604081101561002657600080fd5b508051602090910151600180546001600160a01b039093166001600160a01b0319938416179055600080549092163317909155426002556003556105138061006f6000396000f3fe6080604052600436106100705760003560e01c80632b7fa6be1161004e5780632b7fa6be146100df578063396582451461019e578063614d85e1146101b357806378f305c6146101c857610070565b806304758e7914610072578063075aa0c4146100a35780630b97bc86146100b8575b005b34801561007e57600080fd5b506100876101dd565b604080516001600160a01b039092168252519081900360200190f35b3480156100af57600080fd5b506100876101ec565b3480156100c457600080fd5b506100cd6101fb565b60408051918252519081900360200190f35b3480156100eb57600080fd5b506100706004803603606081101561010257600080fd5b81359160208101359181019060608101604082013564010000000081111561012957600080fd5b82018360208201111561013b57600080fd5b8035906020019184600183028401116401000000008311171561015d57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250929550610201945050505050565b3480156101aa57600080fd5b506100706103c0565b3480156101bf57600080fd5b506100cd6103d3565b3480156101d457600080fd5b506100cd6103d9565b6001546001600160a01b031681565b6000546001600160a01b031681565b60025481565b6001546001600160a01b0316331461026e57604080516020808252601390820152723cb7ba9030b9103737ba103932b1b2bb34b2b960691b8183015290517f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa9181900360600190a16103bb565b604080513060601b602080830191909152603480830186905283518084039091018152605490920190925280519101208381146102fb57604080516020808252600d908201526c686173682069732077726f6e6760981b8183015290517f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa9181900360600190a1506103bb565b600061030785846103df565b6000549091506001600160a01b0380831691161461037b57604080516020808252601290820152717369676e61747572652069732077726f6e6760701b8183015290517f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa9181900360600190a150506103bb565b6001546040516001600160a01b039091169085156108fc029086906000818181858888f193505050506103ad57600080fd5b6000546001600160a01b0316ff5b505050565b426003546002540111156103ad57600080fd5b60035481565b60025490565b600081516041146103f2575060006104d8565b60208201516040830151606084015160001a7f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a082111561043857600093505050506104d8565b601b8160ff16101561044857601b015b8060ff16601b1415801561046057508060ff16601c14155b1561047157600093505050506104d8565b6040805160008152602080820180845289905260ff8416828401526060820186905260808201859052915160019260a0808401939192601f1981019281900390910190855afa1580156104c8573d6000803e3d6000fd5b5050506020604051035193505050505b9291505056fea265627a7a72305820467584d91ae6b0ef2f530abacb5cd59edc63c160d5b7e42c9b50ffad120ad7a464736f6c63430005090032"

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

// ChannelRecipient is a free data retrieval call binding the contract method 0x04758e79.
//
// Solidity: function channelRecipient() constant returns(address)
func (_Channel *ChannelCaller) ChannelRecipient(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Channel.contract.Call(opts, out, "channelRecipient")
	return *ret0, err
}

// ChannelRecipient is a free data retrieval call binding the contract method 0x04758e79.
//
// Solidity: function channelRecipient() constant returns(address)
func (_Channel *ChannelSession) ChannelRecipient() (common.Address, error) {
	return _Channel.Contract.ChannelRecipient(&_Channel.CallOpts)
}

// ChannelRecipient is a free data retrieval call binding the contract method 0x04758e79.
//
// Solidity: function channelRecipient() constant returns(address)
func (_Channel *ChannelCallerSession) ChannelRecipient() (common.Address, error) {
	return _Channel.Contract.ChannelRecipient(&_Channel.CallOpts)
}

// ChannelSender is a free data retrieval call binding the contract method 0x075aa0c4.
//
// Solidity: function channelSender() constant returns(address)
func (_Channel *ChannelCaller) ChannelSender(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Channel.contract.Call(opts, out, "channelSender")
	return *ret0, err
}

// ChannelSender is a free data retrieval call binding the contract method 0x075aa0c4.
//
// Solidity: function channelSender() constant returns(address)
func (_Channel *ChannelSession) ChannelSender() (common.Address, error) {
	return _Channel.Contract.ChannelSender(&_Channel.CallOpts)
}

// ChannelSender is a free data retrieval call binding the contract method 0x075aa0c4.
//
// Solidity: function channelSender() constant returns(address)
func (_Channel *ChannelCallerSession) ChannelSender() (common.Address, error) {
	return _Channel.Contract.ChannelSender(&_Channel.CallOpts)
}

// GetStartDate is a free data retrieval call binding the contract method 0x78f305c6.
//
// Solidity: function getStartDate() constant returns(uint256)
func (_Channel *ChannelCaller) GetStartDate(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Channel.contract.Call(opts, out, "getStartDate")
	return *ret0, err
}

// GetStartDate is a free data retrieval call binding the contract method 0x78f305c6.
//
// Solidity: function getStartDate() constant returns(uint256)
func (_Channel *ChannelSession) GetStartDate() (*big.Int, error) {
	return _Channel.Contract.GetStartDate(&_Channel.CallOpts)
}

// GetStartDate is a free data retrieval call binding the contract method 0x78f305c6.
//
// Solidity: function getStartDate() constant returns(uint256)
func (_Channel *ChannelCallerSession) GetStartDate() (*big.Int, error) {
	return _Channel.Contract.GetStartDate(&_Channel.CallOpts)
}

// StartDate is a free data retrieval call binding the contract method 0x0b97bc86.
//
// Solidity: function startDate() constant returns(uint256)
func (_Channel *ChannelCaller) StartDate(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Channel.contract.Call(opts, out, "startDate")
	return *ret0, err
}

// StartDate is a free data retrieval call binding the contract method 0x0b97bc86.
//
// Solidity: function startDate() constant returns(uint256)
func (_Channel *ChannelSession) StartDate() (*big.Int, error) {
	return _Channel.Contract.StartDate(&_Channel.CallOpts)
}

// StartDate is a free data retrieval call binding the contract method 0x0b97bc86.
//
// Solidity: function startDate() constant returns(uint256)
func (_Channel *ChannelCallerSession) StartDate() (*big.Int, error) {
	return _Channel.Contract.StartDate(&_Channel.CallOpts)
}

// TimeOut is a free data retrieval call binding the contract method 0x614d85e1.
//
// Solidity: function timeOut() constant returns(uint256)
func (_Channel *ChannelCaller) TimeOut(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Channel.contract.Call(opts, out, "timeOut")
	return *ret0, err
}

// TimeOut is a free data retrieval call binding the contract method 0x614d85e1.
//
// Solidity: function timeOut() constant returns(uint256)
func (_Channel *ChannelSession) TimeOut() (*big.Int, error) {
	return _Channel.Contract.TimeOut(&_Channel.CallOpts)
}

// TimeOut is a free data retrieval call binding the contract method 0x614d85e1.
//
// Solidity: function timeOut() constant returns(uint256)
func (_Channel *ChannelCallerSession) TimeOut() (*big.Int, error) {
	return _Channel.Contract.TimeOut(&_Channel.CallOpts)
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

// ChannelCheckHashIterator is returned from FilterCheckHash and is used to iterate over the raw logs and unpacked data for CheckHash events raised by the Channel contract.
type ChannelCheckHashIterator struct {
	Event *ChannelCheckHash // Event containing the contract specifics and raw log

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
func (it *ChannelCheckHashIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelCheckHash)
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
		it.Event = new(ChannelCheckHash)
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
func (it *ChannelCheckHashIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelCheckHashIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelCheckHash represents a CheckHash event raised by the Channel contract.
type ChannelCheckHash struct {
	Proof [32]byte
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterCheckHash is a free log retrieval operation binding the contract event 0x17d999f08836220ec49e62e94e67cc71744bd1abf3749c454443c40547c0acad.
//
// Solidity: event CheckHash(bytes32 proof)
func (_Channel *ChannelFilterer) FilterCheckHash(opts *bind.FilterOpts) (*ChannelCheckHashIterator, error) {

	logs, sub, err := _Channel.contract.FilterLogs(opts, "CheckHash")
	if err != nil {
		return nil, err
	}
	return &ChannelCheckHashIterator{contract: _Channel.contract, event: "CheckHash", logs: logs, sub: sub}, nil
}

// WatchCheckHash is a free log subscription operation binding the contract event 0x17d999f08836220ec49e62e94e67cc71744bd1abf3749c454443c40547c0acad.
//
// Solidity: event CheckHash(bytes32 proof)
func (_Channel *ChannelFilterer) WatchCheckHash(opts *bind.WatchOpts, sink chan<- *ChannelCheckHash) (event.Subscription, error) {

	logs, sub, err := _Channel.contract.WatchLogs(opts, "CheckHash")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelCheckHash)
				if err := _Channel.contract.UnpackLog(event, "CheckHash", log); err != nil {
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

// ParseCheckHash is a log parse operation binding the contract event 0x17d999f08836220ec49e62e94e67cc71744bd1abf3749c454443c40547c0acad.
//
// Solidity: event CheckHash(bytes32 proof)
func (_Channel *ChannelFilterer) ParseCheckHash(log types.Log) (*ChannelCheckHash, error) {
	event := new(ChannelCheckHash)
	if err := _Channel.contract.UnpackLog(event, "CheckHash", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ChannelCheckSigIterator is returned from FilterCheckSig and is used to iterate over the raw logs and unpacked data for CheckSig events raised by the Channel contract.
type ChannelCheckSigIterator struct {
	Event *ChannelCheckSig // Event containing the contract specifics and raw log

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
func (it *ChannelCheckSigIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelCheckSig)
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
		it.Event = new(ChannelCheckSig)
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
func (it *ChannelCheckSigIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelCheckSigIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelCheckSig represents a CheckSig event raised by the Channel contract.
type ChannelCheckSig struct {
	Sig []byte
	Raw types.Log // Blockchain specific contextual infos
}

// FilterCheckSig is a free log retrieval operation binding the contract event 0xb0a719a69f438395b8c98824cb5c498e01d5191a55ee59bc3357011c204e426c.
//
// Solidity: event CheckSig(bytes sig)
func (_Channel *ChannelFilterer) FilterCheckSig(opts *bind.FilterOpts) (*ChannelCheckSigIterator, error) {

	logs, sub, err := _Channel.contract.FilterLogs(opts, "CheckSig")
	if err != nil {
		return nil, err
	}
	return &ChannelCheckSigIterator{contract: _Channel.contract, event: "CheckSig", logs: logs, sub: sub}, nil
}

// WatchCheckSig is a free log subscription operation binding the contract event 0xb0a719a69f438395b8c98824cb5c498e01d5191a55ee59bc3357011c204e426c.
//
// Solidity: event CheckSig(bytes sig)
func (_Channel *ChannelFilterer) WatchCheckSig(opts *bind.WatchOpts, sink chan<- *ChannelCheckSig) (event.Subscription, error) {

	logs, sub, err := _Channel.contract.WatchLogs(opts, "CheckSig")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelCheckSig)
				if err := _Channel.contract.UnpackLog(event, "CheckSig", log); err != nil {
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

// ParseCheckSig is a log parse operation binding the contract event 0xb0a719a69f438395b8c98824cb5c498e01d5191a55ee59bc3357011c204e426c.
//
// Solidity: event CheckSig(bytes sig)
func (_Channel *ChannelFilterer) ParseCheckSig(log types.Log) (*ChannelCheckSig, error) {
	event := new(ChannelCheckSig)
	if err := _Channel.contract.UnpackLog(event, "CheckSig", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ChannelErrorIterator is returned from FilterError and is used to iterate over the raw logs and unpacked data for Error events raised by the Channel contract.
type ChannelErrorIterator struct {
	Event *ChannelError // Event containing the contract specifics and raw log

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
func (it *ChannelErrorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelError)
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
		it.Event = new(ChannelError)
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
func (it *ChannelErrorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelErrorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelError represents a Error event raised by the Channel contract.
type ChannelError struct {
	Data string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterError is a free log retrieval operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Channel *ChannelFilterer) FilterError(opts *bind.FilterOpts) (*ChannelErrorIterator, error) {

	logs, sub, err := _Channel.contract.FilterLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return &ChannelErrorIterator{contract: _Channel.contract, event: "Error", logs: logs, sub: sub}, nil
}

// WatchError is a free log subscription operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Channel *ChannelFilterer) WatchError(opts *bind.WatchOpts, sink chan<- *ChannelError) (event.Subscription, error) {

	logs, sub, err := _Channel.contract.WatchLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelError)
				if err := _Channel.contract.UnpackLog(event, "Error", log); err != nil {
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
func (_Channel *ChannelFilterer) ParseError(log types.Log) (*ChannelError, error) {
	event := new(ChannelError)
	if err := _Channel.contract.UnpackLog(event, "Error", log); err != nil {
		return nil, err
	}
	return event, nil
}

// DebugABI is the input ABI used to generate the binding from.
const DebugABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"data\",\"type\":\"string\"}],\"name\":\"Error\",\"type\":\"event\"}]"

// DebugBin is the compiled bytecode used for deploying new contracts.
var DebugBin = "0x6080604052348015600f57600080fd5b50603e80601d6000396000f3fe6080604052600080fdfea265627a7a723058208ba0e9dea3c04c5e99c36029769b94ef4ed6d502dd413fdaaf1fefe7d558000f64736f6c63430005090032"

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
