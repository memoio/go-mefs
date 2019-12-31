package user

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/libp2p/go-libp2p-core/routing"
	"github.com/memoio/go-mefs/source/data"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/source/instance"
	ad "github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
)

//Info implements user service
type Info struct {
	netID string
	role  string
	ds    data.Service
	fsMap sync.Map // now key is queryID, value is *lfsInfo
	qMap  sync.Map // key is userID, value is *userInfo
}

type userInfo struct {
	querys []string // now, only one
}

// New constructs a new user service
func New(nid string, d data.Service, rt routing.Routing) (instance.Service, error) {
	us := &Info{
		role:  metainfo.RoleUser,
		netID: nid,
		ds:    d,
	}
	err := rt.(*dht.KadDHT).AssignmetahandlerV2(us)
	if err != nil {
		return nil, err
	}
	return us, nil
}

// NewFS add a new user
func (u *Info) NewFS(uid, queryID, sk string, capacity, duration, price int64, ks, ps int, rdo bool) (FileSyetem, error) {
	uinfo, ok := u.qMap.Load(uid)
	if ok {
		queryID := uinfo.(*userInfo).querys[0]
		fs, ok := u.fsMap.Load(queryID)
		if ok {
			return fs.(*LfsInfo), errors.New("user exists")
		}
	}

	// getUK
	uaddr, err := ad.GetAddressFromID(uid)
	if err != nil {
		return nil, err
	}

	ginfo := newGroup(uid, sk, capacity, duration, price, ks, ps, rdo, u.ds)

	// queryID == uid indicats this is a testuser
	if queryID != uid {
		qItem := getQuery(uaddr)
		if qItem != nil {
			ginfo.queryItem = qItem
		} else {
			err := deployQuery(uid, sk, capacity, duration, price, ks, ps, rdo)
			if err != nil {
				return nil, err
			}
			qItem = getQuery(uaddr)
			if qItem == nil {
				return nil, errors.New("fail to get query from chain, please restart")
			}
			ginfo.queryItem = qItem
		}
		qid, _ := ad.GetIDFromAddress(qItem.QueryAddr)
		queryID = qid
		ginfo.groupID = queryID
	} else {
		ginfo.groupID = uid
	}

	_, ok = u.qMap.Load(uid)
	if !ok {
		uq := &userInfo{
			querys: make([]string, 1),
		}
		uq.querys = append(uq.querys, queryID)
		u.qMap.Store(uid, uq)
	}

	uItem := getUpKeeping(uaddr)
	if uItem != nil {
		ginfo.upKeepingItem = uItem
	}

	ctx, cancel := context.WithCancel(context.Background())

	lInfo := &LfsInfo{
		owner:      uid,
		fsID:       queryID,
		context:    ctx,
		cancelFunc: cancel,
		privateKey: []byte(sk),
		gInfo:      ginfo,
		ds:         u.ds,
	}

	u.fsMap.Store(queryID, lInfo)

	return lInfo, nil
}

func (u *Info) Stop() error {
	u.fsMap.Range(func(key, value interface{}) bool {
		uInfo := value.(*LfsInfo)
		if !uInfo.online {
			return true
		}
		err := uInfo.Fsync(true)
		if err != nil {
			log.Printf("Sorry, something wrong in persisting for %s: %v\n", uInfo.owner, err)
		} else {
			log.Printf("User %s Persist completed\n", uInfo.owner)
		}
		uInfo.cancelFunc() //释放资源
		return true
	})
	return nil
}

func (u *Info) KillUser(userID string) error {
	uinfo, ok := u.qMap.Load(userID)
	if ok {
		queryID := uinfo.(*userInfo).querys[0]
		fs, ok := u.fsMap.Load(queryID)
		if ok {
			fs.(*LfsInfo).Stop()
			u.fsMap.Delete(queryID)
		}
	}
	return nil
}

// GetUser gets userInfo
func (u *Info) GetUser(userID string) FileSyetem {
	uinfo, ok := u.qMap.Load(userID)
	if ok {
		queryID := uinfo.(*userInfo).querys[0]
		fs, ok := u.fsMap.Load(queryID)
		if ok {
			return fs.(*LfsInfo)
		}
	}
	return nil
}

// GetAllUser gets userInfo
func (u *Info) GetAllUser() []string {
	res := make([]string, 0)
	u.qMap.Range(func(k, v interface{}) bool {
		res = append(res, k.(string))
		return true
	})
	return res
}

func (u *Info) Fsync() error {
	u.fsMap.Range(func(key, value interface{}) bool {
		uInfo := value.(*LfsInfo)
		if !uInfo.online {
			return true
		}
		err := uInfo.Fsync(true)
		if err != nil {
			log.Printf("Sorry, something wrong in persisting for %s: %v\n", uInfo.owner, err)
		} else {
			log.Printf("User %s Persist completed\n", uInfo.owner)
		}
		return true
	})
	return nil
}
