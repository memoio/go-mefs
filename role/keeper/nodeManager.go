package keeper

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/mgutz/ansi"
)

// store keeper information
type kInfo struct {
	keeperID  string
	online    bool
	availTime int64
}

// store provider information
type pInfo struct {
	providerID string
	maxSpace   uint64 //Bytes from contract
	usedSpace  uint64 //Bytes
	credit     int
	online     bool
	availTime  int64
	offerItem  *contracts.OfferItem // latest?
}

// store user information
type uInfo struct {
	userID string
	querys []string
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

func (k *Info) setQuery(uid, qid string) {
	uinfo, err := k.getUInfo(uid)
	if err != nil {
		return
	}

	has := false
	for _, q := range uinfo.querys {
		if q == qid {
			has = true
		}
	}

	if !has {
		uinfo.querys = append(uinfo.querys, qid)
	}
}

func (k *Info) getKInfo(pid string) (*kInfo, error) {
	if k.netID == pid {
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

func (k *Info) addCredit(provider string) {
	thisInfo, err := k.getPInfo(provider)
	if err != nil {
		return
	}
	thisInfo.credit += 100
}

func (k *Info) setCredit(provider string, val int) {
	thisInfo, err := k.getPInfo(provider)
	if err != nil {
		return
	}
	thisInfo.credit = val
}

func (k *Info) reduceCredit(provider string) {
	thisInfo, err := k.getPInfo(provider)
	if err != nil {
		return
	}
	thisInfo.credit -= 100
}

func (k *Info) checkPeers(ctx context.Context) {
	log.Println("Check connected peer start!")
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
		}

		if ntime-thisInfo.availTime > EXPIRETIME {
			thisInfo.online = false
		}
	}

	tmpKeepers, _ = k.GetProviders()
	for _, pid := range tmpKeepers {
		thisInfoI, ok := k.providers.Load(pid)
		if !ok {
			continue
		}

		thisInfo := thisInfoI.(*pInfo)

		if k.ds.Connect(ctx, pid) {
			thisInfo.online = true
			thisInfo.availTime = ntime
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

		log.Println("try to get new: ", id, " roleinfo from net and chain")
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
				log.Println("Connect to new keeper: ", id)
				thiskInfo, err := k.getKInfo(id)
				if err != nil {
					continue
				}
				thiskInfo.online = true
				thiskInfo.availTime = utils.GetUnixNow()
			}
		} else if string(val) == metainfo.RoleProvider {
			addr, err := address.GetAddressFromID(id)
			if err != nil {
				return err
			}
			isProvider, err := contracts.IsProvider(addr)
			if err != nil {
				return err
			}
			if isProvider {
				log.Println("Connect to new provider: ", id)
				thispInfo, err := k.getPInfo(id)
				if err != nil {
					continue
				}
				thispInfo.online = true
				thispInfo.availTime = utils.GetUnixNow()
				k.saveOffer(id, false)
			}
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

// Flush is
func (k *Info) FlushPeers(ctx context.Context) error {
	return k.checkConnectedPeer(ctx)
}
