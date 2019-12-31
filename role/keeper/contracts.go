package keeper

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	ad "github.com/memoio/go-mefs/utils/address"
)

func (k *Info) saveUpkeeping(uid, gid string, update bool) error {
	gp, ok := u.getGroupsInfo(uid, gid)
	if !ok {
		log.Println("saveUpkeeping getGroupsInfo() error")
		return errNoGroupsInfo
	}

	if gp.upkeeping == nil || update {
		return saveUpkeepingToGP(qid, gp)
	}

	return nil
}

func (g *groupInfo) saveUpkeeping() error {
	// get upkkeeping addr
	userAddr, err := ad.GetAddressFromID(g.owner)
	if err != nil {
		return err
	}
	ukAddr, uk, err := contracts.GetUKFromResolver(userAddr)
	if err != nil {
		log.Println(gp.owner, "has not deployed upkeeping")
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

func (k *Info) getUpkeeping(uid, gid string) (*contracts.UpKeepingItem, error) {
	gp, ok := u.getGroupsInfo(uid, gid)
	if !ok {
		log.Println("saveUpkeeping getGroupsInfo() error")
		return nil, errNoGroupsInfo
	}

	if gp.upkeeping == nil {
		err := saveUpkeepingToGP(gp)
		if err != nil {
			return nil, errGetContractItem
		}
	}

	if gp.upkeeping == nil {
		return nil, errGetContractItem
	}

	return gp.upkeeping, nil
}

func (k *Info) saveQuery(qid string, update bool) error {
	gp, ok := u.getGroupsInfo(qid)
	if !ok {
		return errNoGroupsInfo
	}

	if gp.query == nil || update {
		return saveQueryToGP(qid, gp)
	}

	return nil
}

func saveQueryToGP(qid string, gp *groupInfo) error {
	userAddr, err := ad.GetAddressFromID(gp.owner)
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
	queryItem.UserID = gp.owner
	queryItem.QueryAddr = queryAddr.String()

	gp.query = &queryItem

	return nil
}

func (k *Info) getQuery(qid string) (queryItem *contracts.QueryItem, err error) {
	gp, ok := u.getGroupsInfo(qid)
	if !ok {
		log.Println("saveUpkeeping getGroupsInfo() error")
		return nil, errNoGroupsInfo
	}

	if gp.query == nil {
		err := saveQueryToGP(qid, gp)
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
func (k *Info) ukAddProvider(uid, pid, sk string) error {
	uk, err := u.getUpkeeping(uid)
	if err != nil {
		log.Println("ukAddProvider getUpkeeping() error", err)
		return err
	}
	providerAddr, err := ad.GetAddressFromID(pid)
	if err != nil {
		log.Println("ukAddProvider GetAddressFromID() error", err)
		return err
	}
	for _, localPid := range uk.ProviderIDs { //判断该provider是否为新,如果存在，直接返回
		if strings.Compare(localPid, pid) == 0 {
			return nil
		}
	}

	userAddr, err := ad.GetAddressFromID(uid)
	if err != nil {
		log.Println("ukAddProvider GetAddressFromID() error", err)
		return err
	}

	if u.isMasterKeeper(uid, pid) {
		log.Println("add provider to: ", userAddr)

		err = contracts.AddProvider(sk, userAddr, []common.Address{providerAddr})
		if err != nil {
			log.Println("ukAddProvider AddProvider() error", err)
			return err
		}
	}

	// update uk info
	uk.ProviderIDs = append(uk.ProviderIDs, pid)

	return nil
}

func (k *Info) getKpMapRegular(ctx context.Context) {
	log.Println("Get kpMap from chain start!")

	peerID := k.netID
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
