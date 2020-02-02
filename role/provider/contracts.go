package provider

import (
	"context"
	"math/big"

	"github.com/golang/protobuf/proto"
	"github.com/memoio/go-mefs/contracts"
	pb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func (p *Info) loadContracts() error {
	proID := p.localID

	if p.proContract == nil {
		proItem, err := role.GetProviderInfo(proID, proID)
		if err != nil {
			return err
		}
		p.proContract = &proItem
	}

	proAddr, err := address.GetAddressFromID(proID)
	if err != nil {
		return err
	}

	offers, err := contracts.GetOfferAddrs(proAddr, proAddr)
	if err != nil {
		return err
	}

	for _, offAddr := range offers {
		offerID, err := address.GetIDFromAddress(offAddr.String())
		if err != nil {
			continue
		}

		for _, item := range p.offers {
			if item.OfferID == offerID {
				continue
			}
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
		ctx := context.Background()

		gp.channel.Range(func(key, value interface{}) bool {
			cItem, ok := value.(*role.ChannelItem)
			if !ok {
				return true
			}
			km, err := metainfo.NewKeyMeta(cItem.ChannelID, metainfo.Channel)
			if err != nil {
				return true
			}
			p.ds.PutKey(ctx, km.ToString(), cItem.Sig, "local")
			return true
		})
	}

	return nil
}

func (p *Info) loadChannelValue(userID, groupID string) error {
	if userID == groupID {
		return nil
	}
	gp := p.getGroupInfo(userID, groupID, false)
	if gp != nil && gp.userID != gp.groupID {
		ctx := context.Background()
		gp.channel.Range(func(key, v interface{}) bool {
			cItem, ok := v.(*role.ChannelItem)
			if !ok {
				return true
			}

			km, err := metainfo.NewKeyMeta(cItem.ChannelID, metainfo.Channel)
			if err != nil {
				return true
			}

			valueByte, err := p.ds.GetKey(ctx, km.ToString(), "local")
			if err != nil {
				utils.MLogger.Error("get channel value from local fails: ", err)
				return true
			}

			cSign := &pb.ChannelSign{}
			err = proto.Unmarshal(valueByte, cSign)
			if err != nil {
				utils.MLogger.Error("proto.Unmarshal err:", err)
				return true
			}

			value := new(big.Int).SetBytes(cSign.GetValue())
			utils.MLogger.Info("channel value are:", value.String())
			if value.Cmp(cItem.Value) > 0 {
				cItem.Value = value
				cItem.Sig = valueByte
			}

			if cItem.Value.Cmp(cItem.Money) >= 0 {
				cSign := new(pb.ChannelSign)
				err = proto.Unmarshal(cItem.Sig, cSign)
				if err != nil {
					return true
				}

				// need verify value again
				retry := 3
				for retry > 0 {
					retry--
					err = role.CloseChannel(cItem.ChannelID, p.sk, cSign.GetSig(), cItem.Value)
					if err != nil {
						continue
					}
					break
				}
			}
			return true
		})

	}

	return nil
}

func (g *groupInfo) loadContracts(proID string) error {
	if g.userID == g.groupID {
		return nil
	}

	if g.query == nil {
		qItem, err := role.GetQueryInfo(g.userID, g.groupID)
		if err != nil {
			return err
		}
		g.query = &qItem
	}

	if g.upkeeping == nil {
		uItem, err := role.GetUpKeeping(g.userID, g.groupID)
		if err != nil {
			return err
		}
		g.upkeeping = &uItem
		g.keepers = uItem.KeeperIDs
		g.providers = uItem.ProviderIDs
	}

	cItem, err := role.GetLatestChannel(g.userID, g.groupID, proID)
	if err != nil {
		return err
	}

	g.channel.Store(cItem.ChannelID, &cItem)

	return nil
}

func (g *groupInfo) getChanItem(localID, chanID string) *role.ChannelItem {
	cv, ok := g.channel.Load(chanID)
	if !ok {
		utils.MLogger.Warn("channel is empty, reget it for: ", chanID)
		cI, err := role.GetChannelInfo(localID, chanID)
		if err != nil {
			utils.MLogger.Errorf("channelID %s is not valid", chanID)
			return nil
		}

		g.channel.Store(chanID, &cI)
		return &cI
	}

	return cv.(*role.ChannelItem)
}
