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
	"github.com/memoio/go-mefs/utils/address"
)

//RoleInfo  The basic information of node used for 'role' contract
type RoleInfo struct {
	localID string //local ID
	hexSk   string //local privateKey
}

//NewCR new a instance of contractRole
func NewCR(localID, hexSk string) ContractRole {
	RInfo := &RoleInfo{
		localID: localID,
		hexSk:   hexSk,
	}

	return RInfo
}

// DeployKeeperAdmin deploy a keeper contract
func (r *RoleInfo) DeployKeeperAdmin() (err error) {
	client := GetClient(EndPoint)
	auth, err := MakeAuth(r.hexSk, nil, nil, big.NewInt(defaultGasPrice), 0)
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

	localAddr, err := id.GetAdressFromSk(r.hexSk)
	if err != nil {
		return err
	}

	ma := NewCManage(localAddr, r.hexSk)
	return ma.AddToIndexer(keeperContractAddr, keeperKey, indexer)
}

func GetKeeperContractFromIndexer(localAddress common.Address) (common.Address, *role.Keeper, error) {
	var res common.Address
	ma := NewCManage(localAddress, "")
	keeperContractAddr, _, err := ma.GetResolverAddr(keeperKey)
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
func (r *RoleInfo) SetKeeper(accountAddress common.Address, isKeeper bool) (err error) {
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
		auth, errMA := MakeAuth(r.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
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
			time.Sleep(retryTxSleepTime)
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
func (r *RoleInfo) IsKeeper(accountID string) (bool, error) {
	accountAddress, err := address.GetAddressFromID(accountID)
	if err != nil {
		return false, err
	}

	_, keeperContract, err := GetKeeperContractFromIndexer(accountAddress)
	if err != nil {
		log.Println("get keeperContract err:", err)
		return false, err
	}

	var isKeeper bool
	for i := 0; i < 3; i++ {
		isKeeper, _, _, _, err = keeperContract.Info(&bind.CallOpts{
			From: accountAddress,
		}, accountAddress)
		if err != nil {
			time.Sleep(retryGetInfoSleepTime)
		}
	}

	if err != nil {
		log.Println("get isKeeper info err:", err)
		return false, err
	}

	return isKeeper, nil
}

func (r *RoleInfo) SetKeeperBanned(accountAddress common.Address, isBanned bool) (err error) {
	_, keeperInstance, err := GetKeeperContractFromIndexer(accountAddress)
	if err != nil {
		log.Println("get keeperContract err:", err)
		return err
	}

	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(r.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
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
			time.Sleep(retryTxSleepTime)
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

func (r *RoleInfo) SetKeeperPrice(price *big.Int) (err error) {
	localAddr, err := address.GetAddressFromID(r.localID)
	if err != nil {
		return err
	}

	_, keeperInstance, err := GetKeeperContractFromIndexer(localAddr)
	if err != nil {
		log.Println("keeperContracterr:", err)
		return err
	}

	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(r.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
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
			time.Sleep(retryTxSleepTime)
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

func (r *RoleInfo) GetKeeperPrice() (*big.Int, error) {
	localAddr, err := address.GetAddressFromID(r.localID)
	if err != nil {
		return big.NewInt(0), err
	}

	_, keeperContract, err := GetKeeperContractFromIndexer(localAddr)
	if err != nil {
		log.Println("keeperContracterr:", err)
		return big.NewInt(0), err
	}
	price, err := keeperContract.GetPrice(&bind.CallOpts{
		From: localAddr,
	})
	if err != nil {
		log.Println("getKeeperPrice err:", err)
		return price, err
	}
	return price, nil
}

func (r *RoleInfo) PledgeKeeper(amount *big.Int) (err error) {
	localAddr, err := address.GetAddressFromID(r.localID)
	if err != nil {
		return err
	}

	_, kInstance, err := GetKeeperContractFromIndexer(localAddr)
	if err != nil {
		log.Println("getkeeperContracterr:", err)
		return err
	}

	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(r.hexSk, amount, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
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
			time.Sleep(retryTxSleepTime)
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

// GetAllKeepersAddr gets all keepers from chain
func (r *RoleInfo) GetAllKeepersAddr() ([]common.Address, error) {
	localAddr, err := address.GetAddressFromID(r.localID)
	if err != nil {
		return nil, err
	}

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

//DeployProviderAdmin deploy a keeper contract
func (r *RoleInfo) DeployProviderAdmin() (err error) {
	client := GetClient(EndPoint)
	auth, err := MakeAuth(r.hexSk, nil, nil, big.NewInt(defaultGasPrice), 0)
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

	localAddr, err := id.GetAdressFromSk(r.hexSk)
	if err != nil {
		return err
	}

	ma := NewCManage(localAddr, r.hexSk)
	return ma.AddToIndexer(providerContractAddr, providerKey, indexer)
}

func GetProviderContractFromIndexer(localAddress common.Address) (common.Address, *role.Provider, error) {
	var res common.Address
	ma := NewCManage(localAddress, "")
	providerContractAddr, _, err := ma.GetResolverAddr(providerKey)
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
func (r *RoleInfo) SetProvider(accountAddress common.Address, isProvider bool) (err error) {
	_, providerInstance, err := GetProviderContractFromIndexer(accountAddress)
	if err != nil {
		return err
	}

	log.Println("begin set provider...")
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(r.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
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
			time.Sleep(retryTxSleepTime)
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
func (r *RoleInfo) IsProvider(accountID string) (bool, error) {
	accountAddress, err := address.GetAddressFromID(accountID)
	if err != nil {
		return false, err
	}

	_, providerContract, err := GetProviderContractFromIndexer(accountAddress)
	if err != nil {
		log.Println("providerContracterr:", err)
		return false, err
	}

	var isProvider bool

	for i := 0; i < 3; i++ {
		isProvider, _, _, _, err = providerContract.Info(&bind.CallOpts{
			From: accountAddress,
		}, accountAddress)
		if err != nil {
			time.Sleep(retryGetInfoSleepTime)
		}
	}

	if err != nil {
		log.Println("isProviderErr:", err)
		return false, err
	}
	return isProvider, nil
}

func (r *RoleInfo) SetProviderBanned(accountAddress common.Address, isBanned bool) (err error) {
	_, providerInstance, err := GetProviderContractFromIndexer(accountAddress)
	if err != nil {
		log.Println("providerContracterr:", err)
		return err
	}

	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(r.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
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
			time.Sleep(retryTxSleepTime)
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

func (r *RoleInfo) GetProviderPrice() (*big.Int, error) {
	localAddr, err := address.GetAddressFromID(r.localID)
	if err != nil {
		return big.NewInt(0), err
	}

	_, providerContract, err := GetProviderContractFromIndexer(localAddr)
	if err != nil {
		log.Println("providerContracterr:", err)
		return big.NewInt(0), err
	}
	price, err := providerContract.GetPrice(&bind.CallOpts{
		From: localAddr,
	})
	if err != nil {
		log.Println("getProviderPrice err:", err)
		return price, err
	}
	return price, nil
}

func (r *RoleInfo) SetProviderPrice(price *big.Int) (err error) {
	localAddr, err := address.GetAddressFromID(r.localID)
	if err != nil {
		return err
	}

	_, providerInstance, err := GetProviderContractFromIndexer(localAddr)
	if err != nil {
		log.Println("providerContracterr:", err)
		return err
	}

	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(r.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
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
			time.Sleep(retryTxSleepTime)
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

func (r *RoleInfo) PledgeProvider(size *big.Int) (err error) {
	price, err := r.GetProviderPrice()
	if err != nil {
		return err
	}

	//根据size*price计算出质押金额，赋值给price
	weiPrice := new(big.Float).SetInt(price)
	weiPrice.Quo(weiPrice, GetMemoPrice())
	weiPrice.Int(price)
	price.Mul(price, size)

	localAddr, err := address.GetAddressFromID(r.localID)
	if err != nil {
		return err
	}

	_, providerInstance, err := GetProviderContractFromIndexer(localAddr)
	if err != nil {
		log.Println("providerContracterr:", err)
		return err
	}

	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(r.hexSk, price, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
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
			time.Sleep(retryTxSleepTime)
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

// GetAllProvidersAddr gets all provider addresses from chain
func (r *RoleInfo) GetAllProvidersAddr() ([]common.Address, error) {
	localAddr, err := address.GetAddressFromID(r.localID)
	if err != nil {
		return nil, err
	}

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
func (r *RoleInfo) DeployKPMap() error {
	log.Println("begin deploy keeperProviderMap...")

	//之前没有部署过，部署keeperProviderMap合约
	client := GetClient(EndPoint)
	auth, err := MakeAuth(r.hexSk, nil, nil, big.NewInt(defaultGasPrice), 0)
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
	ma := NewCManage(localAddress, "")
	res, _, err := ma.GetResolverAddr(kpMapKey)
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
func (r *RoleInfo) AddKeeperProvidersToKPMap(keeperAddress common.Address, providerAddresses []common.Address) error {
	log.Println("begin add keeperProviders to kpMap...")
	keeperID, err := address.GetIDFromAddress(keeperAddress.Hex())
	if err != nil {
		return err
	}
	res, err := r.IsKeeper(keeperID)
	if err != nil || res == false {
		log.Println(keeperAddress.String(), "is not a keeper")
		return ErrNotKeeper
	}

	for _, proAddresses := range providerAddresses {
		proID, _ := address.GetIDFromAddress(proAddresses.Hex())
		res, err = r.IsProvider(proID)
		if err != nil || res == false {
			log.Println(proAddresses.String(), "is not a provider")
			return ErrNotProvider
		}
	}

	localAddr, err := address.GetAddressFromID(r.localID)
	if err != nil {
		return err
	}

	_, kpMapInstance, err := getKPMap(localAddr)
	if err != nil {
		log.Println("getKeeperProviderMapInstanceErr:", err)
		return err
	}

	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(r.hexSk, nil, nil, big.NewInt(defaultGasPrice), 0)
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
			time.Sleep(retryTxSleepTime)
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
func (r *RoleInfo) DeleteKeeperFromKPMap(keeperAddress common.Address) error {
	log.Println("begin delete keeper from kpMap...")
	localAddr, err := address.GetAddressFromID(r.localID)
	if err != nil {
		return err
	}

	_, kpMapInstance, err := getKPMap(localAddr)
	if err != nil {
		log.Println("getKeeperProviderMapInstanceErr:", err)
		return err
	}

	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(r.hexSk, nil, nil, big.NewInt(defaultGasPrice), 0)
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
			time.Sleep(retryTxSleepTime)
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
func (r *RoleInfo) DeleteProviderFromKPMap(keeperAddress common.Address, providerAddress common.Address) error {
	log.Println("begin delete provider from kpMap...")
	localAddr, err := address.GetAddressFromID(r.localID)
	if err != nil {
		return err
	}

	_, kpMapInstance, err := getKPMap(localAddr)
	if err != nil {
		log.Println("getKeeperProviderMapInstanceErr:", err)
		return err
	}

	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(r.hexSk, nil, nil, big.NewInt(defaultGasPrice), 0)
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
			time.Sleep(retryTxSleepTime)
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
func (r *RoleInfo) GetAllKeeperInKPMap() ([]common.Address, error) {
	var keeperAddresses []common.Address

	localAddr, err := address.GetAddressFromID(r.localID)
	if err != nil {
		return nil, err
	}
	_, kpMapInstance, err := getKPMap(localAddr)
	if err != nil {
		log.Println("getKeeperProviderMapInstanceErr:", err)
		return nil, err
	}

	keeperAddresses, err = kpMapInstance.GetAllKeeper(&bind.CallOpts{
		From: localAddr,
	})
	if err != nil {
		log.Println("getKeeperInkpMapErr:", err)
		return nil, err
	}
	return keeperAddresses, nil
}

// GetProviderInKPMap gets providers from kpmap
func (r *RoleInfo) GetProviderInKPMap(keeperAddress common.Address) ([]common.Address, error) {
	var providerAddresses []common.Address

	localAddr, err := address.GetAddressFromID(r.localID)
	if err != nil {
		return nil, err
	}

	_, kpMapInstance, err := getKPMap(localAddr)
	if err != nil {
		log.Println("getKeeperProviderMapInstanceErr:", err)
		return nil, err
	}

	providerAddresses, err = kpMapInstance.GetProvider(&bind.CallOpts{
		From: localAddr,
	}, keeperAddress)
	if err != nil {
		log.Println("getProviderInkpMapErr:", err)
		return nil, err
	}
	return providerAddresses, nil
}
