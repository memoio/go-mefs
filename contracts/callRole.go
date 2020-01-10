package contracts

import (
	"errors"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-mefs/contracts/indexer"
	"github.com/memoio/go-mefs/contracts/role"
)

func GetKeeperContractFromIndexer(localAddress common.Address) (keeperContract *role.Keeper, err error) {
	keeperContractAddr, _, err := GetResolver(localAddress, "keeper")
	if err != nil {
		log.Println("get keeper Contract Err:", err)
		return keeperContract, err
	}

	keeperContract, err = role.NewKeeper(keeperContractAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("newKeeperErr:", err)
		return keeperContract, err
	}
	return keeperContract, nil
}

//IsKeeper judge if an account is keeper
func IsKeeper(localAddress common.Address) (bool, error) {
	keeperContract, err := GetKeeperContractFromIndexer(localAddress)
	if err != nil {
		log.Println("keeperContracterr:", err)
		return false, err
	}
	isKeeper, err := keeperContract.IsKeeper(&bind.CallOpts{
		From: localAddress,
	}, localAddress)
	if err != nil {
		log.Println("isKeepererr:", err)
		return false, err
	}
	return isKeeper, nil
}

func GetProviderContractFromIndexer(localAddress common.Address) (providerContract *role.Provider, err error) {
	providerContractAddr, _, err := GetResolver(localAddress, "provider")
	if err != nil {
		log.Println("get provider Contract Err:", err)
		return providerContract, err
	}

	providerContract, err = role.NewProvider(providerContractAddr, GetClient(EndPoint))
	if err != nil {
		log.Println(err)
		return providerContract, err
	}
	return providerContract, nil
}

//IsProvider judge if an account is provider
func IsProvider(localaddress common.Address) (bool, error) {
	providerContract, err := GetProviderContractFromIndexer(localaddress)
	if err != nil {
		log.Println("providerContracterr:", err)
		return false, err
	}
	isProvider, err := providerContract.IsProvider(&bind.CallOpts{
		From: localaddress,
	}, localaddress)
	if err != nil {
		log.Println("isKeepererr:", err)
		return false, err
	}
	return isProvider, nil
}

// KeeperContract deploy a keeper contract
func KeeperContract(hexKey string) (err error) {
	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	client := GetClient(EndPoint)

	//暂时将质押金额设为0
	deposit := big.NewInt(0)
	keeperContractAddr, _, _, err := role.DeployKeeper(auth, client, deposit)
	if err != nil {
		log.Println("deployKeeperErr:", err)
		return err
	}
	log.Println("keeperContractAddr:", keeperContractAddr.String())

	indexerAddr := common.HexToAddress(indexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("newIndexerErr:", err)
		return err
	}

	indexer.Add(auth, "keeper", keeperContractAddr)
	return nil
}

//SetKeeper set "localAddress" keeper in contract if isKeeper is true
func SetKeeper(localAddress common.Address, hexKey string, isKeeper bool) (err error) {
	keeperInstance, err := GetKeeperContractFromIndexer(localAddress)
	if err != nil {
		log.Println("keeperContracterr:", err)
		return err
	}

	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)

	_, err = keeperInstance.Set(auth, localAddress, isKeeper)
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
		log.Println("deployProviderErr:", err)
		return err
	}
	log.Println("providerContractAddr:", providerContractAddr.String())

	indexerAddr := common.HexToAddress(indexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("newIndexerErr:", err)
		return err
	}

	indexer.Add(auth, "provider", providerContractAddr)
	return nil
}

//SetProvider set "localAddress" provider in contract if isProvider is true
func SetProvider(localAddress common.Address, hexKey string, isProvider bool) (err error) {
	provider, err := GetProviderContractFromIndexer(localAddress)
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
func DeployKeeperProviderMap(hexKey string) error {
	log.Println("begin deploy keeperProviderMap...")

	//之前没有部署过，部署keeperProviderMap合约
	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	client := GetClient(EndPoint)

	keeperProviderMapAddr, _, _, err := role.DeployKeeperProviderMap(auth, client)
	if err != nil {
		log.Println("deployKeeperProviderMapErr:", err)
		return err
	}

	indexerAddr := common.HexToAddress(indexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("newIndexerErr:", err)
		return err
	}

	indexer.Add(auth, "keeperProviderMap", keeperProviderMapAddr)

	log.Println("keeperProviderMap-contract have been successfully deployed!")
	return nil
}

func getKeeperProviderMapInstanceFromIndexer(localAddress common.Address) (*role.KeeperProviderMap, error) {
	var keeperProviderInstance *role.KeeperProviderMap

	indexerAddr := common.HexToAddress(indexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("newIndexerErr:", err)
		return keeperProviderInstance, err
	}

	_, keeperproviderMapContractAddr, err := indexer.Get(&bind.CallOpts{
		From: localAddress,
	}, "keeperProviderMap")
	if err != nil {
		log.Println("getkeeperproviderMapContractErr:", err)
		return keeperProviderInstance, err
	}

	keeperProviderInstance, err = role.NewKeeperProviderMap(keeperproviderMapContractAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("newKeeperProviderMapContractErr:", err)
		return keeperProviderInstance, err
	}
	return keeperProviderInstance, nil
}

// AddKeeperProvidersToKPMap adds provider/keeper to kpmap
func AddKeeperProvidersToKPMap(localAddress common.Address, hexKey string, keeperAddress common.Address, providerAddresses []common.Address) error {
	res, err := IsKeeper(keeperAddress)
	if err != nil || res == false {
		log.Println(keeperAddress.String(), "is not a keeper")
		return errors.New("addr is not a keeper")
	}

	for _, proAddresses := range providerAddresses {
		res, err = IsProvider(proAddresses)
		if err != nil || res == false {
			log.Println(proAddresses.String(), "is not a provider")
			return errors.New("addr is not a provider")
		}
	}

	keeperProviderMapInstance, err := getKeeperProviderMapInstanceFromIndexer(localAddress)
	if err != nil {
		log.Println("getKeeperProviderMapInstanceErr:", err)
		return err
	}

	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSAErr:", err)
		return err
	}
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)

	_, err = keeperProviderMapInstance.Add(auth, keeperAddress, providerAddresses)
	if err != nil {
		log.Println("addKeeperProviderTokpMapErr:", err)
		return err
	}
	return nil
}

// DeleteKeeperFromKPMap deletes keeper from kpmap
func DeleteKeeperFromKPMap(localAddress common.Address, hexKey string, keeperAddress common.Address) error {
	keeperProviderMapInstance, err := getKeeperProviderMapInstanceFromIndexer(localAddress)
	if err != nil {
		log.Println("getKeeperProviderMapInstanceErr:", err)
		return err
	}

	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSAErr:", err)
		return err
	}
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)

	_, err = keeperProviderMapInstance.DelKeeper(auth, keeperAddress)
	if err != nil {
		log.Println("deleteKeeperInkpMapErr:", err)
		return err
	}
	return nil
}

// DeleteProviderFromKPMap deletes provider from kpmap
func DeleteProviderFromKPMap(localAddress common.Address, hexKey string, keeperAddress common.Address, providerAddress common.Address) error {
	keeperProviderMapInstance, err := getKeeperProviderMapInstanceFromIndexer(localAddress)
	if err != nil {
		log.Println("getKeeperProviderMapInstanceErr:", err)
		return err
	}

	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSAErr:", err)
		return err
	}
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)

	_, err = keeperProviderMapInstance.DelProvider(auth, keeperAddress, providerAddress)
	if err != nil {
		log.Println("deleteProviderInkpMapErr:", err)
		return err
	}
	return nil
}

// GetAllKeeperInKPMap get keepers in kpmap
func GetAllKeeperInKPMap(localAddress common.Address) ([]common.Address, error) {
	var keeperAddresses []common.Address

	keeperProviderMapInstance, err := getKeeperProviderMapInstanceFromIndexer(localAddress)
	if err != nil {
		log.Println("getKeeperProviderMapInstanceErr:", err)
		return nil, err
	}

	keeperAddresses, err = keeperProviderMapInstance.GetAllKeeper(&bind.CallOpts{
		From: localAddress,
	})
	if err != nil {
		log.Println("getKeeperInkpMapErr:", err)
		return nil, err
	}
	return keeperAddresses, nil
}

// GetProviderInKPMap gets providers from kpmap
func GetProviderInKPMap(localAddress common.Address, keeperAddress common.Address) ([]common.Address, error) {
	var providerAddresses []common.Address

	keeperProviderMapInstance, err := getKeeperProviderMapInstanceFromIndexer(localAddress)
	if err != nil {
		log.Println("getKeeperProviderMapInstanceErr:", err)
		return nil, err
	}

	providerAddresses, err = keeperProviderMapInstance.GetProvider(&bind.CallOpts{
		From: localAddress,
	}, keeperAddress)
	if err != nil {
		log.Println("getProviderInkpMapErr:", err)
		return nil, err
	}
	return providerAddresses, nil
}
