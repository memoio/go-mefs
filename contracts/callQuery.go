package contracts

import (
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/memoio/go-mefs/contracts/market"
)

//DeployQuery user use it to deploy query-contract
func DeployQuery(userAddress common.Address, hexKey string, capacity, duration int64, price *big.Int, ks int, ps int, redo bool) (common.Address, error) {
	var queryAddr, qAddr common.Address

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

	log.Println("begin to deploy query-contract...")
	client := GetClient(EndPoint)
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(hexKey, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return queryAddr, errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		qAddr, tx, _, err = market.DeployQuery(auth, client, big.NewInt(capacity), big.NewInt(duration), price, big.NewInt(int64(ks)), big.NewInt(int64(ps))) //提供存储容量 存储时段 存储单价
		if qAddr.String() != InvalidAddr {
			queryAddr = qAddr
		}
		if err != nil {
			retryCount++
			log.Println("deploy Query Err:", err)
			if err.Error() == core.ErrNonceTooLow.Error() && auth.GasPrice.Cmp(big.NewInt(defaultGasPrice)) > 0 {
				log.Println("previously pending transaction has successfully executed")
				break
			}
			if retryCount > sendTransactionRetryCount {
				return queryAddr, err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("deploy Query transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return queryAddr, err
			}
			continue
		}
		break
	}
	log.Println("query-contract", queryAddr.String(), "have been successfully deployed!")

	err = AddToMapper(queryAddr, hexKey, mapperInstance)
	if err != nil {
		return queryAddr, err
	}
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

//SetQueryCompleted when user has found providers and keepers needed, user call this function
func SetQueryCompleted(hexKey string, queryAddress common.Address) error {
	query, err := market.NewQuery(queryAddress, GetClient(EndPoint))
	if err != nil {
		log.Println("newQueryErr:", err)
		return err
	}

	log.Println("begin to set query completed...")
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(hexKey, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = query.SetCompleted(auth)
		if err != nil {
			retryCount++
			log.Println("set query completed Err:", err)
			if err.Error() == core.ErrNonceTooLow.Error() && auth.GasPrice.Cmp(big.NewInt(defaultGasPrice)) > 0 {
				log.Println("previously pending transaction has successfully executed")
				break
			}
			if retryCount > sendTransactionRetryCount {
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("set query completed transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}

	log.Println("you have called setQueryCompleted successfully!")
	return nil
}
