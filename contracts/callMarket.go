package contracts

import (
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-mefs/contracts/market"
)

//DeployQuery user use it to deploy query-contract
func DeployQuery(userAddress common.Address, hexKey string, capacity int64, duration int64, price int64, ks int, ps int, redo bool) (common.Address, error) {
	log.Println("begin to deploy query-contract...")

	var queryAddr common.Address

	//获得userIndexer, key is userAddr
	_, indexerInstance, err := GetRoleIndexer(userAddress, userAddress)
	if err != nil {
		log.Println("GetResolverErr:", err)
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
				log.Println("deployQueryErr:", err)
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

	log.Println("query-contract have been successfully deployed!")
	return queryAddr, nil
}

//GetQueryAddrs get all querys
func GetQueryAddrs(localAddress, userAddress common.Address) (queryAddr []common.Address, err error) {
	//获得userIndexer, key is userAddr
	_, indexerInstance, err := GetRoleIndexer(localAddress, userAddress)
	if err != nil {
		log.Println("GetResolverErr:", err)
		return nil, err
	}

	//获得mapper, key is upkeeping
	_, mapperInstance, err := getMapperFromIndexer(localAddress, "query", indexerInstance)
	if err != nil {
		return nil, err
	}

	return getAllFromMapper(localAddress, mapperInstance)
}

//GetLatestQuery get latest query
func GetLatestQuery(localAddress, userAddress common.Address) (queryAddr common.Address, queryInstance *market.Query, err error) {
	//获得userIndexer, key is userAddr
	_, indexerInstance, err := GetRoleIndexer(localAddress, userAddress)
	if err != nil {
		log.Println("GetResolverErr:", err)
		return queryAddr, queryInstance, err
	}

	//获得mapper, key is upkeeping
	_, mapperInstance, err := getMapperFromIndexer(localAddress, "query", indexerInstance)
	if err != nil {
		return queryAddr, queryInstance, err
	}

	querys, err := getAllFromMapper(localAddress, mapperInstance)
	if err != nil {
		return queryAddr, queryInstance, err
	}

	queryAddr = querys[len(querys)-1]

	queryInstance, err = market.NewQuery(queryAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("newQueryErr:", err)
		return queryAddr, queryInstance, err
	}

	return queryAddr, queryInstance, nil
}

//SetQueryCompleted when user has found providers and keepers needed, user call this function
func SetQueryCompleted(hexKey string, queryAddress common.Address) error {
	query, err := market.NewQuery(queryAddress, GetClient(EndPoint))
	if err != nil {
		log.Println("newQueryErr:", err)
		return err
	}
	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSAErr:", err)
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
				log.Println("set query Completed Err:", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}
		break
	}

	return nil
}

//DeployOffer provider use it to deploy offer-contract
func DeployOffer(localAddress common.Address, hexKey string, capacity int64, duration int64, price int64, redo bool) (common.Address, error) {
	log.Println("begin to deploy offer-contract...")
	var offerAddr common.Address

	//获得userIndexer, key is userAddr
	_, indexerInstance, err := GetRoleIndexer(localAddress, localAddress)
	if err != nil {
		log.Println("GetResolverErr:", err)
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
		log.Println("HexToECDSAErr:", err)
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
				log.Println("deploy Offer Err:", err)
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

	log.Println("offer-contract have been successfully deployed!")
	return offerAddr, nil
}

//GetOfferAddrs get all offers
func GetOfferAddrs(localAddress, ownerAddress common.Address) ([]common.Address, error) {
	//获得userIndexer, key is userAddr
	_, indexerInstance, err := GetRoleIndexer(localAddress, ownerAddress)
	if err != nil {
		log.Println("GetResolverErr:", err)
		return nil, err
	}

	//获得mapper, key is upkeeping
	_, mapperInstance, err := getMapperFromIndexer(localAddress, "offer", indexerInstance)
	if err != nil {
		return nil, err
	}

	return getAllFromMapper(localAddress, mapperInstance)
}

//GetLatestOffer get latest query
func GetLatestOffer(localAddress, userAddress common.Address) (offerAddr common.Address, offerInstance *market.Offer, err error) {
	//获得userIndexer, key is userAddr
	_, indexerInstance, err := GetRoleIndexer(localAddress, userAddress)
	if err != nil {
		log.Println("GetResolverErr:", err)
		return offerAddr, offerInstance, err
	}

	//获得mapper, key is upkeeping
	_, mapperInstance, err := getMapperFromIndexer(localAddress, "offer", indexerInstance)
	if err != nil {
		return offerAddr, offerInstance, err
	}

	offers, err := getAllFromMapper(localAddress, mapperInstance)
	if err != nil {
		return offerAddr, offerInstance, err
	}

	offerAddr = offers[len(offers)-1]

	offerInstance, err = market.NewOffer(offerAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("newQueryErr:", err)
		return offerAddr, offerInstance, err
	}

	return offerAddr, offerInstance, nil
}
