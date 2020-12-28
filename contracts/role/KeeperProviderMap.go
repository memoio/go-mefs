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

// KeeperProviderMapABI is the input ABI used to generate the binding from.
const KeeperProviderMapABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"keeper\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"provider\",\"type\":\"address[]\"}],\"name\":\"add\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"keeper\",\"type\":\"address\"}],\"name\":\"delKeeper\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"keeper\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"}],\"name\":\"delProvider\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAllKeeper\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"keeper\",\"type\":\"address\"}],\"name\":\"getProvider\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// KeeperProviderMapBin is the compiled bytecode used for deploying new contracts.
var KeeperProviderMapBin = "0x6080604052738391984e2f1cc8f6b916f566c1d0a6bb8a15c73a600160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555034801561006557600080fd5b5061164f806100756000396000f3fe608060405234801561001057600080fd5b50600436106100575760003560e01c8063074a91901461005c57806355f21eb7146100d6578063c484fef31461016f578063c9a5444c146101ce578063f486010614610228575b600080fd5b6100be6004803603604081101561007257600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610316565b60405180821515815260200191505060405180910390f35b610118600480360360208110156100ec57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506104aa565b6040518080602001828103825283818151815260200191508051906020019060200280838360005b8381101561015b578082015181840152602081019050610140565b505050509050019250505060405180910390f35b6101776107fe565b6040518080602001828103825283818151815260200191508051906020019060200280838360005b838110156101ba57808201518184015260208101905061019f565b505050509050019250505060405180910390f35b610210600480360360208110156101e457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610a37565b60405180821515815260200191505060405180910390f35b6102fe6004803603604081101561023e57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019064010000000081111561027b57600080fd5b82018360208201111561028d57600080fd5b803590602001918460208302840111640100000000831117156102af57600080fd5b919080806020026020016040519081016040528093929190818152602001838360200280828437600081840152601f19601f820116905080830192505050505050509192919290505050610c34565b60405180821515815260200191505060405180910390f35b600080600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663d9d6389e6040518163ffffffff1660e01b815260040160206040518083038186803b15801561038157600080fd5b505afa158015610395573d6000803e3d6000fd5b505050506040513d60208110156103ab57600080fd5b810190808051906020019092919050505090508015610432576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260098152602001807f69732062616e6e6564000000000000000000000000000000000000000000000081525060200191505060405180910390fd5b600080600061044187876110db565b925092509250821561049c57600080838154811061045b57fe5b9060005260206000209060020201600101828154811061047757fe5b9060005260206000200160000160146101000a81548160ff0219169083151502179055505b600194505050505092915050565b60606000806104b8846111f2565b915091508161051257600067ffffffffffffffff811180156104d957600080fd5b506040519080825280602002602001820160405280156105085781602001602082028036833780820191505090505b50925050506107f9565b6000805b6000838154811061052357fe5b90600052602060002090600202016001018054905081101561060c57600115156000848154811061055057fe5b9060005260206000209060020201600101828154811061056c57fe5b9060005260206000200160000160149054906101000a900460ff1615151480156105f157506105f0600084815481106105a157fe5b906000526020600020906002020160010182815481106105bd57fe5b9060005260206000200160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1661129f565b5b156105ff5781806001019250505b8080600101915050610516565b5060608167ffffffffffffffff8111801561062657600080fd5b506040519080825280602002602001820160405280156106555781602001602082028036833780820191505090505b5090506000805b6000858154811061066957fe5b9060005260206000209060020201600101805490508110156107ef57600115156000868154811061069657fe5b906000526020600020906002020160010182815481106106b257fe5b9060005260206000200160000160149054906101000a900460ff1615151480156107375750610736600086815481106106e757fe5b9060005260206000209060020201600101828154811061070357fe5b9060005260206000200160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1661129f565b5b156107e2576000858154811061074957fe5b9060005260206000209060020201600101818154811061076557fe5b9060005260206000200160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1683838151811061079f57fe5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505081806001019250505b808060010191505061065c565b5081955050505050505b919050565b60606000805b6000805490508110156108ae57600115156000828154811061082257fe5b906000526020600020906002020160000160149054906101000a900460ff16151514801561089357506108926000828154811061085b57fe5b906000526020600020906002020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1661145c565b5b156108a15781806001019250505b8080600101915050610804565b5060608167ffffffffffffffff811180156108c857600080fd5b506040519080825280602002602001820160405280156108f75781602001602082028036833780820191505090505b5090506000805b600080549050811015610a2d57600115156000828154811061091c57fe5b906000526020600020906002020160000160149054906101000a900460ff16151514801561098d575061098c6000828154811061095557fe5b906000526020600020906002020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1661145c565b5b15610a20576000818154811061099f57fe5b906000526020600020906002020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff168383815181106109dd57fe5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505081806001019250505b80806001019150506108fe565b5081935050505090565b600080600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663d9d6389e6040518163ffffffff1660e01b815260040160206040518083038186803b158015610aa257600080fd5b505afa158015610ab6573d6000803e3d6000fd5b505050506040513d6020811015610acc57600080fd5b810190808051906020019092919050505090508015610b53576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260098152602001807f69732062616e6e6564000000000000000000000000000000000000000000000081525060200191505060405180910390fd5b600080610b5f856111f2565b915091508115610c28576000808281548110610b7757fe5b906000526020600020906002020160000160146101000a81548160ff02191690831515021790555060005b60008281548110610baf57fe5b906000526020600020906002020160010180549050811015610c26576000808381548110610bd957fe5b90600052602060002090600202016001018281548110610bf557fe5b9060005260206000200160000160146101000a81548160ff0219169083151502179055508080600101915050610ba2565b505b60019350505050919050565b600080600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663d9d6389e6040518163ffffffff1660e01b815260040160206040518083038186803b158015610c9f57600080fd5b505afa158015610cb3573d6000803e3d6000fd5b505050506040513d6020811015610cc957600080fd5b810190808051906020019092919050505090508015610d50576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252600d8152602001807f6164642069732062616e6e65640000000000000000000000000000000000000081525060200191505060405180910390fd5b600080610d5c866111f2565b9150915081610f0d5760008054905090508560008281548110610d7b57fe5b906000526020600020906002020160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550600160008281548110610dd957fe5b906000526020600020906002020160000160146101000a81548160ff02191690831515021790555060005b8551811015610f075760008281548110610e1a57fe5b90600052602060002090600202016001016040518060400160405280888481518110610e4257fe5b602002602001015173ffffffffffffffffffffffffffffffffffffffff168152602001600115158152509080600181540180825580915050600190039060005260206000200160009091909190915060008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060208201518160000160146101000a81548160ff02191690831515021790555050508080600101915050610e04565b506110ce565b600160008281548110610f1c57fe5b906000526020600020906002020160000160146101000a81548160ff02191690831515021790555060005b85518110156110cc57600080610f7089898581518110610f6357fe5b60200260200101516110db565b92505091508115610fcf57600160008581548110610f8a57fe5b90600052602060002090600202016001018281548110610fa657fe5b9060005260206000200160000160146101000a81548160ff0219169083151502179055506110bd565b60008481548110610fdc57fe5b906000526020600020906002020160010160405180604001604052808a868151811061100457fe5b602002602001015173ffffffffffffffffffffffffffffffffffffffff168152602001600115158152509080600181540180825580915050600190039060005260206000200160009091909190915060008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060208201518160000160146101000a81548160ff02191690831515021790555050505b50508080600101915050610f47565b505b6001935050505092915050565b60008060008060006110ec876111f2565b915091508161110757600080600094509450945050506111eb565b60005b6000828154811061111757fe5b9060005260206000209060020201600101805490508110156111dc578673ffffffffffffffffffffffffffffffffffffffff166000838154811061115757fe5b9060005260206000209060020201600101828154811061117357fe5b9060005260206000200160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614156111cf57600182829550955095505050506111eb565b808060010191505061110a565b50600080600094509450945050505b9250925092565b60008060005b600080549050811015611291578373ffffffffffffffffffffffffffffffffffffffff166000828154811061122957fe5b906000526020600020906002020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16141561128457600181925092505061129a565b80806001019150506111f8565b50600080915091505b915091565b600080739e4af0964ef92095ca3d2ae0c05b472837d8bd37905060008173ffffffffffffffffffffffffffffffffffffffff1663693ec85e6040518163ffffffff1660e01b81526004018080602001828103825260088152602001807f70726f7669646572000000000000000000000000000000000000000000000000815250602001915050604080518083038186803b15801561133c57600080fd5b505afa158015611350573d6000803e3d6000fd5b505050506040513d604081101561136657600080fd5b810190808051906020019092919080519060200190929190505050915050600081905060008173ffffffffffffffffffffffffffffffffffffffff16630aae7a6b876040518263ffffffff1660e01b8152600401808273ffffffffffffffffffffffffffffffffffffffff16815260200191505060806040518083038186803b1580156113f257600080fd5b505afa158015611406573d6000803e3d6000fd5b505050506040513d608081101561141c57600080fd5b8101908080519060200190929190805190602001909291908051906020019092919080519060200190929190505050505050905080945050505050919050565b600080739e4af0964ef92095ca3d2ae0c05b472837d8bd37905060008173ffffffffffffffffffffffffffffffffffffffff1663693ec85e6040518163ffffffff1660e01b81526004018080602001828103825260068152602001807f6b65657065720000000000000000000000000000000000000000000000000000815250602001915050604080518083038186803b1580156114f957600080fd5b505afa15801561150d573d6000803e3d6000fd5b505050506040513d604081101561152357600080fd5b810190808051906020019092919080519060200190929190505050915050600081905060008173ffffffffffffffffffffffffffffffffffffffff16630aae7a6b876040518263ffffffff1660e01b8152600401808273ffffffffffffffffffffffffffffffffffffffff16815260200191505060806040518083038186803b1580156115af57600080fd5b505afa1580156115c3573d6000803e3d6000fd5b505050506040513d60808110156115d957600080fd5b810190808051906020019092919080519060200190929190805190602001909291908051906020019092919050505050505090508094505050505091905056fea26469706673582212200fbf22e165bb499cfaab15a7ecbcc24904f2587cebe1df26fe56ab743893678464736f6c63430007030033"

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
// Solidity: function getAllKeeper() view returns(address[])
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
// Solidity: function getAllKeeper() view returns(address[])
func (_KeeperProviderMap *KeeperProviderMapSession) GetAllKeeper() ([]common.Address, error) {
	return _KeeperProviderMap.Contract.GetAllKeeper(&_KeeperProviderMap.CallOpts)
}

// GetAllKeeper is a free data retrieval call binding the contract method 0xc484fef3.
//
// Solidity: function getAllKeeper() view returns(address[])
func (_KeeperProviderMap *KeeperProviderMapCallerSession) GetAllKeeper() ([]common.Address, error) {
	return _KeeperProviderMap.Contract.GetAllKeeper(&_KeeperProviderMap.CallOpts)
}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address keeper) view returns(address[])
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
// Solidity: function getProvider(address keeper) view returns(address[])
func (_KeeperProviderMap *KeeperProviderMapSession) GetProvider(keeper common.Address) ([]common.Address, error) {
	return _KeeperProviderMap.Contract.GetProvider(&_KeeperProviderMap.CallOpts, keeper)
}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address keeper) view returns(address[])
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
