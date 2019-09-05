package contracts

import (
	"crypto/ecdsa"
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
func DeployQuery(endPoint string, userAddress common.Address, sk *ecdsa.PrivateKey, capacity int64, duration int64, price int64, ks int, ps int) error {
	fmt.Println("begin to deploy query-contract...")
	//获得resolver
	resolver, err := getResolverFromIndexer(endPoint, userAddress, "query")
	if err != nil {
		fmt.Println("getResolverErr:", err)
		return err
	}

	//获得mapper
	auth := bind.NewKeyedTransactor(sk)
	client := GetClient(endPoint)
	mapper, err := deployMapper(endPoint, userAddress, resolver, auth, client)
	if err != nil {
		return err
	}

	//尝试获取query合约地址，如果已经存在就不部署
	var queryAddresses []common.Address
	queryAddresses, err = mapper.Get(&bind.CallOpts{
		From: userAddress,
	})
	if err != nil {
		fmt.Println("getQueryAddressesErr:", err)
		return err
	}
	length := len(queryAddresses)
	if (length != 0) && (queryAddresses[0].String() != InvalidAddr) { //部署过
		fmt.Println("you have already deployed query-contract, so we will not deploy again")
		return nil
	}

	// 部署query
	auth = bind.NewKeyedTransactor(sk)
	queryAddr, _, _, err := market.DeployQuery(auth, client, big.NewInt(capacity), big.NewInt(duration), big.NewInt(price), big.NewInt(int64(ks)), big.NewInt(int64(ps))) //提供存储容量 存储时段 存储单价
	if err != nil {
		fmt.Println("deployQueryErr:", err)
		return err
	}
	log.Println("queryAddr:", queryAddr.String())

	//queryAddress放进mapper,多尝试几次，以免出现未知错误
	auth = bind.NewKeyedTransactor(sk)
	for addToMapperCount := 0; addToMapperCount < 2; addToMapperCount++ {
		time.Sleep(10 * time.Second)
		_, err = mapper.Add(auth, queryAddr)
		if err == nil {
			break
		}
	}
	if err != nil {
		fmt.Println("addQueryAddressErr:", err)
		return err
	}

	//尝试从mapper中获取queryAddr，以检测queryAddr是否已放进mapper中，user可能部署过其他的query合约，检测方式是查看query合约地址有没有加1
	for {
		queryAddresses, err = mapper.Get(&bind.CallOpts{
			From: userAddress,
		})
		if err != nil {
			fmt.Println("getQueryAddressesErr:", err)
			return err
		}
		if len(queryAddresses) > length && queryAddresses[length].String() != InvalidAddr {
			break
		} else {
			time.Sleep(10 * time.Second)
		}
	}
	fmt.Println("query-contract have been successfully deployed!")
	return nil

}

//SetQueryCompleted when user has found providers and keepers needed, user call this function
func SetQueryCompleted(endPoint string, hexKey string, userAddress common.Address, queryAddress common.Address) error {
	query, err := market.NewQuery(queryAddress, GetClient(endPoint))
	if err != nil {
		fmt.Println("newQueryErr:", err)
		return err
	}
	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		fmt.Println("HexToECDSAErr:", err)
		return err
	}
	auth := bind.NewKeyedTransactor(key)
	_, err = query.SetCompleted(auth)
	if err != nil {
		fmt.Println("setCompletedErr:", err)
		return err
	}
	return nil
}

//GetQueryInfo get user's query-info
// 分别返回申请的容量、持久化时间、价格、keeper数量、provider数量、是否成功放进upkeeping中
func GetQueryInfo(endPoint string, localAddress common.Address, queryAddress common.Address) (QueryItem, error) {
	var item QueryItem
	query, err := market.NewQuery(queryAddress, GetClient(endPoint))
	if err != nil {
		fmt.Println("newQueryErr:", err)
		return item, err
	}
	capacity, duration, price, ks, ps, completed, err := query.Get(&bind.CallOpts{
		From: localAddress,
	})
	if err != nil {
		fmt.Println("getQueryParamsErr:", err)
		return item, err
	}
	item = QueryItem{
		Capacity:     capacity.Int64(),
		Duration:     duration.Int64(),
		Price:        price.Int64(),
		KeeperNums:   int32(ks.Int64()),
		ProviderNums: int32(ps.Int64()),
		Completed:    completed,
	}
	return item, nil
}

//DeployOffer provider use it to deploy offer-contract
func DeployOffer(endPoint string, providerAddress common.Address, hexKey string, capacity int64, duration int64, price int64) error {
	fmt.Println("begin to deploy offer-contract...")
	//获得resolver
	resolver, err := getResolverFromIndexer(endPoint, providerAddress, "offer")
	if err != nil {
		fmt.Println("getResolverErr:", err)
		return err
	}

	//获得mapper
	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		fmt.Println("HexToECDSAErr:", err)
		return err
	}
	auth := bind.NewKeyedTransactor(key)
	client := GetClient(endPoint)
	mapper, err := deployMapper(endPoint, providerAddress, resolver, auth, client)
	if err != nil {
		return err
	}

	// 部署offer
	auth = bind.NewKeyedTransactor(key)
	offerAddr, _, _, err := market.DeployOffer(auth, client, big.NewInt(capacity), big.NewInt(duration), big.NewInt(price)) //提供存储容量 存储时段 存储单价
	if err != nil {
		fmt.Println("deployOfferErr:", err)
		return err
	}
	log.Println("offerAddr:", offerAddr.String())

	//offerAddress放进mapper,多尝试几次，以免出现未知错误
	auth = bind.NewKeyedTransactor(key)
	for addToMapperCount := 0; addToMapperCount < 2; addToMapperCount++ {
		time.Sleep(10 * time.Second)
		_, err = mapper.Add(auth, offerAddr)
		if err == nil {
			break
		}
	}
	if err != nil {
		fmt.Println("addOfferAddressErr:", err)
		return err
	}

	//尝试从mapper中获取offerAddr，以检测offerAddr是否已放进mapper中，provider可能部署过其他的offer合约，检测方式是查看offer合约地址有没有加1
	offerAddresses, err := mapper.Get(&bind.CallOpts{
		From: providerAddress,
	})
	if err != nil {
		fmt.Println("getOfferAddressesErr:", err)
		return err
	}
	length := len(offerAddresses)
	for {
		offerAddresses, err = mapper.Get(&bind.CallOpts{
			From: providerAddress,
		})
		if err != nil {
			fmt.Println("getOfferAddressesErr:", err)
			return err
		}
		if len(offerAddresses) > length && offerAddresses[length].String() != InvalidAddr {
			break
		} else {
			time.Sleep(10 * time.Second)
		}
	}
	fmt.Println("offer-contract have been successfully deployed!")
	return nil

}

//GetOfferInfo get provider's offer-info
func GetOfferInfo(endPoint string, localAddress common.Address, offerAddress common.Address) (OfferItem, error) {
	var item OfferItem
	offer, err := market.NewOffer(offerAddress, GetClient(endPoint))
	if err != nil {
		fmt.Println("newOfferErr:", err)
		return item, err
	}
	capacity, duration, price, err := offer.Get(&bind.CallOpts{
		From: localAddress,
	})
	if err != nil {
		fmt.Println("getOfferParamsErr:", err)
		return item, err
	}
	item = OfferItem{
		Capacity: capacity.Int64(),
		Duration: duration.Int64(),
		Price:    price.Int64(),
	}
	return item, nil
}

// GetMarketAddr get query/offer address by MarketType
func GetMarketAddr(endPoint string, localAddr, ownerAddr common.Address, addrType MarketType) (common.Address, error) {
	var marketAddr common.Address
	var resolverInstance *resolver.Resolver
	var err error
	switch addrType {
	case Offer:
		resolverInstance, err = getResolverFromIndexer(endPoint, localAddr, "offer")
		if err != nil {
			fmt.Println("getResolverErr:", err)
			return marketAddr, err
		}
	case Query:
		resolverInstance, err = getResolverFromIndexer(endPoint, localAddr, "query")
		if err != nil {
			fmt.Println("getResolverErr:", err)
			return marketAddr, err
		}
	default:
		return marketAddr, ErrMarketType
	}
	mapper, err := getMapperInstance(endPoint, localAddr, ownerAddr, resolverInstance)
	if err != nil {
		fmt.Println("getMapperErr:", err)
		return marketAddr, err
	}
	Addresses, err := mapper.Get(&bind.CallOpts{
		From: ownerAddr,
	})
	if err != nil {
		fmt.Println("getMarketAddressesErr:", err)
		return marketAddr, err
	}
	if len(Addresses) == 0 {
		return marketAddr, ErrNotDeployedMarket
	}
	// 返回最新的offer、query地址
	marketAddr = Addresses[len(Addresses)-1]
	return marketAddr, nil
}
