package contracts

import (
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/memoio/go-mefs/contracts/market"
	"github.com/memoio/go-mefs/utils"
)

//DeployQuery user use it to deploy query-contract, price(wei/MB/hour)
func (m *MarketInfo) DeployQuery(capacity, storeDays int64, price *big.Int, ks int, ps int, redo bool) (common.Address, error) {
	utils.MLogger.Info("Begin to deploy query contract...")

	var queryAddr, qAddr common.Address

	// getbalance
	a := NewCA(m.addr, m.hexSk)
	balance, err := a.QueryBalance(m.addr.String())
	if err != nil {
		return queryAddr, err
	}

	utils.MLogger.Infof("%s has balance: %s", m.addr.String(), balance)

	//balance >? query + upKeeping + channel cost
	weiPrice := new(big.Float).SetInt(price)
	weiPrice.Quo(weiPrice, GetMemoPrice())
	newPrice := big.NewInt(0)
	weiPrice.Int(newPrice)

	moneyAccount := big.NewInt(24)
	moneyAccount.Mul(moneyAccount, newPrice)
	moneyAccount.Mul(moneyAccount, big.NewInt(capacity))
	moneyAccount.Mul(moneyAccount, big.NewInt(storeDays))
	// upKeeping cost
	moneyAccount.Add(moneyAccount, big.NewInt(int64(600000000)))
	moneyAccount.Add(moneyAccount, big.NewInt(1128277))

	// channel cost; read 1 times
	readPrice := big.NewInt(utils.READPRICE)
	weiRPrice := new(big.Float).SetInt64(utils.READPRICE)
	weiRPrice.Quo(weiRPrice, GetMemoPrice())
	weiRPrice.Int(readPrice)
	moneyToChannel := new(big.Int).Mul(readPrice, big.NewInt(capacity))

	moneyAccount.Add(moneyAccount, moneyToChannel)
	moneyAccount.Add(moneyAccount, big.NewInt(int64(700000*ps)))

	if moneyAccount.Cmp(balance) > 0 { //余额不足
		utils.MLogger.Infof("user %s has balance %d, but need more balance %d to start: ", m.addr.String(), balance, moneyAccount)
		return queryAddr, ErrNotEnoughBalance
	}

	//将存储时间从‘天’换算成‘秒’
	duration := storeDays * 24 * 60 * 60

	ma := NewCManage(m.addr, m.hexSk)
	_, mapperInstance, err := ma.GetMapperFromAdmin(m.addr, queryKey, true)
	if err != nil {
		return queryAddr, err
	}

	if !redo {
		queryAddr, err = ma.GetLatestFromMapper(mapperInstance)
		if err == nil {
			return queryAddr, nil
		}
	}

	log.Println("begin to deploy query-contract...")
	client := getClient(EndPoint)
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := makeAuth(m.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
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
			time.Sleep(retryTxSleepTime)
			continue
		}

		err = checkTx(tx)
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

	err = ma.AddToMapper(queryAddr, mapperInstance)
	if err != nil {
		return queryAddr, err
	}

	utils.MLogger.Info(m.addr.String(), " Finish deploy query contract: ", queryAddr.String)

	return queryAddr, nil
}

//GetQueryAddrs get all querys
func (m *MarketInfo) GetQueryAddrs(userAddress common.Address) (queryAddr []common.Address, err error) {
	ma := NewCManage(m.addr, m.hexSk)
	//获得userIndexer, key is userAddr
	_, mapperInstance, err := ma.GetMapperFromAdmin(userAddress, queryKey, false)
	if err != nil {
		return nil, err
	}

	return ma.GetAddressFromMapper(mapperInstance)
}

//SetQueryCompleted when user has found providers and keepers needed, user call this function
func (m *MarketInfo) SetQueryCompleted(queryAddress common.Address) error {
	query, err := market.NewQuery(queryAddress, getClient(EndPoint))
	if err != nil {
		log.Println("newQueryErr:", err)
		return err
	}

	log.Println("begin to set query completed...")
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := makeAuth(m.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
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
			time.Sleep(retryTxSleepTime)
			continue
		}

		err = checkTx(tx)
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

//GetQueryInfo get information about query
func (m *MarketInfo) GetQueryInfo(queryAddress common.Address) (int64, int64, *big.Int, int64, int64, bool, error) {
	queryInstance, err := market.NewQuery(queryAddress, getClient(EndPoint))
	if err != nil {
		log.Println("newQueryErr:", err)
		return 0, 0, big.NewInt(0), 0, 0, false, err
	}

	retryCount := 0
	for {
		retryCount++
		capacity, duration, price, ks, ps, completed, err := queryInstance.Get(&bind.CallOpts{
			From: m.addr,
		})
		if err != nil {
			if retryCount > 10 {
				return 0, 0, big.NewInt(0), 0, 0, false, err
			}
			time.Sleep(retryGetInfoSleepTime)
			continue
		}
		return capacity.Int64(), duration.Int64(), price, ks.Int64(), ps.Int64(), completed, nil
	}
}
