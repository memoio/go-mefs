package provider

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/proto"
	u "github.com/ipfs/go-ipfs-util"
	recpb "github.com/libp2p/go-libp2p-record/pb"

	"github.com/memoio/go-mefs/contracts"
	pb "github.com/memoio/go-mefs/role/user/pb"
	blocks "github.com/memoio/go-mefs/source/go-block-format"
	cid "github.com/memoio/go-mefs/source/go-cid"
	ds "github.com/memoio/go-mefs/source/go-datastore"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func handlePutBlock(km *metainfo.KeyMeta, value, from string) error {
	// key is cid|ops|type|begin|end
	splitedNcid := strings.Split(km.ToString(), metainfo.DELIMITER)
	bcid := cid.NewCidV2([]byte(splitedNcid[0]))
	if len(splitedNcid) < 5 {
		Nblk, err := blocks.NewBlockWithCid([]byte(value), bcid)
		if err != nil {
			fmt.Printf("Error create block %s: %s", bcid.String(), err)
			return err
		}
		err = localNode.Blockstore.Put(Nblk)
		if err != nil {
			fmt.Printf("Error writing block to datastore: %s", err)
			return err
		}
		return nil
	}

	typ := splitedNcid[2]

	switch typ {
	case "append":
		if has, err := localNode.Blockstore.Has(bcid); !has || err != nil {
			fmt.Printf("Error append field to block %s: %s", bcid.String(), err)
			return err
		}
		beginOffset, err := strconv.Atoi(splitedNcid[3])
		if err != nil {
			fmt.Printf("Error append field to block %s: %s", bcid.String(), err)
			return err
		}
		endOffset, err := strconv.Atoi(splitedNcid[4])
		if err != nil {
			fmt.Printf("Error append field to block %s: %s", bcid.String(), err)
			return err
		}
		err = localNode.Blockstore.Append(bcid, []byte(value), beginOffset, endOffset)
		if err != nil {
			fmt.Printf("Error append field to block %s: %s", bcid.String(), err)
			return err
		}
		addBlockRecord(bcid.String(), endOffset)
	case "update":
		if has, _ := localNode.Blockstore.Has(bcid); true == has {
			err := localNode.Blockstore.DeleteBlock(bcid)
			if err != nil {
				fmt.Printf("Error delete block %s: %s", bcid.String(), err)
			}
		}
		_, err := strconv.Atoi(splitedNcid[3])
		if err != nil {
			fmt.Printf("Error append field to block %s: %s", bcid.String(), err)
			return err
		}
		endOffset, err := strconv.Atoi(splitedNcid[4])
		if err != nil {
			fmt.Printf("Error append field to block %s: %s", bcid.String(), err)
			return err
		}
		Nblk, err := blocks.NewBlockWithCid([]byte(value), bcid)
		if err != nil {
			fmt.Printf("Error create block %s: %s", bcid.String(), err)
			return err
		}
		err = localNode.Blockstore.Put(Nblk)
		if err != nil {
			fmt.Printf("Error writing block %s to datastore: %s", Nblk.String(), err)
			return err
		}
		addBlockRecord(bcid.String(), endOffset)
	default:
		fmt.Printf("Wrong type in put block")
	}
	return nil
}

func handleGetBlock(km *metainfo.KeyMeta, from string) (string, error) {
	// key is cid|ops|sig
	splitedNcid := strings.Split(km.ToString(), metainfo.DELIMITER)
	if len(splitedNcid) < 3 {
		return "", errors.New("Key is too short")
	}

	res, userID, key, value, err := verify([]byte(splitedNcid[2]))
	if err != nil {
		fmt.Printf("verify block %s failed, err is : %s", splitedNcid[0], err)
	} else if res { //验证通过
		// 内存channel的value变化
		// 然后持久化
		bcid := cid.NewCidV2([]byte(splitedNcid[0]))
		b, err := localNode.Blockstore.Get(bcid)
		if err != nil {
			return "", errors.New("Block is not found")
		}
		if key != "" {
			item, ok := contracts.ProContracts.ChannelBook.Load(userID)
			if !ok {
				return "", errors.New("Find channelItem in channelBook error")
			}
			channelItem, ok := item.(contracts.ChannelItem)
			if !ok {
				return "", errors.New("Transfer item to channelItem error")
			}
			fmt.Println("下载成功，更改内存中channel.value并持久化:", value.String())
			channelItem.Value = value
			contracts.ProContracts.ChannelBook.Store(userID, channelItem)
			err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(key, value.String(), "local")
			if err != nil {
				fmt.Println("cmdPutErr:", err)
				return string(b.RawData()), err
			}
		}
		return string(b.RawData()), nil
	} else { //验证不通过
		fmt.Printf("verify is false %s", splitedNcid[0])
	}

	return "", nil
}

func addBlockRecord(BlockID string, offset int) error {
	key, err := metainfo.NewKeyMeta(BlockID, metainfo.HasBlock)
	if err != nil {
		return err
	}
	rec := new(recpb.Record)
	rec.Key = key.ToByte()
	rec.Value = []byte(strconv.Itoa(offset))
	rec.TimeReceived = u.FormatRFC3339(time.Now())
	data, err := proto.Marshal(rec)
	if err != nil {
		fmt.Printf("addBlockRecord: %s", err)
		return err
	}

	dataStore := localNode.Repo.Datastore()
	return dataStore.Put(ds.NewKey(key.ToString()), data)
}

// verify verifies the transaction
func verify(mes []byte) (bool, string, string, *big.Int, error) {
	signForChannel := &pb.SignForChannel{}
	err := proto.Unmarshal(mes, signForChannel)
	if err != nil {
		fmt.Println("proto.Unmarshal when provider verify err:", err)
		return false, "", "", nil, err
	}

	//解析传过来的参数
	var money = new(big.Int)
	money = money.SetBytes(signForChannel.GetMoney())
	userAddr := common.HexToAddress(signForChannel.GetUserAddress())
	providerAddr := common.HexToAddress(signForChannel.GetProviderAddress())
	sig := signForChannel.GetSig() //传过来的签名信息如果是空，就表明是测试环境，直接返回true
	fmt.Println("====测试sig是否为空====:", sig == nil, " money:", money, " userAddr:", signForChannel.GetUserAddress(), " providerAddr:", signForChannel.GetProviderAddress())
	if sig == nil {
		return true, "", "", nil, nil
	}

	//判断value值是否正确
	if money.Cmp(big.NewInt(0)) == 0 { //传过来的value值如果是0，就表示是测试环境，直接返回true
		return true, "", "", nil, nil
	}

	// 从内存获得value
	userID, err := address.GetIDFromAddress(userAddr.String())
	if err != nil {
		return false, "", "", nil, err
	}
	item, ok := contracts.ProContracts.ChannelBook.Load(userID)
	if !ok {
		fmt.Println("Not find ", userID, "'s channelItem in channelBook.")
		return false, "", "", nil, errors.New("Find channelItem in channelBook error")
	}
	channelItem, ok := item.(contracts.ChannelItem)
	if !ok {
		fmt.Println("Can't transfer item to channelItem.")
		return false, "", "", nil, errors.New("Transfer item to channelItem error")
	}
	//在Value的基础上再加上此次下载需要支付的money，就是此次验证签名的value
	addValue := int64((utils.BlockSize / (1024 * 1024)) * utils.READPRICEPERMB)
	// 默认100 + Value
	value := big.NewInt(0)
	value = value.Add(channelItem.Value, big.NewInt(addValue))
	fmt.Println("==========测试最终value=======:", value.String())

	if money.Cmp(value) != 0 { //比较对方传过来的value和此时的value值是否一样，不一样就返回false
		return false, "", "", nil, nil
	}

	//判断签名是否正确
	channelAddr, err := contracts.GetChannelAddr(providerAddr, providerAddr, userAddr)
	if err != nil {
		return false, "", "", nil, err
	}
	channelValueKeyMeta, err := metainfo.NewKeyMeta(channelAddr.String(), metainfo.Local, metainfo.SyncTypeChannelValue) // HexChannelAddress|13|channelValue
	if err != nil {
		return false, "", "", nil, err
	}
	res, err := contracts.VerifySig(signForChannel.GetUserPK(), sig, channelAddr, value)
	if err != nil {
		return false, "", "", nil, err
	}
	return res, userID, channelValueKeyMeta.ToString(), value, nil
}
