package provider

import (
	"context"
	"errors"
	"log"
	"math/big"

	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
)

func handleUserDeployedContracts(km *metainfo.KeyMeta, metaValue, from string) error {
	log.Println("NewUserDeployedContracts", km.ToString(), metaValue, "From:", from)
	err := saveUpkeeping(km.GetMid())
	if err != nil {
		log.Println("Save ", km.GetMid(), "'s Upkeeping err", err)
	} else {
		log.Println("Save ", km.GetMid(), "'s Upkeeping success")
	}
	err = saveChannel(km.GetMid())
	if err != nil {
		log.Println("Save ", km.GetMid(), "'s Channel err", err)
	} else {
		log.Println("Save ", km.GetMid(), "'s Channel success")
	}
	err = saveQuery(km.GetMid())
	if err != nil {
		log.Println("Save ", km.GetMid(), "'s Query err", err)
	} else {
		log.Println("Save ", km.GetMid(), "'s Query success")
	}
	return nil
}

func saveUpkeeping(userID string) error {
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}
	ukAddr, uk, err := contracts.GetUKFromResolver(userAddr)
	if err != nil {
		return err
	}
	proAddr, err := address.GetAddressFromID(localNode.Identity.Pretty())
	if err != nil {
		return err
	}
	item, err := contracts.GetUpkeepingInfo(proAddr, uk)
	if err != nil {
		return err
	}
	item.UserID = userID
	item.UpKeepingAddr = ukAddr
	proContracts.upKeepingBook.Store(userID, item)
	return nil
}

func getUpkeeping(userID string) (contracts.UpKeepingItem, error) {
	var upkeepingItem contracts.UpKeepingItem
	value, ok := proContracts.upKeepingBook.Load(userID)
	if !ok {
		log.Println("Not find ", userID, "'s upkeepingItem in upKeepingBook")
		return upkeepingItem, errGetContractItem
	}
	upkeepingItem = value.(contracts.UpKeepingItem)
	return upkeepingItem, nil
}

func saveChannel(userID string) error {
	if pos.GetPosId() == userID {
		return nil
	}

	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}
	proID := localNode.Identity.Pretty()
	proAddr, err := address.GetAddressFromID(proID)
	if err != nil {
		return err
	}

	chanItem, err := contracts.GetChannelInfo(proAddr, proAddr, userAddr)
	if err != nil {
		return err
	}

	// 先去本地查
	var value = new(big.Int)
	km, err := metainfo.NewKeyMeta(chanItem.ChannelAddr, metainfo.Channel)
	if err != nil {
		return err
	}
	valueByte, err := localNode.Data.GetKey(context.Background(), km.ToString(), "local")
	if err != nil {
		// 本地没查到，value设为0
		log.Println("Can't get channel value in local,err :", err, ", so  set channel value to 0")
		value = big.NewInt(0)
	} else {
		log.Println("Get channel value in local:", string(valueByte))
		var ok bool
		value, ok = new(big.Int).SetString(string(valueByte), 10)
		if !ok {
			return errors.New("bigint setString error")
		}
	}
	log.Println("保存在内存中的channel.value为:", chanItem.ChannelAddr, value.String())

	chanItem.UserID = userID
	chanItem.ProID = proID
	chanItem.Value = value

	proContracts.channelBook.Store(userID, chanItem)
	return nil
}

func getChannel(userID string) (contracts.ChannelItem, error) {
	var channelItem contracts.ChannelItem
	value, ok := proContracts.channelBook.Load(userID)
	if !ok {
		saveChannel(userID)

		value, ok = proContracts.channelBook.Load(userID)
		if !ok {
			log.Println("Not find ", userID, "'s channelItem in channelBook")
			return channelItem, errGetContractItem
		}
	}
	channelItem = value.(contracts.ChannelItem)
	return channelItem, nil
}

func getChannels() []contracts.ChannelItem {
	var channels []contracts.ChannelItem
	proContracts.channelBook.Range(func(_, channnelItem interface{}) bool {
		channel, ok := channnelItem.(contracts.ChannelItem)
		if !ok {
			return false
		}
		channels = append(channels, channel)
		return true
	})
	return channels
}

func saveQuery(userID string) error {
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}
	proID := localNode.Identity.Pretty()
	proAddr, err := address.GetAddressFromID(proID)
	if err != nil {
		return err
	}
	queryAddr, err := contracts.GetMarketAddr(proAddr, userAddr, contracts.Query)
	if err != nil {
		return err
	}
	queryItem, err := contracts.GetQueryInfo(proAddr, queryAddr)
	if err != nil {
		return err
	}
	queryItem.UserID = userID
	queryItem.QueryAddr = queryAddr.String()
	proContracts.queryBook.Store(userID, queryItem)
	return nil
}

func getQuery(userID string) (contracts.QueryItem, error) {
	var queryItem contracts.QueryItem
	value, ok := proContracts.queryBook.Load(userID)
	if !ok {
		log.Println("Not find ", userID, "'s queryItem in queryBook")
		return queryItem, errGetContractItem
	}
	queryItem = value.(contracts.QueryItem)
	return queryItem, nil
}

func saveOffer() error {
	proID := localNode.Identity.Pretty()
	proAddr, err := address.GetAddressFromID(proID)
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
	offerItem.ProviderID = proID
	offerItem.OfferAddr = offerAddr.String()
	proContracts.offer = offerItem
	return nil
}

func getOffer() (contracts.OfferItem, error) {
	if proContracts.offer.OfferAddr == "" || proContracts.offer.ProviderID == "" {
		log.Println("OfferItem hasn't set")
		return proContracts.offer, errGetContractItem
	}
	return proContracts.offer, nil
}

func saveProInfo() error {
	proID := localNode.Identity.Pretty()
	proAddr, err := address.GetAddressFromID(proID)
	if err != nil {
		return err
	}
	proItem, err := contracts.GetProviderInfo(proAddr, proAddr)
	if err != nil {
		return err
	}
	proItem.ProviderID = proID
	proContracts.proInfo = proItem
	return nil
}

func getProInfo() (contracts.ProviderItem, error) {
	if proContracts.proInfo.ProviderID == "" {
		log.Println("provider info hasn't set")
		return proContracts.proInfo, errGetContractItem
	}
	return proContracts.proInfo, nil
}
