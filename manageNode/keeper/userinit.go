package keeper

import (
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p-core/peer"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
)

// handleUserInit collect keepers and providers for user
// return kv, key: queryID/"UserInit"/userID/keepercount/providercount;
// value: kid1kid2../pid1pid2..
func (k *Info) handleUserInit(km *metainfo.Key, from string) {
	utils.MLogger.Info("handleUserInit: ", km.ToString(), " from: ", from)
	options := km.GetOptions()
	if len(options) != 3 {
		return
	}

	kc, err := strconv.Atoi(options[1])
	if err != nil {
		utils.MLogger.Error("handleUserInitReq: ", err)
		return
	}

	pc, err := strconv.Atoi(options[2])
	if err != nil {
		utils.MLogger.Error("handleUserInitReq: ", err)
		return
	}

	uid := options[0]
	qid := km.GetMid()
	price := big.NewInt(utils.STOREPRICEPEDOLLAR)
	var response string
	if qid != uid {
		utils.MLogger.Infof("Get k/p numbers from query contract %s of user %s ", qid, uid)
		item, err := role.GetQueryInfo(uid, qid)
		if err != nil {
			utils.MLogger.Error("get query: ", qid, " error: ", err)
			return
		}
		kc = int(item.KeeperNums)
		pc = int(item.ProviderNums)
		price = item.Price
	}

	if pos.GetPosId() == uid {
		price = pos.GetPosPrice()
	}

	response, err = k.initUser(uid, qid, kc, pc, price)
	if err != nil {
		if err != nil {
			utils.MLogger.Info("handleUserInitReq err: ", err)
			return
		}
	}
	utils.MLogger.Info("New fs: ", qid, " for user: ", uid, " keeperCount: ", kc, " providerCount: ", pc, " price: ", price)

	k.ds.SendMetaRequest(k.context, int32(mpb.OpType_Put), km.ToString(), []byte(response), nil, from)
}

//response: kid1kid2../pid1pid2..
func (k *Info) initUser(uid, gid string, kc, pc int, price *big.Int) (string, error) {
	var newResponse strings.Builder

	gp := k.getGroupInfo(uid, gid, false)
	if gp == nil {
		localID := k.localID
		// fill self
		newResponse.WriteString(localID)
		kc--
		//fill other keepers
		keepers, err := k.GetKeepers()
		if err != nil {
			return "", err
		}
		for _, kid := range keepers {
			if kc == 0 {
				break
			}
			if kid == localID {
				continue
			}

			thisinfo, ok := k.keepers.Load(kid)
			if ok && thisinfo.(*kInfo).online == true {
				newResponse.WriteString(kid)
				kc--
			}
		}

		newResponse.WriteString(metainfo.DELIMITER)

		pros, err := k.GetProviders()
		if err != nil {
			return "", err
		}
		// fill providers
		for _, proID := range pros {
			if pc == 0 {
				break
			}
			if proID == localID {
				continue
			}

			thisinfo, ok := k.providers.Load(proID)
			if ok {
				thisP := thisinfo.(*pInfo)
				if thisP.online && thisP.offerItem != nil {
					if thisP.offerItem.Price.Cmp(price) <= 0 && thisP.credit > 0 {
						newResponse.WriteString(proID)
						kc--
					} else {
						utils.MLogger.Debugf("provider %s need price %d, but %d; has credit: %d", proID, thisP.offerItem.Price, price, thisP.credit)
					}
				}
			}
		}

		return newResponse.String(), nil
	}

	// user has init
	for _, pid := range gp.keepers {
		newResponse.WriteString(pid)
	}

	newResponse.WriteString(metainfo.DELIMITER)

	for _, pid := range gp.providers {
		newResponse.WriteString(pid)
	}
	return newResponse.String(), nil
}

// handleUserNotify return kv,
// key: queryID/"UserNotify"/userID/keepercount/providercount;
// value: kid1kid2../pid1pid2..
func (k *Info) handleUserNotify(km *metainfo.Key, metaValue []byte, from string) ([]byte, error) {
	utils.MLogger.Info("handleUserNotify: ", km.ToString(), " from:", from)

	splited := strings.Split(string(metaValue), metainfo.DELIMITER)
	if len(splited) < 2 {
		utils.MLogger.Info("UserNotif value is not correct: ", metaValue)
		return []byte("no"), nil
	}

	kids := splited[0]
	for i := 0; i < len(kids)/utils.IDLength; i++ {
		keeper := string(kids[i*utils.IDLength : (i+1)*utils.IDLength])
		if keeper == k.localID {
			return []byte("ok"), nil
		}
	}

	return []byte("ok"), nil
}

func (k *Info) handleHeartBeat(km *metainfo.Key, metaValue []byte, from string) ([]byte, error) {
	utils.MLogger.Info("handleHeartBeat: ", km.ToString(), " from:", from)

	ops := km.GetOptions()
	if len(ops) != 4 {
		return nil, role.ErrWrongKey
	}

	uid := ops[0]
	qid := km.GetMid()

	gp := k.getGroupInfo(uid, qid, false)
	if gp != nil {
		sessID, err := uuid.Parse(ops[3])
		if err != nil {
			return nil, err
		}

		if gp.sessionID == uuid.Nil {
			gp.sessionID = sessID
		}

		if gp.sessionID == sessID {
			gp.sessionTime = time.Now().Unix()
		}

		sessID = gp.sessionID

		return []byte(sessID.String()), nil
	}

	return nil, role.ErrNotMyUser
}

// key: queryID/"UserStop"/userID/keepercount/providercount/sessionID;
func (k *Info) handleUserStop(km *metainfo.Key, metaValue, sig []byte, from string) ([]byte, error) {
	utils.MLogger.Info("handleUserStop: ", km.ToString(), " from:", from)

	ops := km.GetOptions()
	if len(ops) != 4 {
		return nil, role.ErrWrongKey
	}

	uid := ops[0]
	qid := km.GetMid()

	gp := k.getGroupInfo(uid, qid, false)
	if gp != nil {
		ok := k.ds.VerifyKey(k.context, km.ToString(), metaValue, sig)
		if !ok {
			utils.MLogger.Infof("key signature is wrong for %s", km.ToString())
			return nil, role.ErrWrongSign
		}
		sessID, err := uuid.Parse(ops[3])
		if err != nil {
			return nil, err
		}
		if gp.sessionID == sessID {
			gp.sessionID = uuid.Nil
			gp.sessionTime = time.Now().Unix()
		}
		return []byte("ok"), nil
	}

	return nil, role.ErrNotMyUser
}

// key: queryID/"UserStart"/userID/keepercount/providercount/sessionID/"force";
// value: kid1kid2../pid1pid2..
func (k *Info) handleUserStart(km *metainfo.Key, metaValue, sig []byte, from string) ([]byte, error) {
	utils.MLogger.Info("handleUserStart: ", km.ToString(), " from:", from)
	splited := strings.Split(string(metaValue), metainfo.DELIMITER)
	if len(splited) < 2 {
		utils.MLogger.Info("UserStart value is not correct: ", metaValue)
		return nil, role.ErrWrongValue
	}

	ops := km.GetOptions()
	if len(ops) != 5 {
		return nil, role.ErrWrongKey
	}

	kc, err := strconv.Atoi(ops[1])
	if err != nil {
		utils.MLogger.Info("handleUserInitReq: ", err)
		return nil, err
	}

	pc, err := strconv.Atoi(ops[2])
	if err != nil {
		utils.MLogger.Info("handleUserInitReq: ", err)
		return nil, err
	}

	uid := ops[0]
	qid := km.GetMid()

	k.fillPinfo(uid, qid, pc, kc, metaValue, from)

	gp := k.getGroupInfo(uid, qid, true)
	if gp != nil {
		if ops[4] == "0" && gp.sessionID != uuid.Nil && time.Now().Unix()-gp.sessionTime < expireTime {
			return []byte(gp.sessionID.String()), nil
		}
		ok := k.ds.VerifyKey(k.context, km.ToString(), metaValue, sig)
		if !ok {
			utils.MLogger.Infof("key signature is wrong for %s", km.ToString())
			return []byte(uuid.Nil.String()), nil
		}

		sessID, err := uuid.Parse(ops[3])
		if err != nil {
			return nil, err
		}
		gp.sessionTime = time.Now().Unix()
		gp.sessionID = sessID

		return []byte(sessID.String()), nil
	}

	return nil, role.ErrNotMyUser
}

// fillPinfo fill user's uInfo, groupInfo in ukpMap
func (k *Info) fillPinfo(userID, groupID string, kc, pc int, metaValue []byte, from string) (*groupInfo, error) {
	//将value切分，生成好对应的keepers和providers列表
	splited := strings.Split(string(metaValue), metainfo.DELIMITER)
	if len(splited) < 2 {
		utils.MLogger.Info("UserNotif value is not correct: ", metaValue)
		return nil, role.ErrWrongValue
	}

	var keepers, providers []string
	kids := splited[0]
	for i := 0; i < len(kids)/utils.IDLength; i++ {
		keeper := string(kids[i*utils.IDLength : (i+1)*utils.IDLength])
		_, err := peer.IDB58Decode(keeper)
		if err != nil {
			continue
		}
		keepers = append(keepers, keeper)
	}

	pids := splited[1]
	for i := 0; i < len(pids)/utils.IDLength; i++ {
		providerID := string(pids[i*utils.IDLength : (i+1)*utils.IDLength])
		_, err := peer.IDB58Decode(providerID)
		if err != nil {
			continue
		}
		providers = append(providers, providerID)
	}

	gp, err := k.createGroup(userID, groupID, keepers, providers)
	if err != nil {
		return nil, err
	}

	ui, _ := k.getUInfo(userID)
	if ui != nil {
		ui.setQuery(groupID)
	}

	kmkps, err := metainfo.NewKey(groupID, mpb.KeyType_LFS, userID)
	if err != nil {
		return nil, err
	}

	k.putKey(k.context, kmkps.ToString(), metaValue, nil, "local", gp.clusterID, gp.bft)
	return gp, nil
}
