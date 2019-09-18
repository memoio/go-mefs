package contracts

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-mefs/contracts/market"
	"github.com/memoio/go-mefs/contracts/resolver"
)

//MarketType 类型声明
type MarketType int32

const (
	//Query 用来构造metaKey
	Query MarketType = iota
	//Offer 用来构造metaKey
	Offer
)

//DeployQuery user use it to deploy query-contract
func DeployQuery(userAddress common.Address, hexKey string, capacity int64, duration int64, price int64, ks int, ps int, reDeployQuery bool) (common.Address, error) {
	fmt.Println("begin to deploy query-contract...")

	var queryAddr common.Address

	//获得resolver
	_, resolver, err := getResolverFromIndexer(userAddress, "query")
	if err != nil {
		fmt.Println("getResolverErr:", err)
		return queryAddr, err
	}

	//获得mapper
	mapperInstance, err := deployMapper(userAddress, userAddress, resolver, hexKey)
	if err != nil {
		return queryAddr, err
	}

	//尝试获取query合约地址，如果已经存在就不部署
	if !reDeployQuery { //用户不想重新部署offer，那我们首先应该检查以前是否部署过，如果部署过，就直接返回，否则就部署
		queryAddr, err = getLatestAddrFromMapper(userAddress, mapperInstance)
		if err != nil {
			fmt.Println("get query addresses err:", err)
			return queryAddr, err
		}
		return queryAddr, nil
	}

	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSA err: ", err)
		return queryAddr, err
	}
	client := GetClient(EndPoint)
	// 部署query
	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		queryAddr, _, _, err = market.DeployQuery(auth, client, big.NewInt(capacity), big.NewInt(duration), big.NewInt(price), big.NewInt(int64(ks)), big.NewInt(int64(ps))) //提供存储容量 存储时段 存储单价
		if err != nil {
			if retryCount > 5 {
				fmt.Println("deployQueryErr:", err)
				return queryAddr, err
			}
			time.Sleep(time.Minute)
			continue
		}
		break
	}

	err = addToMapper(userAddress, mapperInstance, queryAddr, hexKey)
	if err != nil {
		return queryAddr, err
	}

	fmt.Println("query-contract have been successfully deployed!")
	return queryAddr, nil
}

//SetQueryCompleted when user has found providers and keepers needed, user call this function
func SetQueryCompleted(hexKey string, userAddress common.Address, queryAddress common.Address) error {
	query, err := market.NewQuery(queryAddress, GetClient(EndPoint))
	if err != nil {
		fmt.Println("newQueryErr:", err)
		return err
	}
	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		fmt.Println("HexToECDSAErr:", err)
		return err
	}
	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(key)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		_, err = query.SetCompleted(auth)
		if err != nil {
			if retryCount > 5 {
				fmt.Println("set query Completed Err:", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}
		break
	}

	return nil
}

//GetQueryInfo get user's query-info
// 分别返回申请的容量、持久化时间、价格、keeper数量、provider数量、是否成功放进upkeeping中
func GetQueryInfo(localAddress common.Address, queryAddress common.Address) (QueryItem, error) {
	var item QueryItem
	query, err := market.NewQuery(queryAddress, GetClient(EndPoint))
	if err != nil {
		fmt.Println("newQueryErr:", err)
		return item, err
	}
	retryCount := 0
	for {
		retryCount++
		capacity, duration, price, ks, ps, completed, err := query.Get(&bind.CallOpts{
			From: localAddress,
		})
		if err != nil {
			if retryCount > 10 {
				fmt.Println("getQueryParamsErr:", err)
				return item, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		item = QueryItem{
			Capacity:     capacity.Int64(),
			Duration:     duration.Int64(),
			Price:        price.Int64(),
			KeeperNums:   int32(ks.Int64()),
			ProviderNums: int32(ps.Int64()),
			Completed:    completed,
		}
		break
	}

	return item, nil
}

//DeployOffer provider use it to deploy offer-contract
func DeployOffer(providerAddress common.Address, hexKey string, capacity int64, duration int64, price int64, reDeployOffer bool) (common.Address, error) {
	fmt.Println("begin to deploy offer-contract...")
	var offerAddr common.Address
	//获得resolver实例
	_, resolverInstance, err := getResolverFromIndexer(providerAddress, "offer")
	if err != nil {
		fmt.Println("getResolverErr:", err)
		return offerAddr, err
	}

	//部署mapper，如果部署过就直接返回
	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		fmt.Println("HexToECDSAErr:", err)
		return offerAddr, err
	}

	client := GetClient(EndPoint)
	mapperInstance, err := deployMapper(providerAddress, providerAddress, resolverInstance, hexKey)
	if err != nil {
		fmt.Println("deployMapperErr:", err)
		return offerAddr, err
	}

	if !reDeployOffer { //用户不想重新部署offer，那我们首先应该检查以前是否部署过，如果部署过，就直接返回，否则就部署
		offerAddr, err := getLatestAddrFromMapper(providerAddress, mapperInstance)
		if err != nil {
			fmt.Println("get Offer Addresses Err:", err)
			return offerAddr, err
		}

		return offerAddr, nil
	}

	//部署offer
	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		offerAddr, _, _, err = market.DeployOffer(auth, client, big.NewInt(capacity), big.NewInt(duration), big.NewInt(price)) //提供存储容量 存储时段 存储单价
		if err != nil {
			if retryCount > 5 {
				fmt.Println("deploy Offer Err:", err)
				return offerAddr, err
			}
			time.Sleep(time.Minute)
			continue
		}
		break
	}

	//offerAddress放进mapper
	err = addToMapper(providerAddress, mapperInstance, offerAddr, hexKey)
	if err != nil {
		return offerAddr, err
	}

	fmt.Println("offer-contract have been successfully deployed!")
	return offerAddr, nil

}

//GetOfferInfo get provider's offer-info
func GetOfferInfo(localAddress common.Address, offerAddress common.Address) (OfferItem, error) {
	var item OfferItem
	offer, err := market.NewOffer(offerAddress, GetClient(EndPoint))
	if err != nil {
		fmt.Println("newOfferErr:", err)
		return item, err
	}

	retryCount := 0
	for {
		retryCount++
		capacity, duration, price, err := offer.Get(&bind.CallOpts{
			From: localAddress,
		})
		if err != nil {
			if retryCount > 10 {
				fmt.Println("getOfferParamsErr:", err)
				return item, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		item = OfferItem{
			Capacity: capacity.Int64(),
			Duration: duration.Int64(),
			Price:    price.Int64(),
		}
		break
	}

	return item, nil
}

// GetMarketAddr get query/offer address by MarketType
func GetMarketAddr(localAddr, ownerAddr common.Address, addrType MarketType) (common.Address, error) {
	var marketAddr common.Address
	var resolverInstance *resolver.Resolver
	var err error
	switch addrType {
	case Offer:
		_, resolverInstance, err = getResolverFromIndexer(localAddr, "offer")
		if err != nil {
			fmt.Println("getResolverErr:", err)
			return marketAddr, err
		}
	case Query:
		_, resolverInstance, err = getResolverFromIndexer(localAddr, "query")
		if err != nil {
			fmt.Println("getResolverErr:", err)
			return marketAddr, err
		}
	default:
		return marketAddr, ErrMarketType
	}

	mapperInstance, err := getMapperInstance(localAddr, ownerAddr, resolverInstance)
	if err != nil {
		fmt.Println("getMapperErr:", err)
		return marketAddr, err
	}

	marketAddr, err = getLatestAddrFromMapper(ownerAddr, mapperInstance)
	if err != nil {
		fmt.Println("getMarketAddressesErr:", err)
		return marketAddr, err
	}
	return marketAddr, nil
}
