package keeper

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	"github.com/lni/dragonboat/v3"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/repo/fsrepo"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/source/data"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	recpb "github.com/memoio/go-mefs/source/go-libp2p-kad-dht/pb"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/source/raft"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
)

//Info implements user service
type Info struct {
	localID    string
	role       string
	sk         string
	state      bool
	enableBft  bool
	dnh        *dragonboat.NodeHost
	raftNodeID uint64
	repch      chan string
	ds         data.Service
	keepers    sync.Map // keepers except self; value: *kInfo
	providers  sync.Map // value: *pInfo
	users      sync.Map // value: *uInfo
	ukpGroup   sync.Map // manage user-keeper-provider group, value: *group
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

	rootpath, _ := fsrepo.BestKnownPath()
	m.dnh = raft.StartHost(rootpath)
	if m.dnh != nil {
		m.enableBft = true
		utils.MLogger.Info("Use bft mode")
	} else {
		m.enableBft = false
		utils.MLogger.Info("Use simple mode")
	}

	err = m.load(ctx) //连接节点
	if err != nil {
		utils.MLogger.Error("load err:", err)
		return nil, err
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
	k.save(context.Background())
	if k.dnh != nil {
		k.dnh.Stop()
	}
	return nil
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
		kmLast, err := metainfo.NewKey(qid, mpb.KeyType_LastPay, pid)
		if err != nil {
			return err
		}
		valueLast := strings.Join([]string{utils.UnixToString(beginTime), utils.UnixToString(endTime), spaceTime.String(), "signature", "proof"}, metainfo.DELIMITER)

		clusterID, err := address.GetNodeIDFromID(qid)
		if err != nil {
			return err
		}

		k.putKey(ctx, kmLast.ToString(), []byte(valueLast), nil, "local", clusterID, true)

		//key: `qid/"chalpay"/pid/beginTime/endTime`
		//value: `spacetime/signature/proof`
		//for storing
		km, err := metainfo.NewKey(qid, mpb.KeyType_ChalPay, pid, utils.UnixToString(beginTime), utils.UnixToString(endTime))
		if err != nil {
			return err
		}
		metaValue := strings.Join([]string{spaceTime.String(), "signature", "proof"}, metainfo.DELIMITER)

		k.putKey(ctx, km.ToString(), []byte(metaValue), nil, "local", clusterID, true)
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
	kmUID, err := metainfo.NewKey(localID, mpb.KeyType_Users)
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
				kmfs, err := metainfo.NewKey(userID, mpb.KeyType_Query)
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
	prefix := qid + metainfo.BlockDelimiter
	es, _ := k.ds.Itererate(prefix)
	for _, e := range es {
		rec := new(recpb.Record)
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

		getID := strings.SplitN(km.GetMid(), metainfo.BlockDelimiter, 2)
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
			thisKinfo, err := k.getKInfo(tmpKid)
			if err != nil {
				continue
			}
			if k.ds.Connect(ctx, tmpKid) {
				thisKinfo.availTime = time.Now().Unix()
				thisKinfo.online = true
			}
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

			thisPinfo, err := k.getPInfo(tmpKid)
			if err != nil {
				continue
			}

			if k.ds.Connect(ctx, tmpKid) {
				thisPinfo.availTime = time.Now().Unix()
				thisPinfo.online = true
			}
		}
	}

	return nil
}

/*====================Key Ops========================*/

func (k *Info) putKey(ctx context.Context, key string, data, sig []byte, to string, clusterID uint64, flag bool) error {
	utils.MLogger.Debugf("put %s to %s", key, to)

	k.ds.PutKey(ctx, key, data, sig, "local")

	if k.enableBft && flag {
		rec := &recpb.Record{
			Key:       []byte(key),
			Value:     data,
			Signature: sig,
		}
		recByte, err := proto.Marshal(rec)
		if err != nil {
			utils.MLogger.Error("proto Marshal fails: ", err)
			return err
		}

		// need retry?
		raft.Write(ctx, k.dnh, clusterID, key, recByte)
	}

	return nil
}

func (k *Info) getKey(ctx context.Context, key, to string, clusterID uint64, flag bool) ([]byte, error) {
	utils.MLogger.Debugf("get %s from %s", key, to)

	if k.enableBft && flag {
		res, err := raft.Read(ctx, k.dnh, clusterID, key)
		if err != nil {
			return nil, err
		}

		rec := new(recpb.Record)
		err = proto.Unmarshal(res, rec)
		if err != nil {
			return nil, err
		}

		val, err := k.ds.GetKey(ctx, key, "local")
		if err != nil {
			utils.MLogger.Debugf("get %s fails %s", key, err)
		} else {
			if bytes.Compare(val, rec.GetValue()) == 0 {
				utils.MLogger.Debugf("get %s success", key)
			} else {
				utils.MLogger.Debugf("get %s success, value is not equal", key)
			}
		}

		return rec.GetValue(), nil
	}

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

		if k.enableBft {
			initialMembers := make(map[uint64]string)
			err = raft.StartCluster(k.dnh, gInfo.clusterID, gInfo.nodeID, false, initialMembers)
			if err == nil {
				gInfo.bft = true
				utils.MLogger.Infof("try start cluster %d for %s, success", gInfo.clusterID, gInfo.groupID)
			} else {
				utils.MLogger.Errorf("start cluster %d for %s, fails %s", gInfo.clusterID, gInfo.groupID, err)
				if err == dragonboat.ErrClusterAlreadyExist {
					gInfo.bft = true
				} else {
					for _, kid := range gInfo.keepers {
						if kid == k.localID {
							initialMembers[gInfo.nodeID] = raft.Addr
							continue
						}

						t, err := address.GetNodeIDFromID(kid)
						if err != nil {
							continue
						}

						ipAddr, err := k.ds.GetExternalAddr(kid)
						if err != nil {
							continue
						}

						ips := strings.Split(ipAddr.String(), "/")
						if len(ips) != 5 {
							utils.MLogger.Errorf("ip %s is wrong", ipAddr.String())
							continue
						}

						initialMembers[t] = ips[2] + ":3001"
					}

					if len(initialMembers) == len(gInfo.keepers) {
						err = raft.StartCluster(k.dnh, gInfo.clusterID, gInfo.nodeID, false, initialMembers)
						if err != nil {
							utils.MLogger.Errorf("start cluster %d for %s, fails %s", gInfo.clusterID, gInfo.groupID, err)
							if err != dragonboat.ErrClusterAlreadyExist {
								gInfo.bft = false
							}
						}
						gInfo.bft = true
						utils.MLogger.Infof("start cluster %d for %s, success, has members: %s", gInfo.clusterID, gInfo.groupID, initialMembers)
					}
				}
			}
		}

		k.ukpGroup.Store(qid, gInfo)
		ctx := context.Background()
		for _, pid := range gInfo.providers {
			lin := &lInfo{
				inChallenge:  false,
				lastChalTime: time.Now().Unix(),
			}

			gInfo.ledgerMap.Store(pid, lin)

			kmLast, err := metainfo.NewKey(qid, mpb.KeyType_LastPay, pid)
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
		thisIlinfo, ok := thisGroup.ledgerMap.Load(proID)
		if !ok {
			continue
		}

		thisLinfo := thisIlinfo.(*lInfo)

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

func (k *Info) addBlockMeta(qid, bid, pid string, offset int, mode bool) error {
	utils.MLogger.Info("add block: ", bid, " and its offset: ", offset, " for query: ", qid, " and provider: ", pid)

	gp := k.getGroupInfo(qid, qid, false)
	if gp != nil {
		if mode {
			blockID := qid + metainfo.BlockDelimiter + bid

			km, err := metainfo.NewKey(blockID, mpb.KeyType_BlockPos)
			if err != nil {
				return err
			}

			pidAndOffset := pid + metainfo.DELIMITER + strconv.Itoa(offset)

			err = k.putKey(context.Background(), km.ToString(), []byte(pidAndOffset), nil, "local", gp.clusterID, gp.bft)
			if err != nil {
				utils.MLogger.Info("Add block: ", blockID, " error:", err)
			}
		}

		return gp.addBlockMeta(bid, pid, offset)
	}

	return role.ErrNotMyUser
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
