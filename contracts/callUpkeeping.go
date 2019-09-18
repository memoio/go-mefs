package contracts

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-mefs/contracts/upKeeping"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

//DeployUpkeeping deploy UpKeeping contracts between user, keepers and providers, and save contractAddress in mapper
func DeployUpkeeping(hexKey string, userAddress common.Address, keeperAddress []common.Address, providerAddress []common.Address, days int64, size int64, price int64, moneyAccount *big.Int) error {
	fmt.Println("begin deploy upKeeping...")

	var ukAddr common.Address

	//获得resolver
	_, resolver, err := getResolverFromIndexer(userAddress, "memoriae")
	if err != nil {
		fmt.Println("getResolverErr:", err)
		return err
	}
	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		fmt.Println("HexToECDSAErr:", err)
		return err
	}

	//获得mapper
	_, mapperInstance, err := deployMapper(userAddress, userAddress, resolver, hexKey)
	if err != nil {
		return err
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
		ukAddr, _, _, err = upKeeping.DeployUpKeeping(auth, client, userAddress, keeperAddress, providerAddress, big.NewInt(days), big.NewInt(size), big.NewInt(price))
		if err != nil {
			if retryCount > 5 {
				fmt.Println("deploy Uk Err:", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}
		break
	}

	//uk放进mapper
	err = addToMapper(userAddress, mapperInstance, ukAddr, hexKey)
	if err != nil {
		fmt.Println("add uk Err:", err)
		return err
	}
	fmt.Println("upKeeping-contract have been successfully deployed!")
	return nil
}

//GetUKFromResolver get upKeeping-contract from the mapper, and get the mapper from the resolver
func GetUKFromResolver(localAddress common.Address) (ukaddr string, uk *upKeeping.UpKeeping, err error) {
	//获得resolver
	_, resolverInstance, err := getResolverFromIndexer(localAddress, "memoriae")
	if err != nil {
		fmt.Println("getResolverErr:", err)
		return InvalidAddr, uk, err
	}

	// 获得mapper
	_, mapperInstance, err := getMapperInstance(localAddress, localAddress, resolverInstance)
	if err != nil {
		fmt.Println("get Mapper Instance err:", err)
		return InvalidAddr, uk, err
	}

	// 获得mapper中的合约
	ukAddr, err := getLatestAddrFromMapper(localAddress, mapperInstance)
	if err != nil {
		return InvalidAddr, uk, err
	}
	//获得uk，暂时默认第一个是所需的uk合约地址
	//TODO：优化从mapper中找出uk合约的方法
	uk, err = upKeeping.NewUpKeeping(ukAddr, GetClient(EndPoint))
	if err != nil {
		fmt.Println("newUkErr:", err)
		return InvalidAddr, uk, err
	}
	return ukAddr.String(), uk, nil
}

// SpaceTimePay pay providers for storing data and keepers for service, hexKey is keeper's privateKey
func SpaceTimePay(uk *upKeeping.UpKeeping, userAddress common.Address, providerAddr common.Address, hexKey string, money *big.Int) error {
	//构建auth,用keeper的私钥
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
		break
	}
	return nil
}

// GetUpkeepingInfo get Upkeeping-contract's params
func GetUpkeepingInfo(localAddress common.Address, uk *upKeeping.UpKeeping) (
	UpKeepingItem, error) {
	var item UpKeepingItem

	retryCount := 0
	for {
		retryCount++
		_, keeperAddrs, providerAddrs, duration, capacity, price, startTime, err := uk.GetOrder(&bind.CallOpts{
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
func AddProvider(hexKey string, userAddress common.Address, providerAddress []common.Address) error {
	_, uk, err := GetUKFromResolver(userAddress)
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
	return nil
}
