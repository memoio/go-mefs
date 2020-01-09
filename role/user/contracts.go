package user

import (
	"context"
	"math/big"
	"strings"

	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	b58 "github.com/mr-tron/base58/base58"
)

func (g *groupInfo) getContracts() error {
	if g.queryItem == nil {
		qItem, err := role.GetQueryInfo(g.userID, g.groupID)
		if err != nil {
			return err
		}
		g.queryItem = &qItem
	}

	if g.upKeepingItem == nil {
		uItem, err := role.GetUpKeeping(g.userID, g.groupID)
		if err != nil {
			return err
		}
		g.upKeepingItem = &uItem
	}

	ctx := context.Background()

	for _, proInfo := range g.providers {
		proID := proInfo.providerID

		if proInfo.offerItem == nil {
			oItem, err := role.GetLatestOffer(g.userID, proID)
			if err != nil {
				return err
			}
			proInfo.offerItem = &oItem
		}

		if proInfo.chanItem == nil {
			cItem, err := role.GetLatestChannel(g.userID, g.groupID, proID)
			if err != nil {
				return err
			}
			proInfo.chanItem = &cItem

			var value = new(big.Int)
			km, err := metainfo.NewKeyMeta(proInfo.chanItem.ChannelID, metainfo.Channel)
			if err != nil {
				return nil
			}
			valueByte, err := g.ds.GetKey(ctx, km.ToString(), "local")
			if err != nil {
				utils.MLogger.Info("try to get channel value from: ", proID)
				valueByte, err = g.ds.GetKey(ctx, km.ToString(), proID)
				if err != nil {
					utils.MLogger.Info("Set channel price to 0.")
					value = big.NewInt(0)
				} else {
					value, _ = new(big.Int).SetString(string(valueByte), 10)
				}
			} else {
				value, _ = new(big.Int).SetString(string(valueByte), 10)
			}
			proInfo.chanItem.Value = value
		}
	}

	return nil
}

func (g *groupInfo) loadChannelValue() {

	ctx := context.Background()

	for _, proInfo := range g.providers {
		if proInfo.chanItem != nil {
			proID := proInfo.providerID

			km, err := metainfo.NewKeyMeta(proInfo.chanItem.ChannelID, metainfo.Channel)
			if err != nil {
				continue
			}

			valueByte, err := g.ds.GetKey(ctx, km.ToString(), "local")
			if err != nil {
				utils.MLogger.Info("try to get channel value from: ", proID)
				valueByte, _ = g.ds.GetKey(ctx, km.ToString(), proID)
			}

			value := big.NewInt(0)
			var sig []byte
			vals := strings.Split(string(valueByte), metainfo.DELIMITER)
			if len(vals) == 2 {
				value.SetString(vals[0], 10)
				sig, _ = b58.Decode(vals[1])
			}

			utils.MLogger.Info("channel value are:", value.String())
			if value.Cmp(proInfo.chanItem.Value) > 0 {
				proInfo.chanItem.Value = value
				proInfo.chanItem.Sig = sig
			}
		}
	}

	return
}

func (g *groupInfo) saveChannelValue() {
	ctx := context.Background()
	for _, proInfo := range g.providers {
		if proInfo.chanItem != nil && proInfo.chanItem.Dirty {

			km, err := metainfo.NewKeyMeta(proInfo.chanItem.ChannelID, metainfo.Channel)
			if err != nil {
				continue
			}

			sig := b58.Encode(proInfo.chanItem.Sig)
			value := proInfo.chanItem.Value.String()

			metavalue := value + metainfo.DELIMITER + sig
			g.ds.PutKey(ctx, km.ToString(), []byte(metavalue), "local")
		}
	}

	return
}
