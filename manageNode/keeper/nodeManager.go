package keeper

import (
	"context"
	"strings"
	"sync"
	"time"

	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	datastore "github.com/memoio/go-mefs/source/go-datastore"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/mgutz/ansi"
)

// store user information
type uInfo struct {
	sync.RWMutex
	userID string
	querys map[string]struct{} // key is queryID
}

func (u *uInfo) setQuery(qid string) {
	u.Lock()
	defer u.Unlock()
	if u.querys == nil {
		u.querys = make(map[string]struct{})
	}
	_, ok := u.querys[qid]
	if !ok {
		u.querys[qid] = struct{}{}
	}
}

func (u *uInfo) getQuery() string {
	u.RLock()
	defer u.RUnlock()
	for id := range u.querys {
		return id
	}

	return ""
}

// store keeper information
type kInfo struct {
	// need lock
	keeperID  string
	online    bool
	availTime int64
	keepItem  *role.KeeperItem
}

// todo node queue: accodring to credit, storage used

// store provider information
type pInfo struct {
	sync.RWMutex
	providerID   string
	maxSpace     uint64 // Bytes from contract
	usedSpace    uint64 // Bytes reported by provider
	managedSpace uint64 // Bytes managed by this keeper
	credit       int    // faults
	online       bool
	availTime    int64
	offerItem    *role.OfferItem // "latest"
	proItem      *role.ProviderItem
}

func (p *pInfo) setOffer(flag bool) error {
	if p.proItem == nil || flag {
		proItem, err := role.GetProviderInfo(p.providerID, p.providerID)
		if err != nil {
			utils.MLogger.Infof("%s is not a provider: %s", p.providerID, err)
			return err
		}
		p.proItem = &proItem
	}
	if p.offerItem == nil || flag {
		oItem, err := role.GetLatestOffer(p.providerID, p.providerID)
		if err != nil {
			return err
		}
		p.Lock()
		defer p.Unlock()
		p.offerItem = &oItem
	}

	return nil
}

func (k *Info) getUInfo(pid string) (*uInfo, error) {
	thisInfoI, ok := k.users.Load(pid)
	if !ok {
		tempInfo := &uInfo{
			userID: pid,
		}
		k.ms.userNum.Inc()
		k.users.Store(pid, tempInfo)
		return tempInfo, nil
	}

	return thisInfoI.(*uInfo), nil
}

func (k *Info) getKInfo(pid string) (*kInfo, error) {
	if k.localID == pid {
		return nil, role.ErrNotMyKeeper
	}

	thisInfoI, ok := k.keepers.Load(pid)
	if !ok {
		has, err := role.IsKeeper(pid)
		if err != nil {
			return nil, err
		}

		if !has {
			return nil, role.ErrNotKeeper
		}

		tempInfo := &kInfo{
			keeperID: pid,
		}

		if k.ds.Connect(k.context, pid) {
			tempInfo.availTime = time.Now().Unix()
			tempInfo.online = true
			k.ms.keeperNum.Inc()
			k.keepers.Store(pid, tempInfo)
		}

		return tempInfo, nil
	}

	return thisInfoI.(*kInfo), nil
}

func (k *Info) getPInfo(pid string) (*pInfo, error) {
	thisInfoI, ok := k.providers.Load(pid)
	if !ok {
		has, err := role.IsProvider(pid)
		if err != nil {
			return nil, err
		}

		if !has {
			return nil, role.ErrNotProvider
		}

		tempInfo := &pInfo{
			providerID: pid,
			credit:     50,
		}

		err = tempInfo.setOffer(false)
		if err != nil {
			return nil, err
		}

		if k.ds.Connect(k.context, pid) {
			tempInfo.availTime = time.Now().Unix()
			tempInfo.online = true
			k.ms.providerNum.Inc()
			k.providers.Store(pid, tempInfo)
		}

		return tempInfo, nil
	}

	return thisInfoI.(*pInfo), nil
}

func (k *Info) checkPeers(ctx context.Context) {
	utils.MLogger.Info("Check connected peer start!")
	// sleep 1 minutes and then check
	time.Sleep(time.Minute)
	k.checkConnectedPeer(ctx)
	ticker := time.NewTicker(checkConnTime)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			k.checkLocalPeers(ctx)
			k.checkConnectedPeer(ctx)
			usedCapacity, err := datastore.DiskUsage(k.ds.DataStore())
			if err == nil {
				k.ms.storageUsed.Set(float64(usedCapacity))
			}
		}
	}
}

// check connectness
func (k *Info) checkLocalPeers(ctx context.Context) {
	tmpKeepers, _ := k.GetKeepers()

	for _, kid := range tmpKeepers {
		thisInfoI, ok := k.keepers.Load(kid)
		if !ok {
			continue
		}

		thisInfo := thisInfoI.(*kInfo)

		ntime := time.Now().Unix()
		if k.ds.Connect(ctx, kid) {
			thisInfo.online = true
			thisInfo.availTime = ntime
			continue
		}

		if ntime-thisInfo.availTime > expireTime {
			thisInfo.online = false
		}
	}

	tmpPros, _ := k.GetProviders()
	for _, pid := range tmpPros {
		thisInfoI, ok := k.providers.Load(pid)
		if !ok {
			continue
		}

		thisInfo := thisInfoI.(*pInfo)

		ntime := time.Now().Unix()
		if k.ds.Connect(ctx, pid) {
			thisInfo.online = true
			thisInfo.availTime = ntime
			continue
		}

		if ntime-thisInfo.availTime > expireTime {
			thisInfo.online = false
		}
	}
}

func (k *Info) checkConnectedPeer(ctx context.Context) error {
	//the list of peers we are connected to
	connPeers, err := k.ds.GetPeers()
	if err != nil {
		return err
	}

	for _, pid := range connPeers {
		id := pid.Pretty() //连接结点id的base58编码

		_, exist := k.netIDs[id]
		if exist {
			continue
		}

		_, exist = k.users.Load(id)
		if exist {
			continue
		}

		thisInfoI, exist := k.keepers.Load(id)
		if exist {
			thisInfoI.(*kInfo).online = true
			thisInfoI.(*kInfo).availTime = time.Now().Unix()
			continue
		}

		thisInfoP, exist := k.providers.Load(id)
		if exist {
			thisInfoP.(*pInfo).online = true
			thisInfoP.(*pInfo).availTime = time.Now().Unix()
			continue
		}

		utils.MLogger.Info("try to get new: ", id, " roleinfo from net and chain")
		kmRole, err := metainfo.NewKey(id, mpb.KeyType_Role)
		if err != nil {
			continue
		}
		val, err := k.ds.GetKey(ctx, kmRole.ToString(), id)
		if err != nil {
			continue
		}

		utils.MLogger.Infof("get %s role: %s from net", id, string(val))
		if string(val) == metainfo.RoleKeeper {
			utils.MLogger.Info("Connect to new keeper: ", id)
			k.getKInfo(id)
		} else if string(val) == metainfo.RoleProvider {
			utils.MLogger.Info("Connect to new provider: ", id)
			k.getPInfo(id)
		} else {
			k.netIDs[id] = struct{}{}
		}
	}
	return nil
}

// GetUsers is
func (k *Info) GetUsers() ([]string, error) {
	if !k.state {
		return nil, role.ErrServiceNotReady
	}
	var res []string
	k.ukpGroup.Range(func(key, v interface{}) bool {
		qid, ok := key.(string)
		if !ok {
			return true
		}
		thisGroupsInfo, ok := v.(*groupInfo)
		if !ok {
			return true
		}

		uid := thisGroupsInfo.userID

		temp := ansi.Color(uid+".fsID:"+qid+" has keepers:", "red")
		temp += strings.Join(thisGroupsInfo.keepers, "/")
		res = append(res, temp)
		temp = ansi.Color(uid+".fsID:"+qid+" has providers:", "green")
		temp += strings.Join(thisGroupsInfo.providers, "/")
		res = append(res, temp)
		return true
	})
	return res, nil
}

// GetProviders is
func (k *Info) GetProviders() ([]string, error) {
	if !k.state {
		return nil, role.ErrServiceNotReady
	}

	var res []string
	k.providers.Range(func(k, v interface{}) bool {
		if v.(*pInfo).online && v.(*pInfo).credit > 0 {
			res = append(res, k.(string))
		}
		return true
	})

	return res, nil
}

// GetKeepers is
func (k *Info) GetKeepers() ([]string, error) {
	if !k.state {
		return nil, role.ErrServiceNotReady
	}
	var res []string
	k.keepers.Range(func(k, v interface{}) bool {
		res = append(res, k.(string))
		return true
	})

	return res, nil
}

// FlushPeers is
func (k *Info) FlushPeers(ctx context.Context) error {
	return k.checkConnectedPeer(ctx)
}
