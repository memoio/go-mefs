package keeper

import (
	"errors"
	"log"
	"sort"
	"strconv"
	"strings"

	df "github.com/memoio/go-mefs/data-format"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func newGroup(localID, uid, qid string, keepers, providers []string) (*groupInfo, error) {
	tempInfo := &groupInfo{
		groupID:      qid,
		userID:       uid,
		localKeeper:  qid,
		masterKeeper: qid,
		keepers:      keepers,
		providers:    providers,
	}

	if qid != uid {
		err := tempInfo.getContracts(true)
		if err != nil {
			return nil, err
		}
	} else {
		for _, keeperID := range tempInfo.keepers {
			if localID == keeperID {
				tempInfo.localKeeper = localID
			}
		}

		// not my user
		if tempInfo.localKeeper != localID {
			log.Println(uid, "is not my user")
			return nil, errors.New("Not my user")
		}
	}

	tempInfo.masterKeeper = getMasterID(tempInfo.keepers)

	return tempInfo, nil
}

// if this provider belongs to this keeper, then this keeper is master
func (g *groupInfo) isMaster(pid string) bool {
	var mymaster []string
	mykids, ok := role.GetKeepersOfPro(pid)
	if ok {
		for _, keeperID := range g.keepers {
			for _, nkid := range mykids {
				if nkid == keeperID {
					mymaster = append(mymaster, keeperID)
					break
				}
			}
		}
		if len(mymaster) > 0 {
			return getMasterID(mymaster) == g.localKeeper
		}
	}

	return g.localKeeper == g.masterKeeper
}

//getMasterID choose the middle nodes
func getMasterID(kidlist []string) string {
	sort.Strings(kidlist)
	return kidlist[len(kidlist)/2]
}

func (g *groupInfo) getLInfo(proID string, mode bool) *lInfo {
	if g == nil {
		return nil
	}

	thisIl, ok := g.ledgerMap.Load(proID)
	if !ok {
		if mode {
			templInfo := &lInfo{}
			g.ledgerMap.Store(proID, templInfo)
			return templInfo
		}
		return nil
	}

	return thisIl.(*lInfo)
}

func (g *groupInfo) getBucketInfo(bucketID string, mode bool) *bucketInfo {
	if g == nil {
		return nil
	}

	thisIb, ok := g.buckets.Load(bucketID)
	if !ok {
		if mode {
			tempInfo := &bucketInfo{}
			g.buckets.Store(bucketID, tempInfo)
			return tempInfo
		}
		return nil
	}

	return thisIb.(*bucketInfo)
}

// bid is bucketID_stripeID_chunkID
func (g *groupInfo) addBlockMeta(bid, pid string, offset int) error {

	thisLinfo := g.getLInfo(pid, true)
	if thisLinfo == nil {
		return errors.New("group addBlockMeta err")
	}

	newcidinfo := &blockInfo{
		availtime: utils.GetUnixNow(),
		offset:    offset,
		repair:    0,
		storedOn:  pid,
	}

	// store in cidMap
	oldOffset := -1
	v, ok := thisLinfo.blockMap.Load(bid)
	if ok {
		newcidinfo = v.(*blockInfo)
		oldOffset = newcidinfo.offset
		if offset > oldOffset {
			newcidinfo.offset = oldOffset
		}

		if newcidinfo.storedOn != pid {
			newcidinfo.storedOn = pid
		}
	} else {
		thisLinfo.blockMap.Store(bid, newcidinfo)
	}

	thisLinfo.maxlength += (int64(offset-oldOffset) * df.DefaultSegmentSize)

	// store in buckets
	bucketID, stripeID, chunkID, err := metainfo.GetIDsFromBlock(bid)
	if err != nil {
		return err
	}

	bids := strings.SplitN(bid, metainfo.BLOCK_DELIMITER, 2)

	thisIBucket, ok := g.buckets.Load(bucketID)
	if !ok {
		return errors.New("cannot create bucket info")
	}

	thisBucket := thisIBucket.(*bucketInfo)

	// key: stripeID_chunkID
	thisBucket.stripes.Store(bids[1], newcidinfo)

	snum, err := strconv.Atoi(stripeID)
	if err != nil {
		return err
	}

	if snum > thisBucket.curStripes {
		thisBucket.curStripes = snum
	}

	cnum, err := strconv.Atoi(chunkID)
	if err != nil {
		return err
	}

	if cnum > thisBucket.chunkNum {
		thisBucket.chunkNum = cnum
	}

	return nil
}

func (g *groupInfo) deleteBlockMeta(pid, bid string) {
	thisLinfo := g.getLInfo(pid, false)
	if thisLinfo == nil {
		return
	}

	// delete from blockMap
	thisICid, ok := thisLinfo.blockMap.Load(bid)
	if ok {
		thisCid := thisICid.(*blockInfo)
		thisLinfo.maxlength -= (int64(thisCid.offset+1) * df.DefaultSegmentSize)
		thisLinfo.blockMap.Delete(bid)
	}

	// delete from buckets
	bids := strings.SplitN(bid, metainfo.BLOCK_DELIMITER, 2)

	bui, ok := g.buckets.Load(bids[0])
	if ok {
		bui.(*bucketInfo).stripes.Delete(bids[1])
	}

	return
}

// bucketID_stripeID_chunkID
func (g *groupInfo) getBlockPos(bid string) (string, error) {
	bids := strings.SplitN(bid, metainfo.BLOCK_DELIMITER, 2)

	bui, ok := g.buckets.Load(bids[0])
	if ok {
		sti, ok := bui.(*bucketInfo).stripes.Load(bids[1])
		if ok {
			return sti.(*blockInfo).storedOn, nil
		}
	}

	return "", errors.New("No such block")
}
