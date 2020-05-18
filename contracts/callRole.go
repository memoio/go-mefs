package contracts

import (
	"errors"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-mefs/contracts/indexer"
	"github.com/memoio/go-mefs/contracts/role"
	id "github.com/memoio/go-mefs/crypto/identity"
	"github.com/memoio/go-mefs/utils"
)

// DeployKeeperAdmin deploy a keeper contract
func DeployKeeperAdmin(hexKey string) (err error) {
	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	client := GetClient(EndPoint)

	deposit := big.NewInt(utils.KeeperDeposit)
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

	localAddr, err := id.GetAdressFromSk(hexKey)
	if err != nil {
		return err
	}

	return AddToIndexer(localAddr, keeperContractAddr, keeperKey, hexKey, indexer)
}

func GetKeeperContractFromIndexer(localAddress common.Address) (common.Address, *role.Keeper, error) {
	var res common.Address
	keeperContractAddr, _, err := GetResolverAddr(localAddress, keeperKey)
	if err != nil {
		log.Println("get keeper Contract Err:", err)
		return res, nil, err
	}

	keeperContract, err := role.NewKeeper(keeperContractAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("newKeeperErr:", err)
		return res, nil, err
	}
	return keeperContractAddr, keeperContract, nil
}

//SetKeeper set "localAddress" keeper in contract if isKeeper is true
func SetKeeper(localAddress common.Address, hexKey string, isKeeper bool) (err error) {
	_, keeperInstance, err := GetKeeperContractFromIndexer(localAddress)
	if err != nil {
		log.Println("keeperContracterr:", err)
		return err
	}

	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	auth.GasLimit = defaultGasLimit

	_, err = keeperInstance.Set(auth, localAddress, isKeeper)
	if err != nil {
		return err
	}
	return nil
}

//IsKeeper judge if an account is keeper
func IsKeeper(localAddress common.Address) (bool, error) {
	_, keeperContract, err := GetKeeperContractFromIndexer(localAddress)
	if err != nil {
		log.Println("keeperContracterr:", err)
		return false, err
	}

	isKeeper, _, _, _, err := keeperContract.Info(&bind.CallOpts{
		From: localAddress,
	}, localAddress)
	if err != nil {
		log.Println("isKeepererr:", err)
		return false, err
	}
	return isKeeper, nil
}

func GetKeeperInfo(localAddress common.Address) (bool, bool, *big.Int, *big.Int, error) {
	_, keeperContract, err := GetKeeperContractFromIndexer(localAddress)
	if err != nil {
		log.Println("keeperContracterr:", err)
		return false, false, nil, nil, err
	}
	return keeperContract.Info(&bind.CallOpts{
		From: localAddress,
	}, localAddress)
}

func SetKeeperBanned(localAddress common.Address, hexKey string, isBanned bool) (err error) {
	_, keeperInstance, err := GetKeeperContractFromIndexer(localAddress)
	if err != nil {
		log.Println("keeperContracterr:", err)
		return err
	}

	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	auth.GasLimit = defaultGasLimit

	_, err = keeperInstance.SetBanned(auth, localAddress, isBanned)
	if err != nil {
		return err
	}
	return nil
}

func SetKeeperPrice(localAddress common.Address, hexKey string, price *big.Int) (err error) {
	_, keeperInstance, err := GetKeeperContractFromIndexer(localAddress)
	if err != nil {
		log.Println("keeperContracterr:", err)
		return err
	}

	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	auth.GasLimit = defaultGasLimit

	_, err = keeperInstance.SetPrice(auth, price)
	if err != nil {
		return err
	}
	return nil
}

func GetKeeperPrice(localAddress common.Address) (*big.Int, error) {
	_, keeperContract, err := GetKeeperContractFromIndexer(localAddress)
	if err != nil {
		log.Println("keeperContracterr:", err)
		return big.NewInt(0), err
	}
	price, err := keeperContract.GetPrice(&bind.CallOpts{
		From: localAddress,
	})
	if err != nil {
		log.Println("getKeeperPrice err:", err)
		return price, err
	}
	return price, nil
}

func PledgeKeeper(localAddress common.Address, hexKey string, amount *big.Int) (err error) {
	_, kInstance, err := GetKeeperContractFromIndexer(localAddress)
	if err != nil {
		log.Println("getkeeperContracterr:", err)
		return err
	}

	key, _ := crypto.HexToECDSA(hexKey)
	retryCount := 0
	for {
		retryCount++

		auth := bind.NewKeyedTransactor(key)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		auth.GasLimit = defaultGasLimit
		auth.Value = amount
		tx, err := kInstance.Pledge(auth)
		if err != nil {
			return err
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > checkTxRetryCount {
				log.Println("keeper pledge transaction Err:", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		break
	}

	return nil
}

// GetAllKeepers gets all keepers from chain
func GetAllKeepers(localAddr common.Address) ([]common.Address, error) {
	_, keeperContract, err := GetKeeperContractFromIndexer(localAddr)
	if err != nil {
		log.Println("keeperContracterr:", err)
		return nil, err
	}

	res, err := keeperContract.GetAllAddress(&bind.CallOpts{
		From: localAddr,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

//---------provider----------//

//DeployProvider deploy a keeper contract
func DeployProviderAdmin(hexKey string) (err error) {
	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	client := GetClient(EndPoint)

	//暂时将存储容量、质押金额设为1000
	deposit := big.NewInt(utils.ProviderDeposit)
	providerContractAddr, _, _, err := role.DeployProvider(auth, client, deposit)
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

	localAddr, err := id.GetAdressFromSk(hexKey)
	if err != nil {
		return err
	}

	return AddToIndexer(localAddr, providerContractAddr, providerKey, hexKey, indexer)
}

func GetProviderContractFromIndexer(localAddress common.Address) (common.Address, *role.Provider, error) {
	var res common.Address
	providerContractAddr, _, err := GetResolverAddr(localAddress, providerKey)
	if err != nil {
		log.Println("get provider Contract Err:", err)
		return res, nil, err
	}

	providerContract, err := role.NewProvider(providerContractAddr, GetClient(EndPoint))
	if err != nil {
		log.Println(err)
		return res, nil, err
	}
	return providerContractAddr, providerContract, nil
}

//SetProvider set "localAddress" provider in contract if isProvider is true
func SetProvider(localAddress common.Address, hexKey string, isProvider bool) (err error) {
	_, provider, err := GetProviderContractFromIndexer(localAddress)
	if err != nil {
		return err
	}

	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	auth.GasLimit = defaultGasLimit
	_, err = provider.Set(auth, localAddress, isProvider)
	if err != nil {
		return err
	}
	return nil
}

//IsProvider judge if an account is provider
func IsProvider(localaddress common.Address) (bool, error) {
	_, providerContract, err := GetProviderContractFromIndexer(localaddress)
	if err != nil {
		log.Println("providerContracterr:", err)
		return false, err
	}
	isProvider, _, _, _, err := providerContract.Info(&bind.CallOpts{
		From: localaddress,
	}, localaddress)
	if err != nil {
		log.Println("isKeepererr:", err)
		return false, err
	}
	return isProvider, nil
}

func GetProviderInfo(localaddress common.Address) (bool, bool, *big.Int, *big.Int, error) {
	_, providerContract, err := GetProviderContractFromIndexer(localaddress)
	if err != nil {
		log.Println("providerContracterr:", err)
		return false, false, nil, nil, err
	}
	return providerContract.Info(&bind.CallOpts{
		From: localaddress,
	}, localaddress)
}

func SetProviderBanned(localAddress common.Address, hexKey string, isBanned bool) (err error) {
	_, providerInstance, err := GetProviderContractFromIndexer(localAddress)
	if err != nil {
		log.Println("providerContracterr:", err)
		return err
	}

	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	auth.GasLimit = defaultGasLimit

	_, err = providerInstance.SetBanned(auth, localAddress, isBanned)
	if err != nil {
		return err
	}
	return nil
}

func GetProviderPrice(localAddress common.Address) (*big.Int, error) {
	_, providerContract, err := GetProviderContractFromIndexer(localAddress)
	if err != nil {
		log.Println("providerContracterr:", err)
		return big.NewInt(0), err
	}
	price, err := providerContract.GetPrice(&bind.CallOpts{
		From: localAddress,
	})
	if err != nil {
		log.Println("getProviderPrice err:", err)
		return price, err
	}
	return price, nil
}

func SetProviderPrice(localAddress common.Address, hexKey string, price *big.Int) (err error) {
	_, providerInstance, err := GetProviderContractFromIndexer(localAddress)
	if err != nil {
		log.Println("providerContracterr:", err)
		return err
	}

	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	auth.GasLimit = defaultGasLimit

	_, err = providerInstance.SetPrice(auth, price)
	if err != nil {
		return err
	}
	return nil
}

func PledgeProvider(localAddress common.Address, hexKey string, money *big.Int) (err error) {
	_, providerInstance, err := GetProviderContractFromIndexer(localAddress)
	if err != nil {
		log.Println("providerContracterr:", err)
		return err
	}

	key, _ := crypto.HexToECDSA(hexKey)

	retryCount := 0
	for {
		retryCount++

		auth := bind.NewKeyedTransactor(key)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		auth.Value = money
		auth.GasLimit = defaultGasLimit
		tx, err := providerInstance.Pledge(auth, big.NewInt(0))
		if err != nil {
			return err
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > checkTxRetryCount {
				log.Println("provider pledge transaction Err:", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		break
	}

	return nil
}

// GetAllProviders gets all provider addresses from chain
func GetAllProviders(localAddr common.Address) ([]common.Address, error) {
	_, proContract, err := GetProviderContractFromIndexer(localAddr)
	if err != nil {
		log.Println("providerContracterr:", err)
		return nil, err
	}

	res, err := proContract.GetAllAddress(&bind.CallOpts{
		From: localAddr,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

//----------------------kpmap---------------------------//

//DeployKPMa deploy KeeperProviderMap-contract
func DeployKPMap(hexKey string) error {
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

	indexer.Add(auth, kpMapKey, keeperProviderMapAddr)

	log.Println("keeperProviderMap-contract have been successfully deployed!")
	return nil
}

func getKPMap(localAddress common.Address) (common.Address, *role.KeeperProviderMap, error) {
	res, _, err := GetResolverAddr(localAddress, kpMapKey)
	if err != nil {
		log.Println("get resolver for kpmap err: ", err)
		return res, nil, err
	}

	keeperProviderInstance, err := role.NewKeeperProviderMap(res, GetClient(EndPoint))
	if err != nil {
		log.Println("newKeeperProviderMapContractErr:", err)
		return res, keeperProviderInstance, err
	}
	return res, keeperProviderInstance, nil
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

	_, kpMapInstance, err := getKPMap(localAddress)
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

	_, err = kpMapInstance.Add(auth, keeperAddress, providerAddresses)
	if err != nil {
		log.Println("addKeeperProviderTokpMapErr:", err)
		return err
	}
	return nil
}

// DeleteKeeperFromKPMap deletes keeper from kpmap
func DeleteKeeperFromKPMap(localAddress common.Address, hexKey string, keeperAddress common.Address) error {
	_, kpMapInstance, err := getKPMap(localAddress)
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

	_, err = kpMapInstance.DelKeeper(auth, keeperAddress)
	if err != nil {
		log.Println("deleteKeeperInkpMapErr:", err)
		return err
	}
	return nil
}

// DeleteProviderFromKPMap deletes provider from kpmap
func DeleteProviderFromKPMap(localAddress common.Address, hexKey string, keeperAddress common.Address, providerAddress common.Address) error {
	_, kpMapInstance, err := getKPMap(localAddress)
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

	_, err = kpMapInstance.DelProvider(auth, keeperAddress, providerAddress)
	if err != nil {
		log.Println("deleteProviderInkpMapErr:", err)
		return err
	}
	return nil
}

// GetAllKeeperInKPMap get keepers in kpmap
func GetAllKeeperInKPMap(localAddress common.Address) ([]common.Address, error) {
	var keeperAddresses []common.Address

	_, kpMapInstance, err := getKPMap(localAddress)
	if err != nil {
		log.Println("getKeeperProviderMapInstanceErr:", err)
		return nil, err
	}

	keeperAddresses, err = kpMapInstance.GetAllKeeper(&bind.CallOpts{
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

	_, kpMapInstance, err := getKPMap(localAddress)
	if err != nil {
		log.Println("getKeeperProviderMapInstanceErr:", err)
		return nil, err
	}

	providerAddresses, err = kpMapInstance.GetProvider(&bind.CallOpts{
		From: localAddress,
	}, keeperAddress)
	if err != nil {
		log.Println("getProviderInkpMapErr:", err)
		return nil, err
	}
	return providerAddresses, nil
}
