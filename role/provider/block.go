package provider

import (
	"context"
	"errors"
	"math/big"
	"strings"

	"github.com/golang/protobuf/proto"
	df "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
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
		if Debug {
			blskey, _ := p.getNewUserConfig(gp.userID, qid)
			if blskey != nil && blskey.Pk != nil {
				ok := df.VerifyBlock(value, splitedNcid[0], blskey)
				if !ok {
					utils.MLogger.Warnf("Verify data for %s fails", splitedNcid[0])
				}
			}
		}
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
	utils.MLogger.Info("handleAppendBlock: ", km.ToString(), " from: ", from)
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
	utils.MLogger.Info("handleGetBlock: ", km.ToString(), " from: ", from)

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

	ctx := context.Background()

	if gp.userID != gp.groupID {
		res, chanGot, value, err := verifyChanSign(sig)
		if err != nil {
			utils.MLogger.Errorf("verify sig for block %s failed, err is : %s", splitedNcid[0], err)
			return nil, err
		}

		var cItem *role.ChannelItem
		if chanGot != "" {
			cItem = gp.getChanItem(p.localID, chanGot)
			if cItem == nil {
				utils.MLogger.Warn("channel is empty, reget it for: ", chanGot)
				return nil, errors.New("Channel is not valid")
			}
		}

		if value != nil {
			if value.Cmp(cItem.Money) > 0 {
				utils.MLogger.Errorf("verify block %s failed, money is not enough", splitedNcid[0], err)
				return nil, errors.New("money is not enough")
			}

			ok := verifyChanValue(cItem.Value, value)
			if !ok {
				return nil, errors.New("money is not right")
			}
		}

		if res {
			utils.MLogger.Infof("try to get block %s form local", splitedNcid[0])
			b, err := p.ds.GetBlock(ctx, splitedNcid[0], nil, "local")
			if err != nil {
				utils.MLogger.Errorf("get block %s from local fail: %s", splitedNcid[0], err)
				return nil, err
			}

			if value != nil {
				cItem.Value = value
				cItem.Sig = sig

				key, err := metainfo.NewKeyMeta(cItem.ChannelID, metainfo.Channel)
				if err != nil {
					return nil, err
				}

				p.ds.PutKey(ctx, key.ToString(), sig, nil, "local")
			}

			return b.RawData(), nil
		}
		utils.MLogger.Warnf("sign verify is false for %s", splitedNcid[0])
		return nil, errors.New("Signature is wrong")
	}

	utils.MLogger.Infof("try to get block %s form local", splitedNcid[0])
	b, err := p.ds.GetBlock(ctx, splitedNcid[0], nil, "local")
	if err != nil {
		utils.MLogger.Errorf("get block %s from local fail: %s", splitedNcid[0], err)
		return nil, err
	}

	return b.RawData(), nil
}

// verify verifies the transaction
func verifyChanSign(mes []byte) (bool, string, *big.Int, error) {
	cSign := &pb.ChannelSign{}
	err := proto.Unmarshal(mes, cSign)
	if err != nil {
		utils.MLogger.Error("proto.Unmarshal when provider verify err:", err)
		return false, "", nil, err
	}

	if cSign.GetChannelID() == "test" && string(cSign.GetValue()) == "123" {
		utils.MLogger.Debug("sign for test and repair")
		return true, "", nil, nil
	}

	// verify sign
	res := role.VerifyChannelSign(cSign)
	if !res {
		utils.MLogger.Error("signature is wrong")
		return false, "", nil, nil
	}

	chanGot := cSign.GetChannelID()
	valueGot := new(big.Int).SetBytes(cSign.GetValue())

	return res, chanGot, valueGot, nil
}

func verifyChanValue(oldValue, newValue *big.Int) bool {
	//verify value;： value ?= oldValue + 100
	addValue := int64((df.BlockSize / (1024 * 1024)) * utils.READPRICEPERMB)
	oldValue = oldValue.Add(oldValue, big.NewInt(addValue))
	if newValue.Cmp(oldValue) < 0 {
		utils.MLogger.Warn(newValue.String(), " received is less than calculated: ", oldValue.String())
		return false
	}
	return true
}

func (p *Info) handleDeleteBlock(km *metainfo.KeyMeta, from string) error {
	utils.MLogger.Info("handleDeleteBlock: ", km.ToString(), "from: ", from)
	err := p.ds.DeleteBlock(context.Background(), km.ToString(), "local")
	if err != nil && err != bs.ErrNotFound {
		return err
	}
	return nil
}
