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

func handleUserDeployedContracts(km *metainfo.KeyMeta, metaValue, from string) error {
	fmt.Println("NewUserDeployedContracts", km.ToString(), metaValue, "From:", from)
	err := SaveUpkeeping(km.GetMid())
	if err != nil {
		fmt.Println("Save ", km.GetMid(), "'s Upkeeping err", err)
	} else {
		fmt.Println("Save ", km.GetMid(), "'s Upkeeping success")
	}
	err = SaveChannel(km.GetMid())
	if err != nil {
		fmt.Println("Save ", km.GetMid(), "'s Channel err", err)
	} else {
		fmt.Println("Save ", km.GetMid(), "'s Channel success")
	}
	err = SaveQuery(km.GetMid())
	if err != nil {
		fmt.Println("Save ", km.GetMid(), "'s Query err", err)
	} else {
		fmt.Println("Save ", km.GetMid(), "'s Query success")
	}
	err = SaveOffer()
	if err != nil {
		fmt.Println("Save ", localNode.Identity.Pretty(), "'s Offer err", err)
	} else {
		fmt.Println("Save ", localNode.Identity.Pretty(), "'s Offer success")
	}
	return nil
}

func SaveUpkeeping(userID string) error {
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}
	localAddr, err := address.GetAddressFromID(localNode.Identity.Pretty())
	if err != nil {
		return err
	}
	config, err := localNode.Repo.Config()
	if err != nil {
		return err
	}
	uk, _, err := contracts.GetUKFromResolver(config.Eth, userAddr)
	if err != nil {
		return err
	}
	_, keeperAddrs, providerAddrs, duration, capacity, price, err := contracts.GetUpKeepingParams(config.Eth, localAddr, userAddr)
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

func GetUpkeeping(userID string) (contracts.UpKeepingItem, error) {
	var upkeepingItem contracts.UpKeepingItem
	value, ok := contracts.ProContracts.UpKeepingBook.Load(userID)
	if !ok {
		fmt.Println("Not find ", userID, "'s upkeepingItem in upKeepingBook")
		return upkeepingItem, ErrGetContractItem
	}
	upkeepingItem = value.(contracts.UpKeepingItem)
	return upkeepingItem, nil
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
	config, err := localNode.Repo.Config()
	if err != nil {
		return err
	}
	channelAddr, err := contracts.ProviderGetChannelAddr(config.Eth, localAddr, userAddr)
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

func GetChannel(userID string) (contracts.ChannelItem, error) {
	var channelItem contracts.ChannelItem
	value, ok := contracts.ProContracts.ChannelBook.Load(userID)
	if !ok {
		fmt.Println("Not find ", userID, "'s channelItem in channelBook")
		return channelItem, ErrGetContractItem
	}
	channelItem = value.(contracts.ChannelItem)
	return channelItem, nil
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

func GetQuery(userID string) (contracts.QueryItem, error) {
	var queryItem contracts.QueryItem
	value, ok := contracts.ProContracts.QueryBook.Load(userID)
	if !ok {
		fmt.Println("Not find ", userID, "'s queryItem in queryBook")
		return queryItem, ErrGetContractItem
	}
	queryItem = value.(contracts.QueryItem)
	return queryItem, nil
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

func GetOffer() (contracts.OfferItem, error) {
	if contracts.ProContracts.Offer.OfferAddr == "" || contracts.ProContracts.Offer.ProviderID == "" {
		fmt.Println("OfferItem hasn't set")
		return contracts.ProContracts.Offer, ErrGetContractItem
	}
	return contracts.ProContracts.Offer, nil
}
