package contracts

import (
	"encoding/hex"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/memoio/go-mefs/contracts/root"
)

//RootNodeInfo  The basic information of node used for root contract
type RootNodeInfo struct {
	addr  common.Address //local address
	hexSk string         //local privateKey
}

//NewCR new a instance of contractRoot
func NewCRoot(addr common.Address, hexSk string) ContractRoot {
	RInfo := &RootNodeInfo{
		addr:  addr,
		hexSk: hexSk,
	}

	return RInfo
}

//DeployRoot deploy Root contracts fot users
func (r *RootNodeInfo) DeployRoot(queryAddress common.Address, redo bool) (common.Address, error) {
	var rtAddr, rtAddress common.Address

	ma := NewCManage(r.addr, r.hexSk)
	_, mapperInstance, err := ma.GetMapperFromAdmin(r.addr, rootKey, true)
	if err != nil {
		return rtAddr, err
	}

	if !redo {
		rtAddr, err = ma.GetLatestFromMapper(mapperInstance)
		if err == nil {
			return rtAddr, nil
		}
	}

	log.Println("begin deploy root contract...")
	client := getClient(EndPoint)
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := makeAuth(r.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return rtAddr, errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		rtAddress, tx, _, err = root.DeployRoot(auth, client, queryAddress)
		if rtAddress.String() != InvalidAddr {
			rtAddr = rtAddress
		}
		if err != nil {
			retryCount++
			log.Println("deploy Root Err:", err)
			if err.Error() == core.ErrNonceTooLow.Error() && auth.GasPrice.Cmp(big.NewInt(defaultGasPrice)) > 0 {
				log.Println("previously pending transaction has successfully executed")
				break
			}
			if retryCount > sendTransactionRetryCount {
				return rtAddr, err
			}
			time.Sleep(retryTxSleepTime)
			continue
		}

		err = checkTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("deploy Root transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return rtAddr, err
			}
			continue
		}
		break
	}
	log.Println("root-contract", rtAddr.String(), "have been successfully deployed!")

	//uk放进mapper
	err = ma.AddToMapper(rtAddr, mapperInstance)
	if err != nil {
		log.Println("add root contract addr Err:", err)
		return rtAddr, err
	}
	return rtAddr, nil
}

//GetRootAddrs get all upKeeping address
func (r *RootNodeInfo) GetRootAddrs(userAddress common.Address) ([]common.Address, error) {
	ma := NewCManage(r.addr, "")
	//获得userIndexer, key is userAddr
	_, mapperInstance, err := ma.GetMapperFromAdmin(userAddress, rootKey, false)
	if err != nil {
		return nil, err
	}

	return ma.GetAddressFromMapper(mapperInstance)
}

//GetRoot get root-contract from the mapper, and get the mapper from user's indexer
func (r *RootNodeInfo) GetRoot(userAddress common.Address, key string) (rtaddr common.Address, rt *root.Root, err error) {
	ma := NewCManage(r.addr, "")
	//获得userIndexer, key is userAddr
	_, mapperInstance, err := ma.GetMapperFromAdmin(userAddress, rootKey, false)
	if err != nil {
		return rtaddr, nil, err
	}

	rts, err := ma.GetAddressFromMapper(mapperInstance)
	if err != nil {
		return rtaddr, rt, err
	}

	client := getClient(EndPoint)

	if key == "latest" {
		rtaddr = rts[len(rts)-1]
		rt, err := root.NewRoot(rtaddr, client)
		if err != nil {
			log.Println("new root Err:", err)
			return rtaddr, rt, err
		}
		return rtaddr, rt, nil
	}

	for _, rtAddr := range rts {
		rtaddr = rtAddr
		retryCount := 0
		for {
			retryCount++
			if retryCount > 10 {
				log.Println("Get Root Info err: ", err)
				break
			}

			rt, err = root.NewRoot(rtaddr, client)
			if err != nil {
				continue
			}
			queryAddr, err := rt.QueryAddr(&bind.CallOpts{
				From: r.addr,
			})
			if err != nil {
				time.Sleep(retryGetInfoSleepTime)
				continue
			}

			if queryAddr.String() == key {
				return rtaddr, rt, nil
			}
			break
		}
	}

	return rtaddr, rt, ErrEmpty
}

// SetMerkleRoot sets Merkle root
func (r *RootNodeInfo) SetMerkleRoot(rootAddr common.Address, key int64, value [32]byte) error {
	rt, err := root.NewRoot(rootAddr, getClient(EndPoint))
	if err != nil {
		log.Println("new root Err:", err)
		return err
	}

	log.Println("begin set merkleRoot...")
	tx := &types.Transaction{}
	retryCount := 0
	checkRetryCount := 0
	for {
		auth, errMA := makeAuth(r.hexSk, nil, nil, big.NewInt(defaultGasPrice), defaultGasLimit)
		if errMA != nil {
			return errMA
		}

		if err == ErrTxFail && tx != nil {
			auth.Nonce = big.NewInt(int64(tx.Nonce()))
			auth.GasPrice = new(big.Int).Add(tx.GasPrice(), big.NewInt(defaultGasPrice))
			log.Println("rebuild transaction... nonce is ", auth.Nonce, " gasPrice is ", auth.GasPrice)
		}

		tx, err = rt.SetRoot(auth, key, value)
		if err != nil {
			retryCount++
			log.Println("set MerkleRoot Err:", err)
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

		err = checkTx(tx)
		if err != nil {
			checkRetryCount++
			log.Println("set MerkleRoot transaction fails", err)
			if checkRetryCount > checkTxRetryCount {
				return err
			}
			continue
		}
		break
	}

	log.Println("merkleRoot have been successfuly set!")
	return nil
}

// GetMerkleRoot gets Merkle root
func (r *RootNodeInfo) GetMerkleRoot(rootAddr common.Address, key int64) ([32]byte, error) {
	var value [32]byte
	rt, err := root.NewRoot(rootAddr, getClient(EndPoint))
	if err != nil {
		log.Println("new root Err:", err)
		return value, err
	}

	retryCount := 0
	for {
		retryCount++
		res, err := rt.GetRoot(&bind.CallOpts{
			From: r.addr,
		}, key)

		if err != nil {
			if retryCount > 5 {
				log.Println("get merkel root Err:", err)
				return res, err
			}
			time.Sleep(retryGetInfoSleepTime)
			continue
		}

		if hex.EncodeToString(res[:]) == "0000000000000000000000000000000000000000000000000000000000000000" {
			if retryCount > 5 {
				log.Println("get merkel root Err:", ErrEmpty)
				return res, ErrEmpty
			}
			time.Sleep(retryGetInfoSleepTime)
			continue
		}

		return res, nil
	}
}

// GetMerkleKeys gets Merkle keys
func (r *RootNodeInfo) GetMerkleKeys(rootAddr common.Address) ([]int64, error) {
	rt, err := root.NewRoot(rootAddr, getClient(EndPoint))
	if err != nil {
		log.Println("new root Err:", err)
		return nil, err
	}

	retryCount := 0
	for {
		retryCount++
		res, err := rt.GetAllKey(&bind.CallOpts{
			From: r.addr,
		})
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				log.Println("get merkel keys err:", err)
				return nil, err
			}
			time.Sleep(retryGetInfoSleepTime)
			continue
		}

		if len(res) == 0 {
			return nil, ErrEmpty
		}

		return res, nil
	}
}

// GetLatestMerkleRoot gets Merkle latest root
func (r *RootNodeInfo) GetLatestMerkleRoot(rootAddr common.Address) (int64, [32]byte, error) {
	var val [32]byte
	rt, err := root.NewRoot(rootAddr, getClient(EndPoint))
	if err != nil {
		log.Println("new root Err:", err)
		return 0, val, err
	}

	retryCount := 0
	for {
		retryCount++
		resKey, resVal, err := rt.GetLatest(&bind.CallOpts{
			From: r.addr,
		})
		if err != nil {
			if retryCount > sendTransactionRetryCount {
				log.Println("get merkel key Err:", err)
				return 0, val, err
			}
			time.Sleep(retryGetInfoSleepTime)
			continue
		}

		if resKey == 0 {
			return 0, val, ErrEmpty
		}
		return resKey, resVal, nil
	}
}
