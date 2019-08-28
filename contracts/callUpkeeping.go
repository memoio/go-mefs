package contracts

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/memoio/go-mefs/contracts/upKeeping"
	"github.com/memoio/go-mefs/utils"
)

//GetClient get rpc-client based the endPoint
func GetClient(endPoint string) *ethclient.Client {
	client, err := rpc.Dial(endPoint)
	if err != nil {
		fmt.Println(err)
	}
	return ethclient.NewClient(client)
}

func getResolverFromIndexer(endPoint string, localAddress common.Address, key string) (resolver *upKeeping.Resolver, err error) {
	indexerAddr := common.HexToAddress(IndexerHex)
	indexer, err := upKeeping.NewIndexer(indexerAddr, GetClient(endPoint))
	if err != nil {
		fmt.Println("newIndexerErr:", err)
		return resolver, err
	}

	_, resolverAddr, err := indexer.Get(&bind.CallOpts{
		From: localAddress,
	}, key)
	if err != nil {
		fmt.Println("getResolverErr:", err)
		return resolver, err
	}
	if resolverAddr.String() == InvalidAddr {
		fmt.Println("getResolverErr:", ErrNotDeployedResolver)
		return resolver, ErrNotDeployedResolver
	}

	resolver, err = upKeeping.NewResolver(resolverAddr, GetClient(endPoint))
	if err != nil {
		fmt.Println(err)
		return resolver, err
	}
	return resolver, nil
}

func getMapper(endPoint string, userAddress common.Address, resolver *upKeeping.Resolver, auth *bind.TransactOpts, client *ethclient.Client) (mapper *upKeeping.Mapper, err error) {
	//试图从resolver中取出mapper地址：mapperAddr
	var mapperAddr common.Address
	mapperAddr, err = resolver.Get(&bind.CallOpts{
		From: userAddress,
	}, userAddress)
	if err != nil {
		fmt.Println("getMapperErr:", err)
		return mapper, err
	}

	if len(mapperAddr) == 0 || mapperAddr.String() == InvalidAddr { //没有部署过mapper
		// 部署Mapper
		mapperAddr, _, mapper, err = upKeeping.DeployMapper(auth, client)
		if err != nil {
			fmt.Println("deployMapperErr:", err)
			return mapper, err
		}
		log.Println("mapperAddr:", mapperAddr.String())

		//把mapper放进resolver
		_, err = resolver.Add(auth, mapperAddr)
		if err != nil {
			fmt.Println("addMapperErr:", err)
			return mapper, err
		}
		for { //验证是否放进resolver
			mapperGetted, err := resolver.Get(&bind.CallOpts{
				From: userAddress,
			}, userAddress)
			if err != nil {
				fmt.Println("getMapperErr:", err)
				return mapper, err
			}
			if mapperGetted == mapperAddr {
				break
			}
			time.Sleep(8 * time.Second)
		}
	} else { //部署过mapper，直接根据mapperAddr获得mapper
		mapper, err = upKeeping.NewMapper(mapperAddr, GetClient(endPoint))
		if err != nil {
			fmt.Println("newMapperErr:", err)
			return mapper, err
		}
	}
	return mapper, nil
}

//Deploy deploy UpKeeping contracts between user, keepers and providers, and save contractAddress in mapper
func Deploy(endPoint string, hexKey string, userAddress common.Address, keeperAddress []common.Address, providerAddress []common.Address, days int64, size int64, price int64, moneyAccount *big.Int) error {
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
	mapper, err := getMapper(endPoint, userAddress, resolver, auth, client)
	if err != nil {
		return err
	}

	// 部署UpKeeping
	// 用户需要支付的金额
	auth = bind.NewKeyedTransactor(key)
	auth.Value = moneyAccount
	ukAddr, _, _, err := upKeeping.DeployUpKeeping(auth, client,
		// 用户地址,keeper地址数组,provider地址数组,存储时长 单位 天,存储大小 单位 MB
		userAddress, keeperAddress, providerAddress, big.NewInt(days), big.NewInt(size))
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
	mapperAddr, err := resolver.Get(&bind.CallOpts{
		From: ownerAddress,
	}, ownerAddress)
	if err != nil {
		fmt.Println("getMapperAddrErr:", err)
		return InvalidAddr, uk, err
	}
	if mapperAddr.String() == InvalidAddr {
		fmt.Println("getMapperAddrErr:", ErrNotDeployedMapper)
		return InvalidAddr, uk, ErrNotDeployedMapper
	}
	mapper, err := upKeeping.NewMapper(mapperAddr, GetClient(endPoint))
	if err != nil {
		fmt.Println("newMapperErr:", err)
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

//QueryBalance query the balance of account
func QueryBalance(endPoint string, account string) (balance *big.Int, err error) {
	var result string
	client, err := rpc.Dial(endPoint)
	if err != nil {
		fmt.Println("rpc.dial err:", err)
		return balance, err
	}
	err = client.Call(&result, "eth_getBalance", account, "latest")
	if err != nil {
		fmt.Println("client.call err:", err)
		return balance, err
	}
	balance = utils.HexToBigInt(result)
	return balance, nil
}

//DeployResolver provider deploys resolver to save mapper
func DeployResolver(endPoint string, hexKey string, localAddress common.Address, indexer *upKeeping.Indexer) (err error) {
	fmt.Println("begin deploy resolver...")
	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	client := GetClient(endPoint)

	//查看是否已经部署过
	_, resolverAddrGetted, err := indexer.Get(&bind.CallOpts{
		From: localAddress,
	}, localAddress.String())
	if err != nil {
		fmt.Println("getResolverErr:", err)
		return err
	}
	if resolverAddrGetted.String() != InvalidAddr { //说明部署过
		log.Println("you have deployed resolver already")
		return nil
	}

	//provider部署resolver
	auth = bind.NewKeyedTransactor(key)
	resolverAddr, _, _, err := upKeeping.DeployResolver(auth, client)
	if err != nil {
		fmt.Println("deployResolverErr:", err)
		return err
	}
	log.Println("resolverAddr:", resolverAddr.String())

	//将resolver地址放进indexer中,关键字key为provider的地址
	fmt.Print("wait for resolverAddr added into indexer...")
	auth = bind.NewKeyedTransactor(key)
	_, err = indexer.Add(auth, localAddress.String(), resolverAddr)
	if err != nil {
		fmt.Println("\naddResolverErr:", err)
		return err
	}

	//尝试从indexer中获取resolverAddr，以检测resolverAddr是否已放进indexer中
	for {
		_, resolverAddrGetted, err = indexer.Get(&bind.CallOpts{
			From: localAddress,
		}, localAddress.String())
		if err != nil {
			fmt.Println("\ngetContractsErr:", err)
			return err
		}
		if resolverAddrGetted == resolverAddr { //放进去了
			fmt.Println("done!")
			break
		}
		time.Sleep(10 * time.Second)
	}

	fmt.Println("resolver have been successfully deployed!")
	return nil
}

// GetUpKeepingParams get Upkeeping-contract's params
func GetUpKeepingParams(endPoint string, localAddress, userAddress common.Address) (
	common.Address, []common.Address, []common.Address, int64, int64, int64, error) {
	var userAddr common.Address
	_, uk, err := GetUKFromResolver(endPoint, userAddress)
	if err != nil {
		return userAddr, nil, nil, 0, 0, 0, err
	}
	userAddr, keeperAddrs, providerAddrs, duration, capacity, moneyAccount, err := uk.GetOrder(&bind.CallOpts{
		From: localAddress,
	})
	if err != nil {
		fmt.Println("getOfferParamsErr:", err)
		return userAddr, nil, nil, 0, 0, 0, err
	}
	var price = new(big.Int)
	var moneyPerDay = new(big.Int)
	moneyPerDay = moneyPerDay.Quo(moneyAccount, duration)
	price = price.Quo(moneyPerDay, capacity)
	return userAddr, keeperAddrs, providerAddrs, duration.Int64(), capacity.Int64(), price.Int64(), nil
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
