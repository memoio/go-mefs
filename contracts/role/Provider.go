// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package role

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

// ProviderABI is the input ABI used to generate the binding from.
const ProviderABI = "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_price\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"acc\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"sum\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"cancelPledge\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"cancelPledgeStatus\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAllAddress\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getPrice\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"acc\",\"type\":\"address\"}],\"name\":\"info\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_size\",\"type\":\"uint256\"}],\"name\":\"pledge\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"set\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"setBanned\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_price\",\"type\":\"uint256\"}],\"name\":\"setPrice\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// ProviderBin is the compiled bytecode used for deploying new contracts.
var ProviderBin = "0x608060405234801561001057600080fd5b5060405161145a38038061145a8339818101604052602081101561003357600080fd5b8101908080519060200190929190505050336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555080600281905550506113bf8061009b6000396000f3fe60806040526004361061009c5760003560e01c806388c9bcce1161006457806388c9bcce146102b3578063893d20e81461032857806391b7f5ed1461037f57806398d5fdca146103d2578063ae5e2666146103fd578063e3685c401461041f5761009c565b80630aae7a6b146100a15780630ca05f9f1461012357806335e3b25a1461018c578063715b208b146102015780637326c9c01461026d575b600080fd5b3480156100ad57600080fd5b506100f0600480360360208110156100c457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610491565b60405180851515151581526020018415151515815260200183815260200182815260200194505050505060405180910390f35b34801561012f57600080fd5b506101726004803603602081101561014657600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610573565b604051808215151515815260200191505060405180910390f35b34801561019857600080fd5b506101e7600480360360408110156101af57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080351515906020019092919050505061073e565b604051808215151515815260200191505060405180910390f35b34801561020d57600080fd5b506102166109a7565b6040518080602001828103825283818151815260200191508051906020019060200280838360005b8381101561025957808201518184015260208101905061023e565b505050509050019250505060405180910390f35b6102996004803603602081101561028357600080fd5b8101908080359060200190929190505050610bb9565b604051808215151515815260200191505060405180910390f35b3480156102bf57600080fd5b5061030e600480360360408110156102d657600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803515159060200190929190505050610d32565b604051808215151515815260200191505060405180910390f35b34801561033457600080fd5b5061033d610f62565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34801561038b57600080fd5b506103b8600480360360208110156103a257600080fd5b8101908080359060200190929190505050610f8b565b604051808215151515815260200191505060405180910390f35b3480156103de57600080fd5b506103e761105f565b6040518082815260200191505060405180910390f35b610405611069565b604051808215151515815260200191505060405180910390f35b6104776004803603606081101561043557600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291908035151590602001909291905050506110ce565b604051808215151515815260200191505060405180910390f35b60008060008060006104a2866112dd565b90506001805490508114156104cc576000600160008081915080905094509450945094505061056c565b600181815481106104d957fe5b906000526020600020906003020160000160149054906101000a900460ff166001828154811061050557fe5b906000526020600020906003020160000160159054906101000a900460ff166001838154811061053157fe5b9060005260206000209060030201600101546001848154811061055057fe5b9060005260206000209060030201600201549450945094509450505b9193509193565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610637576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050826000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e908184604051808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a16001915050919050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610802576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b600061080d846112dd565b9050600180549050811461088f57600015156001828154811061082c57fe5b906000526020600020906003020160000160159054906101000a900460ff161515141561088a57826001828154811061086157fe5b906000526020600020906003020160000160146101000a81548160ff0219169083151502179055505b61099c565b60016040518060a001604052808673ffffffffffffffffffffffffffffffffffffffff1681526020018515158152602001600015158152602001600081526020016000815250908060018154018082558091505060019003906000526020600020906003020160009091909190915060008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060208201518160000160146101000a81548160ff02191690831515021790555060408201518160000160156101000a81548160ff021916908315150217905550606082015181600101556080820151816002015550505b600191505092915050565b6060806001805490506040519080825280602002602001820160405280156109de5781602001602082028036833780820191505090505b509050600080905060008090505b600180549050811015610b05576001151560018281548110610a0a57fe5b906000526020600020906003020160000160149054906101000a900460ff1615151415610af8576000151560018281548110610a4257fe5b906000526020600020906003020160000160159054906101000a900460ff1615151415610af75760018181548110610a7657fe5b906000526020600020906003020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16838381518110610ab457fe5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505081806001019250505b5b80806001019150506109ec565b50606081604051908082528060200260200182016040528015610b375781602001602082028036833780820191505090505b50905060008090505b82811015610baf57838181518110610b5457fe5b6020026020010151828281518110610b6857fe5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff16815250508080600101915050610b40565b5080935050505090565b600080610bc5336112dd565b9050600180549050811415610c25573373ffffffffffffffffffffffffffffffffffffffff166108fc349081150290604051600060405180830381858888f19350505050158015610c1a573d6000803e3d6000fd5b506000915050610d2d565b826002540260018281548110610c3757fe5b90600052602060002090600302016001015434011015610ca2573373ffffffffffffffffffffffffffffffffffffffff166108fc349081150290604051600060405180830381858888f19350505050158015610c97573d6000803e3d6000fd5b506000915050610d2d565b6001808281548110610cb057fe5b906000526020600020906003020160000160146101000a81548160ff0219169083151502179055503460018281548110610ce657fe5b9060005260206000209060030201600101600082825401925050819055504260018281548110610d1257fe5b90600052602060002090600302016002018190555060019150505b919050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610df6576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b6000610e01846112dd565b90506001805490508114610e4a578260018281548110610e1d57fe5b906000526020600020906003020160000160156101000a81548160ff021916908315150217905550610f57565b60016040518060a001604052808673ffffffffffffffffffffffffffffffffffffffff1681526020016000151581526020018515158152602001600081526020016000815250908060018154018082558091505060019003906000526020600020906003020160009091909190915060008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060208201518160000160146101000a81548160ff02191690831515021790555060408201518160000160156101000a81548160ff021916908315150217905550606082015181600101556080820151816002015550505b600191505092915050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461104f576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b8160028190555060019050919050565b6000600254905090565b600080611075336112dd565b905060018054905081141561108e5760009150506110cb565b60006001828154811061109d57fe5b906000526020600020906003020160000160146101000a81548160ff02191690831515021790555060019150505b90565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614611192576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b600061119d856112dd565b90506001805490508114156111b65760009150506112d6565b6000600182815481106111c557fe5b90600052602060002090600302016001015485111561120457600182815481106111eb57fe5b9060005260206000209060030201600101549050611208565b8490505b600081141561121c576000925050506112d6565b806001838154811061122a57fe5b90600052602060002090600302016001016000828254039250508190555083156112cf57600015156001838154811061125f57fe5b906000526020600020906003020160000160159054906101000a900460ff16151514156112ce578573ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f193505050501580156112cc573d6000803e3d6000fd5b505b5b6001925050505b9392505050565b600080600090505b60018054905081101561137a578273ffffffffffffffffffffffffffffffffffffffff166001828154811061131657fe5b906000526020600020906003020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16141561136d5780915050611384565b80806001019150506112e5565b5060018054905090505b91905056fea264697066735822122025668f045b50d3c69202f455343e3587f53a7ff478afb8bfa5c0c1e1165eec3864736f6c63430006030033"

// DeployProvider deploys a new Ethereum contract, binding an instance of Provider to it.
func DeployProvider(auth *bind.TransactOpts, backend bind.ContractBackend, _price *big.Int) (common.Address, *types.Transaction, *Provider, error) {
	parsed, err := abi.JSON(strings.NewReader(ProviderABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ProviderBin), backend, _price)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Provider{ProviderCaller: ProviderCaller{contract: contract}, ProviderTransactor: ProviderTransactor{contract: contract}, ProviderFilterer: ProviderFilterer{contract: contract}}, nil
}

// Provider is an auto generated Go binding around an Ethereum contract.
type Provider struct {
	ProviderCaller     // Read-only binding to the contract
	ProviderTransactor // Write-only binding to the contract
	ProviderFilterer   // Log filterer for contract events
}

// ProviderCaller is an auto generated read-only Go binding around an Ethereum contract.
type ProviderCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProviderTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ProviderTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProviderFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ProviderFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProviderSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ProviderSession struct {
	Contract     *Provider         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ProviderCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ProviderCallerSession struct {
	Contract *ProviderCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// ProviderTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ProviderTransactorSession struct {
	Contract     *ProviderTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ProviderRaw is an auto generated low-level Go binding around an Ethereum contract.
type ProviderRaw struct {
	Contract *Provider // Generic contract binding to access the raw methods on
}

// ProviderCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ProviderCallerRaw struct {
	Contract *ProviderCaller // Generic read-only contract binding to access the raw methods on
}

// ProviderTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ProviderTransactorRaw struct {
	Contract *ProviderTransactor // Generic write-only contract binding to access the raw methods on
}

// NewProvider creates a new instance of Provider, bound to a specific deployed contract.
func NewProvider(address common.Address, backend bind.ContractBackend) (*Provider, error) {
	contract, err := bindProvider(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Provider{ProviderCaller: ProviderCaller{contract: contract}, ProviderTransactor: ProviderTransactor{contract: contract}, ProviderFilterer: ProviderFilterer{contract: contract}}, nil
}

// NewProviderCaller creates a new read-only instance of Provider, bound to a specific deployed contract.
func NewProviderCaller(address common.Address, caller bind.ContractCaller) (*ProviderCaller, error) {
	contract, err := bindProvider(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ProviderCaller{contract: contract}, nil
}

// NewProviderTransactor creates a new write-only instance of Provider, bound to a specific deployed contract.
func NewProviderTransactor(address common.Address, transactor bind.ContractTransactor) (*ProviderTransactor, error) {
	contract, err := bindProvider(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ProviderTransactor{contract: contract}, nil
}

// NewProviderFilterer creates a new log filterer instance of Provider, bound to a specific deployed contract.
func NewProviderFilterer(address common.Address, filterer bind.ContractFilterer) (*ProviderFilterer, error) {
	contract, err := bindProvider(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ProviderFilterer{contract: contract}, nil
}

// bindProvider binds a generic wrapper to an already deployed contract.
func bindProvider(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ProviderABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Provider *ProviderRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Provider.Contract.ProviderCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Provider *ProviderRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Provider.Contract.ProviderTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Provider *ProviderRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Provider.Contract.ProviderTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Provider *ProviderCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Provider.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Provider *ProviderTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Provider.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Provider *ProviderTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Provider.Contract.contract.Transact(opts, method, params...)
}

// GetAllAddress is a free data retrieval call binding the contract method 0x715b208b.
//
// Solidity: function getAllAddress() constant returns(address[])
func (_Provider *ProviderCaller) GetAllAddress(opts *bind.CallOpts) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _Provider.contract.Call(opts, out, "getAllAddress")
	return *ret0, err
}

// GetAllAddress is a free data retrieval call binding the contract method 0x715b208b.
//
// Solidity: function getAllAddress() constant returns(address[])
func (_Provider *ProviderSession) GetAllAddress() ([]common.Address, error) {
	return _Provider.Contract.GetAllAddress(&_Provider.CallOpts)
}

// GetAllAddress is a free data retrieval call binding the contract method 0x715b208b.
//
// Solidity: function getAllAddress() constant returns(address[])
func (_Provider *ProviderCallerSession) GetAllAddress() ([]common.Address, error) {
	return _Provider.Contract.GetAllAddress(&_Provider.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Provider *ProviderCaller) GetOwner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Provider.contract.Call(opts, out, "getOwner")
	return *ret0, err
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Provider *ProviderSession) GetOwner() (common.Address, error) {
	return _Provider.Contract.GetOwner(&_Provider.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Provider *ProviderCallerSession) GetOwner() (common.Address, error) {
	return _Provider.Contract.GetOwner(&_Provider.CallOpts)
}

// GetPrice is a free data retrieval call binding the contract method 0x98d5fdca.
//
// Solidity: function getPrice() constant returns(uint256)
func (_Provider *ProviderCaller) GetPrice(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Provider.contract.Call(opts, out, "getPrice")
	return *ret0, err
}

// GetPrice is a free data retrieval call binding the contract method 0x98d5fdca.
//
// Solidity: function getPrice() constant returns(uint256)
func (_Provider *ProviderSession) GetPrice() (*big.Int, error) {
	return _Provider.Contract.GetPrice(&_Provider.CallOpts)
}

// GetPrice is a free data retrieval call binding the contract method 0x98d5fdca.
//
// Solidity: function getPrice() constant returns(uint256)
func (_Provider *ProviderCallerSession) GetPrice() (*big.Int, error) {
	return _Provider.Contract.GetPrice(&_Provider.CallOpts)
}

// Info is a free data retrieval call binding the contract method 0x0aae7a6b.
//
// Solidity: function info(address acc) constant returns(bool, bool, uint256, uint256)
func (_Provider *ProviderCaller) Info(opts *bind.CallOpts, acc common.Address) (bool, bool, *big.Int, *big.Int, error) {
	var (
		ret0 = new(bool)
		ret1 = new(bool)
		ret2 = new(*big.Int)
		ret3 = new(*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
		ret3,
	}
	err := _Provider.contract.Call(opts, out, "info", acc)
	return *ret0, *ret1, *ret2, *ret3, err
}

// Info is a free data retrieval call binding the contract method 0x0aae7a6b.
//
// Solidity: function info(address acc) constant returns(bool, bool, uint256, uint256)
func (_Provider *ProviderSession) Info(acc common.Address) (bool, bool, *big.Int, *big.Int, error) {
	return _Provider.Contract.Info(&_Provider.CallOpts, acc)
}

// Info is a free data retrieval call binding the contract method 0x0aae7a6b.
//
// Solidity: function info(address acc) constant returns(bool, bool, uint256, uint256)
func (_Provider *ProviderCallerSession) Info(acc common.Address) (bool, bool, *big.Int, *big.Int, error) {
	return _Provider.Contract.Info(&_Provider.CallOpts, acc)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Provider *ProviderTransactor) AlterOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "alterOwner", newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Provider *ProviderSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Provider.Contract.AlterOwner(&_Provider.TransactOpts, newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Provider *ProviderTransactorSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Provider.Contract.AlterOwner(&_Provider.TransactOpts, newOwner)
}

// CancelPledge is a paid mutator transaction binding the contract method 0xe3685c40.
//
// Solidity: function cancelPledge(address acc, uint256 sum, bool status) returns(bool)
func (_Provider *ProviderTransactor) CancelPledge(opts *bind.TransactOpts, acc common.Address, sum *big.Int, status bool) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "cancelPledge", acc, sum, status)
}

// CancelPledge is a paid mutator transaction binding the contract method 0xe3685c40.
//
// Solidity: function cancelPledge(address acc, uint256 sum, bool status) returns(bool)
func (_Provider *ProviderSession) CancelPledge(acc common.Address, sum *big.Int, status bool) (*types.Transaction, error) {
	return _Provider.Contract.CancelPledge(&_Provider.TransactOpts, acc, sum, status)
}

// CancelPledge is a paid mutator transaction binding the contract method 0xe3685c40.
//
// Solidity: function cancelPledge(address acc, uint256 sum, bool status) returns(bool)
func (_Provider *ProviderTransactorSession) CancelPledge(acc common.Address, sum *big.Int, status bool) (*types.Transaction, error) {
	return _Provider.Contract.CancelPledge(&_Provider.TransactOpts, acc, sum, status)
}

// CancelPledgeStatus is a paid mutator transaction binding the contract method 0xae5e2666.
//
// Solidity: function cancelPledgeStatus() returns(bool)
func (_Provider *ProviderTransactor) CancelPledgeStatus(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "cancelPledgeStatus")
}

// CancelPledgeStatus is a paid mutator transaction binding the contract method 0xae5e2666.
//
// Solidity: function cancelPledgeStatus() returns(bool)
func (_Provider *ProviderSession) CancelPledgeStatus() (*types.Transaction, error) {
	return _Provider.Contract.CancelPledgeStatus(&_Provider.TransactOpts)
}

// CancelPledgeStatus is a paid mutator transaction binding the contract method 0xae5e2666.
//
// Solidity: function cancelPledgeStatus() returns(bool)
func (_Provider *ProviderTransactorSession) CancelPledgeStatus() (*types.Transaction, error) {
	return _Provider.Contract.CancelPledgeStatus(&_Provider.TransactOpts)
}

// Pledge is a paid mutator transaction binding the contract method 0x7326c9c0.
//
// Solidity: function pledge(uint256 _size) returns(bool)
func (_Provider *ProviderTransactor) Pledge(opts *bind.TransactOpts, _size *big.Int) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "pledge", _size)
}

// Pledge is a paid mutator transaction binding the contract method 0x7326c9c0.
//
// Solidity: function pledge(uint256 _size) returns(bool)
func (_Provider *ProviderSession) Pledge(_size *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.Pledge(&_Provider.TransactOpts, _size)
}

// Pledge is a paid mutator transaction binding the contract method 0x7326c9c0.
//
// Solidity: function pledge(uint256 _size) returns(bool)
func (_Provider *ProviderTransactorSession) Pledge(_size *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.Pledge(&_Provider.TransactOpts, _size)
}

// Set is a paid mutator transaction binding the contract method 0x35e3b25a.
//
// Solidity: function set(address addr, bool status) returns(bool)
func (_Provider *ProviderTransactor) Set(opts *bind.TransactOpts, addr common.Address, status bool) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "set", addr, status)
}

// Set is a paid mutator transaction binding the contract method 0x35e3b25a.
//
// Solidity: function set(address addr, bool status) returns(bool)
func (_Provider *ProviderSession) Set(addr common.Address, status bool) (*types.Transaction, error) {
	return _Provider.Contract.Set(&_Provider.TransactOpts, addr, status)
}

// Set is a paid mutator transaction binding the contract method 0x35e3b25a.
//
// Solidity: function set(address addr, bool status) returns(bool)
func (_Provider *ProviderTransactorSession) Set(addr common.Address, status bool) (*types.Transaction, error) {
	return _Provider.Contract.Set(&_Provider.TransactOpts, addr, status)
}

// SetBanned is a paid mutator transaction binding the contract method 0x88c9bcce.
//
// Solidity: function setBanned(address addr, bool status) returns(bool)
func (_Provider *ProviderTransactor) SetBanned(opts *bind.TransactOpts, addr common.Address, status bool) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "setBanned", addr, status)
}

// SetBanned is a paid mutator transaction binding the contract method 0x88c9bcce.
//
// Solidity: function setBanned(address addr, bool status) returns(bool)
func (_Provider *ProviderSession) SetBanned(addr common.Address, status bool) (*types.Transaction, error) {
	return _Provider.Contract.SetBanned(&_Provider.TransactOpts, addr, status)
}

// SetBanned is a paid mutator transaction binding the contract method 0x88c9bcce.
//
// Solidity: function setBanned(address addr, bool status) returns(bool)
func (_Provider *ProviderTransactorSession) SetBanned(addr common.Address, status bool) (*types.Transaction, error) {
	return _Provider.Contract.SetBanned(&_Provider.TransactOpts, addr, status)
}

// SetPrice is a paid mutator transaction binding the contract method 0x91b7f5ed.
//
// Solidity: function setPrice(uint256 _price) returns(bool)
func (_Provider *ProviderTransactor) SetPrice(opts *bind.TransactOpts, _price *big.Int) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "setPrice", _price)
}

// SetPrice is a paid mutator transaction binding the contract method 0x91b7f5ed.
//
// Solidity: function setPrice(uint256 _price) returns(bool)
func (_Provider *ProviderSession) SetPrice(_price *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.SetPrice(&_Provider.TransactOpts, _price)
}

// SetPrice is a paid mutator transaction binding the contract method 0x91b7f5ed.
//
// Solidity: function setPrice(uint256 _price) returns(bool)
func (_Provider *ProviderTransactorSession) SetPrice(_price *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.SetPrice(&_Provider.TransactOpts, _price)
}

// ProviderAlterOwnerIterator is returned from FilterAlterOwner and is used to iterate over the raw logs and unpacked data for AlterOwner events raised by the Provider contract.
type ProviderAlterOwnerIterator struct {
	Event *ProviderAlterOwner // Event containing the contract specifics and raw log

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
func (it *ProviderAlterOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderAlterOwner)
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
		it.Event = new(ProviderAlterOwner)
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
func (it *ProviderAlterOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderAlterOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderAlterOwner represents a AlterOwner event raised by the Provider contract.
type ProviderAlterOwner struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAlterOwner is a free log retrieval operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Provider *ProviderFilterer) FilterAlterOwner(opts *bind.FilterOpts) (*ProviderAlterOwnerIterator, error) {

	logs, sub, err := _Provider.contract.FilterLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return &ProviderAlterOwnerIterator{contract: _Provider.contract, event: "AlterOwner", logs: logs, sub: sub}, nil
}

// WatchAlterOwner is a free log subscription operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Provider *ProviderFilterer) WatchAlterOwner(opts *bind.WatchOpts, sink chan<- *ProviderAlterOwner) (event.Subscription, error) {

	logs, sub, err := _Provider.contract.WatchLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderAlterOwner)
				if err := _Provider.contract.UnpackLog(event, "AlterOwner", log); err != nil {
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
func (_Provider *ProviderFilterer) ParseAlterOwner(log types.Log) (*ProviderAlterOwner, error) {
	event := new(ProviderAlterOwner)
	if err := _Provider.contract.UnpackLog(event, "AlterOwner", log); err != nil {
		return nil, err
	}
	return event, nil
}
