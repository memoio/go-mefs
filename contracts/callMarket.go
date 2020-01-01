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
func DeployQuery(userAddress common.Address, hexKey string, capacity int64, duration int64, price int64, ks int, ps int, redo bool) (common.Address, error) {
	fmt.Println("begin to deploy query-contract...")

	var queryAddr common.Address

	//获得userIndexer, key is userAddr
	_, indexerInstance, err := GetRoleIndexer(userAddress, userAddress)
	if err != nil {
		fmt.Println("GetResolverErr:", err)
		return queryAddr, err
	}

	//获得mapper, key is query
	_, mapperInstance, err := DeployMapperToIndexer(userAddress, "query", hexKey, indexerInstance)
	if err != nil {
		return queryAddr, err
	}

	if !redo {
		queryAddr, err = getLatestFromMapper(userAddress, mapperInstance)
		if err == nil {
			return queryAddr, nil
		}
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
		qAddr, tx, _, err := market.DeployQuery(auth, client, big.NewInt(capacity), big.NewInt(duration), big.NewInt(price), big.NewInt(int64(ks)), big.NewInt(int64(ps))) //提供存储容量 存储时段 存储单价
		if err != nil {
			if retryCount > 5 {
				fmt.Println("deployQueryErr:", err)
				return queryAddr, err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > 20 {
				log.Println("deploy query transaction fails", err)
				return queryAddr, err
			}
			continue
		}

		queryAddr = qAddr
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
func SetQueryCompleted(hexKey string, queryAddress common.Address) error {
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
func DeployOffer(localAddress common.Address, hexKey string, capacity int64, duration int64, price int64, redo bool) (common.Address, error) {
	fmt.Println("begin to deploy offer-contract...")
	var offerAddr common.Address

	//获得userIndexer, key is userAddr
	_, indexerInstance, err := GetRoleIndexer(localAddress, localAddress)
	if err != nil {
		fmt.Println("GetResolverErr:", err)
		return offerAddr, err
	}

	//获得mapper, key is query
	_, mapperInstance, err := DeployMapperToIndexer(localAddress, "offer", hexKey, indexerInstance)
	if err != nil {
		return offerAddr, err
	}

	if !redo {
		offerAddr, err = getLatestFromMapper(localAddress, mapperInstance)
		if err == nil {
			return offerAddr, nil
		}
	}

	//部署mapper，如果部署过就直接返回
	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		fmt.Println("HexToECDSAErr:", err)
		return offerAddr, err
	}

	//部署offer
	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		oAddr, tx, _, err := market.DeployOffer(auth, GetClient(EndPoint), big.NewInt(capacity), big.NewInt(duration), big.NewInt(price)) //提供存储容量 存储时段 存储单价
		if err != nil {
			if retryCount > 5 {
				fmt.Println("deploy Offer Err:", err)
				return offerAddr, err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > 20 {
				log.Println("deploy offer transaction fails", err)
				return offerAddr, err
			}
			continue
		}

		offerAddr = oAddr
		break
	}

	//offerAddress放进mapper
	err = addToMapper(localAddress, mapperInstance, offerAddr, hexKey)
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
			Capacity:  capacity.Int64(),
			Duration:  duration.Int64(),
			Price:     price.Int64(),
			OfferAddr: offerAddress.String(),
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
		_, resolverInstance, err = GetResolverFromIndexer(localAddr, "offer")
		if err != nil {
			fmt.Println("GetResolverErr:", err)
			return marketAddr, err
		}
	case Query:
		_, resolverInstance, err = GetResolverFromIndexer(localAddr, "query")
		if err != nil {
			fmt.Println("GetResolverErr:", err)
			return marketAddr, err
		}
	default:
		return marketAddr, ErrMarketType
	}

	_, mapperInstance, err := getMapperFromResolver(localAddr, ownerAddr, resolverInstance)
	if err != nil {
		fmt.Println("getMapperErr:", err)
		return marketAddr, err
	}

	marketAddr, err = getLatestFromMapper(ownerAddr, mapperInstance)
	if err != nil {
		fmt.Println("getMarketAddressesErr:", err)
		return marketAddr, err
	}
	return marketAddr, nil
}
