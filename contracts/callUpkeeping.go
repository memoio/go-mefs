package contracts

import (
	"errors"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-mefs/contracts/upKeeping"
)

const ukey = "memoriae"

//DeployUpkeeping deploy UpKeeping contracts between user, keepers and providers, and save contractAddress
func DeployUpkeeping(hexKey string, userAddress, queryAddress common.Address, keeperAddress, providerAddress []common.Address, days, size, price int64, moneyAccount *big.Int, redo bool) (common.Address, error) {
	log.Println("begin deploy upKeeping...")

	var ukAddr common.Address

	_, mapperInstance, err := GetMapperFromAdmin(userAddress, userAddress, ukey, hexKey, true)
	if err != nil {
		return ukAddr, err
	}

	if !redo {
		ukAddr, err = GetLatestFromMapper(userAddress, mapperInstance)
		if err == nil {
			return ukAddr, nil
		}
	}

	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSAErr:", err)
		return ukAddr, err
	}

	// 部署UpKeeping
	// 用户需要支付的金额
	client := GetClient(EndPoint)
	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(key)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		auth.Value = moneyAccount
		// 用户地址,keeper地址数组,provider地址数组,存储时长 单位 天,存储大小 单位 MB
		ukAddress, tx, _, err := upKeeping.DeployUpKeeping(auth, client, queryAddress, keeperAddress, providerAddress, big.NewInt(days), big.NewInt(size), big.NewInt(price))
		if err != nil {
			if retryCount > 5 {
				log.Println("deploy Uk Err:", err)
				return ukAddr, err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > 20 {
				log.Println("deploy upkeeping transaction fails", err)
				return ukAddr, err
			}
			continue
		}
		ukAddr = ukAddress
		break
	}

	//uk放进mapper
	err = AddToMapper(userAddress, ukAddr, hexKey, mapperInstance)
	if err != nil {
		log.Println("add uk Err:", err)
		return ukAddr, err
	}
	log.Println("upKeeping-contract have been successfully deployed!")
	return ukAddr, nil
}

//GetUpkeepingAddrs get all upKeeping address
func GetUpkeepingAddrs(localAddress, userAddress common.Address, key string) ([]common.Address, error) {
	//获得userIndexer, key is userAddr
	_, mapperInstance, err := GetMapperFromAdmin(localAddress, userAddress, ukey, "", false)
	if err != nil {
		return nil, err
	}

	return GetAddrsFromMapper(localAddress, mapperInstance)
}

//GetUpkeeping get upKeeping-contract from the mapper, and get the mapper from user's indexer
func GetUpkeeping(localAddress, userAddress common.Address, key string) (ukaddr common.Address, uk *upKeeping.UpKeeping, err error) {
	//获得userIndexer, key is userAddr
	_, mapperInstance, err := GetMapperFromAdmin(localAddress, userAddress, ukey, "", false)
	if err != nil {
		return ukaddr, nil, err
	}

	uks, err := GetAddrsFromMapper(localAddress, mapperInstance)
	if err != nil {
		return ukaddr, uk, err
	}

	client := GetClient(EndPoint)

	if key == "latest" {
		ukaddr = uks[len(uks)-1]
		uk, err := upKeeping.NewUpKeeping(ukaddr, client)
		if err != nil {
			log.Println("newUkErr:", err)
			return ukaddr, uk, err
		}
		return ukaddr, uk, nil
	}

	for _, ukAddr := range uks {
		ukaddr = ukAddr
		retryCount := 0
		for {
			retryCount++
			if retryCount > 10 {
				log.Println("GetUpkeepingInfo:", err)
				break
			}

			uk, err = upKeeping.NewUpKeeping(ukaddr, client)
			if err != nil {
				continue
			}
			queryAddr, _, _, _, _, _, _, err := uk.GetOrder(&bind.CallOpts{
				From: localAddress,
			})
			if err != nil {
				time.Sleep(60 * time.Second)
				continue
			}

			if queryAddr.String() == key {
				return ukaddr, uk, nil
			}
			break
		}
	}

	return ukaddr, uk, errors.New("No upkeeping")
}

// SpaceTimePay pay providers for storing data and keepers for service, hexKey is keeper's privateKey
func SpaceTimePay(ukAddr, providerAddr common.Address, hexKey string, money *big.Int) error {
	uk, err := upKeeping.NewUpKeeping(ukAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("newUkErr:", err)
		return err
	}

	key, _ := crypto.HexToECDSA(hexKey)
	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(key)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		auth.GasLimit = spaceTimePayGasLimit
		//合约余额不足会自动报错返回
		_, err := uk.SpaceTimePay(auth, providerAddr, money)
		if err != nil {
			if retryCount > 5 {
				log.Println("spaceTimePayErr:", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}
		// need async check, how?
		break
	}
	return nil
}

//AddProvider add a provider to upKeeping
func AddProvider(hexKey string, localAddress, userAddress common.Address, providerAddress []common.Address, key string) error {
	_, uk, err := GetUpkeeping(localAddress, userAddress, key)
	if err != nil {
		return err
	}

	retryCount := 0
	for {
		retryCount++
		key, _ := crypto.HexToECDSA(hexKey)
		auth := bind.NewKeyedTransactor(key)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		_, err = uk.AddProvider(auth, providerAddress)
		if err != nil {
			if retryCount > 5 {
				return err
			}
			time.Sleep(time.Minute)
			continue
		}
		break
	}

	retryCount = 0
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		_, _, proAddr, _, _, _, _, err := uk.GetOrder(&bind.CallOpts{
			From: localAddress,
		})
		if err != nil {
			if retryCount > 5 {
				return err
			}
			continue
		}

		found := false
		for _, pro := range proAddr {
			if pro.String() == providerAddress[0].String() {
				found = true
				break
			}
		}

		if found {
			break
		} else {
			if retryCount > 20 {
				return errors.New("upkeeping add provider fails")
			}
		}
	}
	return nil
}
