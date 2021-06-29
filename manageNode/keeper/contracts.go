package keeper

import (
	"context"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	id "github.com/memoio/go-mefs/crypto/identity"
	mpb "github.com/memoio/go-mefs/pb"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
)

// force update if mode is set true
func (g *groupInfo) loadContracts(mode bool) error {
	if g.groupID == g.userID {
		return nil
	}

	// get upkkeeping addr
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

		if uItem.EndTime < time.Now().Unix() {
			// not load expire uk
			return role.ErrUkExpire
		}

		var keepers []string
		var providers []string
		for _, keeper := range uItem.Keepers {
			if !keeper.Stop {
				kid, err := address.GetIDFromAddress(keeper.Addr.String())
				if err != nil {
					return err
				}
				keepers = append(keepers, kid)
			}
		}

		for _, provider := range uItem.Providers {
			if !provider.Stop {
				pid, err := address.GetIDFromAddress(provider.Addr.String())
				if err != nil {
					return err
				}
				providers = append(providers, pid)
			}
		}

		flag := false
		for _, keeperID := range keepers {
			if g.localKeeper == keeperID {
				flag = true
			}
		}

		// not my user
		if !flag {
			utils.MLogger.Warnf("user %s 's fsID %s not my user", g.userID, g.groupID)
			return role.ErrNotMyUser
		}

		g.providers = providers
		g.keepers = keepers

		g.upkeeping = &uItem
	}

	if g.rootID == "" || mode {
		rootID, err := role.GetRoot(g.userID, g.groupID)
		if err != nil {
			return err
		}
		g.rootID = rootID
	}

	return nil
}

func (k *Info) loadContract(mode bool) error {
	if k.kItem == nil || mode {
		kItem, err := role.GetKeeperInfo(k.localID, k.localID)
		if err != nil {
			return err
		}

		localAddr, _ := address.GetAddressFromID(k.localID)
		r := contracts.NewCR(localAddr, "")
		if err != nil {
			return err
		}

		price, err := r.GetKeeperPrice()
		if err != nil {
			return err
		}

		if kItem.PledgeMoney.Cmp(price) < 0 {
			price.Sub(price, kItem.PledgeMoney)
			utils.MLogger.Infof("pledge keeper %s amount %d", k.localID, price)
			r := contracts.NewCR(localAddr, k.sk)
			err := r.PledgeKeeper(price)
			if err != nil {
				return err
			}
		}

		kItem, err = role.GetKeeperInfo(k.localID, k.localID)
		if err != nil {
			return err
		}

		k.kItem = &kItem
	}
	return nil
}

// addProvider 将传入pid加入posuser的upkeeping合约
func (k *Info) ukAddProvider(uid, gid, pid string) error {
	gp := k.getGroupInfo(uid, gid, true)
	if gp == nil || gp.upkeeping == nil {
		return role.ErrNoContract
	}

	providerAddr, err := address.GetAddressFromID(pid)
	if err != nil {
		return err
	}

	// check pid belongs to this group or not
	for _, gpid := range gp.providers {
		if gpid == pid {
			return nil
		}
	}

	localAddr, err := address.GetAddressFromID(k.localID)
	if err != nil {
		return err
	}

	userAddr, err := address.GetAddressFromID(uid)
	if err != nil {
		return err
	}

	ukAddr, err := address.GetAddressFromID(gp.upkeeping.UpKeepingID)
	if err != nil {
		return err
	}

	if gp.isMaster(pid) {
		utils.MLogger.Info("add provider to: ", userAddr)
		mkey, err := metainfo.NewKey(gp.groupID, mpb.KeyType_ProAddSign, gp.userID, pid)
		if err != nil {
			return err
		}

		key := mkey.ToString()

		sHash, err := role.GetHashForAddProvider(ukAddr, []common.Address{providerAddr})
		if err != nil {
			return err
		}

		sig, err := id.Sign(k.sk, sHash)
		if err != nil {
			return err
		}

		keepers := gp.keepers
		sigs := make([][]byte, len(keepers))
		for i, kid := range keepers {
			if kid == k.localID {
				sigs[i] = sig
				continue
			}
			res, err := k.ds.SendMetaRequest(k.context, int32(mpb.OpType_Get), key, sHash, sig, kid)
			if err != nil {
				return err
			}
			sigs[i] = res
		}

		cu := contracts.NewCU(localAddr, k.sk)
		err = cu.AddProvider(ukAddr, []common.Address{providerAddr}, sigs)
		if err != nil {
			utils.MLogger.Error("ukAddProvider AddProvider error", err)
			return err
		}
	}

	// update uk info
	gp.loadContracts(true)

	return nil
}

//getFromChainRegular loadPeers from keeper/provider/kpMap contract and getIncome
func (k *Info) getFromChainRegular(ctx context.Context) {
	utils.MLogger.Info("Get infos from chain start!")

	localAddr, err := address.GetAddressFromID(k.localID)
	if err != nil {
		return
	}

	lastBlock := int64(0)

	km, err := metainfo.NewKey(k.localID, mpb.KeyType_Income)
	if err == nil {
		res, err := k.ds.GetKey(k.context, km.ToString(), "local")
		if err == nil && len(res) > 0 {
			utils.MLogger.Infof("Load %s income info: %s", km.ToString(), string(res))
			ins := strings.Split(string(res), metainfo.DELIMITER)
			if len(ins) == 4 {
				lb, err := strconv.ParseInt(ins[0], 10, 0)
				if err == nil {
					lastBlock = lb
				}

				mi, ok := new(big.Int).SetString(ins[1], 10)
				if ok {
					k.ManageIncome = mi
				}

				pi, ok := new(big.Int).SetString(ins[2], 10)
				if ok {
					k.PosIncome = pi
				}

				prei, ok := new(big.Int).SetString(ins[3], 10)
				if ok {
					k.PosPreIncome = prei
				}
			}
		}
	}

	k.loadPeersFromChain()

	time.Sleep(5 * time.Minute)
	lb, err := k.getIncome(localAddr, lastBlock)
	if err == nil {
		lastBlock = lb
	}

	ticker := time.NewTicker(kpMapTime)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			k.loadPeersFromChain()
			lb, err := k.getIncome(localAddr, lastBlock)
			if err == nil {
				lastBlock = lb
			}
		}
	}
}

//getIncome get keeper's income from user
func (k *Info) getIncome(localAddr common.Address, pBlock int64) (int64, error) {
	a := contracts.NewCA(localAddr, "")
	b, err := a.GetLatestBlock()
	if err != nil {
		return 0, err
	}

	latestBlock := b.Number().Int64()
	endBlock := b.Number().Int64()
	ukaddrs, posAddrs := k.GetIncomeAddress()

	startBlock := pBlock
	if len(ukaddrs) > 0 && latestBlock > startBlock {
		utils.MLogger.Infof("get manage income from chain")
		endBlock = latestBlock

		for endBlock <= latestBlock {
			if endBlock > startBlock+1024 {
				endBlock = startBlock + 1024
			}

			mIncome, _, err := a.GetStorageIncome(ukaddrs, localAddr, startBlock, endBlock)
			if err != nil {
				utils.MLogger.Info("get ukpay log err:", err)
				break
			}

			k.ManageIncome.Add(k.ManageIncome, mIncome)
			startBlock = endBlock

			if endBlock == latestBlock {
				break
			}

			if endBlock < latestBlock {
				endBlock = latestBlock
			}
		}
	}

	posStartBlock := pBlock
	if len(posAddrs) > 0 && latestBlock > posStartBlock {
		utils.MLogger.Infof("get post income from chain")

		endBlock = latestBlock

		for endBlock <= latestBlock {
			if endBlock > posStartBlock+1024 {
				endBlock = posStartBlock + 1024
			}

			posMIncome, _, err := a.GetStorageIncome(posAddrs, localAddr, posStartBlock, endBlock)
			if err != nil {
				utils.MLogger.Info("get post ukpay log err:", err)
				break
			}

			k.PosIncome.Add(k.PosIncome, posMIncome)
			posStartBlock = endBlock

			if endBlock == latestBlock {
				break
			}

			if endBlock < latestBlock {
				endBlock = latestBlock
			}
		}
	}

	k.PosPreIncome = getPosPreIncome(posAddrs, localAddr)

	km, err := metainfo.NewKey(k.localID, mpb.KeyType_Income)
	if err == nil {
		var res strings.Builder
		res.WriteString(strconv.FormatInt(latestBlock, 10))
		res.WriteString(metainfo.DELIMITER)
		res.WriteString(k.ManageIncome.String())
		res.WriteString(metainfo.DELIMITER)
		res.WriteString(k.PosIncome.String())
		res.WriteString(metainfo.DELIMITER)
		res.WriteString(k.PosPreIncome.String())

		k.ds.PutKey(k.context, km.ToString(), []byte(res.String()), nil, "local")
	}

	utils.MLogger.Infof("get income from chain finished at block %d", latestBlock)
	return latestBlock, nil
}

//GetIncomeAddress get upkeepingAddress and posAddress of this keeper to filter logs in chain
func (k *Info) GetIncomeAddress() ([]common.Address, []common.Address) {
	ukAddr := []common.Address{}
	posAddr := []common.Address{}
	pus := k.getQUKeys()
	for _, pu := range pus {
		if pu.uid == pu.qid { //test
			continue
		}

		gp := k.getGroupInfo(pu.uid, pu.qid, false)
		if gp == nil || gp.upkeeping == nil {
			continue
		}

		tmp, err := address.GetAddressFromID(gp.upkeeping.UpKeepingID)
		if err != nil {
			continue
		}

		if pu.uid == pos.GetPosId() {
			posAddr = append(posAddr, tmp)
			continue
		} else {
			ukAddr = append(ukAddr, tmp)
		}
	}

	if len(posAddr) == 0 {
		qItem, err := role.GetLatestQuery(pos.GetPosId())
		if err != nil {
			return ukAddr, posAddr
		}
		uItem, err := role.GetUpKeeping(pos.GetPosId(), qItem.QueryID)
		if err != nil {
			return ukAddr, posAddr
		}
		localAddr, err := address.GetAddressFromID(k.localID)
		if err != nil {
			return ukAddr, posAddr
		}
		for _, ki := range uItem.Keepers {
			if ki.Addr.String() == localAddr.String() {
				uAddr, err := address.GetAddressFromID(uItem.UpKeepingID)
				if err != nil {
					return ukAddr, posAddr
				}
				posAddr = append(posAddr, uAddr)
			}
		}

	}

	return ukAddr, posAddr
}
