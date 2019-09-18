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

	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSA err: ", err)
		return queryAddr, err
	}

	//获得mapper
	auth := bind.NewKeyedTransactor(sk)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	client := GetClient(EndPoint)
	mapper, err := deployMapper(userAddress, userAddress, resolver, hexKey)
	if err != nil {
		return queryAddr, err
	}

	//尝试获取query合约地址，如果已经存在就不部署
	queryLen := 0
	if !reDeployQuery { //用户不想重新部署offer，那我们首先应该检查以前是否部署过，如果部署过，就直接返回，否则就部署
		queryAddresses, err := mapper.Get(&bind.CallOpts{
			From: userAddress,
		})
		if err != nil {
			fmt.Println("getOfferAddressesErr:", err)
			return queryAddr, err
		}

		queryLen = len(queryAddresses)
		if queryLen > 0 && queryAddresses[queryLen-1].String() != InvalidAddr { //部署过
			fmt.Println("you have already deployed query-contract, so we will not deploy again")
			return queryAddresses[queryLen-1], nil
		}
	}

	// 部署query
	auth = bind.NewKeyedTransactor(sk)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	queryAddr, _, _, err = market.DeployQuery(auth, client, big.NewInt(capacity), big.NewInt(duration), big.NewInt(price), big.NewInt(int64(ks)), big.NewInt(int64(ps))) //提供存储容量 存储时段 存储单价
	if err != nil {
		fmt.Println("deployQueryErr:", err)
		return queryAddr, err
	}
	//queryAddress放进mapper,多尝试几次，以免出现未知错误

	retryCount := 0
	for {
		auth = bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		_, err = mapper.Add(auth, queryAddr)
		if err != nil {
			retryCount++
			if retryCount > 20 {
				return queryAddr, err
			}
			time.Sleep(30 * time.Second)
			continue
		}

		retryCount = 0
		//尝试从mapper中获取queryAddr，以检测queryAddr是否已放进mapper中，user可能部署过其他的query合约，检测方式是查看query合约地址有没有加1
		for {
			time.Sleep(30 * time.Second)
			queryAddresses, err := mapper.Get(&bind.CallOpts{
				From: userAddress,
			})
			if err != nil {
				retryCount++
				if retryCount > 20 {
					return queryAddr, err
				}
				continue
			}
			if len(queryAddresses) > queryLen && queryAddresses[queryLen].String() == queryAddr.String() {
				break
			}

			retryCount++
			if retryCount > 20 {
				return queryAddr, errors.New("add query addr to mapper error")
			}
		}
		break
	}

	fmt.Println("query-contract have been successfully deployed!")
	return queryAddr, err
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
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	_, err = query.SetCompleted(auth)
	if err != nil {
		fmt.Println("setCompletedErr:", err)
		return err
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
func DeployOffer(providerAddress common.Address, hexKey string, capacity int64, duration int64, price int64, reDeployOffer bool) error {
	fmt.Println("begin to deploy offer-contract...")

	//获得resolver实例
	_, resolverInstance, err := getResolverFromIndexer(providerAddress, "offer")
	if err != nil {
		fmt.Println("getResolverErr:", err)
		return err
	}

	//部署mapper，如果部署过就直接返回
	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		fmt.Println("HexToECDSAErr:", err)
		return err
	}
	auth := bind.NewKeyedTransactor(sk)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	client := GetClient(EndPoint)
	mapper, err := deployMapper(providerAddress, providerAddress, resolverInstance, hexKey)
	if err != nil {
		fmt.Println("deployMapperErr:", err)
		return err
	}

	offerLen := 0
	if !reDeployOffer { //用户不想重新部署offer，那我们首先应该检查以前是否部署过，如果部署过，就直接返回，否则就部署
		offerAddressesGetted, err := mapper.Get(&bind.CallOpts{
			From: providerAddress,
		})
		if err != nil {
			fmt.Println("getOfferAddressesErr:", err)
			return err
		}

		offerLen = len(offerAddressesGetted)
		if offerLen != 0 && offerAddressesGetted[0].String() != InvalidAddr { //代表用户之前就部署过offer
			fmt.Println("you have deployed offer already")
			return nil
		}
	}

	//部署offer
	auth = bind.NewKeyedTransactor(sk)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	offerAddr, _, _, err := market.DeployOffer(auth, client, big.NewInt(capacity), big.NewInt(duration), big.NewInt(price)) //提供存储容量 存储时段 存储单价
	if err != nil {
		fmt.Println("deployOfferErr:", err)
		return err
	}
	log.Println("offerAddr:", offerAddr.String())

	//offerAddress放进mapper,多尝试几次，以免出现未知错误
	//尝试从mapper中获取offerAddr，以检测offerAddr是否已放进mapper中，provider可能部署过其他的offer合约，检测方式是查看offer合约地址有没有加1
	retryCount := 0
	for {
		auth = bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		_, err = mapper.Add(auth, offerAddr)
		if err != nil {
			retryCount++
			if retryCount > 20 {
				return err
			}
			time.Sleep(30 * time.Second)
			continue
		}

		retryCount = 0
		//尝试从mapper中获取queryAddr，以检测queryAddr是否已放进mapper中，user可能部署过其他的query合约，检测方式是查看query合约地址有没有加1
		for {
			time.Sleep(30 * time.Second)
			offerAddresses, err := mapper.Get(&bind.CallOpts{
				From: providerAddress,
			})
			if err != nil {
				retryCount++
				if retryCount > 20 {
					return err
				}
				continue
			}
			if len(offerAddresses) > offerLen && offerAddresses[offerLen].String() == offerAddr.String() {
				break
			}

			retryCount++
			if retryCount > 20 {
				return errors.New("add offer addr to mapper error")
			}
		}
		break
	}

	fmt.Println("offer-contract have been successfully deployed!")
	return nil

}

//GetOfferInfo get provider's offer-info
func GetOfferInfo(localAddress common.Address, offerAddress common.Address) (OfferItem, error) {
	var item OfferItem
	offer, err := market.NewOffer(offerAddress, GetClient(EndPoint))
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
func getMarketAddrs(localAddr, ownerAddr common.Address, addrType MarketType) ([]common.Address, error) {
	var marketAddr []common.Address
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
	mapper, err := getMapperInstance(localAddr, ownerAddr, resolverInstance)
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

	marketLen := len(Addresses)

	if marketLen == 0 || Addresses[marketLen-1].String() == InvalidAddr {
		return marketAddr, ErrNotDeployedMarket
	}

	return marketAddr, nil
}

func GetMarketAddr(localAddr, ownerAddr common.Address, addrType MarketType) (common.Address, error) {
	var marketAddr common.Address
	addrs, err := getMarketAddrs(localAddr, ownerAddr, addrType)
	if err != nil {
		return marketAddr, err
	}

	return addrs[len(addrs)-1], nil
}
