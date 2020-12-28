// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package indexer

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

// IndexerABI is the input ABI used to generate the binding from.
const IndexerABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"}],\"name\":\"Add\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"oldOwner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"oldResolver\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newResolver\",\"type\":\"address\"}],\"name\":\"AlterResolver\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"data\",\"type\":\"string\"}],\"name\":\"Error\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"}],\"name\":\"add\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"resolver\",\"type\":\"address\"}],\"name\":\"alterResolver\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"}],\"name\":\"get\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"param\",\"type\":\"bool\"}],\"name\":\"setBanned\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// IndexerBin is the compiled bytecode used for deploying new contracts.
var IndexerBin = "0x60806040526000600260006101000a81548160ff02191690831515021790555034801561002b57600080fd5b50336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506113aa8061007b6000396000f3fe608060405234801561001057600080fd5b506004361061007d5760003560e01c8063693ec85e1161005b578063693ec85e146101fd578063893d20e8146102ff578063cbdc3fe114610333578063f60b53e2146104245761007d565b80630ca05f9f146100825780632bffc7ed146100dc5780634b9c5d3b146101cd575b600080fd5b6100c46004803603602081101561009857600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610515565b60405180821515815260200191505060405180910390f35b6101b5600480360360408110156100f257600080fd5b810190808035906020019064010000000081111561010f57600080fd5b82018360208201111561012157600080fd5b8035906020019184600183028401116401000000008311171561014357600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506106b4565b60405180821515815260200191505060405180910390f35b6101fb600480360360208110156101e357600080fd5b81019080803515159060200190929190505050610aa3565b005b6102b66004803603602081101561021357600080fd5b810190808035906020019064010000000081111561023057600080fd5b82018360208201111561024257600080fd5b8035906020019184600183028401116401000000008311171561026457600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290505050610b81565b604051808373ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390f35b610307610ca7565b604051808273ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b61040c6004803603604081101561034957600080fd5b810190808035906020019064010000000081111561036657600080fd5b82018360208201111561037857600080fd5b8035906020019184600183028401116401000000008311171561039a57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610cd0565b60405180821515815260200191505060405180910390f35b6104fd6004803603604081101561043a57600080fd5b810190808035906020019064010000000081111561045757600080fd5b82018360208201111561046957600080fd5b8035906020019184600183028401116401000000008311171561048b57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050611022565b60405180821515815260200191505060405180910390f35b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146105d9576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050826000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e908184604051808373ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a16001915050919050565b6000600260009054906101000a900460ff1615610739576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260098152602001807f69732062616e6e6564000000000000000000000000000000000000000000000081525060200191505060405180910390fd5b600073ffffffffffffffffffffffffffffffffffffffff166001846040518082805190602001908083835b602083106107875780518252602082019150602081019050602083039250610764565b6001836020036101000a038019825116818451168082178552505050505050905001915050908152602001604051809103902060000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff161461086a577f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa6040518080602001828103825260188152602001807f6b657920686173207265736f6c76657220616c7265616479000000000000000081525060200191505060405180910390a160009050610a9d565b336001846040518082805190602001908083835b602083106108a1578051825260208201915060208101905060208303925061087e565b6001836020036101000a038019825116818451168082178552505050505050905001915050908152602001604051809103902060000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550816001846040518082805190602001908083835b6020831061094c5780518252602082019150602081019050602083039250610929565b6001836020036101000a038019825116818451168082178552505050505050905001915050908152602001604051809103902060010160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507fec689a3871c35587e4800f14216f987ee744b924aff21741edc2e167e2dd43e883338460405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff168152602001828103825285818151815260200191508051906020019080838360005b83811015610a5c578082015181840152602081019050610a41565b50505050905090810190601f168015610a895780820380516001836020036101000a031916815260200191505b5094505050505060405180910390a1600190505b92915050565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610b64576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b80600260006101000a81548160ff02191690831515021790555050565b6000806001836040518082805190602001908083835b60208310610bba5780518252602082019150602081019050602083039250610b97565b6001836020036101000a038019825116818451168082178552505050505050905001915050908152602001604051809103902060000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff166001846040518082805190602001908083835b60208310610c475780518252602082019150602081019050602083039250610c24565b6001836020036101000a038019825116818451168082178552505050505050905001915050908152602001604051809103902060010160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1691509150915091565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b60003373ffffffffffffffffffffffffffffffffffffffff166001846040518082805190602001908083835b60208310610d1f5780518252602082019150602081019050602083039250610cfc565b6001836020036101000a038019825116818451168082178552505050505050905001915050908152602001604051809103902060000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614610e02577f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa6040518080602001828103825260098152602001807f6e6f74206f776e6572000000000000000000000000000000000000000000000081525060200191505060405180910390a16000905061101c565b60006001846040518082805190602001908083835b60208310610e3a5780518252602082019150602081019050602083039250610e17565b6001836020036101000a038019825116818451168082178552505050505050905001915050908152602001604051809103902060010160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050826001856040518082805190602001908083835b60208310610eca5780518252602082019150602081019050602083039250610ea7565b6001836020036101000a038019825116818451168082178552505050505050905001915050908152602001604051809103902060010160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f0a7047ba8be4d874e67aebc953a70ff6db03a81782549290ac646e0738ddfc0484828560405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff168152602001828103825285818151815260200191508051906020019080838360005b83811015610fda578082015181840152602081019050610fbf565b50505050905090810190601f1680156110075780820380516001836020036101000a031916815260200191505b5094505050505060405180910390a160019150505b92915050565b60003373ffffffffffffffffffffffffffffffffffffffff166001846040518082805190602001908083835b60208310611071578051825260208201915060208101905060208303925061104e565b6001836020036101000a038019825116818451168082178552505050505050905001915050908152602001604051809103902060000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614611154577f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa6040518080602001828103825260098152602001807f6e6f74206f776e6572000000000000000000000000000000000000000000000081525060200191505060405180910390a16000905061136e565b60006001846040518082805190602001908083835b6020831061118c5780518252602082019150602081019050602083039250611169565b6001836020036101000a038019825116818451168082178552505050505050905001915050908152602001604051809103902060000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050826001856040518082805190602001908083835b6020831061121c57805182526020820191506020810190506020830392506111f9565b6001836020036101000a038019825116818451168082178552505050505050905001915050908152602001604051809103902060000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f46bd035a76a8302bb74520f9226b59925d8186784298f88ad636a4ea46b85b2184828560405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018373ffffffffffffffffffffffffffffffffffffffff168152602001828103825285818151815260200191508051906020019080838360005b8381101561132c578082015181840152602081019050611311565b50505050905090810190601f1680156113595780820380516001836020036101000a031916815260200191505b5094505050505060405180910390a160019150505b9291505056fea2646970667358221220195a263433a107ca5305bd876da88987cda43b8f87ba65ca066a88932d9d44f564736f6c63430007030033"

// DeployIndexer deploys a new Ethereum contract, binding an instance of Indexer to it.
func DeployIndexer(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Indexer, error) {
	parsed, err := abi.JSON(strings.NewReader(IndexerABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(IndexerBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Indexer{IndexerCaller: IndexerCaller{contract: contract}, IndexerTransactor: IndexerTransactor{contract: contract}, IndexerFilterer: IndexerFilterer{contract: contract}}, nil
}

// Indexer is an auto generated Go binding around an Ethereum contract.
type Indexer struct {
	IndexerCaller     // Read-only binding to the contract
	IndexerTransactor // Write-only binding to the contract
	IndexerFilterer   // Log filterer for contract events
}

// IndexerCaller is an auto generated read-only Go binding around an Ethereum contract.
type IndexerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IndexerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IndexerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IndexerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IndexerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IndexerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IndexerSession struct {
	Contract     *Indexer          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IndexerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IndexerCallerSession struct {
	Contract *IndexerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// IndexerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IndexerTransactorSession struct {
	Contract     *IndexerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// IndexerRaw is an auto generated low-level Go binding around an Ethereum contract.
type IndexerRaw struct {
	Contract *Indexer // Generic contract binding to access the raw methods on
}

// IndexerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IndexerCallerRaw struct {
	Contract *IndexerCaller // Generic read-only contract binding to access the raw methods on
}

// IndexerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IndexerTransactorRaw struct {
	Contract *IndexerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIndexer creates a new instance of Indexer, bound to a specific deployed contract.
func NewIndexer(address common.Address, backend bind.ContractBackend) (*Indexer, error) {
	contract, err := bindIndexer(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Indexer{IndexerCaller: IndexerCaller{contract: contract}, IndexerTransactor: IndexerTransactor{contract: contract}, IndexerFilterer: IndexerFilterer{contract: contract}}, nil
}

// NewIndexerCaller creates a new read-only instance of Indexer, bound to a specific deployed contract.
func NewIndexerCaller(address common.Address, caller bind.ContractCaller) (*IndexerCaller, error) {
	contract, err := bindIndexer(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IndexerCaller{contract: contract}, nil
}

// NewIndexerTransactor creates a new write-only instance of Indexer, bound to a specific deployed contract.
func NewIndexerTransactor(address common.Address, transactor bind.ContractTransactor) (*IndexerTransactor, error) {
	contract, err := bindIndexer(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IndexerTransactor{contract: contract}, nil
}

// NewIndexerFilterer creates a new log filterer instance of Indexer, bound to a specific deployed contract.
func NewIndexerFilterer(address common.Address, filterer bind.ContractFilterer) (*IndexerFilterer, error) {
	contract, err := bindIndexer(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IndexerFilterer{contract: contract}, nil
}

// bindIndexer binds a generic wrapper to an already deployed contract.
func bindIndexer(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IndexerABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Indexer *IndexerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Indexer.Contract.IndexerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Indexer *IndexerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Indexer.Contract.IndexerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Indexer *IndexerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Indexer.Contract.IndexerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Indexer *IndexerCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Indexer.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Indexer *IndexerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Indexer.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Indexer *IndexerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Indexer.Contract.contract.Transact(opts, method, params...)
}

// Get is a free data retrieval call binding the contract method 0x693ec85e.
//
// Solidity: function get(string key) view returns(address, address)
func (_Indexer *IndexerCaller) Get(opts *bind.CallOpts, key string) (common.Address, common.Address, error) {
	var (
		ret0 = new(common.Address)
		ret1 = new(common.Address)
	)
	out := &[]interface{}{
		ret0,
		ret1,
	}
	err := _Indexer.contract.Call(opts, out, "get", key)
	return *ret0, *ret1, err
}

// Get is a free data retrieval call binding the contract method 0x693ec85e.
//
// Solidity: function get(string key) view returns(address, address)
func (_Indexer *IndexerSession) Get(key string) (common.Address, common.Address, error) {
	return _Indexer.Contract.Get(&_Indexer.CallOpts, key)
}

// Get is a free data retrieval call binding the contract method 0x693ec85e.
//
// Solidity: function get(string key) view returns(address, address)
func (_Indexer *IndexerCallerSession) Get(key string) (common.Address, common.Address, error) {
	return _Indexer.Contract.Get(&_Indexer.CallOpts, key)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_Indexer *IndexerCaller) GetOwner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Indexer.contract.Call(opts, out, "getOwner")
	return *ret0, err
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_Indexer *IndexerSession) GetOwner() (common.Address, error) {
	return _Indexer.Contract.GetOwner(&_Indexer.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_Indexer *IndexerCallerSession) GetOwner() (common.Address, error) {
	return _Indexer.Contract.GetOwner(&_Indexer.CallOpts)
}

// Add is a paid mutator transaction binding the contract method 0x2bffc7ed.
//
// Solidity: function add(string key, address resolver) returns(bool)
func (_Indexer *IndexerTransactor) Add(opts *bind.TransactOpts, key string, resolver common.Address) (*types.Transaction, error) {
	return _Indexer.contract.Transact(opts, "add", key, resolver)
}

// Add is a paid mutator transaction binding the contract method 0x2bffc7ed.
//
// Solidity: function add(string key, address resolver) returns(bool)
func (_Indexer *IndexerSession) Add(key string, resolver common.Address) (*types.Transaction, error) {
	return _Indexer.Contract.Add(&_Indexer.TransactOpts, key, resolver)
}

// Add is a paid mutator transaction binding the contract method 0x2bffc7ed.
//
// Solidity: function add(string key, address resolver) returns(bool)
func (_Indexer *IndexerTransactorSession) Add(key string, resolver common.Address) (*types.Transaction, error) {
	return _Indexer.Contract.Add(&_Indexer.TransactOpts, key, resolver)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Indexer *IndexerTransactor) AlterOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Indexer.contract.Transact(opts, "alterOwner", newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Indexer *IndexerSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Indexer.Contract.AlterOwner(&_Indexer.TransactOpts, newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_Indexer *IndexerTransactorSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Indexer.Contract.AlterOwner(&_Indexer.TransactOpts, newOwner)
}

// AlterOwner0 is a paid mutator transaction binding the contract method 0xf60b53e2.
//
// Solidity: function alterOwner(string key, address owner) returns(bool)
func (_Indexer *IndexerTransactor) AlterOwner0(opts *bind.TransactOpts, key string, owner common.Address) (*types.Transaction, error) {
	return _Indexer.contract.Transact(opts, "alterOwner0", key, owner)
}

// AlterOwner0 is a paid mutator transaction binding the contract method 0xf60b53e2.
//
// Solidity: function alterOwner(string key, address owner) returns(bool)
func (_Indexer *IndexerSession) AlterOwner0(key string, owner common.Address) (*types.Transaction, error) {
	return _Indexer.Contract.AlterOwner0(&_Indexer.TransactOpts, key, owner)
}

// AlterOwner0 is a paid mutator transaction binding the contract method 0xf60b53e2.
//
// Solidity: function alterOwner(string key, address owner) returns(bool)
func (_Indexer *IndexerTransactorSession) AlterOwner0(key string, owner common.Address) (*types.Transaction, error) {
	return _Indexer.Contract.AlterOwner0(&_Indexer.TransactOpts, key, owner)
}

// AlterResolver is a paid mutator transaction binding the contract method 0xcbdc3fe1.
//
// Solidity: function alterResolver(string key, address resolver) returns(bool)
func (_Indexer *IndexerTransactor) AlterResolver(opts *bind.TransactOpts, key string, resolver common.Address) (*types.Transaction, error) {
	return _Indexer.contract.Transact(opts, "alterResolver", key, resolver)
}

// AlterResolver is a paid mutator transaction binding the contract method 0xcbdc3fe1.
//
// Solidity: function alterResolver(string key, address resolver) returns(bool)
func (_Indexer *IndexerSession) AlterResolver(key string, resolver common.Address) (*types.Transaction, error) {
	return _Indexer.Contract.AlterResolver(&_Indexer.TransactOpts, key, resolver)
}

// AlterResolver is a paid mutator transaction binding the contract method 0xcbdc3fe1.
//
// Solidity: function alterResolver(string key, address resolver) returns(bool)
func (_Indexer *IndexerTransactorSession) AlterResolver(key string, resolver common.Address) (*types.Transaction, error) {
	return _Indexer.Contract.AlterResolver(&_Indexer.TransactOpts, key, resolver)
}

// SetBanned is a paid mutator transaction binding the contract method 0x4b9c5d3b.
//
// Solidity: function setBanned(bool param) returns()
func (_Indexer *IndexerTransactor) SetBanned(opts *bind.TransactOpts, param bool) (*types.Transaction, error) {
	return _Indexer.contract.Transact(opts, "setBanned", param)
}

// SetBanned is a paid mutator transaction binding the contract method 0x4b9c5d3b.
//
// Solidity: function setBanned(bool param) returns()
func (_Indexer *IndexerSession) SetBanned(param bool) (*types.Transaction, error) {
	return _Indexer.Contract.SetBanned(&_Indexer.TransactOpts, param)
}

// SetBanned is a paid mutator transaction binding the contract method 0x4b9c5d3b.
//
// Solidity: function setBanned(bool param) returns()
func (_Indexer *IndexerTransactorSession) SetBanned(param bool) (*types.Transaction, error) {
	return _Indexer.Contract.SetBanned(&_Indexer.TransactOpts, param)
}

// IndexerAddIterator is returned from FilterAdd and is used to iterate over the raw logs and unpacked data for Add events raised by the Indexer contract.
type IndexerAddIterator struct {
	Event *IndexerAdd // Event containing the contract specifics and raw log

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
func (it *IndexerAddIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IndexerAdd)
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
		it.Event = new(IndexerAdd)
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
func (it *IndexerAddIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IndexerAddIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IndexerAdd represents a Add event raised by the Indexer contract.
type IndexerAdd struct {
	Key      string
	Owner    common.Address
	Resolver common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterAdd is a free log retrieval operation binding the contract event 0xec689a3871c35587e4800f14216f987ee744b924aff21741edc2e167e2dd43e8.
//
// Solidity: event Add(string key, address owner, address resolver)
func (_Indexer *IndexerFilterer) FilterAdd(opts *bind.FilterOpts) (*IndexerAddIterator, error) {

	logs, sub, err := _Indexer.contract.FilterLogs(opts, "Add")
	if err != nil {
		return nil, err
	}
	return &IndexerAddIterator{contract: _Indexer.contract, event: "Add", logs: logs, sub: sub}, nil
}

// WatchAdd is a free log subscription operation binding the contract event 0xec689a3871c35587e4800f14216f987ee744b924aff21741edc2e167e2dd43e8.
//
// Solidity: event Add(string key, address owner, address resolver)
func (_Indexer *IndexerFilterer) WatchAdd(opts *bind.WatchOpts, sink chan<- *IndexerAdd) (event.Subscription, error) {

	logs, sub, err := _Indexer.contract.WatchLogs(opts, "Add")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IndexerAdd)
				if err := _Indexer.contract.UnpackLog(event, "Add", log); err != nil {
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

// ParseAdd is a log parse operation binding the contract event 0xec689a3871c35587e4800f14216f987ee744b924aff21741edc2e167e2dd43e8.
//
// Solidity: event Add(string key, address owner, address resolver)
func (_Indexer *IndexerFilterer) ParseAdd(log types.Log) (*IndexerAdd, error) {
	event := new(IndexerAdd)
	if err := _Indexer.contract.UnpackLog(event, "Add", log); err != nil {
		return nil, err
	}
	return event, nil
}

// IndexerAlterOwnerIterator is returned from FilterAlterOwner and is used to iterate over the raw logs and unpacked data for AlterOwner events raised by the Indexer contract.
type IndexerAlterOwnerIterator struct {
	Event *IndexerAlterOwner // Event containing the contract specifics and raw log

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
func (it *IndexerAlterOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IndexerAlterOwner)
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
		it.Event = new(IndexerAlterOwner)
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
func (it *IndexerAlterOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IndexerAlterOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IndexerAlterOwner represents a AlterOwner event raised by the Indexer contract.
type IndexerAlterOwner struct {
	Key      string
	OldOwner common.Address
	NewOwner common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterAlterOwner is a free log retrieval operation binding the contract event 0x46bd035a76a8302bb74520f9226b59925d8186784298f88ad636a4ea46b85b21.
//
// Solidity: event AlterOwner(string key, address oldOwner, address newOwner)
func (_Indexer *IndexerFilterer) FilterAlterOwner(opts *bind.FilterOpts) (*IndexerAlterOwnerIterator, error) {

	logs, sub, err := _Indexer.contract.FilterLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return &IndexerAlterOwnerIterator{contract: _Indexer.contract, event: "AlterOwner", logs: logs, sub: sub}, nil
}

// WatchAlterOwner is a free log subscription operation binding the contract event 0x46bd035a76a8302bb74520f9226b59925d8186784298f88ad636a4ea46b85b21.
//
// Solidity: event AlterOwner(string key, address oldOwner, address newOwner)
func (_Indexer *IndexerFilterer) WatchAlterOwner(opts *bind.WatchOpts, sink chan<- *IndexerAlterOwner) (event.Subscription, error) {

	logs, sub, err := _Indexer.contract.WatchLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IndexerAlterOwner)
				if err := _Indexer.contract.UnpackLog(event, "AlterOwner", log); err != nil {
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

// ParseAlterOwner is a log parse operation binding the contract event 0x46bd035a76a8302bb74520f9226b59925d8186784298f88ad636a4ea46b85b21.
//
// Solidity: event AlterOwner(string key, address oldOwner, address newOwner)
func (_Indexer *IndexerFilterer) ParseAlterOwner(log types.Log) (*IndexerAlterOwner, error) {
	event := new(IndexerAlterOwner)
	if err := _Indexer.contract.UnpackLog(event, "AlterOwner", log); err != nil {
		return nil, err
	}
	return event, nil
}

// IndexerAlterOwner0Iterator is returned from FilterAlterOwner0 and is used to iterate over the raw logs and unpacked data for AlterOwner0 events raised by the Indexer contract.
type IndexerAlterOwner0Iterator struct {
	Event *IndexerAlterOwner0 // Event containing the contract specifics and raw log

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
func (it *IndexerAlterOwner0Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IndexerAlterOwner0)
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
		it.Event = new(IndexerAlterOwner0)
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
func (it *IndexerAlterOwner0Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IndexerAlterOwner0Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IndexerAlterOwner0 represents a AlterOwner0 event raised by the Indexer contract.
type IndexerAlterOwner0 struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAlterOwner0 is a free log retrieval operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Indexer *IndexerFilterer) FilterAlterOwner0(opts *bind.FilterOpts) (*IndexerAlterOwner0Iterator, error) {

	logs, sub, err := _Indexer.contract.FilterLogs(opts, "AlterOwner0")
	if err != nil {
		return nil, err
	}
	return &IndexerAlterOwner0Iterator{contract: _Indexer.contract, event: "AlterOwner0", logs: logs, sub: sub}, nil
}

// WatchAlterOwner0 is a free log subscription operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Indexer *IndexerFilterer) WatchAlterOwner0(opts *bind.WatchOpts, sink chan<- *IndexerAlterOwner0) (event.Subscription, error) {

	logs, sub, err := _Indexer.contract.WatchLogs(opts, "AlterOwner0")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IndexerAlterOwner0)
				if err := _Indexer.contract.UnpackLog(event, "AlterOwner0", log); err != nil {
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

// ParseAlterOwner0 is a log parse operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_Indexer *IndexerFilterer) ParseAlterOwner0(log types.Log) (*IndexerAlterOwner0, error) {
	event := new(IndexerAlterOwner0)
	if err := _Indexer.contract.UnpackLog(event, "AlterOwner0", log); err != nil {
		return nil, err
	}
	return event, nil
}

// IndexerAlterResolverIterator is returned from FilterAlterResolver and is used to iterate over the raw logs and unpacked data for AlterResolver events raised by the Indexer contract.
type IndexerAlterResolverIterator struct {
	Event *IndexerAlterResolver // Event containing the contract specifics and raw log

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
func (it *IndexerAlterResolverIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IndexerAlterResolver)
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
		it.Event = new(IndexerAlterResolver)
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
func (it *IndexerAlterResolverIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IndexerAlterResolverIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IndexerAlterResolver represents a AlterResolver event raised by the Indexer contract.
type IndexerAlterResolver struct {
	Key         string
	OldResolver common.Address
	NewResolver common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterAlterResolver is a free log retrieval operation binding the contract event 0x0a7047ba8be4d874e67aebc953a70ff6db03a81782549290ac646e0738ddfc04.
//
// Solidity: event AlterResolver(string key, address oldResolver, address newResolver)
func (_Indexer *IndexerFilterer) FilterAlterResolver(opts *bind.FilterOpts) (*IndexerAlterResolverIterator, error) {

	logs, sub, err := _Indexer.contract.FilterLogs(opts, "AlterResolver")
	if err != nil {
		return nil, err
	}
	return &IndexerAlterResolverIterator{contract: _Indexer.contract, event: "AlterResolver", logs: logs, sub: sub}, nil
}

// WatchAlterResolver is a free log subscription operation binding the contract event 0x0a7047ba8be4d874e67aebc953a70ff6db03a81782549290ac646e0738ddfc04.
//
// Solidity: event AlterResolver(string key, address oldResolver, address newResolver)
func (_Indexer *IndexerFilterer) WatchAlterResolver(opts *bind.WatchOpts, sink chan<- *IndexerAlterResolver) (event.Subscription, error) {

	logs, sub, err := _Indexer.contract.WatchLogs(opts, "AlterResolver")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IndexerAlterResolver)
				if err := _Indexer.contract.UnpackLog(event, "AlterResolver", log); err != nil {
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

// ParseAlterResolver is a log parse operation binding the contract event 0x0a7047ba8be4d874e67aebc953a70ff6db03a81782549290ac646e0738ddfc04.
//
// Solidity: event AlterResolver(string key, address oldResolver, address newResolver)
func (_Indexer *IndexerFilterer) ParseAlterResolver(log types.Log) (*IndexerAlterResolver, error) {
	event := new(IndexerAlterResolver)
	if err := _Indexer.contract.UnpackLog(event, "AlterResolver", log); err != nil {
		return nil, err
	}
	return event, nil
}

// IndexerErrorIterator is returned from FilterError and is used to iterate over the raw logs and unpacked data for Error events raised by the Indexer contract.
type IndexerErrorIterator struct {
	Event *IndexerError // Event containing the contract specifics and raw log

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
func (it *IndexerErrorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IndexerError)
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
		it.Event = new(IndexerError)
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
func (it *IndexerErrorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IndexerErrorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IndexerError represents a Error event raised by the Indexer contract.
type IndexerError struct {
	Data string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterError is a free log retrieval operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Indexer *IndexerFilterer) FilterError(opts *bind.FilterOpts) (*IndexerErrorIterator, error) {

	logs, sub, err := _Indexer.contract.FilterLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return &IndexerErrorIterator{contract: _Indexer.contract, event: "Error", logs: logs, sub: sub}, nil
}

// WatchError is a free log subscription operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Indexer *IndexerFilterer) WatchError(opts *bind.WatchOpts, sink chan<- *IndexerError) (event.Subscription, error) {

	logs, sub, err := _Indexer.contract.WatchLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IndexerError)
				if err := _Indexer.contract.UnpackLog(event, "Error", log); err != nil {
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
func (_Indexer *IndexerFilterer) ParseError(log types.Log) (*IndexerError, error) {
	event := new(IndexerError)
	if err := _Indexer.contract.UnpackLog(event, "Error", log); err != nil {
		return nil, err
	}
	return event, nil
}
