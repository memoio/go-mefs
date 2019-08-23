package keeper

import (
	"fmt"

	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/utils/address"
)

func SaveUpkeeping(gp *GroupsInfo, userID string) error {
	config, err := localNode.Repo.Config()
	if err != nil {
		return err
	}
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}
	localAddr, err := address.GetAddressFromID(config.PeerID)
	if err != nil {
		return err
	}
	endPoint := config.Eth //获取endPoint
	ukAddr, _, err := contracts.GetUKFromResolver(endPoint, userAddr)
	if err != nil {
		fmt.Println("get ", userID, "'s ukAddr err:", err)
		return err
	}
	_, keeperAddrs, providerAddrs, duration, capacity, price, err := contracts.GetUpKeepingParams(contracts.EndPoint, localAddr, userAddr)
	if err != nil {
		return err
	}
	var keepers []string
	var providers []string
	for _, keeper := range keeperAddrs {
		keepers = append(keepers, keeper.String())
	}
	for _, provider := range providerAddrs {
		providers = append(providers, provider.String())
	}
	gp.upkeeping = contracts.UpKeepingItem{
		UserID:        userID,
		UpKeepingAddr: ukAddr,
		KeeperAddrs:   keepers,
		KeeperSla:     int32(len(keeperAddrs)),
		ProviderAddrs: providers,
		ProviderSla:   int32(len(providerAddrs)),
		Duration:      duration,
		Capacity:      capacity,
		Price:         price,
	}
	return nil
}

func GetUpkeeping(gp *GroupsInfo) contracts.UpKeepingItem {
	return gp.upkeeping
}

func SaveQuery(userID string) error {
	localID := localNode.Identity.Pretty()
	config, err := localNode.Repo.Config()
	if err != nil {
		return err
	}
	localAddr, err := address.GetAddressFromID(localID)
	if err != nil {
		return err
	}
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}
	endPoint := config.Eth //获取endPoint
	queryAddr, err := contracts.GetMarketAddr(endPoint, localAddr, userAddr, contracts.Query)
	if err != nil {
		return err
	}
	capacity, duration, price, ks, ps, completed, err := contracts.GetQueryParams(endPoint, localAddr, queryAddr)
	if err != nil {
		return err
	}
	queryItem := contracts.QueryItem{
		UserID:       userID,
		QueryAddr:    queryAddr.String(),
		Capacity:     capacity.Int64(),
		Duration:     duration.Int64(),
		Price:        price.Int64(),
		KeeperNums:   int32(ks.Int64()),
		ProviderNums: int32(ps.Int64()),
		Completed:    completed,
	}
	localPeerInfo.queryBook.Store(userID, queryItem)
	return nil
}

func GetQuery(userID string) contracts.QueryItem {
	var queryItem contracts.QueryItem
	value, ok := localPeerInfo.queryBook.Load(userID)
	if !ok {
		fmt.Println("Not find ", userID, "'s queryItem in querybook")
	}
	queryItem = value.(contracts.QueryItem)
	return queryItem
}

func SaveOffer(providerID string) error {
	localID := localNode.Identity.Pretty()
	config, err := localNode.Repo.Config()
	if err != nil {
		return err
	}
	localAddr, err := address.GetAddressFromID(localID)
	if err != nil {
		return err
	}
	endPoint := config.Eth //获取endPoint
	proAddr, err := address.GetAddressFromID(providerID)
	if err != nil {
		return err
	}
	offerAddr, err := contracts.GetMarketAddr(endPoint, localAddr, proAddr, contracts.Offer)
	if err != nil {
		return err
	}
	capacity, duration, price, err := contracts.GetOfferParams(endPoint, localAddr, offerAddr)
	if err != nil {
		return err
	}
	offerItem := contracts.OfferItem{
		ProviderID: providerID,
		OfferAddr:  offerAddr.String(),
		Capacity:   capacity.Int64(),
		Duration:   duration.Int64(),
		Price:      price.Int64(),
	}
	localPeerInfo.offerBook.Store(providerID, offerItem)
	return nil
}

func GetOffer(providerID string) contracts.OfferItem {
	var offerItem contracts.OfferItem
	value, ok := localPeerInfo.offerBook.Load(providerID)
	if !ok {
		fmt.Println("Not find ", providerID, "'s offerItem in offerBook")
	}
	offerItem = value.(contracts.OfferItem)
	return offerItem
}
