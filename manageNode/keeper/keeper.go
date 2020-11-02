package keeper

import (
	"context"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
	metrics "github.com/ipfs/go-metrics-interface"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	id "github.com/memoio/go-mefs/crypto/identity"
	mpb "github.com/memoio/go-mefs/pb"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/source/data"
	datastore "github.com/memoio/go-mefs/source/go-datastore"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
)

//Info implements user service
type Info struct {
	localID       string
	role          string
	sk            string
	state         bool
	enableBft     bool
	context       context.Context
	raftNodeID    uint64
	repch         chan string
	ds            data.Service
	pledgeStorage *big.Int
	keepers       sync.Map // keepers except self; value: *kInfo
	providers     sync.Map // value: *pInfo
	users         sync.Map // value: *uInfo
	ukpGroup      sync.Map // manage user-keeper-provider group, value: *group
	netIDs        map[string]struct{}
	userConfigs   *lru.ARCCache
	kItem         *role.KeeperItem
	ms            *measure
	ManageIncome  *big.Int //keeper's manage-income
	PosIncome     *big.Int //keeper's pos-manage-income
}

type measure struct {
	balance        metrics.Gauge
	userNum        metrics.Gauge
	groupNum       metrics.Gauge
	masterGroupNum metrics.Gauge
	keeperNum      metrics.Gauge
	providerNum    metrics.Gauge
	storageUsed    metrics.Gauge
	repairNum      metrics.Gauge
	faultNum       metrics.Gauge
}

// New is
func New(ctx context.Context, nid, sk string, d data.Service, rt routing.Routing) (instance.Service, error) {
	mea := &measure{
		balance:        metrics.New("keeper.balance", "Balance of this keeper").Gauge(),
		userNum:        metrics.New("keeper.user_num", "User number").Gauge(),
		groupNum:       metrics.New("keeper.group_num", "Group number").Gauge(),
		masterGroupNum: metrics.New("keeper.masterGroup_num", "master group number").Gauge(),
		keeperNum:      metrics.New("keeper.keeper_num", "Keeper number").Gauge(),
		providerNum:    metrics.New("keeper.provider_num", "Providers number").Gauge(),
		storageUsed:    metrics.New("keeper.storage_used", "Storage used(bytes)").Gauge(),
		repairNum:      metrics.New("keeper.repair_num", "Repair number").Gauge(),
		faultNum:       metrics.New("keeper.fault_num", "Fault block number").Gauge(),
	}

	m := &Info{
		localID:       nid,
		sk:            sk,
		state:         false,
		ds:            d,
		repch:         make(chan string, 1024),
		netIDs:        make(map[string]struct{}),
		context:       ctx,
		ms:            mea,
		pledgeStorage: big.NewInt(0),
		ManageIncome:  big.NewInt(0),
		PosIncome:     big.NewInt(0),
	}

	balance := role.GetBalance(m.localID)
	ba, _ := new(big.Float).SetInt(balance).Float64()
	m.ms.balance.Set(ba)

	usedCapacity, err := datastore.DiskUsage(m.ds.DataStore())
	if err == nil {
		m.ms.storageUsed.Set(float64(usedCapacity))
	}

	m.ms.keeperNum.Inc() // add self

	pubKey, err := id.GetCompressPubByte(m.sk)
	if err != nil {
		return nil, err
	}

	kmp, err := metainfo.NewKey(m.localID, mpb.KeyType_PublicKey)
	if err != nil {
		return nil, err
	}

	m.ds.PutKey(ctx, kmp.ToString(), pubKey, nil, "local")

	err = m.loadContract(true)
	if err != nil {
		return nil, err
	}

	// cache userconfigs, key is queryID
	ucache, err := lru.NewARC(1024)
	if err != nil {
		utils.MLogger.Error("new lru err:", err)
		return nil, err
	}
	m.userConfigs = ucache

	err = m.load(ctx) //连接节点
	if err != nil {
		utils.MLogger.Error("load err:", err)
		return nil, err
	}

	go m.persistRegular(ctx)
	go m.challengeRegular(ctx)
	go m.cleanTestUsersRegular(ctx)
	go m.checkLedgerRafi(ctx)
	go m.repairRegular(ctx)
	go m.stPrePayRegular(ctx)
	go m.stPayRegular(ctx)
	go m.checkPeers(ctx) //check if connect
	go m.getFromChainRegular(ctx)

	err = rt.(*dht.KadDHT).AssignmetahandlerV2(m)
	if err != nil {
		return nil, err
	}

	m.state = true
	utils.MLogger.Info("Keeper Service is ready")
	return m, nil
}

// Online is
func (k *Info) Online() bool {
	return k.state
}

// GetRole is
func (k *Info) GetRole() string {
	return metainfo.RoleKeeper
}

// Close is
func (k *Info) Close() error {
	err := k.save(k.context)
	return err
}

/*====================Save and Load========================*/

func (k *Info) persistRegular(ctx context.Context) {
	utils.MLogger.Info("Persist local peerInfo start!")
	ticker := time.NewTicker(persistTime)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := k.save(ctx)
			if err != nil {
				utils.MLogger.Error("Persist local peerInfo err:", err)
			}
		}
	}
}

//persist k.keepers、k.providers、k.users、k.users.uInfo.querys、lastPay to local;
func (k *Info) save(ctx context.Context) error {
	localID := k.localID

	// persist keepers
	kmKID, err := metainfo.NewKey(localID, mpb.KeyType_Keepers)
	if err != nil {
		return err
	}

	var pids strings.Builder
	k.keepers.Range(func(key, value interface{}) bool {
		pids.WriteString(key.(string))
		return true
	})

	if pids.Len() > 0 {
		err = k.ds.PutKey(ctx, kmKID.ToString(), []byte(pids.String()), nil, "local")
		if err != nil {
			return err
		}
	}

	// persist providers
	pids.Reset()
	kmPID, err := metainfo.NewKey(localID, mpb.KeyType_Providers)
	if err != nil {
		return err
	}

	k.providers.Range(func(key, value interface{}) bool {
		pids.WriteString(key.(string))
		return true
	})

	if pids.Len() > 0 {
		err = k.ds.PutKey(ctx, kmPID.ToString(), []byte(pids.String()), nil, "local")
		if err != nil {
			return err
		}
	}

	pids.Reset()

	kmUID, err := metainfo.NewKey(localID, mpb.KeyType_Users)
	if err != nil {
		return err
	}

	var res strings.Builder
	k.users.Range(func(key, value interface{}) bool {
		uid := key.(string)
		pids.WriteString(uid)
		kmfs, err := metainfo.NewKey(uid, mpb.KeyType_Query)
		if err != nil {
			return true
		}

		res.Reset()
		for qid := range value.(*uInfo).querys {
			res.WriteString(qid)
		}

		// persist queryID of one user
		if res.Len() > 0 {
			err = k.ds.PutKey(ctx, kmfs.ToString(), []byte(res.String()), nil, "local")
			if err != nil {
				return true
			}
		}

		return true
	})

	// persist all users
	if pids.Len() > 0 {
		err = k.ds.PutKey(ctx, kmUID.ToString(), []byte(pids.String()), nil, "local")
		if err != nil {
			return err
		}
	}

	// save last pay
	qus := k.getQUKeys()
	stripeNum := -1
	for _, qu := range qus {
		gp := k.getGroupInfo(qu.uid, qu.qid, false)
		if gp == nil {
			continue
		}

		for _, proID := range gp.providers {
			k.savePay(qu.uid, qu.qid, proID)
		}

		if qu.uid == pos.GetPosId() {
			continue
		}

		kmBS, err := metainfo.NewKey(qu.qid, mpb.KeyType_BucketStripes, qu.uid)
		if err != nil {
			return err
		}

		res.Reset()

		for i := 0; i <= int(gp.bucketNum); i++ {
			stripeNum = -1
			bi, ok := gp.buckets.Load(strconv.Itoa(i))
			if ok {
				stripeNum = bi.(*bucketInfo).curStripes
			}
			res.WriteString(strconv.Itoa(stripeNum))
			res.WriteString(metainfo.DELIMITER)
		}

		k.ds.PutKey(ctx, kmBS.ToString(), []byte(res.String()), nil, "local")
	}

	return nil
}

func (k *Info) savePay(userID, qid, pid string) error {
	thisLinfo := k.getLInfo(qid, qid, pid, false)

	if thisLinfo != nil && thisLinfo.lastPay != nil && thisLinfo.lastPay.Status <= 0 {
		ctx := k.context
		lpay := thisLinfo.lastPay
		//key: qid/`lastpay"/pid`
		kmLast, err := metainfo.NewKey(qid, mpb.KeyType_LastPay, pid)
		if err != nil {
			return err
		}

		valueLast, err := proto.Marshal(&lpay.STValue)
		if err != nil {
			return err
		}

		clusterID, err := address.GetNodeIDFromID(qid)
		if err != nil {
			return err
		}

		k.putKey(ctx, kmLast.ToString(), []byte(valueLast), nil, "local", clusterID, true)

		//key: `qid/"chalpay"/userID/pid/kid/beginTime/length`
		km, err := metainfo.NewKey(qid, mpb.KeyType_ChalPay, userID, pid, k.localID, utils.UnixToString(lpay.GetStart()), utils.UnixToString(lpay.GetLength()))
		if err != nil {
			return err
		}

		k.putKey(ctx, km.ToString(), valueLast, nil, "local", clusterID, true)
	}
	return nil
}

func (k *Info) load(ctx context.Context) error {
	k.loadPeers(ctx)
	k.loadPeersFromChain()
	k.loadUser(ctx)
	return nil
}

//重启后重新恢复User现场 读取本地存储的U-K-P信息，构建PInfo结构
func (k *Info) loadUser(ctx context.Context) error {
	utils.MLogger.Info("Load All userID's Information")
	localID := k.localID //本地id
	kmUID, err := metainfo.NewKey(localID, mpb.KeyType_Users)
	if err != nil {
		return err
	}

	if users, err := k.ds.GetKey(ctx, kmUID.ToString(), "local"); users != nil && err == nil {
		for i := 0; i < len(users)/utils.IDLength; i++ { //对user进行循环，逐个恢复user信息
			userID := string(users[i*utils.IDLength : (i+1)*utils.IDLength])
			_, err := peer.IDB58Decode(userID)
			if err != nil {
				continue
			}

			utils.MLogger.Info("Load user: ", userID, " 's infomations begin")
			go func(userID string) {
				kmfs, err := metainfo.NewKey(userID, mpb.KeyType_Query)
				if err != nil {
					return
				}

				ui, err := k.getUInfo(userID)
				if err != nil {
					return
				}

				qs, err := k.ds.GetKey(ctx, kmfs.ToString(), "local")
				if err != nil {
					return
				}

				for i := 0; i < len(qs)/utils.IDLength; i++ {
					qid := string(qs[i*utils.IDLength : (i+1)*utils.IDLength])
					_, err := peer.IDB58Decode(qid)
					if err != nil {
						continue
					}

					utils.MLogger.Info("Load user: ", userID, " 's query: ", qid)

					ui.setQuery(qid)

					err = k.newGroupWithFS(userID, qid, "", true)
					if err != nil {
						utils.MLogger.Error("Load user: ", userID, " 's query: ", qid, " fail: ", err)
						continue
					}
				}
				utils.MLogger.Info("Load user: ", userID, " 's infomations finished")
			}(userID)
		}
	}

	return nil
}

func (k *Info) loadUserBucketStripes(uid, qid string) error {
	// load bucketinfo
	if uid == pos.GetPosId() {
		return nil
	}

	kmBS, err := metainfo.NewKey(qid, mpb.KeyType_BucketStripes, uid)
	if err != nil {
		return err
	}

	res, err := k.ds.GetKey(k.context, kmBS.ToString(), "local")
	if err != nil {
		return err
	}

	gp := k.getGroupInfo(uid, qid, false)
	if gp == nil {
		return role.ErrNotMyUser
	}

	vals := strings.Split(string(res), metainfo.DELIMITER)
	for i, val := range vals {
		buc := gp.getBucketInfo(strconv.Itoa(i), true)
		snum, err := strconv.Atoi(val)
		if err != nil {
			continue
		}
		buc.curStripes = snum
	}

	return nil
}

func (k *Info) loadUserBucket(uid, qid string) error {
	// load bucketinfo
	prefix := qid + metainfo.DELIMITER + strconv.Itoa(int(mpb.KeyType_Bucket)) + metainfo.DELIMITER + uid + metainfo.DELIMITER

	es, _ := k.ds.Itererate(prefix)
	for _, e := range es {
		rec := new(mpb.Record)
		err := proto.Unmarshal(e.Value, rec)
		if err != nil {
			continue
		}

		utils.MLogger.Debug("Load bucket: ", string(rec.GetKey()))
		km, err := metainfo.NewKeyFromString(string(rec.GetKey()))
		if err != nil {
			continue
		}

		ops := km.GetOptions()
		if len(ops) != 2 {
			continue
		}

		if ops[0] != uid {
			continue
		}

		binfo := new(mpb.BucketInfo)
		err = proto.Unmarshal(rec.GetValue(), binfo)
		if err != nil {
			continue
		}
		k.addBucket(km.GetMainID(), ops[1], binfo)
	}
	return nil
}

func (k *Info) loadUserBlock(qid string) error {
	// load blockinfo
	prefix := qid + metainfo.BlockDelimiter
	es, _ := k.ds.Itererate(prefix)
	for _, e := range es {
		rec := new(mpb.Record)
		err := proto.Unmarshal(e.Value, rec)
		if err != nil {
			continue
		}

		utils.MLogger.Debug("Load block: ", string(rec.GetKey()))

		km, err := metainfo.NewKeyFromString(string(rec.GetKey()))
		if err != nil {
			continue
		}

		pids := strings.Split(string(rec.GetValue()), metainfo.DELIMITER)
		if len(pids) < 2 {
			continue
		}

		_, err = peer.IDB58Decode(pids[0])
		if err != nil {
			continue
		}

		off, err := strconv.Atoi(pids[1])
		if err != nil {
			continue
		}

		getID := strings.SplitN(km.GetMainID(), metainfo.BlockDelimiter, 2)
		if len(getID) != 2 || (len(getID) > 0 && getID[0] != qid) {
			continue
		}

		k.addBlockMeta(qid, getID[1], pids[0], off, false)
	}
	return nil
}

func (k *Info) loadUserPay(userID, qid string) error {
	ctx := k.context
	gInfo := k.getGroupInfo(userID, qid, true)
	if gInfo == nil {
		return role.ErrNotMyUser
	}

	if gInfo.userID == gInfo.groupID {
		return nil
	}

	payTime := gInfo.upkeeping.StartTime
	for _, proID := range gInfo.providers {
		lin := gInfo.getLInfo(proID, true)
		if lin == nil {
			continue
		}
		kmLast, err := metainfo.NewKey(qid, mpb.KeyType_LastPay, proID)
		if err != nil {
			return err
		}
		res, err := k.ds.GetKey(ctx, kmLast.ToString(), "local")
		if err == nil && len(res) > 0 {
			val := mpb.STValue{}
			err := proto.Unmarshal(res, &val)
			if err != nil {
				return err
			}
			lin.lastPay = &chalpay{
				STValue: val,
			}
		}

		if lin.lastPay == nil {
			found := false
			for _, pInfo := range gInfo.upkeeping.Providers {
				pid, err := address.GetIDFromAddress(pInfo.Addr.String())
				if err != nil {
					return err
				}
				if pid == proID {
					found = true
					payTime = pInfo.StEnd.Int64()
					break
				}
			}

			if !found {
				continue
			}

			lin.lastPay = &chalpay{
				STValue: mpb.STValue{
					Start:  payTime,
					Length: 0,
				},
			}
		}
	}
	return nil
}

func (k *Info) loadUserChallenge(userID, qid string) error {
	gInfo := k.getGroupInfo(userID, qid, true)
	if gInfo == nil {
		return role.ErrNotMyUser
	}

	if gInfo.userID == gInfo.groupID {
		return nil
	}

	payTime := gInfo.upkeeping.StartTime
	for _, proID := range gInfo.providers {
		lin := gInfo.getLInfo(proID, true)
		if lin == nil {
			continue
		}

		if lin.lastPay == nil {
			continue
		}

		payTime = lin.lastPay.GetStart() + lin.lastPay.GetLength()

		utils.MLogger.Infof("PayTime for %s and %s at %s", userID, proID, time.Unix(payTime, 0).Format(utils.BASETIME))

		km, err := metainfo.NewKey(qid, mpb.KeyType_Challenge, userID, proID, k.localID)
		if err != nil {
			return err
		}
		prefix := km.ToString()
		es, err := k.ds.Itererate(prefix)
		if err != nil {
			continue
		}
		chalRes := new(mpb.ChalInfo)
		for _, e := range es {
			rec := new(mpb.Record)
			err := proto.Unmarshal(e.Value, rec)
			if err != nil {
				continue
			}
			keys := strings.Split(string(rec.GetKey()), metainfo.DELIMITER)
			if len(keys) != 6 {
				continue
			}
			chalTime, err := strconv.ParseInt(keys[5], 10, 0)
			if err != nil {
				continue
			}

			if chalTime < payTime {
				continue
			}

			err = proto.Unmarshal(rec.GetValue(), chalRes)
			if err != nil {
				continue
			}

			if !chalRes.Res {
				continue
			}

			utils.MLogger.Infof("Found chalresult at %s", time.Unix(chalTime, 0).Format(utils.BASETIME))

			// only need successLength for stpay
			chalNeed := &mpb.ChalInfo{
				SuccessLength: chalRes.GetSuccessLength(),
			}

			lin.chalMap.Store(chalTime, chalNeed)
		}
	}
	return nil
}

//查找本地持久化保存的U-K-P信息，并与这些节点尝试连接
func (k *Info) loadPeers(ctx context.Context) error {
	localID := k.localID
	// load keepers
	kmKID, err := metainfo.NewKey(localID, mpb.KeyType_Keepers)
	if err != nil {
		return err
	}

	if kids, err := k.ds.GetKey(ctx, kmKID.ToString(), "local"); kids != nil && err == nil {
		utils.MLogger.Info(localID, " has keepers: ", string(kids))
		for i := 0; i < len(kids)/utils.IDLength; i++ {
			tmpKid := string(kids[i*utils.IDLength : (i+1)*utils.IDLength])
			_, err := peer.IDB58Decode(tmpKid)
			if err != nil {
				continue
			}
			k.getKInfo(tmpKid, true)
		}
	}

	// load providers
	kmPID, err := metainfo.NewKey(localID, mpb.KeyType_Providers)
	if err != nil {
		return err
	}

	if pids, err := k.ds.GetKey(ctx, kmPID.ToString(), "local"); pids != nil && err == nil {
		utils.MLogger.Info(localID, " has providers: ", string(pids))
		for i := 0; i < len(pids)/utils.IDLength; i++ {
			tmpKid := string(pids[i*utils.IDLength : (i+1)*utils.IDLength])
			_, err := peer.IDB58Decode(tmpKid)
			if err != nil {
				continue
			}

			k.getPInfo(tmpKid, true)
		}
	}

	return nil
}

//loadPeersFromChain: load all keepers and providers from keeper/provider contract and save in k.keepers/k.providers
//load kp from kpMap-contract and save in KpMap
func (k *Info) loadPeersFromChain() error {
	keepers, _, err := role.GetAllKeepers(k.localID)
	if err != nil {
		return err
	}

	for _, kItem := range keepers {
		k.getKInfo(kItem.KeeperID, false)
	}

	pros, totalStorage, err := role.GetAllProviders(k.localID)
	if err != nil {
		return err
	}

	for _, pItem := range pros {
		pInfo, err := k.getPInfo(pItem.ProviderID, false)
		if err != nil {
			continue
		}

		pInfo.setOffer(true)
	}

	k.pledgeStorage = totalStorage

	role.SaveKpMap(k.localID)

	return nil
}

func (k *Info) getPosPrice() *big.Int {
	if k.pledgeStorage.Sign() > 0 {
		mm := big.NewInt(MarketingMoney)
		mm.Mul(mm, big.NewInt(utils.Token))
		mm.Quo(mm, big.NewInt(24))  // per hour
		mm.Quo(mm, k.pledgeStorage) // per MB

		// to weiDollar
		mmWei := new(big.Float).SetInt(mm)
		mmWei.Mul(mmWei, role.GetMemoPrice())
		mmWei.Int(mm)
		return mm
	}

	return pos.GetPosPrice()
}

/*====================Key Ops========================*/

func (k *Info) putKey(ctx context.Context, key string, data, sig []byte, to string, clusterID uint64, flag bool) error {
	utils.MLogger.Debugf("put %s to %s", key, to)

	k.ds.PutKey(ctx, key, data, sig, "local")
	return nil
}

func (k *Info) getKey(ctx context.Context, key, to string, clusterID uint64, flag bool) ([]byte, error) {
	utils.MLogger.Debugf("get %s from %s", key, to)

	return k.ds.GetKey(ctx, key, "local")
}

/*====================Group Ops========================*/

//clean unpaid users
func (k *Info) cleanTestUsersRegular(ctx context.Context) {
	utils.MLogger.Info("Clean Test Users start!")
	ticker := time.NewTicker(2 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tNow := time.Now()
			t1 := time.Date(tNow.Year(), tNow.Month(), tNow.Day(), 1, 0, 0, 0, tNow.Location())
			t2 := t1.Add(4 * time.Hour)
			//在一点和五点之间，清理testUsers
			if tNow.After(t1) && tNow.Before(t2) {
				utils.MLogger.Info("Begin to clean test users")
				unpaids := k.getUnpaidUsers()
				for uid, qid := range unpaids {
					k.deleteGroup(ctx, qid)
					k.users.Delete(uid)
				}
			}
		}
	}
}

func (k *Info) createGroup(uid, qid string, keepers, providers []string) (*groupInfo, error) {
	gp, ok := k.ukpGroup.Load(qid)
	if !ok {
		gInfo, err := newGroup(k.localID, uid, qid, keepers, providers)
		if err != nil {
			return nil, err
		}

		k.ms.groupNum.Inc()
		if gInfo.localKeeper == gInfo.masterKeeper {
			k.ms.masterGroupNum.Inc()
		}
		k.ukpGroup.Store(qid, gInfo)

		kmsess, err := metainfo.NewKey(uid, mpb.KeyType_Session, qid)
		if err != nil {
			return gInfo, err
		}

		sessByte, err := k.ds.GetKey(k.context, kmsess.ToString(), "local")
		if err == nil && len(sessByte) > 0 {
			sID, err := uuid.ParseBytes(sessByte)
			if err == nil {
				gInfo.sessionID = sID
				gInfo.sessionTime = time.Now().Unix() - 1500 //
			}
		}

		gInfo.loadContracts(false)

		k.loadUserBucket(uid, qid)
		k.loadUserBucketStripes(uid, qid)
		// need check chunks and query missing from providers
		k.loadUserBlock(qid)
		k.loadUserPay(uid, qid)
		k.loadUserChallenge(uid, qid)

		gInfo.status = true
		return gInfo, nil
	}
	// init userConfig
	return gp.(*groupInfo), nil
}

func (k *Info) newGroupWithFS(userID, groupID string, kpids string, flag bool) error {
	if kpids == "" && flag {
		ctx := k.context
		kmkps, err := metainfo.NewKey(groupID, mpb.KeyType_LFS, userID)
		if err != nil {
			return err
		}

		res, err := k.ds.GetKey(ctx, kmkps.ToString(), "local")
		if err != nil {
			return err
		}
		kpids = string(res)
	}

	splitedMeta := strings.Split(kpids, metainfo.DELIMITER)
	var tmpKps []string
	var tmpPros []string
	if len(splitedMeta) == 2 {
		kps := splitedMeta[0]
		for i := 0; i < len(kps)/utils.IDLength; i++ {
			kid := string(kps[i*utils.IDLength : (i+1)*utils.IDLength])
			_, err := peer.IDB58Decode(kid)
			if err != nil {
				continue
			}
			tmpKps = append(tmpKps, kid)
		}

		kps = splitedMeta[1]
		for i := 0; i < len(kps)/utils.IDLength; i++ {
			kid := string(kps[i*utils.IDLength : (i+1)*utils.IDLength])
			_, err := peer.IDB58Decode(kid)
			if err != nil {
				continue
			}
			tmpPros = append(tmpPros, kid)
		}
	}

	if len(tmpKps) == 0 {
		tmpKps = append(tmpKps, groupID)
	}

	if len(tmpPros) == 0 {
		tmpPros = append(tmpPros, groupID)
	}

	_, err := k.createGroup(userID, groupID, tmpKps, tmpPros)
	return err
}

func (k *Info) deleteGroup(ctx context.Context, qid string) {
	thisGroup := k.getGroupInfo(qid, qid, false)
	if thisGroup == nil {
		return
	}

	err := thisGroup.loadContracts(true)
	if err == nil {
		return
	}

	utils.MLogger.Info(qid, " is a test userID, clean its data")
	for _, proID := range thisGroup.providers {
		thisLinfo := thisGroup.getLInfo(proID, false)
		if thisLinfo == nil {
			continue
		}

		thisLinfo.blockMap.Range(func(key, value interface{}) bool {
			blockID := qid + metainfo.BlockDelimiter + key.(string)
			utils.MLogger.Info("Delete testUser block: ", blockID)
			//先通知Provider删除块
			km, err := metainfo.NewKey(blockID, mpb.KeyType_Block)
			if err != nil {
				return false
			}
			err = k.ds.DeleteBlock(ctx, km.ToString(), proID)
			if err != nil {
				utils.MLogger.Info("Delete testUser block: ", blockID, " error:", err)
			}

			kmBlock, err := metainfo.NewKey(blockID, mpb.KeyType_BlockPos)
			if err != nil {
				return false
			}

			//delete from local
			err = k.ds.DeleteKey(ctx, kmBlock.ToString(), "local")
			if err != nil {
				utils.MLogger.Info("Delete local key error:", err)
			}

			return true
		})
	}

	// delete group
	k.ms.groupNum.Dec()
	k.ukpGroup.Delete(qid)
}

/*====================Block Meta Ops=========================*/

func (k *Info) getBlockPos(qid, bid string) (string, error) {
	gp := k.getGroupInfo(qid, qid, false)
	if gp == nil {
		return "", role.ErrNoBlock
	}

	return gp.getBlockPos(bid)
}

func (k *Info) getBlockAvail(qid, bid string) (int64, error) {
	gp := k.getGroupInfo(qid, qid, false)
	if gp == nil {
		return 0, role.ErrNoBlock
	}

	return gp.getBlockAvail(bid)
}

func (k *Info) addBucket(qid, bid string, binfo *mpb.BucketInfo) error {
	if binfo.GetDeletion() {
		utils.MLogger.Info("add bucket: ", bid, " for query: ", qid, " is deleted")
		return nil
	}
	utils.MLogger.Info("add bucket: ", bid, " for query: ", qid)

	gp := k.getGroupInfo(qid, qid, false)
	if gp != nil {
		return gp.addBucket(bid, binfo)
	}

	return role.ErrNotMyUser
}

func (k *Info) addBlockMeta(qid, bid, pid string, offset int, mode bool) error {
	utils.MLogger.Info("add block: ", bid, " and its offset: ", offset, " for query: ", qid, " and provider: ", pid)

	gp := k.getGroupInfo(qid, qid, false)
	if gp != nil {
		if mode {
			blockID := qid + metainfo.BlockDelimiter + bid

			km, err := metainfo.NewKey(blockID, mpb.KeyType_BlockPos)
			if err == nil {
				pidAndOffset := pid + metainfo.DELIMITER + strconv.Itoa(offset)

				err = k.putKey(k.context, km.ToString(), []byte(pidAndOffset), nil, "local", gp.clusterID, gp.bft)
				if err != nil {
					utils.MLogger.Info("Add block: ", blockID, " error:", err)
				}
			}
		}

		bucketID, _, _, err := metainfo.GetIDsFromBlock(bid)
		if err != nil {
			return err
		}

		bucketNum, err := strconv.Atoi(bucketID)
		if err != nil {
			return err
		}

		if bucketNum <= 0 || gp.userID == pos.GetPosId() {
			return gp.addBlockMeta(bid, pid, offset)
		}

		binfo := gp.getBucketInfo(bucketID, false)
		if binfo == nil {
			bk, err := metainfo.NewKey(qid, mpb.KeyType_Bucket, gp.userID, bucketID)
			if err != nil {
				return err
			}

			res, err := k.ds.GetKey(k.context, bk.ToString(), "local")
			if err != nil {
				for _, kid := range gp.keepers {
					if kid != gp.localKeeper {
						remoteRes, err := k.ds.GetKey(k.context, bk.ToString(), kid)
						if err == nil && len(remoteRes) > 0 {
							res = remoteRes
							break
						}
					}
				}
			}

			if len(res) > 0 {
				binfo := new(mpb.BucketInfo)
				err = proto.Unmarshal(res, binfo)
				if err != nil {
					utils.MLogger.Infof("%s Unmarshal bucketinfo: %s fail: ", qid, bucketID, err)
					return err
				}
				gp.addBucket(bucketID, binfo)
			} else {
				utils.MLogger.Infof("%s has no bucketinfo: %s", qid, bucketID)
			}
		}

		return gp.addBlockMeta(bid, pid, offset)
	}

	return role.ErrNotMyUser
}

// flag: weather noyify provider to actual delete
func (k *Info) deleteBlockMeta(qid, bid string, flag bool) {
	utils.MLogger.Info("delete block: ", bid, " for query: ", qid)

	ctx := k.context

	gp := k.getGroupInfo(qid, qid, false)
	if gp != nil {
		pid, err := gp.getBlockPos(bid)
		if err != nil || pid == "" {
			return
		}
		// delete from mem
		gp.deleteBlockMeta(bid, pid)
	}

	if flag {
		blockID := qid + metainfo.BlockDelimiter + bid

		// notify provider, to delete block
		km, err := metainfo.NewKey(blockID, mpb.KeyType_BlockPos)
		if err != nil {
			return
		}
		err = k.ds.DeleteKey(ctx, km.ToString(), "local")
		if err != nil {
			utils.MLogger.Warn("Delete block: ", blockID, " error:", err)
		}
	}

	return
}
