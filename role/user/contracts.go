package user

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/memoio/go-mefs/contracts"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func ConstructContractService(userID string) *ContractService {
	var upkeeping contracts.UpKeepingItem
	var query contracts.QueryItem
	return &ContractService{
		UserID:        userID,
		channelBook:   make(map[string]contracts.ChannelItem),
		upKeepingItem: upkeeping,
		offerBook:     make(map[string]contracts.OfferItem),
		queryItem:     query,
	}
}

func (cs *ContractService) SaveContracts() error {
	err := cs.SaveChannel()
	if err != nil {
		return err
	}
	err = cs.SaveUpkeeping()
	if err != nil {
		return err
	}
	err = cs.SaveQuery()
	if err != nil {
		return err
	}
	err = cs.SaveOffer()
	if err != nil {
		return err
	}
	return nil
}

func (cs *ContractService) SaveChannel() error {
	userAddr, err := address.GetAddressFromID(cs.UserID)
	if err != nil {
		return err
	}
	gp := GetGroupService(cs.UserID)
	// 获得user的所有provider
	providers, err := gp.GetProviders(-1)
	if err != nil {
		return err
	}
	for _, provider := range providers {
		if _, ok := cs.channelBook[provider]; ok {
			continue
		}
		proAddr, err := address.GetAddressFromID(provider)
		if err != nil {
			return err
		}
		chanAddr, err := contracts.UserGetChannelAddr(userAddr, proAddr)
		if err != nil {
			return err
		}
		// 先从本地找value
		var value = new(big.Int)
		channelValueKeyMeta, err := metainfo.NewKeyMeta(chanAddr.String(), metainfo.Local, metainfo.SyncTypeChannelValue)
		if err != nil {
			return err
		}

		valueByte, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(channelValueKeyMeta.ToString(), "local")
		if err != nil {
			// 本地没找到，从provider上找
			fmt.Println("Can't get channel value in local,err :", err, ", so try to get from ", provider)
			valueByte, err = localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(channelValueKeyMeta.ToString(), provider)
			if err != nil {
				// provider上也没找到，value设为0
				fmt.Println("Can't get channel price from ", provider, ",err :", err, ", so set channel price to 0.")
				value = big.NewInt(0)
			} else {
				//provider上找到了
				var ok bool
				value, ok = new(big.Int).SetString(string(valueByte), 10)
				if !ok {
					return errors.New("bigInt.SetString err")
				}
			}
		} else {
			//本地找到了
			var ok bool
			value, ok = new(big.Int).SetString(string(valueByte), 10)
			if !ok {
				return errors.New("bigInt.SetString err")
			}
		}
		fmt.Println("保存在内存中的channel地址和value为:", chanAddr.String(), value.String())
		channel := contracts.ChannelItem{
			UserID:      gp.Userid,
			ChannelAddr: chanAddr.String(),
			ProID:       provider,
			Value:       value,
		}
		cs.channelBook[provider] = channel
	}
	return nil
}

func (cs *ContractService) SaveUpkeeping() error {
	userAddr, err := address.GetAddressFromID(cs.UserID)
	if err != nil {
		return err
	}
	uk, _, err := contracts.GetUKFromResolver(contracts.EndPoint, userAddr)
	if err != nil {
		return err
	}
	_, keeperAddrs, providerAddrs, duration, capacity, price, err := contracts.GetUpKeepingParams(contracts.EndPoint, userAddr, userAddr)
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
	cs.upKeepingItem = contracts.UpKeepingItem{
		UserID:        cs.UserID,
		UpKeepingAddr: uk,
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

func (cs *ContractService) SaveQuery() error {
	userAddr, err := address.GetAddressFromID(cs.UserID)
	if err != nil {
		return err
	}
	queryAddr, err := contracts.GetMarketAddr(contracts.EndPoint, userAddr, userAddr, contracts.Query)
	if err != nil {
		return err
	}
	capacity, duration, price, ks, ps, completed, err := contracts.GetQueryParams(contracts.EndPoint, userAddr, queryAddr)
	if err != nil {
		return err
	}
	cs.queryItem = contracts.QueryItem{
		UserID:       cs.UserID,
		QueryAddr:    queryAddr.String(),
		Capacity:     capacity.Int64(),
		Duration:     duration.Int64(),
		Price:        price.Int64(),
		KeeperNums:   int32(ks.Int64()),
		ProviderNums: int32(ps.Int64()),
		Completed:    completed,
	}
	return nil
}

func (cs *ContractService) SaveOffer() error {
	userAddr, err := address.GetAddressFromID(cs.UserID)
	if err != nil {
		return err
	}
	gp := GetGroupService(cs.UserID)
	// 获得user的所有provider
	providers, err := gp.GetProviders(-1)
	if err != nil {
		return err
	}
	for _, provider := range providers {
		if _, ok := cs.offerBook[provider]; ok {
			continue
		}
		proAddr, err := address.GetAddressFromID(provider)
		if err != nil {
			return err
		}
		offerAddr, err := contracts.GetMarketAddr(contracts.EndPoint, userAddr, proAddr, contracts.Offer)
		if err != nil {
			fmt.Println("get", provider, "'s offer address err ")
			return err
		}
		capacity, duration, price, err := contracts.GetOfferParams(contracts.EndPoint, userAddr, offerAddr)
		if err != nil {
			fmt.Println("get", provider, "'s offer params err ")
			return err
		}
		offer := contracts.OfferItem{
			ProviderID: provider,
			OfferAddr:  offerAddr.String(),
			Capacity:   capacity.Int64(),
			Duration:   duration.Int64(),
			Price:      price.Int64(),
		}
		cs.offerBook[provider] = offer
	}
	return nil
}

func (cs *ContractService) GetChannelItem(proid string) (contracts.ChannelItem, error) {
	channelItem, ok := cs.channelBook[proid]
	if !ok {
		return channelItem, ErrGetContractItem
	}
	return channelItem, nil
}

func (cs *ContractService) GetOfferItem(proid string) (contracts.OfferItem, error) {
	offerItem, ok := cs.offerBook[proid]
	if !ok {
		return offerItem, ErrGetContractItem
	}
	return offerItem, nil
}

func (cs *ContractService) GetUpkeepingItem() contracts.UpKeepingItem {
	return cs.upKeepingItem
}

func (cs *ContractService) GetQueryItem() contracts.QueryItem {
	return cs.queryItem
}
