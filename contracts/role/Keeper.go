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

// KeeperABI is the input ABI used to generate the binding from.
const KeeperABI = "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_price\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"acc\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"sum\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"cancelPledge\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"cancelPledgeStatus\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAllAddress\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getPrice\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"info\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pledge\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"set\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"setBanned\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_price\",\"type\":\"uint256\"}],\"name\":\"setPrice\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// KeeperBin is the compiled bytecode used for deploying new contracts.
var KeeperBin = "0x608060405234801561001057600080fd5b506040516113ea3803806113ea8339818101604052602081101561003357600080fd5b8101908080519060200190929190505050336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550806002819055505061134f8061009b6000396000f3fe60806040526004361061009c5760003560e01c806388ffe8671161006457806388ffe867146102e2578063893d20e81461030457806391b7f5ed1461035b57806398d5fdca14610396578063ae5e2666146103c1578063e3685c40146103e35761009c565b80630aae7a6b146100a15780630ca05f9f1461012357806335e3b25a1461018c578063715b208b1461020157806388c9bcce1461026d575b600080fd5b3480156100ad57600080fd5b506100f0600480360360208110156100c457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610455565b60405180851515151581526020018415151515815260200183815260200182815260200194505050505060405180910390f35b34801561012f57600080fd5b506101726004803603602081101561014657600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610537565b604051808215151515815260200191505060405180910390f35b34801561019857600080fd5b506101e7600480360360408110156101af57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803515159060200190929190505050610702565b604051808215151515815260200191505060405180910390f35b34801561020d57600080fd5b50610216610963565b6040518080602001828103825283818151815260200191508051906020019060200280838360005b8381101561025957808201518184015260208101905061023e565b505050509050019250505060405180910390f35b34801561027957600080fd5b506102c86004803603604081101561029057600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803515159060200190929190505050610b75565b604051808215151515815260200191505060405180910390f35b6102ea610da5565b604051808215151515815260200191505060405180910390f35b34801561031057600080fd5b50610319610efa565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34801561036757600080fd5b506103946004803603602081101561037e57600080fd5b8101908080359060200190929190505050610f23565b005b3480156103a257600080fd5b506103ab610fef565b6040518082815260200191505060405180910390f35b6103c9610ff9565b604051808215151515815260200191505060405180910390f35b61043b600480360360608110156103f957600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919080351515906020019092919050505061105e565b604051808215151515815260200191505060405180910390f35b60008060008060006104668661126d565b90506001805490508114156104905760006001600080819150809050945094509450945050610530565b6001818154811061049d57fe5b906000526020600020906003020160000160149054906101000a900460ff16600182815481106104c957fe5b906000526020600020906003020160000160159054906101000a900460ff16600183815481106104f557fe5b9060005260206000209060030201600101546001848154811061051457fe5b9060005260206000209060030201600201549450945094509450505b9193509193565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146105fb576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050826000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e908184604051808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a16001915050919050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146107c6576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b60006107d18461126d565b9050600180549050811461084b57600181815481106107ec57fe5b906000526020600020906003020160000160159054906101000a900460ff1661084657826001828154811061081d57fe5b906000526020600020906003020160000160146101000a81548160ff0219169083151502179055505b610958565b60016040518060a001604052808673ffffffffffffffffffffffffffffffffffffffff1681526020018515158152602001600015158152602001600081526020016000815250908060018154018082558091505060019003906000526020600020906003020160009091909190915060008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060208201518160000160146101000a81548160ff02191690831515021790555060408201518160000160156101000a81548160ff021916908315150217905550606082015181600101556080820151816002015550505b600191505092915050565b60608060018054905060405190808252806020026020018201604052801561099a5781602001602082028036833780820191505090505b509050600080905060008090505b600180549050811015610ac15760011515600182815481106109c657fe5b906000526020600020906003020160000160149054906101000a900460ff1615151415610ab45760001515600182815481106109fe57fe5b906000526020600020906003020160000160159054906101000a900460ff1615151415610ab35760018181548110610a3257fe5b906000526020600020906003020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16838381518110610a7057fe5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505081806001019250505b5b80806001019150506109a8565b50606081604051908082528060200260200182016040528015610af35781602001602082028036833780820191505090505b50905060008090505b82811015610b6b57838181518110610b1057fe5b6020026020010151828281518110610b2457fe5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff16815250508080600101915050610afc565b5080935050505090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610c39576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b6000610c448461126d565b90506001805490508114610c8d578260018281548110610c6057fe5b906000526020600020906003020160000160156101000a81548160ff021916908315150217905550610d9a565b60016040518060a001604052808673ffffffffffffffffffffffffffffffffffffffff1681526020016000151581526020018515158152602001600081526020016000815250908060018154018082558091505060019003906000526020600020906003020160009091909190915060008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060208201518160000160146101000a81548160ff02191690831515021790555060408201518160000160156101000a81548160ff021916908315150217905550606082015181600101556080820151816002015550505b600191505092915050565b6000600254341015610e01573373ffffffffffffffffffffffffffffffffffffffff166108fc349081150290604051600060405180830381858888f19350505050158015610df7573d6000803e3d6000fd5b5060009050610ef7565b6000610e0c3361126d565b9050600180549050811415610e6c573373ffffffffffffffffffffffffffffffffffffffff166108fc349081150290604051600060405180830381858888f19350505050158015610e61573d6000803e3d6000fd5b506000915050610ef7565b6001808281548110610e7a57fe5b906000526020600020906003020160000160146101000a81548160ff0219169083151502179055503460018281548110610eb057fe5b9060005260206000209060030201600101600082825401925050819055504260018281548110610edc57fe5b90600052602060002090600302016002018190555060019150505b90565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610fe5576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b8060028190555050565b6000600254905090565b6000806110053361126d565b905060018054905081141561101e57600091505061105b565b60006001828154811061102d57fe5b906000526020600020906003020160000160146101000a81548160ff02191690831515021790555060019150505b90565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614611122576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b600061112d8561126d565b9050600180549050811415611146576000915050611266565b60006001828154811061115557fe5b906000526020600020906003020160010154851115611194576001828154811061117b57fe5b9060005260206000209060030201600101549050611198565b8490505b60008114156111ac57600092505050611266565b80600183815481106111ba57fe5b906000526020600020906003020160010160008282540392505081905550831561125f5760001515600183815481106111ef57fe5b906000526020600020906003020160000160159054906101000a900460ff161515141561125e578573ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f1935050505015801561125c573d6000803e3d6000fd5b505b5b6001925050505b9392505050565b600080600090505b60018054905081101561130a578273ffffffffffffffffffffffffffffffffffffffff16600182815481106112a657fe5b906000526020600020906003020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614156112fd5780915050611314565b8080600101915050611275565b5060018054905090505b91905056fea2646970667358221220a9e98e9a8ed9e71ec822646f6ecc9acd55d4bea685a450890518a6e93c358e4364736f6c63430006030033"

// DeployKeeper deploys a new Ethereum contract, binding an instance of Keeper to it.
func DeployKeeper(auth *bind.TransactOpts, backend bind.ContractBackend, _price *big.Int) (common.Address, *types.Transaction, *Keeper, error) {
	parsed, err := abi.JSON(strings.NewReader(KeeperABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(KeeperBin), backend, _price)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Keeper{KeeperCaller: KeeperCaller{contract: contract}, KeeperTransactor: KeeperTransactor{contract: contract}, KeeperFilterer: KeeperFilterer{contract: contract}}, nil
}

// Keeper is an auto generated Go binding around an Ethereum contract.
type Keeper struct {
	KeeperCaller     // Read-only binding to the contract
	KeeperTransactor // Write-only binding to the contract
	KeeperFilterer   // Log filterer for contract events
}

// KeeperCaller is an auto generated read-only Go binding around an Ethereum contract.
type KeeperCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KeeperTransactor is an auto generated write-only Go binding around an Ethereum contract.
type KeeperTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KeeperFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type KeeperFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KeeperSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type KeeperSession struct {
	Contract     *Keeper           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// KeeperCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type KeeperCallerSession struct {
	Contract *KeeperCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// KeeperTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type KeeperTransactorSession struct {
	Contract     *KeeperTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// KeeperRaw is an auto generated low-level Go binding around an Ethereum contract.
type KeeperRaw struct {
	Contract *Keeper // Generic contract binding to access the raw methods on
}

// KeeperCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type KeeperCallerRaw struct {
	Contract *KeeperCaller // Generic read-only contract binding to access the raw methods on
}

// KeeperTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type KeeperTransactorRaw struct {
	Contract *KeeperTransactor // Generic write-only contract binding to access the raw methods on
}

// NewKeeper creates a new instance of Keeper, bound to a specific deployed contract.
func NewKeeper(address common.Address, backend bind.ContractBackend) (*Keeper, error) {
	contract, err := bindKeeper(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Keeper{KeeperCaller: KeeperCaller{contract: contract}, KeeperTransactor: KeeperTransactor{contract: contract}, KeeperFilterer: KeeperFilterer{contract: contract}}, nil
}

// NewKeeperCaller creates a new read-only instance of Keeper, bound to a specific deployed contract.
func NewKeeperCaller(address common.Address, caller bind.ContractCaller) (*KeeperCaller, error) {
	contract, err := bindKeeper(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &KeeperCaller{contract: contract}, nil
}

// NewKeeperTransactor creates a new write-only instance of Keeper, bound to a specific deployed contract.
func NewKeeperTransactor(address common.Address, transactor bind.ContractTransactor) (*KeeperTransactor, error) {
	contract, err := bindKeeper(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &KeeperTransactor{contract: contract}, nil
}

// NewKeeperFilterer creates a new log filterer instance of Keeper, bound to a specific deployed contract.
func NewKeeperFilterer(address common.Address, filterer bind.ContractFilterer) (*KeeperFilterer, error) {
	contract, err := bindKeeper(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &KeeperFilterer{contract: contract}, nil
}

// bindKeeper binds a generic wrapper to an already deployed contract.
func bindKeeper(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(KeeperABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Keeper *KeeperRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Keeper.Contract.KeeperCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Keeper *KeeperRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Keeper.Contract.KeeperTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Keeper *KeeperRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Keeper.Contract.KeeperTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Keeper *KeeperCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Keeper.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Keeper *KeeperTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Keeper.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Keeper *KeeperTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Keeper.Contract.contract.Transact(opts, method, params...)
}

// GetAllAddress is a free data retrieval call binding the contract method 0x715b208b.
//
// Solidity: function getAllAddress() constant returns(address[])
func (_Keeper *KeeperCaller) GetAllAddress(opts *bind.CallOpts) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _Keeper.contract.Call(opts, out, "getAllAddress")
	return *ret0, err
}

// GetAllAddress is a free data retrieval call binding the contract method 0x715b208b.
//
// Solidity: function getAllAddress() constant returns(address[])
func (_Keeper *KeeperSession) GetAllAddress() ([]common.Address, error) {
	return _Keeper.Contract.GetAllAddress(&_Keeper.CallOpts)
}

// GetAllAddress is a free data retrieval call binding the contract method 0x715b208b.
//
// Solidity: function getAllAddress() constant returns(address[])
func (_Keeper *KeeperCallerSession) GetAllAddress() ([]common.Address, error) {
	return _Keeper.Contract.GetAllAddress(&_Keeper.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Keeper *KeeperCaller) GetOwner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Keeper.contract.Call(opts, out, "getOwner")
	return *ret0, err
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Keeper *KeeperSession) GetOwner() (common.Address, error) {
	return _Keeper.Contract.GetOwner(&_Keeper.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_Keeper *KeeperCallerSession) GetOwner() (common.Address, error) {
	return _Keeper.Contract.GetOwner(&_Keeper.CallOpts)
}

// GetPrice is a free data retrieval call binding the contract method 0x98d5fdca.
//
// Solidity: function getPrice() constant returns(uint256)
func (_Keeper *KeeperCaller) GetPrice(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Keeper.contract.Call(opts, out, "getPrice")
	return *ret0, err
}

// GetPrice is a free data retrieval call binding the contract method 0x98d5fdca.
//
// Solidity: function getPrice() constant returns(uint256)
func (_Keeper *KeeperSession) GetPrice() (*big.Int, error) {
	return _Keeper.Contract.GetPrice(&_Keeper.CallOpts)
}

// GetPrice is a free data retrieval call binding the contract method 0x98d5fdca.
//
// Solidity: function getPrice() constant returns(uint256)
func (_Keeper *KeeperCallerSession) GetPrice() (*big.Int, error) {
	return _Keeper.Contract.GetPrice(&_Keeper.CallOpts)
}

// Info is a free data retrieval call binding the contract method 0x0aae7a6b.
//
// Solidity: function info(address addr) constant returns(bool, bool, uint256, uint256)
func (_Keeper *KeeperCaller) Info(opts *bind.CallOpts, addr common.Address) (bool, bool, *big.Int, *big.Int, error) {
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
	err := _Keeper.contract.Call(opts, out, "info", addr)
	return *ret0, *ret1, *ret2, *ret3, err
}

// Info is a free data retrieval call binding the contract method 0x0aae7a6b.
//
// Solidity: function info(address addr) constant returns(bool, bool, uint256, uint256)
func (_Keeper *KeeperSession) Info(addr common.Address) (bool, bool, *big.Int, *big.Int, error) {
	return _Keeper.Contract.Info(&_Keeper.CallOpts, addr)
}

// Info is a free data retrieval call binding the contract method 0x0aae7a6b.
//
// Solidity: function info(address addr) constant returns(bool, bool, uint256, uint256)
func (_Keeper *KeeperCallerSession) Info(addr common.Address) (bool, bool, *big.Int, *big.Int, error) {
	return _Keeper.Contract.Info(&_Keeper.CallOpts, addr)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Keeper *KeeperTransactor) AlterOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Keeper.contract.Transact(opts, "alterOwner", newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Keeper *KeeperSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Keeper.Contract.AlterOwner(&_Keeper.TransactOpts, newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Keeper *KeeperTransactorSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Keeper.Contract.AlterOwner(&_Keeper.TransactOpts, newOwner)
}

// CancelPledge is a paid mutator transaction binding the contract method 0xe3685c40.
//
// Solidity: function cancelPledge(address acc, uint256 sum, bool status) returns(bool)
func (_Keeper *KeeperTransactor) CancelPledge(opts *bind.TransactOpts, acc common.Address, sum *big.Int, status bool) (*types.Transaction, error) {
	return _Keeper.contract.Transact(opts, "cancelPledge", acc, sum, status)
}

// CancelPledge is a paid mutator transaction binding the contract method 0xe3685c40.
//
// Solidity: function cancelPledge(address acc, uint256 sum, bool status) returns(bool)
func (_Keeper *KeeperSession) CancelPledge(acc common.Address, sum *big.Int, status bool) (*types.Transaction, error) {
	return _Keeper.Contract.CancelPledge(&_Keeper.TransactOpts, acc, sum, status)
}

// CancelPledge is a paid mutator transaction binding the contract method 0xe3685c40.
//
// Solidity: function cancelPledge(address acc, uint256 sum, bool status) returns(bool)
func (_Keeper *KeeperTransactorSession) CancelPledge(acc common.Address, sum *big.Int, status bool) (*types.Transaction, error) {
	return _Keeper.Contract.CancelPledge(&_Keeper.TransactOpts, acc, sum, status)
}

// CancelPledgeStatus is a paid mutator transaction binding the contract method 0xae5e2666.
//
// Solidity: function cancelPledgeStatus() returns(bool)
func (_Keeper *KeeperTransactor) CancelPledgeStatus(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Keeper.contract.Transact(opts, "cancelPledgeStatus")
}

// CancelPledgeStatus is a paid mutator transaction binding the contract method 0xae5e2666.
//
// Solidity: function cancelPledgeStatus() returns(bool)
func (_Keeper *KeeperSession) CancelPledgeStatus() (*types.Transaction, error) {
	return _Keeper.Contract.CancelPledgeStatus(&_Keeper.TransactOpts)
}

// CancelPledgeStatus is a paid mutator transaction binding the contract method 0xae5e2666.
//
// Solidity: function cancelPledgeStatus() returns(bool)
func (_Keeper *KeeperTransactorSession) CancelPledgeStatus() (*types.Transaction, error) {
	return _Keeper.Contract.CancelPledgeStatus(&_Keeper.TransactOpts)
}

// Pledge is a paid mutator transaction binding the contract method 0x88ffe867.
//
// Solidity: function pledge() returns(bool)
func (_Keeper *KeeperTransactor) Pledge(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Keeper.contract.Transact(opts, "pledge")
}

// Pledge is a paid mutator transaction binding the contract method 0x88ffe867.
//
// Solidity: function pledge() returns(bool)
func (_Keeper *KeeperSession) Pledge() (*types.Transaction, error) {
	return _Keeper.Contract.Pledge(&_Keeper.TransactOpts)
}

// Pledge is a paid mutator transaction binding the contract method 0x88ffe867.
//
// Solidity: function pledge() returns(bool)
func (_Keeper *KeeperTransactorSession) Pledge() (*types.Transaction, error) {
	return _Keeper.Contract.Pledge(&_Keeper.TransactOpts)
}

// Set is a paid mutator transaction binding the contract method 0x35e3b25a.
//
// Solidity: function set(address addr, bool status) returns(bool)
func (_Keeper *KeeperTransactor) Set(opts *bind.TransactOpts, addr common.Address, status bool) (*types.Transaction, error) {
	return _Keeper.contract.Transact(opts, "set", addr, status)
}

// Set is a paid mutator transaction binding the contract method 0x35e3b25a.
//
// Solidity: function set(address addr, bool status) returns(bool)
func (_Keeper *KeeperSession) Set(addr common.Address, status bool) (*types.Transaction, error) {
	return _Keeper.Contract.Set(&_Keeper.TransactOpts, addr, status)
}

// Set is a paid mutator transaction binding the contract method 0x35e3b25a.
//
// Solidity: function set(address addr, bool status) returns(bool)
func (_Keeper *KeeperTransactorSession) Set(addr common.Address, status bool) (*types.Transaction, error) {
	return _Keeper.Contract.Set(&_Keeper.TransactOpts, addr, status)
}

// SetBanned is a paid mutator transaction binding the contract method 0x88c9bcce.
//
// Solidity: function setBanned(address addr, bool status) returns(bool)
func (_Keeper *KeeperTransactor) SetBanned(opts *bind.TransactOpts, addr common.Address, status bool) (*types.Transaction, error) {
	return _Keeper.contract.Transact(opts, "setBanned", addr, status)
}

// SetBanned is a paid mutator transaction binding the contract method 0x88c9bcce.
//
// Solidity: function setBanned(address addr, bool status) returns(bool)
func (_Keeper *KeeperSession) SetBanned(addr common.Address, status bool) (*types.Transaction, error) {
	return _Keeper.Contract.SetBanned(&_Keeper.TransactOpts, addr, status)
}

// SetBanned is a paid mutator transaction binding the contract method 0x88c9bcce.
//
// Solidity: function setBanned(address addr, bool status) returns(bool)
func (_Keeper *KeeperTransactorSession) SetBanned(addr common.Address, status bool) (*types.Transaction, error) {
	return _Keeper.Contract.SetBanned(&_Keeper.TransactOpts, addr, status)
}

// SetPrice is a paid mutator transaction binding the contract method 0x91b7f5ed.
//
// Solidity: function setPrice(uint256 _price) returns()
func (_Keeper *KeeperTransactor) SetPrice(opts *bind.TransactOpts, _price *big.Int) (*types.Transaction, error) {
	return _Keeper.contract.Transact(opts, "setPrice", _price)
}

// SetPrice is a paid mutator transaction binding the contract method 0x91b7f5ed.
//
// Solidity: function setPrice(uint256 _price) returns()
func (_Keeper *KeeperSession) SetPrice(_price *big.Int) (*types.Transaction, error) {
	return _Keeper.Contract.SetPrice(&_Keeper.TransactOpts, _price)
}

// SetPrice is a paid mutator transaction binding the contract method 0x91b7f5ed.
//
// Solidity: function setPrice(uint256 _price) returns()
func (_Keeper *KeeperTransactorSession) SetPrice(_price *big.Int) (*types.Transaction, error) {
	return _Keeper.Contract.SetPrice(&_Keeper.TransactOpts, _price)
}

// KeeperAlterOwnerIterator is returned from FilterAlterOwner and is used to iterate over the raw logs and unpacked data for AlterOwner events raised by the Keeper contract.
type KeeperAlterOwnerIterator struct {
	Event *KeeperAlterOwner // Event containing the contract specifics and raw log

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
func (it *KeeperAlterOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperAlterOwner)
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
		it.Event = new(KeeperAlterOwner)
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
func (it *KeeperAlterOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperAlterOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperAlterOwner represents a AlterOwner event raised by the Keeper contract.
type KeeperAlterOwner struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAlterOwner is a free log retrieval operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Keeper *KeeperFilterer) FilterAlterOwner(opts *bind.FilterOpts) (*KeeperAlterOwnerIterator, error) {

	logs, sub, err := _Keeper.contract.FilterLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return &KeeperAlterOwnerIterator{contract: _Keeper.contract, event: "AlterOwner", logs: logs, sub: sub}, nil
}

// WatchAlterOwner is a free log subscription operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Keeper *KeeperFilterer) WatchAlterOwner(opts *bind.WatchOpts, sink chan<- *KeeperAlterOwner) (event.Subscription, error) {

	logs, sub, err := _Keeper.contract.WatchLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperAlterOwner)
				if err := _Keeper.contract.UnpackLog(event, "AlterOwner", log); err != nil {
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
func (_Keeper *KeeperFilterer) ParseAlterOwner(log types.Log) (*KeeperAlterOwner, error) {
	event := new(KeeperAlterOwner)
	if err := _Keeper.contract.UnpackLog(event, "AlterOwner", log); err != nil {
		return nil, err
	}
	return event, nil
}
