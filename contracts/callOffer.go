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

//DeployOffer provider use it to deploy offer-contract
func DeployOffer(localAddress common.Address, hexKey string, capacity, duration int64, price *big.Int, redo bool) (common.Address, error) {
	var offerAddr, oAddr common.Address

	_, mapperInstance, err := GetMapperFromAdmin(localAddress, localAddress, offerKey, hexKey, true)
	if err != nil {
		return offerAddr, err
	}

	if !redo {
		offerAddr, err = GetLatestFromMapper(localAddress, mapperInstance)
		if err == nil {
			log.Println("you have deployed offer-contract")
			return offerAddr, nil
		}
	}

	log.Println("begin to deploy offer-contract...")
	client := GetClient(EndPoint)
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(hexKey, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return offerAddr, errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		oAddr, tx, _, err = market.DeployOffer(auth, client, big.NewInt(capacity), big.NewInt(duration), price) //提供存储容量 存储时段 存储单价
		if oAddr.String() != InvalidAddr{
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
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
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
	err = AddToMapper(offerAddr, hexKey, mapperInstance)
	if err != nil {
		return offerAddr, err
	}
	return offerAddr, nil
}

//GetOfferAddrs get all offers
func GetOfferAddrs(localAddress, ownerAddress common.Address) ([]common.Address, error) {
	//获得userIndexer, key is userAddr
	_, mapperInstance, err := GetMapperFromAdmin(localAddress, ownerAddress, offerKey, "", false)
	if err != nil {
		return nil, err
	}

	return GetAddrsFromMapper(localAddress, mapperInstance)
}

//ExtendOfferTime called by provider to extend the time in offer contract
func ExtendOfferTime(offerAddress common.Address, hexKey string, addTime *big.Int) error {
	offerInstance, err := market.NewOffer(offerAddress, GetClient(EndPoint))
	if err != nil {
		return err
	}

	log.Println("begin to extend offerTime...")
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
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
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
