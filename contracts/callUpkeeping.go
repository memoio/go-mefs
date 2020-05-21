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

//DeployUpkeeping deploy UpKeeping contracts between user, keepers and providers, and save contractAddress
func DeployUpkeeping(hexKey string, userAddress, queryAddress common.Address, keeperAddress, providerAddress []common.Address, duration, size int64, price *big.Int, cycle int64, moneyAccount *big.Int, redo bool) (common.Address, error) {
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
		// 用户地址,keeper地址数组,provider地址数组,存储时长 单位 s,存储大小 单位 MB
		ukAddress, tx, _, err := upKeeping.DeployUpKeeping(auth, client, queryAddress, keeperAddress, providerAddress, big.NewInt(duration), big.NewInt(size), price, big.NewInt(cycle))
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				log.Println("deploy Uk Err:", err)
				return ukAddr, err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > checkTxRetryCount {
				log.Println("deploy UK transaction Err:", err)
				return ukAddr, err
			}
			time.Sleep(time.Minute)
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
func GetUpkeepingAddrs(localAddress, userAddress common.Address) ([]common.Address, error) {
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
			queryAddr, _, _, _, _, _, _, _, _, _, _, err := uk.GetOrder(&bind.CallOpts{
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

	return ukaddr, uk, ErrEmpty
}

// SpaceTimePay pay providers for storing data and keepers for service, hexKey is keeper's privateKey
func SpaceTimePay(ukAddr, providerAddr common.Address, hexKey string, stStart, stLength, stValue *big.Int, merkleRoot [32]byte, share []int64, sign [][]byte) error {
	shareNew := []*big.Int{}
	for _, b := range share {
		shareNew = append(shareNew, big.NewInt(b))
	}

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
		auth.GasPrice = big.NewInt(spaceTimePayGasPrice)
		auth.GasLimit = spaceTimePayGasLimit
		//合约余额不足会自动报错返回
		tx, err := uk.SpaceTimePay(auth, providerAddr, stValue, stStart, stLength, merkleRoot, shareNew, sign)
		if err != nil {
			if retryCount > 2 {
				log.Println("spaceTimePayErr:", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}
		// need async check, how?
		err = CheckTx(tx)
		if err != nil {
			if retryCount > 2 {
				log.Println("spaceTimePay transaction Err:", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}
		break
	}
	return nil
}

//AddProvider add a provider to upKeeping
func AddProvider(hexKey string, localAddress, userAddress, ukAddr common.Address, providerAddress []common.Address) error {
	uk, err := upKeeping.NewUpKeeping(ukAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("newUkErr:", err)
		return err
	}

	retryCount := 0
	for {
		retryCount++
		key, _ := crypto.HexToECDSA(hexKey)
		auth := bind.NewKeyedTransactor(key)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		tx, err := uk.AddProvider(auth, providerAddress)
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > checkTxRetryCount {
				log.Println("upkeeping add provider transaction fails", err)
				return err
			}
			continue
		}
		break
	}

	retryCount = 0
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		_, _, proAddr, _, _, _, _, _, _, _, _, err := uk.GetOrder(&bind.CallOpts{
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
			if pro.Addr.String() == providerAddress[0].String() {
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

//GetOrder get queryAddr、keepers、providers、time、size、price、createDate、proofs、stEnd
func GetOrder(hexKey string, localAddress, userAddress common.Address, key string) (common.Address, []upKeeping.UpKeepingKPInfo, []upKeeping.UpKeepingKPInfo, *big.Int, *big.Int, *big.Int, *big.Int, *big.Int, *big.Int, *big.Int, []upKeeping.UpKeepingProof, error) {
	var queryAddr common.Address
	var keepers, providers []upKeeping.UpKeepingKPInfo
	var t, size, price, createDate, endDate, cycle, needPay *big.Int
	var proofs []upKeeping.UpKeepingProof
	_, uk, err := GetUpkeeping(localAddress, userAddress, key)
	if err != nil {
		return queryAddr, keepers, providers, t, size, price, createDate, endDate, cycle, needPay, proofs, err
	}

	retryCount := 0
	for {
		retryCount++
		queryAddr, keepers, providers, t, size, price, createDate, endDate, cycle, needPay, proofs, err = uk.GetOrder(&bind.CallOpts{
			From: localAddress,
		})
		if err != nil {
			if retryCount > 5 {
				return queryAddr, keepers, providers, t, size, price, createDate, endDate, cycle, needPay, proofs, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		break
	}
	return queryAddr, keepers, providers, t, size, price, createDate, endDate, cycle, needPay, proofs, nil
}

//ExtendTime user extend storage-time in upKeeping-contract
func ExtendTime(hexKey string, localAddress, userAddress common.Address, key string, addTime int64) error {
	_, uk, err := GetUpkeeping(localAddress, userAddress, key)
	if err != nil {
		return err
	}

	retryCount := 0

	for {
		retryCount++
		sk, _ := crypto.HexToECDSA(hexKey)
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		tx, err := uk.ExtendTime(auth, big.NewInt(addTime))
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > checkTxRetryCount {
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		break
	}

	return nil
}

//DestructUpKeeping destruct the upKeeping contract and transfer the balance of contract to user, anyone can call
func DestructUpKeeping(hexKey string, localAddress, userAddress common.Address, key string) error {
	_, uk, err := GetUpkeeping(localAddress, userAddress, key)
	if err != nil {
		return err
	}

	retryCount := 0

	for {
		retryCount++
		sk, _ := crypto.HexToECDSA(hexKey)
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		tx, err := uk.Destruct(auth)
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > checkTxRetryCount {
				log.Println("maybe you cannt destruct the contract now")
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		break
	}

	return nil
}

//SetKeeperStop keeper call to set keeperAddr stop
func SetKeeperStop(hexKey string, localAddress, userAddress, keeperAddr common.Address, key string, sign [][]byte) error {
	_, uk, err := GetUpkeeping(localAddress, userAddress, key)
	if err != nil {
		return err
	}

	retryCount := 0

	for {
		retryCount++
		sk, _ := crypto.HexToECDSA(hexKey)
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		tx, err := uk.SetKeeperStop(auth, keeperAddr, sign)
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				log.Println("setKeeperStop fails, err: ", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > checkTxRetryCount {
				log.Println("setKeeperStop tx fails, err: ", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		break
	}

	return nil
}

//SetProviderStop keeper call to set providerAddr stop
func SetProviderStop(hexKey string, localAddress, userAddress, providerAddr common.Address, key string, sign [][]byte) error {
	_, uk, err := GetUpkeeping(localAddress, userAddress, key)
	if err != nil {
		return err
	}

	retryCount := 0

	for {
		retryCount++
		sk, _ := crypto.HexToECDSA(hexKey)
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		tx, err := uk.SetProviderStop(auth, providerAddr, sign)
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				log.Println("setProviderStop fails, err: ", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > checkTxRetryCount {
				log.Println("setProviderStop tx fails, err: ", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		break
	}

	return nil
}
