package provider

import (
	"errors"
	"strconv"
	"time"

	mcl "github.com/memoio/go-mefs/crypto/bls12"
	mpb "github.com/memoio/go-mefs/pb"
	"github.com/memoio/go-mefs/repo/fsrepo"
	"github.com/memoio/go-mefs/role"
	datastore "github.com/memoio/go-mefs/source/go-datastore"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
)

var (
	ErrGroupNotReady     = errors.New("group is nil")
	ErrUpkeepingNotReady = errors.New("upkeeping contract is not deployed")
)

type GInfoOutput struct {
	GroupID        string
	UserID         string
	Keepers        []string
	Providers      []string
	UpkeepingID    string
	UpkeepingStart string
	UpkeepingEnd   string
	UpkeepingPrice string
	ChannelValue   []string
}

func (p *Info) getNewUserConfig(userID, groupID string) (*mcl.KeySet, error) {
	value, ok := p.userConfigs.Get(groupID)
	if ok {
		return value.(*mcl.KeySet), nil
	}

	if userID == pos.GetPosId() {
		mkey, err := mcl.GenKeySetWithSeed(pos.GetPosSeed(), mcl.TagAtomNumV1, mcl.PDPCountV1)
		if err != nil {
			utils.MLogger.Info("Init bls config for pos user fail: ", err)
			return nil, err
		}

		p.userConfigs.Add(groupID, mkey)
	}

	kmBls12, err := metainfo.NewKey(groupID, mpb.KeyType_Config, userID)
	if err != nil {
		return nil, err
	}

	ctx := p.context
	userconfigkey := kmBls12.ToString()
	userconfigbyte, _ := p.ds.GetKey(ctx, userconfigkey, "local")
	if len(userconfigbyte) > 0 {
		mkey, err := role.BLS12ByteToKeyset(userconfigbyte, nil)
		if err == nil && mkey != nil {
			p.userConfigs.Add(groupID, mkey)
			return mkey, nil
		}
	}

	gp := p.getGroupInfo(userID, groupID, true)
	if gp == nil {
		return nil, role.ErrNotMyUser
	}

	for _, kid := range gp.keepers {
		userconfigbyte, _ := p.ds.GetKey(ctx, userconfigkey, kid)
		if len(userconfigbyte) > 0 {
			mkey, err := role.BLS12ByteToKeyset(userconfigbyte, nil)
			if err != nil {
				return nil, err
			}

			p.ds.PutKey(ctx, userconfigkey, userconfigbyte, nil, "local")

			p.userConfigs.Add(groupID, mkey)
			return mkey, nil
		}
	}

	return nil, role.ErrEmptyBlsKey
}

// getDiskUsage gets the disk usage
func (p *Info) getDiskUsage() (uint64, error) {
	used, err := datastore.DiskUsage(p.ds.DataStore())
	if err != nil {
		return 0, err
	}

	rootpath, err := fsrepo.BestKnownPath()
	if err != nil {
		return 0, err
	}

	localUsed, err := utils.GetDirSize(rootpath)
	if err != nil {
		return 0, err
	}

	if used != localUsed {
		utils.MLogger.Infof("localUsed is %d, while calculate is: %d", localUsed, used)
	}

	return localUsed, nil
}

// getDiskTotal gets the disk total space which is set in config
// default is 1TB
func (p *Info) getDiskTotal() uint64 {
	maxSpaceInByte := uint64(1024 * 1024 * 1024 * 1024)
	if p.proContract != nil && p.proContract.Capacity != 0 {
		maxSpaceInByte = uint64(p.proContract.Capacity) * 1024 * 1024
	}
	return maxSpaceInByte
}

//GetGroupInfoOutput get ginfo for show
func (p *Info) GetGroupInfoOutput(uid, qid string) (*GInfoOutput, error) {
	gpInfo := p.getGroupInfo(uid, qid, false)
	if gpInfo == nil {
		return nil, ErrGroupNotReady
	}

	if gpInfo.upkeeping == nil {
		return nil, ErrUpkeepingNotReady
	}

	var channelValue []string

	gpInfo.channel.Range(func(key, value interface{}) bool {
		cItem, ok := value.(*role.ChannelItem)
		if !ok {
			return true
		}
		channelValue = append(channelValue, utils.FormatWei(cItem.Value))
		return true
	})

	res := &GInfoOutput{
		GroupID:        qid,
		UserID:         uid,
		Keepers:        gpInfo.keepers,
		Providers:      gpInfo.providers,
		UpkeepingID:    gpInfo.upkeeping.UpKeepingID,
		UpkeepingStart: time.Unix(gpInfo.upkeeping.StartTime, 0).Format(utils.SHOWTIME),
		UpkeepingEnd:   time.Unix(gpInfo.upkeeping.EndTime, 0).Format(utils.SHOWTIME),
		UpkeepingPrice: utils.FormatStorePrice(gpInfo.upkeeping.Price),
		ChannelValue:   channelValue,
	}

	return res, nil
}

func (p *Info) ShowUserInfo() []string {
	var res []string
	res = append(res, "uid/qid:")
	num := 0

	p.users.Range(func(key, value interface{}) bool {
		num++
		uid, ok := key.(string)
		if !ok {
			return true
		}

		ui, ok := value.(*uInfo)
		if !ok {
			return true
		}

		qus := ui.getQuery()

		tmp := uid
		for _, qid := range qus {
			tmp = tmp + "/" + qid
		}

		res = append(res, tmp)
		return true
	})

	res = append(res, "sum: "+strconv.Itoa(num))
	return res
}
