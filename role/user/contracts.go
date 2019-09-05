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
	err := cs.SaveUpkeeping()
	if err != nil {
		return err
	}
	err = cs.SaveChannel()
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

func (cs *ContractService) SaveUpkeeping() error {
	userAddr, err := address.GetAddressFromID(cs.UserID)
	if err != nil {
		return err
	}
	config, err := localNode.Repo.Config()
	if err != nil {
		return err
	}
	ukAddr, uk, err := contracts.GetUKFromResolver(config.Eth, userAddr)
	if err != nil {
		return err
	}
	item, err := contracts.GetUpkeepingInfo(config.Eth, userAddr, uk)
	if err != nil {
		return err
	}
	item.UserID = cs.UserID
	item.UpKeepingAddr = ukAddr
	cs.upKeepingItem = item
	return nil
}

func (cs *ContractService) SaveChannel() error {
	userAddr, err := address.GetAddressFromID(cs.UserID)
	if err != nil {
		return err
	}
	uk, err := cs.GetUpkeepingItem()
	if err != nil {
		return err
	}
	config, err := localNode.Repo.Config()
	if err != nil {
		return err
	}
	for _, proId := range uk.ProviderIDs {
		if _, ok := cs.channelBook[proId]; ok {
			continue
		}

		proAddr, err := address.GetAddressFromID(proId)
		if err != nil {
			return err
		}
		chanAddr, err := contracts.GetChannelAddr(config.Eth, userAddr, proAddr, userAddr)
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
			fmt.Println("Can't get channel value in local,err :", err, ", so try to get from ", proId)
			valueByte, err = localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(channelValueKeyMeta.ToString(), proId)
			if err != nil {
				// provider上也没找到，value设为0
				fmt.Println("Can't get channel price from ", proId, ",err :", err, ", so set channel price to 0.")
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
		time, err := contracts.GetChannelStartDate(config.Eth, userAddr, proAddr, userAddr)
		if err != nil {
			return err
		}
		channel := contracts.ChannelItem{
			UserID:      cs.UserID,
			ChannelAddr: chanAddr.String(),
			ProID:       proId,
			Value:       value,
			StartTime:   time,
		}
		cs.channelBook[proId] = channel
	}
	return nil
}

func (cs *ContractService) SaveQuery() error {
	userAddr, err := address.GetAddressFromID(cs.UserID)
	if err != nil {
		return err
	}
	config, err := localNode.Repo.Config()
	if err != nil {
		return err
	}
	queryAddr, err := contracts.GetMarketAddr(config.Eth, userAddr, userAddr, contracts.Query)
	if err != nil {
		return err
	}
	item, err := contracts.GetQueryInfo(config.Eth, userAddr, queryAddr)
	if err != nil {
		return err
	}
	item.UserID = cs.UserID
	item.QueryAddr = queryAddr.String()
	cs.queryItem = item
	return nil
}

func (cs *ContractService) SaveOffer() error {
	userAddr, err := address.GetAddressFromID(cs.UserID)
	if err != nil {
		return err
	}
	uk, err := cs.GetUpkeepingItem()
	if err != nil {
		return err
	}
	config, err := localNode.Repo.Config()
	if err != nil {
		return err
	}
	for _, proId := range uk.ProviderIDs {
		if _, ok := cs.offerBook[proId]; ok {
			continue
		}

		proAddr, err := address.GetAddressFromID(proId)
		if err != nil {
			return err
		}

		offerAddr, err := contracts.GetMarketAddr(config.Eth, userAddr, proAddr, contracts.Offer)
		if err != nil {
			fmt.Println("get", proAddr.String(), "'s offer address err ")
			return err
		}
		offerItem, err := contracts.GetOfferInfo(config.Eth, userAddr, offerAddr)
		if err != nil {
			fmt.Println("get", proAddr.String(), "'s offer params err ")
			return err
		}
		offerItem.ProviderID = proId
		offerItem.OfferAddr = offerAddr.String()
		cs.offerBook[proId] = offerItem
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

func (cs *ContractService) GetUpkeepingItem() (contracts.UpKeepingItem, error) {
	if cs.upKeepingItem.UpKeepingAddr == "" || cs.upKeepingItem.UserID == "" {
		fmt.Println("UpKeepingItem hasn't set")
		return cs.upKeepingItem, ErrGetContractItem
	}
	return cs.upKeepingItem, nil
}

func (cs *ContractService) GetQueryItem() (contracts.QueryItem, error) {
	if cs.queryItem.QueryAddr == "" || cs.queryItem.UserID == "" {
		fmt.Println("QueryItem hasn't set")
		return cs.queryItem, ErrGetContractItem
	}
	return cs.queryItem, nil
}
