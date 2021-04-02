package keeper

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/memoio/go-mefs/contracts"
	id "github.com/memoio/go-mefs/crypto/identity"
	mpb "github.com/memoio/go-mefs/pb"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	ma "github.com/multiformats/go-multiaddr"
	mdns "github.com/multiformats/go-multiaddr-dns"
	mnet "github.com/multiformats/go-multiaddr-net"
)

// HandleMetaMessage callback
func (k *Info) HandleMetaMessage(opType mpb.OpType, metaKey string, metaValue, sig []byte, from string) ([]byte, error) {
	if !k.Online() {
		return nil, role.ErrServiceNotReady
	}

	km, err := metainfo.NewKeyFromString(metaKey)
	if err != nil {
		return nil, err
	}
	dtype := km.GetKeyType()
	switch dtype {
	case mpb.KeyType_UserInit:
		go k.handleUserInit(km, from)
	case mpb.KeyType_UserNotify:
		return k.handleUserNotify(km, metaValue, from)
	case mpb.KeyType_UserStart:
		return k.handleUserStart(km, metaValue, sig, from)
	case mpb.KeyType_UserStop:
		go k.handleUserStop(km, metaValue, sig, from)
	case mpb.KeyType_HeartBeat:
		return k.handleHeartBeat(km, metaValue, from)
	case mpb.KeyType_Bucket:
		go k.handleAddBucket(km, metaValue, sig, from)
	case mpb.KeyType_BlockPos:
		switch opType {
		case mpb.OpType_Put:
			go k.handleAddBlockPos(km, metaValue, sig, from)
		case mpb.OpType_Get:
			return k.handleGetKey(km, metaValue, sig, from)
		case mpb.OpType_Delete:
			go k.handleDeleteBlockPos(km, metaValue, sig, from)
		}
	case mpb.KeyType_Challenge:
		if opType == mpb.OpType_Put {
			go k.handleProof(km, metaValue)
		}
	case mpb.KeyType_Repair: //provider 修复回复
		switch opType {
		case mpb.OpType_Put:
			go k.handleRepairResult(km, metaValue, from)
		case mpb.OpType_BroadCast:
			go k.handleRepairUpdate(km, metaValue, from)
		}
	case mpb.KeyType_Storage:
		go k.handleStorage(km, metaValue, from)
	case mpb.KeyType_ExternalAddress:
		switch opType {
		case mpb.OpType_Put:
			go k.handlePutExAddr(km, metaValue, from)
		case mpb.OpType_Get:
			return k.handleExternalAddr(km)
		case mpb.OpType_Delete:
			go k.handleDelExAddr(km)
		}
	case mpb.KeyType_ChalTime:
		return k.handleChalTime(km)
	case mpb.KeyType_Pos:
		switch opType {
		case mpb.OpType_Put:
			go k.handlePosAdd(km, metaValue, from)
		case mpb.OpType_Delete:
			go k.handlePosDelete(km, metaValue, from)
		case mpb.OpType_Get:
			go k.handlePosGet(km, metaValue, from)
		}
	case mpb.KeyType_StPaySign:
		switch opType {
		case mpb.OpType_Put:
			go k.handlePutStPaySign(km, metaValue, sig, from)
		case mpb.OpType_Get:
			go k.handleGetStPaySign(km, metaValue, sig, from)
		}
	case mpb.KeyType_StPayShare:
		switch opType {
		case mpb.OpType_Put:
			go k.handlePutStPayShare(km, metaValue, from)
		case mpb.OpType_Get:
			go k.handleGetStPayShare(km, metaValue, from)
		}
	case mpb.KeyType_ProAddSign:
		switch opType {
		case mpb.OpType_Get:
			return k.handleGetProAddSign(km, metaValue, sig, from)
		}
	case mpb.KeyType_ProQuit:
		switch opType {
		case mpb.OpType_Put:
			return k.handleProQuit(km, metaValue, from)
		case mpb.OpType_Get:
			return k.handleProQuitSign(km, sig, from)
		}
	case mpb.KeyType_MoveData:
		switch opType {
		case mpb.OpType_Get:
			return k.handleMoveData(km, from)
		}
	default:
		switch opType {
		case mpb.OpType_Put:
			go k.handlePutKey(km, metaValue, sig, from)
		case mpb.OpType_Get:
			return k.handleGetKey(km, metaValue, sig, from)
		case mpb.OpType_Delete:
			go k.handleDeleteKey(km, metaValue, sig, from)
		default:
			return nil, metainfo.ErrWrongType
		}
	}
	return []byte(instance.MetaHandlerComplete), nil
}

func (k *Info) handlePutKey(km *metainfo.Key, metaValue, sig []byte, from string) {
	utils.MLogger.Info("handlePutKey: ", km.ToString())
	ctx := k.context
	ok := k.ds.VerifyKey(ctx, km.ToString(), metaValue, sig)
	if !ok {
		return
	}

	k.ds.PutKey(ctx, km.ToString(), metaValue, sig, "local")
}

func (k *Info) handleGetKey(km *metainfo.Key, metaValue, sig []byte, from string) ([]byte, error) {
	utils.MLogger.Info("handleGetKey: ", km.ToString())

	return k.ds.GetKey(k.context, km.ToString(), "local")
}

func (k *Info) handlePutStPaySign(km *metainfo.Key, metaValue, sig []byte, from string) {
	utils.MLogger.Infof("handlePutSign: %s, from %s", km.ToString(), from)
	// verify sig first
	// putSig
	ops := km.GetOptions()
	if len(ops) < 5 {
		return
	}

	gp := k.getGroupInfo(ops[0], km.GetMainID(), false)
	if gp == nil {
		return
	}

	linfo := gp.getLInfo(ops[1], false)
	// verify currentPay
	if linfo == nil || linfo.currentPay == nil || len(linfo.currentPay.GetSign()) < len(gp.keepers) {
		return
	}

	capy := linfo.currentPay

	for i, kid := range gp.keepers {
		if kid == string(metaValue) {
			capy.Lock()
			capy.Status--
			capy.GetSign()[i] = sig
			capy.Unlock()
			return
		}
	}
}

// key is /qid/"Sign"/uid/pid/kid/stStart/length
// value is hash
func (k *Info) handleGetStPaySign(km *metainfo.Key, metaValue, sig []byte, from string) {
	utils.MLogger.Info("handleGetSign: ", km.ToString())
	// verify sig first
	// verify metaValue
	// sign it
	nsig, err := id.Sign(k.sk, metaValue)
	if err != nil {
		return
	}

	k.ds.SendMetaRequest(k.context, int32(mpb.OpType_Put), km.ToString(), []byte(k.localID), nsig, from)
}

func (k *Info) handlePutStPayShare(km *metainfo.Key, metaValue []byte, from string) {
	utils.MLogger.Info("handlePutShare: ", km.ToString())

	ops := km.GetOptions()
	if len(ops) < 5 {
		return
	}
	gp := k.getGroupInfo(ops[0], km.GetMainID(), false)
	if gp == nil {
		return
	}
	linfo := gp.getLInfo(ops[1], false)
	if linfo == nil || linfo.currentPay == nil || len(linfo.currentPay.Share) < len(gp.keepers) {
		return
	}
	cpay := linfo.currentPay

	buf := bytes.NewBuffer(metaValue)
	var chalfrequency int64
	binary.Read(buf, binary.BigEndian, &chalfrequency)
	utils.MLogger.Debug("handlePutShare: ", km.ToString(), "chalfrequency:", chalfrequency)

	for i, kid := range gp.keepers {
		if kid == ops[2] {
			cpay.Share[i] = chalfrequency
			return
		}
	}
}

func (k *Info) handleGetStPayShare(km *metainfo.Key, metaValue []byte, from string) {
	utils.MLogger.Info("handleGetShare: ", km.ToString())

	ops := km.GetOptions()
	if len(ops) < 5 {
		return
	}
	gp := k.getGroupInfo(ops[0], km.GetMainID(), false)
	if gp == nil {
		return
	}
	linfo := gp.getLInfo(ops[1], false)
	if linfo == nil {
		return
	}

	stStart, err := time.Parse(utils.BASETIME, ops[3])
	if err != nil {
		return
	}
	stEnd, err := time.Parse(utils.BASETIME, ops[4])
	if err != nil {
		return
	}
	chalfrequency := linfo.stShare(stStart.Unix(), stEnd.Unix())

	km.Options[2] = k.localID

	tmp := make([]byte, 0)
	buf := bytes.NewBuffer(tmp)
	binary.Write(buf, binary.BigEndian, chalfrequency)

	utils.MLogger.Debug("handleGetShare: ", km.ToString(), "chalfrequency:", chalfrequency, buf.Bytes())

	k.ds.SendMetaRequest(k.context, int32(mpb.OpType_Put), km.ToString(), buf.Bytes(), nil, from)
}

func (k *Info) handleGetProAddSign(km *metainfo.Key, metaValue, sig []byte, from string) ([]byte, error) {
	utils.MLogger.Info("handleGetSign: ", km.ToString())
	// verify sig first
	// verify metaValue
	// sign it
	nsig, err := id.Sign(k.sk, metaValue)
	if err != nil {
		return nil, err
	}

	return nsig, nil
}

func (k *Info) handleAddBucket(km *metainfo.Key, metaValue, sig []byte, from string) {
	utils.MLogger.Info("handleAddBucket: ", km.ToString())
	ctx := k.context
	ok := k.ds.VerifyKey(ctx, km.ToString(), metaValue, sig)
	if !ok {
		return
	}

	ops := km.GetOptions()
	if len(ops) != 2 {
		return
	}

	binfo := new(mpb.BucketInfo)
	err := proto.Unmarshal(metaValue, binfo)
	if err != nil {
		return
	}

	k.ds.PutKey(ctx, km.ToString(), metaValue, sig, "local")
	k.addBucket(km.GetMainID(), ops[1], binfo)
}

func (k *Info) handleDeleteKey(km *metainfo.Key, metaValue, sig []byte, from string) {
	utils.MLogger.Info("handleDeleteKey: ", km.ToString())
	ctx := k.context
	ok := k.ds.VerifyKey(ctx, km.ToString(), metaValue, sig)
	if !ok {
		return
	}

	k.ds.DeleteKey(ctx, km.ToString(), "local")
}

// key: blockID/"BlockPos"
// value: pid/offset
func (k *Info) handleAddBlockPos(km *metainfo.Key, metaValue, sig []byte, from string) {
	utils.MLogger.Info("handleAddBlockPos: ", km.ToString())

	blockID := km.GetMainID()

	sValue := strings.Split(string(metaValue), metainfo.DELIMITER)
	if len(sValue) != 2 {
		utils.MLogger.Info("handleBlockPos err: ", metainfo.ErrIllegalValue)
		return
	}
	offset, err := strconv.Atoi(sValue[1])
	if err != nil {
		utils.MLogger.Info("handleBlockPos err: ", err)
		return
	}

	bids := strings.SplitN(blockID, metainfo.BlockDelimiter, 2)
	err = k.addBlockMeta(bids[0], bids[1], sValue[0], offset, true)
	if err != nil {
		utils.MLogger.Error("handleBlockPos err: ", err)
	}
	return
}

func (k *Info) handleDeleteBlockPos(km *metainfo.Key, metaValue, sig []byte, from string) {
	utils.MLogger.Info("handleDeleteBlockPos: ", km.ToString())
	blockID := km.GetMainID()

	// delete from local
	k.ds.DeleteKey(k.context, km.ToString(), "local")

	// delete from mem
	bids := strings.SplitN(blockID, metainfo.BlockDelimiter, 2)
	// send to other keepers?
	k.deleteBlockMeta(bids[0], bids[1], false)
}

// key: "Storage"/pid; value: total/used
func (k *Info) handleStorage(km *metainfo.Key, value []byte, pid string) {
	utils.MLogger.Info("handleStorage: ", km.ToString())
	vals := strings.Split(string(value), metainfo.DELIMITER)
	if len(vals) < 2 {
		return
	}

	total, err := strconv.ParseUint(vals[0], 10, 64)
	if err != nil {
		utils.MLogger.Info("handleStorageSync err: ", err)
		return
	}

	used, err := strconv.ParseUint(vals[1], 10, 64)
	if err != nil {
		utils.MLogger.Info("handleStorageSync err: ", err)
		return
	}

	thisInfo, err := k.getPInfo(pid, false)
	if err != nil {
		return
	}
	thisInfo.maxSpace = total
	thisInfo.usedSpace = used
}

func (k *Info) handleDelExAddr(km *metainfo.Key) {
	utils.MLogger.Info("handleDelExternnalAddr: ", km.ToString())

	pid := km.GetMainID()
	pi, err := k.getPInfo(pid, false)
	if err != nil {
		ki, err := k.getKInfo(pid, false)
		if err == nil {
			k.deleteIDByIP(pid, ki.eAddr, true)
			ki.eAddr = ""
		}
	} else {
		k.deleteIDByIP(pid, pi.eAddr, false)
		pi.eAddr = ""
	}
}

func (k *Info) handlePutExAddr(km *metainfo.Key, value []byte, from string) ([]byte, error) {
	utils.MLogger.Info("handlePutExternnalAddr: ", km.ToString(), string(value))

	if string(value) == "" {
		return []byte(instance.MetaHandlerComplete), nil
	}

	ea := strings.Split(string(value), "/")
	if len(ea) >= 5 {
		ipa := ea[2] + ":" + ea[4]
		if !utils.IsReachable(ipa) {
			return nil, role.ErrNotConnectd
		}
	}

	pid := km.GetMainID()
	pi, err := k.getPInfo(pid, false)
	if err != nil {
		ki, err := k.getKInfo(pid, false)
		if err == nil {
			k.putPeerIDByIP(ki.eAddr, string(value), pid, true)
			ki.eAddr = string(value)
		}
	} else {
		k.putPeerIDByIP(pi.eAddr, string(value), pid, false)
		pi.eAddr = string(value)
	}

	return []byte(instance.MetaHandlerComplete), nil
}

func (k *Info) handleExternalAddr(km *metainfo.Key) ([]byte, error) {
	utils.MLogger.Info("handleExternnalAddr: ", km.ToString())
	pid := km.GetMainID()
	var addr string
	thisInfoI, ok := k.providers.Load(pid)
	if ok {
		addr = thisInfoI.(*pInfo).eAddr
	}

	if addr == "" {
		thisInfoI, ok := k.keepers.Load(pid)
		if ok {
			addr = thisInfoI.(*kInfo).eAddr
		}
	}

	if addr != "" {
		maddr, err := ma.NewMultiaddr(addr)
		if err == nil {
			ok := mnet.IsThinWaist(maddr)
			if ok {
				// is ip4/tcp or ip4/udp
				ok = mnet.IsPrivateAddr(maddr)
				if !ok {
					// is public addr
					return maddr.Bytes(), nil
				}
			} else {
				// is /dns/...
				addrs, err := mdns.Resolve(k.context, maddr)
				if err != nil {
					return nil, err
				}

				for _, maddr := range addrs {
					ok = mnet.IsPrivateAddr(maddr)
					if !ok {
						return maddr.Bytes(), nil
					}
				}
			}
		}
	}

	//neither keeper nor provider, users or self
	maddr, err := k.ds.GetExternalAddr(pid)
	if err != nil {
		return nil, err
	}
	return maddr.Bytes(), nil
}

func (k *Info) handleChalTime(km *metainfo.Key) ([]byte, error) {
	utils.MLogger.Info("handleChalTime: ", km.ToString())

	blockID := km.GetMainID()
	if len(blockID) <= utils.IDLength {
		return nil, role.ErrWrongKey
	}

	sValue := strings.SplitN(string(blockID), metainfo.BlockDelimiter, 2)
	if len(sValue) != 2 {
		return nil, role.ErrWrongValue
	}
	qid := sValue[0]
	bid := sValue[1]
	avail, err := k.getBlockAvail(qid, bid)
	if err != nil {
		return nil, err
	}

	return []byte(utils.UnixToString(avail)), nil
}

//value: /bid_sid_cid_offset_newPid/bid_sid_cid_offset_newPid..
func (k *Info) handleProQuit(km *metainfo.Key, value []byte, from string) ([]byte, error) {
	utils.MLogger.Info("handle provider quit: ", km.ToString())

	groupID := km.GetMainID()
	ops := km.GetOptions() //uid
	if len(ops) != 1 {
		utils.MLogger.Debug("handleProQuit km's length is wrong")
		return nil, nil
	}

	utils.MLogger.Debug("handle provider quit, the value: ", string(value))
	responses := strings.Split(string(value), metainfo.DELIMITER)
	for _, res := range responses {
		mes := strings.Split(res, metainfo.BlockDelimiter)
		if len(mes) != 5 {
			continue
		}

		//bid: bucketid_stripeid_chunkid
		bid := mes[0] + metainfo.BlockDelimiter + mes[1] + metainfo.BlockDelimiter + mes[2]
		offset, err := strconv.Atoi(mes[3])
		if err != nil {
			continue
		}
		utils.MLogger.Info("begin update local blockMeta: ", res)
		k.deleteBlockMeta(groupID, bid, true)
		k.addBlockMeta(groupID, bid, mes[4], offset, true)
	}

	gp := k.getGroupInfo(ops[0], groupID, true)
	if gp == nil {
		utils.MLogger.Debug("handleProQuit gp is nil")
		return nil, nil
	}

	proAddr, err := address.GetAddressFromID(from)
	if err != nil {
		utils.MLogger.Debug("transfer proID to addr fails")
		return nil, err
	}

	ukAddr, err := address.GetAddressFromID(gp.upkeeping.UpKeepingID)
	if err != nil {
		utils.MLogger.Debug("transfer ukID to addr fails")
		return nil, err
	}

	localAddr, err := address.GetAddressFromID(k.localID)
	if err != nil {
		utils.MLogger.Debug("transfer localID to addr fails")
		return nil, err
	}

	userAddr, err := address.GetAddressFromID(ops[0])
	if err != nil {
		utils.MLogger.Debug("transfer uid to addr fails")
		return nil, err
	}

	sign, err := role.SignForSetStop(ukAddr, proAddr, k.sk)
	if err != nil {
		utils.MLogger.Debug("signForSetStop fails")
		return nil, err
	}

	if !gp.isMaster(from) {
		utils.MLogger.Info("not master Keeper, ", km.ToString(), "begin send sign of setProviderStop")

		//send setProviderStop sign to masterKeeper

		km, err := metainfo.NewKey(gp.groupID, mpb.KeyType_ProQuit, from)
		if err != nil {
			utils.MLogger.Debug("newKey error:", err)
			return nil, err
		}
		k.ds.SendMetaRequest(k.context, int32(mpb.OpType_Get), km.ToString(), nil, sign, gp.masterKeeper)

		utils.MLogger.Info("send sign of setProviderStop success: ", km.ToString())

		// update uk info
		//有可能更新时，masterKeeper还未完成setStop操作..
		gp.loadContracts(true)
		return nil, nil
	}

	//master keeper, get signs to call setProviderStop in upkeeping
	utils.MLogger.Info("begin get sign for set provider ", from, " stop in group ", groupID)

	linfo := gp.getLInfo(from, false)
	if linfo == nil {
		utils.MLogger.Debug("handleProQuit getLInfo is nil")
		return nil, errors.New("get sign for set provider stop, linfo is nil")
	}
	if linfo.stopSign == nil {
		linfo.stopSign = make(map[string][]byte)
	}
	linfo.stopSign[k.localID] = sign

	for i := 0; i < 20; i++ {
		time.Sleep(time.Minute)
		if len(linfo.stopSign) >= len(gp.keepers)*2/3 {
			break
		}
	}

	if len(linfo.stopSign) >= len(gp.keepers)*2/3 {
		sigs := make([][]byte, len(gp.keepers))
		for i, kid := range gp.keepers {
			sig := linfo.stopSign[kid]
			sigs[i] = sig
		}

		cu := contracts.NewCU(localAddr, k.sk)
		err = cu.SetProviderStop(userAddr, proAddr, ukAddr, "", sigs)
		if err != nil {
			utils.MLogger.Debug("setProviderStop fails, provider: ", proAddr.String(), "ukAddr: ", ukAddr.String(), "err: ", err)
			return nil, err
		}

		utils.MLogger.Info("successfully set provider ", from, " stop in group ", groupID)

		//tell provider finished setStop successfully
		km, err := metainfo.NewKey(gp.groupID, mpb.KeyType_ProQuit)
		if err != nil {
			utils.MLogger.Debug("newKey error:", err)
			return nil, err
		}
		k.ds.SendMetaRequest(k.context, int32(mpb.OpType_Put), km.ToString(), nil, nil, from)

		//add provider:1.find a new provider
		newPro, err := k.findNewProvider(gp.upkeeping.Price, gp.upkeeping.Capacity, gp.upkeeping.Duration, gp.providers)
		if err != nil {
			utils.MLogger.Error("findNewProvider fails:", err)
			return nil, err
		}

		//add provider:2.add
		err = k.ukAddProvider(ops[0], groupID, newPro)
		if err != nil {
			utils.MLogger.Error("ukAddProvider fails:", err)
		}

		return nil, err
	}

	utils.MLogger.Warn("sigs of setProviderStop is not enough")
	return nil, errors.New("signs of setProvidderStop is not enough")
}

func (k *Info) handleProQuitSign(km *metainfo.Key, sig []byte, from string) ([]byte, error) {
	utils.MLogger.Info("handle get sign of proQuit ", km.ToString(), " from ", from)

	groupID := km.GetMainID()
	ops := km.GetOptions() // pid
	if len(ops) != 1 {
		utils.MLogger.Debug("handleProQuitSign km's length is wrong")
		return nil, errors.New("handleProQuitSign km's length is wrong")
	}

	gp := k.getGroupInfo(ops[0], groupID, false)
	if gp == nil {
		utils.MLogger.Debug("handleProQuitSign gp is nil")
		return nil, errors.New("handleProQuitSign gp is nil")
	}

	linfo := gp.getLInfo(ops[0], false)
	if linfo == nil {
		return nil, errors.New("handleProQuitSign linfo is nil")
	}

	if linfo.stopSign == nil {
		linfo.stopSign = make(map[string][]byte)
	}
	linfo.stopSign[from] = sig

	return nil, nil
}

func (k *Info) handleMoveData(km *metainfo.Key, from string) ([]byte, error) {
	utils.MLogger.Info("handle moveData for groupID ", km.GetMainID, "and for provider ", km.GetOptions())

	groupID := km.GetMainID()
	ops := km.GetOptions()
	if len(ops) != 1 {
		utils.MLogger.Debug("handleMoveData km's length is wrong")
		return nil, nil
	}

	gp := k.getGroupInfo(groupID, groupID, false)
	if gp == nil {
		utils.MLogger.Debug("handleProQuit gp is nil")
		return nil, nil
	}

	thisLinfo := gp.getLInfo(ops[0], false)
	if thisLinfo == nil {
		utils.MLogger.Debug("handleMoveData getLInfo fails")
		return nil, errors.New("get linfo fails")
	}

	var res string
	thisLinfo.blockMap.Range(func(key, value interface{}) bool {
		cInfo := value.(*blockInfo)
		//bid_sid_cid_offset
		bID := key.(string) + metainfo.BlockDelimiter + strconv.Itoa(cInfo.offset)
		utils.MLogger.Info("get provider ", ops[0], " block: ", bID)

		binfo := strings.Split(bID, metainfo.BlockDelimiter)
		if len(binfo) != 4 {
			return false
		}

		//search new provider; 1. getproviders
		thisbucket := gp.getBucketInfo(binfo[0], false)
		if thisbucket == nil {
			return false
		}

		count := thisbucket.chunkNum

		var r strings.Builder
		ugid := make([]string, 0, count)
		for i := 0; i < count; i++ {
			r.Reset()
			r.WriteString(binfo[1])
			r.WriteString(metainfo.BlockDelimiter)
			r.WriteString(strconv.Itoa(i))
			thisinfo, ok := thisbucket.stripes.Load(r.String())
			if !ok {
				continue
			}

			pid := thisinfo.(*blockInfo).storedOn
			ugid = append(ugid, pid)
		}

		if len(ugid) == 0 {
			utils.MLogger.Debug("the ugid of block", bID, "is nil, information is not enough")
			return false
		}

		//search new provider; 2. search
		response := k.searchNewProvider(k.context, groupID, ugid)

		//bid_sid_cid_offset_newPid
		bID += metainfo.BlockDelimiter + response
		res = res + bID + metainfo.DELIMITER

		return true
	})

	return []byte(res[:len(res)-1]), nil
}
