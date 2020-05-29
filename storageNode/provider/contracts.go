package provider

import (
	"math/big"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/memoio/go-mefs/contracts"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func (p *Info) loadContracts(capacity, duration, depositSize int64, price *big.Int, reDeployOffer bool) error {
	proID := p.localID

	if p.proContract == nil {
		proItem, err := role.GetProviderInfo(proID, proID)
		if err != nil {
			return err
		}

		// pledge again
		if proItem.Capacity < depositSize {
			utils.MLogger.Infof("your old pledge capacity is %d, now will change to %d", proItem.Capacity, depositSize)
			dsize := new(big.Int).SetInt64(depositSize - proItem.Capacity)
			err := role.PledgeProvider(proID, p.sk, dsize)
			if err != nil {
				return err
			}

			proItem, err = role.GetProviderInfo(proID, proID)
			if err != nil {
				return err
			}
		}

		p.proContract = &proItem
	}

	proAddr, err := address.GetAddressFromID(proID)
	if err != nil {
		return err
	}

	if capacity > p.proContract.Capacity {
		utils.MLogger.Infof("your offer-capacity is %d, more than your deposit capacity %d, so change it to %d", capacity, p.proContract.Capacity, p.proContract.Capacity)
		capacity = p.proContract.Capacity
	}

	_, err = role.DeployOffer(p.localID, p.sk, capacity, duration, price, reDeployOffer)
	if err != nil {
		return err
	}

	offers, err := contracts.GetOfferAddrs(proAddr, proAddr)
	if err != nil {
		return err
	}

	//p.offers保存该provider所有的offer合约信息
	for _, offAddr := range offers {
		offerID, err := address.GetIDFromAddress(offAddr.String())
		if err != nil {
			continue
		}

		has := false
		for _, item := range p.offers {
			if item.OfferID == offerID {
				has = true
				break
			}
		}

		if has {
			continue
		}

		oItem, err := role.GetOfferInfo(proID, offerID)
		if err != nil {
			continue
		}

		p.offers = append(p.offers, &oItem)
	}

	return nil
}

func (p *Info) saveChannelValue(userID, groupID, proID string) error {
	if userID == groupID {
		return nil
	}

	gp := p.getGroupInfo(userID, groupID, false)
	if gp != nil && gp.userID != gp.groupID {
		ctx := p.context

		chs := gp.getChannels()
		for _, ch := range chs {
			ci, ok := gp.channel.Load(ch)
			if !ok {
				continue
			}
			cItem := ci.(*role.ChannelItem)
			km, err := metainfo.NewKey(p.localID, mpb.KeyType_Channel, cItem.ChannelID)
			if err != nil {
				continue
			}
			p.ds.PutKey(ctx, km.ToString(), cItem.Sig, nil, "local")

			if time.Now().Unix()-cItem.StartTime > cItem.Duration-int64(60*60) && cItem.Money.Sign() > 0 {
				cSign := new(mpb.ChannelSign)
				err = proto.Unmarshal(cItem.Sig, cSign)
				if err != nil {
					continue
				}

				// need verify value again;
				err = role.CloseChannel(cItem.ChannelID, p.sk, cSign.GetSig(), cItem.Value)
				if err != nil {
					utils.MLogger.Errorf("close channel %s err: %s", cItem.ChannelID, err)
					continue
				}

				cItem.Money = role.GetBalance(cItem.ChannelID)
			}
		}
	}

	return nil
}

func (p *Info) loadChannelValue(userID, groupID string) error {
	if userID == groupID {
		return nil
	}
	gp := p.getGroupInfo(userID, groupID, false)
	if gp != nil && gp.userID != gp.groupID {
		ctx := p.context

		for _, proID := range gp.providers {
			chanIDs, err := role.GetAllChannels(userID, groupID, proID)
			if err != nil {
				continue
			}

			clen := len(chanIDs)

			for i := 0; i < clen-1; i++ {
				ba, err := role.QueryBalance(chanIDs[i])
				if err != nil {
					continue
				}

				if ba.Sign() <= 0 {
					continue
				}

				km, err := metainfo.NewKey(p.localID, mpb.KeyType_Channel, chanIDs[i])
				if err != nil {
					continue
				}

				valueByte, err := p.ds.GetKey(ctx, km.ToString(), "local")
				if err != nil {
					utils.MLogger.Error("get channel value from local fails: ", err)
					continue
				}

				cSign := &mpb.ChannelSign{}
				err = proto.Unmarshal(valueByte, cSign)
				if err != nil {
					utils.MLogger.Error("proto.Unmarshal channelSign err:", err)
					continue
				}

				err = role.CloseChannel(chanIDs[i], p.sk, cSign.GetSig(), new(big.Int).SetBytes(cSign.GetValue()))
				if err != nil {
					utils.MLogger.Errorf("close channel %s err: %s", chanIDs[i], err)
					continue
				}
			}
		}

		chs := gp.getChannels()
		for _, ch := range chs {
			ci, ok := gp.channel.Load(ch)
			if !ok {
				continue
			}
			cItem := ci.(*role.ChannelItem)
			if cItem.Money.Cmp(big.NewInt(0)) == 0 {
				continue
			}

			km, err := metainfo.NewKey(p.localID, mpb.KeyType_Channel, cItem.ChannelID)
			if err != nil {
				continue
			}

			valueByte, err := p.ds.GetKey(ctx, km.ToString(), "local")
			if err != nil {
				utils.MLogger.Error("get channel value from local fails: ", err)
				continue
			}

			cSign := &mpb.ChannelSign{}
			err = proto.Unmarshal(valueByte, cSign)
			if err != nil {
				utils.MLogger.Error("proto.Unmarshal channelSign err:", err)
				continue
			}

			value := new(big.Int).SetBytes(cSign.GetValue())
			utils.MLogger.Info("channel value are:", value.String())
			if value.Cmp(cItem.Value) > 0 {
				cItem.Value = value
				cItem.Sig = valueByte
			}

			cItem.Money = role.GetBalance(cItem.ChannelID)

			// close before timeout
			if time.Now().Unix()-cItem.StartTime > cItem.Duration-int64(60*60) && cItem.Money.Sign() > 0 {
				cSign := new(mpb.ChannelSign)
				err = proto.Unmarshal(cItem.Sig, cSign)
				if err != nil {
					continue
				}

				// need verify value again;
				err = role.CloseChannel(cItem.ChannelID, p.sk, cSign.GetSig(), cItem.Value)
				if err != nil {
					utils.MLogger.Errorf("close channel %s err: %s", cItem.ChannelID, err)
					continue
				}

				cItem.Money = role.GetBalance(cItem.ChannelID)
			}
		}
	}

	return nil
}

func (g *groupInfo) loadContracts(proID string, mode bool) error {
	if g.userID == g.groupID {
		return nil
	}

	if g.query == nil || mode {
		qItem, err := role.GetQueryInfo(g.userID, g.groupID)
		if err != nil {
			return err
		}
		g.query = &qItem
	}

	if g.upkeeping == nil || mode {
		uItem, err := role.GetUpKeeping(g.userID, g.groupID)
		if err != nil {
			return err
		}

		var keepers []string
		var providers []string
		for _, keeper := range uItem.Keepers {
			kid, err := address.GetIDFromAddress(keeper.Addr.String())
			if err != nil {
				return err
			}
			keepers = append(keepers, kid)
		}

		for _, provider := range uItem.Providers {
			pid, err := address.GetIDFromAddress(provider.Addr.String())
			if err != nil {
				return err
			}
			providers = append(providers, pid)
		}

		g.upkeeping = &uItem
		g.keepers = keepers
		g.providers = providers
	}

	cItem, err := role.GetLatestChannel(g.userID, g.groupID, proID)
	if err != nil {
		return err
	}

	if cItem.Money.Cmp(big.NewInt(0)) > 0 {
		g.channel.Store(cItem.ChannelID, &cItem)
	}

	return nil
}

func (g *groupInfo) getChannels() []string {
	var res []string
	g.channel.Range(func(key, value interface{}) bool {
		res = append(res, key.(string))
		return true
	})
	return res
}

func (g *groupInfo) getChanItem(localID, chanID string) *role.ChannelItem {
	cv, ok := g.channel.Load(chanID)
	if !ok {
		utils.MLogger.Info("channel is empty, reget it for: ", chanID)
		cItem, err := role.GetChannelInfo(localID, chanID)
		if err != nil {
			utils.MLogger.Errorf("channelID %s is not valid: %s", chanID, err)
			return nil
		}

		if cItem.Money.Cmp(big.NewInt(0)) > 0 {
			g.channel.Store(chanID, &cItem)
			return &cItem
		}

		return nil
	}

	cItem := cv.(*role.ChannelItem)

	if cItem.Money.Cmp(big.NewInt(0)) == 0 {
		g.channel.Delete(chanID)
	}

	return cItem
}
