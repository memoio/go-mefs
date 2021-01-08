// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package adminOwned

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

// AdminOwnedABI is the input ABI used to generate the binding from.
const AdminOwnedABI = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterAdminOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"version\",\"type\":\"uint16\"}],\"name\":\"SetBanned\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newAdminOwner\",\"type\":\"address\"}],\"name\":\"alterAdminOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAdminOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getChannelBannedVersion\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getKPMapBannedVersion\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getKeeperBannedVersion\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMapperBannedVersion\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOfferBannedVersion\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getProviderBannedVersion\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getQueryBannedVersion\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getRootBannedVersion\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getUpkeepingBannedVersion\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"v\",\"type\":\"uint16\"}],\"name\":\"setChannelBannedVersion\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"v\",\"type\":\"uint16\"}],\"name\":\"setKPMapBannedVersion\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"v\",\"type\":\"uint16\"}],\"name\":\"setKeeperBannedVersion\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"v\",\"type\":\"uint16\"}],\"name\":\"setMapperBannedVersion\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"v\",\"type\":\"uint16\"}],\"name\":\"setOfferBannedVersion\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"v\",\"type\":\"uint16\"}],\"name\":\"setProviderBannedVersion\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"v\",\"type\":\"uint16\"}],\"name\":\"setQueryBannedVersion\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"v\",\"type\":\"uint16\"}],\"name\":\"setRootBannedVersion\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"v\",\"type\":\"uint16\"}],\"name\":\"setUpkeepingBannedVersion\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// AdminOwnedBin is the compiled bytecode used for deploying new contracts.
var AdminOwnedBin = "0x608060405234801561001057600080fd5b50336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555061148e806100606000396000f3fe608060405234801561001057600080fd5b506004361061012c5760003560e01c80637efa8370116100ad578063d715c85e11610071578063d715c85e146103e7578063de60908a14610409578063e99680b11461042b578063f23cc21c1461045d578063f49ded5a146104915761012c565b80637efa8370146102fd5780638044c8011461032f578063a06b7cfa14610361578063af484b3814610383578063c304b43f146103b55761012c565b806350523e07116100f457806350523e07146101eb57806350d38a991461021d57806353e6d3921461024f578063597e409d146102a95780637ce82a90146102cb5761012c565b8063073eeb531461013157806326b3eb761461015357806333c767721461018557806334b9d634146101a75780634410bb05146101c9575b600080fd5b6101396104b3565b604051808261ffff16815260200191505060405180910390f35b6101836004803603602081101561016957600080fd5b81019080803561ffff1690602001909291905050506104cb565b005b61018d61063e565b604051808261ffff16815260200191505060405180910390f35b6101af610655565b604051808261ffff16815260200191505060405180910390f35b6101d161066c565b604051808261ffff16815260200191505060405180910390f35b61021b6004803603602081101561020157600080fd5b81019080803561ffff169060200190929190505050610683565b005b61024d6004803603602081101561023357600080fd5b81019080803561ffff1690602001909291905050506107f6565b005b6102916004803603602081101561026557600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610969565b60405180821515815260200191505060405180910390f35b6102b1610b08565b604051808261ffff16815260200191505060405180910390f35b6102fb600480360360208110156102e157600080fd5b81019080803561ffff169060200190929190505050610b20565b005b61032d6004803603602081101561031357600080fd5b81019080803561ffff169060200190929190505050610c93565b005b61035f6004803603602081101561034557600080fd5b81019080803561ffff169060200190929190505050610e06565b005b610369610f79565b604051808261ffff16815260200191505060405180910390f35b6103b36004803603602081101561039957600080fd5b81019080803561ffff169060200190929190505050610f90565b005b6103e5600480360360208110156103cb57600080fd5b81019080803561ffff169060200190929190505050611103565b005b6103ef611276565b604051808261ffff16815260200191505060405180910390f35b61041161128e565b604051808261ffff16815260200191505060405180910390f35b61045b6004803603602081101561044157600080fd5b81019080803561ffff1690602001909291905050506112a5565b005b610465611418565b604051808273ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b610499611441565b604051808261ffff16815260200191505060405180910390f35b6000600160009054906101000a900461ffff16905090565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461058c576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b806000601e6101000a81548161ffff021916908361ffff1602179055507fefd4f42e8a20becead4ea7727277fe199cfedb91fea800d50aa1466e01b4a1c9338260405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018361ffff168152602001828103825260078152602001807f6368616e6e656c00000000000000000000000000000000000000000000000000815250602001935050505060405180910390a150565b60008060169054906101000a900461ffff16905090565b600080601c9054906101000a900461ffff16905090565b600080601a9054906101000a900461ffff16905090565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610744576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b80600160046101000a81548161ffff021916908361ffff1602179055507fefd4f42e8a20becead4ea7727277fe199cfedb91fea800d50aa1466e01b4a1c9338260405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018361ffff168152602001828103825260058152602001807f6b704d6170000000000000000000000000000000000000000000000000000000815250602001935050505060405180910390a150565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146108b7576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b80600060146101000a81548161ffff021916908361ffff1602179055507fefd4f42e8a20becead4ea7727277fe199cfedb91fea800d50aa1466e01b4a1c9338260405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018361ffff168152602001828103825260048152602001807f726f6f7400000000000000000000000000000000000000000000000000000000815250602001935050505060405180910390a150565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610a2d576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050826000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f88632f39007912d02dba5583fb689a48338d8a1b0358c8287259a22516517d898184604051808373ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a16001915050919050565b6000600160029054906101000a900461ffff16905090565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610be1576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b80600060186101000a81548161ffff021916908361ffff1602179055507fefd4f42e8a20becead4ea7727277fe199cfedb91fea800d50aa1466e01b4a1c9338260405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018361ffff168152602001828103825260058152602001807f7175657279000000000000000000000000000000000000000000000000000000815250602001935050505060405180910390a150565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610d54576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b80600060166101000a81548161ffff021916908361ffff1602179055507fefd4f42e8a20becead4ea7727277fe199cfedb91fea800d50aa1466e01b4a1c9338260405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018361ffff168152602001828103825260068152602001807f6d61707065720000000000000000000000000000000000000000000000000000815250602001935050505060405180910390a150565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610ec7576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b806000601a6101000a81548161ffff021916908361ffff1602179055507fefd4f42e8a20becead4ea7727277fe199cfedb91fea800d50aa1466e01b4a1c9338260405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018361ffff168152602001828103825260058152602001807f6f66666572000000000000000000000000000000000000000000000000000000815250602001935050505060405180910390a150565b60008060149054906101000a900461ffff16905090565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614611051576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b80600160006101000a81548161ffff021916908361ffff1602179055507fefd4f42e8a20becead4ea7727277fe199cfedb91fea800d50aa1466e01b4a1c9338260405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018361ffff168152602001828103825260068152602001807f6b65657065720000000000000000000000000000000000000000000000000000815250602001935050505060405180910390a150565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146111c4576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b806000601c6101000a81548161ffff021916908361ffff1602179055507fefd4f42e8a20becead4ea7727277fe199cfedb91fea800d50aa1466e01b4a1c9338260405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018361ffff168152602001828103825260098152602001807f75706b656570696e670000000000000000000000000000000000000000000000815250602001935050505060405180910390a150565b6000600160049054906101000a900461ffff16905090565b600080601e9054906101000a900461ffff16905090565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614611366576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b80600160026101000a81548161ffff021916908361ffff1602179055507fefd4f42e8a20becead4ea7727277fe199cfedb91fea800d50aa1466e01b4a1c9338260405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018361ffff168152602001828103825260088152602001807f70726f7669646572000000000000000000000000000000000000000000000000815250602001935050505060405180910390a150565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b60008060189054906101000a900461ffff1690509056fea2646970667358221220c5d3754eb2696df3b3be194d0af6131de5bc67fc7d185526b06bb34d677447e564736f6c63430007030033"

// DeployAdminOwned deploys a new Ethereum contract, binding an instance of AdminOwned to it.
func DeployAdminOwned(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *AdminOwned, error) {
	parsed, err := abi.JSON(strings.NewReader(AdminOwnedABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(AdminOwnedBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &AdminOwned{AdminOwnedCaller: AdminOwnedCaller{contract: contract}, AdminOwnedTransactor: AdminOwnedTransactor{contract: contract}, AdminOwnedFilterer: AdminOwnedFilterer{contract: contract}}, nil
}

// AdminOwned is an auto generated Go binding around an Ethereum contract.
type AdminOwned struct {
	AdminOwnedCaller     // Read-only binding to the contract
	AdminOwnedTransactor // Write-only binding to the contract
	AdminOwnedFilterer   // Log filterer for contract events
}

// AdminOwnedCaller is an auto generated read-only Go binding around an Ethereum contract.
type AdminOwnedCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AdminOwnedTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AdminOwnedTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AdminOwnedFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AdminOwnedFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AdminOwnedSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AdminOwnedSession struct {
	Contract     *AdminOwned       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AdminOwnedCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AdminOwnedCallerSession struct {
	Contract *AdminOwnedCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// AdminOwnedTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AdminOwnedTransactorSession struct {
	Contract     *AdminOwnedTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// AdminOwnedRaw is an auto generated low-level Go binding around an Ethereum contract.
type AdminOwnedRaw struct {
	Contract *AdminOwned // Generic contract binding to access the raw methods on
}

// AdminOwnedCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AdminOwnedCallerRaw struct {
	Contract *AdminOwnedCaller // Generic read-only contract binding to access the raw methods on
}

// AdminOwnedTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AdminOwnedTransactorRaw struct {
	Contract *AdminOwnedTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAdminOwned creates a new instance of AdminOwned, bound to a specific deployed contract.
func NewAdminOwned(address common.Address, backend bind.ContractBackend) (*AdminOwned, error) {
	contract, err := bindAdminOwned(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AdminOwned{AdminOwnedCaller: AdminOwnedCaller{contract: contract}, AdminOwnedTransactor: AdminOwnedTransactor{contract: contract}, AdminOwnedFilterer: AdminOwnedFilterer{contract: contract}}, nil
}

// NewAdminOwnedCaller creates a new read-only instance of AdminOwned, bound to a specific deployed contract.
func NewAdminOwnedCaller(address common.Address, caller bind.ContractCaller) (*AdminOwnedCaller, error) {
	contract, err := bindAdminOwned(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AdminOwnedCaller{contract: contract}, nil
}

// NewAdminOwnedTransactor creates a new write-only instance of AdminOwned, bound to a specific deployed contract.
func NewAdminOwnedTransactor(address common.Address, transactor bind.ContractTransactor) (*AdminOwnedTransactor, error) {
	contract, err := bindAdminOwned(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AdminOwnedTransactor{contract: contract}, nil
}

// NewAdminOwnedFilterer creates a new log filterer instance of AdminOwned, bound to a specific deployed contract.
func NewAdminOwnedFilterer(address common.Address, filterer bind.ContractFilterer) (*AdminOwnedFilterer, error) {
	contract, err := bindAdminOwned(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AdminOwnedFilterer{contract: contract}, nil
}

// bindAdminOwned binds a generic wrapper to an already deployed contract.
func bindAdminOwned(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AdminOwnedABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AdminOwned *AdminOwnedRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _AdminOwned.Contract.AdminOwnedCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AdminOwned *AdminOwnedRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AdminOwned.Contract.AdminOwnedTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AdminOwned *AdminOwnedRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AdminOwned.Contract.AdminOwnedTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AdminOwned *AdminOwnedCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _AdminOwned.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AdminOwned *AdminOwnedTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AdminOwned.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AdminOwned *AdminOwnedTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AdminOwned.Contract.contract.Transact(opts, method, params...)
}

// GetAdminOwner is a free data retrieval call binding the contract method 0xf23cc21c.
//
// Solidity: function getAdminOwner() view returns(address)
func (_AdminOwned *AdminOwnedCaller) GetAdminOwner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _AdminOwned.contract.Call(opts, out, "getAdminOwner")
	return *ret0, err
}

// GetAdminOwner is a free data retrieval call binding the contract method 0xf23cc21c.
//
// Solidity: function getAdminOwner() view returns(address)
func (_AdminOwned *AdminOwnedSession) GetAdminOwner() (common.Address, error) {
	return _AdminOwned.Contract.GetAdminOwner(&_AdminOwned.CallOpts)
}

// GetAdminOwner is a free data retrieval call binding the contract method 0xf23cc21c.
//
// Solidity: function getAdminOwner() view returns(address)
func (_AdminOwned *AdminOwnedCallerSession) GetAdminOwner() (common.Address, error) {
	return _AdminOwned.Contract.GetAdminOwner(&_AdminOwned.CallOpts)
}

// GetChannelBannedVersion is a free data retrieval call binding the contract method 0xde60908a.
//
// Solidity: function getChannelBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedCaller) GetChannelBannedVersion(opts *bind.CallOpts) (uint16, error) {
	var (
		ret0 = new(uint16)
	)
	out := ret0
	err := _AdminOwned.contract.Call(opts, out, "getChannelBannedVersion")
	return *ret0, err
}

// GetChannelBannedVersion is a free data retrieval call binding the contract method 0xde60908a.
//
// Solidity: function getChannelBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedSession) GetChannelBannedVersion() (uint16, error) {
	return _AdminOwned.Contract.GetChannelBannedVersion(&_AdminOwned.CallOpts)
}

// GetChannelBannedVersion is a free data retrieval call binding the contract method 0xde60908a.
//
// Solidity: function getChannelBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedCallerSession) GetChannelBannedVersion() (uint16, error) {
	return _AdminOwned.Contract.GetChannelBannedVersion(&_AdminOwned.CallOpts)
}

// GetKPMapBannedVersion is a free data retrieval call binding the contract method 0xd715c85e.
//
// Solidity: function getKPMapBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedCaller) GetKPMapBannedVersion(opts *bind.CallOpts) (uint16, error) {
	var (
		ret0 = new(uint16)
	)
	out := ret0
	err := _AdminOwned.contract.Call(opts, out, "getKPMapBannedVersion")
	return *ret0, err
}

// GetKPMapBannedVersion is a free data retrieval call binding the contract method 0xd715c85e.
//
// Solidity: function getKPMapBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedSession) GetKPMapBannedVersion() (uint16, error) {
	return _AdminOwned.Contract.GetKPMapBannedVersion(&_AdminOwned.CallOpts)
}

// GetKPMapBannedVersion is a free data retrieval call binding the contract method 0xd715c85e.
//
// Solidity: function getKPMapBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedCallerSession) GetKPMapBannedVersion() (uint16, error) {
	return _AdminOwned.Contract.GetKPMapBannedVersion(&_AdminOwned.CallOpts)
}

// GetKeeperBannedVersion is a free data retrieval call binding the contract method 0x073eeb53.
//
// Solidity: function getKeeperBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedCaller) GetKeeperBannedVersion(opts *bind.CallOpts) (uint16, error) {
	var (
		ret0 = new(uint16)
	)
	out := ret0
	err := _AdminOwned.contract.Call(opts, out, "getKeeperBannedVersion")
	return *ret0, err
}

// GetKeeperBannedVersion is a free data retrieval call binding the contract method 0x073eeb53.
//
// Solidity: function getKeeperBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedSession) GetKeeperBannedVersion() (uint16, error) {
	return _AdminOwned.Contract.GetKeeperBannedVersion(&_AdminOwned.CallOpts)
}

// GetKeeperBannedVersion is a free data retrieval call binding the contract method 0x073eeb53.
//
// Solidity: function getKeeperBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedCallerSession) GetKeeperBannedVersion() (uint16, error) {
	return _AdminOwned.Contract.GetKeeperBannedVersion(&_AdminOwned.CallOpts)
}

// GetMapperBannedVersion is a free data retrieval call binding the contract method 0x33c76772.
//
// Solidity: function getMapperBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedCaller) GetMapperBannedVersion(opts *bind.CallOpts) (uint16, error) {
	var (
		ret0 = new(uint16)
	)
	out := ret0
	err := _AdminOwned.contract.Call(opts, out, "getMapperBannedVersion")
	return *ret0, err
}

// GetMapperBannedVersion is a free data retrieval call binding the contract method 0x33c76772.
//
// Solidity: function getMapperBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedSession) GetMapperBannedVersion() (uint16, error) {
	return _AdminOwned.Contract.GetMapperBannedVersion(&_AdminOwned.CallOpts)
}

// GetMapperBannedVersion is a free data retrieval call binding the contract method 0x33c76772.
//
// Solidity: function getMapperBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedCallerSession) GetMapperBannedVersion() (uint16, error) {
	return _AdminOwned.Contract.GetMapperBannedVersion(&_AdminOwned.CallOpts)
}

// GetOfferBannedVersion is a free data retrieval call binding the contract method 0x4410bb05.
//
// Solidity: function getOfferBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedCaller) GetOfferBannedVersion(opts *bind.CallOpts) (uint16, error) {
	var (
		ret0 = new(uint16)
	)
	out := ret0
	err := _AdminOwned.contract.Call(opts, out, "getOfferBannedVersion")
	return *ret0, err
}

// GetOfferBannedVersion is a free data retrieval call binding the contract method 0x4410bb05.
//
// Solidity: function getOfferBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedSession) GetOfferBannedVersion() (uint16, error) {
	return _AdminOwned.Contract.GetOfferBannedVersion(&_AdminOwned.CallOpts)
}

// GetOfferBannedVersion is a free data retrieval call binding the contract method 0x4410bb05.
//
// Solidity: function getOfferBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedCallerSession) GetOfferBannedVersion() (uint16, error) {
	return _AdminOwned.Contract.GetOfferBannedVersion(&_AdminOwned.CallOpts)
}

// GetProviderBannedVersion is a free data retrieval call binding the contract method 0x597e409d.
//
// Solidity: function getProviderBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedCaller) GetProviderBannedVersion(opts *bind.CallOpts) (uint16, error) {
	var (
		ret0 = new(uint16)
	)
	out := ret0
	err := _AdminOwned.contract.Call(opts, out, "getProviderBannedVersion")
	return *ret0, err
}

// GetProviderBannedVersion is a free data retrieval call binding the contract method 0x597e409d.
//
// Solidity: function getProviderBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedSession) GetProviderBannedVersion() (uint16, error) {
	return _AdminOwned.Contract.GetProviderBannedVersion(&_AdminOwned.CallOpts)
}

// GetProviderBannedVersion is a free data retrieval call binding the contract method 0x597e409d.
//
// Solidity: function getProviderBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedCallerSession) GetProviderBannedVersion() (uint16, error) {
	return _AdminOwned.Contract.GetProviderBannedVersion(&_AdminOwned.CallOpts)
}

// GetQueryBannedVersion is a free data retrieval call binding the contract method 0xf49ded5a.
//
// Solidity: function getQueryBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedCaller) GetQueryBannedVersion(opts *bind.CallOpts) (uint16, error) {
	var (
		ret0 = new(uint16)
	)
	out := ret0
	err := _AdminOwned.contract.Call(opts, out, "getQueryBannedVersion")
	return *ret0, err
}

// GetQueryBannedVersion is a free data retrieval call binding the contract method 0xf49ded5a.
//
// Solidity: function getQueryBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedSession) GetQueryBannedVersion() (uint16, error) {
	return _AdminOwned.Contract.GetQueryBannedVersion(&_AdminOwned.CallOpts)
}

// GetQueryBannedVersion is a free data retrieval call binding the contract method 0xf49ded5a.
//
// Solidity: function getQueryBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedCallerSession) GetQueryBannedVersion() (uint16, error) {
	return _AdminOwned.Contract.GetQueryBannedVersion(&_AdminOwned.CallOpts)
}

// GetRootBannedVersion is a free data retrieval call binding the contract method 0xa06b7cfa.
//
// Solidity: function getRootBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedCaller) GetRootBannedVersion(opts *bind.CallOpts) (uint16, error) {
	var (
		ret0 = new(uint16)
	)
	out := ret0
	err := _AdminOwned.contract.Call(opts, out, "getRootBannedVersion")
	return *ret0, err
}

// GetRootBannedVersion is a free data retrieval call binding the contract method 0xa06b7cfa.
//
// Solidity: function getRootBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedSession) GetRootBannedVersion() (uint16, error) {
	return _AdminOwned.Contract.GetRootBannedVersion(&_AdminOwned.CallOpts)
}

// GetRootBannedVersion is a free data retrieval call binding the contract method 0xa06b7cfa.
//
// Solidity: function getRootBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedCallerSession) GetRootBannedVersion() (uint16, error) {
	return _AdminOwned.Contract.GetRootBannedVersion(&_AdminOwned.CallOpts)
}

// GetUpkeepingBannedVersion is a free data retrieval call binding the contract method 0x34b9d634.
//
// Solidity: function getUpkeepingBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedCaller) GetUpkeepingBannedVersion(opts *bind.CallOpts) (uint16, error) {
	var (
		ret0 = new(uint16)
	)
	out := ret0
	err := _AdminOwned.contract.Call(opts, out, "getUpkeepingBannedVersion")
	return *ret0, err
}

// GetUpkeepingBannedVersion is a free data retrieval call binding the contract method 0x34b9d634.
//
// Solidity: function getUpkeepingBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedSession) GetUpkeepingBannedVersion() (uint16, error) {
	return _AdminOwned.Contract.GetUpkeepingBannedVersion(&_AdminOwned.CallOpts)
}

// GetUpkeepingBannedVersion is a free data retrieval call binding the contract method 0x34b9d634.
//
// Solidity: function getUpkeepingBannedVersion() view returns(uint16)
func (_AdminOwned *AdminOwnedCallerSession) GetUpkeepingBannedVersion() (uint16, error) {
	return _AdminOwned.Contract.GetUpkeepingBannedVersion(&_AdminOwned.CallOpts)
}

// AlterAdminOwner is a paid mutator transaction binding the contract method 0x53e6d392.
//
// Solidity: function alterAdminOwner(address newAdminOwner) returns(bool)
func (_AdminOwned *AdminOwnedTransactor) AlterAdminOwner(opts *bind.TransactOpts, newAdminOwner common.Address) (*types.Transaction, error) {
	return _AdminOwned.contract.Transact(opts, "alterAdminOwner", newAdminOwner)
}

// AlterAdminOwner is a paid mutator transaction binding the contract method 0x53e6d392.
//
// Solidity: function alterAdminOwner(address newAdminOwner) returns(bool)
func (_AdminOwned *AdminOwnedSession) AlterAdminOwner(newAdminOwner common.Address) (*types.Transaction, error) {
	return _AdminOwned.Contract.AlterAdminOwner(&_AdminOwned.TransactOpts, newAdminOwner)
}

// AlterAdminOwner is a paid mutator transaction binding the contract method 0x53e6d392.
//
// Solidity: function alterAdminOwner(address newAdminOwner) returns(bool)
func (_AdminOwned *AdminOwnedTransactorSession) AlterAdminOwner(newAdminOwner common.Address) (*types.Transaction, error) {
	return _AdminOwned.Contract.AlterAdminOwner(&_AdminOwned.TransactOpts, newAdminOwner)
}

// SetChannelBannedVersion is a paid mutator transaction binding the contract method 0x26b3eb76.
//
// Solidity: function setChannelBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedTransactor) SetChannelBannedVersion(opts *bind.TransactOpts, v uint16) (*types.Transaction, error) {
	return _AdminOwned.contract.Transact(opts, "setChannelBannedVersion", v)
}

// SetChannelBannedVersion is a paid mutator transaction binding the contract method 0x26b3eb76.
//
// Solidity: function setChannelBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedSession) SetChannelBannedVersion(v uint16) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetChannelBannedVersion(&_AdminOwned.TransactOpts, v)
}

// SetChannelBannedVersion is a paid mutator transaction binding the contract method 0x26b3eb76.
//
// Solidity: function setChannelBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedTransactorSession) SetChannelBannedVersion(v uint16) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetChannelBannedVersion(&_AdminOwned.TransactOpts, v)
}

// SetKPMapBannedVersion is a paid mutator transaction binding the contract method 0x50523e07.
//
// Solidity: function setKPMapBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedTransactor) SetKPMapBannedVersion(opts *bind.TransactOpts, v uint16) (*types.Transaction, error) {
	return _AdminOwned.contract.Transact(opts, "setKPMapBannedVersion", v)
}

// SetKPMapBannedVersion is a paid mutator transaction binding the contract method 0x50523e07.
//
// Solidity: function setKPMapBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedSession) SetKPMapBannedVersion(v uint16) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetKPMapBannedVersion(&_AdminOwned.TransactOpts, v)
}

// SetKPMapBannedVersion is a paid mutator transaction binding the contract method 0x50523e07.
//
// Solidity: function setKPMapBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedTransactorSession) SetKPMapBannedVersion(v uint16) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetKPMapBannedVersion(&_AdminOwned.TransactOpts, v)
}

// SetKeeperBannedVersion is a paid mutator transaction binding the contract method 0xaf484b38.
//
// Solidity: function setKeeperBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedTransactor) SetKeeperBannedVersion(opts *bind.TransactOpts, v uint16) (*types.Transaction, error) {
	return _AdminOwned.contract.Transact(opts, "setKeeperBannedVersion", v)
}

// SetKeeperBannedVersion is a paid mutator transaction binding the contract method 0xaf484b38.
//
// Solidity: function setKeeperBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedSession) SetKeeperBannedVersion(v uint16) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetKeeperBannedVersion(&_AdminOwned.TransactOpts, v)
}

// SetKeeperBannedVersion is a paid mutator transaction binding the contract method 0xaf484b38.
//
// Solidity: function setKeeperBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedTransactorSession) SetKeeperBannedVersion(v uint16) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetKeeperBannedVersion(&_AdminOwned.TransactOpts, v)
}

// SetMapperBannedVersion is a paid mutator transaction binding the contract method 0x7efa8370.
//
// Solidity: function setMapperBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedTransactor) SetMapperBannedVersion(opts *bind.TransactOpts, v uint16) (*types.Transaction, error) {
	return _AdminOwned.contract.Transact(opts, "setMapperBannedVersion", v)
}

// SetMapperBannedVersion is a paid mutator transaction binding the contract method 0x7efa8370.
//
// Solidity: function setMapperBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedSession) SetMapperBannedVersion(v uint16) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetMapperBannedVersion(&_AdminOwned.TransactOpts, v)
}

// SetMapperBannedVersion is a paid mutator transaction binding the contract method 0x7efa8370.
//
// Solidity: function setMapperBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedTransactorSession) SetMapperBannedVersion(v uint16) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetMapperBannedVersion(&_AdminOwned.TransactOpts, v)
}

// SetOfferBannedVersion is a paid mutator transaction binding the contract method 0x8044c801.
//
// Solidity: function setOfferBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedTransactor) SetOfferBannedVersion(opts *bind.TransactOpts, v uint16) (*types.Transaction, error) {
	return _AdminOwned.contract.Transact(opts, "setOfferBannedVersion", v)
}

// SetOfferBannedVersion is a paid mutator transaction binding the contract method 0x8044c801.
//
// Solidity: function setOfferBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedSession) SetOfferBannedVersion(v uint16) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetOfferBannedVersion(&_AdminOwned.TransactOpts, v)
}

// SetOfferBannedVersion is a paid mutator transaction binding the contract method 0x8044c801.
//
// Solidity: function setOfferBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedTransactorSession) SetOfferBannedVersion(v uint16) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetOfferBannedVersion(&_AdminOwned.TransactOpts, v)
}

// SetProviderBannedVersion is a paid mutator transaction binding the contract method 0xe99680b1.
//
// Solidity: function setProviderBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedTransactor) SetProviderBannedVersion(opts *bind.TransactOpts, v uint16) (*types.Transaction, error) {
	return _AdminOwned.contract.Transact(opts, "setProviderBannedVersion", v)
}

// SetProviderBannedVersion is a paid mutator transaction binding the contract method 0xe99680b1.
//
// Solidity: function setProviderBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedSession) SetProviderBannedVersion(v uint16) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetProviderBannedVersion(&_AdminOwned.TransactOpts, v)
}

// SetProviderBannedVersion is a paid mutator transaction binding the contract method 0xe99680b1.
//
// Solidity: function setProviderBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedTransactorSession) SetProviderBannedVersion(v uint16) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetProviderBannedVersion(&_AdminOwned.TransactOpts, v)
}

// SetQueryBannedVersion is a paid mutator transaction binding the contract method 0x7ce82a90.
//
// Solidity: function setQueryBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedTransactor) SetQueryBannedVersion(opts *bind.TransactOpts, v uint16) (*types.Transaction, error) {
	return _AdminOwned.contract.Transact(opts, "setQueryBannedVersion", v)
}

// SetQueryBannedVersion is a paid mutator transaction binding the contract method 0x7ce82a90.
//
// Solidity: function setQueryBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedSession) SetQueryBannedVersion(v uint16) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetQueryBannedVersion(&_AdminOwned.TransactOpts, v)
}

// SetQueryBannedVersion is a paid mutator transaction binding the contract method 0x7ce82a90.
//
// Solidity: function setQueryBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedTransactorSession) SetQueryBannedVersion(v uint16) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetQueryBannedVersion(&_AdminOwned.TransactOpts, v)
}

// SetRootBannedVersion is a paid mutator transaction binding the contract method 0x50d38a99.
//
// Solidity: function setRootBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedTransactor) SetRootBannedVersion(opts *bind.TransactOpts, v uint16) (*types.Transaction, error) {
	return _AdminOwned.contract.Transact(opts, "setRootBannedVersion", v)
}

// SetRootBannedVersion is a paid mutator transaction binding the contract method 0x50d38a99.
//
// Solidity: function setRootBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedSession) SetRootBannedVersion(v uint16) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetRootBannedVersion(&_AdminOwned.TransactOpts, v)
}

// SetRootBannedVersion is a paid mutator transaction binding the contract method 0x50d38a99.
//
// Solidity: function setRootBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedTransactorSession) SetRootBannedVersion(v uint16) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetRootBannedVersion(&_AdminOwned.TransactOpts, v)
}

// SetUpkeepingBannedVersion is a paid mutator transaction binding the contract method 0xc304b43f.
//
// Solidity: function setUpkeepingBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedTransactor) SetUpkeepingBannedVersion(opts *bind.TransactOpts, v uint16) (*types.Transaction, error) {
	return _AdminOwned.contract.Transact(opts, "setUpkeepingBannedVersion", v)
}

// SetUpkeepingBannedVersion is a paid mutator transaction binding the contract method 0xc304b43f.
//
// Solidity: function setUpkeepingBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedSession) SetUpkeepingBannedVersion(v uint16) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetUpkeepingBannedVersion(&_AdminOwned.TransactOpts, v)
}

// SetUpkeepingBannedVersion is a paid mutator transaction binding the contract method 0xc304b43f.
//
// Solidity: function setUpkeepingBannedVersion(uint16 v) returns()
func (_AdminOwned *AdminOwnedTransactorSession) SetUpkeepingBannedVersion(v uint16) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetUpkeepingBannedVersion(&_AdminOwned.TransactOpts, v)
}

// AdminOwnedAlterAdminOwnerIterator is returned from FilterAlterAdminOwner and is used to iterate over the raw logs and unpacked data for AlterAdminOwner events raised by the AdminOwned contract.
type AdminOwnedAlterAdminOwnerIterator struct {
	Event *AdminOwnedAlterAdminOwner // Event containing the contract specifics and raw log

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
func (it *AdminOwnedAlterAdminOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdminOwnedAlterAdminOwner)
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
		it.Event = new(AdminOwnedAlterAdminOwner)
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
func (it *AdminOwnedAlterAdminOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdminOwnedAlterAdminOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdminOwnedAlterAdminOwner represents a AlterAdminOwner event raised by the AdminOwned contract.
type AdminOwnedAlterAdminOwner struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAlterAdminOwner is a free log retrieval operation binding the contract event 0x88632f39007912d02dba5583fb689a48338d8a1b0358c8287259a22516517d89.
//
// Solidity: event AlterAdminOwner(address from, address to)
func (_AdminOwned *AdminOwnedFilterer) FilterAlterAdminOwner(opts *bind.FilterOpts) (*AdminOwnedAlterAdminOwnerIterator, error) {

	logs, sub, err := _AdminOwned.contract.FilterLogs(opts, "AlterAdminOwner")
	if err != nil {
		return nil, err
	}
	return &AdminOwnedAlterAdminOwnerIterator{contract: _AdminOwned.contract, event: "AlterAdminOwner", logs: logs, sub: sub}, nil
}

// WatchAlterAdminOwner is a free log subscription operation binding the contract event 0x88632f39007912d02dba5583fb689a48338d8a1b0358c8287259a22516517d89.
//
// Solidity: event AlterAdminOwner(address from, address to)
func (_AdminOwned *AdminOwnedFilterer) WatchAlterAdminOwner(opts *bind.WatchOpts, sink chan<- *AdminOwnedAlterAdminOwner) (event.Subscription, error) {

	logs, sub, err := _AdminOwned.contract.WatchLogs(opts, "AlterAdminOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdminOwnedAlterAdminOwner)
				if err := _AdminOwned.contract.UnpackLog(event, "AlterAdminOwner", log); err != nil {
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

// ParseAlterAdminOwner is a log parse operation binding the contract event 0x88632f39007912d02dba5583fb689a48338d8a1b0358c8287259a22516517d89.
//
// Solidity: event AlterAdminOwner(address from, address to)
func (_AdminOwned *AdminOwnedFilterer) ParseAlterAdminOwner(log types.Log) (*AdminOwnedAlterAdminOwner, error) {
	event := new(AdminOwnedAlterAdminOwner)
	if err := _AdminOwned.contract.UnpackLog(event, "AlterAdminOwner", log); err != nil {
		return nil, err
	}
	return event, nil
}

// AdminOwnedSetBannedIterator is returned from FilterSetBanned and is used to iterate over the raw logs and unpacked data for SetBanned events raised by the AdminOwned contract.
type AdminOwnedSetBannedIterator struct {
	Event *AdminOwnedSetBanned // Event containing the contract specifics and raw log

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
func (it *AdminOwnedSetBannedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AdminOwnedSetBanned)
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
		it.Event = new(AdminOwnedSetBanned)
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
func (it *AdminOwnedSetBannedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AdminOwnedSetBannedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AdminOwnedSetBanned represents a SetBanned event raised by the AdminOwned contract.
type AdminOwnedSetBanned struct {
	Key     string
	From    common.Address
	Version uint16
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterSetBanned is a free log retrieval operation binding the contract event 0xefd4f42e8a20becead4ea7727277fe199cfedb91fea800d50aa1466e01b4a1c9.
//
// Solidity: event SetBanned(string key, address from, uint16 version)
func (_AdminOwned *AdminOwnedFilterer) FilterSetBanned(opts *bind.FilterOpts) (*AdminOwnedSetBannedIterator, error) {

	logs, sub, err := _AdminOwned.contract.FilterLogs(opts, "SetBanned")
	if err != nil {
		return nil, err
	}
	return &AdminOwnedSetBannedIterator{contract: _AdminOwned.contract, event: "SetBanned", logs: logs, sub: sub}, nil
}

// WatchSetBanned is a free log subscription operation binding the contract event 0xefd4f42e8a20becead4ea7727277fe199cfedb91fea800d50aa1466e01b4a1c9.
//
// Solidity: event SetBanned(string key, address from, uint16 version)
func (_AdminOwned *AdminOwnedFilterer) WatchSetBanned(opts *bind.WatchOpts, sink chan<- *AdminOwnedSetBanned) (event.Subscription, error) {

	logs, sub, err := _AdminOwned.contract.WatchLogs(opts, "SetBanned")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AdminOwnedSetBanned)
				if err := _AdminOwned.contract.UnpackLog(event, "SetBanned", log); err != nil {
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

// ParseSetBanned is a log parse operation binding the contract event 0xefd4f42e8a20becead4ea7727277fe199cfedb91fea800d50aa1466e01b4a1c9.
//
// Solidity: event SetBanned(string key, address from, uint16 version)
func (_AdminOwned *AdminOwnedFilterer) ParseSetBanned(log types.Log) (*AdminOwnedSetBanned, error) {
	event := new(AdminOwnedSetBanned)
	if err := _AdminOwned.contract.UnpackLog(event, "SetBanned", log); err != nil {
		return nil, err
	}
	return event, nil
}
