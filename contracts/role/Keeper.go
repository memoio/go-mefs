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

// KeeperABI is the input ABI used to generate the binding from.
const KeeperABI = "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_price\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"}],\"name\":\"CancelPledgeStatus\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"acc\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"sum\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"cancelPledge\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"cancelPledgeStatus\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAllAddress\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getPrice\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"info\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pledge\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"set\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"setBanned\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_price\",\"type\":\"uint256\"}],\"name\":\"setPrice\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"transferPledge\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"payable\",\"type\":\"function\"}]"

// KeeperBin is the compiled bytecode used for deploying new contracts.
var KeeperBin = "0x608060405273acf8e37d9e3dcb47423f2938069c11d75de17a20600360006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555034801561006557600080fd5b5060405161186c38038061186c8339818101604052602081101561008857600080fd5b8101908080519060200190929190505050336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550806002819055505061177c806100f06000396000f3fe6080604052600436106100a75760003560e01c806388ffe8671161006457806388ffe8671461033d578063893d20e81461035d57806391b7f5ed1461039e57806398d5fdca146103d9578063ae5e266614610404578063e3685c4014610424576100a7565b80630aae7a6b146100ac5780630ca05f9f1461012a57806335e3b25a146101915780636cb0e8e314610204578063715b208b1461025e57806388c9bcce146102ca575b600080fd5b3480156100b857600080fd5b506100fb600480360360208110156100cf57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610494565b604051808515158152602001841515815260200183815260200182815260200194505050505060405180910390f35b34801561013657600080fd5b506101796004803603602081101561014d57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610570565b60405180821515815260200191505060405180910390f35b34801561019d57600080fd5b506101ec600480360360408110156101b457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080351515906020019092919050505061070f565b60405180821515815260200191505060405180910390f35b6102466004803603602081101561021a57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610970565b60405180821515815260200191505060405180910390f35b34801561026a57600080fd5b50610273610a8a565b6040518080602001828103825283818151815260200191508051906020019060200280838360005b838110156102b657808201518184015260208101905061029b565b505050509050019250505060405180910390f35b3480156102d657600080fd5b50610325600480360360408110156102ed57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803515159060200190929190505050610cc0565b60405180821515815260200191505060405180910390f35b610345610ef0565b60405180821515815260200191505060405180910390f35b34801561036957600080fd5b506103726111b4565b604051808273ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b3480156103aa57600080fd5b506103d7600480360360208110156103c157600080fd5b81019080803590602001909291905050506111dd565b005b3480156103e557600080fd5b506103ee6112a8565b6040518082815260200191505060405180910390f35b61040c6112b2565b60405180821515815260200191505060405180910390f35b61047c6004803603606081101561043a57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919080351515906020019092919050505061148b565b60405180821515815260200191505060405180910390f35b60008060008060006104a58661169a565b90506001805490508114156104c95760006001600080945094509450945050610569565b600181815481106104d657fe5b906000526020600020906003020160000160149054906101000a900460ff166001828154811061050257fe5b906000526020600020906003020160000160159054906101000a900460ff166001838154811061052e57fe5b9060005260206000209060030201600101546001848154811061054d57fe5b9060005260206000209060030201600201549450945094509450505b9193509193565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610634576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050826000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e908184604051808373ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a16001915050919050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146107d3576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b60006107de8461169a565b9050600180549050811461085857600181815481106107f957fe5b906000526020600020906003020160000160159054906101000a900460ff1661085357826001828154811061082a57fe5b906000526020600020906003020160000160146101000a81548160ff0219169083151502179055505b610965565b60016040518060a001604052808673ffffffffffffffffffffffffffffffffffffffff1681526020018515158152602001600015158152602001600081526020016000815250908060018154018082558091505060019003906000526020600020906003020160009091909190915060008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060208201518160000160146101000a81548160ff02191690831515021790555060408201518160000160156101000a81548160ff021916908315150217905550606082015181600101556080820151816002015550505b600191505092915050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610a34576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b60004790508273ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f19350505050158015610a7f573d6000803e3d6000fd5b506001915050919050565b60608060018054905067ffffffffffffffff81118015610aa957600080fd5b50604051908082528060200260200182016040528015610ad85781602001602082028036833780820191505090505b5090506000805b600180549050811015610bf8576001151560018281548110610afd57fe5b906000526020600020906003020160000160149054906101000a900460ff1615151415610beb576000151560018281548110610b3557fe5b906000526020600020906003020160000160159054906101000a900460ff1615151415610bea5760018181548110610b6957fe5b906000526020600020906003020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16838381518110610ba757fe5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505081806001019250505b5b8080600101915050610adf565b5060608167ffffffffffffffff81118015610c1257600080fd5b50604051908082528060200260200182016040528015610c415781602001602082028036833780820191505090505b50905060005b82811015610cb657838181518110610c5b57fe5b6020026020010151828281518110610c6f57fe5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff16815250508080600101915050610c47565b5080935050505090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610d84576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b6000610d8f8461169a565b90506001805490508114610dd8578260018281548110610dab57fe5b906000526020600020906003020160000160156101000a81548160ff021916908315150217905550610ee5565b60016040518060a001604052808673ffffffffffffffffffffffffffffffffffffffff1681526020016000151581526020018515158152602001600081526020016000815250908060018154018082558091505060019003906000526020600020906003020160009091909190915060008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060208201518160000160146101000a81548160ff02191690831515021790555060408201518160000160156101000a81548160ff021916908315150217905550606082015181600101556080820151816002015550505b600191505092915050565b600080600360009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663073eeb536040518163ffffffff1660e01b815260040160206040518083038186803b158015610f5b57600080fd5b505afa158015610f6f573d6000803e3d6000fd5b505050506040513d6020811015610f8557600080fd5b81019080805190602001909291905050509050600161ffff168161ffff1610611016576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260108152602001807f706c656467652069732062616e6e65640000000000000000000000000000000081525060200191505060405180910390fd5b600254341015611071573373ffffffffffffffffffffffffffffffffffffffff166108fc349081150290604051600060405180830381858888f19350505050158015611066573d6000803e3d6000fd5b5060009150506111b1565b600061107c3361169a565b90506001805490508114156110dd573373ffffffffffffffffffffffffffffffffffffffff166108fc349081150290604051600060405180830381858888f193505050501580156110d1573d6000803e3d6000fd5b506000925050506111b1565b60018082815481106110eb57fe5b906000526020600020906003020160000160146101000a81548160ff0219169083151502179055506000346001838154811061112357fe5b9060005260206000209060030201600101540190506001828154811061114557fe5b90600052602060002090600302016001015481101561116357600080fd5b806001838154811061117157fe5b906000526020600020906003020160010181905550426001838154811061119457fe5b906000526020600020906003020160020181905550600193505050505b90565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461129e576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b8060028190555050565b6000600254905090565b600080600360009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663073eeb536040518163ffffffff1660e01b815260040160206040518083038186803b15801561131d57600080fd5b505afa158015611331573d6000803e3d6000fd5b505050506040513d602081101561134757600080fd5b81019080805190602001909291905050509050600161ffff168161ffff16106113d8576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260098152602001807f69732062616e6e6564000000000000000000000000000000000000000000000081525060200191505060405180910390fd5b60006113e33361169a565b90506001805490508114156113fd57600092505050611488565b60006001828154811061140c57fe5b906000526020600020906003020160000160146101000a81548160ff0219169083151502179055507fdc63ce37be5b519f3d4a0c7e239641bea35e801630cac3175f46f92ee32583aa33604051808273ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390a16001925050505b90565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461154f576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f6f6e6c79206f776e65722063616e2063616c6c0000000000000000000000000081525060200191505060405180910390fd5b600061155a8561169a565b9050600180549050811415611573576000915050611693565b60006001828154811061158257fe5b9060005260206000209060030201600101548511156115c157600182815481106115a857fe5b90600052602060002090600302016001015490506115c5565b8490505b60008114156115d957600092505050611693565b80600183815481106115e757fe5b906000526020600020906003020160010160008282540392505081905550831561168c57600015156001838154811061161c57fe5b906000526020600020906003020160000160159054906101000a900460ff161515141561168b578573ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f19350505050158015611689573d6000803e3d6000fd5b505b5b6001925050505b9392505050565b600080600090505b600180549050811015611737578273ffffffffffffffffffffffffffffffffffffffff16600182815481106116d357fe5b906000526020600020906003020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16141561172a5780915050611741565b80806001019150506116a2565b5060018054905090505b91905056fea264697066735822122080f9d50916846e4daef4f782851efc886c3c4699f212827644134dda7bc99da964736f6c63430007030033"

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
// Solidity: function getAllAddress() view returns(address[])
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
// Solidity: function getAllAddress() view returns(address[])
func (_Keeper *KeeperSession) GetAllAddress() ([]common.Address, error) {
	return _Keeper.Contract.GetAllAddress(&_Keeper.CallOpts)
}

// GetAllAddress is a free data retrieval call binding the contract method 0x715b208b.
//
// Solidity: function getAllAddress() view returns(address[])
func (_Keeper *KeeperCallerSession) GetAllAddress() ([]common.Address, error) {
	return _Keeper.Contract.GetAllAddress(&_Keeper.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
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
// Solidity: function getOwner() view returns(address)
func (_Keeper *KeeperSession) GetOwner() (common.Address, error) {
	return _Keeper.Contract.GetOwner(&_Keeper.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() view returns(address)
func (_Keeper *KeeperCallerSession) GetOwner() (common.Address, error) {
	return _Keeper.Contract.GetOwner(&_Keeper.CallOpts)
}

// GetPrice is a free data retrieval call binding the contract method 0x98d5fdca.
//
// Solidity: function getPrice() view returns(uint256)
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
// Solidity: function getPrice() view returns(uint256)
func (_Keeper *KeeperSession) GetPrice() (*big.Int, error) {
	return _Keeper.Contract.GetPrice(&_Keeper.CallOpts)
}

// GetPrice is a free data retrieval call binding the contract method 0x98d5fdca.
//
// Solidity: function getPrice() view returns(uint256)
func (_Keeper *KeeperCallerSession) GetPrice() (*big.Int, error) {
	return _Keeper.Contract.GetPrice(&_Keeper.CallOpts)
}

// Info is a free data retrieval call binding the contract method 0x0aae7a6b.
//
// Solidity: function info(address addr) view returns(bool, bool, uint256, uint256)
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
// Solidity: function info(address addr) view returns(bool, bool, uint256, uint256)
func (_Keeper *KeeperSession) Info(addr common.Address) (bool, bool, *big.Int, *big.Int, error) {
	return _Keeper.Contract.Info(&_Keeper.CallOpts, addr)
}

// Info is a free data retrieval call binding the contract method 0x0aae7a6b.
//
// Solidity: function info(address addr) view returns(bool, bool, uint256, uint256)
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
// Solidity: function cancelPledge(address acc, uint256 sum, bool status) payable returns(bool)
func (_Keeper *KeeperTransactor) CancelPledge(opts *bind.TransactOpts, acc common.Address, sum *big.Int, status bool) (*types.Transaction, error) {
	return _Keeper.contract.Transact(opts, "cancelPledge", acc, sum, status)
}

// CancelPledge is a paid mutator transaction binding the contract method 0xe3685c40.
//
// Solidity: function cancelPledge(address acc, uint256 sum, bool status) payable returns(bool)
func (_Keeper *KeeperSession) CancelPledge(acc common.Address, sum *big.Int, status bool) (*types.Transaction, error) {
	return _Keeper.Contract.CancelPledge(&_Keeper.TransactOpts, acc, sum, status)
}

// CancelPledge is a paid mutator transaction binding the contract method 0xe3685c40.
//
// Solidity: function cancelPledge(address acc, uint256 sum, bool status) payable returns(bool)
func (_Keeper *KeeperTransactorSession) CancelPledge(acc common.Address, sum *big.Int, status bool) (*types.Transaction, error) {
	return _Keeper.Contract.CancelPledge(&_Keeper.TransactOpts, acc, sum, status)
}

// CancelPledgeStatus is a paid mutator transaction binding the contract method 0xae5e2666.
//
// Solidity: function cancelPledgeStatus() payable returns(bool)
func (_Keeper *KeeperTransactor) CancelPledgeStatus(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Keeper.contract.Transact(opts, "cancelPledgeStatus")
}

// CancelPledgeStatus is a paid mutator transaction binding the contract method 0xae5e2666.
//
// Solidity: function cancelPledgeStatus() payable returns(bool)
func (_Keeper *KeeperSession) CancelPledgeStatus() (*types.Transaction, error) {
	return _Keeper.Contract.CancelPledgeStatus(&_Keeper.TransactOpts)
}

// CancelPledgeStatus is a paid mutator transaction binding the contract method 0xae5e2666.
//
// Solidity: function cancelPledgeStatus() payable returns(bool)
func (_Keeper *KeeperTransactorSession) CancelPledgeStatus() (*types.Transaction, error) {
	return _Keeper.Contract.CancelPledgeStatus(&_Keeper.TransactOpts)
}

// Pledge is a paid mutator transaction binding the contract method 0x88ffe867.
//
// Solidity: function pledge() payable returns(bool)
func (_Keeper *KeeperTransactor) Pledge(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Keeper.contract.Transact(opts, "pledge")
}

// Pledge is a paid mutator transaction binding the contract method 0x88ffe867.
//
// Solidity: function pledge() payable returns(bool)
func (_Keeper *KeeperSession) Pledge() (*types.Transaction, error) {
	return _Keeper.Contract.Pledge(&_Keeper.TransactOpts)
}

// Pledge is a paid mutator transaction binding the contract method 0x88ffe867.
//
// Solidity: function pledge() payable returns(bool)
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

// TransferPledge is a paid mutator transaction binding the contract method 0x6cb0e8e3.
//
// Solidity: function transferPledge(address to) payable returns(bool)
func (_Keeper *KeeperTransactor) TransferPledge(opts *bind.TransactOpts, to common.Address) (*types.Transaction, error) {
	return _Keeper.contract.Transact(opts, "transferPledge", to)
}

// TransferPledge is a paid mutator transaction binding the contract method 0x6cb0e8e3.
//
// Solidity: function transferPledge(address to) payable returns(bool)
func (_Keeper *KeeperSession) TransferPledge(to common.Address) (*types.Transaction, error) {
	return _Keeper.Contract.TransferPledge(&_Keeper.TransactOpts, to)
}

// TransferPledge is a paid mutator transaction binding the contract method 0x6cb0e8e3.
//
// Solidity: function transferPledge(address to) payable returns(bool)
func (_Keeper *KeeperTransactorSession) TransferPledge(to common.Address) (*types.Transaction, error) {
	return _Keeper.Contract.TransferPledge(&_Keeper.TransactOpts, to)
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

// KeeperCancelPledgeStatusIterator is returned from FilterCancelPledgeStatus and is used to iterate over the raw logs and unpacked data for CancelPledgeStatus events raised by the Keeper contract.
type KeeperCancelPledgeStatusIterator struct {
	Event *KeeperCancelPledgeStatus // Event containing the contract specifics and raw log

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
func (it *KeeperCancelPledgeStatusIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperCancelPledgeStatus)
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
		it.Event = new(KeeperCancelPledgeStatus)
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
func (it *KeeperCancelPledgeStatusIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperCancelPledgeStatusIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperCancelPledgeStatus represents a CancelPledgeStatus event raised by the Keeper contract.
type KeeperCancelPledgeStatus struct {
	From common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterCancelPledgeStatus is a free log retrieval operation binding the contract event 0xdc63ce37be5b519f3d4a0c7e239641bea35e801630cac3175f46f92ee32583aa.
//
// Solidity: event CancelPledgeStatus(address from)
func (_Keeper *KeeperFilterer) FilterCancelPledgeStatus(opts *bind.FilterOpts) (*KeeperCancelPledgeStatusIterator, error) {

	logs, sub, err := _Keeper.contract.FilterLogs(opts, "CancelPledgeStatus")
	if err != nil {
		return nil, err
	}
	return &KeeperCancelPledgeStatusIterator{contract: _Keeper.contract, event: "CancelPledgeStatus", logs: logs, sub: sub}, nil
}

// WatchCancelPledgeStatus is a free log subscription operation binding the contract event 0xdc63ce37be5b519f3d4a0c7e239641bea35e801630cac3175f46f92ee32583aa.
//
// Solidity: event CancelPledgeStatus(address from)
func (_Keeper *KeeperFilterer) WatchCancelPledgeStatus(opts *bind.WatchOpts, sink chan<- *KeeperCancelPledgeStatus) (event.Subscription, error) {

	logs, sub, err := _Keeper.contract.WatchLogs(opts, "CancelPledgeStatus")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperCancelPledgeStatus)
				if err := _Keeper.contract.UnpackLog(event, "CancelPledgeStatus", log); err != nil {
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
func (_Keeper *KeeperFilterer) ParseCancelPledgeStatus(log types.Log) (*KeeperCancelPledgeStatus, error) {
	event := new(KeeperCancelPledgeStatus)
	if err := _Keeper.contract.UnpackLog(event, "CancelPledgeStatus", log); err != nil {
		return nil, err
	}
	return event, nil
}
