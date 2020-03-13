// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package upKeeping

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

// UpKeepingProof is an auto generated low-level Go binding around an user-defined struct.
type UpKeepingProof struct {
	StStart    *big.Int
	StLength   *big.Int
	StValue    *big.Int
	Keeper     common.Address
	Provider   common.Address
	MerkleRoot [32]byte
}

// UpKeepingABI is the input ABI used to generate the binding from.
const UpKeepingABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_query\",\"type\":\"address\"},{\"internalType\":\"addresspayable[]\",\"name\":\"_keepers\",\"type\":\"address[]\"},{\"internalType\":\"addresspayable[]\",\"name\":\"_providers\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"_time\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_size\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_price\",\"type\":\"uint256\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"AddOrder\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"}],\"name\":\"AddProvider\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"AlterOwner\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"PayKeeper\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"PayProvider\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"addresspayable[]\",\"name\":\"_providers\",\"type\":\"address[]\"}],\"name\":\"addProvider\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"alterOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"destruct\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"addTime\",\"type\":\"uint256\"}],\"name\":\"extendTime\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOrder\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"addresspayable[]\",\"name\":\"\",\"type\":\"address[]\"},{\"internalType\":\"addresspayable[]\",\"name\":\"\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"stStart\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stLength\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stValue\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"keeper\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"merkle_root\",\"type\":\"bytes32\"}],\"internalType\":\"structUpKeeping.Proof[]\",\"name\":\"\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_provider\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_stValue\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_stStart\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_stLength\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"_merkle_root\",\"type\":\"bytes32\"},{\"internalType\":\"uint256[]\",\"name\":\"share\",\"type\":\"uint256[]\"},{\"internalType\":\"bytes[]\",\"name\":\"sign\",\"type\":\"bytes[]\"}],\"name\":\"spaceTimePay\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]"

// UpKeepingBin is the compiled bytecode used for deploying new contracts.
var UpKeepingBin = "0x60806040819052600a80546001600160a01b03191673e0f6a00fb23458731a5c73a02a36f1df2305090b17905562001591388190039081908339810160408190526200004b9162000236565b60008054336001600160a01b031991821617909155600180549091166001600160a01b038816179055845162000089906002906020880190620000ed565b5083516200009f906003906020870190620000ed565b506004839055600582905560068190554260078190556008556040517f0905316f7faca135c292b6e6f8d91c19128d372722215fe029e74e75ef84c08790600090a1505050505050620002e7565b82805482825590600052602060002090810192821562000145579160200282015b828111156200014557825182546001600160a01b0319166001600160a01b039091161782556020909201916001909101906200010e565b506200015392915062000157565b5090565b6200017e91905b80821115620001535780546001600160a01b03191681556001016200015e565b90565b80516200018e81620002ce565b92915050565b600082601f830112620001a5578081fd5b81516001600160401b0380821115620001bc578283fd5b602080830260405182828201018181108582111715620001da578687fd5b604052848152945081850192508582018187018301881015620001fc57600080fd5b600091505b848210156200022b5762000216888262000181565b84529282019260019190910190820162000201565b505050505092915050565b60008060008060008060c087890312156200024f578182fd5b86516200025c81620002ce565b60208801519096506001600160401b038082111562000279578384fd5b620002878a838b0162000194565b965060408901519150808211156200029d578384fd5b50620002ac89828a0162000194565b945050606087015192506080870151915060a087015190509295509295509295565b6001600160a01b0381168114620002e457600080fd5b50565b61129a80620002f76000396000f3fe6080604052600436106100745760003560e01c8063a27aebbc1161004e578063a27aebbc146100ef578063c08081031461010f578063d063e7331461012f578063d36dedd2146101425761007b565b80630ca05f9f146100805780632b68b9c6146100b6578063893d20e8146100cd5761007b565b3661007b57005b600080fd5b34801561008c57600080fd5b506100a061009b366004610c3c565b61016c565b6040516100ad9190611005565b60405180910390f35b3480156100c257600080fd5b506100cb610204565b005b3480156100d957600080fd5b506100e2610245565b6040516100ad9190610f27565b3480156100fb57600080fd5b506100cb61010a366004610e0a565b610254565b34801561011b57600080fd5b506100a061012a366004610d6f565b610289565b6100cb61013d366004610c7b565b6103b2565b34801561014e57600080fd5b506101576108e2565b6040516100ad99989796959493929190610f55565b600080546001600160a01b031633146101a05760405162461bcd60e51b8152600401610197906110be565b60405180910390fd5b600080546001600160a01b038481166001600160a01b03198316179092556040519116907f8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90906101f39083908690610f3b565b60405180910390a150600192915050565b600061020e610245565b6004546007549192506002020142116102395760405162461bcd60e51b8152600401610197906110eb565b806001600160a01b0316ff5b6000546001600160a01b031690565b6000546001600160a01b0316331461027e5760405162461bcd60e51b8152600401610197906110be565b600480549091019055565b600080610294610245565b9050336001600160a01b038216148015906102b557506102b333610a92565b155b156102d25760405162461bcd60e51b815260040161019790611113565b60005b83518110156103a8576102fa8482815181106102ed57fe5b6020026020010151610ae6565b15610304576103a0565b600160020184828151811061031557fe5b60209081029190910181015182546001810184556000938452919092200180546001600160a01b0319166001600160a01b03909216919091179055835133907fa35ad2ad5abe8a31481d418a51abda97be91ba2616927300d0b75a0c340e33079086908490811061038257fe5b60200260200101516040516103979190610f27565b60405180910390a25b6001016102d5565b5060019392505050565b6103bb33610a92565b6103d75760405162461bcd60e51b815260040161019790611172565b854710156103f75760405162461bcd60e51b8152600401610197906111a0565b60085485146104185760405162461bcd60e51b815260040161019790611095565b61042187610ae6565b61043d5760405162461bcd60e51b81526004016101979061106b565b6004546007540185850111156104655760405162461bcd60e51b8152600401610197906111ce565b6000308887878a88886040516020016104849796959493929190610ead565b60408051601f198184030181529190528051602090910120600254835191925060009182805b8281101561058b57600a5487516001600160a01b03909116906319045a259088908a90859081106104d757fe5b60200260200101516040518363ffffffff1660e01b81526004016104fc929190611010565b60206040518083038186803b15801561051457600080fd5b505afa158015610528573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525061054c9190810190610c5f565b915060018001818154811061055d57fe5b6000918252602090912001546001600160a01b0383811691161415610583576001909401935b6001016104aa565b50600360028402048410156105b25760405162461bcd60e51b81526004016101979061113b565b898901600855604051600a8c04906001600160a01b038e16906009830280156108fc02916000818181858888f193505050501580156105f5573d6000803e3d6000fd5b508c6001600160a01b0316336001600160a01b03167f1569130f5bdbde161a213db1c477e4f2670f09e2a9c1c08ca9bafe749b80cb418360090260405161063c91906111fc565b60405180910390a360005b8481101561077457600280548290811061065d57fe5b60009182526020909120015489516001600160a01b03909116906108fc908b908890811061068757fe5b60200260200101518b848151811061069b57fe5b60200260200101518502816106ac57fe5b049081150290604051600060405180830381858888f193505050501580156106d8573d6000803e3d6000fd5b5060028054829081106106e757fe5b60009182526020909120015489516001600160a01b039091169033907faa4c66f6ddfadc835acfabab55148a78bc3e6867ed1cdb36461a10685af4c0c3908c908990811061073157fe5b60200260200101518c858151811061074557fe5b602002602001015186028161075657fe5b0460405161076491906111fc565b60405180910390a3600101610647565b5061077d610b33565b50506040805160c0810182529a8b5260208b01998a528a019a8b52505033606089019081526001600160a01b039a8b1660808a0190815260a08a019788526009805460018101825560009190915299516006909a027f6e1540171b6c0c960b71a7020d9f60077f6af931a8bbf590da0223dacf75c7af81019a909a5597517f6e1540171b6c0c960b71a7020d9f60077f6af931a8bbf590da0223dacf75c7b08a015598517f6e1540171b6c0c960b71a7020d9f60077f6af931a8bbf590da0223dacf75c7b1890155505095517f6e1540171b6c0c960b71a7020d9f60077f6af931a8bbf590da0223dacf75c7b2860180549189166001600160a01b031992831617905593517f6e1540171b6c0c960b71a7020d9f60077f6af931a8bbf590da0223dacf75c7b38601805491909816941693909317909555517f6e1540171b6c0c960b71a7020d9f60077f6af931a8bbf590da0223dacf75c7b490920191909155505050565b60015460045460055460065460075460085460028054604080516020808402820181019092528281526000996060998a998c998a998a998a998e998b996001600160a01b0390981698919760039796959493600993909290918a919083018282801561097757602002820191906000526020600020905b81546001600160a01b03168152600190910190602001808311610959575b50505050509750868054806020026020016040519081016040528092919081815260200182805480156109d357602002820191906000526020600020905b81546001600160a01b031681526001909101906020018083116109b5575b5050505050965081805480602002602001604051908101604052809291908181526020016000905b82821015610a6f5760008481526020908190206040805160c08101825260068602909201805483526001808201548486015260028201549284019290925260038101546001600160a01b03908116606085015260048201541660808401526005015460a083015290835290920191016109fb565b505050509150985098509850985098509850985098509850909192939495969798565b600080805b600254811015610adf576002805482908110610aaf57fe5b6000918252602090912001546001600160a01b0385811691161415610ad75760019150610adf565b600101610a97565b5092915050565b600080805b600354811015610adf576003805482908110610b0357fe5b6000918252602090912001546001600160a01b0385811691161415610b2b5760019150610adf565b600101610aeb565b6040805160c081018252600080825260208201819052918101829052606081018290526080810182905260a081019190915290565b8035610b738161124c565b92915050565b6000601f8381840112610b8a578182fd5b8235610b9d610b988261122c565b611205565b818152925060208084019085810160005b84811015610c30578135880189603f820112610bc957600080fd5b8381013567ffffffffffffffff811115610be257600080fd5b610bf3818901601f19168601611205565b81815260408c81848601011115610c0957600080fd5b82818501888401375060009181018601919091528552509282019290820190600101610bae565b50505050505092915050565b600060208284031215610c4d578081fd5b8135610c588161124c565b9392505050565b600060208284031215610c70578081fd5b8151610c588161124c565b600080600080600080600060e0888a031215610c95578283fd5b8735610ca08161124c565b9650602088810135965060408901359550606089013594506080890135935060a089013567ffffffffffffffff80821115610cd9578485fd5b818b018c601f820112610cea578586fd5b80359250610cfa610b988461122c565b8084825285820191508583018f878888028601011115610d18578889fd5b8893505b85841015610d3a578035835260019390930192918601918601610d1c565b509650505060c08b0135925080831115610d52578384fd5b5050610d608a828b01610b79565b91505092959891949750929550565b60006020808385031215610d81578182fd5b823567ffffffffffffffff811115610d97578283fd5b80840185601f820112610da8578384fd5b80359150610db8610b988361122c565b8281528381019082850185850284018601891015610dd4578687fd5b8693505b84841015610dfe57610dea8982610b68565b835260019390930192918501918501610dd8565b50979650505050505050565b600060208284031215610e1b578081fd5b5035919050565b6000815180845260208085019450808401835b83811015610e5a5781516001600160a01b031687529582019590820190600101610e35565b509495945050505050565b805182526020810151602083015260408101516040830152606081015160018060a01b038082166060850152806080840151166080850152505060a081015160a08301525050565b60006bffffffffffffffffffffffff19808a60601b168352808960601b1660148401525086602883015285604883015284606883015283608883015260a8820183518191506020808601845b83811015610f1557815185529382019390820190600101610ef9565b50929c9b505050505050505050505050565b6001600160a01b0391909116815260200190565b6001600160a01b0392831681529116602082015260400190565b6001600160a01b038a168152610120602080830182905260009190610f7c8483018d610e22565b8481036040860152610f8e818d610e22565b9250508960608501528860808501528760a085015260c0878186015284830360e086015282875180855283850191508389019450855b81811015610fe757610fd7838751610e65565b9484019491830191600101610fc4565b505080945050505050826101008301529a9950505050505050505050565b901515815260200190565b600083825260206040818401528351806040850152825b8181101561104357858101830151858201606001528201611027565b818111156110545783606083870101525b50601f01601f191692909201606001949350505050565b60208082526010908201526f34b63632b3b0b610383937bb34b232b960811b604082015260600190565b6020808252600f908201526e1a5b1b1959d85b081cdd14dd185c9d608a1b604082015260600190565b6020808252601390820152721bdb9b1e481bdddb995c8818d85b8818d85b1b606a1b604082015260600190565b6020808252600e908201526d054696d65206973206e6f742075760941b604082015260600190565b6020808252600e908201526d34b63632b3b0b61031b0b63632b960911b604082015260600190565b6020808252601c908201527f696e73756666696369656e74206c6567616c207369676e617475726500000000604082015260600190565b6020808252601490820152731bdb9b1e481ad9595c195c8818d85b8818d85b1b60621b604082015260600190565b602080825260149082015273696e73756666696369656e742062616c616e636560601b604082015260600190565b60208082526014908201527373744c656e677468206578636565642074696d6560601b604082015260600190565b90815260200190565b60405181810167ffffffffffffffff8111828210171561122457600080fd5b604052919050565b600067ffffffffffffffff821115611242578081fd5b5060209081020190565b6001600160a01b038116811461126157600080fd5b5056fea26469706673582212208ea767ba6986854f2186bf2dd9bb60ec85aa932fc8c540f3e06d2b9b7cd50f3564736f6c63430006030033"

// DeployUpKeeping deploys a new Ethereum contract, binding an instance of UpKeeping to it.
func DeployUpKeeping(auth *bind.TransactOpts, backend bind.ContractBackend, _query common.Address, _keepers []common.Address, _providers []common.Address, _time *big.Int, _size *big.Int, _price *big.Int) (common.Address, *types.Transaction, *UpKeeping, error) {
	parsed, err := abi.JSON(strings.NewReader(UpKeepingABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(UpKeepingBin), backend, _query, _keepers, _providers, _time, _size, _price)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &UpKeeping{UpKeepingCaller: UpKeepingCaller{contract: contract}, UpKeepingTransactor: UpKeepingTransactor{contract: contract}, UpKeepingFilterer: UpKeepingFilterer{contract: contract}}, nil
}

// UpKeeping is an auto generated Go binding around an Ethereum contract.
type UpKeeping struct {
	UpKeepingCaller     // Read-only binding to the contract
	UpKeepingTransactor // Write-only binding to the contract
	UpKeepingFilterer   // Log filterer for contract events
}

// UpKeepingCaller is an auto generated read-only Go binding around an Ethereum contract.
type UpKeepingCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UpKeepingTransactor is an auto generated write-only Go binding around an Ethereum contract.
type UpKeepingTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UpKeepingFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type UpKeepingFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// UpKeepingSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type UpKeepingSession struct {
	Contract     *UpKeeping        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// UpKeepingCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type UpKeepingCallerSession struct {
	Contract *UpKeepingCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// UpKeepingTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type UpKeepingTransactorSession struct {
	Contract     *UpKeepingTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// UpKeepingRaw is an auto generated low-level Go binding around an Ethereum contract.
type UpKeepingRaw struct {
	Contract *UpKeeping // Generic contract binding to access the raw methods on
}

// UpKeepingCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type UpKeepingCallerRaw struct {
	Contract *UpKeepingCaller // Generic read-only contract binding to access the raw methods on
}

// UpKeepingTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type UpKeepingTransactorRaw struct {
	Contract *UpKeepingTransactor // Generic write-only contract binding to access the raw methods on
}

// NewUpKeeping creates a new instance of UpKeeping, bound to a specific deployed contract.
func NewUpKeeping(address common.Address, backend bind.ContractBackend) (*UpKeeping, error) {
	contract, err := bindUpKeeping(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &UpKeeping{UpKeepingCaller: UpKeepingCaller{contract: contract}, UpKeepingTransactor: UpKeepingTransactor{contract: contract}, UpKeepingFilterer: UpKeepingFilterer{contract: contract}}, nil
}

// NewUpKeepingCaller creates a new read-only instance of UpKeeping, bound to a specific deployed contract.
func NewUpKeepingCaller(address common.Address, caller bind.ContractCaller) (*UpKeepingCaller, error) {
	contract, err := bindUpKeeping(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &UpKeepingCaller{contract: contract}, nil
}

// NewUpKeepingTransactor creates a new write-only instance of UpKeeping, bound to a specific deployed contract.
func NewUpKeepingTransactor(address common.Address, transactor bind.ContractTransactor) (*UpKeepingTransactor, error) {
	contract, err := bindUpKeeping(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &UpKeepingTransactor{contract: contract}, nil
}

// NewUpKeepingFilterer creates a new log filterer instance of UpKeeping, bound to a specific deployed contract.
func NewUpKeepingFilterer(address common.Address, filterer bind.ContractFilterer) (*UpKeepingFilterer, error) {
	contract, err := bindUpKeeping(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &UpKeepingFilterer{contract: contract}, nil
}

// bindUpKeeping binds a generic wrapper to an already deployed contract.
func bindUpKeeping(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(UpKeepingABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_UpKeeping *UpKeepingRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _UpKeeping.Contract.UpKeepingCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_UpKeeping *UpKeepingRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UpKeeping.Contract.UpKeepingTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_UpKeeping *UpKeepingRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _UpKeeping.Contract.UpKeepingTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_UpKeeping *UpKeepingCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _UpKeeping.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_UpKeeping *UpKeepingTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UpKeeping.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_UpKeeping *UpKeepingTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _UpKeeping.Contract.contract.Transact(opts, method, params...)
}

// GetOrder is a free data retrieval call binding the contract method 0xd36dedd2.
//
// Solidity: function getOrder() constant returns(address, address[], address[], uint256, uint256, uint256, uint256, []UpKeepingProof, uint256)
func (_UpKeeping *UpKeepingCaller) GetOrder(opts *bind.CallOpts) (common.Address, []common.Address, []common.Address, *big.Int, *big.Int, *big.Int, *big.Int, []UpKeepingProof, *big.Int, error) {
	var (
		ret0 = new(common.Address)
		ret1 = new([]common.Address)
		ret2 = new([]common.Address)
		ret3 = new(*big.Int)
		ret4 = new(*big.Int)
		ret5 = new(*big.Int)
		ret6 = new(*big.Int)
		ret7 = new([]UpKeepingProof)
		ret8 = new(*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
		ret3,
		ret4,
		ret5,
		ret6,
		ret7,
		ret8,
	}
	err := _UpKeeping.contract.Call(opts, out, "getOrder")
	return *ret0, *ret1, *ret2, *ret3, *ret4, *ret5, *ret6, *ret7, *ret8, err
}

// GetOrder is a free data retrieval call binding the contract method 0xd36dedd2.
//
// Solidity: function getOrder() constant returns(address, address[], address[], uint256, uint256, uint256, uint256, []UpKeepingProof, uint256)
func (_UpKeeping *UpKeepingSession) GetOrder() (common.Address, []common.Address, []common.Address, *big.Int, *big.Int, *big.Int, *big.Int, []UpKeepingProof, *big.Int, error) {
	return _UpKeeping.Contract.GetOrder(&_UpKeeping.CallOpts)
}

// GetOrder is a free data retrieval call binding the contract method 0xd36dedd2.
//
// Solidity: function getOrder() constant returns(address, address[], address[], uint256, uint256, uint256, uint256, []UpKeepingProof, uint256)
func (_UpKeeping *UpKeepingCallerSession) GetOrder() (common.Address, []common.Address, []common.Address, *big.Int, *big.Int, *big.Int, *big.Int, []UpKeepingProof, *big.Int, error) {
	return _UpKeeping.Contract.GetOrder(&_UpKeeping.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_UpKeeping *UpKeepingCaller) GetOwner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _UpKeeping.contract.Call(opts, out, "getOwner")
	return *ret0, err
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_UpKeeping *UpKeepingSession) GetOwner() (common.Address, error) {
	return _UpKeeping.Contract.GetOwner(&_UpKeeping.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0x893d20e8.
//
// Solidity: function getOwner() constant returns(address)
func (_UpKeeping *UpKeepingCallerSession) GetOwner() (common.Address, error) {
	return _UpKeeping.Contract.GetOwner(&_UpKeeping.CallOpts)
}

// AddProvider is a paid mutator transaction binding the contract method 0xc0808103.
//
// Solidity: function addProvider(address[] _providers) returns(bool)
func (_UpKeeping *UpKeepingTransactor) AddProvider(opts *bind.TransactOpts, _providers []common.Address) (*types.Transaction, error) {
	return _UpKeeping.contract.Transact(opts, "addProvider", _providers)
}

// AddProvider is a paid mutator transaction binding the contract method 0xc0808103.
//
// Solidity: function addProvider(address[] _providers) returns(bool)
func (_UpKeeping *UpKeepingSession) AddProvider(_providers []common.Address) (*types.Transaction, error) {
	return _UpKeeping.Contract.AddProvider(&_UpKeeping.TransactOpts, _providers)
}

// AddProvider is a paid mutator transaction binding the contract method 0xc0808103.
//
// Solidity: function addProvider(address[] _providers) returns(bool)
func (_UpKeeping *UpKeepingTransactorSession) AddProvider(_providers []common.Address) (*types.Transaction, error) {
	return _UpKeeping.Contract.AddProvider(&_UpKeeping.TransactOpts, _providers)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_UpKeeping *UpKeepingTransactor) AlterOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _UpKeeping.contract.Transact(opts, "alterOwner", newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_UpKeeping *UpKeepingSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _UpKeeping.Contract.AlterOwner(&_UpKeeping.TransactOpts, newOwner)
}

// AlterOwner is a paid mutator transaction binding the contract method 0x0ca05f9f.
//
// Solidity: function alterOwner(address newOwner) returns(bool)
func (_UpKeeping *UpKeepingTransactorSession) AlterOwner(newOwner common.Address) (*types.Transaction, error) {
	return _UpKeeping.Contract.AlterOwner(&_UpKeeping.TransactOpts, newOwner)
}

// Destruct is a paid mutator transaction binding the contract method 0x2b68b9c6.
//
// Solidity: function destruct() returns()
func (_UpKeeping *UpKeepingTransactor) Destruct(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _UpKeeping.contract.Transact(opts, "destruct")
}

// Destruct is a paid mutator transaction binding the contract method 0x2b68b9c6.
//
// Solidity: function destruct() returns()
func (_UpKeeping *UpKeepingSession) Destruct() (*types.Transaction, error) {
	return _UpKeeping.Contract.Destruct(&_UpKeeping.TransactOpts)
}

// Destruct is a paid mutator transaction binding the contract method 0x2b68b9c6.
//
// Solidity: function destruct() returns()
func (_UpKeeping *UpKeepingTransactorSession) Destruct() (*types.Transaction, error) {
	return _UpKeeping.Contract.Destruct(&_UpKeeping.TransactOpts)
}

// ExtendTime is a paid mutator transaction binding the contract method 0xa27aebbc.
//
// Solidity: function extendTime(uint256 addTime) returns()
func (_UpKeeping *UpKeepingTransactor) ExtendTime(opts *bind.TransactOpts, addTime *big.Int) (*types.Transaction, error) {
	return _UpKeeping.contract.Transact(opts, "extendTime", addTime)
}

// ExtendTime is a paid mutator transaction binding the contract method 0xa27aebbc.
//
// Solidity: function extendTime(uint256 addTime) returns()
func (_UpKeeping *UpKeepingSession) ExtendTime(addTime *big.Int) (*types.Transaction, error) {
	return _UpKeeping.Contract.ExtendTime(&_UpKeeping.TransactOpts, addTime)
}

// ExtendTime is a paid mutator transaction binding the contract method 0xa27aebbc.
//
// Solidity: function extendTime(uint256 addTime) returns()
func (_UpKeeping *UpKeepingTransactorSession) ExtendTime(addTime *big.Int) (*types.Transaction, error) {
	return _UpKeeping.Contract.ExtendTime(&_UpKeeping.TransactOpts, addTime)
}

// SpaceTimePay is a paid mutator transaction binding the contract method 0xd063e733.
//
// Solidity: function spaceTimePay(address _provider, uint256 _stValue, uint256 _stStart, uint256 _stLength, bytes32 _merkle_root, uint256[] share, bytes[] sign) returns()
func (_UpKeeping *UpKeepingTransactor) SpaceTimePay(opts *bind.TransactOpts, _provider common.Address, _stValue *big.Int, _stStart *big.Int, _stLength *big.Int, _merkle_root [32]byte, share []*big.Int, sign [][]byte) (*types.Transaction, error) {
	return _UpKeeping.contract.Transact(opts, "spaceTimePay", _provider, _stValue, _stStart, _stLength, _merkle_root, share, sign)
}

// SpaceTimePay is a paid mutator transaction binding the contract method 0xd063e733.
//
// Solidity: function spaceTimePay(address _provider, uint256 _stValue, uint256 _stStart, uint256 _stLength, bytes32 _merkle_root, uint256[] share, bytes[] sign) returns()
func (_UpKeeping *UpKeepingSession) SpaceTimePay(_provider common.Address, _stValue *big.Int, _stStart *big.Int, _stLength *big.Int, _merkle_root [32]byte, share []*big.Int, sign [][]byte) (*types.Transaction, error) {
	return _UpKeeping.Contract.SpaceTimePay(&_UpKeeping.TransactOpts, _provider, _stValue, _stStart, _stLength, _merkle_root, share, sign)
}

// SpaceTimePay is a paid mutator transaction binding the contract method 0xd063e733.
//
// Solidity: function spaceTimePay(address _provider, uint256 _stValue, uint256 _stStart, uint256 _stLength, bytes32 _merkle_root, uint256[] share, bytes[] sign) returns()
func (_UpKeeping *UpKeepingTransactorSession) SpaceTimePay(_provider common.Address, _stValue *big.Int, _stStart *big.Int, _stLength *big.Int, _merkle_root [32]byte, share []*big.Int, sign [][]byte) (*types.Transaction, error) {
	return _UpKeeping.Contract.SpaceTimePay(&_UpKeeping.TransactOpts, _provider, _stValue, _stStart, _stLength, _merkle_root, share, sign)
}

// UpKeepingAddOrderIterator is returned from FilterAddOrder and is used to iterate over the raw logs and unpacked data for AddOrder events raised by the UpKeeping contract.
type UpKeepingAddOrderIterator struct {
	Event *UpKeepingAddOrder // Event containing the contract specifics and raw log

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
func (it *UpKeepingAddOrderIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UpKeepingAddOrder)
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
		it.Event = new(UpKeepingAddOrder)
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
func (it *UpKeepingAddOrderIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UpKeepingAddOrderIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UpKeepingAddOrder represents a AddOrder event raised by the UpKeeping contract.
type UpKeepingAddOrder struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterAddOrder is a free log retrieval operation binding the contract event 0x0905316f7faca135c292b6e6f8d91c19128d372722215fe029e74e75ef84c087.
//
// Solidity: event AddOrder()
func (_UpKeeping *UpKeepingFilterer) FilterAddOrder(opts *bind.FilterOpts) (*UpKeepingAddOrderIterator, error) {

	logs, sub, err := _UpKeeping.contract.FilterLogs(opts, "AddOrder")
	if err != nil {
		return nil, err
	}
	return &UpKeepingAddOrderIterator{contract: _UpKeeping.contract, event: "AddOrder", logs: logs, sub: sub}, nil
}

// WatchAddOrder is a free log subscription operation binding the contract event 0x0905316f7faca135c292b6e6f8d91c19128d372722215fe029e74e75ef84c087.
//
// Solidity: event AddOrder()
func (_UpKeeping *UpKeepingFilterer) WatchAddOrder(opts *bind.WatchOpts, sink chan<- *UpKeepingAddOrder) (event.Subscription, error) {

	logs, sub, err := _UpKeeping.contract.WatchLogs(opts, "AddOrder")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UpKeepingAddOrder)
				if err := _UpKeeping.contract.UnpackLog(event, "AddOrder", log); err != nil {
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

// ParseAddOrder is a log parse operation binding the contract event 0x0905316f7faca135c292b6e6f8d91c19128d372722215fe029e74e75ef84c087.
//
// Solidity: event AddOrder()
func (_UpKeeping *UpKeepingFilterer) ParseAddOrder(log types.Log) (*UpKeepingAddOrder, error) {
	event := new(UpKeepingAddOrder)
	if err := _UpKeeping.contract.UnpackLog(event, "AddOrder", log); err != nil {
		return nil, err
	}
	return event, nil
}

// UpKeepingAddProviderIterator is returned from FilterAddProvider and is used to iterate over the raw logs and unpacked data for AddProvider events raised by the UpKeeping contract.
type UpKeepingAddProviderIterator struct {
	Event *UpKeepingAddProvider // Event containing the contract specifics and raw log

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
func (it *UpKeepingAddProviderIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UpKeepingAddProvider)
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
		it.Event = new(UpKeepingAddProvider)
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
func (it *UpKeepingAddProviderIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UpKeepingAddProviderIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UpKeepingAddProvider represents a AddProvider event raised by the UpKeeping contract.
type UpKeepingAddProvider struct {
	From     common.Address
	Provider common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterAddProvider is a free log retrieval operation binding the contract event 0xa35ad2ad5abe8a31481d418a51abda97be91ba2616927300d0b75a0c340e3307.
//
// Solidity: event AddProvider(address indexed from, address provider)
func (_UpKeeping *UpKeepingFilterer) FilterAddProvider(opts *bind.FilterOpts, from []common.Address) (*UpKeepingAddProviderIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _UpKeeping.contract.FilterLogs(opts, "AddProvider", fromRule)
	if err != nil {
		return nil, err
	}
	return &UpKeepingAddProviderIterator{contract: _UpKeeping.contract, event: "AddProvider", logs: logs, sub: sub}, nil
}

// WatchAddProvider is a free log subscription operation binding the contract event 0xa35ad2ad5abe8a31481d418a51abda97be91ba2616927300d0b75a0c340e3307.
//
// Solidity: event AddProvider(address indexed from, address provider)
func (_UpKeeping *UpKeepingFilterer) WatchAddProvider(opts *bind.WatchOpts, sink chan<- *UpKeepingAddProvider, from []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}

	logs, sub, err := _UpKeeping.contract.WatchLogs(opts, "AddProvider", fromRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UpKeepingAddProvider)
				if err := _UpKeeping.contract.UnpackLog(event, "AddProvider", log); err != nil {
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

// ParseAddProvider is a log parse operation binding the contract event 0xa35ad2ad5abe8a31481d418a51abda97be91ba2616927300d0b75a0c340e3307.
//
// Solidity: event AddProvider(address indexed from, address provider)
func (_UpKeeping *UpKeepingFilterer) ParseAddProvider(log types.Log) (*UpKeepingAddProvider, error) {
	event := new(UpKeepingAddProvider)
	if err := _UpKeeping.contract.UnpackLog(event, "AddProvider", log); err != nil {
		return nil, err
	}
	return event, nil
}

// UpKeepingAlterOwnerIterator is returned from FilterAlterOwner and is used to iterate over the raw logs and unpacked data for AlterOwner events raised by the UpKeeping contract.
type UpKeepingAlterOwnerIterator struct {
	Event *UpKeepingAlterOwner // Event containing the contract specifics and raw log

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
func (it *UpKeepingAlterOwnerIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UpKeepingAlterOwner)
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
		it.Event = new(UpKeepingAlterOwner)
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
func (it *UpKeepingAlterOwnerIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UpKeepingAlterOwnerIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UpKeepingAlterOwner represents a AlterOwner event raised by the UpKeeping contract.
type UpKeepingAlterOwner struct {
	From common.Address
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAlterOwner is a free log retrieval operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_UpKeeping *UpKeepingFilterer) FilterAlterOwner(opts *bind.FilterOpts) (*UpKeepingAlterOwnerIterator, error) {

	logs, sub, err := _UpKeeping.contract.FilterLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return &UpKeepingAlterOwnerIterator{contract: _UpKeeping.contract, event: "AlterOwner", logs: logs, sub: sub}, nil
}

// WatchAlterOwner is a free log subscription operation binding the contract event 0x8c153ecee6895f15da72e646b4029e0ef7cbf971986d8d9cfe48c5563d368e90.
//
// Solidity: event AlterOwner(address from, address to)
func (_UpKeeping *UpKeepingFilterer) WatchAlterOwner(opts *bind.WatchOpts, sink chan<- *UpKeepingAlterOwner) (event.Subscription, error) {

	logs, sub, err := _UpKeeping.contract.WatchLogs(opts, "AlterOwner")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UpKeepingAlterOwner)
				if err := _UpKeeping.contract.UnpackLog(event, "AlterOwner", log); err != nil {
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
func (_UpKeeping *UpKeepingFilterer) ParseAlterOwner(log types.Log) (*UpKeepingAlterOwner, error) {
	event := new(UpKeepingAlterOwner)
	if err := _UpKeeping.contract.UnpackLog(event, "AlterOwner", log); err != nil {
		return nil, err
	}
	return event, nil
}

// UpKeepingPayKeeperIterator is returned from FilterPayKeeper and is used to iterate over the raw logs and unpacked data for PayKeeper events raised by the UpKeeping contract.
type UpKeepingPayKeeperIterator struct {
	Event *UpKeepingPayKeeper // Event containing the contract specifics and raw log

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
func (it *UpKeepingPayKeeperIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UpKeepingPayKeeper)
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
		it.Event = new(UpKeepingPayKeeper)
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
func (it *UpKeepingPayKeeperIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UpKeepingPayKeeperIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UpKeepingPayKeeper represents a PayKeeper event raised by the UpKeeping contract.
type UpKeepingPayKeeper struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterPayKeeper is a free log retrieval operation binding the contract event 0xaa4c66f6ddfadc835acfabab55148a78bc3e6867ed1cdb36461a10685af4c0c3.
//
// Solidity: event PayKeeper(address indexed from, address indexed to, uint256 value)
func (_UpKeeping *UpKeepingFilterer) FilterPayKeeper(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*UpKeepingPayKeeperIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _UpKeeping.contract.FilterLogs(opts, "PayKeeper", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &UpKeepingPayKeeperIterator{contract: _UpKeeping.contract, event: "PayKeeper", logs: logs, sub: sub}, nil
}

// WatchPayKeeper is a free log subscription operation binding the contract event 0xaa4c66f6ddfadc835acfabab55148a78bc3e6867ed1cdb36461a10685af4c0c3.
//
// Solidity: event PayKeeper(address indexed from, address indexed to, uint256 value)
func (_UpKeeping *UpKeepingFilterer) WatchPayKeeper(opts *bind.WatchOpts, sink chan<- *UpKeepingPayKeeper, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _UpKeeping.contract.WatchLogs(opts, "PayKeeper", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UpKeepingPayKeeper)
				if err := _UpKeeping.contract.UnpackLog(event, "PayKeeper", log); err != nil {
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

// ParsePayKeeper is a log parse operation binding the contract event 0xaa4c66f6ddfadc835acfabab55148a78bc3e6867ed1cdb36461a10685af4c0c3.
//
// Solidity: event PayKeeper(address indexed from, address indexed to, uint256 value)
func (_UpKeeping *UpKeepingFilterer) ParsePayKeeper(log types.Log) (*UpKeepingPayKeeper, error) {
	event := new(UpKeepingPayKeeper)
	if err := _UpKeeping.contract.UnpackLog(event, "PayKeeper", log); err != nil {
		return nil, err
	}
	return event, nil
}

// UpKeepingPayProviderIterator is returned from FilterPayProvider and is used to iterate over the raw logs and unpacked data for PayProvider events raised by the UpKeeping contract.
type UpKeepingPayProviderIterator struct {
	Event *UpKeepingPayProvider // Event containing the contract specifics and raw log

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
func (it *UpKeepingPayProviderIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(UpKeepingPayProvider)
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
		it.Event = new(UpKeepingPayProvider)
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
func (it *UpKeepingPayProviderIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *UpKeepingPayProviderIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// UpKeepingPayProvider represents a PayProvider event raised by the UpKeeping contract.
type UpKeepingPayProvider struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterPayProvider is a free log retrieval operation binding the contract event 0x1569130f5bdbde161a213db1c477e4f2670f09e2a9c1c08ca9bafe749b80cb41.
//
// Solidity: event PayProvider(address indexed from, address indexed to, uint256 value)
func (_UpKeeping *UpKeepingFilterer) FilterPayProvider(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*UpKeepingPayProviderIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _UpKeeping.contract.FilterLogs(opts, "PayProvider", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &UpKeepingPayProviderIterator{contract: _UpKeeping.contract, event: "PayProvider", logs: logs, sub: sub}, nil
}

// WatchPayProvider is a free log subscription operation binding the contract event 0x1569130f5bdbde161a213db1c477e4f2670f09e2a9c1c08ca9bafe749b80cb41.
//
// Solidity: event PayProvider(address indexed from, address indexed to, uint256 value)
func (_UpKeeping *UpKeepingFilterer) WatchPayProvider(opts *bind.WatchOpts, sink chan<- *UpKeepingPayProvider, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _UpKeeping.contract.WatchLogs(opts, "PayProvider", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(UpKeepingPayProvider)
				if err := _UpKeeping.contract.UnpackLog(event, "PayProvider", log); err != nil {
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

// ParsePayProvider is a log parse operation binding the contract event 0x1569130f5bdbde161a213db1c477e4f2670f09e2a9c1c08ca9bafe749b80cb41.
//
// Solidity: event PayProvider(address indexed from, address indexed to, uint256 value)
func (_UpKeeping *UpKeepingFilterer) ParsePayProvider(log types.Log) (*UpKeepingPayProvider, error) {
	event := new(UpKeepingPayProvider)
	if err := _UpKeeping.contract.UnpackLog(event, "PayProvider", log); err != nil {
		return nil, err
	}
	return event, nil
}
