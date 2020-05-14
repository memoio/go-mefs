package user

import (
	"context"
	"math/big"
	"sync"

	"github.com/libp2p/go-libp2p-core/routing"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/source/data"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	"golang.org/x/sync/semaphore"
)

const defaultWeighted = int64(100)

//Info implements user service
type Info struct {
	localID string
	ds      data.Service
	fsMap   sync.Map // now key is queryID, value is *lfsInfo
	qMap    sync.Map // key is userID, value is *userInfo
	context context.Context
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
func New(ctx context.Context, nid string, d data.Service, rt routing.Routing) (instance.Service, error) {
	us := &Info{
		localID: nid,
		ds:      d,
		context: ctx,
	}
	err := rt.(*dht.KadDHT).AssignmetahandlerV2(us)
	if err != nil {
		return nil, err
	}

	return us, nil
}

// NewFS add a new user
func (u *Info) NewFS(userID, shareTo, queryID, sk string, capacity, duration int64, price *big.Int, ks, ps int, rdo, force bool) (FileSyetem, error) {
	utils.MLogger.Infof("create lfs service: %s for user %s", queryID, userID)
	if !rdo {
		// check stats
		fs := u.GetUser(userID)
		if fs != nil {
			return fs, nil
		}
	}

	ginfo := newGroup(userID, shareTo, sk, capacity, duration, price, ks, ps, u.ds)

	ginfo.force = force
	ginfo.reDeploy = rdo

	// queryID == userID indicats this is a testuser
	if queryID != userID {
		if sk == "" {
			return nil, role.ErrEmptyPrivateKey
		}

		if queryID == "" || rdo {
			qid, err := role.DeployQuery(userID, sk, capacity, duration, price, ks, ps, rdo)
			if err != nil {
				return nil, err
			}
			queryID = qid
		}

		qItem, err := role.GetQueryInfo(userID, queryID)
		if err != nil {
			utils.MLogger.Infof("get query %s for user %s from chain failed: %s, please restart", queryID, userID, err)
			return nil, err
		}

		ginfo.queryItem = &qItem
		ginfo.keeperSLA = int(qItem.KeeperNums)
		ginfo.providerSLA = int(qItem.ProviderNums)
		ginfo.groupID = queryID
		uItem, err := role.GetUpKeeping(userID, queryID)
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

	ctx, cancel := context.WithCancel(u.context)

	lInfo := &LfsInfo{
		userID:     userID,
		fsID:       queryID,
		context:    ctx,
		cancelFunc: cancel,
		privateKey: sk,
		gInfo:      ginfo,
		ds:         u.ds,
		Sm:         semaphore.NewWeighted(defaultWeighted),
	}

	u.fsMap.Store(queryID, lInfo)

	return lInfo, nil
}

// GetRole gets role
func (u *Info) GetRole() string {
	return metainfo.RoleUser
}

// Stop stops service after persist
func (u *Info) Close() error {
	u.fsMap.Range(func(key, value interface{}) bool {
		uInfo := value.(*LfsInfo)
		if !uInfo.online {
			return true
		}
		err := uInfo.Fsync(true)
		if err != nil {
			utils.MLogger.Warnf("Sorry, something wrong in persisting for %s: %s", uInfo.userID, err)
		} else {
			utils.MLogger.Infof("User %s Persist completed", uInfo.userID)
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
			err := fs.(*LfsInfo).Stop()
			if err != nil {
				return err
			}
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
	utils.MLogger.Debug("get all users")
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
