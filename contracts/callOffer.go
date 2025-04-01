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

//MarketInfo  The basic information of node used for 'market' contract
type MarketInfo struct {
	addr  common.Address //local address
	hexSk string         //local privateKey
}

//NewCM new a instance of contractMarket
func NewCM(addr common.Address, hexSk string) ContractMarket {
	MInfo := &MarketInfo{
		addr:  addr,
		hexSk: hexSk,
	}

	return MInfo
}

//DeployOffer provider use it to deploy offer-contract
func (m *MarketInfo) DeployOffer(capacity, duration int64, price *big.Int, redo bool) (common.Address, error) {
	var offerAddr, oAddr common.Address
	utils.MLogger.Info("Begin to deploy offer contract...")

	//获得用户的账户余额
	a := NewCA(m.addr, m.hexSk)
	balance, err := a.QueryBalance(m.addr.Hex())
	if err != nil {
		return offerAddr, err
	}

	utils.MLogger.Infof("%s has balance: %s", m.addr.Hex(), balance)

	ma := NewCManage(m.addr, m.hexSk)
	_, mapperInstance, err := ma.GetMapperFromAdmin(m.addr, offerKey, true)
	if err != nil {
		return offerAddr, err
	}

	if !redo {
		offerAddr, err = ma.GetLatestFromMapper(mapperInstance)
		if err == nil {
			log.Println("you have deployed offer-contract")
			return offerAddr, nil
		}
	}

	//开始部署offer合约，失败则多尝试几次
	client := getClient(EndPoint)
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := makeAuth(m.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return offerAddr, errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		oAddr, tx, _, err = market.DeployOffer(auth, client, big.NewInt(capacity), big.NewInt(duration), price) //提供存储容量 存储时段 存储单价
		if oAddr.String() != InvalidAddr {
			offerAddr = oAddr
		}
		if err != nil {
			retryCount++
			log.Println("deploy Offer Err:", err)
			if err.Error() == core.ErrNonceTooLow.Error() && auth.GasPrice.Cmp(big.NewInt(defaultGasPrice)) > 0 {
				log.Println("previously pending transaction has successfully executed")
				break
			}
			if retryCount > sendTransactionRetryCount {
				return offerAddr, err
			}
			time.Sleep(retryTxSleepTime)
			continue
		}

		err = checkTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("deploy Offer transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return offerAddr, err
			}
			continue
		}
		break
	}
	log.Println("offer-contract", offerAddr.String(), "have been successfully deployed!")

	//offerAddress放进mapper
	err = ma.AddToMapper(offerAddr, mapperInstance)
	if err != nil {
		return offerAddr, err
	}

	utils.MLogger.Info("Finish deploy offer contract: ", offerAddr.Hex())

	return offerAddr, nil
}

//GetOfferAddrs get all offers
func (m *MarketInfo) GetOfferAddrs(ownerAddress common.Address) ([]common.Address, error) {
	ma := NewCManage(m.addr, m.hexSk)
	//获得userIndexer, key is userAddr
	_, mapperInstance, err := ma.GetMapperFromAdmin(ownerAddress, offerKey, false)
	if err != nil {
		return nil, err
	}

	return ma.GetAddressFromMapper(mapperInstance)
}

//ExtendOfferTime called by provider to extend the time in offer contract
func (m *MarketInfo) ExtendOfferTime(offerAddress common.Address, addTime *big.Int) error {
	offerInstance, err := market.NewOffer(offerAddress, getClient(EndPoint))
	if err != nil {
		return err
	}

	log.Println("begin to extend offerTime...")
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

		tx, err = offerInstance.Extend(auth, addTime)
		if err != nil {
			retryCount++
			log.Println("extend Offer time Err:", err)
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
			log.Println("extend Offer time transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}

	log.Println("you have called extendOfferTime successfully!")
	return nil
}

//GetOfferInfo get information about offer
func (m *MarketInfo) GetOfferInfo(offerAddress common.Address) (int64, int64, *big.Int, int64, error) {
	offerInstance, err := market.NewOffer(offerAddress, getClient(EndPoint))
	if err != nil {
		return 0, 0, big.NewInt(0), 0, err
	}

	retryCount := 0
	for {
		retryCount++
		capacity, duration, price, createDate, err := offerInstance.Get(&bind.CallOpts{
			From: m.addr,
		})
		if err != nil {
			if retryCount > 10 {
				return 0, 0, big.NewInt(0), 0, err
			}
			time.Sleep(retryGetInfoSleepTime)
			continue
		}

		return capacity.Int64(), duration.Int64(), price, createDate.Int64(), nil
	}
}
