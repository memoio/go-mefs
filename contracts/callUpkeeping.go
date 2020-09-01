package contracts

import (
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/memoio/go-mefs/contracts/upKeeping"
)

//DeployUpkeeping deploy UpKeeping contracts between user, keepers and providers, and save contractAddress
func DeployUpkeeping(hexKey string, userAddress, queryAddress common.Address, keeperAddress, providerAddress []common.Address, duration, size int64, price *big.Int, cycle int64, moneyAccount *big.Int, redo bool) (common.Address, error) {
	var ukAddr, ukAddress common.Address

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

	log.Println("begin deploy upKeeping...")
	client := GetClient(EndPoint)
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(hexKey, moneyAccount, nil, big.NewInt(defaultGasPrice), 0)
		if errMA != nil {
			return ukAddr, errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		// 用户地址,keeper地址数组,provider地址数组,存储时长 单位 s,存储大小 单位 MB
		ukAddress, tx, _, err = upKeeping.DeployUpKeeping(auth, client, queryAddress, keeperAddress, providerAddress, big.NewInt(duration), big.NewInt(size), price, big.NewInt(cycle))
		if ukAddress.String() != InvalidAddr {
			ukAddr = ukAddress
		}
		if err != nil {
			retryCount++
			log.Println("deploy UpKeeping Err:", err)
			if err.Error() == core.ErrNonceTooLow.Error() && auth.GasPrice.Cmp(big.NewInt(defaultGasPrice)) > 0 {
				log.Println("previously pending transaction has successfully executed")
				break
			}
			if retryCount > sendTransactionRetryCount {
				return ukAddr, err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("deploy UpKeeping transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return ukAddr, err
			}
			continue
		}
		break
	}
	log.Println("upKeeping-contract", ukAddr.String(), "have been successfully deployed!")

	//uk放进mapper
	err = AddToMapper(ukAddr, hexKey, mapperInstance)
	if err != nil {
		log.Println("add uk Err:", err)
		return ukAddr, err
	}
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
	uks, err := GetUpkeepingAddrs(localAddress, userAddress)
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

	log.Println("begin call spaceTimePay...")
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(hexKey, nil, nil, big.NewInt(spaceTimePayGasPrice), spaceTimePayGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		//合约余额不足会自动报错返回
		tx, err = uk.SpaceTimePay(auth, providerAddr, stValue, stStart, stLength, merkleRoot, shareNew, sign)
		if err != nil {
			retryCount++
			log.Println("spaceTimePay Err:", err)
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
			log.Println("spaceTimePay transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			time.Sleep(time.Minute)
			continue
		}
		break
	}
	log.Println("spaceTimePay have been successfully called!")
	return nil
}

//AddProvider add a provider to upKeeping
func AddProvider(hexKey string, localAddress, ukAddr common.Address, providerAddress []common.Address, sign [][]byte) error {
	uk, err := upKeeping.NewUpKeeping(ukAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("newUkErr:", err)
		return err
	}

	log.Println("begin add provider...")
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

		//合约余额不足会自动报错返回
		tx, err = uk.AddProvider(auth, providerAddress, sign)
		if err != nil {
			retryCount++
			log.Println("addProvider Err:", err)
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
			log.Println("addProvider transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}

	log.Println("provider have been successfully added to upKeeping!")
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

	log.Println("begin extend upkeeping time...")
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

		//合约余额不足会自动报错返回
		tx, err = uk.ExtendTime(auth, big.NewInt(addTime))
		if err != nil {
			retryCount++
			log.Println("extendUKTime Err:", err)
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
			log.Println("extendUKTime transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}

	log.Println("UKTime have been successfully extended!")
	return nil
}

//DestructUpKeeping destruct the upKeeping contract and transfer the balance of contract to user, anyone can call
func DestructUpKeeping(hexKey string, localAddress, userAddress common.Address, key string) error {
	_, uk, err := GetUpkeeping(localAddress, userAddress, key)
	if err != nil {
		return err
	}

	log.Println("begin destruct upkeeping...")
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

		tx, err = uk.Destruct(auth)
		if err != nil {
			retryCount++
			log.Println("destruct UK Err:", err)
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
			log.Println("destruct UK transaction fails", err, "maybe you cannot destruct UK now")
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}

	log.Println("UK have been successfully destruct!")
	return nil
}

//SetKeeperStop keeper call to set keeperAddr stop
func SetKeeperStop(hexKey string, localAddress, userAddress, keeperAddr common.Address, key string, sign [][]byte) error {
	_, uk, err := GetUpkeeping(localAddress, userAddress, key)
	if err != nil {
		return err
	}

	log.Println("begin set keeper in upkeeping stop...")
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(hexKey, nil, nil, big.NewInt(spaceTimePayGasPrice), defaultGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = uk.SetKeeperStop(auth, keeperAddr, sign)
		if err != nil {
			retryCount++
			log.Println("set keeper stop Err:", err)
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
			log.Println("set keeper stop transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}

	log.Println("keeper have been successfully set stop!")
	return nil
}

//SetProviderStop keeper call to set providerAddr stop
func SetProviderStop(hexKey string, localAddress, userAddress, providerAddr common.Address, key string, sign [][]byte) error {
	_, uk, err := GetUpkeeping(localAddress, userAddress, key)
	if err != nil {
		return err
	}

	log.Println("begin set provider in upkeeping stop...")
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(hexKey, nil, nil, big.NewInt(spaceTimePayGasPrice), defaultGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = uk.SetProviderStop(auth, providerAddr, sign)
		if err != nil {
			retryCount++
			log.Println("set provider stop Err:", err)
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
			log.Println("set provider stop transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}

	log.Println("provider have been successfully set stop!")
	return nil
}
