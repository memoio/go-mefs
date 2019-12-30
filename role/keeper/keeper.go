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
	"github.com/memoio/go-mefs/contracts"
	pb "github.com/memoio/go-mefs/role/pb"
	"github.com/memoio/go-mefs/source/data"
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
	netID      string
	role       string
	sk         string
	state      bool
	enableBft  bool
	repch      chan string
	ds         data.Service
	keepers    sync.Map // keepers except self
	providers  sync.Map // providers
	users      sync.Map // users
	ukpManager *ukp
	lManager   *ledger // key: PU，value: *chalinfo
}

// New is
// TODO:Keeper出问题重启后，应该能自动将所有user的信息恢复到内存中
func New(ctx context.Context, nid, sk string, d data.Service) instance.Service {
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
		return nil
	}

	//tendermint启动相关
	m.enableBft = false
	if !m.enableBft {
		log.Println("Use simple mode")
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
	return m
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

func (k *Info) GetLedger() *ledger {
	return k.lManager
}

func (k *Info) GetUKP() *ukp {
	return k.ukpManager
}

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

	k.ukpManager.gMap.Range(func(qid, groupsinfo interface{}) bool {
		pids.WriteString(qid.(string))
		return true
	})

	if pids.Len() > 0 {
		err = k.ds.PutKey(ctx, kmQID.ToString(), []byte(pids.String()), "local")
		if err != nil {
			return err
		}
	}

	// persist ledgerInfos
	kmLedger, err := metainfo.NewKeyMeta(localID, metainfo.LedgerMap)
	if err != nil {
		return err
	}

	tmpLedgerinfo := make(map[string]*pb.Chalin)

	pus := k.ukpManager.getPUKeys()
	for _, pu := range pus {
		thisInfo, ok := k.lManager.lMap.Load(pu)
		if !ok {
			continue
		}
		thischalinfo := thisInfo.(*chalinfo)
		tmpCid := make(map[string]*pb.Cidin)
		puProto := &pb.Pu{
			Provider: pu.pid,
			User:     pu.qid,
		}
		puByte, err := proto.Marshal(puProto) //*格式修改
		if err != nil {
			log.Println("proto.Marshal error:", err)
		}
		thischalinfo.cidMap.Range(func(k, v interface{}) bool {
			tmpCidin := &pb.Cidin{
				Res:      v.(*cidInfo).res,
				Repair:   v.(*cidInfo).repair,
				Offset:   int64(v.(*cidInfo).offset),
				Avaltime: utils.UnixToString(v.(*cidInfo).availtime),
			}
			tmpCid[k.(string)] = tmpCidin
			return true
		})
		chalinProto := &pb.Chalin{
			Cidin:     tmpCid,
			Maxlength: thischalinfo.maxlength,
		}
		tmpLedgerinfo[string(puByte)] = chalinProto
	}

	ledgerin := &pb.LedgerInfo{
		Chalinfo: tmpLedgerinfo,
	}
	ledgerByte, err := proto.Marshal(ledgerin) //*格式修改
	if err != nil {
		log.Println("proto.Marshal error:", err)
	}

	err = k.ds.PutKey(ctx, kmLedger.ToString(), ledgerByte, "local")
	if err != nil {
		return err
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
					k.fillGroup(qid, userID, []string{qid}, []string{qid})
					k.loadUserBlock(qid)
				}
			}(userID)
		}
	}

	// load ledgerinfo
	kmLedger, err := metainfo.NewKeyMeta(localID, metainfo.LedgerMap)
	if err != nil {
		return err
	}

	if ledgers, err := k.ds.GetKey(ctx, kmLedger.ToString(), "local"); ledgers != nil && err == nil {
		ledgerinProto := &pb.LedgerInfo{}
		err = proto.Unmarshal(ledgers, ledgerinProto)
		if err != nil {
			return err
		}
		for pustr, thischalinfoinProto := range ledgerinProto.Chalinfo {
			puinProto := &pb.Pu{}
			err = proto.Unmarshal([]byte(pustr), puinProto)
			if err != nil {
				return err
			}
			newpu := pqKey{
				pid: puinProto.Provider,
				qid: puinProto.User,
			}

			for blockid, thiscidinfoinProto := range thischalinfoinProto.Cidin {
				err = k.addBlockMeta(newpu.qid, newpu.pid, blockid, int(thiscidinfoinProto.Offset))
				if err != nil {
					continue
				}
			}

			thisChal, ok := k.lManager.lMap.Load(newpu)
			if !ok {
				continue
			}

			if thischalinfoinProto.Maxlength != thisChal.(*chalinfo).maxlength {
				log.Println(newpu.qid, "stores on pid: ", newpu.pid, " calculate length and stored length is: ", thisChal.(*chalinfo).maxlength, thischalinfoinProto.Maxlength)
			}
		}
	}
	wg.Wait()

	return nil
}

func (k *Info) fillGroup(qid, uid string, keepers, providers []string) (*groupsInfo, error) {
	tempInfo := &groupsInfo{
		keepers:      keepers,
		providers:    providers,
		groupID:      qid,
		owner:        uid,
		localKeeper:  k.netID,
		masterKeeper: qid,
	}

	if qid != uid {
		err := saveUpkeepingToGP(qid, tempInfo)
		if err != nil {
			return nil, err
		}
	} else {
		flag := false
		for _, keeperID := range tempInfo.keepers {
			if k.netID == keeperID {
				flag = true
			}
		}

		// not my user
		if !flag {
			log.Println(uid, "is not my user")
			return nil, errors.New("Not my user")
		}
	}

	return tempInfo, nil
}

func (k *Info) addBlockMeta(qid, pid, bid string, offset int) error {

	newCidInfo, err := k.lManager.addBlockMeta(qid, pid, bid, offset)
	if err != nil {
		return err
	}
	return k.ukpManager.addBlockMeta(qid, bid, newCidInfo)
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

		k.addBlockMeta(qid, pids[0], km.GetMid(), off)
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
				k.cleanTestUsers(ctx)
			}
		}
	}
}

func (k *Info) cleanTestUsers(ctx context.Context) {
	unpaids := k.ukpManager.getUnpaidUsers()
	for _, userID := range unpaids {
		thisKPInfo, ok := k.ukpManager.gMap.Load(userID)
		if !ok {
			continue
		}

		thiskp := thisKPInfo.(*groupsInfo)

		//recheck the user's status
		addr, err := address.GetAddressFromID(userID)
		if err != nil {
			continue
		}

		_, _, err = contracts.GetUKFromResolver(addr)
		if err != contracts.ErrNotDeployedMapper && err != contracts.ErrNotDeployedUk {
			saveUpkeepingToGP(userID, thiskp)
			continue
		}

		for _, proID := range thiskp.providers {
			pu := pqKey{
				pid: proID,
				qid: userID,
			}

			thisInfo, ok := k.lManager.lMap.Load(pu)
			if !ok {
				continue
			}

			thischalinfo := thisInfo.(*chalinfo)

			log.Println(pu.qid, "is a test userID, clean its data")
			thischalinfo.cidMap.Range(func(key, value interface{}) bool {
				blockID := key.(string)
				log.Println("Delete testUser block-", blockID)
				//先通知Provider删除块
				km, err := metainfo.NewKeyMeta(blockID, metainfo.Block)
				if err != nil {
					log.Println("construct delete block KV error :", err)
					return false
				}
				err = k.ds.DeleteBlock(ctx, km.ToString(), pu.pid)
				if err != nil {
					log.Println("Delete testUser block failed-", blockID, "error:", err)
				}

				//再在本地删除记录
				kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.BlockPos)
				if err != nil {
					log.Println("NewKeyMeta()error!", err, "blockID:", blockID)
				}
				err = k.ds.DeleteKey(ctx, kmBlock.ToString(), "local")
				if err != nil {
					log.Println("Delete local Message error:", err)
				}
				return true
			})

			k.ukpManager.gMap.Delete(pu.qid)
			k.lManager.lMap.Delete(pu)
		}
	}
}

func (k *Info) getBlockPos(qid, blockID string) (string, error) {
	gri, ok := k.ukpManager.gMap.Load(qid)
	if !ok {
		return "", errors.New("No block")
	}

	bis := strings.SplitN(blockID, metainfo.BLOCK_DELIMITER, 2)

	bui, ok := gri.(*groupsInfo).buckets.Load(bis[0])
	if !ok {
		return "", errors.New("No block")
	}

	sti, ok := bui.(*bucketInfo).stripes.Load(bis[1])
	if !ok {
		return "", errors.New("No block")
	}

	st := sti.(*cidInfo)
	return st.storedOn, nil
}

func (k *Info) deleteBlockMeta(qid, blockID string) {
	pid, err := k.getBlockPos(qid, blockID)
	if err != nil {
		return
	}
	k.lManager.deleteBlockMeta(qid, pid, blockID)
	k.ukpManager.deleteBlockMeta(pid, blockID)
}
