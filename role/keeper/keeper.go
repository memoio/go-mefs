package keeper

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/source/data"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	recpb "github.com/memoio/go-mefs/source/go-libp2p-kad-dht/pb"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
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
	netID     string
	role      string
	sk        string
	state     bool
	enableBft bool
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
		netID: nid,
		sk:    sk,
		state: false,
		ds:    d,
		repch: make(chan string, 1024),
	}

	err := m.load(ctx) //连接节点
	if err != nil {
		log.Println("searchAllKeepersAndProviders err:", err)
		return nil, err
	}

	//tendermint启动相关
	m.enableBft = false
	if !m.enableBft {
		log.Println("Use simple mode")
	}

	err = rt.(*dht.KadDHT).AssignmetahandlerV2(m)
	if err != nil {
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
	log.Println("Keeper Service is ready")
	return m, nil
}

func (k *Info) Online() bool {
	return k.state
}

func (k *Info) GetRole() string {
	return metainfo.RoleKeeper
}

func (k *Info) Stop() error {
	return k.save(context.Background())
}

/*====================Save and Load========================*/

func (k *Info) persistRegular(ctx context.Context) {
	log.Println("Persist LocalPeerInfo start!")
	ticker := time.NewTicker(PERSISTTIME)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := k.save(ctx)
			if err != nil {
				log.Println("PersistlocalPeerInfo err:", err)
			}
		}
	}
}

func (k *Info) save(ctx context.Context) error {
	localID := k.netID

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
		err = k.ds.PutKey(ctx, kmKID.ToString(), []byte(pids.String()), "local")
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
		err = k.ds.PutKey(ctx, kmPID.ToString(), []byte(pids.String()), "local")
		if err != nil {
			return err
		}
	}

	// persist users
	pids.Reset()

	kmUID, err := metainfo.NewKeyMeta(localID, metainfo.Users)
	if err != nil {
		return err
	}

	var res strings.Builder
	k.users.Range(func(key, value interface{}) bool {
		uid := key.(string)
		pids.WriteString(uid)
		kmfs, err := metainfo.NewKeyMeta(uid, metainfo.LogFS)
		if err != nil {
			return true
		}

		for _, qid := range value.(*uInfo).querys {
			res.WriteString(qid)
		}

		if res.Len() > 0 {
			err = k.ds.PutKey(ctx, kmfs.ToString(), []byte(res.String()), "local")
			if err != nil {
				return true
			}
		}

		res.Reset()
		return true
	})

	if pids.Len() > 0 {
		err = k.ds.PutKey(ctx, kmUID.ToString(), []byte(pids.String()), "local")
		if err != nil {
			return err
		}
	}

	// persist ManagedUsers: query id
	pids.Reset()

	kmQID, err := metainfo.NewKeyMeta(localID, metainfo.Users)
	if err != nil {
		return err
	}

	k.ukpGroup.Range(func(qid, groupsinfo interface{}) bool {
		pids.WriteString(qid.(string))
		return true
	})

	if pids.Len() > 0 {
		err = k.ds.PutKey(ctx, kmQID.ToString(), []byte(pids.String()), "local")
		if err != nil {
			return err
		}
	}

	// save last pay
	pus := k.getUQKeys()
	for _, pu := range pus {
		gp := k.getGroupInfo(pu.qid, pu.uid, false)
		if gp == nil {
			continue
		}

		for _, proID := range gp.providers {
			k.savePay(pu.qid, proID)
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
			log.Println("doSpaceTimePay()NewKeyMeta()err: ", err)
			return err
		}
		valueLast := strings.Join([]string{utils.UnixToString(beginTime), utils.UnixToString(endTime), spaceTime.String(), "signature", "proof"}, metainfo.DELIMITER)
		k.ds.PutKey(ctx, kmLast.ToString(), []byte(valueLast), "local")

		//key: `qid/"chalpay"/pid/beginTime/endTime`
		//value: `spacetime/signature/proof`
		//for storing
		km, err := metainfo.NewKeyMeta(qid, metainfo.ChalPay, pid, utils.UnixToString(beginTime), utils.UnixToString(endTime))
		if err != nil {
			log.Println("doSpaceTimePay()NewKeyMeta()err: ", err)
			return err
		}
		metaValue := strings.Join([]string{spaceTime.String(), "signature", "proof"}, metainfo.DELIMITER)
		k.ds.PutKey(ctx, km.ToString(), []byte(metaValue), "local")
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
	log.Println("Load All userID's Information")
	localID := k.netID //本地id
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

			log.Println("Load user", userID, "'s infomations")
			wg.Add(1)
			go func(userID string) {
				defer wg.Done()
				kmfs, err := metainfo.NewKeyMeta(userID, metainfo.LogFS)
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
					err = k.createGroup(qid, userID, []string{qid}, []string{qid})
					if err != nil {
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
		proto.Unmarshal(e.Value, rec)
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

		if len(km.GetMid()) > utils.IDLength {
			if km.GetMid()[:utils.IDLength] != qid {
				continue
			}
		}

		k.addBlockMeta(qid, km.GetMid(), pids[0], off)
	}
	return nil
}

//查找本地持久化保存的U-K-P信息，并与这些节点尝试连接
func (k *Info) loadPeers(ctx context.Context) error {
	localID := k.netID
	// load keepers
	kmKID, err := metainfo.NewKeyMeta(localID, metainfo.Keepers)
	if err != nil {

		return err
	}

	if kids, err := k.ds.GetKey(ctx, kmKID.ToString(), "local"); kids != nil && err == nil {
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

			err = saveOfferToPinfo(tmpKid, thisPinfo)
			if err != nil {
				log.Println("Save ", tmpKid, "'s Offer error: ", err)
			}
		}
	}

	return nil
}

/*====================Group Ops========================*/

//clean unpaid users
func (k *Info) cleanTestUsersRegular(ctx context.Context) {
	log.Println("Clean Test Users start!")
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
				log.Println("Begin to clean test users")
				unpaids := k.getUnpaidUsers()
				for uid, qid := range unpaids {
					k.deleteGroup(ctx, qid)
					k.users.Delete(uid)
				}
			}
		}
	}
}

func (k *Info) createGroup(qid, uid string, keepers, providers []string) error {
	_, ok := k.ukpGroup.Load(qid)
	if !ok {
		gInfo, err := newGroup(k.netID, qid, uid, keepers, providers)
		if err != nil {
			return err
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
				log.Println(err)
				return err
			}

			kms := kmLast.ToString()
			// get from leveldb
			res, err := k.ds.GetKey(ctx, kms, "local")
			if err != nil {
				log.Println("no lastTime data, return Unix(0)")
				return err
			}
			err = lin.parseLastPayKV(res)
			if err != nil {
				log.Println("checkLastPayTime() parseLastPayKV() err: ", err)
				return err
			}
		}
	}

	// init userConfig

	return nil
}

func (k *Info) deleteGroup(ctx context.Context, qid string) {
	thisGroup := k.getGroupInfo(qid, qid, false)
	if thisGroup == nil {
		return
	}

	//recheck the user's status
	addr, err := address.GetAddressFromID(thisGroup.owner)
	if err != nil {
		return
	}

	_, _, err = contracts.GetUKFromResolver(addr)
	if err != contracts.ErrNotDeployedMapper && err != contracts.ErrNotDeployedUk {
		thisGroup.saveUpkeeping()
		return
	}

	log.Println(qid, "is a test userID, clean its data")
	for _, proID := range thisGroup.providers {
		thisIlinfo, ok := thisGroup.ledgerMap.Load(proID)
		if !ok {
			continue
		}

		thisLinfo := thisIlinfo.(*lInfo)

		thisLinfo.blockMap.Range(func(key, value interface{}) bool {
			blockID := qid + metainfo.BLOCK_DELIMITER + key.(string)
			log.Println("Delete testUser block-", blockID)
			//先通知Provider删除块
			km, err := metainfo.NewKeyMeta(blockID, metainfo.Block)
			if err != nil {
				log.Println("construct delete block KV error :", err)
				return false
			}
			err = k.ds.DeleteBlock(ctx, km.ToString(), proID)
			if err != nil {
				log.Println("Delete testUser block failed-", blockID, "error:", err)
			}

			kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.BlockPos)
			if err != nil {
				log.Println("NewKeyMeta()error!", err, "blockID:", blockID)
			}

			//delete from local
			err = k.ds.DeleteKey(ctx, kmBlock.ToString(), "local")
			if err != nil {
				log.Println("Delete local key error:", err)
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
		return "", errors.New("No block")
	}

	return gp.getBlockPos(bid)
}

func (k *Info) addBlockMeta(qid, bid, pid string, offset int) error {
	log.Println("add block: ", bid, "for query: ", qid, " and provider: ", pid)

	blockID := qid + metainfo.BLOCK_DELIMITER + bid

	kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.BlockPos)
	if err != nil {
		log.Println("NewKeyMeta()error!", err, "blockID:", blockID)
	}

	ctx := context.Background()
	//put to local
	value := pid + metainfo.DELIMITER + strconv.Itoa(offset)
	err = k.ds.PutKey(ctx, kmBlock.ToString(), []byte(value), "local")
	if err != nil {
		log.Println("Delete local key error:", err)
	}

	gp := k.getGroupInfo(qid, qid, false)
	if gp != nil {
		return gp.addBlockMeta(bid, pid, offset)
	}

	return errors.New("Not my user")
}

// flag: weather noyify provider to actual delete
func (k *Info) deleteBlockMeta(qid, bid string, flag bool) {
	blockID := qid + metainfo.BLOCK_DELIMITER + bid

	kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.BlockPos)
	if err != nil {
		log.Println("NewKeyMeta()error!", err, "blockID:", blockID)
	}

	ctx := context.Background()
	//delete from local
	err = k.ds.DeleteKey(ctx, kmBlock.ToString(), "local")
	if err != nil {
		log.Println("Delete local key error:", err)
	}

	var pid string
	gp := k.getGroupInfo(qid, qid, false)
	if gp != nil {
		pid, err = gp.getBlockPos(bid)
		if err != nil || pid == "" {
			// need to get from local kv
			res, err := k.ds.GetKey(ctx, kmBlock.ToString(), "local")
			if err != nil {
				return
			}
			po := strings.Split(string(res), metainfo.DELIMITER)
			pid = po[0]
			return
		}
		// delete from mem
		gp.deleteBlockMeta(pid, bid)
	}

	if flag {
		// notify provider, to delete block
		km, err := metainfo.NewKeyMeta(blockID, metainfo.Block)
		if err != nil {
			log.Println("construct delete block KV error :", err)
			return
		}
		err = k.ds.DeleteBlock(ctx, km.ToString(), pid)
		if err != nil {
			log.Println("Delete testUser block failed-", blockID, "error:", err)
		}
	}

	return
}
