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

//DeployOffer provider use it to deploy offer-contract
func DeployOffer(localAddress common.Address, hexKey string, capacity int64, duration int64, price int64, redo bool) (common.Address, error) {
	log.Println("begin to deploy offer-contract...")
	var offerAddr common.Address

	_, mapperInstance, err := GetMapperFromAdmin(localAddress, localAddress, "offer", hexKey, true)
	if err != nil {
		return offerAddr, err
	}

	if !redo {
		offerAddr, err = GetLatestFromMapper(localAddress, mapperInstance)
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
	err = AddToMapper(localAddress, offerAddr, hexKey, mapperInstance)
	if err != nil {
		return offerAddr, err
	}

	log.Println("offer-contract have been successfully deployed!")
	return offerAddr, nil
}

//GetOfferAddrs get all offers
func GetOfferAddrs(localAddress, ownerAddress common.Address) ([]common.Address, error) {
	//获得userIndexer, key is userAddr
	_, mapperInstance, err := GetMapperFromAdmin(localAddress, ownerAddress, "offer", "", false)
	if err != nil {
		return nil, err
	}

	return GetAddrsFromMapper(localAddress, mapperInstance)
}

//GetLatestOffer get latest query
func GetLatestOffer(localAddress, userAddress common.Address) (offerAddr common.Address, offerInstance *market.Offer, err error) {
	//获得userIndexer, key is userAddr
	_, mapperInstance, err := GetMapperFromAdmin(localAddress, userAddress, "offer", "", false)
	if err != nil {
		return offerAddr, nil, err
	}

	offers, err := GetAddrsFromMapper(localAddress, mapperInstance)
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
