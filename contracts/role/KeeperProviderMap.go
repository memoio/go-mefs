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
	//_ = abi.U256
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// KeeperProviderMapABI is the input ABI used to generate the binding from.
const KeeperProviderMapABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"data\",\"type\":\"string\"}],\"name\":\"Error\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"name\":\"keeper\",\"type\":\"address\"},{\"name\":\"provider\",\"type\":\"address[]\"}],\"name\":\"add\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"keeper\",\"type\":\"address\"}],\"name\":\"delKeeper\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"keeper\",\"type\":\"address\"},{\"name\":\"provider\",\"type\":\"address\"}],\"name\":\"delProvider\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getAllKeeper\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"keeper\",\"type\":\"address\"}],\"name\":\"getProvider\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// KeeperProviderMapBin is the compiled bytecode used for deploying new contracts.
const KeeperProviderMapBin = `0x6080604052336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506117f0806100536000396000f3fe608060405260043610610083576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063074a9190146100885780630ca05f9f1461011157806355f21eb71461017a578063893d20e814610220578063c484fef314610277578063c9a5444c146102e3578063f48601061461034c575b600080fd5b34801561009457600080fd5b506100f7600480360360408110156100ab57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610449565b604051808215151515815260200191505060405180910390f35b34801561011d57600080fd5b506101606004803603602081101561013457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506104c6565b604051808215151515815260200191505060405180910390f35b34801561018657600080fd5b506101c96004803603602081101561019d57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610697565b6040518080602001828103825283818151815260200191508051906020019060200280838360005b8381101561020c5780820151818401526020810190506101f1565b505050509050019250505060405180910390f35b34801561022c57600080fd5b506102356109e9565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34801561028357600080fd5b5061028c610a12565b6040518080602001828103825283818151815260200191508051906020019060200280838360005b838110156102cf5780820151818401526020810190506102b4565b505050509050019250505060405180910390f35b3480156102ef57600080fd5b506103326004803603602081101561030657600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610c50565b604051808215151515815260200191505060405180910390f35b34801561035857600080fd5b5061042f6004803603604081101561036f57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001906401000000008111156103ac57600080fd5b8201836020820111156103be57600080fd5b803590602001918460208302840111640100000000831117156103e057600080fd5b919080806020026020016040519081016040528093929190818152602001838360200280828437600081840152601f19601f820116905080830192505050505050509192919290505050610d3f565b604051808215151515815260200191505060405180910390f35b600080600080610459868661110d565b92509250925082156104b957600060018381548110151561047657fe5b90600052602060002090600202016001018281548110151561049457fe5b9060005260206000200160000160146101000a81548160ff0219169083151502179055505b6001935050505092915050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614156106245760008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050826000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e908184604051808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a16001915050610691565b7f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa60405180806020018281038252600e8152602001807fe4bda0e4b88de698af6f776e657200000000000000000000000000000000000081525060200191505060405180910390a1610692565b5b919050565b60606000806106a584611238565b915091508115156106ea5760006040519080825280602002602001820160405280156106e05781602001602082028038833980820191505090505b50925050506109e4565b600080905060008090505b60018381548110151561070457fe5b9060005260206000209060020201600101805490508110156107f5576001151560018481548110151561073357fe5b90600052602060002090600202016001018281548110151561075157fe5b9060005260206000200160000160149054906101000a900460ff1615151480156107da57506107d960018481548110151561078857fe5b9060005260206000209060020201600101828154811015156107a657fe5b9060005260206000200160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff166112ed565b5b156107e85781806001019250505b80806001019150506106f5565b506060816040519080825280602002602001820160405280156108275781602001602082028038833980820191505090505b509050600080905060008090505b60018581548110151561084457fe5b9060005260206000209060020201600101805490508110156109da576001151560018681548110151561087357fe5b90600052602060002090600202016001018281548110151561089157fe5b9060005260206000200160000160149054906101000a900460ff16151514801561091a57506109196001868154811015156108c857fe5b9060005260206000209060020201600101828154811015156108e657fe5b9060005260206000200160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff166112ed565b5b156109cd5760018581548110151561092e57fe5b90600052602060002090600202016001018181548110151561094c57fe5b9060005260206000200160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16838381518110151561098857fe5b9060200190602002019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505081806001019250505b8080600101915050610835565b5081955050505050505b919050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b6060600080905060008090505b600180549050811015610acd5760011515600182815481101515610a3f57fe5b906000526020600020906002020160000160149054906101000a900460ff161515148015610ab25750610ab1600182815481101515610a7a57fe5b906000526020600020906002020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff166114cd565b5b15610ac05781806001019250505b8080600101915050610a1f565b50606081604051908082528060200260200182016040528015610aff5781602001602082028038833980820191505090505b509050600080905060008090505b600180549050811015610c465760011515600182815481101515610b2d57fe5b906000526020600020906002020160000160149054906101000a900460ff161515148015610ba05750610b9f600182815481101515610b6857fe5b906000526020600020906002020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff166114cd565b5b15610c3957600181815481101515610bb457fe5b906000526020600020906002020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff168383815181101515610bf457fe5b9060200190602002019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505081806001019250505b8080600101915050610b0d565b5081935050505090565b6000806000610c5e84611238565b915091508115610d34576000600182815481101515610c7957fe5b906000526020600020906002020160000160146101000a81548160ff02191690831515021790555060008090505b600182815481101515610cb657fe5b906000526020600020906002020160010180549050811015610d32576000600183815481101515610ce357fe5b906000526020600020906002020160010182815481101515610d0157fe5b9060005260206000200160000160146101000a81548160ff0219169083151502179055508080600101915050610ca7565b505b600192505050919050565b6000806000610d4d85611238565b91509150811515610f2857600060018054905090506001808054905001600181610d7791906116ad565b5085600182815481101515610d8857fe5b906000526020600020906002020160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060018082815481101515610de757fe5b906000526020600020906002020160000160146101000a81548160ff02191690831515021790555060008090505b8551811015610f2157600182815481101515610e2d57fe5b906000526020600020906002020160010160408051908101604052808884815181101515610e5757fe5b9060200190602002015173ffffffffffffffffffffffffffffffffffffffff1681526020016001151581525090806001815401808255809150509060018203906000526020600020016000909192909190915060008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060208201518160000160146101000a81548160ff0219169083151502179055505050508080600101915050610e15565b5050611101565b60018082815481101515610f3857fe5b906000526020600020906002020160000160146101000a81548160ff02191690831515021790555060008090505b84518110156110ff576000806000610f95898986815181101515610f8657fe5b9060200190602002015161110d565b9250925092508215610ff85760018083815481101515610fb157fe5b906000526020600020906002020160010182815481101515610fcf57fe5b9060005260206000200160000160146101000a81548160ff0219169083151502179055506110ef565b60018281548110151561100757fe5b906000526020600020906002020160010160408051908101604052808a8781518110151561103157fe5b9060200190602002015173ffffffffffffffffffffffffffffffffffffffff1681526020016001151581525090806001815401808255809150509060018203906000526020600020016000909192909190915060008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060208201518160000160146101000a81548160ff0219169083151502179055505050505b5050508080600101915050610f66565b505b60019250505092915050565b600080600080600061111e87611238565b915091508115156111415760008060008191508090509450945094505050611231565b60008090505b60018281548110151561115657fe5b90600052602060002090600202016001018054905081101561121f578673ffffffffffffffffffffffffffffffffffffffff1660018381548110151561119857fe5b9060005260206000209060020201600101828154811015156111b657fe5b9060005260206000200160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614156112125760018282955095509550505050611231565b8080600101915050611147565b50600081600080905094509450945050505b9250925092565b60008060008090505b6001805490508110156112dc578373ffffffffffffffffffffffffffffffffffffffff1660018281548110151561127457fe5b906000526020600020906002020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614156112cf5760018192509250506112e8565b8080600101915050611241565b50600080809050915091505b915091565b600080739e4af0964ef92095ca3d2ae0c05b472837d8bd37905060008173ffffffffffffffffffffffffffffffffffffffff1663e4ae7d776040518163ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018080602001828103825260088152602001807f70726f766964657200000000000000000000000000000000000000000000000081525060200191505060206040518083038186803b1580156113a757600080fd5b505afa1580156113bb573d6000803e3d6000fd5b505050506040513d60208110156113d157600080fd5b81019080805190602001909291905050509050600081905060008173ffffffffffffffffffffffffffffffffffffffff1663c2bc2efc876040518263ffffffff167c0100000000000000000000000000000000000000000000000000000000028152600401808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060206040518083038186803b15801561148457600080fd5b505afa158015611498573d6000803e3d6000fd5b505050506040513d60208110156114ae57600080fd5b8101908080519060200190929190505050905080945050505050919050565b600080739e4af0964ef92095ca3d2ae0c05b472837d8bd37905060008173ffffffffffffffffffffffffffffffffffffffff1663e4ae7d776040518163ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018080602001828103825260068152602001807f6b6565706572000000000000000000000000000000000000000000000000000081525060200191505060206040518083038186803b15801561158757600080fd5b505afa15801561159b573d6000803e3d6000fd5b505050506040513d60208110156115b157600080fd5b81019080805190602001909291905050509050600081905060008173ffffffffffffffffffffffffffffffffffffffff1663c2bc2efc876040518263ffffffff167c0100000000000000000000000000000000000000000000000000000000028152600401808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060206040518083038186803b15801561166457600080fd5b505afa158015611678573d6000803e3d6000fd5b505050506040513d602081101561168e57600080fd5b8101908080519060200190929190505050905080945050505050919050565b8154818355818111156116da576002028160020283600052602060002091820191016116d991906116df565b5b505050565b61174691905b8082111561174257600080820160006101000a81549073ffffffffffffffffffffffffffffffffffffffff02191690556000820160146101000a81549060ff02191690556001820160006117399190611749565b506002016116e5565b5090565b90565b5080546000825590600052602060002090810190611767919061176a565b50565b6117c191905b808211156117bd57600080820160006101000a81549073ffffffffffffffffffffffffffffffffffffffff02191690556000820160146101000a81549060ff021916905550600101611770565b5090565b9056fea165627a7a723058203b829b89975c9740b4ac946478217a969b9bcc0b264ca7d9e1932578a9674be30029`

// DeployKeeperProviderMap deploys a new Ethereum contract, binding an instance of KeeperProviderMap to it.
func DeployKeeperProviderMap(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *KeeperProviderMap, error) {
	parsed, err := abi.JSON(strings.NewReader(KeeperProviderMapABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(KeeperProviderMapBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &KeeperProviderMap{KeeperProviderMapCaller: KeeperProviderMapCaller{contract: contract}, KeeperProviderMapTransactor: KeeperProviderMapTransactor{contract: contract}, KeeperProviderMapFilterer: KeeperProviderMapFilterer{contract: contract}}, nil
}

// KeeperProviderMap is an auto generated Go binding around an Ethereum contract.
type KeeperProviderMap struct {
	KeeperProviderMapCaller     // Read-only binding to the contract
	KeeperProviderMapTransactor // Write-only binding to the contract
	KeeperProviderMapFilterer   // Log filterer for contract events
}

// KeeperProviderMapCaller is an auto generated read-only Go binding around an Ethereum contract.
type KeeperProviderMapCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KeeperProviderMapTransactor is an auto generated write-only Go binding around an Ethereum contract.
type KeeperProviderMapTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KeeperProviderMapFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type KeeperProviderMapFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KeeperProviderMapSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type KeeperProviderMapSession struct {
	Contract     *KeeperProviderMap // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// KeeperProviderMapCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type KeeperProviderMapCallerSession struct {
	Contract *KeeperProviderMapCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// KeeperProviderMapTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type KeeperProviderMapTransactorSession struct {
	Contract     *KeeperProviderMapTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// KeeperProviderMapRaw is an auto generated low-level Go binding around an Ethereum contract.
type KeeperProviderMapRaw struct {
	Contract *KeeperProviderMap // Generic contract binding to access the raw methods on
}

// KeeperProviderMapCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type KeeperProviderMapCallerRaw struct {
	Contract *KeeperProviderMapCaller // Generic read-only contract binding to access the raw methods on
}

// KeeperProviderMapTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type KeeperProviderMapTransactorRaw struct {
	Contract *KeeperProviderMapTransactor // Generic write-only contract binding to access the raw methods on
}

// NewKeeperProviderMap creates a new instance of KeeperProviderMap, bound to a specific deployed contract.
func NewKeeperProviderMap(address common.Address, backend bind.ContractBackend) (*KeeperProviderMap, error) {
	contract, err := bindKeeperProviderMap(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &KeeperProviderMap{KeeperProviderMapCaller: KeeperProviderMapCaller{contract: contract}, KeeperProviderMapTransactor: KeeperProviderMapTransactor{contract: contract}, KeeperProviderMapFilterer: KeeperProviderMapFilterer{contract: contract}}, nil
}

// NewKeeperProviderMapCaller creates a new read-only instance of KeeperProviderMap, bound to a specific deployed contract.
func NewKeeperProviderMapCaller(address common.Address, caller bind.ContractCaller) (*KeeperProviderMapCaller, error) {
	contract, err := bindKeeperProviderMap(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &KeeperProviderMapCaller{contract: contract}, nil
}

// NewKeeperProviderMapTransactor creates a new write-only instance of KeeperProviderMap, bound to a specific deployed contract.
func NewKeeperProviderMapTransactor(address common.Address, transactor bind.ContractTransactor) (*KeeperProviderMapTransactor, error) {
	contract, err := bindKeeperProviderMap(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &KeeperProviderMapTransactor{contract: contract}, nil
}

// NewKeeperProviderMapFilterer creates a new log filterer instance of KeeperProviderMap, bound to a specific deployed contract.
func NewKeeperProviderMapFilterer(address common.Address, filterer bind.ContractFilterer) (*KeeperProviderMapFilterer, error) {
	contract, err := bindKeeperProviderMap(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &KeeperProviderMapFilterer{contract: contract}, nil
}

// bindKeeperProviderMap binds a generic wrapper to an already deployed contract.
func bindKeeperProviderMap(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(KeeperProviderMapABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_KeeperProviderMap *KeeperProviderMapRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _KeeperProviderMap.Contract.KeeperProviderMapCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_KeeperProviderMap *KeeperProviderMapRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _KeeperProviderMap.Contract.KeeperProviderMapTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_KeeperProviderMap *KeeperProviderMapRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _KeeperProviderMap.Contract.KeeperProviderMapTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_KeeperProviderMap *KeeperProviderMapCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _KeeperProviderMap.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_KeeperProviderMap *KeeperProviderMapTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _KeeperProviderMap.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_KeeperProviderMap *KeeperProviderMapTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _KeeperProviderMap.Contract.contract.Transact(opts, method, params...)
}

// GetAllKeeper is a free data retrieval call binding the contract method 0xc484fef3.
//
// Solidity: function getAllKeeper() constant returns(address[])
func (_KeeperProviderMap *KeeperProviderMapCaller) GetAllKeeper(opts *bind.CallOpts) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _KeeperProviderMap.contract.Call(opts, out, "getAllKeeper")
	return *ret0, err
}

// GetAllKeeper is a free data retrieval call binding the contract method 0xc484fef3.
//
// Solidity: function getAllKeeper() constant returns(address[])
func (_KeeperProviderMap *KeeperProviderMapSession) GetAllKeeper() ([]common.Address, error) {
	return _KeeperProviderMap.Contract.GetAllKeeper(&_KeeperProviderMap.CallOpts)
}

// GetAllKeeper is a free data retrieval call binding the contract method 0xc484fef3.
//
// Solidity: function getAllKeeper() constant returns(address[])
func (_KeeperProviderMap *KeeperProviderMapCallerSession) GetAllKeeper() ([]common.Address, error) {
	return _KeeperProviderMap.Contract.GetAllKeeper(&_KeeperProviderMap.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_KeeperProviderMap *KeeperProviderMapCaller) GetOwner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _KeeperProviderMap.contract.Call(opts, out, "getOwner")
	return *ret0, err
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_KeeperProviderMap *KeeperProviderMapSession) GetOwner() (common.Address, error) {
	return _KeeperProviderMap.Contract.GetOwner(&_KeeperProviderMap.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_KeeperProviderMap *KeeperProviderMapCallerSession) GetOwner() (common.Address, error) {
	return _KeeperProviderMap.Contract.GetOwner(&_KeeperProviderMap.CallOpts)
}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address keeper) constant returns(address[])
func (_KeeperProviderMap *KeeperProviderMapCaller) GetProvider(opts *bind.CallOpts, keeper common.Address) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _KeeperProviderMap.contract.Call(opts, out, "getProvider", keeper)
	return *ret0, err
}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address keeper) constant returns(address[])
func (_KeeperProviderMap *KeeperProviderMapSession) GetProvider(keeper common.Address) ([]common.Address, error) {
	return _KeeperProviderMap.Contract.GetProvider(&_KeeperProviderMap.CallOpts, keeper)
}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address keeper) constant returns(address[])
func (_KeeperProviderMap *KeeperProviderMapCallerSession) GetProvider(keeper common.Address) ([]common.Address, error) {
	return _KeeperProviderMap.Contract.GetProvider(&_KeeperProviderMap.CallOpts, keeper)
}

// Add is a paid mutator transaction binding the contract method 0xf4860106.
//
// Solidity: function add(address keeper, address[] provider) returns(bool)
func (_KeeperProviderMap *KeeperProviderMapTransactor) Add(opts *bind.TransactOpts, keeper common.Address, provider []common.Address) (*types.Transaction, error) {
	return _KeeperProviderMap.contract.Transact(opts, "add", keeper, provider)
}

// Add is a paid mutator transaction binding the contract method 0xf4860106.
//
// Solidity: function add(address keeper, address[] provider) returns(bool)
func (_KeeperProviderMap *KeeperProviderMapSession) Add(keeper common.Address, provider []common.Address) (*types.Transaction, error) {
	return _KeeperProviderMap.Contract.Add(&_KeeperProviderMap.TransactOpts, keeper, provider)
}

// Add is a paid mutator transaction binding the contract method 0xf4860106.
//
// Solidity: function add(address keeper, address[] provider) returns(bool)
func (_KeeperProviderMap *KeeperProviderMapTransactorSession) Add(keeper common.Address, provider []common.Address) (*types.Transaction, error) {
	return _KeeperProviderMap.Contract.Add(&_KeeperProviderMap.TransactOpts, keeper, provider)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_KeeperProviderMap *KeeperProviderMapTransactor) AlterOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _KeeperProviderMap.contract.Transact(opts, "alterOwner", newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_KeeperProviderMap *KeeperProviderMapSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _KeeperProviderMap.Contract.AlterOwner(&_KeeperProviderMap.TransactOpts, newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_KeeperProviderMap *KeeperProviderMapTransactorSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _KeeperProviderMap.Contract.AlterOwner(&_KeeperProviderMap.TransactOpts, newOwner)
}

// DelKeeper is a paid mutator transaction binding the contract method 0xc9a5444c.
//
// Solidity: function delKeeper(address keeper) returns(bool)
func (_KeeperProviderMap *KeeperProviderMapTransactor) DelKeeper(opts *bind.TransactOpts, keeper common.Address) (*types.Transaction, error) {
	return _KeeperProviderMap.contract.Transact(opts, "delKeeper", keeper)
}

// DelKeeper is a paid mutator transaction binding the contract method 0xc9a5444c.
//
// Solidity: function delKeeper(address keeper) returns(bool)
func (_KeeperProviderMap *KeeperProviderMapSession) DelKeeper(keeper common.Address) (*types.Transaction, error) {
	return _KeeperProviderMap.Contract.DelKeeper(&_KeeperProviderMap.TransactOpts, keeper)
}

// DelKeeper is a paid mutator transaction binding the contract method 0xc9a5444c.
//
// Solidity: function delKeeper(address keeper) returns(bool)
func (_KeeperProviderMap *KeeperProviderMapTransactorSession) DelKeeper(keeper common.Address) (*types.Transaction, error) {
	return _KeeperProviderMap.Contract.DelKeeper(&_KeeperProviderMap.TransactOpts, keeper)
}

// DelProvider is a paid mutator transaction binding the contract method 0x074a9190.
//
// Solidity: function delProvider(address keeper, address provider) returns(bool)
func (_KeeperProviderMap *KeeperProviderMapTransactor) DelProvider(opts *bind.TransactOpts, keeper common.Address, provider common.Address) (*types.Transaction, error) {
	return _KeeperProviderMap.contract.Transact(opts, "delProvider", keeper, provider)
}

// DelProvider is a paid mutator transaction binding the contract method 0x074a9190.
//
// Solidity: function delProvider(address keeper, address provider) returns(bool)
func (_KeeperProviderMap *KeeperProviderMapSession) DelProvider(keeper common.Address, provider common.Address) (*types.Transaction, error) {
	return _KeeperProviderMap.Contract.DelProvider(&_KeeperProviderMap.TransactOpts, keeper, provider)
}

// DelProvider is a paid mutator transaction binding the contract method 0x074a9190.
//
// Solidity: function delProvider(address keeper, address provider) returns(bool)
func (_KeeperProviderMap *KeeperProviderMapTransactorSession) DelProvider(keeper common.Address, provider common.Address) (*types.Transaction, error) {
	return _KeeperProviderMap.Contract.DelProvider(&_KeeperProviderMap.TransactOpts, keeper, provider)
}

// KeeperProviderMapAlterOwnerIterator is returned from FilterAlterOwner and is used to iterate over the raw logs and unpacked data for AlterOwner events raised by the KeeperProviderMap contract.
type KeeperProviderMapAlterOwnerIterator struct {
	Event *KeeperProviderMapAlterOwner // Event containing the contract specifics and raw log

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
func (it *KeeperProviderMapAlterOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperProviderMapAlterOwner)
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
		it.Event = new(KeeperProviderMapAlterOwner)
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
func (it *KeeperProviderMapAlterOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperProviderMapAlterOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperProviderMapAlterOwner represents a AlterOwner event raised by the KeeperProviderMap contract.
type KeeperProviderMapAlterOwner struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAlterOwner is a free log retrieval operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_KeeperProviderMap *KeeperProviderMapFilterer) FilterAlterOwner(opts *bind.FilterOpts) (*KeeperProviderMapAlterOwnerIterator, error) {

	logs, sub, err := _KeeperProviderMap.contract.FilterLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return &KeeperProviderMapAlterOwnerIterator{contract: _KeeperProviderMap.contract, event: "AlterOwner", logs: logs, sub: sub}, nil
}

// WatchAlterOwner is a free log subscription operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_KeeperProviderMap *KeeperProviderMapFilterer) WatchAlterOwner(opts *bind.WatchOpts, sink chan<- *KeeperProviderMapAlterOwner) (event.Subscription, error) {

	logs, sub, err := _KeeperProviderMap.contract.WatchLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperProviderMapAlterOwner)
				if err := _KeeperProviderMap.contract.UnpackLog(event, "AlterOwner", log); err != nil {
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

// KeeperProviderMapErrorIterator is returned from FilterError and is used to iterate over the raw logs and unpacked data for Error events raised by the KeeperProviderMap contract.
type KeeperProviderMapErrorIterator struct {
	Event *KeeperProviderMapError // Event containing the contract specifics and raw log

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
func (it *KeeperProviderMapErrorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperProviderMapError)
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
		it.Event = new(KeeperProviderMapError)
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
func (it *KeeperProviderMapErrorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperProviderMapErrorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperProviderMapError represents a Error event raised by the KeeperProviderMap contract.
type KeeperProviderMapError struct {
	Data string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterError is a free log retrieval operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_KeeperProviderMap *KeeperProviderMapFilterer) FilterError(opts *bind.FilterOpts) (*KeeperProviderMapErrorIterator, error) {

	logs, sub, err := _KeeperProviderMap.contract.FilterLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return &KeeperProviderMapErrorIterator{contract: _KeeperProviderMap.contract, event: "Error", logs: logs, sub: sub}, nil
}

// WatchError is a free log subscription operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_KeeperProviderMap *KeeperProviderMapFilterer) WatchError(opts *bind.WatchOpts, sink chan<- *KeeperProviderMapError) (event.Subscription, error) {

	logs, sub, err := _KeeperProviderMap.contract.WatchLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperProviderMapError)
				if err := _KeeperProviderMap.contract.UnpackLog(event, "Error", log); err != nil {
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
