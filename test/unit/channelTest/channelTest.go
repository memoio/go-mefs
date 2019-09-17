package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"flag"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	config "github.com/memoio/go-mefs/config"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/contracts/indexer"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

const (
	userAddr = "0x208649111Fd9253B76950e9f827a5A6dd616340d"
	userSk   = "8f9eb151ffaebf2fe963e6185f0d1f8c1e8397e5905b616958d765e7753329ea"
	moneyTo  = 1000000000000000
)

var ethEndPoint string

func main() {
	eth := flag.String("eth", "http://212.64.28.207:8101", "eth api address")
	flag.Parse()
	ethEndPoint = *eth

	contracts.EndPoint = ethEndPoint

	balance := queryBalance(userAddr)
	if balance.Cmp(big.NewInt(moneyTo)) <= 0 {
		transferTo(big.NewInt(moneyTo), userAddr)
	}

	for {
		time.Sleep(30 * time.Second)
		balance := queryBalance(userAddr)
		if balance.Cmp(big.NewInt(moneyTo)) >= 0 {
			break
		}

		log.Println(userAddr, "'s Balance now:", balance.String(), ", waiting for transfer success")
	}

	if err := testCloseChannel(); err != nil {
		log.Fatal(err)
	}

	if err := testChannelTimeout(); err != nil {
		log.Fatal(err)
	}
}

func testChannelTimeout() (err error) {
	log.Println("==========test channel timeout=========")
	proAddr, proSk, err := createAddr()
	if err != nil {
		log.Fatal("create provider fails")
	}

	indexerAddr := common.HexToAddress(contracts.IndexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, contracts.GetClient(contracts.EndPoint))
	if err != nil {
		log.Fatal("newIndexerErr:", err)
		return err
	}

	providerAddr := common.HexToAddress(proAddr[2:])

	log.Println("test deploy channel resolver")
	_, err = contracts.DeployResolverForChannel(proSk, providerAddr, indexer)
	if err != nil {
		log.Fatal("deployResolverErr:", err)
		return err
	}

	localAddr := common.HexToAddress(userAddr[2:])

	log.Println("test deploy channel contract")
	timeout := big.NewInt(5 * 60)
	moneyToChannel := big.NewInt(1000000)
	channelAddr, err := contracts.DeployChannelContract(userSk, localAddr, providerAddr, timeout, moneyToChannel)
	if err != nil {
		log.Fatal("deployChannelErr:", err)
		return err
	}

	log.Println("test query channel balance: ", channelAddr.String())
	retryCount := 0
	cbalance := queryBalance(channelAddr.String())
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		cbalance = queryBalance(channelAddr.String()) //查看部署的channel合约的账户余额
		if cbalance.Cmp(big.NewInt(0)) == 0 {
			if retryCount > 20 {
				log.Fatal("channel contract has no balance")
			}
			continue
		}
		if cbalance.Cmp(moneyToChannel) != 0 {
			log.Fatal("channel contract has wrong balance")
		}
		break
	}

	log.Println("test query channel start date")
	_, err = contracts.GetChannelStartDate(localAddr, providerAddr, localAddr)
	if err != nil {
		log.Println("Get Channel StartDate Err: ", err)
		return err
	}

	log.Println("test channel timeout before enddate, should return err")
	//触发channelTimeout()
	err = contracts.ChannelTimeout(userSk, localAddr, providerAddr)
	if err == nil {
		log.Println("call channelTimeout success, but time is early")
		return err
	}

	log.Println("test channel timeout after enddate")
	time.Sleep(300 * time.Second)
	err = contracts.ChannelTimeout(userSk, localAddr, providerAddr)
	if err != nil {
		log.Println("call channelTimeout err:", err)
		return err
	}

	retryCount = 0
	nbalance := queryBalance(channelAddr.String())
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		nbalance = queryBalance(channelAddr.String()) //查看部署的channel合约的账户余额
		if nbalance.Cmp(cbalance) == 0 {
			if retryCount > 20 {
				log.Fatal("call channel timeout failed")
			}
			continue
		}

		if nbalance.Cmp(big.NewInt(0)) != 0 {
			log.Fatal("call channel timeout has wrong balance")
		}

		log.Println("test channel timeout success")

		break
	}

	return nil
}

func testCloseChannel() (err error) {
	log.Println("test close channel")
	proAddr, proSk, err := createAddr()
	if err != nil {
		log.Fatal("create provider fails")
	}

	indexerAddr := common.HexToAddress(contracts.IndexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, contracts.GetClient(contracts.EndPoint))
	if err != nil {
		log.Fatal("newIndexerErr:", err)
		return err
	}

	providerAddr := common.HexToAddress(proAddr[2:])

	log.Println("test deploy channel resolver")
	_, err = contracts.DeployResolverForChannel(userSk, providerAddr, indexer)
	if err != nil {
		log.Fatal("deployResolverErr:", err)
		return err
	}

	localAddr := common.HexToAddress(userAddr[2:])

	log.Println("test deploy channel contract")
	timeout := big.NewInt(5 * 60)
	moneyToChannel := big.NewInt(1000000)
	channelAddr, err := contracts.DeployChannelContract(userSk, localAddr, providerAddr, timeout, moneyToChannel)
	if err != nil {
		log.Fatal("deployChannelErr:", err)
		return err
	}

	log.Println("test query channel balance: ", channelAddr.String())
	retryCount := 0
	cbalance := queryBalance(channelAddr.String())
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		cbalance = queryBalance(channelAddr.String()) //查看部署的channel合约的账户余额
		if cbalance.Cmp(big.NewInt(0)) == 0 {
			if retryCount > 20 {
				log.Fatal("channel contract has no balance")
			}
			continue
		}
		if cbalance.Cmp(moneyToChannel) != 0 {
			log.Fatal("channel contract has wrong balance")
		}
		break
	}

	chanAddr, err := contracts.GetChannelAddr(localAddr, providerAddr, localAddr)
	if err != nil {
		log.Fatal("GetChannelAddr fails:", err)
		return err
	}

	if chanAddr.String() != channelAddr.String() {
		log.Fatal("Get Wrong ChannelAddr")
	}

	balance := queryBalance(userAddr) //查看账户余额
	//签名
	value := big.NewInt(11111)
	sig, err := contracts.SignForChannel(channelAddr, value, userSk)
	if err != nil {
		log.Fatal("SignForChannelErr:", err)
		return err
	}

	//账户验证签名
	pubKey, err := utils.GetCompressedPkFromHexSk(userSk)
	if err != nil {
		log.Fatal("GetCompressedPkFromHexSkErr:", err)
		return err
	}
	verify, err := contracts.VerifySig(pubKey, sig, channelAddr, value)
	if err != nil || !verify {
		log.Fatal("verifyErr:", err)
		return err
	}

	//provider触发CloseChannel()
	err = contracts.CloseChannel(proSk, providerAddr, localAddr, sig, value)
	if err != nil {
		log.Fatal("CloseChannelErr:", err)
		return err
	}

	retryCount = 0
	nbalance := queryBalance(channelAddr.String())
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		nbalance = queryBalance(channelAddr.String()) //查看部署的channel合约的账户余额
		if nbalance.Cmp(big.NewInt(0)) > 0 {
			if retryCount > 20 {
				log.Println("call close channel, balance has: ", nbalance.String())
				break
				//log.Fatal("channel contract has balance")
			}
			continue
		}
		log.Println("call close channel success: channel has no balance")
		break
	}

	retryCount = 0
	ubalance := queryBalance(userAddr)
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		ubalance = queryBalance(userAddr)
		if ubalance.Cmp(balance) == 0 {
			if retryCount > 20 {
				log.Fatal("channel contract has balance")
			}
			continue
		}

		ubalance.Add(ubalance, value)
		ubalance.Sub(ubalance, moneyToChannel)

		if ubalance.Cmp(balance) != 0 {
			log.Fatal("call close channel failed")
		}

		log.Println("call close channel success: user has refund his remain value")
		break
	}
	return nil
}

func createAddr() (string, string, error) {
	identity, err := config.CreateID(os.Stdout, 2048)
	if err != nil {
		return "", "", err
	}
	address, err := address.GetAddressFromID(identity.PeerID)
	if err != nil {
		return "", "", err
	}
	addressHex := address.Hex()
	skByteEth, err := utils.IPFSskToEthskByte(identity.PrivKey)
	if err != nil {
		return "", "", err
	}
	enc := make([]byte, len(skByteEth)*2)
	//对私钥进行十六进制编码，得到以太坊格式的私钥，此处不加上"0x"前缀
	hex.Encode(enc, skByteEth)

	balance := queryBalance(addressHex)
	if balance.Cmp(big.NewInt(moneyTo)) <= 0 {
		transferTo(big.NewInt(moneyTo), addressHex)
	}

	for {
		time.Sleep(30 * time.Second)
		balance := queryBalance(addressHex)
		if balance.Cmp(big.NewInt(moneyTo)) >= 0 {
			break
		}

		log.Println(addressHex, "'s Balance now:", balance.String(), ", waiting for transfer success")
	}

	return addressHex, string(enc), nil
}

func transferTo(value *big.Int, addr string) {
	client, err := ethclient.Dial(ethEndPoint)
	if err != nil {
		log.Println("rpc.Dial err", err)
		log.Fatal(err)
	}
	log.Println("ethclient.Dial success")

	privateKey, err := crypto.HexToECDSA("928969b4eb7fbca964a41024412702af827cbc950dbe9268eae9f5df668c85b4")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("crypto.HexToECDSA success")

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}
	log.Println("cast public key to ECDSA success")

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("client.PendingNonceAt success")
	gasLimit := uint64(21000) // in units

	gasPrice := big.NewInt(30000000000) // in wei (30 gwei)
	gasPrice, err = client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("client.SuggestGasPrice success")

	toAddress := common.HexToAddress(addr[2:])
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Println("client.NetworkID error,use the default chainID")
		chainID = big.NewInt(666)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("types.SignTx success")

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("transfer ", value.String(), "to", addr)
	log.Printf("tx sent: %s\n", signedTx.Hash().Hex())
}

func queryBalance(addr string) *big.Int {
	var result string
	client, err := rpc.Dial(ethEndPoint)
	if err != nil {
		log.Fatal("rpc.dial err:", err)
	}
	err = client.Call(&result, "eth_getBalance", addr, "latest")
	if err != nil {
		log.Fatal("client.call err:", err)
	}
	return utils.HexToBigInt(result)
}
