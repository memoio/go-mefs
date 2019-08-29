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
	"github.com/memoio/go-mefs/contracts/mapper"
	"github.com/memoio/go-mefs/utils"
)

//DeployChannelContract deploy channel-contract, timeOut's unit is second
func DeployChannelContract(endPoint string, hexKey string, localAddress common.Address, providerAddress common.Address, timeOut *big.Int, moneyToChannel *big.Int) (common.Address, error) {
	fmt.Println("begin deploy channel-contract with", providerAddress.String(), "...")
	var channelAddr common.Address
	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	client := GetClient(endPoint)

	//根据key(provider的地址)从indexer中获得对应的resolver
	resolver, err := getResolverFromIndexer(endPoint, localAddress, providerAddress.String())
	if err != nil {
		fmt.Println("getResolverErr:", err)
		return channelAddr, err
	}

	//从上面的resolver中，获得本user的mapper，如果没有，则部署mapper
	auth = bind.NewKeyedTransactor(key)
	mapper, err := deployMapper(endPoint, localAddress, resolver, auth, client)
	if err != nil {
		return channelAddr, err
	}

	//本user与指定的provider部署channel合约
	auth = bind.NewKeyedTransactor(key)
	auth.Value = moneyToChannel //放进合约里的钱
	channelAddr, _, _, err = channel.DeployChannel(auth, client, providerAddress, timeOut)
	if err != nil {
		fmt.Println("deployChannelErr:", err)
		return channelAddr, err
	}

	log.Println("channelContractAddr:", channelAddr.String())

	//将channel合约地址channelAddr放进上述的mapper中
	auth = bind.NewKeyedTransactor(key)
	_, err = mapper.Add(auth, channelAddr)
	if err != nil {
		fmt.Println("addChannelErr:", err)
		return channelAddr, err
	}

	//验证channelAddr是否已经放进了mapper中
	for i := 1; ; i++ {
		if i%10 == 0 { //每隔10次如果还get不到合约地址，就再触发一次添加合约到mapper
			auth = bind.NewKeyedTransactor(key)
			_, err = mapper.Add(auth, channelAddr)
			if err != nil {
				fmt.Println("addChannelErr:", err)
				return channelAddr, err
			}
		}
		channelGetted, err := mapper.Get(&bind.CallOpts{
			From: localAddress,
		})
		if err != nil {
			fmt.Println("getChannelErr:", err)
			return channelAddr, err
		}
		length := len(channelGetted)
		if length != 0 && channelGetted[length-1] == channelAddr {
			break
		}
		time.Sleep(8 * time.Second)
	}

	fmt.Println("channel-contract with", providerAddress.String(), "have been successfully deployed!")
	return channelAddr, nil
}

//ChannelTimeout called by user to discontinue the channel-contract
func ChannelTimeout(endPoint string, hexKey string, localAddress common.Address, providerAddress common.Address) (err error) {
	key, _ := crypto.HexToECDSA(hexKey)

	resolver, err := getResolverFromIndexer(endPoint, localAddress, providerAddress.String())
	if err != nil {
		fmt.Println("getResolverErr:", err)
		return err
	}

	mapper, err := getMapperInstance(endPoint, localAddress, localAddress, resolver)
	if err != nil {
		return err
	}

	_, channelContract, err := getChannel(endPoint, mapper, localAddress)
	if err != nil {
		return err
	}

	auth := bind.NewKeyedTransactor(key)
	_, err = channelContract.ChannelTimeout(auth)
	if err != nil {
		fmt.Println("channelTimeOutErr:", err)
		return err
	}

	fmt.Println("you have called ChannelTimeout successfully!")
	return nil
}

//CloseChannel called by provider to stop the channel-contract,the ownerAddress implements the deployer
func CloseChannel(endPoint string, hexKey string, localAddress common.Address, ownerAddress common.Address, sig []byte, value *big.Int) (err error) {
	//获得channel合约地址
	resolver, err := getResolverFromIndexer(endPoint, localAddress, localAddress.String())
	if err != nil {
		fmt.Println("getResolverErr:", err)
		return err
	}
	mapper, err := getMapperInstance(endPoint, localAddress, ownerAddress, resolver)
	if err != nil {
		return err
	}
	channelAddr, channelContract, err := getChannel(endPoint, mapper, localAddress)
	if err != nil {
		return err
	}

	//(channelAddress, value)的哈希值
	var hashNew [32]byte
	valueNew := common.LeftPadBytes(value.Bytes(), 32)
	hash := crypto.Keccak256(channelAddr.Bytes(), valueNew) //32Byte
	copy(hashNew[:], hash[:32])

	//用user的签名来触发closeChannel()
	key, _ := crypto.HexToECDSA(hexKey)
	auth := bind.NewKeyedTransactor(key)
	auth.GasLimit = 8000000
	_, err = channelContract.CloseChannel(auth, hashNew, value, sig)
	if err != nil {
		fmt.Println("closeChannelErr:", err)
		return err
	}

	fmt.Println("you have called CloseChannel successfully!")
	return nil
}

//getChannel()当在ChannelTimeOut()中被调用，则localAddress为userAddr；
// 当在CloseChannel()中被调用，则localAddress是providerAddr
func getChannel(endPoint string, mapper *mapper.Mapper, localAddress common.Address) (channelAddr common.Address, channelContract *channel.Channel, err error) {
	channels, err := mapper.Get(&bind.CallOpts{
		From: localAddress,
	})
	if err != nil {
		fmt.Println("getChannelsErr:", err)
		return channelAddr, nil, err
	}
	if len(channels) == 0 {
		fmt.Println("getChannelErr:", ErrNotDeployedChannel)
	}

	//返回最新的channel地址
	channelAddr = channels[len(channels)-1]
	channelContract, err = channel.NewChannel(channelAddr, GetClient(endPoint))
	if err != nil {
		fmt.Println("getChannelsErr:", err)
		return channelAddr, nil, err
	}
	return channelAddr, channelContract, nil
}

//GetChannelAddr get the channel contract's address
func GetChannelAddr(endPoint string, localAddr, providerAddr, ownerAddr common.Address) (common.Address, error) {
	var ChannelAddr common.Address
	resolver, err := getResolverFromIndexer(endPoint, localAddr, providerAddr.String())
	if err != nil {
		return ChannelAddr, err
	}

	mapper, err := getMapperInstance(endPoint, localAddr, ownerAddr, resolver)
	if err != nil {
		return ChannelAddr, err
	}

	channelAddr, _, err := getChannel(endPoint, mapper, localAddr)
	if err != nil {
		return ChannelAddr, err
	}
	return channelAddr, nil
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
