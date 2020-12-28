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
const AdminOwnedABI = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterAdminOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"param\",\"type\":\"bool\"}],\"name\":\"SetBanned\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newAdminOwner\",\"type\":\"address\"}],\"name\":\"alterAdminOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAdminOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getChannelBanned\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getKPMapBanned\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getKeeperBanned\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMapperBanned\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOfferBanned\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getProviderBanned\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getQueryBanned\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getRootBanned\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getUpkeepingBanned\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"param\",\"type\":\"bool\"}],\"name\":\"setChannelBanned\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"param\",\"type\":\"bool\"}],\"name\":\"setKPMapBanned\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"param\",\"type\":\"bool\"}],\"name\":\"setKeeperBanned\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"param\",\"type\":\"bool\"}],\"name\":\"setMapperBanned\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"param\",\"type\":\"bool\"}],\"name\":\"setOfferBanned\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"param\",\"type\":\"bool\"}],\"name\":\"setProviderBanned\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"param\",\"type\":\"bool\"}],\"name\":\"setQueryBanned\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"param\",\"type\":\"bool\"}],\"name\":\"setRootBanned\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"param\",\"type\":\"bool\"}],\"name\":\"setUpkeepingBanned\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// AdminOwnedBin is the compiled bytecode used for deploying new contracts.
var AdminOwnedBin = "0x608060405234801561001057600080fd5b50336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550611431806100606000396000f3fe608060405234801561001057600080fd5b506004361061012c5760003560e01c806372857ce2116100ad578063d9d6389e11610071578063d9d6389e146103bb578063e44efb90146103db578063e922e4ce146103fb578063f23cc21c1461042b578063fc3b6a391461045f5761012c565b806372857ce2146102eb578063902b218d1461030b5780639db22ae61461033b578063c04ff8491461036b578063cc893e971461038b5761012c565b8063349a3de2116100f4578063349a3de214610211578063368e63211461023157806353e6d3921461025157806357b6bde6146102ab5780636778d3cb146102cb5761012c565b806306909ba71461013157806316ec989c1461016157806317ef9444146101815780631fed1561146101b15780632321b8ae146101e1575b600080fd5b61015f6004803603602081101561014757600080fd5b8101908080351515906020019092919050505061048f565b005b6101696105fd565b60405180821515815260200191505060405180910390f35b6101af6004803603602081101561019757600080fd5b81019080803515159060200190929190505050610613565b005b6101df600480360360208110156101c757600080fd5b81019080803515159060200190929190505050610781565b005b61020f600480360360208110156101f757600080fd5b810190808035151590602001909291905050506108ef565b005b610219610a5d565b60405180821515815260200191505060405180910390f35b610239610a73565b60405180821515815260200191505060405180910390f35b6102936004803603602081101561026757600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610a89565b60405180821515815260200191505060405180910390f35b6102b3610c28565b60405180821515815260200191505060405180910390f35b6102d3610c3e565b60405180821515815260200191505060405180910390f35b6102f3610c54565b60405180821515815260200191505060405180910390f35b6103396004803603602081101561032157600080fd5b81019080803515159060200190929190505050610c6a565b005b6103696004803603602081101561035157600080fd5b81019080803515159060200190929190505050610dd8565b005b610373610f46565b60405180821515815260200191505060405180910390f35b6103b9600480360360208110156103a157600080fd5b81019080803515159060200190929190505050610f5c565b005b6103c36110ca565b60405180821515815260200191505060405180910390f35b6103e36110e0565b60405180821515815260200191505060405180910390f35b6104296004803603602081101561041157600080fd5b810190808035151590602001909291905050506110f6565b005b610433611264565b604051808273ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b61048d6004803603602081101561047557600080fd5b8101908080351515906020019092919050505061128d565b005b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610550576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b80600060176101000a81548160ff0219169083151502179055507f88a2f5ad849982851810463cc052ff32213c45ba95d13cf29e38f09c5c0c4eca338260405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018315158152602001828103825260058152602001807f6f66666572000000000000000000000000000000000000000000000000000000815250602001935050505060405180910390a150565b60008060169054906101000a900460ff16905090565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146106d4576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b806000601c6101000a81548160ff0219169083151502179055507f88a2f5ad849982851810463cc052ff32213c45ba95d13cf29e38f09c5c0c4eca338260405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018315158152602001828103825260058152602001807f6b704d6170000000000000000000000000000000000000000000000000000000815250602001935050505060405180910390a150565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610842576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b80600060186101000a81548160ff0219169083151502179055507f88a2f5ad849982851810463cc052ff32213c45ba95d13cf29e38f09c5c0c4eca338260405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018315158152602001828103825260098152602001807f75706b656570696e670000000000000000000000000000000000000000000000815250602001935050505060405180910390a150565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146109b0576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b80600060166101000a81548160ff0219169083151502179055507f88a2f5ad849982851810463cc052ff32213c45ba95d13cf29e38f09c5c0c4eca338260405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018315158152602001828103825260058152602001807f7175657279000000000000000000000000000000000000000000000000000000815250602001935050505060405180910390a150565b600080601a9054906101000a900460ff16905090565b60008060159054906101000a900460ff16905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610b4d576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050826000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f88632f39007912d02dba5583fb689a48338d8a1b0358c8287259a22516517d898184604051808373ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a16001915050919050565b600080601b9054906101000a900460ff16905090565b60008060149054906101000a900460ff16905090565b60008060189054906101000a900460ff16905090565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610d2b576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b806000601a6101000a81548160ff0219169083151502179055507f88a2f5ad849982851810463cc052ff32213c45ba95d13cf29e38f09c5c0c4eca338260405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018315158152602001828103825260068152602001807f6b65657065720000000000000000000000000000000000000000000000000000815250602001935050505060405180910390a150565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610e99576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b806000601b6101000a81548160ff0219169083151502179055507f88a2f5ad849982851810463cc052ff32213c45ba95d13cf29e38f09c5c0c4eca338260405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018315158152602001828103825260088152602001807f70726f7669646572000000000000000000000000000000000000000000000000815250602001935050505060405180910390a150565b60008060179054906101000a900460ff16905090565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461101d576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b80600060196101000a81548160ff0219169083151502179055507f88a2f5ad849982851810463cc052ff32213c45ba95d13cf29e38f09c5c0c4eca338260405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018315158152602001828103825260078152602001807f6368616e6e656c00000000000000000000000000000000000000000000000000815250602001935050505060405180910390a150565b600080601c9054906101000a900460ff16905090565b60008060199054906101000a900460ff16905090565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146111b7576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b80600060156101000a81548160ff0219169083151502179055507f88a2f5ad849982851810463cc052ff32213c45ba95d13cf29e38f09c5c0c4eca338260405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018315158152602001828103825260068152602001807f6d61707065720000000000000000000000000000000000000000000000000000815250602001935050505060405180910390a150565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461134e576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260188152602001807f6f6e6c792061646d696e4f776e65722063616e2063616c6c000000000000000081525060200191505060405180910390fd5b80600060146101000a81548160ff0219169083151502179055507f88a2f5ad849982851810463cc052ff32213c45ba95d13cf29e38f09c5c0c4eca338260405180806020018473ffffffffffffffffffffffffffffffffffffffff1681526020018315158152602001828103825260048152602001807f726f6f7400000000000000000000000000000000000000000000000000000000815250602001935050505060405180910390a15056fea2646970667358221220c62c4f5c712095a520bc4aabd3f92778c741d37e9654c014e88fe41ea07b362764736f6c63430007030033"

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

// GetChannelBanned is a free data retrieval call binding the contract method 0xe44efb90.
//
// Solidity: function getChannelBanned() view returns(bool)
func (_AdminOwned *AdminOwnedCaller) GetChannelBanned(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _AdminOwned.contract.Call(opts, out, "getChannelBanned")
	return *ret0, err
}

// GetChannelBanned is a free data retrieval call binding the contract method 0xe44efb90.
//
// Solidity: function getChannelBanned() view returns(bool)
func (_AdminOwned *AdminOwnedSession) GetChannelBanned() (bool, error) {
	return _AdminOwned.Contract.GetChannelBanned(&_AdminOwned.CallOpts)
}

// GetChannelBanned is a free data retrieval call binding the contract method 0xe44efb90.
//
// Solidity: function getChannelBanned() view returns(bool)
func (_AdminOwned *AdminOwnedCallerSession) GetChannelBanned() (bool, error) {
	return _AdminOwned.Contract.GetChannelBanned(&_AdminOwned.CallOpts)
}

// GetKPMapBanned is a free data retrieval call binding the contract method 0xd9d6389e.
//
// Solidity: function getKPMapBanned() view returns(bool)
func (_AdminOwned *AdminOwnedCaller) GetKPMapBanned(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _AdminOwned.contract.Call(opts, out, "getKPMapBanned")
	return *ret0, err
}

// GetKPMapBanned is a free data retrieval call binding the contract method 0xd9d6389e.
//
// Solidity: function getKPMapBanned() view returns(bool)
func (_AdminOwned *AdminOwnedSession) GetKPMapBanned() (bool, error) {
	return _AdminOwned.Contract.GetKPMapBanned(&_AdminOwned.CallOpts)
}

// GetKPMapBanned is a free data retrieval call binding the contract method 0xd9d6389e.
//
// Solidity: function getKPMapBanned() view returns(bool)
func (_AdminOwned *AdminOwnedCallerSession) GetKPMapBanned() (bool, error) {
	return _AdminOwned.Contract.GetKPMapBanned(&_AdminOwned.CallOpts)
}

// GetKeeperBanned is a free data retrieval call binding the contract method 0x349a3de2.
//
// Solidity: function getKeeperBanned() view returns(bool)
func (_AdminOwned *AdminOwnedCaller) GetKeeperBanned(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _AdminOwned.contract.Call(opts, out, "getKeeperBanned")
	return *ret0, err
}

// GetKeeperBanned is a free data retrieval call binding the contract method 0x349a3de2.
//
// Solidity: function getKeeperBanned() view returns(bool)
func (_AdminOwned *AdminOwnedSession) GetKeeperBanned() (bool, error) {
	return _AdminOwned.Contract.GetKeeperBanned(&_AdminOwned.CallOpts)
}

// GetKeeperBanned is a free data retrieval call binding the contract method 0x349a3de2.
//
// Solidity: function getKeeperBanned() view returns(bool)
func (_AdminOwned *AdminOwnedCallerSession) GetKeeperBanned() (bool, error) {
	return _AdminOwned.Contract.GetKeeperBanned(&_AdminOwned.CallOpts)
}

// GetMapperBanned is a free data retrieval call binding the contract method 0x368e6321.
//
// Solidity: function getMapperBanned() view returns(bool)
func (_AdminOwned *AdminOwnedCaller) GetMapperBanned(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _AdminOwned.contract.Call(opts, out, "getMapperBanned")
	return *ret0, err
}

// GetMapperBanned is a free data retrieval call binding the contract method 0x368e6321.
//
// Solidity: function getMapperBanned() view returns(bool)
func (_AdminOwned *AdminOwnedSession) GetMapperBanned() (bool, error) {
	return _AdminOwned.Contract.GetMapperBanned(&_AdminOwned.CallOpts)
}

// GetMapperBanned is a free data retrieval call binding the contract method 0x368e6321.
//
// Solidity: function getMapperBanned() view returns(bool)
func (_AdminOwned *AdminOwnedCallerSession) GetMapperBanned() (bool, error) {
	return _AdminOwned.Contract.GetMapperBanned(&_AdminOwned.CallOpts)
}

// GetOfferBanned is a free data retrieval call binding the contract method 0xc04ff849.
//
// Solidity: function getOfferBanned() view returns(bool)
func (_AdminOwned *AdminOwnedCaller) GetOfferBanned(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _AdminOwned.contract.Call(opts, out, "getOfferBanned")
	return *ret0, err
}

// GetOfferBanned is a free data retrieval call binding the contract method 0xc04ff849.
//
// Solidity: function getOfferBanned() view returns(bool)
func (_AdminOwned *AdminOwnedSession) GetOfferBanned() (bool, error) {
	return _AdminOwned.Contract.GetOfferBanned(&_AdminOwned.CallOpts)
}

// GetOfferBanned is a free data retrieval call binding the contract method 0xc04ff849.
//
// Solidity: function getOfferBanned() view returns(bool)
func (_AdminOwned *AdminOwnedCallerSession) GetOfferBanned() (bool, error) {
	return _AdminOwned.Contract.GetOfferBanned(&_AdminOwned.CallOpts)
}

// GetProviderBanned is a free data retrieval call binding the contract method 0x57b6bde6.
//
// Solidity: function getProviderBanned() view returns(bool)
func (_AdminOwned *AdminOwnedCaller) GetProviderBanned(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _AdminOwned.contract.Call(opts, out, "getProviderBanned")
	return *ret0, err
}

// GetProviderBanned is a free data retrieval call binding the contract method 0x57b6bde6.
//
// Solidity: function getProviderBanned() view returns(bool)
func (_AdminOwned *AdminOwnedSession) GetProviderBanned() (bool, error) {
	return _AdminOwned.Contract.GetProviderBanned(&_AdminOwned.CallOpts)
}

// GetProviderBanned is a free data retrieval call binding the contract method 0x57b6bde6.
//
// Solidity: function getProviderBanned() view returns(bool)
func (_AdminOwned *AdminOwnedCallerSession) GetProviderBanned() (bool, error) {
	return _AdminOwned.Contract.GetProviderBanned(&_AdminOwned.CallOpts)
}

// GetQueryBanned is a free data retrieval call binding the contract method 0x16ec989c.
//
// Solidity: function getQueryBanned() view returns(bool)
func (_AdminOwned *AdminOwnedCaller) GetQueryBanned(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _AdminOwned.contract.Call(opts, out, "getQueryBanned")
	return *ret0, err
}

// GetQueryBanned is a free data retrieval call binding the contract method 0x16ec989c.
//
// Solidity: function getQueryBanned() view returns(bool)
func (_AdminOwned *AdminOwnedSession) GetQueryBanned() (bool, error) {
	return _AdminOwned.Contract.GetQueryBanned(&_AdminOwned.CallOpts)
}

// GetQueryBanned is a free data retrieval call binding the contract method 0x16ec989c.
//
// Solidity: function getQueryBanned() view returns(bool)
func (_AdminOwned *AdminOwnedCallerSession) GetQueryBanned() (bool, error) {
	return _AdminOwned.Contract.GetQueryBanned(&_AdminOwned.CallOpts)
}

// GetRootBanned is a free data retrieval call binding the contract method 0x6778d3cb.
//
// Solidity: function getRootBanned() view returns(bool)
func (_AdminOwned *AdminOwnedCaller) GetRootBanned(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _AdminOwned.contract.Call(opts, out, "getRootBanned")
	return *ret0, err
}

// GetRootBanned is a free data retrieval call binding the contract method 0x6778d3cb.
//
// Solidity: function getRootBanned() view returns(bool)
func (_AdminOwned *AdminOwnedSession) GetRootBanned() (bool, error) {
	return _AdminOwned.Contract.GetRootBanned(&_AdminOwned.CallOpts)
}

// GetRootBanned is a free data retrieval call binding the contract method 0x6778d3cb.
//
// Solidity: function getRootBanned() view returns(bool)
func (_AdminOwned *AdminOwnedCallerSession) GetRootBanned() (bool, error) {
	return _AdminOwned.Contract.GetRootBanned(&_AdminOwned.CallOpts)
}

// GetUpkeepingBanned is a free data retrieval call binding the contract method 0x72857ce2.
//
// Solidity: function getUpkeepingBanned() view returns(bool)
func (_AdminOwned *AdminOwnedCaller) GetUpkeepingBanned(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _AdminOwned.contract.Call(opts, out, "getUpkeepingBanned")
	return *ret0, err
}

// GetUpkeepingBanned is a free data retrieval call binding the contract method 0x72857ce2.
//
// Solidity: function getUpkeepingBanned() view returns(bool)
func (_AdminOwned *AdminOwnedSession) GetUpkeepingBanned() (bool, error) {
	return _AdminOwned.Contract.GetUpkeepingBanned(&_AdminOwned.CallOpts)
}

// GetUpkeepingBanned is a free data retrieval call binding the contract method 0x72857ce2.
//
// Solidity: function getUpkeepingBanned() view returns(bool)
func (_AdminOwned *AdminOwnedCallerSession) GetUpkeepingBanned() (bool, error) {
	return _AdminOwned.Contract.GetUpkeepingBanned(&_AdminOwned.CallOpts)
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

// SetChannelBanned is a paid mutator transaction binding the contract method 0xcc893e97.
//
// Solidity: function setChannelBanned(bool param) returns()
func (_AdminOwned *AdminOwnedTransactor) SetChannelBanned(opts *bind.TransactOpts, param bool) (*types.Transaction, error) {
	return _AdminOwned.contract.Transact(opts, "setChannelBanned", param)
}

// SetChannelBanned is a paid mutator transaction binding the contract method 0xcc893e97.
//
// Solidity: function setChannelBanned(bool param) returns()
func (_AdminOwned *AdminOwnedSession) SetChannelBanned(param bool) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetChannelBanned(&_AdminOwned.TransactOpts, param)
}

// SetChannelBanned is a paid mutator transaction binding the contract method 0xcc893e97.
//
// Solidity: function setChannelBanned(bool param) returns()
func (_AdminOwned *AdminOwnedTransactorSession) SetChannelBanned(param bool) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetChannelBanned(&_AdminOwned.TransactOpts, param)
}

// SetKPMapBanned is a paid mutator transaction binding the contract method 0x17ef9444.
//
// Solidity: function setKPMapBanned(bool param) returns()
func (_AdminOwned *AdminOwnedTransactor) SetKPMapBanned(opts *bind.TransactOpts, param bool) (*types.Transaction, error) {
	return _AdminOwned.contract.Transact(opts, "setKPMapBanned", param)
}

// SetKPMapBanned is a paid mutator transaction binding the contract method 0x17ef9444.
//
// Solidity: function setKPMapBanned(bool param) returns()
func (_AdminOwned *AdminOwnedSession) SetKPMapBanned(param bool) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetKPMapBanned(&_AdminOwned.TransactOpts, param)
}

// SetKPMapBanned is a paid mutator transaction binding the contract method 0x17ef9444.
//
// Solidity: function setKPMapBanned(bool param) returns()
func (_AdminOwned *AdminOwnedTransactorSession) SetKPMapBanned(param bool) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetKPMapBanned(&_AdminOwned.TransactOpts, param)
}

// SetKeeperBanned is a paid mutator transaction binding the contract method 0x902b218d.
//
// Solidity: function setKeeperBanned(bool param) returns()
func (_AdminOwned *AdminOwnedTransactor) SetKeeperBanned(opts *bind.TransactOpts, param bool) (*types.Transaction, error) {
	return _AdminOwned.contract.Transact(opts, "setKeeperBanned", param)
}

// SetKeeperBanned is a paid mutator transaction binding the contract method 0x902b218d.
//
// Solidity: function setKeeperBanned(bool param) returns()
func (_AdminOwned *AdminOwnedSession) SetKeeperBanned(param bool) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetKeeperBanned(&_AdminOwned.TransactOpts, param)
}

// SetKeeperBanned is a paid mutator transaction binding the contract method 0x902b218d.
//
// Solidity: function setKeeperBanned(bool param) returns()
func (_AdminOwned *AdminOwnedTransactorSession) SetKeeperBanned(param bool) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetKeeperBanned(&_AdminOwned.TransactOpts, param)
}

// SetMapperBanned is a paid mutator transaction binding the contract method 0xe922e4ce.
//
// Solidity: function setMapperBanned(bool param) returns()
func (_AdminOwned *AdminOwnedTransactor) SetMapperBanned(opts *bind.TransactOpts, param bool) (*types.Transaction, error) {
	return _AdminOwned.contract.Transact(opts, "setMapperBanned", param)
}

// SetMapperBanned is a paid mutator transaction binding the contract method 0xe922e4ce.
//
// Solidity: function setMapperBanned(bool param) returns()
func (_AdminOwned *AdminOwnedSession) SetMapperBanned(param bool) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetMapperBanned(&_AdminOwned.TransactOpts, param)
}

// SetMapperBanned is a paid mutator transaction binding the contract method 0xe922e4ce.
//
// Solidity: function setMapperBanned(bool param) returns()
func (_AdminOwned *AdminOwnedTransactorSession) SetMapperBanned(param bool) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetMapperBanned(&_AdminOwned.TransactOpts, param)
}

// SetOfferBanned is a paid mutator transaction binding the contract method 0x06909ba7.
//
// Solidity: function setOfferBanned(bool param) returns()
func (_AdminOwned *AdminOwnedTransactor) SetOfferBanned(opts *bind.TransactOpts, param bool) (*types.Transaction, error) {
	return _AdminOwned.contract.Transact(opts, "setOfferBanned", param)
}

// SetOfferBanned is a paid mutator transaction binding the contract method 0x06909ba7.
//
// Solidity: function setOfferBanned(bool param) returns()
func (_AdminOwned *AdminOwnedSession) SetOfferBanned(param bool) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetOfferBanned(&_AdminOwned.TransactOpts, param)
}

// SetOfferBanned is a paid mutator transaction binding the contract method 0x06909ba7.
//
// Solidity: function setOfferBanned(bool param) returns()
func (_AdminOwned *AdminOwnedTransactorSession) SetOfferBanned(param bool) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetOfferBanned(&_AdminOwned.TransactOpts, param)
}

// SetProviderBanned is a paid mutator transaction binding the contract method 0x9db22ae6.
//
// Solidity: function setProviderBanned(bool param) returns()
func (_AdminOwned *AdminOwnedTransactor) SetProviderBanned(opts *bind.TransactOpts, param bool) (*types.Transaction, error) {
	return _AdminOwned.contract.Transact(opts, "setProviderBanned", param)
}

// SetProviderBanned is a paid mutator transaction binding the contract method 0x9db22ae6.
//
// Solidity: function setProviderBanned(bool param) returns()
func (_AdminOwned *AdminOwnedSession) SetProviderBanned(param bool) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetProviderBanned(&_AdminOwned.TransactOpts, param)
}

// SetProviderBanned is a paid mutator transaction binding the contract method 0x9db22ae6.
//
// Solidity: function setProviderBanned(bool param) returns()
func (_AdminOwned *AdminOwnedTransactorSession) SetProviderBanned(param bool) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetProviderBanned(&_AdminOwned.TransactOpts, param)
}

// SetQueryBanned is a paid mutator transaction binding the contract method 0x2321b8ae.
//
// Solidity: function setQueryBanned(bool param) returns()
func (_AdminOwned *AdminOwnedTransactor) SetQueryBanned(opts *bind.TransactOpts, param bool) (*types.Transaction, error) {
	return _AdminOwned.contract.Transact(opts, "setQueryBanned", param)
}

// SetQueryBanned is a paid mutator transaction binding the contract method 0x2321b8ae.
//
// Solidity: function setQueryBanned(bool param) returns()
func (_AdminOwned *AdminOwnedSession) SetQueryBanned(param bool) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetQueryBanned(&_AdminOwned.TransactOpts, param)
}

// SetQueryBanned is a paid mutator transaction binding the contract method 0x2321b8ae.
//
// Solidity: function setQueryBanned(bool param) returns()
func (_AdminOwned *AdminOwnedTransactorSession) SetQueryBanned(param bool) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetQueryBanned(&_AdminOwned.TransactOpts, param)
}

// SetRootBanned is a paid mutator transaction binding the contract method 0xfc3b6a39.
//
// Solidity: function setRootBanned(bool param) returns()
func (_AdminOwned *AdminOwnedTransactor) SetRootBanned(opts *bind.TransactOpts, param bool) (*types.Transaction, error) {
	return _AdminOwned.contract.Transact(opts, "setRootBanned", param)
}

// SetRootBanned is a paid mutator transaction binding the contract method 0xfc3b6a39.
//
// Solidity: function setRootBanned(bool param) returns()
func (_AdminOwned *AdminOwnedSession) SetRootBanned(param bool) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetRootBanned(&_AdminOwned.TransactOpts, param)
}

// SetRootBanned is a paid mutator transaction binding the contract method 0xfc3b6a39.
//
// Solidity: function setRootBanned(bool param) returns()
func (_AdminOwned *AdminOwnedTransactorSession) SetRootBanned(param bool) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetRootBanned(&_AdminOwned.TransactOpts, param)
}

// SetUpkeepingBanned is a paid mutator transaction binding the contract method 0x1fed1561.
//
// Solidity: function setUpkeepingBanned(bool param) returns()
func (_AdminOwned *AdminOwnedTransactor) SetUpkeepingBanned(opts *bind.TransactOpts, param bool) (*types.Transaction, error) {
	return _AdminOwned.contract.Transact(opts, "setUpkeepingBanned", param)
}

// SetUpkeepingBanned is a paid mutator transaction binding the contract method 0x1fed1561.
//
// Solidity: function setUpkeepingBanned(bool param) returns()
func (_AdminOwned *AdminOwnedSession) SetUpkeepingBanned(param bool) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetUpkeepingBanned(&_AdminOwned.TransactOpts, param)
}

// SetUpkeepingBanned is a paid mutator transaction binding the contract method 0x1fed1561.
//
// Solidity: function setUpkeepingBanned(bool param) returns()
func (_AdminOwned *AdminOwnedTransactorSession) SetUpkeepingBanned(param bool) (*types.Transaction, error) {
	return _AdminOwned.Contract.SetUpkeepingBanned(&_AdminOwned.TransactOpts, param)
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
	Key   string
	From  common.Address
	Param bool
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterSetBanned is a free log retrieval operation binding the contract event 0x88a2f5ad849982851810463cc052ff32213c45ba95d13cf29e38f09c5c0c4eca.
//
// Solidity: event SetBanned(string key, address from, bool param)
func (_AdminOwned *AdminOwnedFilterer) FilterSetBanned(opts *bind.FilterOpts) (*AdminOwnedSetBannedIterator, error) {

	logs, sub, err := _AdminOwned.contract.FilterLogs(opts, "SetBanned")
	if err != nil {
		return nil, err
	}
	return &AdminOwnedSetBannedIterator{contract: _AdminOwned.contract, event: "SetBanned", logs: logs, sub: sub}, nil
}

// WatchSetBanned is a free log subscription operation binding the contract event 0x88a2f5ad849982851810463cc052ff32213c45ba95d13cf29e38f09c5c0c4eca.
//
// Solidity: event SetBanned(string key, address from, bool param)
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

// ParseSetBanned is a log parse operation binding the contract event 0x88a2f5ad849982851810463cc052ff32213c45ba95d13cf29e38f09c5c0c4eca.
//
// Solidity: event SetBanned(string key, address from, bool param)
func (_AdminOwned *AdminOwnedFilterer) ParseSetBanned(log types.Log) (*AdminOwnedSetBanned, error) {
	event := new(AdminOwnedSetBanned)
	if err := _AdminOwned.contract.UnpackLog(event, "SetBanned", log); err != nil {
		return nil, err
	}
	return event, nil
}
