package contracts

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-mefs/contracts/indexer"
	"github.com/memoio/go-mefs/contracts/role"
)

func getKeeperContractFromIndexer(localAddress common.Address) (keeperContract *role.Keeper, err error) {
	indexerAddr := common.HexToAddress(IndexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, GetClient(EndPoint))
	if err != nil {
		fmt.Println("newIndexerErr:", err)
		return keeperContract, err
	}

	_, keeperContractAddr, err := indexer.Get(&bind.CallOpts{
		From: localAddress,
	}, "keeper")
	if err != nil {
		fmt.Println("getkeeperContractErr:", err)
		return keeperContract, err
	}

	keeperContract, err = role.NewKeeper(keeperContractAddr, GetClient(EndPoint))
	if err != nil {
		fmt.Println("newKeeperErr:", err)
		return keeperContract, err
	}
	return keeperContract, nil
}

//IsKeeper judge if an account is keeper
func IsKeeper(localAddress common.Address) (bool, error) {
	keeperContract, err := getKeeperContractFromIndexer(localAddress)
	if err != nil {
		fmt.Println("keeperContracterr:", err)
		return false, err
	}
	isKeeper, err := keeperContract.IsKeeper(&bind.CallOpts{
		From: localAddress,
	}, localAddress)
	if err != nil {
		fmt.Println("isKeepererr:", err)
		return false, err
	}
	return isKeeper, nil
}

func getProviderContractFromIndexer(localAddress common.Address) (providerContract *role.Provider, err error) {
	indexerAddr := common.HexToAddress(IndexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, GetClient(EndPoint))
	if err != nil {
		fmt.Println("newIndexerErr:", err)
		return providerContract, err
	}

	_, providerContractAddr, err := indexer.Get(&bind.CallOpts{
		From: localAddress,
	}, "provider")
	if err != nil {
		fmt.Println("getproviderContractErr:", err)
		return providerContract, err
	}

	providerContract, err = role.NewProvider(providerContractAddr, GetClient(EndPoint))
	if err != nil {
		fmt.Println(err)
		return providerContract, err
	}
	return providerContract, nil
}

//IsProvider judge if an account is provider
func IsProvider(localaddress common.Address) (bool, error) {
	providerContract, err := getProviderContractFromIndexer(localaddress)
	if err != nil {
		fmt.Println("providerContracterr:", err)
		return false, err
	}
	isProvider, err := providerContract.IsProvider(&bind.CallOpts{
		From: localaddress,
	}, localaddress)
	if err != nil {
		fmt.Println("isKeepererr:", err)
		return false, err
	}
	return isProvider, nil
}

//KeeperContract deploy a keeper contract
func KeeperContract(hexKey string) (err error) {
	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	client := GetClient(EndPoint)

	//暂时将质押金额设为0
	deposit := big.NewInt(0)
	keeperContractAddr, _, _, err := role.DeployKeeper(auth, client, deposit)
	if err != nil {
		fmt.Println("deployKeeperErr:", err)
		return err
	}
	log.Println("keeperContractAddr:", keeperContractAddr.String())

	indexerAddr := common.HexToAddress(IndexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, GetClient(EndPoint))
	if err != nil {
		fmt.Println("newIndexerErr:", err)
		return err
	}

	indexer.Add(auth, "keeper", keeperContractAddr)
	return nil
}

//SetKeeper set "localAddress" keeper in contract if isKeeper is true
func SetKeeper(localAddress common.Address, hexKey string, isKeeper bool) (err error) {
	keeper, err := getKeeperContractFromIndexer(localAddress)
	if err != nil {
		fmt.Println("keeperContracterr:", err)
		return err
	}

	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)

	_, err = keeper.Set(auth, localAddress, isKeeper)
	if err != nil {
		return err
	}
	return nil
}

//ProviderContract deploy a keeper contract
func ProviderContract(hexKey string) (err error) {
	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	client := GetClient(EndPoint)

	//暂时将存储容量、质押金额设为0
	size := big.NewInt(0)
	deposit := big.NewInt(0)
	providerContractAddr, _, _, err := role.DeployProvider(auth, client, size, deposit)
	if err != nil {
		fmt.Println("deployProviderErr:", err)
		return err
	}
	log.Println("providerContractAddr:", providerContractAddr.String())

	indexerAddr := common.HexToAddress(IndexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, GetClient(EndPoint))
	if err != nil {
		fmt.Println("newIndexerErr:", err)
		return err
	}

	indexer.Add(auth, "provider", providerContractAddr)
	return nil
}

//SetProvider set "localAddress" provider in contract if isProvider is true
func SetProvider(localAddress common.Address, hexKey string, isProvider bool) (err error) {
	provider, err := getProviderContractFromIndexer(localAddress)
	if err != nil {
		return err
	}

	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	_, err = provider.Set(auth, localAddress, isProvider)
	if err != nil {
		return err
	}
	return nil
}

//DeployKeeperProviderMap deploy KeeperProviderMap-contract
func DeployKeeperProviderMap(hexKey string, localAddress common.Address) (*role.KeeperProviderMap, error) {
	fmt.Println("begin deploy keeperProviderMap...")

	var keeperProviderMapInstance *role.KeeperProviderMap

	//获得resolver
	resolver, err := getResolverFromIndexer(localAddress, "keeperProviderMap")
	if err != nil {
		fmt.Println("getResolverErr:", err)
		return nil, err
	}

	//获得mapper
	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		fmt.Println("HexToECDSAErr:", err)
		return nil, err
	}
	auth := bind.NewKeyedTransactor(key)
	client := GetClient(EndPoint)
	mapper, err := deployMapper(localAddress, resolver, auth, client)
	if err != nil {
		return nil, err
	}

	//查看是否已经部署过keeperProviderMap，如果部署过就直接返回
	keeperProviderMapAddressesGetted, err := mapper.Get(&bind.CallOpts{
		From: localAddress,
	})
	if err != nil {
		fmt.Println("getOfferAddressesErr:", err)
		return nil, err
	}
	if len(keeperProviderMapAddressesGetted) != 0 && keeperProviderMapAddressesGetted[0].String() != InvalidAddr { //代表用户之前就部署过keeperProviderMap
		fmt.Println("you have deployed keeperProviderMap already")
		keeperProviderMapInstance, err = role.NewKeeperProviderMap(keeperProviderMapAddressesGetted[0], client)
		if err != nil {
			fmt.Println("newKeeperProviderMapInstanceErr:", err)
			return nil, ErrNewContractInstance
		}
		return keeperProviderMapInstance, nil
	}

	//之前没有部署过，部署keeperProviderMap合约
	auth = bind.NewKeyedTransactor(key)
	keeperProviderMapAddr, _, keeperProviderMapInstance, err := role.DeployKeeperProviderMap(auth, client)
	if err != nil {
		fmt.Println("deployKeeperProviderMapErr:", err)
		return nil, err
	}

	//keeperProviderMap放进mapper
	auth = bind.NewKeyedTransactor(key)
	for addToMapperCount := 0; addToMapperCount < 2; addToMapperCount++ {
		time.Sleep(10 * time.Second)
		_, err = mapper.Add(auth, keeperProviderMapAddr)
		if err == nil {
			break
		}
	}
	if err != nil {
		fmt.Println("addukErr:", err)
		return nil, err
	}
	//尝试从mapper中获取keeperProviderMap，以检测keeperProviderMap是否已放进mapper中
	var contracts []common.Address
	for i := 0; i < 30; i++ {
		contracts, err = mapper.Get(&bind.CallOpts{
			From: localAddress,
		})
		if err != nil {
			fmt.Println("getContractsErr:", err)
			return nil, err
		}
		if len(contracts) == 0 || contracts[0].String() == InvalidAddr { //ukAddr还没放进mapper
			time.Sleep(10 * time.Second)
		} else {
			break
		}
	}
	if len(contracts) == 0 || contracts[0].String() == InvalidAddr {
		fmt.Println("keeperProviderMap-contract have not been put to mapper!")
		return nil, ErrContractNotPutToMapper
	}
	fmt.Println("keeperProviderMap-contract have been successfully deployed!")
	return keeperProviderMapInstance, nil
}

func getKeeperProviderMapInstanceFromIndexer(localAddress common.Address) (*role.KeeperProviderMap, error) {
	var keeperProviderInstance *role.KeeperProviderMap

	indexerAddr := common.HexToAddress(IndexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, GetClient(EndPoint))
	if err != nil {
		fmt.Println("newIndexerErr:", err)
		return keeperProviderInstance, err
	}

	_, keeperproviderMapContractAddr, err := indexer.Get(&bind.CallOpts{
		From: localAddress,
	}, "keeperProviderMap")
	if err != nil {
		fmt.Println("getkeeperproviderMapContractErr:", err)
		return keeperProviderInstance, err
	}

	keeperProviderInstance, err = role.NewKeeperProviderMap(keeperproviderMapContractAddr, GetClient(EndPoint))
	if err != nil {
		fmt.Println("newKeeperProviderMapContractErr:", err)
		return keeperProviderInstance, err
	}
	return keeperProviderInstance, nil
}

func addKeeperProvidersToKPMap(localAddress common.Address, hexKey string, keeperAddress common.Address, providerAddresses []common.Address) error {
	keeperProviderMapInstance, err := getKeeperProviderMapInstanceFromIndexer(localAddress)
	if err != nil {
		fmt.Println("getKeeperProviderMapInstanceErr:", err)
		return err
	}

	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		fmt.Println("HexToECDSAErr:", err)
		return err
	}
	auth := bind.NewKeyedTransactor(key)

	_, err = keeperProviderMapInstance.Add(auth, keeperAddress, providerAddresses)
	if err != nil {
		fmt.Println("addKeeperProviderTokpMapErr:", err)
		return err
	}
	return nil
}

func deleteKeeper(localAddress common.Address, hexKey string, keeperAddress common.Address) error {
	keeperProviderMapInstance, err := getKeeperProviderMapInstanceFromIndexer(localAddress)
	if err != nil {
		fmt.Println("getKeeperProviderMapInstanceErr:", err)
		return err
	}

	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		fmt.Println("HexToECDSAErr:", err)
		return err
	}
	auth := bind.NewKeyedTransactor(key)

	_, err = keeperProviderMapInstance.DelKeeper(auth, keeperAddress)
	if err != nil {
		fmt.Println("deleteKeeperInkpMapErr:", err)
		return err
	}
	return nil
}

func deleteProvider(localAddress common.Address, hexKey string, keeperAddress common.Address, providerAddress common.Address) error {
	keeperProviderMapInstance, err := getKeeperProviderMapInstanceFromIndexer(localAddress)
	if err != nil {
		fmt.Println("getKeeperProviderMapInstanceErr:", err)
		return err
	}

	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		fmt.Println("HexToECDSAErr:", err)
		return err
	}
	auth := bind.NewKeyedTransactor(key)

	_, err = keeperProviderMapInstance.DelProvider(auth, keeperAddress, providerAddress)
	if err != nil {
		fmt.Println("deleteProviderInkpMapErr:", err)
		return err
	}
	return nil
}

func getAllKeeperInKPMap(localAddress common.Address) ([]common.Address, error) {
	var keeperAddresses []common.Address

	keeperProviderMapInstance, err := getKeeperProviderMapInstanceFromIndexer(localAddress)
	if err != nil {
		fmt.Println("getKeeperProviderMapInstanceErr:", err)
		return nil, err
	}

	keeperAddresses, err = keeperProviderMapInstance.GetAllKeeper(&bind.CallOpts{
		From: localAddress,
	})
	if err != nil {
		fmt.Println("deleteProviderInkpMapErr:", err)
		return nil, err
	}
	return keeperAddresses, nil
}

func getProviderInKPMap(localAddress common.Address, keeperAddress common.Address) ([]common.Address, error) {
	var providerAddresses []common.Address

	keeperProviderMapInstance, err := getKeeperProviderMapInstanceFromIndexer(localAddress)
	if err != nil {
		fmt.Println("getKeeperProviderMapInstanceErr:", err)
		return nil, err
	}

	providerAddresses, err = keeperProviderMapInstance.GetProvider(&bind.CallOpts{
		From: localAddress,
	}, keeperAddress)
	if err != nil {
		fmt.Println("deleteProviderInkpMapErr:", err)
		return nil, err
	}
	return providerAddresses, nil
}
