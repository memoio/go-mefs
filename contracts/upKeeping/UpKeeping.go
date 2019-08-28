// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package upKeeping

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

// UpKeepingABI is the input ABI used to generate the binding from.
const UpKeepingABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"user\",\"type\":\"address\"},{\"name\":\"keeper\",\"type\":\"address[]\"},{\"name\":\"provider\",\"type\":\"address[]\"},{\"name\":\"time\",\"type\":\"uint256\"},{\"name\":\"size\",\"type\":\"uint256\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"AddOrder\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"user\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"ReadPay\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"PayKeeper\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"PayProvider\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"data\",\"type\":\"string\"}],\"name\":\"Error\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"provider\",\"type\":\"address[]\"}],\"name\":\"addProvider\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"provider\",\"type\":\"address\"},{\"name\":\"money\",\"type\":\"uint256\"}],\"name\":\"spaceTimePay\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"provider\",\"type\":\"address\"}],\"name\":\"readPay\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOrder\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"address[]\"},{\"name\":\"\",\"type\":\"address[]\"},{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// UpKeepingBin is the compiled bytecode used for deploying new contracts.
const UpKeepingBin = `0x6080604052604051620011c0380380620011c0833981018060405260a08110156200002957600080fd5b810190808051906020019092919080516401000000008111156200004c57600080fd5b828101905060208101848111156200006357600080fd5b81518560208202830111640100000000821117156200008157600080fd5b505092919060200180516401000000008111156200009e57600080fd5b82810190506020810184811115620000b557600080fd5b8151856020820283011164010000000082111715620000d357600080fd5b50509291906020018051906020019092919080519060200190929190505050336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060c0604051908101604052808673ffffffffffffffffffffffffffffffffffffffff16815260200185815260200184815260200183815260200182815260200134815250600160008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506020820151816001019080519060200190620001dd92919062000255565b506040820151816002019080519060200190620001fc92919062000255565b50606082015181600301556080820151816004015560a082015181600501559050507f0905316f7faca135c292b6e6f8d91c19128d372722215fe029e74e75ef84c08760405160405180910390a150505050506200032a565b828054828255906000526020600020908101928215620002d1579160200282015b82811115620002d05782518260006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055509160200191906001019062000276565b5b509050620002e09190620002e4565b5090565b6200032791905b808211156200032357600081816101000a81549073ffffffffffffffffffffffffffffffffffffffff021916905550600101620002eb565b5090565b90565b610e86806200033a6000396000f3fe608060405260043610610078576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680630ca05f9f1461007a5780631e042234146100e35780632eb0346f14610156578063893d20e8146101b2578063c080810314610209578063d36dedd2146102e6575b005b34801561008657600080fd5b506100c96004803603602081101561009d57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506103e2565b604051808215151515815260200191505060405180910390f35b3480156100ef57600080fd5b5061013c6004803603604081101561010657600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291905050506105b3565b604051808215151515815260200191505060405180910390f35b6101986004803603602081101561016c57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506109f8565b604051808215151515815260200191505060405180910390f35b3480156101be57600080fd5b506101c7610aa5565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34801561021557600080fd5b506102cc6004803603602081101561022c57600080fd5b810190808035906020019064010000000081111561024957600080fd5b82018360208201111561025b57600080fd5b8035906020019184602083028401116401000000008311171561027d57600080fd5b919080806020026020016040519081016040528093929190818152602001838360200280828437600081840152601f19601f820116905080830192505050505050509192919290505050610ace565b604051808215151515815260200191505060405180910390f35b3480156102f257600080fd5b506102fb610c3e565b604051808773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018060200180602001868152602001858152602001848152602001838103835288818151815260200191508051906020019060200280838360005b8381101561038657808201518184015260208101905061036b565b50505050905001838103825287818151815260200191508051906020019060200280838360005b838110156103c85780820151818401526020810190506103ad565b505050509050019850505050505050505060405180910390f35b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614156105405760008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050826000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e908184604051808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a160019150506105ad565b7f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa60405180806020018281038252600e8152602001807fe4bda0e4b88de698af6f776e657200000000000000000000000000000000000081525060200191505060405180910390a16105ae565b5b919050565b6000806000905060008090505b60018001805490508110156106505760018001818154811015156105e057fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16141561064357600191505b80806001019150506105c0565b508015610982573073ffffffffffffffffffffffffffffffffffffffff16318311156106e7577f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa6040518080602001828103825260188152602001807fe59088e7baa6e4b8ade79a84e4bd99e9a29de4b88de8b6b3000000000000000081525060200191505060405180910390a16000915061097d565b6106f084610dad565b1515610767577f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa6040518080602001828103825260208152602001807fe8a681e8bdace8b4a6e79a84e59cb0e59d80e4b88de698af70726f766964657281525060200191505060405180910390a16000915061097d565b6000600a8481151561077557fe5b0490508473ffffffffffffffffffffffffffffffffffffffff166108fc600983029081150290604051600060405180830381858888f193505050501580156107c1573d6000803e3d6000fd5b50600981028573ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167f1569130f5bdbde161a213db1c477e4f2670f09e2a9c1c08ca9bafe749b80cb4160405160405180910390a460006001800180549050905060008090505b8181101561097557600180018181548110151561084b57fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc838581151561089a57fe5b049081150290604051600060405180830381858888f193505050501580156108c6573d6000803e3d6000fd5b5081838115156108d257fe5b0460018001828154811015156108e457fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167faa4c66f6ddfadc835acfabab55148a78bc3e6867ed1cdb36461a10685af4c0c360405160405180910390a48080600101915050610832565b506001935050505b6109f0565b7f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa60405180806020018281038252600f8152602001807fe4bda0e4b88de698af6b6565706572000000000000000000000000000000000081525060200191505060405180910390a1506109f2565b505b92915050565b60008173ffffffffffffffffffffffffffffffffffffffff166108fc349081150290604051600060405180830381858888f19350505050158015610a40573d6000803e3d6000fd5b50348273ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167f412887bd474e56e243eb289e55bd2cc3fb5023d072e45e9541a3963107e3fe7c60405160405180910390a460019050919050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161415610bcb5760008090505b8251811015610bc15760016002018382815181101515610b4757fe5b9060200190602002015190806001815401808255809150509060018203906000526020600020016000909192909190916101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550508080600101915050610b2b565b5060019050610c38565b7f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa60405180806020018281038252600e8152602001807fe4bda0e4b88de698af6f776e657200000000000000000000000000000000000081525060200191505060405180910390a1610c39565b5b919050565b60006060806000806000600160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1660018001600160020160016003015460016004015460016005015484805480602002602001604051908101604052809291908181526020018280548015610d0a57602002820191906000526020600020905b8160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019060010190808311610cc0575b5050505050945083805480602002602001604051908101604052809291908181526020018280548015610d9257602002820191906000526020600020905b8160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019060010190808311610d48575b50505050509350955095509550955095509550909192939495565b6000806000905060008090505b600160020180549050811015610e5057600160020181815481101515610ddc57fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff161415610e435760019150610e50565b8080600101915050610dba565b508091505091905056fea165627a7a723058207e07415ad7f4c3b99b6afc009397f3f4481f0c248537283261d54f4872d7362b0029`

// DeployUpKeeping deploys a new Ethereum contract, binding an instance of UpKeeping to it.
func DeployUpKeeping(auth *bind.TransactOpts, backend bind.ContractBackend, user common.Address, keeper []common.Address, provider []common.Address, time *big.Int, size *big.Int) (common.Address, *types.Transaction, *UpKeeping, error) {
	parsed, err := abi.JSON(strings.NewReader(UpKeepingABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(UpKeepingBin), backend, user, keeper, provider, time, size)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &UpKeeping{UpKeepingCaller: UpKeepingCaller{contract: contract}, UpKeepingTransactor: UpKeepingTransactor{contract: contract}, UpKeepingFilterer: UpKeepingFilterer{contract: contract}}, nil
}

// UpKeeping is an auto generated Go binding around an Ethereum contract.
type UpKeeping struct {
	UpKeepingCaller     // Read-only binding to the contract
	UpKeepingTransactor // Write-only binding to the contract
	UpKeepingFilterer   // Log filterer for contract events
}

// UpKeepingCaller is an auto generated read-only Go binding around an Ethereum contract.
type UpKeepingCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UpKeepingTransactor is an auto generated write-only Go binding around an Ethereum contract.
type UpKeepingTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UpKeepingFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type UpKeepingFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UpKeepingSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type UpKeepingSession struct {
	Contract     *UpKeeping        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// UpKeepingCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type UpKeepingCallerSession struct {
	Contract *UpKeepingCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// UpKeepingTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type UpKeepingTransactorSession struct {
	Contract     *UpKeepingTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// UpKeepingRaw is an auto generated low-level Go binding around an Ethereum contract.
type UpKeepingRaw struct {
	Contract *UpKeeping // Generic contract binding to access the raw methods on
}

// UpKeepingCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type UpKeepingCallerRaw struct {
	Contract *UpKeepingCaller // Generic read-only contract binding to access the raw methods on
}

// UpKeepingTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type UpKeepingTransactorRaw struct {
	Contract *UpKeepingTransactor // Generic write-only contract binding to access the raw methods on
}

// NewUpKeeping creates a new instance of UpKeeping, bound to a specific deployed contract.
func NewUpKeeping(address common.Address, backend bind.ContractBackend) (*UpKeeping, error) {
	contract, err := bindUpKeeping(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &UpKeeping{UpKeepingCaller: UpKeepingCaller{contract: contract}, UpKeepingTransactor: UpKeepingTransactor{contract: contract}, UpKeepingFilterer: UpKeepingFilterer{contract: contract}}, nil
}

// NewUpKeepingCaller creates a new read-only instance of UpKeeping, bound to a specific deployed contract.
func NewUpKeepingCaller(address common.Address, caller bind.ContractCaller) (*UpKeepingCaller, error) {
	contract, err := bindUpKeeping(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &UpKeepingCaller{contract: contract}, nil
}

// NewUpKeepingTransactor creates a new write-only instance of UpKeeping, bound to a specific deployed contract.
func NewUpKeepingTransactor(address common.Address, transactor bind.ContractTransactor) (*UpKeepingTransactor, error) {
	contract, err := bindUpKeeping(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &UpKeepingTransactor{contract: contract}, nil
}

// NewUpKeepingFilterer creates a new log filterer instance of UpKeeping, bound to a specific deployed contract.
func NewUpKeepingFilterer(address common.Address, filterer bind.ContractFilterer) (*UpKeepingFilterer, error) {
	contract, err := bindUpKeeping(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &UpKeepingFilterer{contract: contract}, nil
}

// bindUpKeeping binds a generic wrapper to an already deployed contract.
func bindUpKeeping(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(UpKeepingABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_UpKeeping *UpKeepingRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _UpKeeping.Contract.UpKeepingCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_UpKeeping *UpKeepingRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UpKeeping.Contract.UpKeepingTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_UpKeeping *UpKeepingRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _UpKeeping.Contract.UpKeepingTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_UpKeeping *UpKeepingCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _UpKeeping.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_UpKeeping *UpKeepingTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UpKeeping.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_UpKeeping *UpKeepingTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _UpKeeping.Contract.contract.Transact(opts, method, params...)
}

// GetOrder is a free data retrieval call binding the contract method 0xd36dedd2.
//
// Solidity: function getOrder() constant returns(address, address[], address[], uint256, uint256, uint256)
func (_UpKeeping *UpKeepingCaller) GetOrder(opts *bind.CallOpts) (common.Address, []common.Address, []common.Address, *big.Int, *big.Int, *big.Int, error) {
	var (
		ret0 = new(common.Address)
		ret1 = new([]common.Address)
		ret2 = new([]common.Address)
		ret3 = new(*big.Int)
		ret4 = new(*big.Int)
		ret5 = new(*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
		ret3,
		ret4,
		ret5,
	}
	err := _UpKeeping.contract.Call(opts, out, "getOrder")
	return *ret0, *ret1, *ret2, *ret3, *ret4, *ret5, err
}

// GetOrder is a free data retrieval call binding the contract method 0xd36dedd2.
//
// Solidity: function getOrder() constant returns(address, address[], address[], uint256, uint256, uint256)
func (_UpKeeping *UpKeepingSession) GetOrder() (common.Address, []common.Address, []common.Address, *big.Int, *big.Int, *big.Int, error) {
	return _UpKeeping.Contract.GetOrder(&_UpKeeping.CallOpts)
}

// GetOrder is a free data retrieval call binding the contract method 0xd36dedd2.
//
// Solidity: function getOrder() constant returns(address, address[], address[], uint256, uint256, uint256)
func (_UpKeeping *UpKeepingCallerSession) GetOrder() (common.Address, []common.Address, []common.Address, *big.Int, *big.Int, *big.Int, error) {
	return _UpKeeping.Contract.GetOrder(&_UpKeeping.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_UpKeeping *UpKeepingCaller) GetOwner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _UpKeeping.contract.Call(opts, out, "getOwner")
	return *ret0, err
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_UpKeeping *UpKeepingSession) GetOwner() (common.Address, error) {
	return _UpKeeping.Contract.GetOwner(&_UpKeeping.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_UpKeeping *UpKeepingCallerSession) GetOwner() (common.Address, error) {
	return _UpKeeping.Contract.GetOwner(&_UpKeeping.CallOpts)
}

// AddProvider is a paid mutator transaction binding the contract method 0xc0808103.
//
// Solidity: function addProvider(address[] provider) returns(bool)
func (_UpKeeping *UpKeepingTransactor) AddProvider(opts *bind.TransactOpts, provider []common.Address) (*types.Transaction, error) {
	return _UpKeeping.contract.Transact(opts, "addProvider", provider)
}

// AddProvider is a paid mutator transaction binding the contract method 0xc0808103.
//
// Solidity: function addProvider(address[] provider) returns(bool)
func (_UpKeeping *UpKeepingSession) AddProvider(provider []common.Address) (*types.Transaction, error) {
	return _UpKeeping.Contract.AddProvider(&_UpKeeping.TransactOpts, provider)
}

// AddProvider is a paid mutator transaction binding the contract method 0xc0808103.
//
// Solidity: function addProvider(address[] provider) returns(bool)
func (_UpKeeping *UpKeepingTransactorSession) AddProvider(provider []common.Address) (*types.Transaction, error) {
	return _UpKeeping.Contract.AddProvider(&_UpKeeping.TransactOpts, provider)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_UpKeeping *UpKeepingTransactor) AlterOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _UpKeeping.contract.Transact(opts, "alterOwner", newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_UpKeeping *UpKeepingSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _UpKeeping.Contract.AlterOwner(&_UpKeeping.TransactOpts, newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_UpKeeping *UpKeepingTransactorSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _UpKeeping.Contract.AlterOwner(&_UpKeeping.TransactOpts, newOwner)
}

// ReadPay is a paid mutator transaction binding the contract method 0x2eb0346f.
//
// Solidity: function readPay(address provider) returns(bool)
func (_UpKeeping *UpKeepingTransactor) ReadPay(opts *bind.TransactOpts, provider common.Address) (*types.Transaction, error) {
	return _UpKeeping.contract.Transact(opts, "readPay", provider)
}

// ReadPay is a paid mutator transaction binding the contract method 0x2eb0346f.
//
// Solidity: function readPay(address provider) returns(bool)
func (_UpKeeping *UpKeepingSession) ReadPay(provider common.Address) (*types.Transaction, error) {
	return _UpKeeping.Contract.ReadPay(&_UpKeeping.TransactOpts, provider)
}

// ReadPay is a paid mutator transaction binding the contract method 0x2eb0346f.
//
// Solidity: function readPay(address provider) returns(bool)
func (_UpKeeping *UpKeepingTransactorSession) ReadPay(provider common.Address) (*types.Transaction, error) {
	return _UpKeeping.Contract.ReadPay(&_UpKeeping.TransactOpts, provider)
}

// SpaceTimePay is a paid mutator transaction binding the contract method 0x1e042234.
//
// Solidity: function spaceTimePay(address provider, uint256 money) returns(bool)
func (_UpKeeping *UpKeepingTransactor) SpaceTimePay(opts *bind.TransactOpts, provider common.Address, money *big.Int) (*types.Transaction, error) {
	return _UpKeeping.contract.Transact(opts, "spaceTimePay", provider, money)
}

// SpaceTimePay is a paid mutator transaction binding the contract method 0x1e042234.
//
// Solidity: function spaceTimePay(address provider, uint256 money) returns(bool)
func (_UpKeeping *UpKeepingSession) SpaceTimePay(provider common.Address, money *big.Int) (*types.Transaction, error) {
	return _UpKeeping.Contract.SpaceTimePay(&_UpKeeping.TransactOpts, provider, money)
}

// SpaceTimePay is a paid mutator transaction binding the contract method 0x1e042234.
//
// Solidity: function spaceTimePay(address provider, uint256 money) returns(bool)
func (_UpKeeping *UpKeepingTransactorSession) SpaceTimePay(provider common.Address, money *big.Int) (*types.Transaction, error) {
	return _UpKeeping.Contract.SpaceTimePay(&_UpKeeping.TransactOpts, provider, money)
}

// UpKeepingAddOrderIterator is returned from FilterAddOrder and is used to iterate over the raw logs and unpacked data for AddOrder events raised by the UpKeeping contract.
type UpKeepingAddOrderIterator struct {
	Event *UpKeepingAddOrder // Event containing the contract specifics and raw log

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
func (it *UpKeepingAddOrderIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UpKeepingAddOrder)
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
		it.Event = new(UpKeepingAddOrder)
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
func (it *UpKeepingAddOrderIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UpKeepingAddOrderIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UpKeepingAddOrder represents a AddOrder event raised by the UpKeeping contract.
type UpKeepingAddOrder struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterAddOrder is a free log retrieval operation binding the contract event 0x0905316f7faca135c292b6e6f8d91c19128d372722215fe029e74e75ef84c087.
//
// Solidity: event AddOrder()
func (_UpKeeping *UpKeepingFilterer) FilterAddOrder(opts *bind.FilterOpts) (*UpKeepingAddOrderIterator, error) {

	logs, sub, err := _UpKeeping.contract.FilterLogs(opts, "AddOrder")
	if err != nil {
		return nil, err
	}
	return &UpKeepingAddOrderIterator{contract: _UpKeeping.contract, event: "AddOrder", logs: logs, sub: sub}, nil
}

// WatchAddOrder is a free log subscription operation binding the contract event 0x0905316f7faca135c292b6e6f8d91c19128d372722215fe029e74e75ef84c087.
//
// Solidity: event AddOrder()
func (_UpKeeping *UpKeepingFilterer) WatchAddOrder(opts *bind.WatchOpts, sink chan<- *UpKeepingAddOrder) (event.Subscription, error) {

	logs, sub, err := _UpKeeping.contract.WatchLogs(opts, "AddOrder")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UpKeepingAddOrder)
				if err := _UpKeeping.contract.UnpackLog(event, "AddOrder", log); err != nil {
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

// UpKeepingAlterOwnerIterator is returned from FilterAlterOwner and is used to iterate over the raw logs and unpacked data for AlterOwner events raised by the UpKeeping contract.
type UpKeepingAlterOwnerIterator struct {
	Event *UpKeepingAlterOwner // Event containing the contract specifics and raw log

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
func (it *UpKeepingAlterOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UpKeepingAlterOwner)
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
		it.Event = new(UpKeepingAlterOwner)
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
func (it *UpKeepingAlterOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UpKeepingAlterOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UpKeepingAlterOwner represents a AlterOwner event raised by the UpKeeping contract.
type UpKeepingAlterOwner struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAlterOwner is a free log retrieval operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_UpKeeping *UpKeepingFilterer) FilterAlterOwner(opts *bind.FilterOpts) (*UpKeepingAlterOwnerIterator, error) {

	logs, sub, err := _UpKeeping.contract.FilterLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return &UpKeepingAlterOwnerIterator{contract: _UpKeeping.contract, event: "AlterOwner", logs: logs, sub: sub}, nil
}

// WatchAlterOwner is a free log subscription operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_UpKeeping *UpKeepingFilterer) WatchAlterOwner(opts *bind.WatchOpts, sink chan<- *UpKeepingAlterOwner) (event.Subscription, error) {

	logs, sub, err := _UpKeeping.contract.WatchLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UpKeepingAlterOwner)
				if err := _UpKeeping.contract.UnpackLog(event, "AlterOwner", log); err != nil {
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

// UpKeepingErrorIterator is returned from FilterError and is used to iterate over the raw logs and unpacked data for Error events raised by the UpKeeping contract.
type UpKeepingErrorIterator struct {
	Event *UpKeepingError // Event containing the contract specifics and raw log

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
func (it *UpKeepingErrorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UpKeepingError)
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
		it.Event = new(UpKeepingError)
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
func (it *UpKeepingErrorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UpKeepingErrorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UpKeepingError represents a Error event raised by the UpKeeping contract.
type UpKeepingError struct {
	Data string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterError is a free log retrieval operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_UpKeeping *UpKeepingFilterer) FilterError(opts *bind.FilterOpts) (*UpKeepingErrorIterator, error) {

	logs, sub, err := _UpKeeping.contract.FilterLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return &UpKeepingErrorIterator{contract: _UpKeeping.contract, event: "Error", logs: logs, sub: sub}, nil
}

// WatchError is a free log subscription operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_UpKeeping *UpKeepingFilterer) WatchError(opts *bind.WatchOpts, sink chan<- *UpKeepingError) (event.Subscription, error) {

	logs, sub, err := _UpKeeping.contract.WatchLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UpKeepingError)
				if err := _UpKeeping.contract.UnpackLog(event, "Error", log); err != nil {
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

// UpKeepingPayKeeperIterator is returned from FilterPayKeeper and is used to iterate over the raw logs and unpacked data for PayKeeper events raised by the UpKeeping contract.
type UpKeepingPayKeeperIterator struct {
	Event *UpKeepingPayKeeper // Event containing the contract specifics and raw log

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
func (it *UpKeepingPayKeeperIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UpKeepingPayKeeper)
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
		it.Event = new(UpKeepingPayKeeper)
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
func (it *UpKeepingPayKeeperIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UpKeepingPayKeeperIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UpKeepingPayKeeper represents a PayKeeper event raised by the UpKeeping contract.
type UpKeepingPayKeeper struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterPayKeeper is a free log retrieval operation binding the contract event 0xaa4c66f6ddfadc835acfabab55148a78bc3e6867ed1cdb36461a10685af4c0c3.
//
// Solidity: event PayKeeper(address indexed from, address indexed to, uint256 indexed value)
func (_UpKeeping *UpKeepingFilterer) FilterPayKeeper(opts *bind.FilterOpts, from []common.Address, to []common.Address, value []*big.Int) (*UpKeepingPayKeeperIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var valueRule []interface{}
	for _, valueItem := range value {
		valueRule = append(valueRule, valueItem)
	}

	logs, sub, err := _UpKeeping.contract.FilterLogs(opts, "PayKeeper", fromRule, toRule, valueRule)
	if err != nil {
		return nil, err
	}
	return &UpKeepingPayKeeperIterator{contract: _UpKeeping.contract, event: "PayKeeper", logs: logs, sub: sub}, nil
}

// WatchPayKeeper is a free log subscription operation binding the contract event 0xaa4c66f6ddfadc835acfabab55148a78bc3e6867ed1cdb36461a10685af4c0c3.
//
// Solidity: event PayKeeper(address indexed from, address indexed to, uint256 indexed value)
func (_UpKeeping *UpKeepingFilterer) WatchPayKeeper(opts *bind.WatchOpts, sink chan<- *UpKeepingPayKeeper, from []common.Address, to []common.Address, value []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var valueRule []interface{}
	for _, valueItem := range value {
		valueRule = append(valueRule, valueItem)
	}

	logs, sub, err := _UpKeeping.contract.WatchLogs(opts, "PayKeeper", fromRule, toRule, valueRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UpKeepingPayKeeper)
				if err := _UpKeeping.contract.UnpackLog(event, "PayKeeper", log); err != nil {
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

// UpKeepingPayProviderIterator is returned from FilterPayProvider and is used to iterate over the raw logs and unpacked data for PayProvider events raised by the UpKeeping contract.
type UpKeepingPayProviderIterator struct {
	Event *UpKeepingPayProvider // Event containing the contract specifics and raw log

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
func (it *UpKeepingPayProviderIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UpKeepingPayProvider)
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
		it.Event = new(UpKeepingPayProvider)
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
func (it *UpKeepingPayProviderIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UpKeepingPayProviderIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UpKeepingPayProvider represents a PayProvider event raised by the UpKeeping contract.
type UpKeepingPayProvider struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterPayProvider is a free log retrieval operation binding the contract event 0x1569130f5bdbde161a213db1c477e4f2670f09e2a9c1c08ca9bafe749b80cb41.
//
// Solidity: event PayProvider(address indexed from, address indexed to, uint256 indexed value)
func (_UpKeeping *UpKeepingFilterer) FilterPayProvider(opts *bind.FilterOpts, from []common.Address, to []common.Address, value []*big.Int) (*UpKeepingPayProviderIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var valueRule []interface{}
	for _, valueItem := range value {
		valueRule = append(valueRule, valueItem)
	}

	logs, sub, err := _UpKeeping.contract.FilterLogs(opts, "PayProvider", fromRule, toRule, valueRule)
	if err != nil {
		return nil, err
	}
	return &UpKeepingPayProviderIterator{contract: _UpKeeping.contract, event: "PayProvider", logs: logs, sub: sub}, nil
}

// WatchPayProvider is a free log subscription operation binding the contract event 0x1569130f5bdbde161a213db1c477e4f2670f09e2a9c1c08ca9bafe749b80cb41.
//
// Solidity: event PayProvider(address indexed from, address indexed to, uint256 indexed value)
func (_UpKeeping *UpKeepingFilterer) WatchPayProvider(opts *bind.WatchOpts, sink chan<- *UpKeepingPayProvider, from []common.Address, to []common.Address, value []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var valueRule []interface{}
	for _, valueItem := range value {
		valueRule = append(valueRule, valueItem)
	}

	logs, sub, err := _UpKeeping.contract.WatchLogs(opts, "PayProvider", fromRule, toRule, valueRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UpKeepingPayProvider)
				if err := _UpKeeping.contract.UnpackLog(event, "PayProvider", log); err != nil {
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

// UpKeepingReadPayIterator is returned from FilterReadPay and is used to iterate over the raw logs and unpacked data for ReadPay events raised by the UpKeeping contract.
type UpKeepingReadPayIterator struct {
	Event *UpKeepingReadPay // Event containing the contract specifics and raw log

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
func (it *UpKeepingReadPayIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UpKeepingReadPay)
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
		it.Event = new(UpKeepingReadPay)
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
func (it *UpKeepingReadPayIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UpKeepingReadPayIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UpKeepingReadPay represents a ReadPay event raised by the UpKeeping contract.
type UpKeepingReadPay struct {
	User     common.Address
	Provider common.Address
	Value    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterReadPay is a free log retrieval operation binding the contract event 0x412887bd474e56e243eb289e55bd2cc3fb5023d072e45e9541a3963107e3fe7c.
//
// Solidity: event ReadPay(address indexed user, address indexed provider, uint256 indexed value)
func (_UpKeeping *UpKeepingFilterer) FilterReadPay(opts *bind.FilterOpts, user []common.Address, provider []common.Address, value []*big.Int) (*UpKeepingReadPayIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var valueRule []interface{}
	for _, valueItem := range value {
		valueRule = append(valueRule, valueItem)
	}

	logs, sub, err := _UpKeeping.contract.FilterLogs(opts, "ReadPay", userRule, providerRule, valueRule)
	if err != nil {
		return nil, err
	}
	return &UpKeepingReadPayIterator{contract: _UpKeeping.contract, event: "ReadPay", logs: logs, sub: sub}, nil
}

// WatchReadPay is a free log subscription operation binding the contract event 0x412887bd474e56e243eb289e55bd2cc3fb5023d072e45e9541a3963107e3fe7c.
//
// Solidity: event ReadPay(address indexed user, address indexed provider, uint256 indexed value)
func (_UpKeeping *UpKeepingFilterer) WatchReadPay(opts *bind.WatchOpts, sink chan<- *UpKeepingReadPay, user []common.Address, provider []common.Address, value []*big.Int) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var valueRule []interface{}
	for _, valueItem := range value {
		valueRule = append(valueRule, valueItem)
	}

	logs, sub, err := _UpKeeping.contract.WatchLogs(opts, "ReadPay", userRule, providerRule, valueRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UpKeepingReadPay)
				if err := _UpKeeping.contract.UnpackLog(event, "ReadPay", log); err != nil {
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
