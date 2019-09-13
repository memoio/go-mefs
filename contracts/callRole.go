package contracts

import (
	"fmt"
	"log"
	"math/big"

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
func DeployKeeperProviderMap(hexKey string) (*role.KeeperProviderMap, error) {
	fmt.Println("begin deploy keeperProviderMap...")

	//之前没有部署过，部署keeperProviderMap合约
	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	client := GetClient(EndPoint)

	keeperProviderMapAddr, _, keeperProviderMapInstance, err := role.DeployKeeperProviderMap(auth, client)
	if err != nil {
		fmt.Println("deployKeeperProviderMapErr:", err)
		return nil, err
	}

	indexerAddr := common.HexToAddress(IndexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, GetClient(EndPoint))
	if err != nil {
		fmt.Println("newIndexerErr:", err)
		return nil, err
	}

	indexer.Add(auth, "keeperProviderMap", keeperProviderMapAddr)

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

func AddKeeperProvidersToKPMap(localAddress common.Address, hexKey string, keeperAddress common.Address, providerAddresses []common.Address) error {
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
	auth.GasPrice = big.NewInt(defaultGasPrice)

	_, err = keeperProviderMapInstance.Add(auth, keeperAddress, providerAddresses)
	if err != nil {
		fmt.Println("addKeeperProviderTokpMapErr:", err)
		return err
	}
	return nil
}

func DeleteKeeperFromKPMap(localAddress common.Address, hexKey string, keeperAddress common.Address) error {
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
	auth.GasPrice = big.NewInt(defaultGasPrice)

	_, err = keeperProviderMapInstance.DelKeeper(auth, keeperAddress)
	if err != nil {
		fmt.Println("deleteKeeperInkpMapErr:", err)
		return err
	}
	return nil
}

func DeleteProviderFromKPMap(localAddress common.Address, hexKey string, keeperAddress common.Address, providerAddress common.Address) error {
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
	auth.GasPrice = big.NewInt(defaultGasPrice)

	_, err = keeperProviderMapInstance.DelProvider(auth, keeperAddress, providerAddress)
	if err != nil {
		fmt.Println("deleteProviderInkpMapErr:", err)
		return err
	}
	return nil
}

func GetAllKeeperInKPMap(localAddress common.Address) ([]common.Address, error) {
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
		fmt.Println("getKeeperInkpMapErr:", err)
		return nil, err
	}
	return keeperAddresses, nil
}

func GetProviderInKPMap(localAddress common.Address, keeperAddress common.Address) ([]common.Address, error) {
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
		fmt.Println("getProviderInkpMapErr:", err)
		return nil, err
	}
	return providerAddresses, nil
}
