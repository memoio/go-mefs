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
	"github.com/memoio/go-mefs/contracts/upKeeping"
	"github.com/memoio/go-mefs/utils"
)

// EndPoint config中的ETH，在daemon中赋值
var EndPoint string

const (
	//indexerHex indexerAddress, it is well known
	indexerHex = "0xA36D0F4e56b76B89532eBbca8108d90d8cA006c2"
	//InvalidAddr implements invalid contracts-address
	InvalidAddr               = "0x0000000000000000000000000000000000000000"
	spaceTimePayGasLimit      = uint64(8000000)
	spaceTimePayGasPrice      = 2 * defaultGasPrice
	defaultGasPrice           = 200
	defaultGasLimit           = uint64(8000000)
	sendTransactionRetryCount = 5
	checkTxRetryCount         = 8
	checkTxSleepTime          = 5
	retryTxSleepTime          = time.Minute
	retryGetInfoSleepTime     = time.Minute
	waitTime                  = 3 * time.Second
)

const (
	keeperKey   = "keeperV0"
	providerKey = "providerV0"
	kpMapKey    = "kpMapV0"

	offerKey   = "offerV0"
	queryKey   = "queryV0"
	ukey       = "upKeepingV0"
	rootKey    = "rootV0"
	channelKey = "channelV0"
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
	ErrNewContractInstance = errors.New("new contract Instance failed")
	// ErrNotDeployedKPMap is
	ErrNotDeployedKPMap = errors.New("has not deployed keeperProviderMap contract")
	ErrTxFail           = errors.New("transaction fails")
	ErrTxExecu          = errors.New("Transaction mined but execution failed")
	ErrNotKeeper        = errors.New("addr is not a keeper")
	ErrNotProvider      = errors.New("addr is not a provider")
	ErrNotEnoughBalance = errors.New("balance is insufficient")
	ErrNotRight         = errors.New("the results in the contract don't match expectations")
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
	EndPoint = "http://119.147.213.219:8101"
}

//GetClient get rpc-client based the endPoint
func GetClient(endPoint string) *ethclient.Client {
	client, err := rpc.Dial(endPoint)
	if err != nil {
		log.Println(err)
	}
	return ethclient.NewClient(client)
}

//MakeAuth make the transactOpts to call contract
func MakeAuth(hexSk string, moneyToContract, nonce, gasPrice *big.Int, gasLimit uint64) (*bind.TransactOpts, error) {
	auth := &bind.TransactOpts{}
	sk, err := crypto.HexToECDSA(hexSk)
	if err != nil {
		log.Println("HexToECDSA err: ", err)
		return auth, err
	}

	auth = bind.NewKeyedTransactor(sk)
	auth.GasPrice = gasPrice
	auth.Value = moneyToContract //放进合约里的钱
	auth.Nonce = nonce
	auth.GasLimit = gasLimit
	return auth, nil
}

//QueryBalance query the balance of account
func QueryBalance(account string) (*big.Int, error) {
	var result string

	client, err := rpc.Dial(EndPoint)
	if err != nil {
		log.Println("rpc.dial err:", err)
		return big.NewInt(0), err
	}

	retryCount := 0
	for {
		retryCount++

		err = client.Call(&result, "eth_getBalance", account, "latest")

		if err != nil {
			if retryCount > sendTransactionRetryCount {
				return big.NewInt(0), err
			}
			time.Sleep(retryGetInfoSleepTime)
			continue
		}
		balance := utils.HexToBigInt(result)
		return balance, nil
	}
}

//GetLatestBlock get latest block from chain
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

// GetMemoPrice gets memo price
func GetMemoPrice() *big.Float {
	return big.NewFloat(utils.Memo2Dollar)
}

//CheckTx 通过交易详情检查交易是否成功
func CheckTx(tx *types.Transaction) error {
	log.Println("Check Tx hash:", tx.Hash().Hex(), "nonce:", tx.Nonce(), "gasPrice:", tx.GasPrice())

	var receipt *types.Receipt
	for i := 0; i < 10; i++ {
		receipt = GetTransactionReceipt(tx.Hash())
		if receipt != nil {
			break
		}
		t := checkTxSleepTime * (i + 1)
		time.Sleep(time.Duration(t) * time.Second)
	}

	if receipt == nil { //5分钟获取不到交易信息，判定交易失败
		return ErrTxFail
	}

	if receipt.Status == 0 { //等于0表示交易失败，等于1表示成功
		log.Println("Transaction mined but execution failed")
		txReceipt, err := receipt.MarshalJSON()
		if err != nil {
			return err
		}
		log.Println("TxReceipt:", string(txReceipt))
		return ErrTxExecu
	}

	log.Println("GasUsed:", receipt.GasUsed, "CumulativeGasUsed:", receipt.CumulativeGasUsed)

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
