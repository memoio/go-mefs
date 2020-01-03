package provider

import (
	"context"
	"log"
	"math/big"
	"strings"

	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	b58 "github.com/mr-tron/base58/base58"
)

func (p *Info) getContracts() error {
	proID := p.localID
	proItem, err := role.GetProviderInfo(proID, proID)
	if err != nil {
		return err
	}
	p.proContract = &proItem

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
		oItem, err := role.GetOfferInfo(proID, offerID)
		if err != nil {
			continue
		}
		p.offers = append(p.offers, &oItem)
	}

	return nil
}

func (p *Info) saveChannelValue(userID, groupID, proID string) error {
	gp := p.getGroupInfo(userID, groupID, false)
	if gp != nil && gp.channel != nil {
		ctx := context.Background()
		km, err := metainfo.NewKeyMeta(gp.channel.ChannelID, metainfo.Channel)
		if err != nil {
			return err
		}

		sig := b58.Encode(gp.channel.Sig)
		value := gp.channel.Value.String()

		metavalue := value + metainfo.DELIMITER + sig
		p.ds.PutKey(ctx, km.ToString(), []byte(metavalue), "local")
	}

	return nil
}

func (p *Info) loadChannelValue(userID, groupID string) error {
	gp := p.getGroupInfo(userID, groupID, false)
	if gp != nil && gp.userID != gp.groupID && gp.channel != nil {
		ctx := context.Background()
		km, err := metainfo.NewKeyMeta(gp.channel.ChannelID, metainfo.Channel)
		if err != nil {
			return err
		}

		valueByte, err := p.ds.GetKey(ctx, km.ToString(), "local")
		if err != nil {
			log.Println("try to get channel value from: ", gp.userID)
			valueByte, _ = p.ds.GetKey(ctx, km.ToString(), gp.userID)
		}

		value := big.NewInt(0)
		var sig []byte
		vals := strings.Split(string(valueByte), metainfo.DELIMITER)
		if len(vals) == 2 {
			value.SetString(vals[0], 10)
			sig, _ = b58.Decode(vals[1])
		}

		log.Println("channel value are:", value.String())
		if value.Cmp(gp.channel.Value) > 0 {
			gp.channel.Value = value
			gp.channel.Sig = sig
		}
	}

	return nil
}

func (g *groupInfo) getContracts(proID string) error {
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
