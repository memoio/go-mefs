package keeper

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
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

		price, err := role.GetKeeperPrice(k.localID)
		if err != nil {
			return err
		}

		if kItem.PledgeMoney.Cmp(price) < 0 {
			price.Sub(price, kItem.PledgeMoney)
			utils.MLogger.Infof("pledge keeper %s amount %d", k.localID, price)
			err := role.PledgeKeeper(k.localID, k.sk, price)
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

	//TODO
	sig := [][]byte{}
	if gp.isMaster(pid) {
		utils.MLogger.Info("add provider to: ", userAddr)
		err = contracts.AddProvider(k.sk, localAddr, userAddr, ukAddr, []common.Address{providerAddr}, sig)
		if err != nil {
			utils.MLogger.Error("ukAddProvider AddProvider error", err)
			return err
		}
	}

	// update uk info
	gp.loadContracts(true)

	return nil
}

func (k *Info) getFromChainRegular(ctx context.Context) {
	utils.MLogger.Info("Get infos from chain start!")
	ticker := time.NewTicker(kpMapTime)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			k.loadPeersFromChain()
		}
	}
}
