package provider

import (
	"context"
	"errors"
	"math/big"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/memoio/go-mefs/role"
	pb "github.com/memoio/go-mefs/role/user/pb"
	bs "github.com/memoio/go-mefs/source/go-ipfs-blockstore"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func (p *Info) handlePutBlock(km *metainfo.KeyMeta, value []byte, from string) error {
	utils.MLogger.Info("handlePutBlock: ", km.ToString(), "from: ", from)
	// key is blockID/"block"
	splitedNcid := strings.Split(km.ToString(), metainfo.DELIMITER)
	if len(splitedNcid) != 2 {
		return errors.New("Wrong value for put block")
	}

	bids := strings.SplitN(splitedNcid[0], metainfo.BLOCK_DELIMITER, 2)
	qid := bids[0]

	gp := p.getGroupInfo(qid, qid, true)
	if gp == nil {
		return errors.New("NotMyUser")
	}

	ctx := context.Background()

	go func() {
		err := p.ds.PutBlock(ctx, splitedNcid[0], value, "local")
		if err != nil {
			utils.MLogger.Info("Error writing block to datastore: %s", err)
			return
		}
		return
	}()

	return nil
}

func (p *Info) handleAppendBlock(km *metainfo.KeyMeta, value []byte, from string) error {
	utils.MLogger.Info("handleAppendBlock: ", km.ToString(), "from: ", from)
	// key is blockID/"Block"/begin/end
	splitedNcid := strings.Split(km.ToString(), metainfo.DELIMITER)
	if len(splitedNcid) != 4 {
		return errors.New("Wrong value for put block")
	}

	bids := strings.SplitN(splitedNcid[0], metainfo.BLOCK_DELIMITER, 2)
	qid := bids[0]

	gp := p.getGroupInfo(qid, qid, true)
	if gp == nil {
		return errors.New("NotMyUser")
	}

	ctx := context.Background()
	go func() {
		err := p.ds.AppendBlock(ctx, km.ToString(), value, "local")
		if err != nil {
			utils.MLogger.Errorf("append to block %s: %s", km.ToString(), err)
			return
		}
	}()
	return nil
}

func (p *Info) handleGetBlock(km *metainfo.KeyMeta, metaValue, sig []byte, from string) ([]byte, error) {
	utils.MLogger.Info("handleGetBlock: ", km.ToString(), "from: ", from)

	splitedNcid := strings.Split(km.ToString(), metainfo.DELIMITER)
	if len(splitedNcid) != 2 {
		return nil, errors.New("Wrong value for get block")
	}

	bids := strings.SplitN(splitedNcid[0], metainfo.BLOCK_DELIMITER, 2)
	qid := bids[0]

	gp := p.getGroupInfo(qid, qid, true)
	if gp == nil {
		return nil, errors.New("NotMyUser")
	}

	if gp.userID != gp.groupID && gp.channel != nil {
		chanID := gp.channel.ChannelID
		value := gp.channel.Value

		res, value, sig, err := p.verify(chanID, value, sig)
		if err != nil {
			utils.MLogger.Info("verify block %s failed, err is : %s", splitedNcid[0], err)
			return nil, err
		}

		if res {
			b, err := p.ds.GetBlock(context.Background(), splitedNcid[0], nil, "local")
			if err != nil {
				return nil, errors.New("Block is not found")
			}

			key, err := metainfo.NewKeyMeta(gp.channel.ChannelID, metainfo.Channel) // HexChannelAddress|13|channelValue
			if err != nil {
				return nil, err
			}

			err = p.ds.PutKey(context.Background(), key.ToString(), value.Bytes(), "local")
			if err != nil {
				utils.MLogger.Info("cmdPutErr:", err)
			}

			gp.channel.Value = value
			gp.channel.Sig = sig

			return b.RawData(), nil
		}

		utils.MLogger.Info("verify is false %s", splitedNcid[0])
	} else {
		b, err := p.ds.GetBlock(context.Background(), splitedNcid[0], nil, "local")
		if err != nil {
			return nil, errors.New("Block is not found")
		}

		return b.RawData(), nil
	}
	return nil, errors.New("Signature is wrong")
}

// verify verifies the transaction
func (p *Info) verify(chanID string, oldValue *big.Int, mes []byte) (bool, *big.Int, []byte, error) {
	signForChannel := &pb.SignForChannel{}
	err := proto.Unmarshal(mes, signForChannel)
	if err != nil {
		utils.MLogger.Error("proto.Unmarshal when provider verify err:", err)
		return false, nil, nil, err
	}

	//解析传过来的参数
	var money = new(big.Int)
	money = money.SetBytes(signForChannel.GetMoney())
	sig := signForChannel.GetSig()
	utils.MLogger.Info("====测试sig是否为空====:", sig == nil, " money:", money, " userAddr:", signForChannel.GetUserAddress(), " providerAddr:", signForChannel.GetProviderAddress())
	if sig == nil {
		return true, nil, nil, nil
	}

	//verify value： value ?= oldValue + 100
	addValue := int64((utils.BlockSize / (1024 * 1024)) * utils.READPRICEPERMB)
	value := big.NewInt(0)
	value = value.Add(oldValue, big.NewInt(addValue))
	if money.Cmp(value) < 0 {
		utils.MLogger.Info("value is less than money,  value is: ", value.String())
		// to test
		//return false, nil, nil, nil
	}

	//判断签名是否正确
	res, err := role.VerifySig(chanID, money, sig, signForChannel.GetUserPK())
	if err != nil {
		return false, nil, nil, err
	}
	return res, value, sig, nil
}

func (p *Info) handleDeleteBlock(km *metainfo.KeyMeta, from string) error {
	utils.MLogger.Info("handleDeleteBlock: ", km.ToString(), "from: ", from)
	err := p.ds.DeleteBlock(context.Background(), km.ToString(), "local")
	if err != nil && err != bs.ErrNotFound {
		return err
	}
	return nil
}
