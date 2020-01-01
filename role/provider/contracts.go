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

func (p *Info) saveProvider() error {
	proID := p.localID
	proAddr, err := address.GetAddressFromID(proID)
	if err != nil {
		return err
	}
	proItem, err := contracts.GetProviderInfo(proAddr, proAddr)
	if err != nil {
		return err
	}
	proItem.ProviderID = proID
	p.proContract = &proItem
	return nil
}

func (p *Info) saveOffer() error {
	proID := p.localID
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
	p.offers = append(p.offers, &offerItem)
	return nil
}

func (p *Info) saveChannelValue(groupID, userID, proID string) error {
	gp := p.getGroupInfo(groupID, userID, false)
	if gp != nil && gp.channel != nil {
		km, err := metainfo.NewKeyMeta(gp.channel.ChannelAddr, metainfo.Channel)
		if err != nil {
			return err
		}

		// in future value: value/sig
		p.ds.PutKey(context.Background(), km.ToString(), gp.channel.Value.Bytes(), "local")
	}

	return nil
}

func (p *Info) loadChannelValue(groupID, userID string) error {
	gp := p.getGroupInfo(groupID, userID, false)
	if gp != nil {
		gp.saveChannel(p.localID)
		if gp.channel != nil {
			km, err := metainfo.NewKeyMeta(gp.channel.ChannelAddr, metainfo.Channel)
			if err != nil {
				return err
			}

			valueByte, err := p.ds.GetKey(context.Background(), km.ToString(), "local")
			if err != nil {
				log.Println("Can't get channel value in local,err :", err, ", set channel value to 0")
				gp.channel.Value = big.NewInt(0)
			} else {
				gp.channel.Value = new(big.Int).SetBytes(valueByte)
			}
		}
	}

	return nil
}

func (g *groupInfo) saveUpkeeping() error {
	userAddr, err := address.GetAddressFromID(g.userID)
	if err != nil {
		return err
	}
	ukAddr, uk, err := contracts.GetUKFromResolver(userAddr)
	if err != nil {
		return err
	}

	item, err := contracts.GetUpkeepingInfo(userAddr, uk)
	if err != nil {
		return err
	}
	item.UserID = g.userID
	item.UpKeepingAddr = ukAddr
	g.upkeeping = &item
	return nil
}

func (g *groupInfo) saveQuery() error {
	userAddr, err := address.GetAddressFromID(g.userID)
	if err != nil {
		return err
	}
	queryAddr, err := contracts.GetMarketAddr(userAddr, userAddr, contracts.Query)
	if err != nil {
		return err
	}
	queryItem, err := contracts.GetQueryInfo(userAddr, queryAddr)
	if err != nil {
		return err
	}
	queryItem.UserID = g.userID
	queryItem.QueryAddr = queryAddr.String()
	g.query = &queryItem
	return nil
}

func (g *groupInfo) saveChannel(proID string) error {
	if pos.GetPosId() == g.userID {
		return nil
	}

	userAddr, err := address.GetAddressFromID(g.userID)
	if err != nil {
		return err
	}

	proAddr, err := address.GetAddressFromID(proID)
	if err != nil {
		return err
	}

	chanItem, err := contracts.GetChannelInfo(proAddr, proAddr, userAddr)
	if err != nil {
		return err
	}

	chanItem.UserID = g.userID
	chanItem.ProID = proID
	g.channel = &chanItem
	return nil
}
