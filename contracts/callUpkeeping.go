package contracts

import (
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

//DeployUpkeeping deploy UpKeeping contracts between user, keepers and providers, and save contractAddress in mapper
func DeployUpkeeping(endPoint string, hexKey string, userAddress common.Address, keeperAddress []common.Address, providerAddress []common.Address, days int64, size int64, price int64, moneyAccount *big.Int) error {
	fmt.Println("begin deploy upKeeping...")
	//获得resolver
	resolver, err := getResolverFromIndexer(endPoint, userAddress, "memoriae")
	if err != nil {
		fmt.Println("getResolverErr:", err)
		return err
	}
	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		fmt.Println("HexToECDSAErr:", err)
		return err
	}
	auth := bind.NewKeyedTransactor(key)
	client := GetClient(endPoint)

	//获得mapper
	mapper, err := deployMapper(endPoint, userAddress, resolver, auth, client)
	if err != nil {
		return err
	}

	// 部署UpKeeping
	// 用户需要支付的金额
	auth = bind.NewKeyedTransactor(key)
	auth.Value = moneyAccount
	ukAddr, _, _, err := upKeeping.DeployUpKeeping(auth, client,
		// 用户地址,keeper地址数组,provider地址数组,存储时长 单位 天,存储大小 单位 MB
		userAddress, keeperAddress, providerAddress, big.NewInt(days), big.NewInt(size), big.NewInt(price))
	if err != nil {
		fmt.Println("deployUkErr:", err)
		return err
	}
	log.Println("ukAddr:", ukAddr.String())

	//uk放进mapper
	auth = bind.NewKeyedTransactor(key)
	for addToMapperCount := 0; addToMapperCount < 2; addToMapperCount++ {
		time.Sleep(10 * time.Second)
		_, err = mapper.Add(auth, ukAddr)
		if err == nil {
			break
		}
	}
	if err != nil {
		fmt.Println("addukErr:", err)
		return err
	}

	//尝试从mapper中获取ukAddr，以检测ukAddr是否已放进mapper中
	var contracts []common.Address
	for i := 0; i < 30; i++ {
		contracts, err = mapper.Get(&bind.CallOpts{
			From: userAddress,
		})
		if err != nil {
			fmt.Println("getContractsErr:", err)
			return err
		}
		if len(contracts) == 0 || contracts[0].String() == InvalidAddr { //ukAddr还没放进mapper
			time.Sleep(10 * time.Second)
		} else {
			break
		}
	}
	if len(contracts) == 0 || contracts[0].String() == InvalidAddr {
		fmt.Println("upKeeping-contract have not been put to mapper!")
		return ErrContractNotPutToMapper
	}
	fmt.Println("upKeeping-contract have been successfully deployed!")
	return nil
}

//GetUKFromResolver get upKeeping-contract from the mapper, and get the mapper from the resolver
func GetUKFromResolver(endPoint string, ownerAddress common.Address) (ukaddr string, uk *upKeeping.UpKeeping, err error) {
	//获得resolver
	resolver, err := getResolverFromIndexer(endPoint, ownerAddress, "memoriae")
	if err != nil {
		fmt.Println("getResolverErr:", err)
		return InvalidAddr, uk, err
	}

	//获得mapper
	mapper, err := getMapperInstance(endPoint, ownerAddress, ownerAddress, resolver)
	if err != nil {
		fmt.Println("getMapperInstance err:", err)
		return InvalidAddr, uk, err
	}

	//获得mapper中的合约
	contracts, err := mapper.Get(&bind.CallOpts{
		From: ownerAddress,
	})
	if err != nil {
		fmt.Println("getContractsErr:", err)
		return InvalidAddr, uk, err
	}
	if len(contracts) == 0 || contracts[0].String() == InvalidAddr {
		fmt.Println("getContractsErr:", ErrNotDeployedUk)
		return InvalidAddr, uk, ErrNotDeployedUk
	}

	//获得uk，暂时默认第一个是所需的uk合约地址
	//TODO：优化从mapper中找出uk合约的方法
	uk, err = upKeeping.NewUpKeeping(contracts[0], GetClient(endPoint))
	if err != nil {
		fmt.Println("newUkErr:", err)
		return InvalidAddr, uk, err
	}
	ukaddr = contracts[0].String()
	return ukaddr, uk, nil
}

//SpaceTimePay pay providers for storing data and keepers for service, hexKey is keeper's privateKey
func SpaceTimePay(uk *upKeeping.UpKeeping, endPoint string, userAddress common.Address, providerAddr common.Address, hexKey string, money *big.Int) error {
	//构建auth,用keeper的私钥
	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(spaceTimePayGasPrice)
	auth.GasLimit = spaceTimePayGasLimit
	tran, err := uk.SpaceTimePay(auth, providerAddr, money) //合约余额不足会自动报错返回

	if err != nil {
		fmt.Println("spaceTimePayErr:", err)
		tranM, _ := tran.MarshalJSON()
		fmt.Println("tran,", string(tranM))
		return err
	}
	fmt.Println("tran.hash:", tran.Hash().Hex())
	return nil
}

// GetUpkeepingInfo get Upkeeping-contract's params
func GetUpkeepingInfo(endPoint string, localAddress common.Address, uk *upKeeping.UpKeeping) (
	UpKeepingItem, error) {
	var item UpKeepingItem
	_, keeperAddrs, providerAddrs, duration, capacity, price, time, err := uk.GetOrder(&bind.CallOpts{
		From: localAddress,
	})
	if err != nil {
		fmt.Println("getOfferParamsErr:", err)
		return item, err
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
		StartTime:   utils.UnixToTime(time.Int64()).Format("2006-01-02 15:04:05"),
	}
	return item, nil
}

//AddProvider add a provider to upKeeping
func AddProvider(endPoint string, hexKey string, userAddress common.Address, providerAddress []common.Address) error {
	_, uk, err := GetUKFromResolver(endPoint, userAddress)
	if err != nil {
		return err
	}

	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)

	_, err = uk.AddProvider(auth, providerAddress)
	if err != nil {
		return err
	}
	return nil
}
