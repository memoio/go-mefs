package provider

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/memoio/go-mefs/contracts"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func SaveUpkeeping(userID string) error {
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}
	localAddr, err := address.GetAddressFromID(localNode.Identity.Pretty())
	if err != nil {
		return err
	}
	uk, _, err := contracts.GetUKFromResolver(contracts.EndPoint, userAddr)
	if err != nil {
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
	upKeepingItem := contracts.UpKeepingItem{
		UserID:        userID,
		UpKeepingAddr: uk,
		KeeperAddrs:   keepers,
		KeeperSla:     int32(len(keeperAddrs)),
		ProviderAddrs: providers,
		ProviderSla:   int32(len(providerAddrs)),
		Duration:      duration,
		Capacity:      capacity,
		Price:         price,
	}
	contracts.ProContracts.UpKeepingBook.Store(userID, upKeepingItem)
	return nil
}

func GetUpkeeping(userID string) contracts.UpKeepingItem {
	var upkeepingItem contracts.UpKeepingItem
	value, ok := contracts.ProContracts.UpKeepingBook.Load(userID)
	if !ok {
		fmt.Println("Not find ", userID, "'s upkeepingItem in upKeepingBook")
	}
	upkeepingItem = value.(contracts.UpKeepingItem)
	return upkeepingItem
}

func SaveChannel(userID string) error {
	localID := localNode.Identity.Pretty()
	localAddr, err := address.GetAddressFromID(localID)
	if err != nil {
		return err
	}
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}
	channelAddr, err := contracts.ProviderGetChannelAddr(localAddr, userAddr)
	if err != nil {
		return err
	}
	// 先去本地查
	var value = new(big.Int)
	km, err := metainfo.NewKeyMeta(channelAddr.String(), metainfo.Local, metainfo.SyncTypeChannelValue)
	if err != nil {
		return err
	}
	valueByte, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(km.ToString(), "local")
	if err != nil {
		// 本地没查到，value设为0
		fmt.Println("Can't get channel value in local,err :", err, ", so  set channel value to 0")
		value = big.NewInt(0)
	} else {
		fmt.Println("Get channel value in local:", string(valueByte))
		var ok bool
		value, ok = new(big.Int).SetString(string(valueByte), 10)
		if !ok {
			return errors.New("bigint setString error")
		}
	}
	fmt.Println("保存在内存中的channel.value为:", channelAddr.String(), value.String())
	channel := contracts.ChannelItem{
		UserID:      userID,
		ChannelAddr: channelAddr.String(),
		ProID:       localID,
		Value:       value,
	}
	contracts.ProContracts.ChannelBook.Store(userID, channel)
	return nil
}

func GetChannel(userID string) contracts.ChannelItem {
	var channelItem contracts.ChannelItem
	value, ok := contracts.ProContracts.ChannelBook.Load(userID)
	if !ok {
		fmt.Println("Not find ", userID, "'s channelItem in channelBook")
	}
	channelItem = value.(contracts.ChannelItem)
	return channelItem
}

func GetChannels() []contracts.ChannelItem {
	var channels []contracts.ChannelItem
	contracts.ProContracts.ChannelBook.Range(func(_, channnelItem interface{}) bool {
		channel, ok := channnelItem.(contracts.ChannelItem)
		if !ok {
			return false
		}
		channels = append(channels, channel)
		return true
	})
	return channels
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
	contracts.ProContracts.QueryBook.Store(userID, queryItem)
	return nil
}

func GetQuery(userID string) contracts.QueryItem {
	var queryItem contracts.QueryItem
	value, ok := contracts.ProContracts.QueryBook.Load(userID)
	if !ok {
		fmt.Println("Not find ", userID, "'s queryItem in queryBook")
	}
	queryItem = value.(contracts.QueryItem)
	return queryItem
}

func SaveOffer() error {
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
	offerAddr, err := contracts.GetMarketAddr(endPoint, localAddr, localAddr, contracts.Offer)
	if err != nil {
		return err
	}
	capacity, duration, price, err := contracts.GetOfferParams(endPoint, localAddr, offerAddr)
	if err != nil {
		return err
	}
	contracts.ProContracts.Offer = contracts.OfferItem{
		ProviderID: localID,
		OfferAddr:  offerAddr.String(),
		Capacity:   capacity.Int64(),
		Duration:   duration.Int64(),
		Price:      price.Int64(),
	}
	return nil
}

func GetOffer() contracts.OfferItem {
	return contracts.ProContracts.Offer
}
