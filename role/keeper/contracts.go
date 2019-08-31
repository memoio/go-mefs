package keeper

import (
	"fmt"

	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/utils/address"
)

func SaveUpkeeping(gp *GroupsInfo, userID string) error {
	if gp == nil {
		return ErrIncorrectParams
	}
	// get upkkeeping addr
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}
	config, err := localNode.Repo.Config()
	if err != nil {
		return err
	}
	endPoint := config.Eth
	ukAddr, uk, err := contracts.GetUKFromResolver(endPoint, userAddr)
	if err != nil {
		fmt.Println("get ", userID, "'s ukAddr err:", err)
		return err
	}
	// get upkkeeping params
	keeperID := localNode.Identity.Pretty()
	keeperAddr, err := address.GetAddressFromID(keeperID)
	if err != nil {
		return err
	}
	item, err := contracts.GetUpkeepingInfo(endPoint, keeperAddr, uk)
	if err != nil {
		return err
	}
	item.UserID = userID
	item.UpKeepingAddr = ukAddr
	gp.upkeeping = item
	return nil
}

func GetUpkeeping(gp *GroupsInfo) (contracts.UpKeepingItem, error) {
	if gp.upkeeping.UserID == "" || gp.upkeeping.UpKeepingAddr == "" {
		fmt.Println("OfferItem hasn't set")
		return gp.upkeeping, ErrGetContractItem
	}
	return gp.upkeeping, nil
}

func SaveQuery(userID string) error {
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}
	config, err := localNode.Repo.Config()
	if err != nil {
		return err
	}
	endPoint := config.Eth
	keeperID := localNode.Identity.Pretty()
	keeperAddr, err := address.GetAddressFromID(keeperID)
	if err != nil {
		return err
	}
	queryAddr, err := contracts.GetMarketAddr(endPoint, keeperAddr, userAddr, contracts.Query)
	if err != nil {
		return err
	}
	queryItem, err := contracts.GetQueryInfo(endPoint, keeperAddr, queryAddr)
	if err != nil {
		return err
	}
	queryItem.UserID = userID
	queryItem.QueryAddr = queryAddr.String()
	localPeerInfo.queryBook.Store(userID, queryItem)
	return nil
}

func GetQuery(userID string) (contracts.QueryItem, error) {
	var queryItem contracts.QueryItem
	value, ok := localPeerInfo.queryBook.Load(userID)
	if !ok {
		fmt.Println("Not find ", userID, "'s queryItem in querybook")
		return queryItem, ErrGetContractItem
	}
	queryItem = value.(contracts.QueryItem)
	return queryItem, nil
}

func SaveOffer(providerID string) error {
	config, err := localNode.Repo.Config()
	if err != nil {
		return err
	}
	endPoint := config.Eth //获取endPoint
	proAddr, err := address.GetAddressFromID(providerID)
	if err != nil {
		return err
	}
	keeperID := localNode.Identity.Pretty()
	keeperAddr, err := address.GetAddressFromID(keeperID)
	if err != nil {
		return err
	}
	offerAddr, err := contracts.GetMarketAddr(endPoint, keeperAddr, proAddr, contracts.Offer)
	if err != nil {
		return err
	}
	offerItem, err := contracts.GetOfferInfo(endPoint, keeperAddr, offerAddr)
	if err != nil {
		return err
	}
	offerItem.ProviderID = providerID
	offerItem.OfferAddr = offerAddr.String()
	localPeerInfo.offerBook.Store(providerID, offerItem)
	return nil
}

func GetOffer(providerID string) (contracts.OfferItem, error) {
	var offerItem contracts.OfferItem
	value, ok := localPeerInfo.offerBook.Load(providerID)
	if !ok {
		fmt.Println("Not find ", providerID, "'s offerItem in offerBook")
		return offerItem, ErrGetContractItem
	}
	offerItem = value.(contracts.OfferItem)
	return offerItem, nil
}
