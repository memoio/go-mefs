package contracts

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-mefs/contracts/upKeeping"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

//DeployUpkeeping deploy UpKeeping contracts between user, keepers and providers, and save contractAddress
func DeployUpkeeping(hexKey string, userAddress, queryAddress common.Address, keeperAddress, providerAddress []common.Address, days, size, price int64, moneyAccount *big.Int, redo bool) (common.Address, error) {
	fmt.Println("begin deploy upKeeping...")

	var ukAddr common.Address

	//获得userIndexer, key is userAddr
	_, indexerInstance, err := GetRoleIndexer(userAddress, userAddress)
	if err != nil {
		fmt.Println("GetResolverErr:", err)
		return ukAddr, err
	}
	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		fmt.Println("HexToECDSAErr:", err)
		return ukAddr, err
	}

	//获得mapper, key is upkeeping
	_, mapperInstance, err := DeployMapperToIndexer(userAddress, "upkeeping", hexKey, indexerInstance)
	if err != nil {
		return ukAddr, err
	}

	if !redo {
		ukAddr, err = getLatestFromMapper(userAddress, mapperInstance)
		if err == nil {
			return ukAddr, nil
		}
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
		ukAddress, tx, _, err := upKeeping.DeployUpKeeping(auth, client, userAddress, keeperAddress, providerAddress, big.NewInt(days), big.NewInt(size), big.NewInt(price))
		if err != nil {
			if retryCount > 5 {
				fmt.Println("deploy Uk Err:", err)
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
	err = addToMapper(userAddress, mapperInstance, ukAddr, hexKey)
	if err != nil {
		fmt.Println("add uk Err:", err)
		return ukAddr, err
	}
	fmt.Println("upKeeping-contract have been successfully deployed!")
	return ukAddr, nil
}

//GetUpkeeping get upKeeping-contract from the mapper, and get the mapper from user's indexer
func GetUpkeeping(localAddress, userAddress common.Address, key string) (ukaddr string, uk *upKeeping.UpKeeping, err error) {
	//获得userIndexer, key is userAddr
	_, indexerInstance, err := GetRoleIndexer(localAddress, userAddress)
	if err != nil {
		fmt.Println("GetResolverErr:", err)
		return InvalidAddr, uk, err
	}

	//获得mapper, key is upkeeping
	_, mapperInstance, err := getMapperFromIndexer(localAddress, "upkeeping", indexerInstance)
	if err != nil {
		return InvalidAddr, uk, err
	}

	uks, err := getAllFromMapper(localAddress, mapperInstance)
	if err != nil {
		return InvalidAddr, uk, err
	}

	client := GetClient(EndPoint)

	if key == "latest" {
		ukAddr := uks[len(uks)-1]
		uk, err := upKeeping.NewUpKeeping(ukAddr, client)
		if err != nil {
			fmt.Println("newUkErr:", err)
			return InvalidAddr, uk, err
		}
		return ukAddr.String(), uk, nil
	}

	for _, ukAddr := range uks {
		retryCount := 0
		for {
			retryCount++
			if retryCount > 10 {
				fmt.Println("GetUpkeepingInfo:", err)
				break
			}

			uk, err = upKeeping.NewUpKeeping(ukAddr, client)
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
				return ukAddr.String(), uk, nil
			}
			break
		}
	}

	return InvalidAddr, uk, err
}

// SpaceTimePay pay providers for storing data and keepers for service, hexKey is keeper's privateKey
func SpaceTimePay(ukAddr, providerAddr common.Address, hexKey string, money *big.Int) error {
	uk, err := upKeeping.NewUpKeeping(ukAddr, GetClient(EndPoint))
	if err != nil {
		fmt.Println("newUkErr:", err)
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
				fmt.Println("spaceTimePayErr:", err)
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

// GetUpkeepingInfo get Upkeeping-contract's params
func GetUpkeepingInfo(localAddress common.Address, uk *upKeeping.UpKeeping) (UpKeepingItem, error) {
	var item UpKeepingItem

	retryCount := 0
	for {
		retryCount++
		queryAddr, keeperAddrs, providerAddrs, duration, capacity, price, startTime, err := uk.GetOrder(&bind.CallOpts{
			From: localAddress,
		})
		if err != nil {
			if retryCount > 10 {
				fmt.Println("GetUpkeepingInfo:", err)
				return item, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		var keepers []string
		var providers []string
		for _, keeper := range keeperAddrs {
			kid, err := address.GetIDFromAddress(keeper.String())
			if err != nil {
				return item, err
			}
			keepers = append(keepers, kid)
		}
		for _, provider := range providerAddrs {
			pid, err := address.GetIDFromAddress(provider.String())
			if err != nil {
				return item, err
			}
			providers = append(providers, pid)
		}
		item = UpKeepingItem{
			QueryID:     queryAddr.String(),
			KeeperIDs:   keepers,
			KeeperSLA:   int32(len(keeperAddrs)),
			ProviderIDs: providers,
			ProviderSLA: int32(len(providerAddrs)),
			Duration:    duration.Int64(),
			Capacity:    capacity.Int64(),
			Price:       price.Int64(),
			StartTime:   utils.UnixToTime(startTime.Int64()).Format(utils.SHOWTIME),
		}
		break
	}

	return item, nil
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

	pid, _ := address.GetIDFromAddress(providerAddress[0].String())
	retryCount = 0
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		upItem, err := GetUpkeepingInfo(userAddress, uk)
		if err != nil {
			if retryCount > 5 {
				return err
			}
			continue
		}

		found := false
		for _, proID := range upItem.ProviderIDs {
			if proID == pid {
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
