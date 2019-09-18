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
const ProviderABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\",\"signature\":\"0x0ca05f9f\"},{\"constant\":true,\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\",\"signature\":\"0x893d20e8\"},{\"inputs\":[{\"name\":\"_size\",\"type\":\"uint256\"},{\"name\":\"_deposit\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\",\"signature\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"addr\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"bool\"}],\"name\":\"Set\",\"type\":\"event\",\"signature\":\"0xa09d518561e304be3f7de32d470dadb560b3bc168a5bad632dba82666dda9589\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"acc\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"money\",\"type\":\"uint256\"}],\"name\":\"Pledge\",\"type\":\"event\",\"signature\":\"0x5e91ea8ea1c46300eb761859be01d7b16d44389ef91e03a163a87413cbf55b95\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"acc\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"money\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"CancelPledge\",\"type\":\"event\",\"signature\":\"0xa70461ebff4d11e6f321ed483fa2998132842461adfbbae6d10dc5f3b9b23058\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"acc\",\"type\":\"address\"}],\"name\":\"ApplyCancelPledge\",\"type\":\"event\",\"signature\":\"0x8444089032f19005a69aa5d9fa82009aee6241c0c090430f53abe2d6d890697c\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"data\",\"type\":\"string\"}],\"name\":\"Error\",\"type\":\"event\",\"signature\":\"0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\",\"signature\":\"0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90\"},{\"constant\":false,\"inputs\":[{\"name\":\"_size\",\"type\":\"uint256\"}],\"name\":\"setSize\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\",\"signature\":\"0x170ab405\"},{\"constant\":true,\"inputs\":[],\"name\":\"getSize\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\",\"signature\":\"0xde8fa431\"},{\"constant\":false,\"inputs\":[{\"name\":\"_deposit\",\"type\":\"uint256\"}],\"name\":\"setDeposit\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\",\"signature\":\"0xf5bade66\"},{\"constant\":true,\"inputs\":[],\"name\":\"getDeposit\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\",\"signature\":\"0xc399ec88\"},{\"constant\":false,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"bool\"}],\"name\":\"set\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\",\"signature\":\"0x35e3b25a\"},{\"constant\":true,\"inputs\":[{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"isProvider\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\",\"signature\":\"0x6b074a07\"},{\"constant\":true,\"inputs\":[],\"name\":\"getAllAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\",\"signature\":\"0x715b208b\"},{\"constant\":true,\"inputs\":[{\"name\":\"_size\",\"type\":\"uint256\"}],\"name\":\"getPledgeMoney\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\",\"signature\":\"0xe1493261\"},{\"constant\":false,\"inputs\":[{\"name\":\"_size\",\"type\":\"uint256\"}],\"name\":\"pledge\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\",\"signature\":\"0x7326c9c0\"},{\"constant\":false,\"inputs\":[],\"name\":\"cancelPledge\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\",\"signature\":\"0x28c418cf\"},{\"constant\":false,\"inputs\":[{\"name\":\"acc\",\"type\":\"address\"},{\"name\":\"sum\",\"type\":\"uint256\"},{\"name\":\"status\",\"type\":\"bool\"}],\"name\":\"setCancelPledgeStatus\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\",\"signature\":\"0xd78ed5da\"},{\"constant\":true,\"inputs\":[{\"name\":\"acc\",\"type\":\"address\"}],\"name\":\"info\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"},{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\",\"signature\":\"0x0aae7a6b\"},{\"constant\":true,\"inputs\":[{\"name\":\"acc\",\"type\":\"address\"}],\"name\":\"get\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\",\"signature\":\"0xc2bc2efc\"}]"

// ProviderBin is the compiled bytecode used for deploying new contracts.
const ProviderBin = `0x608060405234801561001057600080fd5b50604051604080611b718339810180604052604081101561003057600080fd5b810190808051906020019092919080519060200190929190505050336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555081600281905550806003819055505050611ac7806100aa6000396000f3fe6080604052600436106100db576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680630aae7a6b146100e05780630ca05f9f1461015e578063170ab405146101c757806328c418cf1461021a57806335e3b25a1461023c5780636b074a07146102b1578063715b208b1461031a5780637326c9c014610386578063893d20e8146103cc578063c2bc2efc14610423578063c399ec881461048c578063d78ed5da146104b7578063de8fa43114610529578063e149326114610554578063f5bade66146105a3575b600080fd5b3480156100ec57600080fd5b5061012f6004803603602081101561010357600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506105f6565b604051808515151515815260200184815260200183815260200182815260200194505050505060405180910390f35b34801561016a57600080fd5b506101ad6004803603602081101561018157600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506106f0565b604051808215151515815260200191505060405180910390f35b3480156101d357600080fd5b50610200600480360360208110156101ea57600080fd5b81019080803590602001909291905050506108c1565b604051808215151515815260200191505060405180910390f35b61022261099b565b604051808215151515815260200191505060405180910390f35b34801561024857600080fd5b506102976004803603604081101561025f57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803515159060200190929190505050610b68565b604051808215151515815260200191505060405180910390f35b3480156102bd57600080fd5b50610300600480360360208110156102d457600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610e18565b604051808215151515815260200191505060405180910390f35b34801561032657600080fd5b5061032f610e90565b6040518080602001828103825283818151815260200191508051906020019060200280838360005b83811015610372578082015181840152602081019050610357565b505050509050019250505060405180910390f35b6103b26004803603602081101561039c57600080fd5b8101908080359060200190929190505050611079565b604051808215151515815260200191505060405180910390f35b3480156103d857600080fd5b506103e1611377565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34801561042f57600080fd5b506104726004803603602081101561044657600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506113a0565b604051808215151515815260200191505060405180910390f35b34801561049857600080fd5b506104a1611417565b6040518082815260200191505060405180910390f35b61050f600480360360608110156104cd57600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190803515159060200190929190505050611421565b604051808215151515815260200191505060405180910390f35b34801561053557600080fd5b5061053e6116e0565b6040518082815260200191505060405180910390f35b34801561056057600080fd5b5061058d6004803603602081101561057757600080fd5b81019080803590602001909291905050506116ea565b6040518082815260200191505060405180910390f35b3480156105af57600080fd5b506105dc600480360360208110156105c657600080fd5b8101908080359060200190929190505050611718565b604051808215151515815260200191505060405180910390f35b6000806000806000610607866117f2565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff81141561064e576000806000808292508191508090509450945094509450506106e9565b60018181548110151561065d57fe5b906000526020600020906004020160000160149054906101000a900460ff1660018281548110151561068b57fe5b9060005260206000209060040201600101546001838154811015156106ac57fe5b9060005260206000209060040201600201546001848154811015156106cd57fe5b9060005260206000209060040201600301549450945094509450505b9193509193565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16141561084e5760008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050826000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055507f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e908184604051808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019250505060405180910390a160019150506108bb565b7f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa60405180806020018281038252600e8152602001807fe4bda0e4b88de698af6f776e657200000000000000000000000000000000000081525060200191505060405180910390a16108bc565b5b919050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161415610928578160028190555060019050610995565b7f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa60405180806020018281038252600e8152602001807fe4bda0e4b88de698af6f776e657200000000000000000000000000000000000081525060200191505060405180910390a1610996565b5b919050565b6000806109a7336117f2565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff811415610a43577f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa6040518080602001828103825260118152602001807fe682a8e4b88de698af70726f766964657200000000000000000000000000000081525060200191505060405180910390a16000915050610b65565b610a4c336118bb565b15610ac3577f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa6040518080602001828103825260208152602001807fe682a8e698af6d696e65722ce5bf85e9a1bbe58588e98080e587ba6d696e657281525060200191505060405180910390a16000915050610b65565b6000600182815481101515610ad457fe5b906000526020600020906004020160000160146101000a81548160ff0219169083151502179055507f8444089032f19005a69aa5d9fa82009aee6241c0c090430f53abe2d6d890697c33604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390a160019150505b90565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161415610da4576000610bca846117f2565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff81141515610c325782600182815481101515610c0557fe5b906000526020600020906004020160000160146101000a81548160ff021916908315150217905550610d2b565b600160a0604051908101604052808673ffffffffffffffffffffffffffffffffffffffff1681526020018515158152602001600081526020016000815260200160008152509080600181540180825580915050906001820390600052602060002090600402016000909192909190915060008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060208201518160000160146101000a81548160ff0219169083151502179055506040820151816001015560608201518160020155608082015181600301555050505b7fa09d518561e304be3f7de32d470dadb560b3bc168a5bad632dba82666dda95898484604051808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001821515151581526020019250505060405180910390a16001915050610e11565b7f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa60405180806020018281038252600e8152602001807fe4bda0e4b88de698af6f776e657200000000000000000000000000000000000081525060200191505060405180910390a1610e12565b5b92915050565b600080610e24836117f2565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff81141515610e8557600181815481101515610e5e57fe5b906000526020600020906004020160000160149054906101000a900460ff16915050610e8b565b60009150505b919050565b606080600180549050604051908082528060200260200182016040528015610ec75781602001602082028038833980820191505090505b509050600080905060008090505b600180549050811015610fbd5760011515600182815481101515610ef557fe5b906000526020600020906004020160000160149054906101000a900460ff1615151415610fb057600181815481101515610f2b57fe5b906000526020600020906004020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff168383815181101515610f6b57fe5b9060200190602002019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505081806001019250505b8080600101915050610ed5565b50606081604051908082528060200260200182016040528015610fef5781602001602082028038833980820191505090505b50905060008090505b8281101561106f57838181518110151561100e57fe5b90602001906020020151828281518110151561102657fe5b9060200190602002019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff16815250508080600101915050610ff8565b5080935050505090565b6000611084826116ea565b341015611169577f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa60405180806020018281038252602f8152602001807fe682a8e4baa4e79a84e992b1e4b88de5a49fe8b4a8e68abce79a84e98791e9a281526020017f9d2ce99c80e8a6816465706f736974000000000000000000000000000000000081525060400191505060405180910390a13373ffffffffffffffffffffffffffffffffffffffff166108fc349081150290604051600060405180830381858888f1935050505015801561115f573d6000803e3d6000fd5b5060009050611372565b6000611174336117f2565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8114151561120a57600180828154811015156111af57fe5b906000526020600020906004020160000160146101000a81548160ff021916908315150217905550346001828154811015156111e757fe5b906000526020600020906004020160010160008282540192505081905550611301565b600160a0604051908101604052803373ffffffffffffffffffffffffffffffffffffffff168152602001600115158152602001348152602001858152602001428152509080600181540180825580915050906001820390600052602060002090600402016000909192909190915060008201518160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060208201518160000160146101000a81548160ff0219169083151502179055506040820151816001015560608201518160020155608082015181600301555050505b7f5e91ea8ea1c46300eb761859be01d7b16d44389ef91e03a163a87413cbf55b953334604051808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018281526020019250505060405180910390a160019150505b919050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b6000806113ac836117f2565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8114156113e0576000915050611412565b6001818154811015156113ef57fe5b906000526020600020906004020160000160149054906101000a900460ff169150505b919050565b6000600354905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16141561166b576000611483336117f2565b905060008085141561149a57600092505050611666565b6001828154811015156114a957fe5b9060005260206000209060040201600101548511156114ea576001828154811015156114d157fe5b90600052602060002090600402016001015490506114ee565b8490505b806001838154811015156114fe57fe5b90600052602060002090600402016001016000828254039250508190555083156115e6578573ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f19350505050158015611568573d6000803e3d6000fd5b507fa70461ebff4d11e6f321ed483fa2998132842461adfbbae6d10dc5f3b9b2305886826001604051808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200183815260200182151515158152602001935050505060405180910390a161165f565b7fa70461ebff4d11e6f321ed483fa2998132842461adfbbae6d10dc5f3b9b2305886826000604051808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200183815260200182151515158152602001935050505060405180910390a15b6001925050505b6116d8565b7f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa60405180806020018281038252600e8152602001807fe4bda0e4b88de698af6f776e657200000000000000000000000000000000000081525060200191505060405180910390a16116d9565b5b9392505050565b6000600254905090565b600080600254838115156116fa57fe5b049050600081141561170b57600190505b8060035402915050919050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16141561177f5781600381905550600190506117ec565b7f08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa60405180806020018281038252600e8152602001807fe4bda0e4b88de698af6f776e657200000000000000000000000000000000000081525060200191505060405180910390a16117ed565b5b919050565b600080600090505b600180549050811015611891578273ffffffffffffffffffffffffffffffffffffffff1660018281548110151561182d57fe5b906000526020600020906004020160000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16141561188457809150506118b6565b80806001019150506117fa565b507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff90505b919050565b600080739e4af0964ef92095ca3d2ae0c05b472837d8bd37905060008173ffffffffffffffffffffffffffffffffffffffff1663e4ae7d776040518163ffffffff167c01000000000000000000000000000000000000000000000000000000000281526004018080602001828103825260058152602001807f6d696e657200000000000000000000000000000000000000000000000000000081525060200191505060206040518083038186803b15801561197557600080fd5b505afa158015611989573d6000803e3d6000fd5b505050506040513d602081101561199f57600080fd5b81019080805190602001909291905050509050600081905060008173ffffffffffffffffffffffffffffffffffffffff1663c2bc2efc876040518263ffffffff167c0100000000000000000000000000000000000000000000000000000000028152600401808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060206040518083038186803b158015611a5257600080fd5b505afa158015611a66573d6000803e3d6000fd5b505050506040513d6020811015611a7c57600080fd5b810190808051906020019092919050505090508094505050505091905056fea165627a7a72305820a3482e7743f7552d6af78dd8f7e12500f816395a3bd16ccac97ec53be3b6c9a20029`

// DeployProvider deploys a new Ethereum contract, binding an instance of Provider to it.
func DeployProvider(auth *bind.TransactOpts, backend bind.ContractBackend, _size *big.Int, _deposit *big.Int) (common.Address, *types.Transaction, *Provider, error) {
	parsed, err := abi.JSON(strings.NewReader(ProviderABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ProviderBin), backend, _size, _deposit)
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

// Get is a free data retrieval call binding the contract method 0xc2bc2efc.
//
// Solidity: function get(address acc) constant returns(bool)
func (_Provider *ProviderCaller) Get(opts *bind.CallOpts, acc common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Provider.contract.Call(opts, out, "get", acc)
	return *ret0, err
}

// Get is a free data retrieval call binding the contract method 0xc2bc2efc.
//
// Solidity: function get(address acc) constant returns(bool)
func (_Provider *ProviderSession) Get(acc common.Address) (bool, error) {
	return _Provider.Contract.Get(&_Provider.CallOpts, acc)
}

// Get is a free data retrieval call binding the contract method 0xc2bc2efc.
//
// Solidity: function get(address acc) constant returns(bool)
func (_Provider *ProviderCallerSession) Get(acc common.Address) (bool, error) {
	return _Provider.Contract.Get(&_Provider.CallOpts, acc)
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

// GetDeposit is a free data retrieval call binding the contract method 0xc399ec88.
//
// Solidity: function getDeposit() constant returns(uint256)
func (_Provider *ProviderCaller) GetDeposit(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Provider.contract.Call(opts, out, "getDeposit")
	return *ret0, err
}

// GetDeposit is a free data retrieval call binding the contract method 0xc399ec88.
//
// Solidity: function getDeposit() constant returns(uint256)
func (_Provider *ProviderSession) GetDeposit() (*big.Int, error) {
	return _Provider.Contract.GetDeposit(&_Provider.CallOpts)
}

// GetDeposit is a free data retrieval call binding the contract method 0xc399ec88.
//
// Solidity: function getDeposit() constant returns(uint256)
func (_Provider *ProviderCallerSession) GetDeposit() (*big.Int, error) {
	return _Provider.Contract.GetDeposit(&_Provider.CallOpts)
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

// GetPledgeMoney is a free data retrieval call binding the contract method 0xe1493261.
//
// Solidity: function getPledgeMoney(uint256 _size) constant returns(uint256)
func (_Provider *ProviderCaller) GetPledgeMoney(opts *bind.CallOpts, _size *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Provider.contract.Call(opts, out, "getPledgeMoney", _size)
	return *ret0, err
}

// GetPledgeMoney is a free data retrieval call binding the contract method 0xe1493261.
//
// Solidity: function getPledgeMoney(uint256 _size) constant returns(uint256)
func (_Provider *ProviderSession) GetPledgeMoney(_size *big.Int) (*big.Int, error) {
	return _Provider.Contract.GetPledgeMoney(&_Provider.CallOpts, _size)
}

// GetPledgeMoney is a free data retrieval call binding the contract method 0xe1493261.
//
// Solidity: function getPledgeMoney(uint256 _size) constant returns(uint256)
func (_Provider *ProviderCallerSession) GetPledgeMoney(_size *big.Int) (*big.Int, error) {
	return _Provider.Contract.GetPledgeMoney(&_Provider.CallOpts, _size)
}

// GetSize is a free data retrieval call binding the contract method 0xde8fa431.
//
// Solidity: function getSize() constant returns(uint256)
func (_Provider *ProviderCaller) GetSize(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Provider.contract.Call(opts, out, "getSize")
	return *ret0, err
}

// GetSize is a free data retrieval call binding the contract method 0xde8fa431.
//
// Solidity: function getSize() constant returns(uint256)
func (_Provider *ProviderSession) GetSize() (*big.Int, error) {
	return _Provider.Contract.GetSize(&_Provider.CallOpts)
}

// GetSize is a free data retrieval call binding the contract method 0xde8fa431.
//
// Solidity: function getSize() constant returns(uint256)
func (_Provider *ProviderCallerSession) GetSize() (*big.Int, error) {
	return _Provider.Contract.GetSize(&_Provider.CallOpts)
}

// Info is a free data retrieval call binding the contract method 0x0aae7a6b.
//
// Solidity: function info(address acc) constant returns(bool, uint256, uint256, uint256)
func (_Provider *ProviderCaller) Info(opts *bind.CallOpts, acc common.Address) (bool, *big.Int, *big.Int, *big.Int, error) {
	var (
		ret0 = new(bool)
		ret1 = new(*big.Int)
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
// Solidity: function info(address acc) constant returns(bool, uint256, uint256, uint256)
func (_Provider *ProviderSession) Info(acc common.Address) (bool, *big.Int, *big.Int, *big.Int, error) {
	return _Provider.Contract.Info(&_Provider.CallOpts, acc)
}

// Info is a free data retrieval call binding the contract method 0x0aae7a6b.
//
// Solidity: function info(address acc) constant returns(bool, uint256, uint256, uint256)
func (_Provider *ProviderCallerSession) Info(acc common.Address) (bool, *big.Int, *big.Int, *big.Int, error) {
	return _Provider.Contract.Info(&_Provider.CallOpts, acc)
}

// IsProvider is a free data retrieval call binding the contract method 0x6b074a07.
//
// Solidity: function isProvider(address addr) constant returns(bool)
func (_Provider *ProviderCaller) IsProvider(opts *bind.CallOpts, addr common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Provider.contract.Call(opts, out, "isProvider", addr)
	return *ret0, err
}

// IsProvider is a free data retrieval call binding the contract method 0x6b074a07.
//
// Solidity: function isProvider(address addr) constant returns(bool)
func (_Provider *ProviderSession) IsProvider(addr common.Address) (bool, error) {
	return _Provider.Contract.IsProvider(&_Provider.CallOpts, addr)
}

// IsProvider is a free data retrieval call binding the contract method 0x6b074a07.
//
// Solidity: function isProvider(address addr) constant returns(bool)
func (_Provider *ProviderCallerSession) IsProvider(addr common.Address) (bool, error) {
	return _Provider.Contract.IsProvider(&_Provider.CallOpts, addr)
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

// CancelPledge is a paid mutator transaction binding the contract method 0x28c418cf.
//
// Solidity: function cancelPledge() returns(bool)
func (_Provider *ProviderTransactor) CancelPledge(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "cancelPledge")
}

// CancelPledge is a paid mutator transaction binding the contract method 0x28c418cf.
//
// Solidity: function cancelPledge() returns(bool)
func (_Provider *ProviderSession) CancelPledge() (*types.Transaction, error) {
	return _Provider.Contract.CancelPledge(&_Provider.TransactOpts)
}

// CancelPledge is a paid mutator transaction binding the contract method 0x28c418cf.
//
// Solidity: function cancelPledge() returns(bool)
func (_Provider *ProviderTransactorSession) CancelPledge() (*types.Transaction, error) {
	return _Provider.Contract.CancelPledge(&_Provider.TransactOpts)
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
// Solidity: function set(address addr, bool value) returns(bool)
func (_Provider *ProviderTransactor) Set(opts *bind.TransactOpts, addr common.Address, value bool) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "set", addr, value)
}

// Set is a paid mutator transaction binding the contract method 0x35e3b25a.
//
// Solidity: function set(address addr, bool value) returns(bool)
func (_Provider *ProviderSession) Set(addr common.Address, value bool) (*types.Transaction, error) {
	return _Provider.Contract.Set(&_Provider.TransactOpts, addr, value)
}

// Set is a paid mutator transaction binding the contract method 0x35e3b25a.
//
// Solidity: function set(address addr, bool value) returns(bool)
func (_Provider *ProviderTransactorSession) Set(addr common.Address, value bool) (*types.Transaction, error) {
	return _Provider.Contract.Set(&_Provider.TransactOpts, addr, value)
}

// SetCancelPledgeStatus is a paid mutator transaction binding the contract method 0xd78ed5da.
//
// Solidity: function setCancelPledgeStatus(address acc, uint256 sum, bool status) returns(bool)
func (_Provider *ProviderTransactor) SetCancelPledgeStatus(opts *bind.TransactOpts, acc common.Address, sum *big.Int, status bool) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "setCancelPledgeStatus", acc, sum, status)
}

// SetCancelPledgeStatus is a paid mutator transaction binding the contract method 0xd78ed5da.
//
// Solidity: function setCancelPledgeStatus(address acc, uint256 sum, bool status) returns(bool)
func (_Provider *ProviderSession) SetCancelPledgeStatus(acc common.Address, sum *big.Int, status bool) (*types.Transaction, error) {
	return _Provider.Contract.SetCancelPledgeStatus(&_Provider.TransactOpts, acc, sum, status)
}

// SetCancelPledgeStatus is a paid mutator transaction binding the contract method 0xd78ed5da.
//
// Solidity: function setCancelPledgeStatus(address acc, uint256 sum, bool status) returns(bool)
func (_Provider *ProviderTransactorSession) SetCancelPledgeStatus(acc common.Address, sum *big.Int, status bool) (*types.Transaction, error) {
	return _Provider.Contract.SetCancelPledgeStatus(&_Provider.TransactOpts, acc, sum, status)
}

// SetDeposit is a paid mutator transaction binding the contract method 0xf5bade66.
//
// Solidity: function setDeposit(uint256 _deposit) returns(bool)
func (_Provider *ProviderTransactor) SetDeposit(opts *bind.TransactOpts, _deposit *big.Int) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "setDeposit", _deposit)
}

// SetDeposit is a paid mutator transaction binding the contract method 0xf5bade66.
//
// Solidity: function setDeposit(uint256 _deposit) returns(bool)
func (_Provider *ProviderSession) SetDeposit(_deposit *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.SetDeposit(&_Provider.TransactOpts, _deposit)
}

// SetDeposit is a paid mutator transaction binding the contract method 0xf5bade66.
//
// Solidity: function setDeposit(uint256 _deposit) returns(bool)
func (_Provider *ProviderTransactorSession) SetDeposit(_deposit *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.SetDeposit(&_Provider.TransactOpts, _deposit)
}

// SetSize is a paid mutator transaction binding the contract method 0x170ab405.
//
// Solidity: function setSize(uint256 _size) returns(bool)
func (_Provider *ProviderTransactor) SetSize(opts *bind.TransactOpts, _size *big.Int) (*types.Transaction, error) {
	return _Provider.contract.Transact(opts, "setSize", _size)
}

// SetSize is a paid mutator transaction binding the contract method 0x170ab405.
//
// Solidity: function setSize(uint256 _size) returns(bool)
func (_Provider *ProviderSession) SetSize(_size *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.SetSize(&_Provider.TransactOpts, _size)
}

// SetSize is a paid mutator transaction binding the contract method 0x170ab405.
//
// Solidity: function setSize(uint256 _size) returns(bool)
func (_Provider *ProviderTransactorSession) SetSize(_size *big.Int) (*types.Transaction, error) {
	return _Provider.Contract.SetSize(&_Provider.TransactOpts, _size)
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

// ProviderApplyCancelPledgeIterator is returned from FilterApplyCancelPledge and is used to iterate over the raw logs and unpacked data for ApplyCancelPledge events raised by the Provider contract.
type ProviderApplyCancelPledgeIterator struct {
	Event *ProviderApplyCancelPledge // Event containing the contract specifics and raw log

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
func (it *ProviderApplyCancelPledgeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderApplyCancelPledge)
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
		it.Event = new(ProviderApplyCancelPledge)
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
func (it *ProviderApplyCancelPledgeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderApplyCancelPledgeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderApplyCancelPledge represents a ApplyCancelPledge event raised by the Provider contract.
type ProviderApplyCancelPledge struct {
	Acc common.Address
	Raw types.Log // Blockchain specific contextual infos
}

// FilterApplyCancelPledge is a free log retrieval operation binding the contract event 0x8444089032f19005a69aa5d9fa82009aee6241c0c090430f53abe2d6d890697c.
//
// Solidity: event ApplyCancelPledge(address acc)
func (_Provider *ProviderFilterer) FilterApplyCancelPledge(opts *bind.FilterOpts) (*ProviderApplyCancelPledgeIterator, error) {

	logs, sub, err := _Provider.contract.FilterLogs(opts, "ApplyCancelPledge")
	if err != nil {
		return nil, err
	}
	return &ProviderApplyCancelPledgeIterator{contract: _Provider.contract, event: "ApplyCancelPledge", logs: logs, sub: sub}, nil
}

// WatchApplyCancelPledge is a free log subscription operation binding the contract event 0x8444089032f19005a69aa5d9fa82009aee6241c0c090430f53abe2d6d890697c.
//
// Solidity: event ApplyCancelPledge(address acc)
func (_Provider *ProviderFilterer) WatchApplyCancelPledge(opts *bind.WatchOpts, sink chan<- *ProviderApplyCancelPledge) (event.Subscription, error) {

	logs, sub, err := _Provider.contract.WatchLogs(opts, "ApplyCancelPledge")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderApplyCancelPledge)
				if err := _Provider.contract.UnpackLog(event, "ApplyCancelPledge", log); err != nil {
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

// ProviderCancelPledgeIterator is returned from FilterCancelPledge and is used to iterate over the raw logs and unpacked data for CancelPledge events raised by the Provider contract.
type ProviderCancelPledgeIterator struct {
	Event *ProviderCancelPledge // Event containing the contract specifics and raw log

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
func (it *ProviderCancelPledgeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderCancelPledge)
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
		it.Event = new(ProviderCancelPledge)
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
func (it *ProviderCancelPledgeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderCancelPledgeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderCancelPledge represents a CancelPledge event raised by the Provider contract.
type ProviderCancelPledge struct {
	Acc    common.Address
	Money  *big.Int
	Status bool
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterCancelPledge is a free log retrieval operation binding the contract event 0xa70461ebff4d11e6f321ed483fa2998132842461adfbbae6d10dc5f3b9b23058.
//
// Solidity: event CancelPledge(address acc, uint256 money, bool status)
func (_Provider *ProviderFilterer) FilterCancelPledge(opts *bind.FilterOpts) (*ProviderCancelPledgeIterator, error) {

	logs, sub, err := _Provider.contract.FilterLogs(opts, "CancelPledge")
	if err != nil {
		return nil, err
	}
	return &ProviderCancelPledgeIterator{contract: _Provider.contract, event: "CancelPledge", logs: logs, sub: sub}, nil
}

// WatchCancelPledge is a free log subscription operation binding the contract event 0xa70461ebff4d11e6f321ed483fa2998132842461adfbbae6d10dc5f3b9b23058.
//
// Solidity: event CancelPledge(address acc, uint256 money, bool status)
func (_Provider *ProviderFilterer) WatchCancelPledge(opts *bind.WatchOpts, sink chan<- *ProviderCancelPledge) (event.Subscription, error) {

	logs, sub, err := _Provider.contract.WatchLogs(opts, "CancelPledge")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderCancelPledge)
				if err := _Provider.contract.UnpackLog(event, "CancelPledge", log); err != nil {
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

// ProviderErrorIterator is returned from FilterError and is used to iterate over the raw logs and unpacked data for Error events raised by the Provider contract.
type ProviderErrorIterator struct {
	Event *ProviderError // Event containing the contract specifics and raw log

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
func (it *ProviderErrorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderError)
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
		it.Event = new(ProviderError)
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
func (it *ProviderErrorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderErrorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderError represents a Error event raised by the Provider contract.
type ProviderError struct {
	Data string
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterError is a free log retrieval operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Provider *ProviderFilterer) FilterError(opts *bind.FilterOpts) (*ProviderErrorIterator, error) {

	logs, sub, err := _Provider.contract.FilterLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return &ProviderErrorIterator{contract: _Provider.contract, event: "Error", logs: logs, sub: sub}, nil
}

// WatchError is a free log subscription operation binding the contract event 0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa.
//
// Solidity: event Error(string data)
func (_Provider *ProviderFilterer) WatchError(opts *bind.WatchOpts, sink chan<- *ProviderError) (event.Subscription, error) {

	logs, sub, err := _Provider.contract.WatchLogs(opts, "Error")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderError)
				if err := _Provider.contract.UnpackLog(event, "Error", log); err != nil {
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

// ProviderPledgeIterator is returned from FilterPledge and is used to iterate over the raw logs and unpacked data for Pledge events raised by the Provider contract.
type ProviderPledgeIterator struct {
	Event *ProviderPledge // Event containing the contract specifics and raw log

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
func (it *ProviderPledgeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPledge)
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
		it.Event = new(ProviderPledge)
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
func (it *ProviderPledgeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPledgeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPledge represents a Pledge event raised by the Provider contract.
type ProviderPledge struct {
	Acc   common.Address
	Money *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterPledge is a free log retrieval operation binding the contract event 0x5e91ea8ea1c46300eb761859be01d7b16d44389ef91e03a163a87413cbf55b95.
//
// Solidity: event Pledge(address acc, uint256 money)
func (_Provider *ProviderFilterer) FilterPledge(opts *bind.FilterOpts) (*ProviderPledgeIterator, error) {

	logs, sub, err := _Provider.contract.FilterLogs(opts, "Pledge")
	if err != nil {
		return nil, err
	}
	return &ProviderPledgeIterator{contract: _Provider.contract, event: "Pledge", logs: logs, sub: sub}, nil
}

// WatchPledge is a free log subscription operation binding the contract event 0x5e91ea8ea1c46300eb761859be01d7b16d44389ef91e03a163a87413cbf55b95.
//
// Solidity: event Pledge(address acc, uint256 money)
func (_Provider *ProviderFilterer) WatchPledge(opts *bind.WatchOpts, sink chan<- *ProviderPledge) (event.Subscription, error) {

	logs, sub, err := _Provider.contract.WatchLogs(opts, "Pledge")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPledge)
				if err := _Provider.contract.UnpackLog(event, "Pledge", log); err != nil {
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

// ProviderSetIterator is returned from FilterSet and is used to iterate over the raw logs and unpacked data for Set events raised by the Provider contract.
type ProviderSetIterator struct {
	Event *ProviderSet // Event containing the contract specifics and raw log

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
func (it *ProviderSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderSet)
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
		it.Event = new(ProviderSet)
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
func (it *ProviderSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderSet represents a Set event raised by the Provider contract.
type ProviderSet struct {
	Addr  common.Address
	Value bool
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterSet is a free log retrieval operation binding the contract event 0xa09d518561e304be3f7de32d470dadb560b3bc168a5bad632dba82666dda9589.
//
// Solidity: event Set(address addr, bool value)
func (_Provider *ProviderFilterer) FilterSet(opts *bind.FilterOpts) (*ProviderSetIterator, error) {

	logs, sub, err := _Provider.contract.FilterLogs(opts, "Set")
	if err != nil {
		return nil, err
	}
	return &ProviderSetIterator{contract: _Provider.contract, event: "Set", logs: logs, sub: sub}, nil
}

// WatchSet is a free log subscription operation binding the contract event 0xa09d518561e304be3f7de32d470dadb560b3bc168a5bad632dba82666dda9589.
//
// Solidity: event Set(address addr, bool value)
func (_Provider *ProviderFilterer) WatchSet(opts *bind.WatchOpts, sink chan<- *ProviderSet) (event.Subscription, error) {

	logs, sub, err := _Provider.contract.WatchLogs(opts, "Set")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderSet)
				if err := _Provider.contract.UnpackLog(event, "Set", log); err != nil {
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
