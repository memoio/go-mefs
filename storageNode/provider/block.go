package provider

import (
	"math/big"
	"strings"

	"github.com/gogo/protobuf/proto"
	df "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/source/data"
	bs "github.com/memoio/go-mefs/source/go-ipfs-blockstore"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func (p *Info) handlePutBlock(km *metainfo.Key, value []byte, from string) error {
	utils.MLogger.Info("handlePutBlock: ", km.ToString(), "from: ", from)
	// key is blockID/"block"
	splitedNcid := strings.Split(km.ToString(), metainfo.DELIMITER)
	if len(splitedNcid) != 2 {
		return role.ErrWrongValue
	}

	bids := strings.SplitN(splitedNcid[0], metainfo.BlockDelimiter, 2)
	qid := bids[0]

	gp := p.getGroupInfo(qid, qid, true)
	if gp == nil {
		return role.ErrNotMyUser
	}

	go func() {
		ctx := p.context
		err := p.ds.PutBlock(ctx, splitedNcid[0], value, "local")
		if err != nil {
			utils.MLogger.Info("Error writing block to datastore: %s", err)
			return
		}

		if role.Debug {
			blskey, _ := p.getNewUserConfig(gp.userID, qid)
			if blskey != nil && blskey.Pk != nil {
				ok := df.VerifyBlock(value, splitedNcid[0], blskey)
				if !ok {
					utils.MLogger.Errorf("Verify data for %s fails, delete it", splitedNcid[0])
					p.ds.DeleteBlock(ctx, splitedNcid[0], "local")
				}
			}
		}
	}()

	return nil
}

func (p *Info) handleAppendBlock(km *metainfo.Key, value []byte, from string) error {
	utils.MLogger.Info("handleAppendBlock: ", km.ToString(), " from: ", from)
	// key is blockID/"Block"/begin/end
	splitedNcid := strings.Split(km.ToString(), metainfo.DELIMITER)
	if len(splitedNcid) != 4 {
		return role.ErrWrongValue
	}

	bids := strings.SplitN(splitedNcid[0], metainfo.BlockDelimiter, 2)
	qid := bids[0]

	gp := p.getGroupInfo(qid, qid, true)
	if gp == nil {
		return role.ErrNotMyUser
	}

	ctx := p.context
	go func() {
		err := p.ds.AppendBlock(ctx, km.ToString(), value, "local")
		if err != nil {
			utils.MLogger.Errorf("append to block %s: %s", km.ToString(), err)
			return
		}
	}()
	return nil
}

func (p *Info) handleGetBlock(km *metainfo.Key, metaValue, sig []byte, from string) ([]byte, error) {
	utils.MLogger.Info("handleGetBlock: ", km.ToString(), " from: ", from)
	kms := km.ToString()
	splitedNcid := strings.Split(kms, metainfo.DELIMITER)
	if len(splitedNcid) < 2 {
		return nil, role.ErrWrongValue
	}

	bids := strings.SplitN(splitedNcid[0], metainfo.BlockDelimiter, 2)
	qid := bids[0]

	gp := p.getGroupInfo(qid, qid, true)
	if gp == nil {
		return nil, role.ErrNotMyUser
	}

	ctx := p.context

	if gp.userID != gp.groupID {
		res, chanGot, value, err := verifyChanSign(sig)
		if err != nil {
			utils.MLogger.Errorf("verify sig for block %s failed, err is : %s", splitedNcid[0], err)
			return nil, err
		}

		var cItem *role.ChannelItem
		if chanGot != "" {
			gotItem, ok := gp.channel.Load(chanGot)
			if !ok {
				go gp.getChanItem(p.localID, chanGot)
				return nil, data.ErrRetry
			}

			cItem = gotItem.(*role.ChannelItem)
		}

		if value != nil {
			if cItem.Money.Cmp(big.NewInt(0)) == 0 || value.Cmp(cItem.Money) > 0 {
				utils.MLogger.Errorf("verify sig for block %s failed, money is not enough, has %s, expected %s", splitedNcid[0], cItem.Money.String(), value.String())
				go p.loadChannelValue(gp.userID, gp.groupID)
				return nil, role.ErrNotEnoughMoney
			}
		}

		if res {
			utils.MLogger.Infof("try to get block %s form local", splitedNcid[0])
			b, err := p.ds.GetBlock(ctx, kms, nil, "local")
			if err != nil {
				utils.MLogger.Errorf("get block %s from local fail: %s", splitedNcid[0], err)
				return nil, err
			}

			readLen := len(b.RawData())
			if value != nil {
				ok := verifyChanValue(cItem.Value, value, readLen)
				if !ok {
					return nil, role.ErrWrongMoney
				}

				cItem.Value = value
				cItem.Sig = sig

				key, err := metainfo.NewKey(p.localID, mpb.KeyType_Channel, cItem.ChannelID)
				if err != nil {
					return nil, err
				}

				p.ds.PutKey(ctx, key.ToString(), sig, nil, "local")
			}

			return b.RawData(), nil
		}
		utils.MLogger.Warnf("sign verify is false for %s", splitedNcid[0])
		return nil, role.ErrWrongSign
	}

	utils.MLogger.Infof("try to get block %s form local", splitedNcid[0])
	b, err := p.ds.GetBlock(ctx, kms, nil, "local")
	if err != nil {
		utils.MLogger.Errorf("get block %s from local fail: %s", splitedNcid[0], err)
		return nil, err
	}

	return b.RawData(), nil
}

// verify verifies the transaction
func verifyChanSign(mes []byte) (bool, string, *big.Int, error) {
	cSign := &mpb.ChannelSign{}
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

func verifyChanValue(oldValue, newValue *big.Int, readLen int) bool {
	//verify value;： value ?= oldValue + 100
	addValue := big.NewInt(int64(readLen) * utils.READPRICE / (1024 * 1024 * 1024))
	addValue.Add(addValue, oldValue)
	if newValue.Cmp(addValue) < 0 {
		utils.MLogger.Warn(newValue.String(), " received is less than calculated: ", addValue.String())
		return false
	}
	return true
}

func (p *Info) handleDeleteBlock(km *metainfo.Key, from string) error {
	utils.MLogger.Info("handleDeleteBlock: ", km.ToString(), "from: ", from)
	err := p.ds.DeleteBlock(p.context, km.ToString(), "local")
	if err != nil && err != bs.ErrNotFound {
		return err
	}
	return nil
}
