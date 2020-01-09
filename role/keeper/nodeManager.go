package keeper

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
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

// store provider information
type pInfo struct {
	sync.RWMutex
	providerID string
	maxSpace   uint64 //Bytes from contract
	usedSpace  uint64 //Bytes
	credit     int
	online     bool
	availTime  int64
	offerItem  *role.OfferItem // "latest"
	proItem    *role.ProviderItem
}

func (p *pInfo) setOffer() {
	p.Lock()
	defer p.Unlock()

}

func (k *Info) getUInfo(pid string) (*uInfo, error) {
	thisInfoI, ok := k.users.Load(pid)
	if !ok {
		tempInfo := &uInfo{
			userID: pid,
		}
		k.users.Store(pid, tempInfo)
		return tempInfo, nil
	}

	return thisInfoI.(*uInfo), nil
}

func (k *Info) getKInfo(pid string) (*kInfo, error) {
	if k.localID == pid {
		return nil, errors.New("is local keeper")
	}

	thisInfoI, ok := k.keepers.Load(pid)
	if !ok {
		tempInfo := &kInfo{
			keeperID: pid,
		}
		k.keepers.Store(pid, tempInfo)
		return tempInfo, nil
	}

	return thisInfoI.(*kInfo), nil
}

func (k *Info) getPInfo(pid string) (*pInfo, error) {
	thisInfoI, ok := k.providers.Load(pid)
	if !ok {
		tempInfo := &pInfo{
			providerID: pid,
		}
		k.providers.Store(pid, tempInfo)
		return tempInfo, nil
	}

	return thisInfoI.(*pInfo), nil
}

func (k *Info) checkPeers(ctx context.Context) {
	utils.MLogger.Info("Check connected peer start!")
	// sleep 1 minutes and then check
	time.Sleep(time.Minute)
	k.checkConnectedPeer(ctx)
	ticker := time.NewTicker(CONPEERTIME)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			k.checkLocalPeers(ctx)
			k.checkConnectedPeer(ctx)
		}
	}
}

// check connectness
func (k *Info) checkLocalPeers(ctx context.Context) {
	tmpKeepers, _ := k.GetKeepers()

	ntime := utils.GetUnixNow()
	for _, kid := range tmpKeepers {
		thisInfoI, ok := k.keepers.Load(kid)
		if !ok {
			continue
		}

		thisInfo := thisInfoI.(*kInfo)

		if k.ds.Connect(ctx, kid) {
			thisInfo.online = true
			thisInfo.availTime = ntime
			continue
		}

		if ntime-thisInfo.availTime > EXPIRETIME {
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

		if k.ds.Connect(ctx, pid) {
			thisInfo.online = true
			thisInfo.availTime = ntime
			continue
		}

		if ntime-thisInfo.availTime > EXPIRETIME {
			thisInfo.online = false
		}
	}
}

func (k *Info) checkConnectedPeer(ctx context.Context) error {
	connPeers := k.ds.GetPeers() //the list of peers we are connected to
	for _, pid := range connPeers {
		id := pid.Pretty() //连接结点id的base58编码

		_, exist := k.users.Load(id)
		if exist {
			continue
		}

		thisInfoI, exist := k.keepers.Load(id)
		if exist {
			thisInfoI.(*kInfo).online = true
			thisInfoI.(*kInfo).availTime = utils.GetUnixNow()
			continue
		}

		thisInfoP, exist := k.providers.Load(id)
		if exist {
			thisInfoP.(*pInfo).online = true
			thisInfoP.(*pInfo).availTime = utils.GetUnixNow()
			continue
		}

		utils.MLogger.Info("try to get new: ", id, " roleinfo from net and chain")
		kmRole, err := metainfo.NewKeyMeta(id, metainfo.Role)
		if err != nil {
			return err
		}
		val, _ := k.ds.GetKey(ctx, kmRole.ToString(), id)
		if string(val) == metainfo.RoleKeeper {
			addr, err := address.GetAddressFromID(id)
			if err != nil {
				return err
			}
			isKeeper, err := contracts.IsKeeper(addr)
			if err != nil {
				return err
			}
			if isKeeper {
				utils.MLogger.Info("Connect to new keeper: ", id)
				thiskInfo, err := k.getKInfo(id)
				if err != nil {
					continue
				}
				thiskInfo.online = true
				thiskInfo.availTime = utils.GetUnixNow()
			}
		} else if string(val) == metainfo.RoleProvider {
			utils.MLogger.Info("Connect to new provider: ", id)
			thispInfo, err := k.getPInfo(id)
			if err != nil {
				continue
			}

			thispInfo.online = true
			thispInfo.availTime = utils.GetUnixNow()
		}
	}
	return nil
}

// GetUsers is
func (k *Info) GetUsers() ([]string, error) {
	if !k.state {
		return nil, errKeeperServiceNotReady
	}
	var res []string
	k.ukpGroup.Range(func(uid, v interface{}) bool {
		thisuid, ok := uid.(string)
		if !ok {
			return false
		}
		thisGroupsInfo, ok := v.(*groupInfo)
		if !ok {
			return false
		}

		temp := ansi.Color(thisuid+".keepers:", "green")
		for i, keeperID := range thisGroupsInfo.keepers {
			if i != 0 {
				temp += "_"
			}
			temp += keeperID
		}
		res = append(res, temp)
		temp = ansi.Color(thisuid+".providers:", "green")
		temp += strings.Join(thisGroupsInfo.providers, "_")
		res = append(res, temp)
		return true
	})
	return res, nil
}

// GetProviders is
func (k *Info) GetProviders() ([]string, error) {
	if !k.state {
		return nil, errKeeperServiceNotReady
	}

	var res []string
	k.providers.Range(func(k, v interface{}) bool {
		res = append(res, k.(string))
		return true
	})

	return res, nil
}

// GetKeepers is
func (k *Info) GetKeepers() ([]string, error) {
	if !k.state {
		return nil, errKeeperServiceNotReady
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
