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

//DeployChannelContract deploy channel-contract, timeOut's unit is second
func DeployChannelContract(hexKey string, userAddress, queryAddress, providerAddress common.Address, timeOut *big.Int, moneyToChannel *big.Int, redo bool) (common.Address, error) {
	var channelAddr common.Address

	client := GetClient(EndPoint)

	//获得userIndexer, key is userAddr
	_, indexerInstance, err := GetRoleIndexer(userAddress, userAddress)
	if err != nil {
		fmt.Println("GetResolverErr:", err)
		return channelAddr, err
	}

	key := queryAddress.String() + "channel" + providerAddress.String()
	_, mapperInstance, err := DeployMapperToIndexer(userAddress, key, hexKey, indexerInstance)
	if err != nil {
		return channelAddr, err
	}

	if !redo {
		channelAddr, err = getLatestFromMapper(userAddress, mapperInstance)
		if err == nil {
			return channelAddr, nil
		}
	}

	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSA err: ", err)
		return channelAddr, err
	}

	//本user与指定的provider部署channel合约
	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		auth.Value = moneyToChannel //放进合约里的钱
		cAddr, tx, _, err := channel.DeployChannel(auth, client, providerAddress, timeOut)
		if err != nil {
			if retryCount > 5 {
				fmt.Println("deploy Channel Err:", err)
				return channelAddr, err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > 20 {
				log.Println("deploy channel transaction fails", err)
				return channelAddr, err
			}
			continue
		}

		channelAddr = cAddr
		break
	}

	//将channel合约地址channelAddr放进上述的mapper中
	err = addToMapper(userAddress, mapperInstance, channelAddr, hexKey)
	if err != nil {
		return channelAddr, nil
	}

	fmt.Println("channel-contract with", providerAddress.String(), "have been successfully deployed!")
	return channelAddr, nil
}

//GetChannelAddrs get the channel contract's address
func GetChannelAddrs(localAddress, userAddress, providerAddress, queryAddress common.Address) ([]common.Address, error) {
	//获得userIndexer, key is userAddr
	_, indexerInstance, err := GetRoleIndexer(userAddress, userAddress)
	if err != nil {
		fmt.Println("GetResolverErr:", err)
		return nil, err
	}

	key := queryAddress.String() + "channel" + providerAddress.String()
	_, mapperInstance, err := getMapperFromIndexer(localAddress, key, indexerInstance)
	if err != nil {
		return nil, err
	}

	return getAllFromMapper(localAddress, mapperInstance)
}

//GetLatestChannel get the channel contract's address
func GetLatestChannel(localAddress, userAddress, providerAddress, queryAddress common.Address) (common.Address, *channel.Channel, error) {
	var channelAddr common.Address

	//获得userIndexer, key is userAddr
	_, indexerInstance, err := GetRoleIndexer(userAddress, userAddress)
	if err != nil {
		fmt.Println("GetResolverErr:", err)
		return channelAddr, nil, err
	}

	key := queryAddress.String() + "channel" + providerAddress.String()
	_, mapperInstance, err := getMapperFromIndexer(localAddress, key, indexerInstance)
	if err != nil {
		return channelAddr, nil, err
	}

	channelAddr, err = getLatestFromMapper(localAddress, mapperInstance)
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

//ChannelTimeout called by user to discontinue the channel-contract
func ChannelTimeout(channelAddress common.Address, hexKey string) (err error) {
	channelInstance, err := channel.NewChannel(channelAddress, GetClient(EndPoint))
	if err != nil {
		return err
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
func CloseChannel(channelAddress common.Address, hexKey string, sig []byte, value *big.Int) (err error) {
	channelInstance, err := channel.NewChannel(channelAddress, GetClient(EndPoint))
	if err != nil {
		return err
	}
	//(channelAddress, value)的哈希值
	var hashNew [32]byte
	valueNew := common.LeftPadBytes(value.Bytes(), 32)
	hash := crypto.Keccak256(channelAddress.Bytes(), valueNew) //32Byte
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
