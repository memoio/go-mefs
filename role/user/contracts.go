package user

import (
	"context"
	"log"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func deployQuery(userID, sk string, storeDays, storeSize, storePrice int64, ks, ps int, rdo bool) error {
	uaddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}
	// getbalance
	balance, err := contracts.QueryBalance(uaddr.String())
	if err != nil {
		return err
	}
	log.Println(uaddr.String(), " has balance (wei): ", balance)

	//判断账户余额能否部署query合约 + upKeeping + channel合约
	var moneyPerDay = new(big.Int)
	moneyPerDay = moneyPerDay.Mul(big.NewInt(storePrice), big.NewInt(storeSize))
	var moneyAccount = new(big.Int)
	moneyAccount = moneyAccount.Mul(moneyPerDay, big.NewInt(storeDays))

	deployPrice := big.NewInt(int64(740621000000000))
	deployPrice.Add(big.NewInt(1128277), big.NewInt(int64(652346*ps)))
	var leastMoney = new(big.Int)
	leastMoney = leastMoney.Add(moneyAccount, deployPrice)
	if balance.Cmp(leastMoney) < 0 { //余额不足
		log.Println(uaddr.String(), " need more balance to start")
		return ErrBalance
	}

	// deploy query

	queryAddr, err := contracts.DeployQuery(uaddr, sk, storeSize, storeDays, storePrice, ks, ps, rdo)
	if err != nil {
		log.Println("fail to deploy query contract")
		return err
	}

	log.Println(uaddr.String(), "has new query: ", queryAddr.String())
	return nil
}

func deployUpKeepingAndChannel(userID, queryID, hexSk string, ks, ps []string, storeDays, storeSize, storePrice int64) error {
	localAddress, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}

	queryAddress, err := address.GetAddressFromID(queryID)
	if err != nil {
		return err
	}

	var keepers, providers []common.Address
	for _, keeper := range ks {
		keeperAddress, err := address.GetAddressFromID(keeper)
		if err != nil {
			return err
		}
		keepers = append(keepers, keeperAddress)
	}

	for _, provider := range ps {
		providerAddress, err := address.GetAddressFromID(provider)
		if err != nil {
			return err
		}
		providers = append(providers, providerAddress)
	}

	var moneyPerDay = new(big.Int)
	var moneyAccount = new(big.Int)
	moneyPerDay = moneyPerDay.Mul(big.NewInt(storePrice), big.NewInt(storeSize))
	moneyAccount = moneyAccount.Mul(moneyPerDay, big.NewInt(storeDays))

	log.Println("Begin to dploy upkeeping contract...")

	_, err = contracts.DeployUpkeeping(hexSk, localAddress, queryAddress, keepers, providers, storeDays, storeSize, storePrice, moneyAccount, true)
	if err != nil {
		return err
	}

	err = contracts.SetQueryCompleted(hexSk, queryAddress)
	if err != nil {
		return err
	}

	//依次与各provider签署channel合约
	timeOut := big.NewInt(int64(storeDays * 24 * 60 * 60)) //秒，存储时间
	var moneyToChannel = new(big.Int)
	moneyToChannel = moneyToChannel.Mul(big.NewInt(storeSize), big.NewInt(int64(utils.READPRICEPERMB))) //暂定往每个channel合约中存储金额为：存储大小 x 每MB单价

	log.Println("Begin to deploy channel contract...")

	var wg sync.WaitGroup
	for _, proAddr := range providers {
		wg.Add(1)
		providerAddr := proAddr
		go func() {
			defer wg.Done()
			_, err := contracts.DeployChannelContract(hexSk, localAddress, queryAddress, providerAddr, timeOut, moneyToChannel, true)
			if err != nil {
				return
			}
		}()
	}
	wg.Wait()
	log.Println("user has deployed all contracts successfully!")
	return nil
}

func (g *groupInfo) saveContracts() error {
	userAddr, err := address.GetAddressFromID(g.owner)
	if err != nil {
		return err
	}

	queryAddr, err := address.GetAddressFromID(g.groupID)
	if err != nil {
		return err
	}

	if g.queryItem == nil {
		g.queryItem = getQuery(userAddr)
	}

	if g.upKeepingItem == nil {
		g.upKeepingItem = getUpKeeping(g.owner, g.groupID)
	}

	ctx := context.Background()

	for _, proInfo := range g.providers {
		proID := proInfo.providerID

		proAddr, err := address.GetAddressFromID(proID)
		if err != nil {
			return err
		}

		if proInfo.offerItem == nil {
			proInfo.offerItem = getOffer(userAddr, proAddr)
		}

		if proInfo.chanItem == nil {
			item, err := contracts.GetChannelInfo(userAddr, userAddr, proAddr, queryAddr)
			if err != nil {
				return err
			}

			// 先从本地找value
			var value = new(big.Int)
			km, err := metainfo.NewKeyMeta(item.ChannelAddr, metainfo.Channel)
			if err != nil {
				return nil
			}

			valueByte, err := g.ds.GetKey(ctx, km.ToString(), "local")
			if err != nil {
				log.Println("try to get channel value from: ", proID)
				valueByte, err = g.ds.GetKey(ctx, km.ToString(), proID)
				if err != nil {
					log.Println("Set channel price to 0.")
					value = big.NewInt(0)
				} else {
					value, _ = new(big.Int).SetString(string(valueByte), 10)
				}
			} else {
				value, _ = new(big.Int).SetString(string(valueByte), 10)
			}
			log.Println("保存在内存中的channel地址和value为:", item.ChannelAddr, value.String())
			item.Value = value
			proInfo.chanItem = &item
		}
	}

	return nil
}

func (g *groupInfo) getChannel(proID string) *contracts.ChannelItem {
	userAddr, err := address.GetAddressFromID(g.owner)
	if err != nil {
		return nil
	}

	queryAddr, err := address.GetAddressFromID(g.groupID)
	if err != nil {
		return nil
	}

	ctx := context.Background()

	for _, proInfo := range g.providers {
		pro := proInfo.providerID

		if pro != proID {
			continue
		}

		if proInfo.chanItem == nil {
			proAddr, err := address.GetAddressFromID(pro)
			if err != nil {
				return nil
			}

			item, err := contracts.GetChannelInfo(userAddr, userAddr, proAddr, queryAddr)
			if err != nil {
				return nil
			}

			// 先从本地找value
			var value = new(big.Int)
			km, err := metainfo.NewKeyMeta(item.ChannelAddr, metainfo.Channel)
			if err != nil {
				return nil
			}

			valueByte, err := g.ds.GetKey(ctx, km.ToString(), "local")
			if err != nil {
				log.Println("try to get channel value from: ", proID)
				valueByte, err = g.ds.GetKey(ctx, km.ToString(), proID)
				if err != nil {
					log.Println("Set channel price to 0.")
					value = big.NewInt(0)
				} else {
					value, _ = new(big.Int).SetString(string(valueByte), 10)
				}
			} else {
				value, _ = new(big.Int).SetString(string(valueByte), 10)
			}
			log.Println("channel addr and value are:", item.ChannelAddr, value.String())
			item.Value = value
			proInfo.chanItem = &item
		}
		return proInfo.chanItem
	}

	return nil
}

func getQuery(userAddr common.Address) *contracts.QueryItem {
	queryAddr, err := contracts.GetMarketAddr(userAddr, userAddr, contracts.Query)
	if err != nil {
		return nil
	}
	item, err := contracts.GetQueryInfo(userAddr, queryAddr)
	if err != nil {
		return nil
	}
	item.QueryAddr = queryAddr.String()

	return &item
}

func getUpKeeping(userID, groupID string) *contracts.UpKeepingItem {
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return nil
	}

	queryAddr, err := address.GetAddressFromID(groupID)
	if err != nil {
		return nil
	}

	ukAddr, uk, err := contracts.GetUpkeeping(userAddr, userAddr, queryAddr.String())
	if err != nil {
		return nil
	}

	item, err := contracts.GetUpkeepingInfo(userAddr, uk)
	if err != nil {
		return nil
	}
	item.UpKeepingAddr = ukAddr
	return &item
}

func getOffer(userAddr, proAddr common.Address) *contracts.OfferItem {
	offerAddr, err := contracts.GetMarketAddr(userAddr, proAddr, contracts.Offer)
	if err != nil {
		log.Println("get", proAddr.String(), "'s offer address err ")
		return nil
	}
	item, err := contracts.GetOfferInfo(userAddr, offerAddr)
	if err != nil {
		log.Println("get", proAddr.String(), "'s offer params err ")
		return nil
	}
	return &item
}
