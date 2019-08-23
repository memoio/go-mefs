package contracts

import (
	"errors"
	"math/big"
	"sync"
)

const (
	//IndexerHex indexerAddress, it is well known
	IndexerHex = "0x9e4af0964ef92095ca3d2ae0c05b472837d8bd37"
	//InvalidAddr implements invalid contracts-address
	InvalidAddr          = "0x0000000000000000000000000000000000000000"
	spaceTimePayGasLimit = uint64(400000)
	spaceTimePayGasPrice = 100
)

//EndPoint eth端口的全局变量
var EndPoint string

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
	Price         int64 // 部署的价格
}

type ChannelItem struct {
	UserID      string // 部署Channel的userid
	ChannelAddr string
	Value       *big.Int
	ProID       string
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

type RoleItem struct{}

type ProviderContracts struct {
	UpKeepingBook sync.Map // K-user的id, V-upkeeping
	ChannelBook   sync.Map // K-user的id, V-Channel
	QueryBook     sync.Map // K-user的id, V-Query
	Offer         OfferItem
}

var ProContracts *ProviderContracts
