package provider

import (
	"errors"
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/proto"
	"github.com/memoio/go-mefs/contracts"
	pb "github.com/memoio/go-mefs/role/user/pb"
	blocks "github.com/memoio/go-mefs/source/go-block-format"
	cid "github.com/memoio/go-mefs/source/go-cid"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	b58 "github.com/mr-tron/base58/base58"
)

func handlePutBlock(km *metainfo.KeyMeta, value, from string) error {
	// key is cid|ops|type|begin|end
	splitedNcid := strings.Split(km.ToString(), metainfo.DELIMITER)
	if len(splitedNcid) == 0 {
		return errors.New("No value in km")
	}
	bmeta, err := metainfo.GetBlockMeta(splitedNcid[0])
	if err != nil {
		return nil
	}

	isMyuser := false
	// 保存合约
	upItem, err := getUpkeeping(bmeta.GetUid())
	if err != nil {
		go saveUpkeeping(bmeta.GetUid())
	} else {
		localID := localNode.Identity.Pretty()
		for _, proID := range upItem.ProviderIDs {
			if localID == proID {
				isMyuser = true
				offerItem, err := getOffer()
				if err != nil {
					return err
				}
				if upItem.Price < offerItem.Price {
					return errors.New("price is lower now")
				}
				break
			}
		}
	}

	if !isMyuser {
		return errors.New("NotMyUser")
	}

	go func() {
		bcid := cid.NewCidV2([]byte(splitedNcid[0]))
		if len(splitedNcid) < 5 {
			Nblk, err := blocks.NewBlockWithCid([]byte(value), bcid)
			if err != nil {
				log.Printf("Error create block %s: %s", bcid.String(), err)
				return
			}
			err = localNode.Blockstore.Put(Nblk)
			if err != nil {
				log.Printf("Error writing block to datastore: %s", err)
				return
			}
			return
		}

		typ := splitedNcid[2]

		switch typ {
		case "append":
			if has, err := localNode.Blockstore.Has(bcid); !has || err != nil {
				log.Printf("Error append field to block %s: %s", bcid.String(), err)
				return
			}
			beginOffset, err := strconv.Atoi(splitedNcid[3])
			if err != nil {
				log.Printf("Error append field to block %s: %s", bcid.String(), err)
				return
			}
			endOffset, err := strconv.Atoi(splitedNcid[4])
			if err != nil {
				log.Printf("Error append field to block %s: %s", bcid.String(), err)
				return
			}
			err = localNode.Blockstore.Append(bcid, []byte(value), beginOffset, endOffset)
			if err != nil {
				log.Printf("Error append field to block %s: %s", bcid.String(), err)
				return
			}
		case "update":
			_, err := strconv.Atoi(splitedNcid[3])
			if err != nil {
				log.Printf("Error append field to block %s: %s", bcid.String(), err)
				return
			}
			_, err = strconv.Atoi(splitedNcid[4])
			if err != nil {
				log.Printf("Error append field to block %s: %s", bcid.String(), err)
				return
			}
			if has, _ := localNode.Blockstore.Has(bcid); true == has {
				err := localNode.Blockstore.DeleteBlock(bcid)
				if err != nil {
					log.Printf("Error delete block %s: %s", bcid.String(), err)
				}
			}
			Nblk, err := blocks.NewBlockWithCid([]byte(value), bcid)
			if err != nil {
				log.Printf("Error create block %s: %s", bcid.String(), err)
				return
			}
			err = localNode.Blockstore.Put(Nblk)
			if err != nil {
				log.Printf("Error writing block %s to datastore: %s", Nblk.String(), err)
				return
			}
		default:
			log.Printf("Wrong type in put block")
		}
	}()

	return nil
}

func handleGetBlock(km *metainfo.KeyMeta, from string) (string, error) {
	// key is cid|ops|sig
	splitedNcid := strings.Split(km.ToString(), metainfo.DELIMITER)
	if len(splitedNcid) < 3 {
		return "", errors.New("Key is too short")
	}

	sigByte, err := b58.Decode(splitedNcid[2])
	if err != nil {
		return "", errors.New("Signature format is wrong")
	}

	res, userID, key, value, err := verify(sigByte)
	if err != nil {
		log.Printf("verify block %s failed, err is : %s", splitedNcid[0], err)
		return "", err
	}

	if res {
		// 验证通过
		// 内存channel的value变化
		// 然后持久化
		bcid := cid.NewCidV2([]byte(splitedNcid[0]))
		b, err := localNode.Blockstore.Get(bcid)
		if err != nil {
			return "", errors.New("Block is not found")
		}
		if key != "" {
			channelItem, err := getChannel(userID)
			if err != nil {
				return "", errors.New("Find channelItem in channelBook error")
			}

			log.Println("Downlaod success，change channel.value and persist: ", value.String())
			channelItem.Value = value
			ProContracts.channelBook.Store(userID, channelItem)
			err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(key, value.String(), "local")
			if err != nil {
				log.Println("cmdPutErr:", err)
			}
		}
		return string(b.RawData()), nil
	}

	log.Printf("verify is false %s", splitedNcid[0])

	return "", errors.New("Signature is wrong")
}

// verify verifies the transaction
func verify(mes []byte) (bool, string, string, *big.Int, error) {
	signForChannel := &pb.SignForChannel{}
	err := proto.Unmarshal(mes, signForChannel)
	if err != nil {
		log.Println("proto.Unmarshal when provider verify err:", err)
		return false, "", "", nil, err
	}

	//解析传过来的参数
	var money = new(big.Int)
	money = money.SetBytes(signForChannel.GetMoney())
	userAddr := common.HexToAddress(signForChannel.GetUserAddress())
	providerAddr := common.HexToAddress(signForChannel.GetProviderAddress())
	sig := signForChannel.GetSig() //传过来的签名信息如果是空，就表明是测试环境，直接返回true
	log.Println("====测试sig是否为空====:", sig == nil, " money:", money, " userAddr:", signForChannel.GetUserAddress(), " providerAddr:", signForChannel.GetProviderAddress())
	if sig == nil {
		return true, "", "", nil, nil
	}

	// 从内存获得value
	userID, err := address.GetIDFromAddress(userAddr.String())
	if err != nil {
		return false, "", "", nil, err
	}
	item, ok := ProContracts.channelBook.Load(userID)
	if !ok {
		log.Println("Not find ", userID, "'s channelItem in channelBook.")
		return false, "", "", nil, errors.New("Find channelItem in channelBook error")
	}
	channelItem, ok := item.(contracts.ChannelItem)
	if !ok {
		log.Println("Can't transfer item to channelItem.")
		return false, "", "", nil, errors.New("Transfer item to channelItem error")
	}
	//在Value的基础上再加上此次下载需要支付的money，就是此次验证签名的value
	addValue := int64((utils.BlockSize / (1024 * 1024)) * utils.READPRICEPERMB)
	// 默认100 + Value
	value := big.NewInt(0)
	value = value.Add(channelItem.Value, big.NewInt(addValue))

	if money.Cmp(value) < 0 { //比较对方传过来的value和此时的value值是否一样，不一样就返回false
		log.Println("value is different from money,  value is: ", value.String())
		//return false, "", "", nil, nil
	}

	//判断签名是否正确
	channelAddr, _, err := contracts.GetChannelAddr(providerAddr, providerAddr, userAddr)
	if err != nil {
		return false, "", "", nil, err
	}
	channelValueKeyMeta, err := metainfo.NewKeyMeta(channelAddr.String(), metainfo.Local, metainfo.SyncTypeChannelValue) // HexChannelAddress|13|channelValue
	if err != nil {
		return false, "", "", nil, err
	}
	res, err := contracts.VerifySig(signForChannel.GetUserPK(), sig, channelAddr, money)
	if err != nil {
		return false, "", "", nil, err
	}
	return res, userID, channelValueKeyMeta.ToString(), value, nil
}
