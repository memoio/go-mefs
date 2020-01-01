package keeper

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	ad "github.com/memoio/go-mefs/utils/address"
)

// force update when mode is true
func (k *Info) saveUpkeeping(gid, uid string, mode bool) error {
	gp := k.getGroupInfo(gid, uid, false)
	if gp == nil {
		log.Println("saveUpkeeping getGroupsInfo() error")
		return errNoGroupsInfo
	}

	if gp.upkeeping == nil || (mode && gid != uid) {
		return gp.saveUpkeeping()
	}

	return nil
}

func (g *groupInfo) saveUpkeeping() error {
	// get upkkeeping addr
	userAddr, err := ad.GetAddressFromID(g.owner)
	if err != nil {
		return err
	}

	queryAddr, err := ad.GetAddressFromID(g.groupID)
	if err != nil {
		return err
	}

	ukAddr, uk, err := contracts.GetUpkeeping(userAddr, userAddr, queryAddr.String())
	if err != nil {
		log.Println(g.owner, "has not deployed upkeeping")
		return err
	}
	// get upkkeeping params
	item, err := contracts.GetUpkeepingInfo(userAddr, uk)
	if err != nil {
		return err
	}

	flag := false
	for _, keeperID := range item.KeeperIDs {
		if g.localKeeper == keeperID {
			flag = true
		}
	}

	// not my user
	if !flag {
		log.Println(g.owner, "is not my user")
		return errors.New("Not my user")
	}

	g.providers = item.ProviderIDs
	g.keepers = item.KeeperIDs
	item.UserID = g.owner
	item.UpKeepingAddr = ukAddr
	g.upkeeping = &item

	return nil
}

func (k *Info) getUpkeeping(gid, uid string) (*contracts.UpKeepingItem, error) {
	gp := k.getGroupInfo(gid, uid, false)
	if gp == nil {
		return nil, errNoGroupsInfo
	}

	if gp.upkeeping == nil {
		err := gp.saveUpkeeping()
		if err != nil {
			return nil, errGetContractItem
		}
	}

	if gp.upkeeping == nil {
		return nil, errGetContractItem
	}

	return gp.upkeeping, nil
}

func (k *Info) saveQuery(qid, uid string, mode bool) error {
	gp := k.getGroupInfo(qid, uid, false)
	if gp == nil {
		return errNoGroupsInfo
	}

	if gp.query == nil || mode {
		return gp.saveQuery()
	}

	return nil
}

func (g *groupInfo) saveQuery() error {
	userAddr, err := ad.GetAddressFromID(g.owner)
	if err != nil {
		return err
	}

	queryAddr, err := contracts.GetMarketAddr(userAddr, userAddr, contracts.Query)
	if err != nil {
		return err
	}
	queryItem, err := contracts.GetQueryInfo(userAddr, queryAddr)
	if err != nil {
		return err
	}
	queryItem.UserID = g.owner
	queryItem.QueryAddr = queryAddr.String()

	g.query = &queryItem

	return nil
}

func (k *Info) getQuery(qid, uid string) (queryItem *contracts.QueryItem, err error) {
	gp := k.getGroupInfo(qid, uid, false)
	if gp == nil {
		return nil, errNoGroupsInfo
	}

	if gp.query == nil {
		err := gp.saveQuery()
		if err != nil {
			return nil, errGetContractItem
		}
	}

	if gp.query == nil {
		return nil, errGetContractItem
	}

	return gp.query, nil
}

func (k *Info) saveOffer(providerID string, update bool) error {
	thisInfo, err := k.getPInfo(providerID)
	if err != nil {
		return err
	}

	if thisInfo.offerItem == nil || update {
		return saveOfferToPinfo(providerID, thisInfo)
	}

	return nil
}

func saveOfferToPinfo(providerID string, thisInfo *pInfo) error {
	proAddr, err := ad.GetAddressFromID(providerID)
	if err != nil {
		return err
	}

	offerAddr, err := contracts.GetMarketAddr(proAddr, proAddr, contracts.Offer)
	if err != nil {
		return err
	}
	offerItem, err := contracts.GetOfferInfo(proAddr, offerAddr)
	if err != nil {
		return err
	}
	offerItem.ProviderID = providerID
	offerItem.OfferAddr = offerAddr.String()

	thisInfo.offerItem = &offerItem

	return nil
}

func (k *Info) getOffer(providerID string) (offerItem *contracts.OfferItem, err error) {
	thisInfo, err := k.getPInfo(providerID)
	if err != nil {
		return offerItem, err
	}

	if thisInfo.offerItem == nil {
		err = saveOfferToPinfo(providerID, thisInfo)
		if err != nil {
			return nil, errGetContractItem
		}
	}

	if thisInfo.offerItem == nil {
		log.Println("cannot get offerItem")
		return nil, errGetContractItem
	}

	return thisInfo.offerItem, nil
}

// addProvider 将传入pid加入posuser的upkeeping合约
func (k *Info) ukAddProvider(qid, uid, pid, sk string) error {
	gp := k.getGroupInfo(qid, uid, false)
	if gp == nil || gp.upkeeping == nil {
		return errors.New("No upkeeping")
	}

	providerAddr, err := ad.GetAddressFromID(pid)
	if err != nil {
		log.Println("ukAddProvider GetAddressFromID() error", err)
		return err
	}

	// check pid belongs to this group or not
	for _, gpid := range gp.providers {
		if gpid == pid {
			return nil
		}
	}

	userAddr, err := ad.GetAddressFromID(uid)
	if err != nil {
		log.Println("ukAddProvider GetAddressFromID() error", err)
		return err
	}

	queryAddr, err := ad.GetAddressFromID(qid)
	if err != nil {
		log.Println("ukAddProvider GetAddressFromID() error", err)
		return err
	}

	if gp.isMaster(pid) {
		log.Println("add provider to: ", userAddr)
		err = contracts.AddProvider(sk, userAddr, userAddr, []common.Address{providerAddr}, queryAddr.String())
		if err != nil {
			log.Println("ukAddProvider AddProvider() error", err)
			return err
		}
	}

	// update uk info
	gp.saveUpkeeping()

	return nil
}

func (k *Info) getKpMapRegular(ctx context.Context) {
	log.Println("Get kpMap from chain start!")

	peerID := k.localID
	contracts.SaveKpMap(peerID)
	ticker := time.NewTicker(KPMAPTIME)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go func() {
				contracts.SaveKpMap(peerID)
			}()
		}
	}
}
