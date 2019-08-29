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
	_, keeperAddrs, providerAddrs, duration, capacity, price, err := contracts.GetUKInfoFromUK(endPoint, keeperAddr, uk)
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
	capacity, duration, price, ks, ps, completed, err := contracts.GetQueryInfo(endPoint, keeperAddr, queryAddr)
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
	capacity, duration, price, err := contracts.GetOfferInfo(endPoint, keeperAddr, offerAddr)
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
