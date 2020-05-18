package contracts

import (
	"context"
	"errors"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/memoio/go-mefs/contracts/channel"
	"github.com/memoio/go-mefs/contracts/indexer"
	"github.com/memoio/go-mefs/contracts/mapper"
	"github.com/memoio/go-mefs/contracts/resolver"
	"github.com/memoio/go-mefs/contracts/upKeeping"
	"github.com/memoio/go-mefs/utils"
)

// EndPoint config中的ETH，在daemon中赋值
var EndPoint string

const (
	//indexerHex indexerAddress, it is well known
	indexerHex = "0x9e4af0964ef92095ca3d2ae0c05b472837d8bd37"
	//InvalidAddr implements invalid contracts-address
	InvalidAddr               = "0x0000000000000000000000000000000000000000"
	spaceTimePayGasLimit      = uint64(8000000)
	spaceTimePayGasPrice      = 100
	defaultGasPrice           = 100
	defaultGasLimit           = uint64(8000000)
	sendTransactionRetryCount = 5
	checkTxRetryCount         = 10
)

const (
	keeperKey   = "keeperV1"
	providerKey = "providerV1"
	kpMapKey    = "kpMapV1"

	offerKey   = "offerV1"
	queryKey   = "queryV1"
	ukey       = "upKeepingV1"
	rootKey    = "rootV1"
	channelKey = "channelV1"
)

var (
	ErrEmpty              = errors.New("has not addr")
	ErrMisType            = errors.New("mistype contract")
	ErrNotDeployedIndexer = errors.New("has not deployed indexer")
	//ErrNotDeployedMapper the user has not deployed mapper in the specified resolver
	ErrNotDeployedMapper = errors.New("has not deployed mapper")
	//ErrNotDeployedResolver the provider has not deployed resolver
	ErrNotDeployedResolver = errors.New("has not deployed resolver")
	//ErrNotDeployedUk the user has not deployed uk in the specified mapper
	ErrNotDeployedUk = errors.New("has not deployed upKeeping")
	// ErrNotDeployedChannel is
	ErrNotDeployedChannel = errors.New("the user has not deployed channel-contract with you")
	// ErrContractNotPutToMapper is
	ErrContractNotPutToMapper = errors.New("the upKeeping-contract has not been added to mapper within a specified period of time")
	// ErrMarketType is
	ErrMarketType = errors.New("The market type is error, please input correct market type")
	// ErrNotDeployedMarket is
	ErrNotDeployedMarket = errors.New("has not deployed query or offer")
	// ErrNewContractInstance is
	ErrNewContractInstance = errors.New("new contract Instace failed")
	// ErrNotDeployedKPMap is
	ErrNotDeployedKPMap = errors.New("has not deployed keeperProviderMap contract")
	ErrTxFail           = errors.New("transaction fails")
)

type LogPay struct {
	From  common.Address
	To    common.Address
	Value *big.Int
}

type LogCloseChannel struct {
	From  common.Address
	Value *big.Int
}

func init() {
	EndPoint = "http://212.64.28.207:8101"
}

//GetClient get rpc-client based the endPoint
func GetClient(endPoint string) *ethclient.Client {
	client, err := rpc.Dial(endPoint)
	if err != nil {
		log.Println(err)
	}
	return ethclient.NewClient(client)
}

//QueryBalance query the balance of account
func QueryBalance(account string) (balance *big.Int, err error) {
	var result string
	client, err := rpc.Dial(EndPoint)
	if err != nil {
		log.Println("rpc.dial err:", err)
		return balance, err
	}
	err = client.Call(&result, "eth_getBalance", account, "latest")
	if err != nil {
		log.Println("client.call err:", err)
		return balance, err
	}
	balance = utils.HexToBigInt(result)
	return balance, nil
}

//GetLatestBlock query the balance of account
func GetLatestBlock() (*types.Block, error) {
	client, err := ethclient.Dial(EndPoint)
	if err != nil {
		log.Println("rpc.Dial err", err)
		return nil, err
	}

	b, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		log.Println("client.call err:", err)
		return nil, err
	}

	return b, nil
}

// DeployIndexer deploy role's indexer and add it to admin indexer
func DeployIndexer(hexKey string) (common.Address, *indexer.Indexer, error) {
	var indexerAddr common.Address

	client := GetClient(EndPoint)
	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSA err: ", err)
		return indexerAddr, nil, err
	}

	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		indexerAddr, tx, indexerInstance, err := indexer.DeployIndexer(auth, client)
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				log.Println("deploy Indexer Err:", err)
				return indexerAddr, indexerInstance, err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			log.Println("deploy user indexer transaction fails:", err)
			if retryCount > checkTxRetryCount {
				return indexerAddr, indexerInstance, err
			}
			time.Sleep(time.Minute)
			continue
		}

		return indexerAddr, indexerInstance, nil
	}
}

// AddToIndexer adds
func AddToIndexer(localAddress, addAddr common.Address, key, hexKey string, adminIndexer *indexer.Indexer) error {
	_, ownAddr, err := GetAddrFromIndexer(localAddress, key, adminIndexer)
	if ownAddr.String() == localAddress.String() {
		return AlterAddrInIndexer(localAddress, addAddr, key, hexKey, adminIndexer)
	}

	if err == nil {
		return nil
	}

	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSA err: ", err)
		return err
	}

	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		tx, err := adminIndexer.Add(auth, key, addAddr)
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				log.Println("add role indexer err:", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > checkTxRetryCount {
				log.Println("add usroleer indexer transaction fails: ", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		addrGetted, _, err := GetAddrFromIndexer(localAddress, key, adminIndexer)
		if err != nil {
			time.Sleep(time.Minute)
			continue
		}

		if addrGetted == addAddr {
			return nil
		}

		if retryCount > 20 {
			return ErrNotDeployedIndexer
		}
	}
}

// AlterAddrInIndexer alters
func AlterAddrInIndexer(localAddress, addAddr common.Address, key, hexKey string, adminIndexer *indexer.Indexer) error {
	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSA err: ", err)
		return err
	}

	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		tx, err := adminIndexer.AlterResolver(auth, key, addAddr)
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				log.Println("add role indexer err:", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > checkTxRetryCount {
				log.Println("add usroleer indexer transaction fails: ", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		if retryCount > 20 {
			return ErrNotDeployedIndexer
		}

		addrGetted, _, err := GetAddrFromIndexer(localAddress, key, adminIndexer)
		if err != nil {
			time.Sleep(time.Minute)
			continue
		}

		if addrGetted == addAddr {
			return nil
		}
	}
}

// GetAddrFromIndexer gets addr
func GetAddrFromIndexer(localAddress common.Address, key string, indexerInstance *indexer.Indexer) (common.Address, common.Address, error) {
	retryCount := 0
	for {
		retryCount++
		ownAddr, indexerAddr, err := indexerInstance.Get(&bind.CallOpts{
			From: localAddress,
		}, key)
		if err != nil {
			if retryCount > 10 {
				log.Println("get addr from indexer err: ", err)
				return indexerAddr, ownAddr, err
			}
			time.Sleep(30 * time.Second)
			continue
		}

		if len(indexerAddr) == 0 || indexerAddr.String() == InvalidAddr {
			log.Println("get empty addr from indexer")
			return indexerAddr, ownAddr, ErrEmpty
		}

		return indexerAddr, ownAddr, nil
	}
}

// DeployResolver deploys
func DeployResolver(hexKey string) (common.Address, *resolver.Resolver, error) {
	var resolverAddr common.Address
	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSA err: ", err)
		return resolverAddr, nil, err
	}

	client := GetClient(EndPoint)
	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		resAddr, tx, resolverInstance, err := resolver.DeployResolver(auth, client)
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				log.Println("deploy Resolver Err:", err)
				return resolverAddr, resolverInstance, err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > checkTxRetryCount {
				log.Println("deploy Resolver transaction fails:", err)
				return resolverAddr, resolverInstance, err
			}
			time.Sleep(time.Minute)
			continue
		}

		return resAddr, resolverInstance, nil
	}
}

// AddToResolver adds
// ownerAddress is according to hexKey
func AddToResolver(ownerAddress, addAddr common.Address, hexKey string, resolverInstance *resolver.Resolver) error {
	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSA err: ", err)
		return err
	}

	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		tx, err := resolverInstance.Add(auth, addAddr)
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				log.Println("add to resolver err:", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > checkTxRetryCount {
				log.Println("add to resolver transaction fails: ", err)
				return err
			}
			continue
		}

		addrGetted, err := GetAddrFromResolver(ownerAddress, ownerAddress, resolverInstance)
		if err != nil {
			time.Sleep(time.Minute)
			continue
		}

		if addrGetted == addAddr {
			return nil
		}

		if retryCount > 20 {
			return ErrNotDeployedIndexer
		}
	}
}

// GetAddrFromResolver gets addr from resolver
func GetAddrFromResolver(localAddress common.Address, ownerAddress common.Address, resolverInstance *resolver.Resolver) (common.Address, error) {
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
			log.Println("get empty addr from resolver")
			return mapperAddr, ErrEmpty
		}
		return mapperAddr, nil
	}
}

// DeployMapper deploy a new mapper
func DeployMapper(localAddress common.Address, hexKey string) (common.Address, *mapper.Mapper, error) {
	var mapperAddr common.Address
	var mapperInstance *mapper.Mapper

	client := GetClient(EndPoint)
	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("Hex To ECDSA err: ", err)
		return mapperAddr, mapperInstance, err
	}
	// 部署Mapper
	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		mapperAddr, tx, mapperInstance, err := mapper.DeployMapperToResolver(auth, client)
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				log.Println("deploy Mapper to indexer Err:", err)
				return mapperAddr, mapperInstance, err
			}
			time.Sleep(30 * time.Second)
			continue
		}

		//检查交易
		err = CheckTx(tx)
		if err != nil {
			log.Println("DeployMapper transaction fails:", err)
			if retryCount > checkTxRetryCount {
				return mapperAddr, mapperInstance, err
			}
			continue
		}

		return mapperAddr, mapperInstance, nil
	}
}

func AddToMapper(localAddress, addr common.Address, hexKey string, mapperInstance *mapper.Mapper) error {
	key, _ := crypto.HexToECDSA(hexKey)

	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(key)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		tx, err := mapperInstance.Add(auth, addr)
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				log.Println("add addr to Mapper Err:", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > checkTxRetryCount {
				log.Println("add to mapper fails", err)
				return err
			}
			continue
		}

		if retryCount > 20 {
			return errors.New("add address to mapper fail")
		}

		addrGetted, err := GetAddrsFromMapper(localAddress, mapperInstance)
		if err != nil {
			time.Sleep(30 * time.Second)
			continue
		}

		length := len(addrGetted)
		if length != 0 && addrGetted[length-1] == addr {
			return nil
		}
	}
}

// GetAddrsFromMapper gets
func GetAddrsFromMapper(localAddress common.Address, mapperInstance *mapper.Mapper) ([]common.Address, error) {
	retryCount := 0
	for {
		retryCount++
		channels, err := mapperInstance.Get(&bind.CallOpts{
			From: localAddress,
		})
		if err != nil {
			if retryCount > 20 {
				log.Println("get addr from mapper:", err)
				return nil, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		if len(channels) != 0 && channels[len(channels)-1].String() != InvalidAddr {
			return channels, nil
		}

		log.Println("get empty addr from mapper")
		return nil, ErrEmpty
	}
}

func GetLatestFromMapper(localAddress common.Address, mapperInstance *mapper.Mapper) (common.Address, error) {
	var addr common.Address
	addrs, err := GetAddrsFromMapper(localAddress, mapperInstance)
	if err != nil {
		return addr, err
	}
	return addrs[len(addrs)-1], nil
}

// GetResolver gets role indexer
func GetResolverAddr(localAddress common.Address, key string) (common.Address, common.Address, error) {
	var resAddr common.Address

	client := GetClient(EndPoint)
	adminIndexerAddr := common.HexToAddress(indexerHex)
	adminIndexer, err := indexer.NewIndexer(adminIndexerAddr, client)
	if err != nil {
		log.Println("new admin Indexer err: ", err)
		return resAddr, resAddr, err
	}

	resAddr, ownAddr, err := GetAddrFromIndexer(localAddress, key, adminIndexer)
	if err != nil {
		return resAddr, resAddr, err
	}

	return resAddr, ownAddr, nil
}

func GetResolver(localAddress common.Address, key string) (common.Address, *resolver.Resolver, error) {
	resAddr, _, err := GetResolverAddr(localAddress, key)
	if err != nil {
		return resAddr, nil, err
	}

	resInstance, err := resolver.NewResolver(resAddr, GetClient(EndPoint))
	if err != nil {
		return resAddr, nil, err
	}

	return resAddr, resInstance, nil
}

// GetRoleIndexer gets role indexer
func GetRoleIndexer(localAddress, userAddress common.Address) (common.Address, *indexer.Indexer, error) {
	var indexerAddr common.Address
	var indexerInstance *indexer.Indexer

	client := GetClient(EndPoint)
	adminIndexerAddr := common.HexToAddress(indexerHex)
	adminIndexer, err := indexer.NewIndexer(adminIndexerAddr, client)
	if err != nil {
		log.Println("new Indexer err: ", err)
		return indexerAddr, indexerInstance, err
	}

	key := userAddress.String()

	indexerAddr, ownAddr, err := GetAddrFromIndexer(localAddress, key, adminIndexer)
	if err != nil {
		return indexerAddr, indexerInstance, err
	}

	if ownAddr.String() != key {
		log.Println("set owner is: ", key, ", but got owner is: ", ownAddr.String())
		return indexerAddr, indexerInstance, ErrMisType
	}

	indexerInstance, err = indexer.NewIndexer(indexerAddr, client)
	if err != nil {
		return indexerAddr, indexerInstance, err
	}

	return indexerAddr, indexerInstance, nil
}

// getMapperFromResolver 返回已经部署的Mapper，若Mapper没部署则返回err
// 特别地，当在ChannelTimeOut()中被调用，则localAddress和ownerAddress都是userAddr；
// 当在CloseChannel()中被调用，则localAddress为providerAddr, ownerAddress为userAddr
func GetMapperFromIndexer(localAddress common.Address, key string, indexerInstance *indexer.Indexer) (common.Address, *mapper.Mapper, error) {
	mapperAddr, _, err := GetAddrFromIndexer(localAddress, key, indexerInstance)
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

func GetMapperFromResolver(localAddress common.Address, ownerAddress common.Address, resolverInstance *resolver.Resolver) (common.Address, *mapper.Mapper, error) {
	mapperAddr, err := GetAddrFromResolver(localAddress, ownerAddress, resolverInstance)
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

// GetMapperFromAdmin get mapper
// key is adminIndexer->resolver
// userAddr is resolver->mapper
// flag indicates set or not;
// when set: userAddr depends on hexKey
func GetMapperFromAdmin(localAddr, userAddr common.Address, key, hexKey string, flag bool) (common.Address, *mapper.Mapper, error) {
	var mapperAddr common.Address

	//获得userIndexer, key is userAddr.String()
	_, resInstance, err := GetResolverFromAdmin(localAddr, userAddr, key, hexKey, flag)
	if err != nil {
		return mapperAddr, nil, err
	}

	mapperAddr, mapperInstance, err := GetMapperFromResolver(localAddr, userAddr, resInstance)
	if err != nil {
		if !flag {
			return mapperAddr, nil, err
		}
		mapperAddr, mInstance, err := DeployMapper(localAddr, hexKey)
		if err != nil {
			log.Println("deploy mapper err:", err)
			return mapperAddr, nil, err
		}

		mapperInstance = mInstance

		err = AddToResolver(userAddr, mapperAddr, hexKey, resInstance)
		if err != nil {
			log.Println("add mapper to resolver err:", err)
			return mapperAddr, nil, err
		}
		return mapperAddr, mapperInstance, nil
	}

	return mapperAddr, mapperInstance, nil
}

func GetResolverFromAdmin(localAddr, userAddr common.Address, key, hexKey string, flag bool) (common.Address, *resolver.Resolver, error) {
	if hexKey == "" {
		flag = false
	}

	//获得userIndexer, key is userAddr.String()
	resolverAddr, resInstance, err := GetResolver(localAddr, key)
	if err == ErrMisType {
		return resolverAddr, nil, err
	}

	if err != nil {
		if !flag {
			return resolverAddr, nil, err
		}
		client := GetClient(EndPoint)
		adminIndexerAddr := common.HexToAddress(indexerHex)
		adminIndexer, err := indexer.NewIndexer(adminIndexerAddr, client)
		if err != nil {
			log.Println("New Admin Indexer Err: ", err)
			return resolverAddr, nil, err
		}

		resAddr, rInstance, err := DeployResolver(hexKey)
		if err != nil {
			log.Println("Deploy Role Indexer Err:", err)
			return resolverAddr, nil, err
		}

		log.Println("add resolver")
		err = AddToIndexer(localAddr, resAddr, key, hexKey, adminIndexer)
		if err != nil {
			log.Println("add Role Indexer Err:", err)
			return resolverAddr, nil, err
		}

		return resAddr, rInstance, nil
	}

	return resolverAddr, resInstance, nil
}

// GetMapperFromAdminV1 get mapper
// userAddr is adminIndexer->indexer
// key is indexer->mapper
// flag indicates set or not
func GetMapperFromAdminV1(localAddr, userAddr common.Address, key, hexKey string, flag bool) (common.Address, *mapper.Mapper, error) {
	var mapperAddr common.Address

	if hexKey == "" {
		flag = false
	}

	//获得userIndexer, key is userAddr.String()
	_, indexerInstance, err := GetRoleIndexer(localAddr, userAddr)
	if err == ErrMisType {
		return mapperAddr, nil, err
	}

	if err != nil {
		if !flag {
			return mapperAddr, nil, err
		}
		client := GetClient(EndPoint)
		adminIndexerAddr := common.HexToAddress(indexerHex)
		adminIndexer, err := indexer.NewIndexer(adminIndexerAddr, client)
		if err != nil {
			log.Println("New Admin Indexer Err: ", err)
			return mapperAddr, nil, err
		}

		indexerAddr, iInstance, err := DeployIndexer(hexKey)
		if err != nil {
			log.Println("Deploy Role Indexer Err:", err)
			return mapperAddr, nil, err
		}

		indexerInstance = iInstance

		log.Println("add Role Indexer")
		err = AddToIndexer(localAddr, indexerAddr, userAddr.String(), hexKey, adminIndexer)
		if err != nil {
			log.Println("add Role Indexer Err:", err)
			return mapperAddr, nil, err
		}
	}

	mapperAddr, mapperInstance, err := GetMapperFromIndexer(localAddr, key, indexerInstance)
	if err != nil {
		if !flag {
			return mapperAddr, nil, err
		}
		mapperAddr, mInstance, err := DeployMapper(localAddr, hexKey)
		if err != nil {
			log.Println("deploy mapper err:", err)
			return mapperAddr, nil, err
		}

		mapperInstance = mInstance

		err = AddToIndexer(localAddr, mapperAddr, key, hexKey, indexerInstance)
		if err != nil {
			log.Println("add mapper to indexer err:", err)
			return mapperAddr, nil, err
		}
		return mapperAddr, mapperInstance, nil
	}

	return mapperAddr, mapperInstance, nil
}

//CheckTx 通过交易详情检查交易是否成功
func CheckTx(tx *types.Transaction) error {
	log.Println("Check Tx hash:", tx.Hash().Hex())

	var receipt *types.Receipt
	for i := 0; i < 60; i++ {
		receipt = GetTransactionReceipt(tx.Hash())
		if receipt != nil {
			break
		}
		time.Sleep(30 * time.Second)
	}

	if receipt == nil { //30分钟获取不到交易信息，判定交易失败
		log.Println("transaction fails")
		return ErrTxFail
	}

	if receipt.Status == 0 { //等于0表示交易失败，等于1表示成功
		log.Println("Transaction mined but execution failed")
		return ErrTxFail
	}

	return nil
}

//GetTransactionReceipt 通过交易hash获得交易详情
func GetTransactionReceipt(hash common.Hash) *types.Receipt {
	client, err := ethclient.Dial(EndPoint)
	if err != nil {
		log.Fatal("rpc.Dial err", err)
	}
	receipt, err := client.TransactionReceipt(context.Background(), hash)
	return receipt
}

//GetLogs filter logs according to
func GetLogs(restrictAddress []common.Address, fromBlock, toBlock *big.Int) ([]types.Log, error) {
	log.Println("begin to filter logs in chain...")

	client, err := ethclient.Dial(EndPoint)
	if err != nil {
		log.Println("rpc.Dial err", err)
		return nil, err
	}

	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Addresses: restrictAddress,
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Println("filterLogs err:", err)
		return nil, err
	}

	return logs, nil
}

//GetStorageIncome filter upkeeping-contract Pay-logs to calculate provider's income
func GetStorageIncome(restrictAddress []common.Address, providerAddr common.Address, fromBlock, toBlock int64) (*big.Int, []types.Log, error) {
	log.Println("begin to filter upkeeping Pay logs in chain...")

	totalIncome := big.NewInt(0)

	logs, err := GetLogs(restrictAddress, big.NewInt(fromBlock), big.NewInt(toBlock))
	if err != nil {
		return totalIncome, nil, err
	}

	contractAbi, err := abi.JSON(strings.NewReader(string(upKeeping.UpKeepingABI)))
	if err != nil {
		log.Println("abi json err:", err)
		return totalIncome, nil, err
	}

	logPaySignHash := crypto.Keccak256Hash([]byte("Pay(address,address,uint256)"))

	var resLogs []types.Log

	for _, vLog := range logs {
		if vLog.Topics[0].Hex() == logPaySignHash.Hex() && common.HexToAddress(vLog.Topics[2].Hex()).Hex() == providerAddr.Hex() {
			var payLog LogPay
			err := contractAbi.Unpack(&payLog, "Pay", vLog.Data)
			if err != nil {
				log.Println("unpack log err: ", err)
				return totalIncome, nil, err
			}

			totalIncome.Add(totalIncome, payLog.Value)

			resLogs = append(resLogs, vLog)
		}
	}
	return totalIncome, resLogs, nil
}

//GetReadIncome filter channel-contract CloseChannel-logs to calculate provider's income
func GetReadIncome(restrictAddress []common.Address, providerAddr common.Address, fromBlock, toBlock int64) (*big.Int, []types.Log, error) {
	log.Println("begin to filter channel closeChannel logs in chain...")

	totalIncome := big.NewInt(0)

	logs, err := GetLogs(restrictAddress, big.NewInt(fromBlock), big.NewInt(toBlock))
	if err != nil {
		return totalIncome, nil, err
	}

	contractAbi, err := abi.JSON(strings.NewReader(string(channel.ChannelABI)))
	if err != nil {
		log.Println("abi json err:", err)
		return totalIncome, nil, err
	}

	logCloseChannelSignHash := crypto.Keccak256Hash([]byte("closeChannel(address,uint256)"))

	var resLogs []types.Log

	for _, vLog := range logs {
		if vLog.Topics[0].Hex() == logCloseChannelSignHash.Hex() && common.HexToAddress(vLog.Topics[1].Hex()).Hex() == providerAddr.Hex() {
			var channelLog LogCloseChannel
			err := contractAbi.Unpack(&channelLog, "closeChannel", vLog.Data)
			if err != nil {
				log.Println("unpack log err: ", err)
				return totalIncome, nil, err
			}

			totalIncome.Add(totalIncome, channelLog.Value)

			resLogs = append(resLogs, vLog)
		}
	}
	return totalIncome, resLogs, nil
}

//GetBlockTime get block's timeStamp
func GetBlockTime(blockHash common.Hash) (uint64, error) {
	client, err := ethclient.Dial(EndPoint)
	if err != nil {
		return 0, err
	}

	blockHeader, err := client.HeaderByHash(context.Background(), blockHash)
	time := blockHeader.Time
	return time, nil
}

func isToday(t int64) bool {
	currentTime := time.Now()
	startTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
	oneDay := int64(24 * 60 * 60)
	if t >= startTime.Unix() && t <= startTime.Unix()+oneDay {
		return true
	}
	return false
}
