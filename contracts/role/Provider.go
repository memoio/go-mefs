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
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// ProviderABI is the input ABI used to generate the binding from.
const ProviderABI = "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_price\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"}],\"name\":\"CancelPledgeStatus\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"acc\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"sum\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"cancelPledge\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"cancelPledgeStatus\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAllAddress\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getPrice\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"acc\",\"type\":\"address\"}],\"name\":\"info\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_size\",\"type\":\"uint256\"}],\"name\":\"pledge\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"set\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"setBanned\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_price\",\"type\":\"uint256\"}],\"name\":\"setPrice\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"transferPledge\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"payable\",\"type\":\"function\"}]"

// ProviderBin is the compiled bytecode used for deploying new contracts.
var ProviderBin = "0x6080604052738026796fd7ce63eae824314aa5bacf55643e893d600360006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555034801561006557600080fd5b506040516118bc3803806118bc8339818101604052602081101561008857600080fd5b8101908080519060200190929190505050336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555080600281905550506117cc806100f06000396000f3fe6080604052600436106100a75760003560e01c806388c9bcce1161006457806388c9bcce1461030e578063893d20e81461038157806391b7f5ed146103c257806398d5fdca14610413578063ae5e26661461043e578063e3685c401461045e576100a7565b80630aae7a6b146100ac5780630ca05f9f1461012a57806335e3b25a146101915780636cb0e8e314610204578063715b208b1461025e5780637326c9c0146102ca575b600080fd5b3480156100b857600080fd5b506100fb600480360360208110156100cf57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506104ce565b604051808515158152602001841515815260200183815260200182815260200194505050505060405180910390f35b34801561013657600080fd5b506101796004803603602081101561014d57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506105aa565b60405180821515815260200191505060405180910390f35b34801561019d57600080fd5b506101ec600480360360408110156101b457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803515159060200190929190505050610749565b60405180821515815260200191505060405180910390f35b6102466004803603602081101561021a57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506109b2565b60405180821515815260200191505060405180910390f35b34801561026a57600080fd5b50610273610acc565b6040518080602001828103825283818151815260200191508051906020019060200280838360005b838110156102b657808201518184015260208101905061029b565b505050509050019250505060405180910390f35b6102f6600480360360208110156102e057600080fd5b8101908080359060200190929190505050610d02565b60405180821515815260200191505060405180910390f35b34801561031a57600080fd5b506103696004803603604081101561033157600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803515159060200190929190505050610fcb565b60405180821515815260200191505060405180910390f35b34801561038d57600080fd5b506103966111fb565b604051808273ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b3480156103ce57600080fd5b506103fb600480360360208110156103e557600080fd5b8101908080359060200190929190505050611224565b60405180821515815260200191505060405180910390f35b34801561041f57600080fd5b506104286112f8565b6040518082815260200191505060405180910390f35b610446611302565b60405180821515815260200191505060405180910390f35b6104b66004803603606081101561047457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291908035151590602001909291905050506114db565b60405180821515815260200191505060405180910390f35b60008060008060006104df866116ea565b905060018054905081141561050357600060016000809450945094509450506105a3565b6001818154811061051057fe5b906000526020600020906003020160000160149054906101000a900460ff166001828154811061053c57fe5b906000526020600020906003020160000160159054906101000a900460ff166001838154811061056857fe5b9060005260206000209060030201600101546001848154811061058757fe5b9060005260206000209060030201600201549450945094509450505b9193509193565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461066e576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050826000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e908184604051808373ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a16001915050919050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461080d576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b6000610818846116ea565b9050600180549050811461089a57600015156001828154811061083757fe5b906000526020600020906003020160000160159054906101000a900460ff161515141561089557826001828154811061086c57fe5b906000526020600020906003020160000160146101000a81548160ff0219169083151502179055505b6109a7565b60016040518060a001604052808673ffffffffffffffffffffffffffffffffffffffff1681526020018515158152602001600015158152602001600081526020016000815250908060018154018082558091505060019003906000526020600020906003020160009091909190915060008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060208201518160000160146101000a81548160ff02191690831515021790555060408201518160000160156101000a81548160ff021916908315150217905550606082015181600101556080820151816002015550505b600191505092915050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610a76576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b60004790508273ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f19350505050158015610ac1573d6000803e3d6000fd5b506001915050919050565b60608060018054905067ffffffffffffffff81118015610aeb57600080fd5b50604051908082528060200260200182016040528015610b1a5781602001602082028036833780820191505090505b5090506000805b600180549050811015610c3a576001151560018281548110610b3f57fe5b906000526020600020906003020160000160149054906101000a900460ff1615151415610c2d576000151560018281548110610b7757fe5b906000526020600020906003020160000160159054906101000a900460ff1615151415610c2c5760018181548110610bab57fe5b906000526020600020906003020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16838381518110610be957fe5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505081806001019250505b5b8080600101915050610b21565b5060608167ffffffffffffffff81118015610c5457600080fd5b50604051908082528060200260200182016040528015610c835781602001602082028036833780820191505090505b50905060005b82811015610cf857838181518110610c9d57fe5b6020026020010151828281518110610cb157fe5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff16815250508080600101915050610c89565b5080935050505090565b600080600360009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663597e409d6040518163ffffffff1660e01b815260040160206040518083038186803b158015610d6d57600080fd5b505afa158015610d81573d6000803e3d6000fd5b505050506040513d6020811015610d9757600080fd5b81019080805190602001909291905050509050600161ffff168161ffff1610610e28576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260108152602001807f706c656467652069732062616e6e65640000000000000000000000000000000081525060200191505060405180910390fd5b6000610e33336116ea565b9050600180549050811415610e94573373ffffffffffffffffffffffffffffffffffffffff166108fc349081150290604051600060405180830381858888f19350505050158015610e88573d6000803e3d6000fd5b50600092505050610fc6565b8360025402341015610ef2573373ffffffffffffffffffffffffffffffffffffffff166108fc349081150290604051600060405180830381858888f19350505050158015610ee6573d6000803e3d6000fd5b50600092505050610fc6565b6001808281548110610f0057fe5b906000526020600020906003020160000160146101000a81548160ff02191690831515021790555060003460018381548110610f3857fe5b90600052602060002090600302016001015401905060018281548110610f5a57fe5b906000526020600020906003020160010154811015610f7857600080fd5b8060018381548110610f8657fe5b9060005260206000209060030201600101819055504260018381548110610fa957fe5b906000526020600020906003020160020181905550600193505050505b919050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461108f576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b600061109a846116ea565b905060018054905081146110e35782600182815481106110b657fe5b906000526020600020906003020160000160156101000a81548160ff0219169083151502179055506111f0565b60016040518060a001604052808673ffffffffffffffffffffffffffffffffffffffff1681526020016000151581526020018515158152602001600081526020016000815250908060018154018082558091505060019003906000526020600020906003020160009091909190915060008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060208201518160000160146101000a81548160ff02191690831515021790555060408201518160000160156101000a81548160ff021916908315150217905550606082015181600101556080820151816002015550505b600191505092915050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146112e8576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b8160028190555060019050919050565b6000600254905090565b600080600360009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663597e409d6040518163ffffffff1660e01b815260040160206040518083038186803b15801561136d57600080fd5b505afa158015611381573d6000803e3d6000fd5b505050506040513d602081101561139757600080fd5b81019080805190602001909291905050509050600161ffff168161ffff1610611428576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260098152602001807f69732062616e6e6564000000000000000000000000000000000000000000000081525060200191505060405180910390fd5b6000611433336116ea565b905060018054905081141561144d576000925050506114d8565b60006001828154811061145c57fe5b906000526020600020906003020160000160146101000a81548160ff0219169083151502179055507fdc63ce37be5b519f3d4a0c7e239641bea35e801630cac3175f46f92ee32583aa33604051808273ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390a16001925050505b90565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461159f576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b60006115aa856116ea565b90506001805490508114156115c35760009150506116e3565b6000600182815481106115d257fe5b90600052602060002090600302016001015485111561161157600182815481106115f857fe5b9060005260206000209060030201600101549050611615565b8490505b6000811415611629576000925050506116e3565b806001838154811061163757fe5b90600052602060002090600302016001016000828254039250508190555083156116dc57600015156001838154811061166c57fe5b906000526020600020906003020160000160159054906101000a900460ff16151514156116db578573ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f193505050501580156116d9573d6000803e3d6000fd5b505b5b6001925050505b9392505050565b600080600090505b600180549050811015611787578273ffffffffffffffffffffffffffffffffffffffff166001828154811061172357fe5b906000526020600020906003020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16141561177a5780915050611791565b80806001019150506116f2565b5060018054905090505b91905056fea26469706673582212208ad96dc0f00cc919aa847f306b510349d7e723c3700cb1a4d1c6612289d4157a64736f6c63430007030033"

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
// Solidity: function getAllAddress() view returns(address[])
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
// Solidity: function getAllAddress() view returns(address[])
func (_Provider *ProviderSession) GetAllAddress() ([]common.Address, error) {
	return _Provider.Contract.GetAllAddress(&_Provider.CallOpts)
}

// GetAllAddress is a free data retrieval call binding the contract method 0x715b208b.
//
// Solidity: function getAllAddress() view returns(address[])
func (_Provider *ProviderCallerSession) GetAllAddress() ([]common.Address, error) {
	return _Provider.Contract.GetAllAddress(&_Provider.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
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
// Solidity: function getOwner() view returns(address)
func (_Provider *ProviderSession) GetOwner() (common.Address, error) {
	return _Provider.Contract.GetOwner(&_Provider.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_Provider *ProviderCallerSession) GetOwner() (common.Address, error) {
	return _Provider.Contract.GetOwner(&_Provider.CallOpts)
}

// GetPrice is a free data retrieval call binding the contract method 0x98d5fdca.
//
// Solidity: function getPrice() view returns(uint256)
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
// Solidity: function getPrice() view returns(uint256)
func (_Provider *ProviderSession) GetPrice() (*big.Int, error) {
	return _Provider.Contract.GetPrice(&_Provider.CallOpts)
}

// GetPrice is a free data retrieval call binding the contract method 0x98d5fdca.
//
// Solidity: function getPrice() view returns(uint256)
func (_Provider *ProviderCallerSession) GetPrice() (*big.Int, error) {
	return _Provider.Contract.GetPrice(&_Provider.CallOpts)
}

// Info is a free data retrieval call binding the contract method 0x0aae7a6b.
//
// Solidity: function info(address acc) view returns(bool, bool, uint256, uint256)
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
// Solidity: function info(address acc) view returns(bool, bool, uint256, uint256)
func (_Provider *ProviderSession) Info(acc common.Address) (bool, bool, *big.Int, *big.Int, error) {
	return _Provider.Contract.Info(&_Provider.CallOpts, acc)
}

// Info is a free data retrieval call binding the contract method 0x0aae7a6b.
//
// Solidity: function info(address acc) view returns(bool, bool, uint256, uint256)
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
// Solidity: function cancelPledge(address acc, uint256 sum, bool status) payable returns(bool)
func (_Provider *ProviderTransactor) CancelPledge(opts *bind.TransactOpts, acc common.Address, sum *big.Int, status bool) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "cancelPledge", acc, sum, status)
}

// CancelPledge is a paid mutator transaction binding the contract method 0xe3685c40.
//
// Solidity: function cancelPledge(address acc, uint256 sum, bool status) payable returns(bool)
func (_Provider *ProviderSession) CancelPledge(acc common.Address, sum *big.Int, status bool) (*types.Transaction, error) {
	return _Provider.Contract.CancelPledge(&_Provider.TransactOpts, acc, sum, status)
}

// CancelPledge is a paid mutator transaction binding the contract method 0xe3685c40.
//
// Solidity: function cancelPledge(address acc, uint256 sum, bool status) payable returns(bool)
func (_Provider *ProviderTransactorSession) CancelPledge(acc common.Address, sum *big.Int, status bool) (*types.Transaction, error) {
	return _Provider.Contract.CancelPledge(&_Provider.TransactOpts, acc, sum, status)
}

// CancelPledgeStatus is a paid mutator transaction binding the contract method 0xae5e2666.
//
// Solidity: function cancelPledgeStatus() payable returns(bool)
func (_Provider *ProviderTransactor) CancelPledgeStatus(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "cancelPledgeStatus")
}

// CancelPledgeStatus is a paid mutator transaction binding the contract method 0xae5e2666.
//
// Solidity: function cancelPledgeStatus() payable returns(bool)
func (_Provider *ProviderSession) CancelPledgeStatus() (*types.Transaction, error) {
	return _Provider.Contract.CancelPledgeStatus(&_Provider.TransactOpts)
}

// CancelPledgeStatus is a paid mutator transaction binding the contract method 0xae5e2666.
//
// Solidity: function cancelPledgeStatus() payable returns(bool)
func (_Provider *ProviderTransactorSession) CancelPledgeStatus() (*types.Transaction, error) {
	return _Provider.Contract.CancelPledgeStatus(&_Provider.TransactOpts)
}

// Pledge is a paid mutator transaction binding the contract method 0x7326c9c0.
//
// Solidity: function pledge(uint256 _size) payable returns(bool)
func (_Provider *ProviderTransactor) Pledge(opts *bind.TransactOpts, _size *big.Int) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "pledge", _size)
}

// Pledge is a paid mutator transaction binding the contract method 0x7326c9c0.
//
// Solidity: function pledge(uint256 _size) payable returns(bool)
func (_Provider *ProviderSession) Pledge(_size *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.Pledge(&_Provider.TransactOpts, _size)
}

// Pledge is a paid mutator transaction binding the contract method 0x7326c9c0.
//
// Solidity: function pledge(uint256 _size) payable returns(bool)
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

// TransferPledge is a paid mutator transaction binding the contract method 0x6cb0e8e3.
//
// Solidity: function transferPledge(address to) payable returns(bool)
func (_Provider *ProviderTransactor) TransferPledge(opts *bind.TransactOpts, to common.Address) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "transferPledge", to)
}

// TransferPledge is a paid mutator transaction binding the contract method 0x6cb0e8e3.
//
// Solidity: function transferPledge(address to) payable returns(bool)
func (_Provider *ProviderSession) TransferPledge(to common.Address) (*types.Transaction, error) {
	return _Provider.Contract.TransferPledge(&_Provider.TransactOpts, to)
}

// TransferPledge is a paid mutator transaction binding the contract method 0x6cb0e8e3.
//
// Solidity: function transferPledge(address to) payable returns(bool)
func (_Provider *ProviderTransactorSession) TransferPledge(to common.Address) (*types.Transaction, error) {
	return _Provider.Contract.TransferPledge(&_Provider.TransactOpts, to)
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

// ProviderCancelPledgeStatusIterator is returned from FilterCancelPledgeStatus and is used to iterate over the raw logs and unpacked data for CancelPledgeStatus events raised by the Provider contract.
type ProviderCancelPledgeStatusIterator struct {
	Event *ProviderCancelPledgeStatus // Event containing the contract specifics and raw log

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
func (it *ProviderCancelPledgeStatusIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderCancelPledgeStatus)
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
		it.Event = new(ProviderCancelPledgeStatus)
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
func (it *ProviderCancelPledgeStatusIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderCancelPledgeStatusIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderCancelPledgeStatus represents a CancelPledgeStatus event raised by the Provider contract.
type ProviderCancelPledgeStatus struct {
	From common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterCancelPledgeStatus is a free log retrieval operation binding the contract event 0xdc63ce37be5b519f3d4a0c7e239641bea35e801630cac3175f46f92ee32583aa.
//
// Solidity: event CancelPledgeStatus(address from)
func (_Provider *ProviderFilterer) FilterCancelPledgeStatus(opts *bind.FilterOpts) (*ProviderCancelPledgeStatusIterator, error) {

	logs, sub, err := _Provider.contract.FilterLogs(opts, "CancelPledgeStatus")
	if err != nil {
		return nil, err
	}
	return &ProviderCancelPledgeStatusIterator{contract: _Provider.contract, event: "CancelPledgeStatus", logs: logs, sub: sub}, nil
}

// WatchCancelPledgeStatus is a free log subscription operation binding the contract event 0xdc63ce37be5b519f3d4a0c7e239641bea35e801630cac3175f46f92ee32583aa.
//
// Solidity: event CancelPledgeStatus(address from)
func (_Provider *ProviderFilterer) WatchCancelPledgeStatus(opts *bind.WatchOpts, sink chan<- *ProviderCancelPledgeStatus) (event.Subscription, error) {

	logs, sub, err := _Provider.contract.WatchLogs(opts, "CancelPledgeStatus")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderCancelPledgeStatus)
				if err := _Provider.contract.UnpackLog(event, "CancelPledgeStatus", log); err != nil {
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

// ParseCancelPledgeStatus is a log parse operation binding the contract event 0xdc63ce37be5b519f3d4a0c7e239641bea35e801630cac3175f46f92ee32583aa.
//
// Solidity: event CancelPledgeStatus(address from)
func (_Provider *ProviderFilterer) ParseCancelPledgeStatus(log types.Log) (*ProviderCancelPledgeStatus, error) {
	event := new(ProviderCancelPledgeStatus)
	if err := _Provider.contract.UnpackLog(event, "CancelPledgeStatus", log); err != nil {
		return nil, err
	}
	return event, nil
}
