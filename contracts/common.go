package contracts

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/memoio/go-mefs/contracts/indexer"
	"github.com/memoio/go-mefs/contracts/mapper"
	"github.com/memoio/go-mefs/contracts/resolver"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

// EndPoint config中的ETH，在daemon中赋值
var EndPoint string

const (
	//indexerHex indexerAddress, it is well known
	indexerHex = "0x9e4af0964ef92095ca3d2ae0c05b472837d8bd37"
	//InvalidAddr implements invalid contracts-address
	InvalidAddr          = "0x0000000000000000000000000000000000000000"
	spaceTimePayGasLimit = uint64(400000)
	spaceTimePayGasPrice = 100
	defaultGasPrice      = 100
)

var (
	ErrNotDeployedIndexer = errors.New("has not deployed indexer")
	//ErrNotDeployedMapper the user has not deployed mapper in the specified resolver
	ErrNotDeployedMapper = errors.New("has not deployed mapper")
	//ErrNotDeployedResolver the provider has not deployed resolver
	ErrNotDeployedResolver = errors.New("has not deployed resolver")
	//ErrNotDeployedUk the user has not deployed uk in the specified mapper
	ErrNotDeployedUk = errors.New("has not deployed upKeeping")
	// ErrNotDeployedChannel is
	ErrNotDeployedChannel = errors.New("the user has not deployed channel-contract with you")
	// ErrContractNotPutToMapper is
	ErrContractNotPutToMapper = errors.New("the upKeeping-contract has not been added to mapper within a specified period of time")
	// ErrMarketType is
	ErrMarketType = errors.New("The market type is error, please input correct market type")
	// ErrNotDeployedMarket is
	ErrNotDeployedMarket = errors.New("has not deployed query or offer")
	// ErrNewContractInstance is
	ErrNewContractInstance = errors.New("new contract Instace failed")
	// ErrNotDeployedKPMap is
	ErrNotDeployedKPMap = errors.New("has not deployed keeperProviderMap contract")
)

// UpKeepingItem has upkeeping information
type UpKeepingItem struct {
	UserID        string // 部署upkeeping的userid
	QueryID       string // 部署upkeeping的queryID
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

// ChannelItem has channel information
type ChannelItem struct {
	UserID      string // 部署Channel的userid
	ProID       string
	ChannelAddr string
	Value       *big.Int
	Sig         []byte   // signature(channel addr, value)
	Money       *big.Int // channel has
	StartTime   string   // 部署的时间
	Duration    int64    // timeout
}

// QueryItem has query information
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

// OfferItem has offer information
type OfferItem struct {
	ProviderID string // 部署Offer的providerid
	OfferAddr  string
	Capacity   int64
	Duration   int64
	Price      int64 // 合约给出的单价
}

// ProviderItem has provider's info
type ProviderItem struct {
	ProviderID string   // providerid
	Capacity   int64    // MB
	Money      *big.Int // pledge time
	StartTime  string   // start time
}

type kpItem struct {
	keeperIDs []string
}

var kpMap sync.Map

func init() {
	EndPoint = "http://212.64.28.207:8101"
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
func QueryBalance(account string) (balance *big.Int, err error) {
	var result string
	client, err := rpc.Dial(EndPoint)
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

// DeployAdminIndexer should be resolver; to modify
func DeployAdminIndexer(hexKey string) (common.Address, *indexer.Indexer, error) {
	var indexerAddr common.Address
	var indexerInstance *indexer.Indexer

	client := GetClient(EndPoint)

	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSA err: ", err)
		return indexerAddr, indexerInstance, err
	}

	retryCount := 0
	for {
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		iAddr, tx, iInstance, err := indexer.DeployIndexer(auth, client)
		if err != nil {
			retryCount++
			if retryCount > 20 {
				fmt.Println("deploy Indexer Err:", err)
				return indexerAddr, indexerInstance, err
			}
			time.Sleep(30 * time.Second)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			log.Println("deploy indexer transaction fails", err)
			if retryCount > 20 {
				return indexerAddr, indexerInstance, err
			}
			continue
		}

		indexerAddr = iAddr
		indexerInstance = iInstance
		break
	}
	return indexerAddr, indexerInstance, nil
}

// DeployRoleIndexer deploy role's indexer and add it to admin indexer
func DeployRoleIndexer(localAddress, userAddress common.Address, hexKey string) (common.Address, *indexer.Indexer, error) {
	var indexerAddr common.Address
	var indexerInstance *indexer.Indexer

	client := GetClient(EndPoint)
	adminIndexerAddr := common.HexToAddress(indexerHex)
	adminIndexer, err := indexer.NewIndexer(adminIndexerAddr, client)
	if err != nil {
		log.Println("new Indexer err: ", err)
		return indexerAddr, indexerInstance, err
	}

	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSA err: ", err)
		return indexerAddr, indexerInstance, err
	}

	retryCount := 0
	for {
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		iAddr, tx, iInstance, err := indexer.DeployIndexer(auth, client)
		if err != nil {
			retryCount++
			if retryCount > 20 {
				fmt.Println("deploy Indexer Err:", err)
				return indexerAddr, indexerInstance, err
			}
			time.Sleep(30 * time.Second)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			log.Println("deploy user indexer transaction fails", err)
			if retryCount > 20 {
				return indexerAddr, indexerInstance, err
			}
			continue
		}

		indexerAddr = iAddr
		indexerInstance = iInstance
		break
	}

	key := userAddress.String()
	retryCount = 0
	for {
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		tx, err := adminIndexer.Add(auth, key, indexerAddr)
		if err != nil {
			retryCount++
			if retryCount > 20 {
				fmt.Println("\naddResolverErr:", err)
				return indexerAddr, indexerInstance, err
			}
			time.Sleep(30 * time.Second)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > 20 {
				log.Println("add user indexer transaction fails", err)
				return indexerAddr, indexerInstance, err
			}
			continue
		}

		retryCount = 0
		//尝试从indexer中获取resolverAddr，以检测resolverAddr是否已放进indexer中
		for {
			retryCount++
			time.Sleep(30 * time.Second)
			_, addrGetted, err := adminIndexer.Get(&bind.CallOpts{
				From: localAddress,
			}, key)
			if err != nil {
				if retryCount > 20 {
					fmt.Println("add then get Resolver Err:", err)
					return indexerAddr, indexerInstance, err
				}
				continue
			}
			if addrGetted == indexerAddr { //放进去了
				break
			}
		}
		break
	}

	return indexerAddr, indexerInstance, nil
}

// GetRoleIndexer gets role indexer
func GetRoleIndexer(localAddress, userAddress common.Address) (common.Address, *indexer.Indexer, error) {
	var indexerAddr common.Address
	var indexerInstance *indexer.Indexer

	client := GetClient(EndPoint)
	adminIndexerAddr := common.HexToAddress(indexerHex)
	adminIndexer, err := indexer.NewIndexer(adminIndexerAddr, client)
	if err != nil {
		log.Println("new Indexer err: ", err)
		return indexerAddr, indexerInstance, err
	}

	key := userAddress.String()
	retryCount := 0
	for {
		retryCount++
		_, indexerAddr, err := adminIndexer.Get(&bind.CallOpts{
			From: localAddress,
		}, key)
		if err != nil {
			if retryCount > 20 {
				return indexerAddr, indexerInstance, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		if len(indexerAddr) == 0 || indexerAddr.String() == InvalidAddr {
			return indexerAddr, indexerInstance, ErrNotDeployedIndexer
		}
		indexerInstance, err = indexer.NewIndexer(indexerAddr, client)
		if err != nil {
			return indexerAddr, indexerInstance, err
		}
		return indexerAddr, indexerInstance, nil
	}
}

// GetResolverFromIndexer gets
func GetResolverFromIndexer(localAddress common.Address, key string) (common.Address, *resolver.Resolver, error) {
	var resolverAddr common.Address
	var resolverInstance *resolver.Resolver

	client := GetClient(EndPoint)
	indexerAddr := common.HexToAddress(indexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, client)
	if err != nil {
		log.Println("new Indexer err: ", err)
		return resolverAddr, resolverInstance, err
	}

	retryCount := 0
	for {
		retryCount++
		_, resolverAddr, err := indexer.Get(&bind.CallOpts{
			From: localAddress,
		}, key)
		if err != nil {
			if retryCount > 20 {
				return resolverAddr, resolverInstance, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		if len(resolverAddr) == 0 || resolverAddr.String() == InvalidAddr {
			return resolverAddr, resolverInstance, ErrNotDeployedResolver
		}
		resolverInstance, err = resolver.NewResolver(resolverAddr, client)
		if err != nil {
			return resolverAddr, resolverInstance, err
		}
		return resolverAddr, resolverInstance, nil
	}
}

func DeployResolver(localAddress common.Address, hexKey, key string) (common.Address, *resolver.Resolver, error) {
	resolverAddr, resolverInstance, err := GetResolverFromIndexer(localAddress, key)
	if err == nil {
		return resolverAddr, resolverInstance, nil
	}

	client := GetClient(EndPoint)

	indexerAddr := common.HexToAddress(indexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, client)
	if err != nil {
		log.Println("new Indexer err: ", err)
		return resolverAddr, resolverInstance, err
	}

	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSA err: ", err)
		return resolverAddr, resolverInstance, err
	}

	retryCount := 0
	for {
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		resolverAddr, _, resolverInstance, err = resolver.DeployResolver(auth, client)
		if err != nil {
			retryCount++
			if retryCount > 20 {
				fmt.Println("deploy Resolver Err:", err)
				return resolverAddr, resolverInstance, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		break
	}

	//将resolver地址放进indexer中,关键字key可以理解为resolverAddress的索引
	//resolver-for-channel的key为providerAddr.string()
	retryCount = 0
	for {
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		tx, err := indexer.Add(auth, key, resolverAddr)
		if err != nil {
			retryCount++
			if retryCount > 20 {
				fmt.Println("\naddResolverErr:", err)
				return resolverAddr, resolverInstance, err
			}
			time.Sleep(30 * time.Second)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > 20 {
				log.Println("add user indexer transaction fails", err)
				return resolverAddr, resolverInstance, err
			}
			continue
		}

		retryCount = 0
		//尝试从indexer中获取resolverAddr，以检测resolverAddr是否已放进indexer中
		for {
			retryCount++
			time.Sleep(30 * time.Second)
			_, resolverAddrGetted, err := indexer.Get(&bind.CallOpts{
				From: localAddress,
			}, key)
			if err != nil {
				if retryCount > 20 {
					fmt.Println("add then get Resolver Err:", err)
					return resolverAddr, resolverInstance, err
				}
				continue
			}
			if resolverAddrGetted == resolverAddr { //放进去了
				break
			}
		}
		break
	}
	return resolverAddr, resolverInstance, nil
}

func getResolverFromResolver(localAddress, ownerAddress common.Address, resolverInstance *resolver.Resolver) (common.Address, *resolver.Resolver, error) {
	retryCount := 0
	for {
		retryCount++
		resolverAddr, err := resolverInstance.Get(&bind.CallOpts{
			From: localAddress,
		}, ownerAddress)
		if err != nil {
			if retryCount > 20 {
				fmt.Println("get resolve Addr err: ", err)
				return resolverAddr, nil, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		if len(resolverAddr) == 0 || resolverAddr.String() == InvalidAddr {
			return resolverAddr, nil, ErrNotDeployedResolver
		}

		secondInstance, err := resolver.NewResolver(resolverAddr, GetClient(EndPoint))
		if err != nil {
			return resolverAddr, nil, err
		}
		return resolverAddr, secondInstance, nil
	}
}

func deployResolverToResolver(localAddress common.Address, resolverInstance *resolver.Resolver, hexKey string) (common.Address, *resolver.Resolver, error) {
	resolverAddr, secondInstance, err := getResolverFromResolver(localAddress, localAddress, resolverInstance)
	if err == nil {
		return resolverAddr, secondInstance, nil
	}

	client := GetClient(EndPoint)
	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSA err: ", err)
		return resolverAddr, nil, err
	}

	retryCount := 0
	for {
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		resolverAddr, _, secondInstance, err = resolver.DeployResolver(auth, client)
		if err != nil {
			retryCount++
			if retryCount > 20 {
				fmt.Println("deploy Resolver Err:", err)
				return resolverAddr, secondInstance, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		break
	}

	//将resolver地址放进indexer中,关键字key可以理解为resolverAddress的索引
	//resolver-for-channel的key为providerAddr.string()
	retryCount = 0
	for {
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		tx, err := resolverInstance.Add(auth, resolverAddr)
		if err != nil {
			retryCount++
			if retryCount > 20 {
				fmt.Println("\naddResolverErr:", err)
				return resolverAddr, secondInstance, err
			}
			time.Sleep(30 * time.Second)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > 20 {
				log.Println("add user indexer transaction fails", err)
				return resolverAddr, resolverInstance, err
			}
			continue
		}

		retryCount = 0
		//尝试从indexer中获取resolverAddr，以检测resolverAddr是否已放进indexer中
		for {
			retryCount++
			time.Sleep(30 * time.Second)
			resolverAddrGetted, err := resolverInstance.Get(&bind.CallOpts{
				From: localAddress,
			}, localAddress)
			if err != nil {
				if retryCount > 20 {
					fmt.Println("add then get Resolver Err:", err)
					return resolverAddr, secondInstance, err
				}
				continue
			}
			if resolverAddrGetted == resolverAddr { //放进去了
				break
			}
		}
		break
	}
	return resolverAddr, secondInstance, nil
}

func GetMapperAddrFromResolver(localAddress common.Address, ownerAddress common.Address, resolverInstance *resolver.Resolver) (common.Address, error) {
	retryCount := 0
	for {
		retryCount++
		mapperAddr, err := resolverInstance.Get(&bind.CallOpts{
			From: localAddress,
		}, ownerAddress)
		if err != nil {
			if retryCount > 20 {
				fmt.Println("getMapperAddrErr:", err)
				return mapperAddr, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		if len(mapperAddr) == 0 || mapperAddr.String() == InvalidAddr {
			return mapperAddr, ErrNotDeployedMapper
		}
		return mapperAddr, nil
	}
}

// getMapperFromResolver 返回已经部署的Mapper，若Mapper没部署则返回err
// 特别地，当在ChannelTimeOut()中被调用，则localAddress和ownerAddress都是userAddr；
// 当在CloseChannel()中被调用，则localAddress为providerAddr, ownerAddress为userAddr
func getMapperFromResolver(localAddress common.Address, ownerAddress common.Address, resolverInstance *resolver.Resolver) (common.Address, *mapper.Mapper, error) {
	mapperAddr, err := GetMapperAddrFromResolver(localAddress, ownerAddress, resolverInstance)
	if err != nil {
		return mapperAddr, nil, err
	}

	mapperInstance, err := mapper.NewMapper(mapperAddr, GetClient(EndPoint))
	if err != nil {
		fmt.Println("newMapperErr:", err)
		return mapperAddr, nil, err
	}
	return mapperAddr, mapperInstance, nil
}

// DeployMapperToResolver 部署Mapper合约，若Mapper已经部署过，则返回已部署好的Mapper
func DeployMapperToResolver(localAddress common.Address, ownerAddress common.Address, resolverInstance *resolver.Resolver, hexKey string) (common.Address, *mapper.Mapper, error) {
	//试图从resolver中取出mapper地址：mapperAddr
	mapperAddr, mapperInstance, err := getMapperFromResolver(localAddress, ownerAddress, resolverInstance)
	if err == nil {
		return mapperAddr, mapperInstance, nil
	}

	client := GetClient(EndPoint)
	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("Hex To ECDSA err: ", err)
		return mapperAddr, mapperInstance, err
	}
	// 部署Mapper
	retryCount := 0
	for {
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		mAddr, tx, mInstance, err := mapper.DeployMapperToResolver(auth, client)
		if err != nil {
			if retryCount > 20 {
				fmt.Println("deployMapperErr:", err)
				return mapperAddr, mapperInstance, err
			}
			retryCount++
			time.Sleep(30 * time.Second)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			log.Println("addMapper transaction fails", err)
			if retryCount > 20 {
				return mapperAddr, mapperInstance, err
			}
			continue
		}
		mapperAddr = mAddr
		mapperInstance = mInstance
		break
	}

	//把mapper放进resolver
	retryCount = 0
	for {
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		tx, err := resolverInstance.Add(auth, mapperAddr)
		if err != nil {
			retryCount++
			if retryCount > 20 {
				return mapperAddr, mapperInstance, err
			}
			time.Sleep(30 * time.Second)
			continue
		}

		//检查交易
		err = CheckTx(tx)
		if err != nil {
			log.Println("addMapper transaction fails", err)
			if retryCount > 20 {
				return mapperAddr, mapperInstance, err
			}
			continue
		}

		retryCount = 0
		for { //验证是否放进resolver
			retryCount++
			time.Sleep(30 * time.Second)
			mapperGetted, err := resolverInstance.Get(&bind.CallOpts{
				From: localAddress,
			}, ownerAddress)
			if err != nil {
				if retryCount > 20 {
					return mapperAddr, mapperInstance, err
				}
				continue
			}
			if mapperGetted == mapperAddr {
				break
			}
		}
		break
	}

	return mapperAddr, mapperInstance, nil
}

func GetMapperAddrFromIndexer(localAddress common.Address, key string, indexerInstance *indexer.Indexer) (common.Address, error) {

	retryCount := 0
	for {
		retryCount++
		_, mapperAddr, err := indexerInstance.Get(&bind.CallOpts{
			From: localAddress,
		}, key)
		if err != nil {
			if retryCount > 20 {
				fmt.Println("getMapperAddrErr:", err)
				return mapperAddr, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		if len(mapperAddr) == 0 || mapperAddr.String() == InvalidAddr {
			return mapperAddr, ErrNotDeployedMapper
		}
		return mapperAddr, nil
	}
}

// getMapperFromResolver 返回已经部署的Mapper，若Mapper没部署则返回err
// 特别地，当在ChannelTimeOut()中被调用，则localAddress和ownerAddress都是userAddr；
// 当在CloseChannel()中被调用，则localAddress为providerAddr, ownerAddress为userAddr
func getMapperFromIndexer(localAddress common.Address, key string, indexerInstance *indexer.Indexer) (common.Address, *mapper.Mapper, error) {
	mapperAddr, err := GetMapperAddrFromIndexer(localAddress, key, indexerInstance)
	if err != nil {
		return mapperAddr, nil, err
	}

	mapperInstance, err := mapper.NewMapper(mapperAddr, GetClient(EndPoint))
	if err != nil {
		fmt.Println("newMapperErr:", err)
		return mapperAddr, nil, err
	}
	return mapperAddr, mapperInstance, nil
}

// DeployMapperToIndexer 部署Mapper合约，若Mapper已经部署过，则返回已部署好的Mapper
func DeployMapperToIndexer(localAddress common.Address, key, hexKey string, indexerInstance *indexer.Indexer) (common.Address, *mapper.Mapper, error) {
	//试图从resolver中取出mapper地址：mapperAddr
	mapperAddr, mapperInstance, err := getMapperFromIndexer(localAddress, key, indexerInstance)
	if err == nil {
		return mapperAddr, mapperInstance, nil
	}

	client := GetClient(EndPoint)
	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("Hex To ECDSA err: ", err)
		return mapperAddr, mapperInstance, err
	}
	// 部署Mapper
	retryCount := 0
	for {
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		mapperAddr, _, mapperInstance, err = mapper.DeployMapperToResolver(auth, client)
		if err != nil {
			if retryCount > 20 {
				fmt.Println("deployMapperErr:", err)
				return mapperAddr, mapperInstance, err
			}
			retryCount++
			time.Sleep(30 * time.Second)
			continue
		}
		break
	}

	//把mapper放进resolver
	retryCount = 0
	for {
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		tx, err := indexerInstance.Add(auth, key, mapperAddr)
		if err != nil {
			retryCount++
			if retryCount > 20 {
				return mapperAddr, mapperInstance, err
			}
			time.Sleep(30 * time.Second)
			continue
		}

		//检查交易
		err = CheckTx(tx)
		if err != nil {
			log.Println("addMapper transaction fails", err)
			if retryCount > 20 {
				return mapperAddr, mapperInstance, err
			}
			continue
		}

		retryCount = 0
		for { //验证是否放进resolver
			retryCount++
			time.Sleep(30 * time.Second)
			_, mapperGetted, err := indexerInstance.Get(&bind.CallOpts{
				From: localAddress,
			}, key)
			if err != nil {
				if retryCount > 20 {
					return mapperAddr, mapperInstance, err
				}
				continue
			}
			if mapperGetted == mapperAddr {
				return mapperAddr, mapperInstance, nil
			}
		}
	}

	return mapperAddr, mapperInstance, ErrNotDeployedMapper

}

func addToMapper(localAddress common.Address, mapperInstance *mapper.Mapper, addr common.Address, hexKey string) error {
	key, _ := crypto.HexToECDSA(hexKey)

	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(key)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		tx, err := mapperInstance.Add(auth, addr)
		if err != nil {
			if retryCount > 10 {
				fmt.Println("add addr to Mapper Err:", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > 20 {
				log.Println("add to mapper fails", err)
				return err
			}
			continue
		}

		retryCount = 0
		for {
			retryCount++
			time.Sleep(30 * time.Second)
			addrGetted, err := mapperInstance.Get(&bind.CallOpts{
				From: localAddress,
			})
			if err != nil {
				if retryCount > 20 {
					fmt.Println("get addr from Mapper Err:", err)
					return err
				}
				continue
			}
			length := len(addrGetted)
			if length != 0 && addrGetted[length-1] == addr {
				return nil
			}
			if retryCount > 20 {
				break
			}
		}
		break
	}
	return errors.New("add address to mapper fail")
}

func getAllFromMapper(localAddress common.Address, mapperInstance *mapper.Mapper) ([]common.Address, error) {
	var addr []common.Address
	retryCount := 0
	for {
		retryCount++
		channels, err := mapperInstance.Get(&bind.CallOpts{
			From: localAddress,
		})
		if err != nil {
			if retryCount > 20 {
				fmt.Println("get addr from mapper:", err)
				return addr, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		if len(channels) != 0 && channels[len(channels)-1].String() != InvalidAddr {
			return channels, nil
		}
		return addr, errors.New("get addr from mapper error")
	}
}

func getLatestFromMapper(localAddress common.Address, mapperInstance *mapper.Mapper) (common.Address, error) {
	var addr common.Address
	addrs, err := getAllFromMapper(localAddress, mapperInstance)
	if err != nil {
		return addr, err
	}
	return addrs[len(addrs)-1], nil
}

//CheckTx 通过交易详情检查交易是否触发了errorEvent
func CheckTx(tx *types.Transaction) error {
	log.Println("Tx hash:", tx.Hash().Hex())

	var receipt *types.Receipt
	for i := 0; i < 60; i++ {
		receipt = GetTransactionReceipt(tx.Hash())
		if receipt != nil {
			TxReceipt, _ := receipt.MarshalJSON()
			fmt.Println("TxReceipt:", string(TxReceipt))
			break
		}
		time.Sleep(30 * time.Second)
	}
	if receipt == nil { //30分钟获取不到交易信息，判定交易失败
		log.Println("transaction fails")
		return errors.New("transaction fails")
	}

	if receipt.Logs == nil || len(receipt.Logs) == 0 {
		log.Println("the transaction didn't trigger an event")
		return nil
	}
	topics := receipt.Logs[0].Topics[0].Hex()
	fmt.Println("topics:", topics)

	if topics == "0x08c379a0afcc32b1a39302f7cb8073359698411ab5fd6e3edb2c02c0b5fba8aa" {
		str := string(receipt.Logs[0].Data)
		fmt.Println(str)
		return errors.New(str)
	}
	return nil
}

//GetTransactionReceipt 通过交易hash获得交易详情
func GetTransactionReceipt(hash common.Hash) *types.Receipt {
	client, err := ethclient.Dial(EndPoint)
	if err != nil {
		log.Println("rpc.Dial err", err)
		log.Fatal(err)
	}
	receipt, err := client.TransactionReceipt(context.Background(), hash)
	return receipt
}

// SaveKpMap saves kpmap
func SaveKpMap(peerID string) error {
	localAddr, err := address.GetAddressFromID(peerID)
	if err != nil {
		log.Println("saveKpMap GetAddressFromID() error", err)
		return err
	}
	kps, err := GetAllKeeperInKPMap(localAddr)
	if err != nil {
		log.Println("saveKpMap GetAllKeepers() error", err)
		return err
	}

	for _, kpaddr := range kps {
		pids, err := GetProviderInKPMap(localAddr, kpaddr)
		if err != nil {
			log.Println("get provider from kpmap err:", err)
		}
		if len(pids) > 0 {
			keeperID, _ := address.GetIDFromAddress(kpaddr.String())
			kidList := []string{keeperID}
			for _, paddr := range pids {
				pid, _ := address.GetIDFromAddress(paddr.String())
				res, ok := kpMap.Load(pid)
				if ok {
					res.(*kpItem).keeperIDs = append(res.(*kpItem).keeperIDs, keeperID)
				} else {
					kidres := &kpItem{
						keeperIDs: kidList,
					}
					kpMap.Store(keeperID, kidres)
				}
			}
		}
	}
	return nil
}

// GetKeepersOfPro get keepers of some provider
func GetKeepersOfPro(peerID string) ([]string, bool) {
	res, ok := kpMap.Load(peerID)
	if !ok {
		return nil, false
	}
	return res.(*kpItem).keeperIDs, true
}
