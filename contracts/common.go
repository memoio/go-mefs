package contracts

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/memoio/go-mefs/utils"
)

const (
	//IndexerHex indexerAddress, it is well known
	IndexerHex = "0x9e4af0964ef92095ca3d2ae0c05b472837d8bd37"
	//InvalidAddr implements invalid contracts-address
	InvalidAddr          = "0x0000000000000000000000000000000000000000"
	spaceTimePayGasLimit = uint64(400000)
	spaceTimePayGasPrice = 100
)

var (
	//ErrNotDeployedMapper the user has not deployed mapper in the specified resolver
	ErrNotDeployedMapper = errors.New("has not deployed mapper")
	//ErrNotDeployedResolver the provider has not deployed resolver
	ErrNotDeployedResolver = errors.New("has not deployed resolver")
	//ErrNotDeployedUk the user has not deployed uk in the specified mapper
	ErrNotDeployedUk          = errors.New("has not deployed upKeeping")
	ErrNotDeployedChannel     = errors.New("the user has not deployed channel-contract with you")
	ErrContractNotPutToMapper = errors.New("the upKeeping-contract has not been added to mapper within a specified period of time")
	ErrMarketType             = errors.New("The market type is error, please input correct market type")
)

type UpKeepingItem struct {
	UserID        string // 部署upkeeping的userid
	UpKeepingAddr string // 合约地址
	KeeperAddrs   []string
	ProviderAddrs []string
	KeeperSla     int32
	ProviderSla   int32
	Duration      int64
	Capacity      int64
	Price         int64     // 部署的价格
	StartTime     time.Time // 部署的时间
}

type ChannelItem struct {
	UserID      string // 部署Channel的userid
	ProID       string
	ChannelAddr string
	Value       *big.Int
	StartTime   time.Time // 部署的时间
	Duration    int64     // timeout
}

type QueryItem struct {
	UserID       string // 部署Query的userid
	QueryAddr    string
	Capacity     int64
	Duration     int64
	Price        int64 // 合约给出的单价
	KeeperNums   int32
	ProviderNums int32
	Completed    bool
}

type OfferItem struct {
	ProviderID string // 部署Offer的providerid
	OfferAddr  string
	Capacity   int64
	Duration   int64
	Price      int64 // 合约给出的单价
}

// move to provider directory
type ProviderContracts struct {
	UpKeepingBook sync.Map // K-user的id, V-upkeeping
	ChannelBook   sync.Map // K-user的id, V-Channel
	QueryBook     sync.Map // K-user的id, V-Query
	Offer         OfferItem
}

var ProContracts *ProviderContracts

//GetClient get rpc-client based the endPoint
// common method
func GetClient(endPoint string) *ethclient.Client {
	client, err := rpc.Dial(endPoint)
	if err != nil {
		fmt.Println(err)
	}
	return ethclient.NewClient(client)
}

//QueryBalance query the balance of account
func QueryBalance(endPoint string, account string) (balance *big.Int, err error) {
	var result string
	client, err := rpc.Dial(endPoint)
	if err != nil {
		fmt.Println("rpc.dial err:", err)
		return balance, err
	}
	err = client.Call(&result, "eth_getBalance", account, "latest")
	if err != nil {
		fmt.Println("client.call err:", err)
		return balance, err
	}
	balance = utils.HexToBigInt(result)
	return balance, nil
}
