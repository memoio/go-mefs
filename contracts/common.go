package contracts

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/memoio/go-mefs/contracts/indexer"
	"github.com/memoio/go-mefs/contracts/mapper"
	"github.com/memoio/go-mefs/contracts/resolver"
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
	ErrNotDeployedMarket      = errors.New("has not deployed query or offer")
)

type UpKeepingItem struct {
	UserID        string // 部署upkeeping的userid
	UpKeepingAddr string // 合约地址
	KeeperIDs     []string
	ProviderIDs   []string
	KeeperSLA     int32
	ProviderSLA   int32
	Duration      int64
	Capacity      int64
	Price         int64  // 部署的价格
	StartTime     string // 部署的时间
}

type ChannelItem struct {
	UserID      string // 部署Channel的userid
	ProID       string
	ChannelAddr string
	Value       *big.Int
	StartTime   string // 部署的时间
	Duration    int64  // timeout
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

//GetClient get rpc-client based the endPoint
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

//DeployResolver provider deploys resolver to save mapper
func DeployResolver(endPoint string, hexKey string, localAddress common.Address, indexer *indexer.Indexer) error {
	fmt.Println("begin deploy resolver...")
	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	client := GetClient(endPoint)

	//查看是否已经部署过
	_, resolverAddrGetted, err := indexer.Get(&bind.CallOpts{
		From: localAddress,
	}, localAddress.String())
	if err != nil {
		fmt.Println("getResolverErr:", err)
		return err
	}
	if resolverAddrGetted.String() != InvalidAddr { //说明部署过
		log.Println("you have deployed resolver already")
		return nil
	}

	//provider部署resolver
	auth = bind.NewKeyedTransactor(key)
	resolverAddr, _, _, err := resolver.DeployResolver(auth, client)
	if err != nil {
		fmt.Println("deployResolverErr:", err)
		return err
	}
	log.Println("resolverAddr:", resolverAddr.String())

	//将resolver地址放进indexer中,关键字key为provider的地址
	fmt.Print("wait for resolverAddr added into indexer...")
	auth = bind.NewKeyedTransactor(key)
	_, err = indexer.Add(auth, localAddress.String(), resolverAddr)
	if err != nil {
		fmt.Println("\naddResolverErr:", err)
		return err
	}

	//尝试从indexer中获取resolverAddr，以检测resolverAddr是否已放进indexer中
	for {
		_, resolverAddrGetted, err = indexer.Get(&bind.CallOpts{
			From: localAddress,
		}, localAddress.String())
		if err != nil {
			fmt.Println("\ngetContractsErr:", err)
			return err
		}
		if resolverAddrGetted == resolverAddr { //放进去了
			fmt.Println("done!")
			break
		}
		time.Sleep(10 * time.Second)
	}

	fmt.Println("resolver have been successfully deployed!")
	return nil
}

func getResolverFromIndexer(endPoint string, localAddress common.Address, key string) (*resolver.Resolver, error) {
	var resolverInstance *resolver.Resolver
	indexerAddr := common.HexToAddress(IndexerHex)
	indexerInstance, err := indexer.NewIndexer(indexerAddr, GetClient(endPoint))
	if err != nil {
		fmt.Println("newIndexerErr:", err)
		return resolverInstance, err
	}

	_, resolverAddr, err := indexerInstance.Get(&bind.CallOpts{
		From: localAddress,
	}, key)
	if err != nil {
		fmt.Println("getResolverErr:", err)
		return resolverInstance, err
	}
	if resolverAddr.String() == InvalidAddr {
		fmt.Println("getResolverErr:", ErrNotDeployedResolver)
		return resolverInstance, ErrNotDeployedResolver
	}

	resolverInstance, err = resolver.NewResolver(resolverAddr, GetClient(endPoint))
	if err != nil {
		fmt.Println(err)
		return resolverInstance, err
	}
	return resolverInstance, nil
}

// deployMapper 部署Mapper合约，若Mapper已经部署过，则返回已部署好的Mapper
func deployMapper(endPoint string, userAddress common.Address, resolver *resolver.Resolver, auth *bind.TransactOpts, client *ethclient.Client) (*mapper.Mapper, error) {
	//试图从resolver中取出mapper地址：mapperAddr
	var mapperAddr common.Address
	var mapperInstance *mapper.Mapper
	mapperAddr, err := resolver.Get(&bind.CallOpts{
		From: userAddress,
	}, userAddress)
	if err != nil {
		fmt.Println("getMapperErr:", err)
		return mapperInstance, err
	}

	if len(mapperAddr) == 0 || mapperAddr.String() == InvalidAddr { //没有部署过mapper
		// 部署Mapper
		mapperAddr, _, mapperInstance, err = mapper.DeployMapper(auth, client)
		if err != nil {
			fmt.Println("deployMapperErr:", err)
			return mapperInstance, err
		}
		log.Println("mapperAddr:", mapperAddr.String())

		//把mapper放进resolver
		_, err = resolver.Add(auth, mapperAddr)
		if err != nil {
			fmt.Println("addMapperErr:", err)
			return mapperInstance, err
		}
		for { //验证是否放进resolver
			mapperGetted, err := resolver.Get(&bind.CallOpts{
				From: userAddress,
			}, userAddress)
			if err != nil {
				fmt.Println("getMapperErr:", err)
				return mapperInstance, err
			}
			if mapperGetted == mapperAddr {
				break
			}
			time.Sleep(8 * time.Second)
		}
	} else { //部署过mapper，直接根据mapperAddr获得mapper
		mapperInstance, err = mapper.NewMapper(mapperAddr, GetClient(endPoint))
		if err != nil {
			fmt.Println("newMapperErr:", err)
			return mapperInstance, err
		}
	}
	return mapperInstance, nil
}

// getMapperInstance 返回已经部署的Mapper，若Mapper没部署则返回err
// 特别地，当在ChannelTimeOut()中被调用，则localAddress和ownerAddress都是userAddr；
// 当在CloseChannel()中被调用，则localAddress为providerAddr, ownerAddress为userAddr
func getMapperInstance(endPoint string, localAddress common.Address, ownerAddress common.Address, resolver *resolver.Resolver) (*mapper.Mapper, error) {
	mapperAddr, err := resolver.Get(&bind.CallOpts{
		From: localAddress,
	}, ownerAddress)
	if err != nil {
		fmt.Println("getMapperAddrErr:", err)
		return nil, err
	}
	if len(mapperAddr) == 0 || mapperAddr.String() == InvalidAddr {
		fmt.Println(ErrNotDeployedMapper)
		return nil, ErrNotDeployedMapper
	}
	mapperInstance, err := mapper.NewMapper(mapperAddr, GetClient(endPoint))
	if err != nil {
		fmt.Println("newMapperErr:", err)
		return nil, err
	}
	return mapperInstance, nil
}
