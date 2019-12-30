package provider

import (
	"context"
	"errors"
	"log"
	"math/big"

	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/utils/address"
	ad "github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
)

var errBalance = errors.New("your account's balance is insufficient, we will not deploy resolver")

func providerDeployResolverAndOffer(id, sk string, capacity, duration, price int64, reDeployOffer bool) error {
	localAddress, err := ad.GetAddressFromID(id)
	if err != nil {
		log.Println("getLocalAddr err: ", err)
		return err
	}

	//获得用户的账户余额
	balance, _ := contracts.QueryBalance(localAddress.Hex())
	log.Println("balance is: ", balance)
	//先部署resolver-for-channel
	//如果部署过resolver-for-channel，那接下来就可以直接检查是否部署过offer合约，没有的话就部署
	//DeployResolver()函数内部会进行判断是否部署过
	_, err = contracts.DeployResolverForChannel(localAddress, sk)
	if err != nil {
		return err
	}

	//获得用户的账户余额
	balance, _ = contracts.QueryBalance(localAddress.Hex())
	log.Println("after deploying resolver for channel, balance is: ", balance)
	//再开始部署offer合约
	if reDeployOffer { //用户想要重新部署offer合约
		log.Println("provider wants to redeploy offer-contract")
	}
	_, err = contracts.DeployOffer(localAddress, sk, capacity, duration, price, reDeployOffer)
	if err != nil {
		return err
	}

	return nil
}

func (p *Info) handleUserDeployedContracts(km *metainfo.KeyMeta, metaValue, from string) error {
	log.Println("NewUserDeployedContracts", km.ToString(), metaValue, "From:", from)
	err := p.saveUpkeeping(km.GetMid())
	if err != nil {
		log.Println("Save ", km.GetMid(), "'s Upkeeping err", err)
	} else {
		log.Println("Save ", km.GetMid(), "'s Upkeeping success")
	}
	err = p.saveChannel(km.GetMid())
	if err != nil {
		log.Println("Save ", km.GetMid(), "'s Channel err", err)
	} else {
		log.Println("Save ", km.GetMid(), "'s Channel success")
	}
	err = p.saveQuery(km.GetMid())
	if err != nil {
		log.Println("Save ", km.GetMid(), "'s Query err", err)
	} else {
		log.Println("Save ", km.GetMid(), "'s Query success")
	}
	return nil
}

func (p *Info) saveUpkeeping(userID string) error {
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}
	ukAddr, uk, err := contracts.GetUKFromResolver(userAddr)
	if err != nil {
		return err
	}
	proAddr, err := address.GetAddressFromID(p.netID)
	if err != nil {
		return err
	}
	item, err := contracts.GetUpkeepingInfo(proAddr, uk)
	if err != nil {
		return err
	}
	item.UserID = userID
	item.UpKeepingAddr = ukAddr
	p.conManager.upKeepingBook.Store(userID, item)
	return nil
}

func (p *Info) getUpkeeping(userID string) (contracts.UpKeepingItem, error) {
	var upkeepingItem contracts.UpKeepingItem
	value, ok := p.conManager.upKeepingBook.Load(userID)
	if !ok {
		log.Println("Not find ", userID, "'s upkeepingItem in upKeepingBook")
		return upkeepingItem, errGetContractItem
	}
	upkeepingItem = value.(contracts.UpKeepingItem)
	return upkeepingItem, nil
}

func (p *Info) saveChannel(userID string) error {
	if pos.GetPosId() == userID {
		return nil
	}

	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}
	proID := p.netID
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
	valueByte, err := p.ds.GetKey(context.Background(), km.ToString(), "local")
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

	p.conManager.channelBook.Store(userID, chanItem)
	return nil
}

func (p *Info) getChannel(userID string) (contracts.ChannelItem, error) {
	var channelItem contracts.ChannelItem
	value, ok := p.conManager.channelBook.Load(userID)
	if !ok {
		p.saveChannel(userID)

		value, ok = p.conManager.channelBook.Load(userID)
		if !ok {
			log.Println("Not find ", userID, "'s channelItem in channelBook")
			return channelItem, errGetContractItem
		}
	}
	channelItem = value.(contracts.ChannelItem)
	return channelItem, nil
}

func (p *Info) getChannels() []contracts.ChannelItem {
	var channels []contracts.ChannelItem
	p.conManager.channelBook.Range(func(_, channnelItem interface{}) bool {
		channel, ok := channnelItem.(contracts.ChannelItem)
		if !ok {
			return false
		}
		channels = append(channels, channel)
		return true
	})
	return channels
}

func (p *Info) saveQuery(userID string) error {
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}
	proID := p.netID
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
	p.conManager.queryBook.Store(userID, queryItem)
	return nil
}

func (p *Info) getQuery(userID string) (contracts.QueryItem, error) {
	var queryItem contracts.QueryItem
	value, ok := p.conManager.queryBook.Load(userID)
	if !ok {
		log.Println("Not find ", userID, "'s queryItem in queryBook")
		return queryItem, errGetContractItem
	}
	queryItem = value.(contracts.QueryItem)
	return queryItem, nil
}

func (p *Info) saveOffer() error {
	proID := p.netID
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
	p.conManager.offer = offerItem
	return nil
}

func (p *Info) getOffer() (contracts.OfferItem, error) {
	if p.conManager.offer.OfferAddr == "" || p.conManager.offer.ProviderID == "" {
		log.Println("OfferItem hasn't set")
		return p.conManager.offer, errGetContractItem
	}
	return p.conManager.offer, nil
}

func (p *Info) saveProInfo() error {
	proID := p.netID
	proAddr, err := address.GetAddressFromID(proID)
	if err != nil {
		return err
	}
	proItem, err := contracts.GetProviderInfo(proAddr, proAddr)
	if err != nil {
		return err
	}
	proItem.ProviderID = proID
	p.conManager.proInfo = proItem
	return nil
}

func (p *Info) getProInfo() (contracts.ProviderItem, error) {
	if p.conManager.proInfo.ProviderID == "" {
		log.Println("provider info hasn't set")
		return p.conManager.proInfo, errGetContractItem
	}
	return p.conManager.proInfo, nil
}
