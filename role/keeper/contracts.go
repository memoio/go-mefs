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

func saveUpkeeping(userID string, update bool) error {
	gp, ok := getGroupsInfo(userID)
	if !ok {
		log.Println("saveUpkeeping getGroupsInfo() error")
		return errNoGroupsInfo
	}

	if gp.upkeeping == nil || update {
		return saveUpkeepingToGP(userID, gp)
	}

	return nil
}

func saveUpkeepingToGP(userID string, gp *groupsInfo) error {
	// get upkkeeping addr
	userAddr, err := ad.GetAddressFromID(userID)
	if err != nil {
		return err
	}
	ukAddr, uk, err := contracts.GetUKFromResolver(userAddr)
	if err != nil {
		log.Println(userID, "has not deployed upkeeping")
		return err
	}
	// get upkkeeping params
	keeperID := localNode.Identity.Pretty()
	keeperAddr, err := ad.GetAddressFromID(keeperID)
	if err != nil {
		return err
	}
	item, err := contracts.GetUpkeepingInfo(keeperAddr, uk)
	if err != nil {
		return err
	}

	for _, kp := range item.KeeperIDs {
		if kp == keeperID {
			gp.localKeeper = kp
		}
	}

	if gp.localKeeper == userID {
		log.Println(userID, "is not my user")
		return errors.New("not my user")
	}

	item.UserID = userID
	item.UpKeepingAddr = ukAddr
	gp.upkeeping = &item
	gp.providers = item.ProviderIDs
	gp.keepers = item.KeeperIDs

	return nil
}

func getUpkeeping(userID string) (*contracts.UpKeepingItem, error) {
	gp, ok := getGroupsInfo(userID)
	if !ok {
		log.Println("saveUpkeeping getGroupsInfo() error")
		return nil, errNoGroupsInfo
	}

	if gp.upkeeping == nil {
		err := saveUpkeepingToGP(userID, gp)
		if err != nil {
			return nil, errGetContractItem
		}
	}

	if gp.upkeeping == nil {
		return nil, errGetContractItem
	}

	return gp.upkeeping, nil
}

func saveQuery(userID string, update bool) error {
	thisInfo, err := getUInfo(userID)
	if err != nil {
		return err
	}

	if thisInfo.queryItem == nil || update {
		return saveQueryToUinfo(userID, thisInfo)
	}

	return nil
}

func saveQueryToUinfo(userID string, thisInfo *uInfo) error {
	userAddr, err := ad.GetAddressFromID(userID)
	if err != nil {
		return err
	}
	keeperID := localNode.Identity.Pretty()
	keeperAddr, err := ad.GetAddressFromID(keeperID)
	if err != nil {
		return err
	}
	queryAddr, err := contracts.GetMarketAddr(keeperAddr, userAddr, contracts.Query)
	if err != nil {
		return err
	}
	queryItem, err := contracts.GetQueryInfo(keeperAddr, queryAddr)
	if err != nil {
		return err
	}
	queryItem.UserID = userID
	queryItem.QueryAddr = queryAddr.String()

	thisInfo.queryItem = &queryItem

	return nil
}

func getQuery(userID string) (queryItem *contracts.QueryItem, err error) {
	thisInfo, err := getUInfo(userID)
	if err != nil {
		return queryItem, err
	}

	if thisInfo.queryItem == nil {
		err = saveQueryToUinfo(userID, thisInfo)
		if err != nil {
			return nil, errGetContractItem
		}
	}

	if thisInfo.queryItem == nil {
		log.Println("cannot get queryItem")
		return nil, errGetContractItem
	}

	return thisInfo.queryItem, nil
}

func saveOffer(providerID string, update bool) error {
	thisInfo, err := getPInfo(providerID)
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
	keeperID := localNode.Identity.Pretty()
	keeperAddr, err := ad.GetAddressFromID(keeperID)
	if err != nil {
		return err
	}
	offerAddr, err := contracts.GetMarketAddr(keeperAddr, proAddr, contracts.Offer)
	if err != nil {
		return err
	}
	offerItem, err := contracts.GetOfferInfo(keeperAddr, offerAddr)
	if err != nil {
		return err
	}
	offerItem.ProviderID = providerID
	offerItem.OfferAddr = offerAddr.String()

	thisInfo.offerItem = &offerItem

	return nil
}

func getOffer(providerID string) (offerItem *contracts.OfferItem, err error) {
	thisInfo, err := getPInfo(providerID)
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
func ukAddProvider(uid, pid, sk string) error {
	uk, err := getUpkeeping(uid)
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

	if isMasterKeeper(uid, pid) {
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

func getKpMapRegular(ctx context.Context) {
	log.Println("Get kpMap from chain start!")

	peerID := localNode.Identity.Pretty()
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
