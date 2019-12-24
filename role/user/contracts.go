package user

import (
	"encoding/hex"
	"errors"
	"log"
	"math/big"
	"strconv"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func deployUpKeepingAndChannel(userID string, ks []*keeperInfo, ps []*providerInfo, storeDays int64, storeSize int64, storePrice int64) error {
	hexSk := utils.EthSkByteToEthString(getSk(userID))
	localAddress, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}

	var keepers, providers []common.Address
	for _, keeper := range ks {
		keeperAddress, err := address.GetAddressFromID(keeper.keeperID)
		if err != nil {
			return err
		}
		keepers = append(keepers, keeperAddress)
	}

	for _, provider := range ps {
		providerAddress, err := address.GetAddressFromID(provider.providerID)
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

	err = contracts.DeployUpkeeping(hexSk, localAddress, keepers, providers, storeDays, storeSize, storePrice, moneyAccount)
	if err != nil {
		return err
	}

	//部署好upKeeping合约后，将user部署的query合约的completed参数设为true
	queryAddr, err := contracts.GetMarketAddr(localAddress, localAddress, contracts.Query)
	if err != nil {
		return err
	}
	err = contracts.SetQueryCompleted(hexSk, queryAddr)
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
			channelAddr, err := contracts.DeployChannelContract(hexSk, localAddress, providerAddr, timeOut, moneyToChannel)
			if err != nil {
				return
			}
			//设置channel的value初始值为0
			//存到本地
			channelValueKeyMeta, err := metainfo.NewKeyMeta(channelAddr.String(), metainfo.Local, metainfo.SyncTypeChannelValue)
			if err != nil {
				return
			}
			key := channelValueKeyMeta.ToString() // hexChannelAddress|13|channelvalue
			err = putKeyTo(key, strconv.FormatInt(0, 10), "local")
			if err != nil {
				return
			}
			//存到provider上
			providerID, err := address.GetIDFromAddress(providerAddr.String())
			if err != nil {
				return
			}
			err = putKeyTo(key, strconv.FormatInt(0, 10), providerID)
			if err != nil {
				return
			}
		}()
	}
	wg.Wait()
	log.Println("user has deployed all contracts successfully!")
	return nil
}

func saveContracts(userID string) error {
	gp := getGroup(userID)
	if gp == nil {
		return errors.New("does not exist or has not started")
	}
	err := saveUpkeeping(userID, gp)
	if err != nil {
		return err
	}
	err = saveChannel(userID, gp)
	if err != nil {
		return err
	}
	err = saveQuery(userID, gp)
	if err != nil {
		return err
	}
	err = saveOffer(userID, gp)
	if err != nil {
		return err
	}
	return nil
}

func saveUpkeeping(userID string, gp *groupInfo) error {
	userAddr, err := address.GetAddressFromID(userID)
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
	item.UserID = userID
	item.UpKeepingAddr = ukAddr
	gp.upKeepingItem = &item
	return nil
}

func saveChannel(userID string, gp *groupInfo) error {
	if gp.upKeepingItem == nil {
		return errors.New("get upkeeing first")
	}

	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}

	for _, proInfo := range gp.providers {
		if proInfo.chanItem != nil {
			continue
		}
		proID := proInfo.providerID

		proAddr, err := address.GetAddressFromID(proID)
		if err != nil {
			return err
		}

		item, err := contracts.GetChannelInfo(userAddr, proAddr, userAddr)
		if err != nil {
			return err
		}

		// 先从本地找value
		var value = new(big.Int)
		channelValueKeyMeta, err := metainfo.NewKeyMeta(item.ChannelAddr, metainfo.Local, metainfo.SyncTypeChannelValue)
		if err != nil {
			return err
		}

		valueByte, err := getKeyFrom(channelValueKeyMeta.ToString(), "local")
		if err != nil {
			// 本地没找到，从provider上找
			log.Println("Can't get channel value in local,err :", err, ", so try to get from ", proID)
			valueByte, err = getKeyFrom(channelValueKeyMeta.ToString(), proID)
			if err != nil {
				// provider上也没找到，value设为0
				log.Println("Can't get channel price from ", proID, ",err :", err, ", so set channel price to 0.")
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
		log.Println("保存在内存中的channel地址和value为:", item.ChannelAddr, value.String())
		item.UserID = userID
		item.ProID = proID
		item.Value = value
		proInfo.chanItem = &item
	}
	return nil
}

func saveChannelValue(userID string) error {
	gp := getGroup(userID)
	if gp == nil {
		return errors.New("does not exist or has not started")
	}
	for _, proInfo := range gp.providers {
		if proInfo.chanItem != nil {
			// 保存本地形式：K-provider，V-channel此时的value
			km, err := metainfo.NewKeyMeta(proInfo.chanItem.ChannelAddr, metainfo.Local, metainfo.SyncTypeChannelValue)
			if err != nil {
				log.Println("NewKeyMeta err:", proInfo.providerID, err)
				continue
			}
			err = putKeyTo(km.ToString(), proInfo.chanItem.Value.String(), "local")
			if err != nil {
				log.Println("CmdPutTo error", proInfo.providerID, err)
				continue
			}
		}

	}
	return nil
}

func saveOffer(userID string, gp *groupInfo) error {
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}
	for _, proInfo := range gp.providers {
		if proInfo.offerItem != nil {
			continue
		}
		proID := proInfo.providerID
		proAddr, err := address.GetAddressFromID(proID)
		if err != nil {
			return err
		}

		offerAddr, err := contracts.GetMarketAddr(userAddr, proAddr, contracts.Offer)
		if err != nil {
			log.Println("get", proAddr.String(), "'s offer address err ")
			return err
		}
		item, err := contracts.GetOfferInfo(userAddr, offerAddr)
		if err != nil {
			log.Println("get", proAddr.String(), "'s offer params err ")
			return err
		}
		item.ProviderID = proID
		proInfo.offerItem = &item
	}
	return nil
}

func saveQuery(userID string, gp *groupInfo) error {
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return err
	}

	queryAddr, err := contracts.GetMarketAddr(userAddr, userAddr, contracts.Query)
	if err != nil {
		return err
	}
	item, err := contracts.GetQueryInfo(userAddr, queryAddr)
	if err != nil {
		return err
	}
	item.QueryAddr = queryAddr.String()
	gp.queryItem = &item
	return nil
}

func getUpkeepingItem(userID, proid string) (*contracts.UpKeepingItem, error) {
	gp := getGroup(userID)
	if gp == nil {
		return nil, errors.New("does not exist or has not started")
	}

	if gp.upKeepingItem != nil {
		return gp.upKeepingItem, nil
	}
	saveUpkeeping(userID, gp)
	return gp.upKeepingItem, nil
}

func getQueryItem(userID string) (*contracts.QueryItem, error) {
	gp := getGroup(userID)
	if gp == nil {
		return nil, errors.New("does not exist or has not started")
	}

	if gp.queryItem != nil {
		return gp.queryItem, nil
	}

	saveQuery(userID, gp)
	return gp.queryItem, nil
}

func getChannelItem(userID, proid string) (*contracts.ChannelItem, error) {
	gp := getGroup(userID)
	if gp == nil {
		return nil, errors.New("does not exist or has not started")
	}

	for _, proInfo := range gp.providers {
		if proInfo.providerID == proid {
			if proInfo.chanItem == nil {
				saveChannel(userID, gp)
			}
			return proInfo.chanItem, nil
		}
	}

	return nil, nil
}

func getOfferItem(userID, proid string) (*contracts.OfferItem, error) {
	gp := getGroup(userID)
	if gp == nil {
		return nil, errors.New("does not exist or has not started")
	}

	for _, proInfo := range gp.providers {
		if proInfo.providerID == proid {
			if proInfo.offerItem == nil {
				saveOffer(userID, gp)
			}
			return proInfo.offerItem, nil
		}
	}

	return nil, nil
}

func buildSignParams(userID string, providerID string, privateKey []byte) (common.Address, common.Address, string, error) {
	var userAddress, providerAddress common.Address
	providerAddress, err := address.GetAddressFromID(providerID)
	if err != nil {
		log.Println("GetProAddr err: ", err)
		return userAddress, providerAddress, "", err
	}

	userAddress, err = address.GetAddressFromID(userID)
	if err != nil {
		log.Println("GetLocalAddr err: ", err)
		return userAddress, providerAddress, "", err
	}

	pk := crypto.ToECDSAUnsafe(privateKey)
	pkByte := math.PaddedBigBytes(pk.D, pk.Params().BitSize/8)
	enc := make([]byte, len(pkByte)*2)
	hex.Encode(enc, pkByte)

	return userAddress, providerAddress, string(enc), nil
}
