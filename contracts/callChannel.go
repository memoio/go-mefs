package contracts

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-mefs/contracts/channel"
	"github.com/memoio/go-mefs/utils"
)

// DeployResolverForChannel provider deploys mapper to save user's mapper
// provider owns second sesolverß
func DeployResolverForChannel(localAddress common.Address, hexKey string) (common.Address, error) {
	resolverAddr, resolverInstance, err := deployResolver(localAddress, hexKey, "channel")
	if err != nil {
		return resolverAddr, err
	}

	secondAddr, _, err := deployResolverToResolver(localAddress, resolverInstance, hexKey)
	if err != nil {
		log.Println("deploy resolver for Channel Err:", err)
		return secondAddr, err
	}

	return secondAddr, nil
}

//DeployChannelContract deploy channel-contract, timeOut's unit is second
func DeployChannelContract(hexKey string, localAddress common.Address, providerAddress common.Address, timeOut *big.Int, moneyToChannel *big.Int) (common.Address, error) {
	var channelAddr common.Address

	key, _ := crypto.HexToECDSA(hexKey)

	client := GetClient(EndPoint)

	//根据key(provider的地址)从indexer中获得对应的resolver
	_, resolverInstance, err := getResolverFromIndexer(localAddress, "channel")
	if err != nil {
		fmt.Println("getResolverErr:", err)
		return channelAddr, err
	}

	_, secondInstance, err := getResolverFromResolver(localAddress, providerAddress, resolverInstance)
	if err != nil {
		return channelAddr, err
	}

	//本user与指定的provider部署channel合约
	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(key)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		auth.Value = moneyToChannel //放进合约里的钱
		channelAddr, _, _, err = channel.DeployChannel(auth, client, providerAddress, timeOut)
		if err != nil {
			if retryCount > 5 {
				fmt.Println("deploy Channel Err:", err)
				return channelAddr, err
			}
			time.Sleep(time.Minute)
			continue
		}
		break
	}

	//从上面的resolver中，获得本user的mapper，如果没有，则部署mapper
	// user owns mapper
	_, mapperInstance, err := deployMapper(localAddress, localAddress, secondInstance, hexKey)
	if err != nil {
		log.Println("deploy Mapper for Channel Err:", err)
		return channelAddr, err
	}

	//将channel合约地址channelAddr放进上述的mapper中
	err = addToMapper(localAddress, mapperInstance, channelAddr, hexKey)
	if err != nil {
		return channelAddr, nil
	}

	fmt.Println("channel-contract with", providerAddress.String(), "have been successfully deployed!")
	return channelAddr, nil
}

//ChannelTimeout called by user to discontinue the channel-contract
func ChannelTimeout(localAddress common.Address, providerAddress common.Address, hexKey string) (err error) {
	_, channelInstance, err := GetChannelAddr(localAddress, providerAddress, localAddress)
	if err != nil {
		return nil
	}

	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	_, err = channelInstance.ChannelTimeout(auth)
	if err != nil {
		fmt.Println("channelTimeOutErr:", err)
		return err
	}

	fmt.Println("you have called ChannelTimeout successfully!")
	return nil
}

//CloseChannel called by provider to stop the channel-contract,the ownerAddress implements the mapper
func CloseChannel(localAddress common.Address, userAddress common.Address, hexKey string, sig []byte, value *big.Int) (err error) {
	channelAddr, channelInstance, err := GetChannelAddr(localAddress, localAddress, userAddress)
	if err != nil {
		return nil
	}

	//(channelAddress, value)的哈希值
	var hashNew [32]byte
	valueNew := common.LeftPadBytes(value.Bytes(), 32)
	hash := crypto.Keccak256(channelAddr.Bytes(), valueNew) //32Byte
	copy(hashNew[:], hash[:32])

	//用user的签名来触发closeChannel()
	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasPrice = big.NewInt(defaultGasPrice)
	auth.GasLimit = 8000000
	_, err = channelInstance.CloseChannel(auth, hashNew, value, sig)
	if err != nil {
		fmt.Println("closeChannelErr:", err)
		return err
	}

	fmt.Println("you have called CloseChannel successfully!")
	return nil
}

// getChannel()当在ChannelTimeOut()中被调用，则localAddress为userAddr；
// 当在CloseChannel()中被调用，则localAddress是providerAddr

//GetChannelAddr get the channel contract's address
func GetChannelAddr(localAddress, providerAddress, userAddress common.Address) (common.Address, *channel.Channel, error) {
	var channelAddr common.Address

	_, resolverInstance, err := getResolverFromIndexer(localAddress, "channel")
	if err != nil {
		fmt.Println("get Resolver Err:", err)
		return channelAddr, nil, err
	}

	_, secondInstance, err := getResolverFromResolver(localAddress, providerAddress, resolverInstance)
	if err != nil {
		fmt.Println("get second Resolver Err:", err)
		return channelAddr, nil, err
	}

	_, mapperInstance, err := getMapperInstance(localAddress, userAddress, secondInstance)
	if err != nil {
		return channelAddr, nil, err
	}

	channelAddr, err = getLatestAddrFromMapper(localAddress, mapperInstance)
	if err != nil {
		return channelAddr, nil, err
	}

	channelInstance, err := channel.NewChannel(channelAddr, GetClient(EndPoint))
	if err != nil {
		fmt.Println("getChannelsErr:", err)
		return channelAddr, nil, err
	}
	return channelAddr, channelInstance, nil
}

//SignForChannel user sends a private key signature to the provider
func SignForChannel(channelAddr common.Address, value *big.Int, hexKey string) (sig []byte, err error) {
	//(channelAddress, value)的哈希值
	valueNew := common.LeftPadBytes(value.Bytes(), 32)
	hash := crypto.Keccak256(channelAddr.Bytes(), valueNew) //32Byte

	//私钥格式转换
	skECDSA, err := utils.HexskToECDSAsk(hexKey)
	if err != nil {
		fmt.Println("HexskToECDSAskErr:", err)
		return sig, err
	}

	//私钥对上述哈希值签名
	sig, err = crypto.Sign(hash, skECDSA)
	if err != nil {
		fmt.Println("signForChannelErr:", err)
		return sig, err
	}
	return sig, nil
}

//VerifySig provider used to verify user's signature for channel-contract
func VerifySig(userPubKey, sig []byte, channelAddr common.Address, value *big.Int) (verify bool, err error) {
	//(channelAddress, value)的哈希值
	valueNew := common.LeftPadBytes(value.Bytes(), 32)
	hash := crypto.Keccak256(channelAddr.Bytes(), valueNew)

	//验证签名
	verify = crypto.VerifySignature(userPubKey, hash, sig[:64])
	return verify, nil
}

//GetChannelStartDate used to get the startDate of channel-contract
func GetChannelStartDate(localAddr, providerAddr, userAddr common.Address) (string, error) {
	_, channelContract, err := GetChannelAddr(localAddr, providerAddr, userAddr)
	if err != nil {
		return "", err
	}

	startDate, err := channelContract.GetStartDate(&bind.CallOpts{
		From: localAddr,
	})
	if err != nil {
		fmt.Println("GetStartDateErr:", err)
		return "", err
	}

	return utils.UnixToTime(startDate.Int64()).Format(utils.SHOWTIME), nil
}
