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

	if len(offers) >= len(p.offers) {
		return nil
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
	if gp != nil && gp.userID != gp.groupID && gp.channel != nil && gp.channel.Sig != nil {
		ctx := context.Background()
		km, err := metainfo.NewKeyMeta(gp.channel.ChannelID, metainfo.Channel)
		if err != nil {
			return err
		}

		p.ds.PutKey(ctx, km.ToString(), gp.channel.Sig, "local")
	}

	return nil
}

func (p *Info) loadChannelValue(userID, groupID string) error {
	if userID == groupID {
		return nil
	}
	gp := p.getGroupInfo(userID, groupID, false)
	if gp != nil && gp.userID != gp.groupID && gp.channel != nil {
		ctx := context.Background()
		km, err := metainfo.NewKeyMeta(gp.channel.ChannelID, metainfo.Channel)
		if err != nil {
			return err
		}

		valueByte, err := p.ds.GetKey(ctx, km.ToString(), "local")
		if err != nil {
			utils.MLogger.Info("try to get channel value from: ", gp.userID)
			valueByte, _ = p.ds.GetKey(ctx, km.ToString(), gp.userID)
		}

		cSign := &pb.ChannelSign{}
		err = proto.Unmarshal(valueByte, cSign)
		if err != nil {
			utils.MLogger.Error("proto.Unmarshal err:", err)
			return err
		}

		value := new(big.Int).SetBytes(cSign.GetValue())
		utils.MLogger.Info("channel value are:", value.String())
		if value.Cmp(gp.channel.Value) > 0 {
			gp.channel.Value = value
			gp.channel.Sig = cSign.GetSig()
		}
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
	}

	if g.channel == nil {
		cItem, err := role.GetLatestChannel(g.userID, g.groupID, proID)
		if err != nil {
			return err
		}
		g.channel = &cItem
	}

	return nil
}
