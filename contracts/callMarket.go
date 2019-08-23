package contracts

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-mefs/contracts/market"
	"github.com/memoio/go-mefs/contracts/upKeeping"
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
	mapper, err := getMapper(endPoint, userAddress, resolver, auth, client)
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

func GetQueryAddress(endPoint string, userAddress common.Address) (common.Address, error) {
	var queryAddr common.Address
	//获得resolver
	resolver, err := getResolverFromIndexer(endPoint, userAddress, "query")
	if err != nil {
		fmt.Println("getResolverErr:", err)
		return queryAddr, err
	}

	//获得mapper
	mapperAddr, err := resolver.Get(&bind.CallOpts{
		From: userAddress,
	}, userAddress)
	if err != nil {
		fmt.Println("getMapperErr:", err)
		return queryAddr, err
	}
	mapper, err := upKeeping.NewMapper(mapperAddr, GetClient(endPoint))
	if err != nil {
		fmt.Println("newMapperErr:", err)
		return queryAddr, err
	}

	//获得queryAddress
	queryAddresses, err := mapper.Get(&bind.CallOpts{
		From: userAddress,
	})
	if err != nil {
		fmt.Println("getQueryAddressesErr:", err)
		return queryAddr, err
	}
	length := len(queryAddresses)
	if length < 1 || queryAddresses[length-1].String() == InvalidAddr {
		fmt.Println("user don't have query-contract")
		return queryAddr, errors.New("user don't have query-contract")
	}
	return queryAddresses[length-1], nil
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
	mapper, err := getMapper(endPoint, providerAddress, resolver, auth, client)
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

//GetQueryParams get user's query-params
// 分别返回申请的容量、持久化时间、价格、keeper数量、provider数量、是否成功放进upkeeping中
func GetQueryParams(endPoint string, localAddress common.Address, queryAddress common.Address) (*big.Int, *big.Int, *big.Int, *big.Int, *big.Int, bool, error) {
	query, err := market.NewQuery(queryAddress, GetClient(endPoint))
	if err != nil {
		fmt.Println("newQueryErr:", err)
		return nil, nil, nil, nil, nil, false, err
	}
	capacity, duration, price, ks, ps, completed, err := query.Get(&bind.CallOpts{
		From: localAddress,
	})
	if err != nil {
		fmt.Println("getQueryParamsErr:", err)
		return nil, nil, nil, nil, nil, false, err
	}
	return capacity, duration, price, ks, ps, completed, nil
}

//GetOfferParams get provider's offer-params
func GetOfferParams(endPoint string, localAddress common.Address, offerAddress common.Address) (*big.Int, *big.Int, *big.Int, error) {
	offer, err := market.NewOffer(offerAddress, GetClient(endPoint))
	if err != nil {
		fmt.Println("newOfferErr:", err)
		return nil, nil, nil, err
	}
	capacity, duration, price, err := offer.Get(&bind.CallOpts{
		From: localAddress,
	})
	if err != nil {
		fmt.Println("getOfferParamsErr:", err)
		return nil, nil, nil, err
	}
	return capacity, duration, price, nil
}

func GetMarketAddr(endPoint string, localAddr, ownerAddr common.Address, addrType MarketType) (common.Address, error) {
	var marketAddr common.Address
	var resolver *upKeeping.Resolver
	var err error
	switch addrType {
	case Offer:
		resolver, err = getResolverFromIndexer(endPoint, localAddr, "offer")
		if err != nil {
			fmt.Println("getResolverErr:", err)
			return marketAddr, err
		}
	case Query:
		resolver, err = getResolverFromIndexer(endPoint, localAddr, "query")
		if err != nil {
			fmt.Println("getResolverErr:", err)
			return marketAddr, err
		}
	default:
		return marketAddr, ErrMarketType
	}
	mapper, err := getDeployedMapper(endPoint, localAddr, ownerAddr, resolver)
	if err != nil {
		fmt.Println("getMapperErr:", err)
		return marketAddr, err
	}
	offerAddresses, err := mapper.Get(&bind.CallOpts{
		From: ownerAddr,
	})
	if err != nil {
		fmt.Println("getOfferAddressesErr:", err)
		return marketAddr, err
	}
	// 返回最新的offer地址
	marketAddr = offerAddresses[len(offerAddresses)-1]
	return marketAddr, nil
}
