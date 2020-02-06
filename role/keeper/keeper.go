package keeper

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	"github.com/lni/dragonboat/v3"
	"github.com/memoio/go-mefs/repo/fsrepo"
	"github.com/memoio/go-mefs/source/data"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	recpb "github.com/memoio/go-mefs/source/go-libp2p-kad-dht/pb"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/source/raft"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

const (
	EXPIRETIME       = int64(30 * 60) //超过这个时间，触发修复，单位：秒
	CHALTIME         = 5 * time.Minute
	CHECKTIME        = 7 * time.Minute
	PERSISTTIME      = 3 * time.Minute
	SPACETIMEPAYTIME = 61 * time.Minute
	CONPEERTIME      = 5 * time.Minute
	KPMAPTIME        = 11 * time.Minute
)

//Info implements user service
type Info struct {
	localID   string
	role      string
	sk        string
	state     bool
	enableBft bool
	dnh       *dragonboat.NodeHost
	repch     chan string
	ds        data.Service
	keepers   sync.Map // keepers except self; value: *kInfo
	providers sync.Map // value: *pInfo
	users     sync.Map // value: *uInfo
	ukpGroup  sync.Map // manage user-keeper-provider group, value: *group
}

// New is
// TODO:Keeper出问题重启后，应该能自动将所有user的信息恢复到内存中
func New(ctx context.Context, nid, sk string, d data.Service, rt routing.Routing) (instance.Service, error) {
	m := &Info{
		localID: nid,
		sk:      sk,
		state:   false,
		ds:      d,
		repch:   make(chan string, 1024),
	}

	err := rt.(*dht.KadDHT).AssignmetahandlerV2(m)
	if err != nil {
		return nil, err
	}

	err = m.load(ctx) //连接节点
	if err != nil {
		utils.MLogger.Error("load err:", err)
		return nil, err
	}

	rootpath, _ := fsrepo.BestKnownPath()
	m.dnh = raft.StartRaftHost(rootpath)

	//tendermint启动相关
	m.enableBft = false
	if !m.enableBft {
		utils.MLogger.Info("Use simple mode")
	}

	go m.persistRegular(ctx)
	go m.challengeRegular(ctx)
	go m.cleanTestUsersRegular(ctx)
	go m.checkLedger(ctx)
	go m.repairRegular(ctx)
	go m.stPayRegular(ctx)
	go m.checkPeers(ctx)
	go m.getKpMapRegular(ctx)
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

// Stop is
func (k *Info) Stop() error {
	return k.save(context.Background())
}

/*====================Save and Load========================*/

func (k *Info) persistRegular(ctx context.Context) {
	utils.MLogger.Info("Persist local peerInfo start!")
	ticker := time.NewTicker(PERSISTTIME)
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

func (k *Info) save(ctx context.Context) error {
	localID := k.localID

	// persist keepers
	kmKID, err := metainfo.NewKeyMeta(localID, metainfo.Keepers)
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
	kmPID, err := metainfo.NewKeyMeta(localID, metainfo.Providers)
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

	kmUID, err := metainfo.NewKeyMeta(localID, metainfo.Users)
	if err != nil {
		return err
	}

	var res strings.Builder
	k.users.Range(func(key, value interface{}) bool {
		uid := key.(string)
		pids.WriteString(uid)
		kmfs, err := metainfo.NewKeyMeta(uid, metainfo.Query)
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
	for _, qu := range qus {
		gp := k.getGroupInfo(qu.uid, qu.qid, false)
		if gp == nil {
			continue
		}

		for _, proID := range gp.providers {
			k.savePay(qu.qid, proID)
		}
	}

	return nil
}

func (k *Info) savePay(qid, pid string) error {
	thisLinfo := k.getLInfo(qid, qid, pid, false)

	if thisLinfo != nil && thisLinfo.lastPay != nil {
		beginTime := thisLinfo.lastPay.beginTime
		endTime := thisLinfo.lastPay.endTime
		spaceTime := thisLinfo.lastPay.spacetime
		ctx := context.Background()

		//key: qid/`lastpay"/pid`
		//value: `beginTime/endTime/spacetime/signature/proof`
		kmLast, err := metainfo.NewKeyMeta(qid, metainfo.LastPay, pid)
		if err != nil {
			return err
		}
		valueLast := strings.Join([]string{utils.UnixToString(beginTime), utils.UnixToString(endTime), spaceTime.String(), "signature", "proof"}, metainfo.DELIMITER)
		k.ds.PutKey(ctx, kmLast.ToString(), []byte(valueLast), nil, "local")

		//key: `qid/"chalpay"/pid/beginTime/endTime`
		//value: `spacetime/signature/proof`
		//for storing
		km, err := metainfo.NewKeyMeta(qid, metainfo.ChalPay, pid, utils.UnixToString(beginTime), utils.UnixToString(endTime))
		if err != nil {
			return err
		}
		metaValue := strings.Join([]string{spaceTime.String(), "signature", "proof"}, metainfo.DELIMITER)
		k.ds.PutKey(ctx, km.ToString(), []byte(metaValue), nil, "local")
	}
	return nil
}

func (k *Info) load(ctx context.Context) error {
	k.loadPeers(ctx)
	k.loadUser(ctx)
	return nil
}

//重启后重新恢复User现场 读取本地存储的U-K-P信息，构建PInfo结构
func (k *Info) loadUser(ctx context.Context) error {
	utils.MLogger.Info("Load All userID's Information")
	localID := k.localID //本地id
	kmUID, err := metainfo.NewKeyMeta(localID, metainfo.Users)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	if users, err := k.ds.GetKey(ctx, kmUID.ToString(), "local"); users != nil && err == nil {
		for i := 0; i < len(users)/utils.IDLength; i++ { //对user进行循环，逐个恢复user信息
			userID := string(users[i*utils.IDLength : (i+1)*utils.IDLength])
			_, err := peer.IDB58Decode(userID)
			if err != nil {
				continue
			}

			utils.MLogger.Info("Load user: ", userID, " 's infomations")
			wg.Add(1)
			go func(userID string) {
				defer wg.Done()
				kmfs, err := metainfo.NewKeyMeta(userID, metainfo.Query)
				if err != nil {
					return
				}

				ui := &uInfo{
					userID: userID,
				}

				k.users.Store(userID, ui)

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

					k.loadUserBlock(qid)
				}
			}(userID)
		}
	}

	wg.Wait()

	return nil
}

func (k *Info) loadUserBlock(qid string) error {
	prefix := qid + metainfo.BLOCK_DELIMITER
	es, _ := k.ds.Itererate(prefix)
	for _, e := range es {
		rec := new(recpb.Record)
		err := proto.Unmarshal(e.Value, rec)
		if err != nil {
			continue
		}

		utils.MLogger.Debug("Load block: ", string(rec.GetKey()))

		km, err := metainfo.GetKeyMeta(string(rec.GetKey()))
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

		getID := strings.SplitN(km.GetMid(), metainfo.BLOCK_DELIMITER, 2)
		if len(getID) != 2 || (len(getID) > 0 && getID[0] != qid) {
			continue
		}

		k.addBlockMeta(qid, getID[1], pids[0], off, false)
	}
	return nil
}

//查找本地持久化保存的U-K-P信息，并与这些节点尝试连接
func (k *Info) loadPeers(ctx context.Context) error {
	localID := k.localID
	// load keepers
	kmKID, err := metainfo.NewKeyMeta(localID, metainfo.Keepers)
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
			thisKinfo, err := k.getKInfo(tmpKid)
			if err != nil {
				continue
			}
			if k.ds.Connect(ctx, tmpKid) {
				thisKinfo.availTime = utils.GetUnixNow()
				thisKinfo.online = true
			}
		}
	}

	// load providers
	kmPID, err := metainfo.NewKeyMeta(localID, metainfo.Providers)
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

			thisPinfo, err := k.getPInfo(tmpKid)
			if err != nil {
				continue
			}

			if k.ds.Connect(ctx, tmpKid) {
				thisPinfo.availTime = utils.GetUnixNow()
				thisPinfo.online = true
			}
		}
	}

	return nil
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
		k.ukpGroup.Store(qid, gInfo)
		ctx := context.Background()
		for _, pid := range gInfo.providers {
			lin := &lInfo{
				inChallenge:  false,
				lastChalTime: utils.GetUnixNow(),
			}

			gInfo.ledgerMap.Store(pid, lin)

			kmLast, err := metainfo.NewKeyMeta(qid, metainfo.LastPay, pid)
			if err != nil {
				continue
			}

			res, err := k.ds.GetKey(ctx, kmLast.ToString(), "local")
			if err == nil && len(res) > 0 {
				err = lin.parseLastPayKV(res)
				if err != nil {
					utils.MLogger.Error("parseLastPayKV err: ", err)
				}
			}
		}
		go gInfo.loadContracts(true)
		go k.loadUserBlock(qid)
		return gInfo, nil
	}
	// init userConfig
	return gp.(*groupInfo), nil
}

func (k *Info) newGroupWithFS(userID, groupID string, kpids string, flag bool) error {
	if kpids == "" && flag {
		ctx := context.Background()
		kmkps, err := metainfo.NewKeyMeta(groupID, metainfo.LogFS, userID)
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
		thisIlinfo, ok := thisGroup.ledgerMap.Load(proID)
		if !ok {
			continue
		}

		thisLinfo := thisIlinfo.(*lInfo)

		thisLinfo.blockMap.Range(func(key, value interface{}) bool {
			blockID := qid + metainfo.BLOCK_DELIMITER + key.(string)
			utils.MLogger.Info("Delete testUser block: ", blockID)
			//先通知Provider删除块
			km, err := metainfo.NewKeyMeta(blockID, metainfo.Block)
			if err != nil {
				return false
			}
			err = k.ds.DeleteBlock(ctx, km.ToString(), proID)
			if err != nil {
				utils.MLogger.Info("Delete testUser block: ", blockID, " error:", err)
			}

			kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.BlockPos)
			if err != nil {
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
	k.ukpGroup.Delete(qid)
}

/*====================Block Meta Ops=========================*/

func (k *Info) getBlockPos(qid, bid string) (string, error) {
	gp := k.getGroupInfo(qid, qid, false)
	if gp == nil {
		return "", errBlockNotExist
	}

	return gp.getBlockPos(bid)
}

func (k *Info) getBlockAvail(qid, bid string) (int64, error) {
	gp := k.getGroupInfo(qid, qid, false)
	if gp == nil {
		return 0, errBlockNotExist
	}

	return gp.getBlockAvail(bid)
}

func (k *Info) addBlockMeta(qid, bid, pid string, offset int, mode bool) error {
	utils.MLogger.Info("add block: ", bid, " and its offset: ", offset, " for query: ", qid, " and provider: ", pid)

	if mode {
		blockID := qid + metainfo.BLOCK_DELIMITER + bid

		// notify provider, to delete block
		km, err := metainfo.NewKeyMeta(blockID, metainfo.BlockPos)
		if err != nil {
			return err
		}

		pidAndOffset := pid + metainfo.DELIMITER + strconv.Itoa(offset)

		err = k.ds.PutKey(context.Background(), km.ToString(), []byte(pidAndOffset), nil, "local")
		if err != nil {
			utils.MLogger.Info("Delete testUser block: ", blockID, " error:", err)
		}
	}

	gp := k.getGroupInfo(qid, qid, false)
	if gp != nil {
		return gp.addBlockMeta(bid, pid, offset)
	}

	return errors.New("Not my user")
}

// flag: weather noyify provider to actual delete
func (k *Info) deleteBlockMeta(qid, bid string, flag bool) {
	utils.MLogger.Info("delete block: ", bid, "for query: ", qid)

	ctx := context.Background()

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
		blockID := qid + metainfo.BLOCK_DELIMITER + bid

		// notify provider, to delete block
		km, err := metainfo.NewKeyMeta(blockID, metainfo.BlockPos)
		if err != nil {
			return
		}
		err = k.ds.DeleteKey(ctx, km.ToString(), "local")
		if err != nil {
			utils.MLogger.Warn("Delete testUser block: ", blockID, "  error:", err)
		}
	}

	return
}
