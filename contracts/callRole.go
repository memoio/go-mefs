package contracts

import (
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/memoio/go-mefs/contracts/indexer"
	"github.com/memoio/go-mefs/contracts/role"
	id "github.com/memoio/go-mefs/crypto/identity"
	"github.com/memoio/go-mefs/utils"
)

// DeployKeeperAdmin deploy a keeper contract
func DeployKeeperAdmin(hexKey string) (err error) {
	client := GetClient(EndPoint)
	auth, err := MakeAuth(hexKey, nil, nil, big.NewInt(defaultGasPrice), 0)
	if err != nil {
		return err
	}

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

//SetKeeper set "accountAddress" keeper in contract if isKeeper is true
func SetKeeper(accountAddress common.Address, hexKey string, isKeeper bool) (err error) {
	_, keeperInstance, err := GetKeeperContractFromIndexer(accountAddress)
	if err != nil {
		log.Println("keeperContracterr:", err)
		return err
	}

	log.Println("begin set keeper...")
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(hexKey, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = keeperInstance.Set(auth, accountAddress, isKeeper)
		if err != nil {
			retryCount++
			log.Println("set keeper error:", err)
			if retryCount > sendTransactionRetryCount {
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("set keeper transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}
	log.Println("keeper has been successfully set!")
	return nil
}

//IsKeeper judge if localAddress is keeper
func IsKeeper(localAddress common.Address) (bool, error) {
	_, keeperContract, err := GetKeeperContractFromIndexer(localAddress)
	if err != nil {
		log.Println("get keeperContract err:", err)
		return false, err
	}

	isKeeper, _, _, _, err := keeperContract.Info(&bind.CallOpts{
		From: localAddress,
	}, localAddress)
	if err != nil {
		log.Println("get isKeeper info err:", err)
		return false, err
	}
	return isKeeper, nil
}

func SetKeeperBanned(accountAddress common.Address, hexKey string, isBanned bool) (err error) {
	_, keeperInstance, err := GetKeeperContractFromIndexer(accountAddress)
	if err != nil {
		log.Println("get keeperContract err:", err)
		return err
	}

	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(hexKey, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = keeperInstance.SetBanned(auth, accountAddress, isBanned)
		if err != nil {
			retryCount++
			log.Println("ban keeper error:", err)
			if retryCount > sendTransactionRetryCount {
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("ban keeper transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}
	return nil
}

func SetKeeperPrice(localAddress common.Address, hexKey string, price *big.Int) (err error) {
	_, keeperInstance, err := GetKeeperContractFromIndexer(localAddress)
	if err != nil {
		log.Println("keeperContracterr:", err)
		return err
	}

	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(hexKey, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = keeperInstance.SetPrice(auth, price)
		if err != nil {
			retryCount++
			log.Println("set keeper price error:", err)
			if retryCount > sendTransactionRetryCount {
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("set keeper price transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
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

	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(hexKey, amount, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = kInstance.Pledge(auth)
		if err != nil {
			retryCount++
			log.Println("keeper pledge error:", err)
			if err.Error() == core.ErrNonceTooLow.Error() && auth.GasPrice.Cmp(big.NewInt(defaultGasPrice)) > 0 {
				log.Println("previously pending transaction has successfully executed")
				break
			}
			if retryCount > sendTransactionRetryCount {
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("keeper pledge transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}

	log.Println("pledge keeper success")
	return nil
}

// GetAllKeepers gets all keepers from chain
func GetAllKeepersAddr(localAddr common.Address) ([]common.Address, error) {
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
	client := GetClient(EndPoint)
	auth, err := MakeAuth(hexKey, nil, nil, big.NewInt(defaultGasPrice), 0)
	if err != nil {
		return err
	}

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

//SetProvider set "accountAddress" provider in contract if isProvider is true
func SetProvider(accountAddress common.Address, hexKey string, isProvider bool) (err error) {
	_, providerInstance, err := GetProviderContractFromIndexer(accountAddress)
	if err != nil {
		return err
	}

	log.Println("begin set provider...")
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(hexKey, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = providerInstance.Set(auth, accountAddress, isProvider)
		if err != nil {
			retryCount++
			log.Println("set provider error:", err)
			if retryCount > sendTransactionRetryCount {
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("set provider transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}
	log.Println("provider has been successfully set!")
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
		log.Println("isProviderErr:", err)
		return false, err
	}
	return isProvider, nil
}

func SetProviderBanned(accountAddress common.Address, hexKey string, isBanned bool) (err error) {
	_, providerInstance, err := GetProviderContractFromIndexer(accountAddress)
	if err != nil {
		log.Println("providerContracterr:", err)
		return err
	}

	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(hexKey, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = providerInstance.SetBanned(auth, accountAddress, isBanned)
		if err != nil {
			retryCount++
			log.Println("ban provider error:", err)
			if retryCount > sendTransactionRetryCount {
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("ban provider transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
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

	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(hexKey, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = providerInstance.SetPrice(auth, price)
		if err != nil {
			retryCount++
			log.Println("set provider price error:", err)
			if retryCount > sendTransactionRetryCount {
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("set provider price transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}
	return nil
}

func PledgeProvider(localAddress common.Address, hexKey string, money *big.Int) (err error) {
	_, providerInstance, err := GetProviderContractFromIndexer(localAddress)
	if err != nil {
		log.Println("providerContracterr:", err)
		return err
	}

	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(hexKey, money, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = providerInstance.Pledge(auth, big.NewInt(0))
		if err != nil {
			retryCount++
			log.Println("provider pledge error:", err)
			if err.Error() == core.ErrNonceTooLow.Error() && auth.GasPrice.Cmp(big.NewInt(defaultGasPrice)) > 0 {
				log.Println("previously pending transaction has successfully executed")
				break
			}
			if retryCount > sendTransactionRetryCount {
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("provider pledge transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}

	log.Println("pledge provider success")
	return nil
}

// GetAllProviders gets all provider addresses from chain
func GetAllProvidersAddr(localAddr common.Address) ([]common.Address, error) {
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
	client := GetClient(EndPoint)
	auth, err := MakeAuth(hexKey, nil, nil, big.NewInt(defaultGasPrice), 0)
	if err != nil {
		return err
	}

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
	log.Println("begin add keeperProviders to kpMap...")
	res, err := IsKeeper(keeperAddress)
	if err != nil || res == false {
		log.Println(keeperAddress.String(), "is not a keeper")
		return ErrNotKeeper
	}

	for _, proAddresses := range providerAddresses {
		res, err = IsProvider(proAddresses)
		if err != nil || res == false {
			log.Println(proAddresses.String(), "is not a provider")
			return ErrNotProvider
		}
	}

	_, kpMapInstance, err := getKPMap(localAddress)
	if err != nil {
		log.Println("getKeeperProviderMapInstanceErr:", err)
		return err
	}

	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(hexKey, nil, nil, big.NewInt(defaultGasPrice), 0)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = kpMapInstance.Add(auth, keeperAddress, providerAddresses)
		if err != nil {
			retryCount++
			log.Println("add kpMap error:", err)
			if err.Error() == core.ErrNonceTooLow.Error() && tx.GasPrice().Cmp(big.NewInt(defaultGasPrice)) > 0 {
				log.Println("previously pending transaction has successfully executed")
				break
			}
			if retryCount > sendTransactionRetryCount {
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("add kpMap transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}

	log.Println("kp have been successfully added to kpMap!")
	return nil
}

// DeleteKeeperFromKPMap deletes keeper from kpmap
func DeleteKeeperFromKPMap(localAddress common.Address, hexKey string, keeperAddress common.Address) error {
	log.Println("begin delete keeper from kpMap...")
	_, kpMapInstance, err := getKPMap(localAddress)
	if err != nil {
		log.Println("getKeeperProviderMapInstanceErr:", err)
		return err
	}

	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(hexKey, nil, nil, big.NewInt(defaultGasPrice), 0)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = kpMapInstance.DelKeeper(auth, keeperAddress)
		if err != nil {
			retryCount++
			log.Println("delete keeper from kpMap error:", err)
			if err.Error() == core.ErrNonceTooLow.Error() && tx.GasPrice().Cmp(big.NewInt(defaultGasPrice)) > 0 {
				log.Println("previously pending transaction has successfully executed")
				break
			}
			if retryCount > sendTransactionRetryCount {
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("delete keeper from kpMap transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}

	log.Println("keeper have been successfuly deleted from kpMap!")
	return nil
}

// DeleteProviderFromKPMap deletes provider from kpmap
func DeleteProviderFromKPMap(localAddress common.Address, hexKey string, keeperAddress common.Address, providerAddress common.Address) error {
	log.Println("begin delete provider from kpMap...")
	_, kpMapInstance, err := getKPMap(localAddress)
	if err != nil {
		log.Println("getKeeperProviderMapInstanceErr:", err)
		return err
	}

	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(hexKey, nil, nil, big.NewInt(defaultGasPrice), 0)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = kpMapInstance.DelProvider(auth, keeperAddress, providerAddress)
		if err != nil {
			retryCount++
			log.Println("delete provider from kpMap error:", err)
			if err.Error() == core.ErrNonceTooLow.Error() && tx.GasPrice().Cmp(big.NewInt(defaultGasPrice)) > 0 {
				log.Println("previously pending transaction has successfully executed")
				break
			}
			if retryCount > sendTransactionRetryCount {
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("delete provider from kpMap transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}

	log.Println("provider have been successfuly deleted from kpMap!")
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
