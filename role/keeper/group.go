package keeper

import (
	"log"

	df "github.com/memoio/go-mefs/data-format"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func newGroup(localID, qid, uid string, keepers, providers []string) (*groupInfo, error) {
	tempInfo := &groupInfo{
		keepers:      keepers,
		providers:    providers,
		groupID:      qid,
		owner:        uid,
		localKeeper:  qid,
		masterKeeper: qid,
	}

	if qid != uid {
		err := saveUpkeepingToGP(qid, tempInfo)
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
	mykids, ok := contracts.GetKeepersOfPro(pid)
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
			return getMasterID(mymaster) == u.localID
		}
	}

	return g.localKeeper == g.masterKeeper
}

//getMasterID choose the middle nodes
func getMasterID(kidlist []string) string {
	sort.Strings(kidlist)
	return kidlist[len(kidlist)/2]
}

// bid is bucketID_stripeID_chunkID
func (g *groupInfo) addBlockMeta(pid, bid string, offset int) error {
	log.Println("add block: ", bid, "for query: ", qid, " and provider: ", pid)

	thisLinfo := &lInfo{}

	thisILinfo, ok := g.ledgerMap.Load(pid)
	if ok {
		thisLinfo = thisILinfo.(*lInfo)
	} else {
		g.ledgerMap.Store(pid, thisLinfo)
	}

	newcidinfo := &blockInfo{
		availtime: utils.GetUnixNow(),
		offset:    offset,
		repair:    0,
		storedOn:  pid,
	}

	// store in cidMap
	oldOffset := -1
	v, ok := thisLinfo.cidMap.Load(bid)
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
		thisLinfo.cidMap.Store(bid, newcidinfo)
	}

	thisLinfo.maxlength += (int64(offset-oldOffset) * df.DefaultSegmentSize)

	// store in buckets
	bucketID, stripeID, chunkID, err := metainfo.GetIDsFromBlock(bid)
	if err != nil {
		return err
	}

	bids := strings.SplitN(bid, metainfo.BLOCK_DELIMITER, 2)

	thisIBucket, ok := thisLinfo.buckets.Load(bucketID)
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
	thisILinfo, ok := g.ledgerMap.Load(pid)
	if !ok {
		return
	}

	thisLinfo = thisILinfo.(*lInfo)

	// delete from blockMap
	thisICid, ok := thisLinfo.blockMap.Load(bid)
	if ok {
		thisCid := thisICid.(*blockInfo)
		thisLinfo.maxlength -= (int64(thisCid.offset+1) * df.DefaultSegmentSize)
		thisLinfo.blockMap.Delete(bid)
	}

	// delete from buckets
	bids := strings.SplitN(bid, metainfo.BLOCK_DELIMITER, 2)

	bui, ok := thisLinfo.buckets.Load(bids[0])
	if ok {
		bui.(*bucketInfo).stripes.Delete(bis[1])
	}

	return
}

// bucketID_stripeID_chunkID
func (g *groupInfo) getBlockPos(bid string) (string, error) {
	bids := strings.SplitN(bid, metainfo.BLOCK_DELIMITER, 2)

	bui, ok := gri.(*groupInfo).buckets.Load(bis[0])
	if ok {
		sti, ok := bui.(*bucketInfo).stripes.Load(bis[1])
		if ok {
			return sti.(*cidInfo).storedOn, nil
		}
	}
}
