package keeper

import (
	"context"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/memoio/go-mefs/contracts"
	pb "github.com/memoio/go-mefs/role/pb"
	ds "github.com/memoio/go-mefs/source/go-datastore"
	dsq "github.com/memoio/go-mefs/source/go-datastore/query"
	recpb "github.com/memoio/go-mefs/source/go-libp2p-kad-dht/pb"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func persistLocalPeerInfoRegular(ctx context.Context) {
	log.Println("Persist LocalPeerInfo start!")
	ticker := time.NewTicker(PERSISTTIME)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := PersistlocalPeerInfo()
			if err != nil {
				log.Println("PersistlocalPeerInfo err:", err)
			}
		}
	}
}

// PersistlocalPeerInfo is
func PersistlocalPeerInfo() error {
	if !isKeeperServiceRunning() {
		return errKeeperServiceNotReady
	}
	localID := localNode.Identity.Pretty() //本地id

	// persist keepers
	kmKID, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypeKid)
	if err != nil {
		return err
	}

	var pids strings.Builder
	localPeerInfo.keepersInfo.Range(func(key, value interface{}) bool {
		pids.WriteString(key.(string))
		return true
	})

	if pids.Len() > 0 {
		err = localNode.Data.PutKey(context.Backgroud(), kmKID.ToString(), []byte(pids.String()), "local")
		if err != nil {
			return err
		}

	}

	// persist providers
	pids.Reset()
	kmPID, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypePid)
	if err != nil {
		return err
	}

	localPeerInfo.providersInfo.Range(func(key, value interface{}) bool {
		pids.WriteString(key.(string))
		return true
	})

	if pids.Len() > 0 {
		err = localNode.Data.PutKey(context.Backgroud(), kmPID.ToString(), []byte(pids.String()), "local")
		if err != nil {
			return err
		}
	}

	// persist users
	pids.Reset()
	kmUID, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypeUID)
	if err != nil {
		return err
	}

	ukpInfo.Range(func(uid, groupsinfo interface{}) bool {
		pids.WriteString(uid.(string))
		return true
	})

	if pids.Len() > 0 {
		err = localNode.Data.PutKey(context.Backgroud(), kmUID.ToString(), []byte(pids.String()), "local")
		if err != nil {
			return err
		}
	}

	// persist ledgerInfos
	kmLedger, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypeLedger)
	if err != nil {
		return err
	}

	tmpLedgerinfo := make(map[string]*pb.Chalin)

	pus := getPUKeysFromukpInfo()
	for _, pu := range pus {
		thisInfo, ok := ledgerInfo.Load(pu)
		if !ok {
			continue
		}
		thischalinfo := thisInfo.(*chalinfo)
		tmpCid := make(map[string]*pb.Cidin)
		puProto := &pb.Pu{
			Provider: pu.pid,
			User:     pu.uid,
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

	err = localNode.Data.PutKey(context.Backgroud(), kmLedger.ToString(), ledgerByte, "local")
	if err != nil {
		return err
	}

	return nil
}

//重启后重新恢复User现场 读取本地存储的U-K-P信息，构建PInfo结构
func loadAllUser() error {
	log.Println("Load All userID's Information")
	localID := localNode.Identity.Pretty() //本地id

	kmUID, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypeUID)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	if users, err := localNode.Data.GetKey(kmUID.ToString(), "local"); users != nil && err == nil {
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
				fillUserInfo(userID, []string{userID}, []string{userID})
				loadUserBlock(userID)
			}(userID)
		}
	}

	// load ledgerinfo
	kmLedger, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypeLedger)
	if err != nil {
		return err
	}

	if ledgers, err := localNode.Data.GetKey(kmLedger.ToString(), "local"); ledgers != nil && err == nil {
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
			newpu := puKey{
				pid: puinProto.Provider,
				uid: puinProto.User,
			}

			newchalinfo := &chalinfo{}
			ledgerInfo.Store(newpu, newchalinfo)

			for blockid, thiscidinfoinProto := range thischalinfoinProto.Cidin {
				newcidinfo := &cidInfo{
					res:       thiscidinfoinProto.Res,
					repair:    thiscidinfoinProto.Repair,
					availtime: utils.StringToUnix(thiscidinfoinProto.Avaltime),
					offset:    int(thiscidinfoinProto.Offset),
					storedOn:  puinProto.Provider,
				}
				err = addCidinfotoMem(newpu.uid, newpu.pid, blockid, newcidinfo)
				if err != nil {
					continue
				}
			}

			if thischalinfoinProto.Maxlength != newchalinfo.maxlength {
				log.Println(newpu.uid, "stores on pid: ", newpu.pid, " calculate length and stored length is: ", newchalinfo.maxlength, thischalinfoinProto.Maxlength)
			}
		}
	}
	wg.Wait()

	return nil
}

func loadUserBlock(userID string) error {
	key := ds.NewKey(userID + metainfo.BLOCK_DELIMITER).String()
	q := dsq.Query{Prefix: key}
	qr, _ := localNode.Repo.Datastore().Query(q) //进行查询操作
	es, _ := qr.Rest()

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
			if km.GetMid()[:utils.IDLength] != userID {
				continue
			}
		}

		addBlocktoMem(userID, pids[0], km.GetMid(), off)
	}
	return nil
}

//查找本地持久化保存的U-K-P信息，并与这些节点尝试连接
func loadKnownKeepersAndProviders(ctx context.Context) error {
	if !isKeeperServiceRunning() {
		return errKeeperServiceNotReady
	}
	localID := localNode.Identity.Pretty() //本地id

	// load keepers
	kmKID, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypeKid)
	if err != nil {

		return err
	}

	if kids, err := localNode.Data.GetKey(kmKID.ToString(), "local"); kids != nil && err == nil {
		for i := 0; i < len(kids)/utils.IDLength; i++ {
			tmpKid := string(kids[i*utils.IDLength : (i+1)*utils.IDLength])
			_, err := peer.IDB58Decode(tmpKid)
			if err != nil {
				continue
			}
			thisKinfo, err := getKInfo(tmpKid)
			if err != nil {
				continue
			}
			if localNode.Data.Connect(ctx, tmpKid) {
				thisKinfo.lastAvailTime = utils.GetUnixNow()
				thisKinfo.online = true
			}
		}
	}

	// load providers
	kmPID, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypePid)
	if err != nil {
		return err
	}

	if pids, err := localNode.Data.GetKey(kmPID.ToString(), "local"); pids != nil && err == nil {
		for i := 0; i < len(pids)/utils.IDLength; i++ {
			tmpKid := string(pids[i*utils.IDLength : (i+1)*utils.IDLength])
			_, err := peer.IDB58Decode(tmpKid)
			if err != nil {
				continue
			}
			thisPinfo, err := getPInfo(tmpKid)
			if err != nil {
				continue
			}

			if sc.ConnectTo(ctx, localNode, tmpKid) {
				thisPinfo.lastAvailTime = utils.GetUnixNow()
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
func cleanTestUsersRegular(ctx context.Context) {
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
				cleanTestUsers()
			}
		}
	}
}

func cleanTestUsers() {
	unpaids := getUnpaidUsers()
	for _, userID := range unpaids {
		thisKPInfo, ok := ukpInfo.Load(userID)
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
			pu := puKey{
				pid: proID,
				uid: userID,
			}

			thisInfo, ok := ledgerInfo.Load(pu)
			if !ok {
				continue
			}

			thischalinfo := thisInfo.(*chalinfo)

			log.Println(pu.uid, "is a test userID, clean its data")
			thischalinfo.cidMap.Range(func(key, value interface{}) bool {
				blockID := key.(string)
				log.Println("Delete testUser block-", blockID)
				//先通知Provider删除块
				km, err := metainfo.NewKeyMeta(blockID, metainfo.DeleteBlock)
				if err != nil {
					log.Println("construct delete block KV error :", err)
					return false
				}
				_, err = localNode.Data.SendMetaRequest(km, "", pu.pid)
				if err != nil {
					retryCount := 3
					for i := 0; i < retryCount; i++ {
						_, err = localNode.Data.SendMetaRequest(km, "", pu.pid)
						if err == nil {
							break
						}
					}
					if err != nil {
						log.Println("Delete testUser block failed-", blockID, "error:", err)
					}
				}

				//再在本地删除记录
				kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.Local, metainfo.SyncTypeBlock)
				if err != nil {
					log.Println("NewKeyMeta()error!", err, "blockID:", blockID)
				}
				err = localNode.Data.DeleteKey(kmBlock.ToString(), "local")
				if err != nil {
					log.Println("Delete local Message error:", err)
				}
				return true
			})

			//将其从账本中删除
			kmKID, err := metainfo.NewKeyMeta(pu.uid, metainfo.Local, metainfo.SyncTypeKid)
			if err != nil {
				return
			}
			err = localNode.Data.DeleteKey(kmKID.ToString(), "local")
			if err != nil {
				log.Println("Delete local Message error:", err)
			}
			kmPID, err := metainfo.NewKeyMeta(pu.uid, metainfo.Local, metainfo.SyncTypePid)
			if err != nil {
				return
			}
			err = localNode.Data.DeleteKey(kmPID.ToString(), "local")
			if err != nil {
				log.Println("Delete local Message error:", err)
			}
			ukpInfo.Delete(pu.uid)
			ledgerInfo.Delete(pu)
		}
	}
}
