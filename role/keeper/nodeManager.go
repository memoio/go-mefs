package keeper

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/mgutz/ansi"
)

// store keeper information
type kInfo struct {
	keeperID      string
	online        bool
	lastAvailTime int64
}

// store provider information
type pInfo struct {
	providerID    string
	maxSpace      uint64 //Bytes
	usedSpace     uint64
	credit        int
	online        bool
	lastAvailTime int64
	offerItem     *contracts.OfferItem
}

// store user information
type uInfo struct {
	userID    string
	pubKey    *mcl.PublicKey
	queryItem *contracts.QueryItem
}

// local node information
type peerInfo struct {
	keepersInfo   sync.Map // keepers except self
	providersInfo sync.Map // providers
	usersInfo     sync.Map // users
	enableBft     bool
}

var localPeerInfo *peerInfo

func getUInfo(pid string) (*uInfo, error) {
	thisInfoI, ok := localPeerInfo.usersInfo.Load(pid)
	if !ok {
		tempInfo := &uInfo{
			userID: pid,
		}
		localPeerInfo.usersInfo.Store(pid, tempInfo)
		return tempInfo, nil
	}

	return thisInfoI.(*uInfo), nil
}

func getKInfo(pid string) (*kInfo, error) {
	if localNode.Identity.Pretty() == pid {
		return nil, errors.New("is local keeper")
	}

	thisInfoI, ok := localPeerInfo.keepersInfo.Load(pid)
	if !ok {
		tempInfo := &kInfo{
			keeperID: pid,
		}
		localPeerInfo.keepersInfo.Store(pid, tempInfo)
		return tempInfo, nil
	}

	return thisInfoI.(*kInfo), nil
}

func getPInfo(pid string) (*pInfo, error) {
	thisInfoI, ok := localPeerInfo.providersInfo.Load(pid)
	if !ok {
		tempInfo := &pInfo{
			providerID: pid,
		}
		localPeerInfo.providersInfo.Store(pid, tempInfo)
		return tempInfo, nil
	}

	return thisInfoI.(*pInfo), nil
}

func addCredit(provider string) {
	thisInfo, err := getPInfo(provider)
	if err != nil {
		return
	}
	thisInfo.credit += 100
}

func setCredit(provider string, val int) {
	thisInfo, err := getPInfo(provider)
	if err != nil {
		return
	}
	thisInfo.credit = val
}

func reduceCredit(provider string) {
	thisInfo, err := getPInfo(provider)
	if err != nil {
		return
	}
	thisInfo.credit -= 100
}

// check connectness
func checkLocalPeers(ctx context.Context) {
	var tmpKeepers []string
	localPeerInfo.keepersInfo.Range(func(key, value interface{}) bool {
		tmpKeepers = append(tmpKeepers, key.(string))
		return true
	})
	for _, keeper := range tmpKeepers {
		thisInfo, err := getKInfo(keeper)
		if err != nil || thisInfo == nil {
			continue
		}

		if localNode.Data.Connect(ctx, keeper) {
			thisInfo.online = true
			thisInfo.lastAvailTime = utils.GetUnixNow()
		} else {
			thisInfo.online = false
		}
	}

	tmpKeepers = tmpKeepers[:0]
	localPeerInfo.providersInfo.Range(func(key, value interface{}) bool {
		tmpKeepers = append(tmpKeepers, key.(string))
		return true
	})

	for _, keeper := range tmpKeepers {
		thisInfo, err := getPInfo(keeper)
		if err != nil || thisInfo == nil {
			continue
		}

		if localNode.Data.Connect(ctx, keeper) {
			thisInfo.online = true
			thisInfo.lastAvailTime = utils.GetUnixNow()
		} else {
			thisInfo.online = false
		}
	}
}

func checkConnectedPeer(ctx context.Context) error {
	checkLocalPeers(ctx)

	connPeers := localNode.PeerHost.Network().Peers() //the list of peers we are connected to

	for _, ID := range connPeers {
		id := ID.Pretty() //连接结点id的base58编码

		thisInfoI, exist := localPeerInfo.keepersInfo.Load(id)

		if exist {
			thisInfoI.(*kInfo).online = true
			thisInfoI.(*kInfo).lastAvailTime = utils.GetUnixNow()
			continue
		}

		thisInfoP, exist := localPeerInfo.providersInfo.Load(id)
		if exist {
			thisInfoP.(*pInfo).online = true
			thisInfoP.(*pInfo).lastAvailTime = utils.GetUnixNow()
			continue
		}

		log.Println("try to get new: ", id, " roleinfo from net and chain")
		kmRole, err := metainfo.NewKeyMeta(id, metainfo.Role)
		if err != nil {
			return err
		}
		val, _ := localNode.Data.GetKey(context.Background(), kmRole.ToString(), id)
		if string(val) == instance.RoleKeeper {
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
				thisInfoI, err := getKInfo(id)
				if err != nil {
					continue
				}
				thisInfoI.online = true
				thisInfoI.lastAvailTime = utils.GetUnixNow()
			}
		} else if string(val) == instance.RoleProvider {
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
				thisInfoP, err := getPInfo(id)
				if err != nil {
					continue
				}
				thisInfoP.online = true
				thisInfoP.lastAvailTime = utils.GetUnixNow()
				saveOffer(id, false)
			}
		}
	}
	return nil
}

func checkPeers(ctx context.Context) {
	log.Println("Check connected peer start!")
	// sleep 1 minutes and then check
	time.Sleep(time.Minute)
	checkConnectedPeer(ctx)
	ticker := time.NewTicker(CONPEERTIME)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			checkConnectedPeer(ctx)
		}
	}
}

// GetUsers is
func GetUsers() ([]string, error) {
	if !isKeeperServiceRunning() {
		return nil, errKeeperServiceNotReady
	}
	var res []string
	ukpInfo.Range(func(uid, v interface{}) bool {
		thisuid, ok := uid.(string)
		if !ok {
			return false
		}
		thisGroupsInfo, ok := v.(*groupsInfo)
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
func GetProviders() ([]string, error) {
	if !isKeeperServiceRunning() {
		return nil, errKeeperServiceNotReady
	}

	var res []string
	localPeerInfo.providersInfo.Range(func(k, v interface{}) bool {
		res = append(res, k.(string))
		return true
	})

	return res, nil
}

// GetKeepers is
func GetKeepers() ([]string, error) {
	if !isKeeperServiceRunning() {
		return nil, errKeeperServiceNotReady
	}
	var res []string
	localPeerInfo.keepersInfo.Range(func(k, v interface{}) bool {
		res = append(res, k.(string))
		return true
	})

	return res, nil
}

// FlushKeepersAndProviders is
func FlushKeepersAndProviders() error {
	if !isKeeperServiceRunning() {
		return errKeeperServiceNotReady
	}
	return checkConnectedPeer(context.Background())
}

// GetUsersInfomation gets user's informatin
func GetUsersInfomation(userid string) {
	gp, ok := getGroupsInfo(userid)
	if !ok {
		return
	}
	for _, proID := range gp.providers {
		thisPU := puKey{
			uid: userid,
			pid: proID,
		}
		thisinfo, ok := ledgerInfo.Load(thisPU)
		if !ok {
			continue
		}

		thisChal := thisinfo.(*chalinfo)
		fmt.Println("last challenge time is: ", utils.UnixToTime(thisChal.lastChalTime))
		if thisChal.lastPay != nil {
			fmt.Println(fmt.Println("last pay time is: ", utils.UnixToTime(thisChal.lastPay.endTime)))
		}
		thisChal.cidMap.Range(func(key, value interface{}) bool {
			cinfo := value.(*cidInfo)
			fmt.Println("cid is: ", key.(string))
			fmt.Println("availtime is: ", utils.UnixToTime(cinfo.availtime))
			fmt.Println("cid length is: ", cinfo.offset)
			return true
		})
	}
	return
}
