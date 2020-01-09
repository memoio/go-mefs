package contracts

import (
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-mefs/contracts/indexer"
	"github.com/memoio/go-mefs/contracts/mapper"
	"github.com/memoio/go-mefs/contracts/resolver"
)

// GetResolverFromIndexer gets
func GetResolverFromIndexer(localAddress common.Address, key string) (common.Address, *resolver.Resolver, error) {
	var resolverAddr common.Address
	var resolverInstance *resolver.Resolver

	client := GetClient(EndPoint)
	indexerAddr := common.HexToAddress(indexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, client)
	if err != nil {
		log.Println("new Indexer err: ", err)
		return resolverAddr, resolverInstance, err
	}

	retryCount := 0
	for {
		retryCount++
		_, resolverAddr, err := indexer.Get(&bind.CallOpts{
			From: localAddress,
		}, key)
		if err != nil {
			if retryCount > 20 {
				return resolverAddr, resolverInstance, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		if len(resolverAddr) == 0 || resolverAddr.String() == InvalidAddr {
			return resolverAddr, resolverInstance, ErrNotDeployedResolver
		}
		resolverInstance, err = resolver.NewResolver(resolverAddr, client)
		if err != nil {
			return resolverAddr, resolverInstance, err
		}
		return resolverAddr, resolverInstance, nil
	}
}

func DeployResolver(localAddress common.Address, hexKey, key string) (common.Address, *resolver.Resolver, error) {
	resolverAddr, resolverInstance, err := GetResolverFromIndexer(localAddress, key)
	if err == nil {
		return resolverAddr, resolverInstance, nil
	}

	client := GetClient(EndPoint)

	indexerAddr := common.HexToAddress(indexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, client)
	if err != nil {
		log.Println("new Indexer err: ", err)
		return resolverAddr, resolverInstance, err
	}

	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSA err: ", err)
		return resolverAddr, resolverInstance, err
	}

	retryCount := 0
	for {
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		resolverAddr, _, resolverInstance, err = resolver.DeployResolver(auth, client)
		if err != nil {
			retryCount++
			if retryCount > 20 {
				log.Println("deploy Resolver Err:", err)
				return resolverAddr, resolverInstance, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		break
	}

	//将resolver地址放进indexer中,关键字key可以理解为resolverAddress的索引
	//resolver-for-channel的key为providerAddr.string()
	retryCount = 0
	for {
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		tx, err := indexer.Add(auth, key, resolverAddr)
		if err != nil {
			retryCount++
			if retryCount > 20 {
				log.Println("\naddResolverErr:", err)
				return resolverAddr, resolverInstance, err
			}
			time.Sleep(30 * time.Second)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > 20 {
				log.Println("add user indexer transaction fails", err)
				return resolverAddr, resolverInstance, err
			}
			continue
		}

		retryCount = 0
		//尝试从indexer中获取resolverAddr，以检测resolverAddr是否已放进indexer中
		for {
			retryCount++
			time.Sleep(30 * time.Second)
			_, resolverAddrGetted, err := indexer.Get(&bind.CallOpts{
				From: localAddress,
			}, key)
			if err != nil {
				if retryCount > 20 {
					log.Println("add then get Resolver Err:", err)
					return resolverAddr, resolverInstance, err
				}
				continue
			}
			if resolverAddrGetted == resolverAddr { //放进去了
				break
			}
		}
		break
	}
	return resolverAddr, resolverInstance, nil
}

func getResolverFromResolver(localAddress, ownerAddress common.Address, resolverInstance *resolver.Resolver) (common.Address, *resolver.Resolver, error) {
	retryCount := 0
	for {
		retryCount++
		resolverAddr, err := resolverInstance.Get(&bind.CallOpts{
			From: localAddress,
		}, ownerAddress)
		if err != nil {
			if retryCount > 20 {
				log.Println("get resolve Addr err: ", err)
				return resolverAddr, nil, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		if len(resolverAddr) == 0 || resolverAddr.String() == InvalidAddr {
			return resolverAddr, nil, ErrNotDeployedResolver
		}

		secondInstance, err := resolver.NewResolver(resolverAddr, GetClient(EndPoint))
		if err != nil {
			return resolverAddr, nil, err
		}
		return resolverAddr, secondInstance, nil
	}
}

func deployResolverToResolver(localAddress common.Address, resolverInstance *resolver.Resolver, hexKey string) (common.Address, *resolver.Resolver, error) {
	resolverAddr, secondInstance, err := getResolverFromResolver(localAddress, localAddress, resolverInstance)
	if err == nil {
		return resolverAddr, secondInstance, nil
	}

	client := GetClient(EndPoint)
	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSA err: ", err)
		return resolverAddr, nil, err
	}

	retryCount := 0
	for {
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		resolverAddr, _, secondInstance, err = resolver.DeployResolver(auth, client)
		if err != nil {
			retryCount++
			if retryCount > 20 {
				log.Println("deploy Resolver Err:", err)
				return resolverAddr, secondInstance, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		break
	}

	//将resolver地址放进indexer中,关键字key可以理解为resolverAddress的索引
	//resolver-for-channel的key为providerAddr.string()
	retryCount = 0
	for {
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		tx, err := resolverInstance.Add(auth, resolverAddr)
		if err != nil {
			retryCount++
			if retryCount > 20 {
				log.Println("\naddResolverErr:", err)
				return resolverAddr, secondInstance, err
			}
			time.Sleep(30 * time.Second)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > 20 {
				log.Println("add user indexer transaction fails", err)
				return resolverAddr, resolverInstance, err
			}
			continue
		}

		retryCount = 0
		//尝试从indexer中获取resolverAddr，以检测resolverAddr是否已放进indexer中
		for {
			retryCount++
			time.Sleep(30 * time.Second)
			resolverAddrGetted, err := resolverInstance.Get(&bind.CallOpts{
				From: localAddress,
			}, localAddress)
			if err != nil {
				if retryCount > 20 {
					log.Println("add then get Resolver Err:", err)
					return resolverAddr, secondInstance, err
				}
				continue
			}
			if resolverAddrGetted == resolverAddr { //放进去了
				break
			}
		}
		break
	}
	return resolverAddr, secondInstance, nil
}

func GetMapperAddrFromResolver(localAddress common.Address, ownerAddress common.Address, resolverInstance *resolver.Resolver) (common.Address, error) {
	retryCount := 0
	for {
		retryCount++
		mapperAddr, err := resolverInstance.Get(&bind.CallOpts{
			From: localAddress,
		}, ownerAddress)
		if err != nil {
			if retryCount > 20 {
				log.Println("getMapperAddrErr:", err)
				return mapperAddr, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		if len(mapperAddr) == 0 || mapperAddr.String() == InvalidAddr {
			return mapperAddr, ErrNotDeployedMapper
		}
		return mapperAddr, nil
	}
}

// getMapperFromResolver 返回已经部署的Mapper，若Mapper没部署则返回err
// 特别地，当在ChannelTimeOut()中被调用，则localAddress和ownerAddress都是userAddr；
// 当在CloseChannel()中被调用，则localAddress为providerAddr, ownerAddress为userAddr
func getMapperFromResolver(localAddress common.Address, ownerAddress common.Address, resolverInstance *resolver.Resolver) (common.Address, *mapper.Mapper, error) {
	mapperAddr, err := GetMapperAddrFromResolver(localAddress, ownerAddress, resolverInstance)
	if err != nil {
		return mapperAddr, nil, err
	}

	mapperInstance, err := mapper.NewMapper(mapperAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("newMapperErr:", err)
		return mapperAddr, nil, err
	}
	return mapperAddr, mapperInstance, nil
}

// DeployMapperToResolver 部署Mapper合约，若Mapper已经部署过，则返回已部署好的Mapper
func DeployMapperToResolver(localAddress common.Address, ownerAddress common.Address, resolverInstance *resolver.Resolver, hexKey string) (common.Address, *mapper.Mapper, error) {
	//试图从resolver中取出mapper地址：mapperAddr
	mapperAddr, mapperInstance, err := getMapperFromResolver(localAddress, ownerAddress, resolverInstance)
	if err == nil {
		return mapperAddr, mapperInstance, nil
	}

	client := GetClient(EndPoint)
	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("Hex To ECDSA err: ", err)
		return mapperAddr, mapperInstance, err
	}
	// 部署Mapper
	retryCount := 0
	for {
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		mAddr, tx, mInstance, err := mapper.DeployMapperToResolver(auth, client)
		if err != nil {
			if retryCount > 20 {
				log.Println("deployMapperErr:", err)
				return mapperAddr, mapperInstance, err
			}
			retryCount++
			time.Sleep(30 * time.Second)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			log.Println("addMapper transaction fails", err)
			if retryCount > 20 {
				return mapperAddr, mapperInstance, err
			}
			continue
		}
		mapperAddr = mAddr
		mapperInstance = mInstance
		break
	}

	//把mapper放进resolver
	retryCount = 0
	for {
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		tx, err := resolverInstance.Add(auth, mapperAddr)
		if err != nil {
			retryCount++
			if retryCount > 20 {
				return mapperAddr, mapperInstance, err
			}
			time.Sleep(30 * time.Second)
			continue
		}

		//检查交易
		err = CheckTx(tx)
		if err != nil {
			log.Println("addMapper transaction fails", err)
			if retryCount > 20 {
				return mapperAddr, mapperInstance, err
			}
			continue
		}

		retryCount = 0
		for { //验证是否放进resolver
			retryCount++
			time.Sleep(30 * time.Second)
			mapperGetted, err := resolverInstance.Get(&bind.CallOpts{
				From: localAddress,
			}, ownerAddress)
			if err != nil {
				if retryCount > 20 {
					return mapperAddr, mapperInstance, err
				}
				continue
			}
			if mapperGetted == mapperAddr {
				break
			}
		}
		break
	}

	return mapperAddr, mapperInstance, nil
}
