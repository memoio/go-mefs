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

func getKeeperContractFromIndexer(endPoint string, localAddress common.Address) (keeperContract *role.Keeper, err error) {
	indexerAddr := common.HexToAddress(IndexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, GetClient(endPoint))
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

	keeperContract, err = role.NewKeeper(keeperContractAddr, GetClient(endPoint))
	if err != nil {
		fmt.Println("newKeeperErr:", err)
		return keeperContract, err
	}
	return keeperContract, nil
}

//IsKeeper judge if an account is keeper
func IsKeeper(endPoint string, localAddress common.Address) (bool, error) {
	keeperContract, err := getKeeperContractFromIndexer(endPoint, localAddress)
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

func getProviderContractFromIndexer(endPoint string, localAddress common.Address) (providerContract *role.Provider, err error) {
	indexerAddr := common.HexToAddress(IndexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, GetClient(endPoint))
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

	providerContract, err = role.NewProvider(providerContractAddr, GetClient(endPoint))
	if err != nil {
		fmt.Println(err)
		return providerContract, err
	}
	return providerContract, nil
}

//IsProvider judge if an account is provider
func IsProvider(endPoint string, localaddress common.Address) (bool, error) {
	providerContract, err := getProviderContractFromIndexer(endPoint, localaddress)
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
func KeeperContract(endPoint string, hexKey string) (err error) {
	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	client := GetClient(endPoint)

	//暂时将质押金额设为0
	deposit := big.NewInt(0)
	keeperContractAddr, _, _, err := role.DeployKeeper(auth, client, deposit)
	if err != nil {
		fmt.Println("deployKeeperErr:", err)
		return err
	}
	log.Println("keeperContractAddr:", keeperContractAddr.String())

	indexerAddr := common.HexToAddress(IndexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, GetClient(endPoint))
	if err != nil {
		fmt.Println("newIndexerErr:", err)
		return err
	}

	indexer.Add(auth, "keeper", keeperContractAddr)
	return nil
}

//SetKeeper set "localAddress" keeper in contract if isKeeper is true
func SetKeeper(endPoint string, localAddress common.Address, hexKey string, isKeeper bool) (err error) {
	keeper, err := getKeeperContractFromIndexer(endPoint, localAddress)
	if err != nil {
		fmt.Println("keeperContracterr:", err)
		return err
	}

	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)

	_, err = keeper.Set(auth, localAddress, isKeeper)
	if err != nil {
		return err
	}
	return nil
}

//ProviderContract deploy a keeper contract
func ProviderContract(endPoint string, hexKey string) (err error) {
	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	client := GetClient(endPoint)

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
	indexer, err := indexer.NewIndexer(indexerAddr, GetClient(endPoint))
	if err != nil {
		fmt.Println("newIndexerErr:", err)
		return err
	}

	indexer.Add(auth, "provider", providerContractAddr)
	return nil
}

//SetProvider set "localAddress" provider in contract if isProvider is true
func SetProvider(endPoint string, localAddress common.Address, hexKey string, isProvider bool) (err error) {
	provider, err := getProviderContractFromIndexer(endPoint, localAddress)
	if err != nil {
		return err
	}

	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	_, err = provider.Set(auth, localAddress, isProvider)
	if err != nil {
		return err
	}
	return nil
}
