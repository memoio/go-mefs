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
func DeployQuery(userAddress common.Address, hexKey string, capacity, duration int64, price *big.Int, ks int, ps int, redo bool) (common.Address, error) {
	log.Println("begin to deploy query-contract...")

	var queryAddr common.Address

	_, mapperInstance, err := GetMapperFromAdmin(userAddress, userAddress, queryKey, hexKey, true)
	if err != nil {
		return queryAddr, err
	}

	if !redo {
		queryAddr, err = GetLatestFromMapper(userAddress, mapperInstance)
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
		qAddr, tx, _, err := market.DeployQuery(auth, client, big.NewInt(capacity), big.NewInt(duration), price, big.NewInt(int64(ks)), big.NewInt(int64(ps))) //提供存储容量 存储时段 存储单价
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				log.Println("deployQueryErr:", err)
				return queryAddr, err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > checkTxRetryCount {
				log.Println("deploy query transaction fails", err)
				return queryAddr, err
			}
			continue
		}

		queryAddr = qAddr
		break
	}

	err = AddToMapper(userAddress, queryAddr, hexKey, mapperInstance)
	if err != nil {
		return queryAddr, err
	}

	log.Println("query-contract have been successfully deployed!")
	return queryAddr, nil
}

//GetQueryAddrs get all querys
func GetQueryAddrs(localAddress, userAddress common.Address) (queryAddr []common.Address, err error) {
	//获得userIndexer, key is userAddr
	_, mapperInstance, err := GetMapperFromAdmin(localAddress, userAddress, queryKey, "", false)
	if err != nil {
		return nil, err
	}

	return GetAddrsFromMapper(localAddress, mapperInstance)
}

//GetLatestQuery get latest query
func GetLatestQuery(localAddress, userAddress common.Address) (queryAddr common.Address, queryInstance *market.Query, err error) {
	//获得userIndexer, key is userAddr
	_, mapperInstance, err := GetMapperFromAdmin(localAddress, userAddress, queryKey, "", false)
	if err != nil {
		return queryAddr, queryInstance, err
	}

	querys, err := GetAddrsFromMapper(localAddress, mapperInstance)
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
		tx, err := query.SetCompleted(auth)
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				log.Println("set query Completed fails: ", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > checkTxRetryCount {
				log.Println("set query completed transaction fails", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		break
	}

	return nil
}
