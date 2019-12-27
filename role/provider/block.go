package provider

import (
	"context"
	"errors"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/proto"
	"github.com/memoio/go-mefs/contracts"
	pb "github.com/memoio/go-mefs/role/user/pb"
	bs "github.com/memoio/go-mefs/source/go-ipfs-blockstore"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	b58 "github.com/mr-tron/base58/base58"
)

func handlePutBlock(km *metainfo.KeyMeta, value []byte, from string) error {
	// key is "block"/cid
	splitedNcid := strings.Split(km.ToString(), metainfo.DELIMITER)
	if len(splitedNcid) != 2 {
		return errors.New("Wrong value for put block")
	}

	bmeta, err := metainfo.GetBlockMeta(splitedNcid[1])
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

	ctx := context.Background()

	go func() {
		err = localNode.Data.PutBlock(ctx, km.ToString(), value, "local")
		if err != nil {
			log.Printf("Error writing block to datastore: %s", err)
			return
		}
		return
	}()

	return nil
}

func handleAppendBlock(km *metainfo.KeyMeta, value []byte, from string) error {
	// key is "block"/cid/begin/end
	splitedNcid := strings.Split(km.ToString(), metainfo.DELIMITER)
	if len(splitedNcid) != 4 {
		return errors.New("Wrong value for put block")
	}

	bmeta, err := metainfo.GetBlockMeta(splitedNcid[1])
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

	ctx := context.Background()
	go func() {
		err = localNode.Data.AppendBlock(ctx, km.ToString(), value, "local")
		if err != nil {
			log.Printf("Error append field to block %s: %s", km.ToString(), err)
			return
		}
	}()
	return nil
}

func handleGetBlock(km *metainfo.KeyMeta, from string) ([]byte, error) {
	// key is cid|ops|sig
	splitedNcid := strings.Split(km.ToString(), metainfo.DELIMITER)
	if len(splitedNcid) < 3 {
		return nil, errors.New("Key is too short")
	}

	sigByte, err := b58.Decode(splitedNcid[2])
	if err != nil {
		return nil, errors.New("Signature format is wrong")
	}

	res, userID, key, value, err := verify(sigByte)
	if err != nil {
		log.Printf("verify block %s failed, err is : %s", splitedNcid[0], err)
		return nil, err
	}

	if res {
		// 验证通过
		// 内存channel的value变化
		// 然后持久化
		b, err := localNode.Data.GetBlock(context.Background(), splitedNcid[0], nil, "local")
		if err != nil {
			return nil, errors.New("Block is not found")
		}
		if key != "" {
			channelItem, err := getChannel(userID)
			if err != nil {
				return nil, errors.New("Find channelItem in channelBook error")
			}

			log.Println("Downlaod success，change channel.value and persist: ", value.String())
			channelItem.Value = value
			proContracts.channelBook.Store(userID, channelItem)
			err = localNode.Data.PutKey(context.Background(), key, []byte(value.String()), "local")
			if err != nil {
				log.Println("cmdPutErr:", err)
			}
		}
		return b.RawData(), nil
	}

	log.Printf("verify is false %s", splitedNcid[0])

	return nil, errors.New("Signature is wrong")
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
	item, ok := proContracts.channelBook.Load(userID)
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
	channelValueKeyMeta, err := metainfo.NewKeyMeta(channelAddr.String(), metainfo.Channel) // HexChannelAddress|13|channelValue
	if err != nil {
		return false, "", "", nil, err
	}
	res, err := contracts.VerifySig(signForChannel.GetUserPK(), sig, channelAddr, money)
	if err != nil {
		return false, "", "", nil, err
	}
	return res, userID, channelValueKeyMeta.ToString(), value, nil
}

func handleDeleteBlock(km *metainfo.KeyMeta, from string) error {
	blockID := km.GetMid()
	err := localNode.Data.DeleteBlock(context.Background(), blockID, "local")
	if err != nil && err != bs.ErrNotFound {
		return err
	}
	return nil
}
