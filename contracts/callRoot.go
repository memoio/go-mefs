package contracts

import (
	"encoding/hex"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-mefs/contracts/root"
)

//DeployRoot deploy Root contracts fot users
func DeployRoot(hexKey string, userAddress, queryAddress common.Address, redo bool) (common.Address, error) {
	log.Println("begin deploy root contract...")

	var ukAddr common.Address

	_, mapperInstance, err := GetMapperFromAdmin(userAddress, userAddress, "root", hexKey, true)
	if err != nil {
		return ukAddr, err
	}

	if !redo {
		ukAddr, err = GetLatestFromMapper(userAddress, mapperInstance)
		if err == nil {
			return ukAddr, nil
		}
	}

	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSAErr:", err)
		return ukAddr, err
	}

	// 部署UpKeeping
	// 用户需要支付的金额
	client := GetClient(EndPoint)
	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(key)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		// 用户地址,keeper地址数组,provider地址数组,存储时长 单位 天,存储大小 单位 MB
		ukAddress, tx, _, err := root.DeployRoot(auth, client, queryAddress)
		if err != nil {
			if retryCount > 5 {
				log.Println("deploy root contract fails:", err)
				return ukAddr, err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > 20 {
				log.Println("deploy root transaction fails", err)
				return ukAddr, err
			}
			continue
		}
		ukAddr = ukAddress
		break
	}

	//uk放进mapper
	err = AddToMapper(userAddress, ukAddr, hexKey, mapperInstance)
	if err != nil {
		log.Println("add root contract addr Err:", err)
		return ukAddr, err
	}
	log.Println("root-contract have been successfully deployed!")
	return ukAddr, nil
}

//GetRootAddrs get all upKeeping address
func GetRootAddrs(localAddress, userAddress common.Address) ([]common.Address, error) {
	//获得userIndexer, key is userAddr
	_, mapperInstance, err := GetMapperFromAdmin(localAddress, userAddress, "root", "", false)
	if err != nil {
		return nil, err
	}

	return GetAddrsFromMapper(localAddress, mapperInstance)
}

//GetRoot get root-contract from the mapper, and get the mapper from user's indexer
func GetRoot(localAddress, userAddress common.Address, key string) (ukaddr common.Address, uk *root.Root, err error) {
	//获得userIndexer, key is userAddr
	_, mapperInstance, err := GetMapperFromAdmin(localAddress, userAddress, "root", "", false)
	if err != nil {
		return ukaddr, nil, err
	}

	uks, err := GetAddrsFromMapper(localAddress, mapperInstance)
	if err != nil {
		return ukaddr, uk, err
	}

	client := GetClient(EndPoint)

	if key == "latest" {
		ukaddr = uks[len(uks)-1]
		uk, err := root.NewRoot(ukaddr, client)
		if err != nil {
			log.Println("new root Err:", err)
			return ukaddr, uk, err
		}
		return ukaddr, uk, nil
	}

	for _, ukAddr := range uks {
		ukaddr = ukAddr
		retryCount := 0
		for {
			retryCount++
			if retryCount > 10 {
				log.Println("Get Root Info err: ", err)
				break
			}

			uk, err = root.NewRoot(ukaddr, client)
			if err != nil {
				continue
			}
			queryAddr, err := uk.QueryAddr(&bind.CallOpts{
				From: localAddress,
			})
			if err != nil {
				time.Sleep(60 * time.Second)
				continue
			}

			if queryAddr.String() == key {
				return ukaddr, uk, nil
			}
			break
		}
	}

	return ukaddr, uk, ErrEmpty
}

// SetMerkleRoot sets Merkle root
func SetMerkleRoot(hexKey string, rootAddr common.Address, key int64, value [32]byte) error {
	uk, err := root.NewRoot(rootAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("new root Err:", err)
		return err
	}

	skey, _ := crypto.HexToECDSA(hexKey)
	retryCount := 0
	for {
		retryCount++
		auth := bind.NewKeyedTransactor(skey)
		auth.GasPrice = big.NewInt(defaultGasPrice)
		auth.GasLimit = spaceTimePayGasLimit

		tx, err := uk.SetRoot(auth, key, value)
		if err != nil {
			if retryCount > 5 {
				log.Println("set root Err:", err)
				return err
			}
			time.Sleep(time.Minute)
			continue
		}

		err = CheckTx(tx)
		if err != nil {
			if retryCount > 20 {
				log.Println("set merkle root transaction fails", err)
				return err
			}
			continue
		}

		// need async check, how?
		break
	}
	return nil
}

// GetMerkleRoot gets Merkle root
func GetMerkleRoot(localAddress, rootAddr common.Address, key int64) ([32]byte, error) {
	var value [32]byte
	uk, err := root.NewRoot(rootAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("new root Err:", err)
		return value, err
	}

	retryCount := 0
	for {
		retryCount++
		res, err := uk.GetRoot(&bind.CallOpts{
			From: localAddress,
		}, key)
		if err != nil {
			if retryCount > 5 {
				log.Println("get merkel root Err:", err)
				return value, err
			}
			time.Sleep(time.Minute)
			continue
		}

		if hex.EncodeToString(res[:]) == "0000000000000000000000000000000000000000000000000000000000000000" {
			return res, ErrEmpty
		}

		return res, err
	}
}

// GetMerkleKeys gets Merkle keys
func GetMerkleKeys(localAddress, rootAddr common.Address) ([]int64, error) {
	uk, err := root.NewRoot(rootAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("new root Err:", err)
		return nil, err
	}

	retryCount := 0
	for {
		retryCount++
		res, err := uk.GetAllKey(&bind.CallOpts{
			From: localAddress,
		})
		if err != nil {
			if retryCount > 5 {
				log.Println("get merkel keys err:", err)
				return nil, err
			}
			time.Sleep(time.Minute)
			continue
		}

		if len(res) == 0 {
			return nil, ErrEmpty
		}

		return res, err
	}
}

// GetLatestMerkleRoot gets Merkle latest root
func GetLatestMerkleRoot(localAddress, rootAddr common.Address) (int64, [32]byte, error) {
	var val [32]byte
	uk, err := root.NewRoot(rootAddr, GetClient(EndPoint))
	if err != nil {
		log.Println("new root Err:", err)
		return 0, val, err
	}

	retryCount := 0
	for {
		retryCount++
		resKey, resVal, err := uk.GetLatest(&bind.CallOpts{
			From: localAddress,
		})
		if err != nil {
			if retryCount > 5 {
				log.Println("get merkel key Err:", err)
				return 0, val, err
			}
			time.Sleep(time.Minute)
			continue
		}

		if resKey == 0 {
			return 0, val, ErrEmpty
		}
		return resKey, resVal, err
	}
}
