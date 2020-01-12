package user

import (
	"context"
	"errors"
	"sync"

	"github.com/libp2p/go-libp2p-core/routing"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/source/data"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

//Info implements user service
type Info struct {
	localID string
	ds      data.Service
	fsMap   sync.Map // now key is queryID, value is *lfsInfo
	qMap    sync.Map // key is userID, value is *userInfo
}

type queryInfo struct {
	sync.RWMutex
	querys map[string]struct{} // now, only one
}

func (q *queryInfo) setQuery(qid string) {
	q.Lock()
	defer q.Unlock()
	if q.querys == nil {
		q.querys = make(map[string]struct{})
	}
	_, ok := q.querys[qid]
	if !ok {
		q.querys[qid] = struct{}{}
	}
}

func (q *queryInfo) getQuery() string {
	q.RLock()
	defer q.RUnlock()
	for qid := range q.querys {
		return qid
	}
	return ""
}

// New constructs a new user service
func New(nid string, d data.Service, rt routing.Routing) (instance.Service, error) {
	us := &Info{
		localID: nid,
		ds:      d,
	}
	err := rt.(*dht.KadDHT).AssignmetahandlerV2(us)
	if err != nil {
		return nil, err
	}

	return us, nil
}

// NewFS add a new user
func (u *Info) NewFS(queryID, userID, sk string, capacity, duration, price int64, ks, ps int, rdo bool) (FileSyetem, error) {
	// check stats
	if queryID != "" {
		fs, ok := u.fsMap.Load(queryID)
		if ok {
			return fs.(*LfsInfo), errors.New("user exists")
		}
	}

	ginfo := newGroup(userID, sk, capacity, duration, price, ks, ps, rdo, u.ds)

	// queryID == userID indicats this is a testuser
	if queryID != userID {
		qItem, err := role.GetQueryInfo(userID, queryID)
		if err == nil {
			ginfo.queryItem = &qItem
		} else {
			qid, err := role.DeployQuery(userID, sk, capacity, duration, price, ks, ps, rdo)
			if err != nil {
				return nil, err
			}
			qItem, err := role.GetQueryInfo(userID, queryID)
			if err != nil {
				return nil, errors.New("fail to get query from chain, please restart")
			}
			queryID = qid
			ginfo.queryItem = &qItem
		}
		ginfo.groupID = queryID
		uItem, err := role.GetUpKeeping(ginfo.userID, ginfo.groupID)
		if err == nil {
			ginfo.upKeepingItem = &uItem
		}
	} else {
		ginfo.groupID = userID
	}

	qInfo, ok := u.qMap.Load(userID)
	if ok {
		qInfo.(*queryInfo).setQuery(queryID)
	} else {
		uq := &queryInfo{
			querys: make(map[string]struct{}),
		}
		uq.setQuery(queryID)
		u.qMap.Store(userID, uq)
	}

	ctx, cancel := context.WithCancel(context.Background())

	lInfo := &LfsInfo{
		userID:     userID,
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

// GetRole gets role
func (u *Info) GetRole() string {
	return metainfo.RoleUser
}

// Stop stops service after persist
func (u *Info) Stop() error {
	u.fsMap.Range(func(key, value interface{}) bool {
		uInfo := value.(*LfsInfo)
		if !uInfo.online {
			return true
		}
		err := uInfo.Fsync(true)
		if err != nil {
			utils.MLogger.Warnf("Sorry, something wrong in persisting for %s: %s", uInfo.userID, err)
		} else {
			utils.MLogger.Infof("User %s Persist completed\n", uInfo.userID)
		}
		uInfo.cancelFunc() //释放资源
		return true
	})
	return nil
}

func (u *Info) KillUser(userID string) error {
	uinfo, ok := u.qMap.Load(userID)
	if ok {
		queryID := uinfo.(*queryInfo).getQuery()
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
		queryID := uinfo.(*queryInfo).getQuery()
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
			utils.MLogger.Infof("Sorry, something wrong in persisting for %s: %v", uInfo.userID, err)
		} else {
			utils.MLogger.Infof("User %s Persist completed", uInfo.userID)
		}
		return true
	})
	return nil
}
