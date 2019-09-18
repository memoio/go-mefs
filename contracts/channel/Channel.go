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
const ChannelABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"channelRecipient\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"channelSender\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"startDate\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"},{\"name\":\"value\",\"type\":\"uint256\"},{\"name\":\"sign\",\"type\":\"bytes\"}],\"name\":\"CloseChannel\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"ChannelTimeout\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getInfo\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"timeOut\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"timeout\",\"type\":\"uint256\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"}]"

// ChannelBin is the compiled bytecode used for deploying new contracts.
const ChannelBin = `0x60806040526040516104c73803806104c78339818101604052604081101561002657600080fd5b508051602090910151600180546001600160a01b039093166001600160a01b0319938416179055600080549092163317909155426002556003556104588061006f6000396000f3fe6080604052600436106100705760003560e01c80632b7fa6be1161004e5780632b7fa6be146100df578063396582451461019e5780635a9b0b89146101b3578063614d85e1146101f757610070565b806304758e7914610072578063075aa0c4146100a35780630b97bc86146100b8575b005b34801561007e57600080fd5b5061008761020c565b604080516001600160a01b039092168252519081900360200190f35b3480156100af57600080fd5b5061008761021b565b3480156100c457600080fd5b506100cd61022a565b60408051918252519081900360200190f35b3480156100eb57600080fd5b506100706004803603606081101561010257600080fd5b81359160208101359181019060608101604082013564010000000081111561012957600080fd5b82018360208201111561013b57600080fd5b8035906020019184600183028401116401000000008311171561015d57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250929550610230945050505050565b3480156101aa57600080fd5b506100706102ec565b3480156101bf57600080fd5b506101c86102ff565b6040805194855260208501939093526001600160a01b0391821684840152166060830152519081900360800190f35b34801561020357600080fd5b506100cd61031e565b6001546001600160a01b031681565b6000546001600160a01b031681565b60025481565b6001546001600160a01b0316331461024757600080fd5b604080513060601b6020808301919091526034808301869052835180840390910181526054909201909252805191012083811461028357600080fd5b600061028f8584610324565b6000549091506001600160a01b038083169116146102ac57600080fd5b6001546040516001600160a01b039091169085156108fc029086906000818181858888f193505050506102de57600080fd5b6000546001600160a01b0316ff5b426003546002540111156102de57600080fd5b6002546003546000546001546001600160a01b03918216911690919293565b60035481565b600081516041146103375750600061041d565b60208201516040830151606084015160001a7f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a082111561037d576000935050505061041d565b601b8160ff16101561038d57601b015b8060ff16601b141580156103a557508060ff16601c14155b156103b6576000935050505061041d565b6040805160008152602080820180845289905260ff8416828401526060820186905260808201859052915160019260a0808401939192601f1981019281900390910190855afa15801561040d573d6000803e3d6000fd5b5050506020604051035193505050505b9291505056fea265627a7a7230582035c9113e352d06f963777e5cbf46f7892eaffca8a045ff5dbe5e935cc72e939464736f6c63430005090032`

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
