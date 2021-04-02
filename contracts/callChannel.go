package contracts

import (
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-mefs/contracts/channel"
)

//ChannelInfo  The basic information of node used for channel contract
type ChannelNodeInfo struct {
	addr  common.Address //local address
	hexSk string         //local privateKey
}

//NewCH new a instance of contractChannel
func NewCH(addr common.Address, hexSk string) ContractChannel {
	ChInfo := &ChannelNodeInfo{
		addr:  addr,
		hexSk: hexSk,
	}

	return ChInfo
}

//DeployChannelContract deploy channel-contract, timeOut's unit is second
func (ch *ChannelNodeInfo) DeployChannelContract(queryAddress, providerAddress common.Address, timeOut *big.Int, moneyToChannel *big.Int, redo bool) (common.Address, error) {
	var channelAddr common.Address

	key := queryAddress.String() + channelKey + providerAddress.String()

	ma := NewCManage(ch.addr, ch.hexSk)
	_, mapperInstance, err := ma.GetMapperFromAdmin(ch.addr, key, true)
	if err != nil {
		return channelAddr, err
	}

	if !redo {
		channelAddr, err = ma.GetLatestFromMapper(mapperInstance)
		if err == nil {
			return channelAddr, nil
		}
	}

	log.Println("begin deploy channel contract...")
	client := GetClient(EndPoint)

	//本user与指定的provider部署channel合约
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	var cAddr common.Address
	for {
		auth, errMA := MakeAuth(ch.hexSk, moneyToChannel, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return channelAddr, errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		cAddr, tx, _, err = channel.DeployChannel(auth, client, providerAddress, timeOut)
		if cAddr.String() != InvalidAddr {
			channelAddr = cAddr
		}
		if err != nil {
			retryCount++
			log.Println("deploy Channel Err:", err)
			if err.Error() == core.ErrNonceTooLow.Error() && auth.GasPrice.Cmp(big.NewInt(defaultGasPrice)) > 0 {
				log.Println("previously pending transaction has successfully executed")
				break
			}
			if retryCount > sendTransactionRetryCount {
				return channelAddr, err
			}
			time.Sleep(retryTxSleepTime)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("deploy channel transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return channelAddr, err
			}
			continue
		}
		break
	}

	log.Println("channel contract", channelAddr.String(), "with", providerAddress.String(), "have been successfuly deployed!")

	//将channel合约地址channelAddr放进上述的mapper中
	err = ma.AddToMapper(channelAddr, mapperInstance)
	if err != nil {
		return channelAddr, err
	}

	return channelAddr, nil
}

func (ch *ChannelNodeInfo) GetChannelInfo(chanAddress common.Address) (int64, int64, common.Address, common.Address, error) {
	var sender, receiver common.Address
	var startDate, timeOut *big.Int
	channelInstance, err := channel.NewChannel(chanAddress, GetClient(EndPoint))
	if err != nil {
		return 0, 0, sender, receiver, err
	}
	retryCount := 0
	for {
		retryCount++
		startDate, timeOut, sender, receiver, err = channelInstance.GetInfo(&bind.CallOpts{
			From: ch.addr,
		})
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				return 0, 0, sender, receiver, err
			}
			time.Sleep(retryGetInfoSleepTime)
			continue
		}

		return startDate.Int64(), timeOut.Int64(), sender, receiver, nil
	}
}

//GetChannelAddrs get the channel contract's address
func (ch *ChannelNodeInfo) GetChannelAddrs(userAddress, providerAddress, queryAddress common.Address) ([]common.Address, error) {
	key := queryAddress.String() + channelKey + providerAddress.String()
	ma := NewCManage(ch.addr, ch.hexSk)
	_, mapperInstance, err := ma.GetMapperFromAdmin(userAddress, key, false)
	if err != nil {
		return nil, err
	}

	return ma.GetAddressFromMapper(mapperInstance)
}

//GetLatestChannel get the channel contract's address
func (ch *ChannelNodeInfo) GetLatestChannel(userAddress, providerAddress, queryAddress common.Address) (common.Address, *channel.Channel, error) {
	var channelAddr common.Address
	key := queryAddress.String() + channelKey + providerAddress.String()
	ma := NewCManage(ch.addr, ch.hexSk)
	_, mapperInstance, err := ma.GetMapperFromAdmin(userAddress, key, false)
	if err != nil {
		return channelAddr, nil, err
	}

	channelAddr, err = ma.GetLatestFromMapper(mapperInstance)
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
func (ch *ChannelNodeInfo) ChannelTimeout(channelAddress common.Address) (err error) {
	channelInstance, err := channel.NewChannel(channelAddress, GetClient(EndPoint))
	if err != nil {
		return err
	}

	log.Println("begin call channelTimeout...")
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(ch.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = channelInstance.ChannelTimeout(auth)
		if err != nil {
			retryCount++
			log.Println("channelTimeout Err:", err)
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
			log.Println("channelTimeout transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
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
func (ch *ChannelNodeInfo) CloseChannel(channelAddress common.Address, sig []byte, value *big.Int) (err error) {
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
	log.Println("begin call closeChannel...")
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(ch.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = channelInstance.CloseChannel(auth, hashNew, value, sig)
		if err != nil {
			retryCount++
			log.Println("closeChannel Err:", err)
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
			log.Println("closeChannel transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}

	log.Println("you have called CloseChannel successfully!")
	return nil
}

//ExtendChannelTime called by user to extend the time in channel contract
func (ch *ChannelNodeInfo) ExtendChannelTime(channelAddress common.Address, addTime *big.Int) error {
	channelInstance, err := channel.NewChannel(channelAddress, GetClient(EndPoint))
	if err != nil {
		return err
	}

	log.Println("begin call extendChannelTime...")
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := MakeAuth(ch.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = channelInstance.Extend(auth, addTime)
		if err != nil {
			retryCount++
			log.Println("extendChannelTime Err:", err)
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
			log.Println("extendChannelTime transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}

	log.Println("you have called extendChannelTime successfully!")
	return nil
}
