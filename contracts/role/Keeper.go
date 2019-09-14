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
const KeeperABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\",\"signature\":\"0x0ca05f9f\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\",\"signature\":\"0x893d20e8\"},{\"inputs\":[{\"name\":\"_deposit\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\",\"signature\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"bool\"}],\"name\":\"Set\",\"type\":\"event\",\"signature\":\"0xa09d518561e304be3f7de32d470dadb560b3bc168a5bad632dba82666dda9589\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"acc\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"money\",\"type\":\"uint256\"}],\"name\":\"Pledge\",\"type\":\"event\",\"signature\":\"0x5e91ea8ea1c46300eb761859be01d7b16d44389ef91e03a163a87413cbf55b95\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"acc\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"money\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"CancelPledge\",\"type\":\"event\",\"signature\":\"0xa70461ebff4d11e6f321ed483fa2998132842461adfbbae6d10dc5f3b9b23058\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"acc\",\"type\":\"address\"}],\"name\":\"ApplyCancelPledge\",\"type\":\"event\",\"signature\":\"0x8444089032f19005a69aa5d9fa82009aee6241c0c090430f53abe2d6d890697c\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"data\",\"type\":\"string\"}],\"name\":\"Error\",\"type\":\"event\",\"signature\":\"0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\",\"signature\":\"0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"bool\"}],\"name\":\"setWhitelist\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\",\"signature\":\"0x53d6fd59\"},{\"constant\":false,\"inputs\":[{\"name\":\"_deposit\",\"type\":\"uint256\"}],\"name\":\"setDeposit\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\",\"signature\":\"0xf5bade66\"},{\"constant\":true,\"inputs\":[],\"name\":\"getDeposit\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\",\"signature\":\"0xc399ec88\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"bool\"}],\"name\":\"set\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\",\"signature\":\"0x35e3b25a\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"isKeeper\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\",\"signature\":\"0x6ba42aaa\"},{\"constant\":true,\"inputs\":[],\"name\":\"getAllAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\",\"signature\":\"0x715b208b\"},{\"constant\":false,\"inputs\":[],\"name\":\"pledge\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\",\"signature\":\"0x88ffe867\"},{\"constant\":false,\"inputs\":[],\"name\":\"cancelPledge\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\",\"signature\":\"0x28c418cf\"},{\"constant\":false,\"inputs\":[{\"name\":\"acc\",\"type\":\"address\"},{\"name\":\"sum\",\"type\":\"uint256\"},{\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"setCancelPledgeStatus\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\",\"signature\":\"0xd78ed5da\"},{\"constant\":true,\"inputs\":[{\"name\":\"acc\",\"type\":\"address\"}],\"name\":\"info\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"},{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\",\"signature\":\"0x0aae7a6b\"}]"

// KeeperBin is the compiled bytecode used for deploying new contracts.
const KeeperBin = `0x608060405234801561001057600080fd5b506040516020806117d58339810180604052602081101561003057600080fd5b8101908080519060200190929190505050336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550806002819055505061173d806100986000396000f3fe6080604052600436106100ba576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680630aae7a6b146100bf5780630ca05f9f1461012f57806328c418cf1461019857806335e3b25a146101ba57806353d6fd591461022f5780636ba42aaa146102a4578063715b208b1461030d57806388ffe86714610379578063893d20e81461039b578063c399ec88146103f2578063d78ed5da1461041d578063f5bade661461048f575b600080fd5b3480156100cb57600080fd5b5061010e600480360360208110156100e257600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506104e2565b60405180831515151581526020018281526020019250505060405180910390f35b34801561013b57600080fd5b5061017e6004803603602081101561015257600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610584565b604051808215151515815260200191505060405180910390f35b6101a0610755565b604051808215151515815260200191505060405180910390f35b3480156101c657600080fd5b50610215600480360360408110156101dd57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035151590602001909291905050506108a2565b604051808215151515815260200191505060405180910390f35b34801561023b57600080fd5b5061028a6004803603604081101561025257600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803515159060200190929190505050610b30565b604051808215151515815260200191505060405180910390f35b3480156102b057600080fd5b506102f3600480360360208110156102c757600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610c5b565b604051808215151515815260200191505060405180910390f35b34801561031957600080fd5b50610322610cd3565b6040518080602001828103825283818151815260200191508051906020019060200280838360005b8381101561036557808201518184015260208101905061034a565b505050509050019250505060405180910390f35b610381610ebc565b604051808215151515815260200191505060405180910390f35b3480156103a757600080fd5b506103b061127c565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b3480156103fe57600080fd5b506104076112a5565b6040518082815260200191505060405180910390f35b6104756004803603606081101561043357600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291908035151590602001909291905050506112af565b604051808215151515815260200191505060405180910390f35b34801561049b57600080fd5b506104c8600480360360208110156104b257600080fd5b810190808035906020019092919050505061156e565b604051808215151515815260200191505060405180910390f35b60008060006104f084611648565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff81141561052a57600080809050925092505061057f565b60018181548110151561053957fe5b906000526020600020906002020160000160149054906101000a900460ff1660018281548110151561056757fe5b90600052602060002090600202016001015492509250505b915091565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614156106e25760008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050826000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e908184604051808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a1600191505061074f565b7f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa60405180806020018281038252600e8152602001807fe4bda0e4b88de698af6f776e657200000000000000000000000000000000000081525060200191505060405180910390a1610750565b5b919050565b60008061076133611648565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8114156107fd577f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa60405180806020018281038252600f8152602001807fe682a8e4b88de698af6b6565706572000000000000000000000000000000000081525060200191505060405180910390a1600091505061089f565b600060018281548110151561080e57fe5b906000526020600020906002020160000160146101000a81548160ff0219169083151502179055507f8444089032f19005a69aa5d9fa82009aee6241c0c090430f53abe2d6d890697c33604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390a160019150505b90565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161415610abc57600061090484611648565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8114151561096c578260018281548110151561093f57fe5b906000526020600020906002020160000160146101000a81548160ff021916908315150217905550610a43565b60016060604051908101604052808673ffffffffffffffffffffffffffffffffffffffff168152602001851515815260200160008152509080600181540180825580915050906001820390600052602060002090600202016000909192909190915060008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060208201518160000160146101000a81548160ff021916908315150217905550604082015181600101555050505b7fa09d518561e304be3f7de32d470dadb560b3bc168a5bad632dba82666dda95898484604051808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001821515151581526020019250505060405180910390a16001915050610b29565b7f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa60405180806020018281038252600e8152602001807fe4bda0e4b88de698af6f776e657200000000000000000000000000000000000081525060200191505060405180910390a1610b2a565b5b92915050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161415610be75781600360008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff02191690831515021790555060019050610c54565b7f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa60405180806020018281038252600e8152602001807fe4bda0e4b88de698af6f776e657200000000000000000000000000000000000081525060200191505060405180910390a1610c55565b5b92915050565b600080610c6783611648565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff81141515610cc857600181815481101515610ca157fe5b906000526020600020906002020160000160149054906101000a900460ff16915050610cce565b60009150505b919050565b606080600180549050604051908082528060200260200182016040528015610d0a5781602001602082028038833980820191505090505b509050600080905060008090505b600180549050811015610e005760011515600182815481101515610d3857fe5b906000526020600020906002020160000160149054906101000a900460ff1615151415610df357600181815481101515610d6e57fe5b906000526020600020906002020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff168383815181101515610dae57fe5b9060200190602002019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505081806001019250505b8080600101915050610d18565b50606081604051908082528060200260200182016040528015610e325781602001602082028038833980820191505090505b50905060008090505b82811015610eb2578381815181101515610e5157fe5b906020019060200201518282815181101515610e6957fe5b9060200190602002019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff16815250508080600101915050610e3b565b5080935050505090565b6000600360003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff161515610fa8577f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa6040518080602001828103825260228152602001807fe4bda0e4b88de59ca8e799bde5908de58d95e4b8ad2ce4b88de883bde794b3e881526020017fafb700000000000000000000000000000000000000000000000000000000000081525060400191505060405180910390a160009050611279565b600254341015611090577f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa60405180806020018281038252602f8152602001807fe682a8e4baa4e79a84e992b1e4b88de5a49fe8b4a8e68abce79a84e98791e9a281526020017f9d2ce99c80e8a6816465706f736974000000000000000000000000000000000081525060400191505060405180910390a13373ffffffffffffffffffffffffffffffffffffffff166108fc349081150290604051600060405180830381858888f19350505050158015611086573d6000803e3d6000fd5b5060009050611279565b600061109b33611648565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8114151561113157600180828154811015156110d657fe5b906000526020600020906002020160000160146101000a81548160ff0219169083151502179055503460018281548110151561110e57fe5b906000526020600020906002020160010160008282540192505081905550611208565b60016060604051908101604052803373ffffffffffffffffffffffffffffffffffffffff168152602001600115158152602001348152509080600181540180825580915050906001820390600052602060002090600202016000909192909190915060008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060208201518160000160146101000a81548160ff021916908315150217905550604082015181600101555050505b7f5e91ea8ea1c46300eb761859be01d7b16d44389ef91e03a163a87413cbf55b953334604051808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018281526020019250505060405180910390a160019150505b90565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b6000600254905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614156114f957600061131133611648565b9050600080851415611328576000925050506114f4565b60018281548110151561133757fe5b9060005260206000209060020201600101548511156113785760018281548110151561135f57fe5b906000526020600020906002020160010154905061137c565b8490505b8060018381548110151561138c57fe5b9060005260206000209060020201600101600082825403925050819055508315611474578573ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f193505050501580156113f6573d6000803e3d6000fd5b507fa70461ebff4d11e6f321ed483fa2998132842461adfbbae6d10dc5f3b9b2305886826001604051808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200183815260200182151515158152602001935050505060405180910390a16114ed565b7fa70461ebff4d11e6f321ed483fa2998132842461adfbbae6d10dc5f3b9b2305886826000604051808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200183815260200182151515158152602001935050505060405180910390a15b6001925050505b611566565b7f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa60405180806020018281038252600e8152602001807fe4bda0e4b88de698af6f776e657200000000000000000000000000000000000081525060200191505060405180910390a1611567565b5b9392505050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614156115d5578160028190555060019050611642565b7f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa60405180806020018281038252600e8152602001807fe4bda0e4b88de698af6f776e657200000000000000000000000000000000000081525060200191505060405180910390a1611643565b5b919050565b600080600090505b6001805490508110156116e7578273ffffffffffffffffffffffffffffffffffffffff1660018281548110151561168357fe5b906000526020600020906002020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614156116da578091505061170c565b8080600101915050611650565b507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff90505b91905056fea165627a7a72305820700ecdfc45a6786dfab3652279f35b351b9721d20cf1b0357d284406735a1a3a0029`

// DeployKeeper deploys a new Ethereum contract, binding an instance of Keeper to it.
func DeployKeeper(auth *bind.TransactOpts, backend bind.ContractBackend, _deposit *big.Int) (common.Address, *types.Transaction, *Keeper, error) {
	parsed, err := abi.JSON(strings.NewReader(KeeperABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(KeeperBin), backend, _deposit)
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

// GetDeposit is a free data retrieval call binding the contract method 0xc399ec88.
//
// Solidity: function getDeposit() constant returns(uint256)
func (_Keeper *KeeperCaller) GetDeposit(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Keeper.contract.Call(opts, out, "getDeposit")
	return *ret0, err
}

// GetDeposit is a free data retrieval call binding the contract method 0xc399ec88.
//
// Solidity: function getDeposit() constant returns(uint256)
func (_Keeper *KeeperSession) GetDeposit() (*big.Int, error) {
	return _Keeper.Contract.GetDeposit(&_Keeper.CallOpts)
}

// GetDeposit is a free data retrieval call binding the contract method 0xc399ec88.
//
// Solidity: function getDeposit() constant returns(uint256)
func (_Keeper *KeeperCallerSession) GetDeposit() (*big.Int, error) {
	return _Keeper.Contract.GetDeposit(&_Keeper.CallOpts)
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

// Info is a free data retrieval call binding the contract method 0x0aae7a6b.
//
// Solidity: function info(address acc) constant returns(bool, uint256)
func (_Keeper *KeeperCaller) Info(opts *bind.CallOpts, acc common.Address) (bool, *big.Int, error) {
	var (
		ret0 = new(bool)
		ret1 = new(*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
	}
	err := _Keeper.contract.Call(opts, out, "info", acc)
	return *ret0, *ret1, err
}

// Info is a free data retrieval call binding the contract method 0x0aae7a6b.
//
// Solidity: function info(address acc) constant returns(bool, uint256)
func (_Keeper *KeeperSession) Info(acc common.Address) (bool, *big.Int, error) {
	return _Keeper.Contract.Info(&_Keeper.CallOpts, acc)
}

// Info is a free data retrieval call binding the contract method 0x0aae7a6b.
//
// Solidity: function info(address acc) constant returns(bool, uint256)
func (_Keeper *KeeperCallerSession) Info(acc common.Address) (bool, *big.Int, error) {
	return _Keeper.Contract.Info(&_Keeper.CallOpts, acc)
}

// IsKeeper is a free data retrieval call binding the contract method 0x6ba42aaa.
//
// Solidity: function isKeeper(address addr) constant returns(bool)
func (_Keeper *KeeperCaller) IsKeeper(opts *bind.CallOpts, addr common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Keeper.contract.Call(opts, out, "isKeeper", addr)
	return *ret0, err
}

// IsKeeper is a free data retrieval call binding the contract method 0x6ba42aaa.
//
// Solidity: function isKeeper(address addr) constant returns(bool)
func (_Keeper *KeeperSession) IsKeeper(addr common.Address) (bool, error) {
	return _Keeper.Contract.IsKeeper(&_Keeper.CallOpts, addr)
}

// IsKeeper is a free data retrieval call binding the contract method 0x6ba42aaa.
//
// Solidity: function isKeeper(address addr) constant returns(bool)
func (_Keeper *KeeperCallerSession) IsKeeper(addr common.Address) (bool, error) {
	return _Keeper.Contract.IsKeeper(&_Keeper.CallOpts, addr)
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

// CancelPledge is a paid mutator transaction binding the contract method 0x28c418cf.
//
// Solidity: function cancelPledge() returns(bool)
func (_Keeper *KeeperTransactor) CancelPledge(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Keeper.contract.Transact(opts, "cancelPledge")
}

// CancelPledge is a paid mutator transaction binding the contract method 0x28c418cf.
//
// Solidity: function cancelPledge() returns(bool)
func (_Keeper *KeeperSession) CancelPledge() (*types.Transaction, error) {
	return _Keeper.Contract.CancelPledge(&_Keeper.TransactOpts)
}

// CancelPledge is a paid mutator transaction binding the contract method 0x28c418cf.
//
// Solidity: function cancelPledge() returns(bool)
func (_Keeper *KeeperTransactorSession) CancelPledge() (*types.Transaction, error) {
	return _Keeper.Contract.CancelPledge(&_Keeper.TransactOpts)
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
// Solidity: function set(address addr, bool value) returns(bool)
func (_Keeper *KeeperTransactor) Set(opts *bind.TransactOpts, addr common.Address, value bool) (*types.Transaction, error) {
	return _Keeper.contract.Transact(opts, "set", addr, value)
}

// Set is a paid mutator transaction binding the contract method 0x35e3b25a.
//
// Solidity: function set(address addr, bool value) returns(bool)
func (_Keeper *KeeperSession) Set(addr common.Address, value bool) (*types.Transaction, error) {
	return _Keeper.Contract.Set(&_Keeper.TransactOpts, addr, value)
}

// Set is a paid mutator transaction binding the contract method 0x35e3b25a.
//
// Solidity: function set(address addr, bool value) returns(bool)
func (_Keeper *KeeperTransactorSession) Set(addr common.Address, value bool) (*types.Transaction, error) {
	return _Keeper.Contract.Set(&_Keeper.TransactOpts, addr, value)
}

// SetCancelPledgeStatus is a paid mutator transaction binding the contract method 0xd78ed5da.
//
// Solidity: function setCancelPledgeStatus(address acc, uint256 sum, bool status) returns(bool)
func (_Keeper *KeeperTransactor) SetCancelPledgeStatus(opts *bind.TransactOpts, acc common.Address, sum *big.Int, status bool) (*types.Transaction, error) {
	return _Keeper.contract.Transact(opts, "setCancelPledgeStatus", acc, sum, status)
}

// SetCancelPledgeStatus is a paid mutator transaction binding the contract method 0xd78ed5da.
//
// Solidity: function setCancelPledgeStatus(address acc, uint256 sum, bool status) returns(bool)
func (_Keeper *KeeperSession) SetCancelPledgeStatus(acc common.Address, sum *big.Int, status bool) (*types.Transaction, error) {
	return _Keeper.Contract.SetCancelPledgeStatus(&_Keeper.TransactOpts, acc, sum, status)
}

// SetCancelPledgeStatus is a paid mutator transaction binding the contract method 0xd78ed5da.
//
// Solidity: function setCancelPledgeStatus(address acc, uint256 sum, bool status) returns(bool)
func (_Keeper *KeeperTransactorSession) SetCancelPledgeStatus(acc common.Address, sum *big.Int, status bool) (*types.Transaction, error) {
	return _Keeper.Contract.SetCancelPledgeStatus(&_Keeper.TransactOpts, acc, sum, status)
}

// SetDeposit is a paid mutator transaction binding the contract method 0xf5bade66.
//
// Solidity: function setDeposit(uint256 _deposit) returns(bool)
func (_Keeper *KeeperTransactor) SetDeposit(opts *bind.TransactOpts, _deposit *big.Int) (*types.Transaction, error) {
	return _Keeper.contract.Transact(opts, "setDeposit", _deposit)
}

// SetDeposit is a paid mutator transaction binding the contract method 0xf5bade66.
//
// Solidity: function setDeposit(uint256 _deposit) returns(bool)
func (_Keeper *KeeperSession) SetDeposit(_deposit *big.Int) (*types.Transaction, error) {
	return _Keeper.Contract.SetDeposit(&_Keeper.TransactOpts, _deposit)
}

// SetDeposit is a paid mutator transaction binding the contract method 0xf5bade66.
//
// Solidity: function setDeposit(uint256 _deposit) returns(bool)
func (_Keeper *KeeperTransactorSession) SetDeposit(_deposit *big.Int) (*types.Transaction, error) {
	return _Keeper.Contract.SetDeposit(&_Keeper.TransactOpts, _deposit)
}

// SetWhitelist is a paid mutator transaction binding the contract method 0x53d6fd59.
//
// Solidity: function setWhitelist(address addr, bool value) returns(bool)
func (_Keeper *KeeperTransactor) SetWhitelist(opts *bind.TransactOpts, addr common.Address, value bool) (*types.Transaction, error) {
	return _Keeper.contract.Transact(opts, "setWhitelist", addr, value)
}

// SetWhitelist is a paid mutator transaction binding the contract method 0x53d6fd59.
//
// Solidity: function setWhitelist(address addr, bool value) returns(bool)
func (_Keeper *KeeperSession) SetWhitelist(addr common.Address, value bool) (*types.Transaction, error) {
	return _Keeper.Contract.SetWhitelist(&_Keeper.TransactOpts, addr, value)
}

// SetWhitelist is a paid mutator transaction binding the contract method 0x53d6fd59.
//
// Solidity: function setWhitelist(address addr, bool value) returns(bool)
func (_Keeper *KeeperTransactorSession) SetWhitelist(addr common.Address, value bool) (*types.Transaction, error) {
	return _Keeper.Contract.SetWhitelist(&_Keeper.TransactOpts, addr, value)
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

// KeeperApplyCancelPledgeIterator is returned from FilterApplyCancelPledge and is used to iterate over the raw logs and unpacked data for ApplyCancelPledge events raised by the Keeper contract.
type KeeperApplyCancelPledgeIterator struct {
	Event *KeeperApplyCancelPledge // Event containing the contract specifics and raw log

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
func (it *KeeperApplyCancelPledgeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperApplyCancelPledge)
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
		it.Event = new(KeeperApplyCancelPledge)
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
func (it *KeeperApplyCancelPledgeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperApplyCancelPledgeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperApplyCancelPledge represents a ApplyCancelPledge event raised by the Keeper contract.
type KeeperApplyCancelPledge struct {
	Acc common.Address
	Raw types.Log // Blockchain specific contextual infos
}

// FilterApplyCancelPledge is a free log retrieval operation binding the contract event 0x8444089032f19005a69aa5d9fa82009aee6241c0c090430f53abe2d6d890697c.
//
// Solidity: event ApplyCancelPledge(address acc)
func (_Keeper *KeeperFilterer) FilterApplyCancelPledge(opts *bind.FilterOpts) (*KeeperApplyCancelPledgeIterator, error) {

	logs, sub, err := _Keeper.contract.FilterLogs(opts, "ApplyCancelPledge")
	if err != nil {
		return nil, err
	}
	return &KeeperApplyCancelPledgeIterator{contract: _Keeper.contract, event: "ApplyCancelPledge", logs: logs, sub: sub}, nil
}

// WatchApplyCancelPledge is a free log subscription operation binding the contract event 0x8444089032f19005a69aa5d9fa82009aee6241c0c090430f53abe2d6d890697c.
//
// Solidity: event ApplyCancelPledge(address acc)
func (_Keeper *KeeperFilterer) WatchApplyCancelPledge(opts *bind.WatchOpts, sink chan<- *KeeperApplyCancelPledge) (event.Subscription, error) {

	logs, sub, err := _Keeper.contract.WatchLogs(opts, "ApplyCancelPledge")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperApplyCancelPledge)
				if err := _Keeper.contract.UnpackLog(event, "ApplyCancelPledge", log); err != nil {
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

// KeeperCancelPledgeIterator is returned from FilterCancelPledge and is used to iterate over the raw logs and unpacked data for CancelPledge events raised by the Keeper contract.
type KeeperCancelPledgeIterator struct {
	Event *KeeperCancelPledge // Event containing the contract specifics and raw log

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
func (it *KeeperCancelPledgeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperCancelPledge)
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
		it.Event = new(KeeperCancelPledge)
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
func (it *KeeperCancelPledgeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperCancelPledgeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperCancelPledge represents a CancelPledge event raised by the Keeper contract.
type KeeperCancelPledge struct {
	Acc    common.Address
	Money  *big.Int
	Status bool
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterCancelPledge is a free log retrieval operation binding the contract event 0xa70461ebff4d11e6f321ed483fa2998132842461adfbbae6d10dc5f3b9b23058.
//
// Solidity: event CancelPledge(address acc, uint256 money, bool status)
func (_Keeper *KeeperFilterer) FilterCancelPledge(opts *bind.FilterOpts) (*KeeperCancelPledgeIterator, error) {

	logs, sub, err := _Keeper.contract.FilterLogs(opts, "CancelPledge")
	if err != nil {
		return nil, err
	}
	return &KeeperCancelPledgeIterator{contract: _Keeper.contract, event: "CancelPledge", logs: logs, sub: sub}, nil
}

// WatchCancelPledge is a free log subscription operation binding the contract event 0xa70461ebff4d11e6f321ed483fa2998132842461adfbbae6d10dc5f3b9b23058.
//
// Solidity: event CancelPledge(address acc, uint256 money, bool status)
func (_Keeper *KeeperFilterer) WatchCancelPledge(opts *bind.WatchOpts, sink chan<- *KeeperCancelPledge) (event.Subscription, error) {

	logs, sub, err := _Keeper.contract.WatchLogs(opts, "CancelPledge")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperCancelPledge)
				if err := _Keeper.contract.UnpackLog(event, "CancelPledge", log); err != nil {
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

// KeeperErrorIterator is returned from FilterError and is used to iterate over the raw logs and unpacked data for Error events raised by the Keeper contract.
type KeeperErrorIterator struct {
	Event *KeeperError // Event containing the contract specifics and raw log

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
func (it *KeeperErrorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperError)
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
		it.Event = new(KeeperError)
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
func (it *KeeperErrorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperErrorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperError represents a Error event raised by the Keeper contract.
type KeeperError struct {
	Data string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterError is a free log retrieval operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Keeper *KeeperFilterer) FilterError(opts *bind.FilterOpts) (*KeeperErrorIterator, error) {

	logs, sub, err := _Keeper.contract.FilterLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return &KeeperErrorIterator{contract: _Keeper.contract, event: "Error", logs: logs, sub: sub}, nil
}

// WatchError is a free log subscription operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Keeper *KeeperFilterer) WatchError(opts *bind.WatchOpts, sink chan<- *KeeperError) (event.Subscription, error) {

	logs, sub, err := _Keeper.contract.WatchLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperError)
				if err := _Keeper.contract.UnpackLog(event, "Error", log); err != nil {
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

// KeeperPledgeIterator is returned from FilterPledge and is used to iterate over the raw logs and unpacked data for Pledge events raised by the Keeper contract.
type KeeperPledgeIterator struct {
	Event *KeeperPledge // Event containing the contract specifics and raw log

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
func (it *KeeperPledgeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperPledge)
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
		it.Event = new(KeeperPledge)
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
func (it *KeeperPledgeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperPledgeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperPledge represents a Pledge event raised by the Keeper contract.
type KeeperPledge struct {
	Acc   common.Address
	Money *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterPledge is a free log retrieval operation binding the contract event 0x5e91ea8ea1c46300eb761859be01d7b16d44389ef91e03a163a87413cbf55b95.
//
// Solidity: event Pledge(address acc, uint256 money)
func (_Keeper *KeeperFilterer) FilterPledge(opts *bind.FilterOpts) (*KeeperPledgeIterator, error) {

	logs, sub, err := _Keeper.contract.FilterLogs(opts, "Pledge")
	if err != nil {
		return nil, err
	}
	return &KeeperPledgeIterator{contract: _Keeper.contract, event: "Pledge", logs: logs, sub: sub}, nil
}

// WatchPledge is a free log subscription operation binding the contract event 0x5e91ea8ea1c46300eb761859be01d7b16d44389ef91e03a163a87413cbf55b95.
//
// Solidity: event Pledge(address acc, uint256 money)
func (_Keeper *KeeperFilterer) WatchPledge(opts *bind.WatchOpts, sink chan<- *KeeperPledge) (event.Subscription, error) {

	logs, sub, err := _Keeper.contract.WatchLogs(opts, "Pledge")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperPledge)
				if err := _Keeper.contract.UnpackLog(event, "Pledge", log); err != nil {
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

// KeeperSetIterator is returned from FilterSet and is used to iterate over the raw logs and unpacked data for Set events raised by the Keeper contract.
type KeeperSetIterator struct {
	Event *KeeperSet // Event containing the contract specifics and raw log

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
func (it *KeeperSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(KeeperSet)
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
		it.Event = new(KeeperSet)
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
func (it *KeeperSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *KeeperSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// KeeperSet represents a Set event raised by the Keeper contract.
type KeeperSet struct {
	Addr  common.Address
	Value bool
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterSet is a free log retrieval operation binding the contract event 0xa09d518561e304be3f7de32d470dadb560b3bc168a5bad632dba82666dda9589.
//
// Solidity: event Set(address addr, bool value)
func (_Keeper *KeeperFilterer) FilterSet(opts *bind.FilterOpts) (*KeeperSetIterator, error) {

	logs, sub, err := _Keeper.contract.FilterLogs(opts, "Set")
	if err != nil {
		return nil, err
	}
	return &KeeperSetIterator{contract: _Keeper.contract, event: "Set", logs: logs, sub: sub}, nil
}

// WatchSet is a free log subscription operation binding the contract event 0xa09d518561e304be3f7de32d470dadb560b3bc168a5bad632dba82666dda9589.
//
// Solidity: event Set(address addr, bool value)
func (_Keeper *KeeperFilterer) WatchSet(opts *bind.WatchOpts, sink chan<- *KeeperSet) (event.Subscription, error) {

	logs, sub, err := _Keeper.contract.WatchLogs(opts, "Set")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(KeeperSet)
				if err := _Keeper.contract.UnpackLog(event, "Set", log); err != nil {
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
