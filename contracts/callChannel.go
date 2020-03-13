package contracts

import (
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-mefs/contracts/channel"
)

//DeployChannelContract deploy channel-contract, timeOut's unit is second
func DeployChannelContract(hexKey string, userAddress, queryAddress, providerAddress common.Address, timeOut *big.Int, moneyToChannel *big.Int, redo bool) (common.Address, error) {
	var channelAddr common.Address

	key := queryAddress.String() + "channel" + providerAddress.String()

	_, mapperInstance, err := GetMapperFromAdmin(userAddress, userAddress, key, hexKey, true)
	if err != nil {
		return channelAddr, err
	}

	if !redo {
		channelAddr, err = GetLatestFromMapper(userAddress, mapperInstance)
		if err == nil {
			return channelAddr, nil
		}
	}

	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSA err: ", err)
		return channelAddr, err
	}

	client := GetClient(EndPoint)

	//本user与指定的provider部署channel合约
	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(sk)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		auth.Value = moneyToChannel //放进合约里的钱
		cAddr, tx, _, err := channel.DeployChannel(auth, client, providerAddress, timeOut)
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				log.Println("deploy Channel Err:", err)
				return channelAddr, err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > checkTxRetryCount {
				log.Println("deploy channel transaction fails", err)
				return channelAddr, err
			}
			continue
		}

		channelAddr = cAddr
		break
	}

	//将channel合约地址channelAddr放进上述的mapper中
	err = AddToMapper(userAddress, channelAddr, hexKey, mapperInstance)
	if err != nil {
		return channelAddr, nil
	}

	log.Println("channel-contract with", providerAddress.String(), "have been successfully deployed!")
	return channelAddr, nil
}

//GetChannelAddrs get the channel contract's address
func GetChannelAddrs(localAddress, userAddress, providerAddress, queryAddress common.Address) ([]common.Address, error) {
	key := queryAddress.String() + "channel" + providerAddress.String()
	_, mapperInstance, err := GetMapperFromAdmin(localAddress, userAddress, key, "", false)
	if err != nil {
		return nil, err
	}

	return GetAddrsFromMapper(localAddress, mapperInstance)
}

//GetLatestChannel get the channel contract's address
func GetLatestChannel(localAddress, userAddress, providerAddress, queryAddress common.Address) (common.Address, *channel.Channel, error) {
	var channelAddr common.Address
	key := queryAddress.String() + "channel" + providerAddress.String()
	_, mapperInstance, err := GetMapperFromAdmin(localAddress, userAddress, key, "", false)
	if err != nil {
		return channelAddr, nil, err
	}

	channelAddr, err = GetLatestFromMapper(localAddress, mapperInstance)
	if err != nil {
		return channelAddr, nil, err
	}

	channelInstance, err := channel.NewChannel(channelAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("getChannelsErr:", err)
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

	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(key)
		auth.GasPrice = big.NewInt(defaultGasPrice)

		tx, err := channelInstance.ChannelTimeout(auth)
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				log.Println("channelTimeOutErr:", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > checkTxRetryCount {
				log.Println("close channel fails", err)
				return err
			}
			continue
		}
		break
	}

	log.Println("you have called ChannelTimeout successfully!")
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

	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(key)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		auth.GasLimit = 8000000
		tx, err := channelInstance.CloseChannel(auth, hashNew, value, sig)
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				log.Println("closeChannelErr:", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > checkTxRetryCount {
				log.Println("close channel fails", err)
				return err
			}
			continue
		}
		break
	}

	log.Println("you have called CloseChannel successfully!")
	return nil
}
