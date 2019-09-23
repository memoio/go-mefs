package keeper

import (
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/utils/address"
	ad "github.com/memoio/go-mefs/utils/address"
)

func saveUpkeeping(gp *GroupsInfo, userID string) error {
	if gp == nil {
		return ErrIncorrectParams
	}
	// get upkkeeping addr
	userAddr, err := address.GetAddressFromID(userID)
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
	keeperAddr, err := address.GetAddressFromID(keeperID)
	if err != nil {
		return err
	}
	item, err := contracts.GetUpkeepingInfo(keeperAddr, uk)
	if err != nil {
		return err
	}
	item.UserID = userID
	item.UpKeepingAddr = ukAddr
	gp.upkeeping = item
	gp.Providers = item.ProviderIDs

	return nil
}

func getUpkeeping(gp *GroupsInfo) (contracts.UpKeepingItem, error) {
	if gp.upkeeping.UserID == "" || gp.upkeeping.UpKeepingAddr == "" {
		return gp.upkeeping, ErrGetContractItem
	}
	return gp.upkeeping, nil
}

func saveQuery(userID string) error {
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}
	keeperID := localNode.Identity.Pretty()
	keeperAddr, err := address.GetAddressFromID(keeperID)
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
	localPeerInfo.queryBook.Store(userID, queryItem)
	return nil
}

func getQuery(userID string) (contracts.QueryItem, error) {
	var queryItem contracts.QueryItem
	value, ok := localPeerInfo.queryBook.Load(userID)
	if !ok {
		log.Println("Not find ", userID, "'s queryItem in querybook")
		return queryItem, ErrGetContractItem
	}
	queryItem = value.(contracts.QueryItem)
	return queryItem, nil
}

func saveOffer(providerID string) error {
	proAddr, err := address.GetAddressFromID(providerID)
	if err != nil {
		return err
	}
	keeperID := localNode.Identity.Pretty()
	keeperAddr, err := address.GetAddressFromID(keeperID)
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
	localPeerInfo.offerBook.Store(providerID, offerItem)
	return nil
}

func getOffer(providerID string) (contracts.OfferItem, error) {
	var offerItem contracts.OfferItem
	value, ok := localPeerInfo.offerBook.Load(providerID)
	if !ok {
		log.Println("Not find ", providerID, "'s offerItem in offerBook")
		return offerItem, ErrGetContractItem
	}
	offerItem = value.(contracts.OfferItem)
	return offerItem, nil
}

// addProvider 将传入pid加入posuser的upkeeping合约
func ukAddProvider(uid, pid, sk string) error {
	gp, ok := getGroupsInfo(uid)
	if !ok {
		log.Println("ukAddProvider getGroupsInfo() error")
		return ErrNoGroupsInfo
	}
	uk, err := getUpkeeping(gp)
	if err != nil {
		err := saveUpkeeping(gp, uid)
		if err != nil {
			log.Println("ukAddProvider getUpkeeping() error", err)
			return err
		}

		uk, err = getUpkeeping(gp) //保存之后重试。还是出错就返回
		if err != nil {
			log.Println("ukAddProvider getUpkeeping() error", err)
			return err
		}
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
