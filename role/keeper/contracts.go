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

		flag := false
		for _, keeperID := range uItem.KeeperIDs {
			if g.localKeeper == keeperID {
				flag = true
			}
		}

		// not my user
		if !flag {
			utils.MLogger.Warnf("user %s 's fsID %s not my user", g.userID, g.groupID)
			return role.ErrNotMyUser
		}

		g.providers = uItem.ProviderIDs
		g.keepers = uItem.KeeperIDs

		g.upkeeping = &uItem
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
		err = contracts.AddProvider(k.sk, localAddr, userAddr, ukAddr, []common.Address{providerAddr})
		if err != nil {
			utils.MLogger.Error("ukAddProvider AddProvider error", err)
			return err
		}
	}

	// update uk info
	gp.loadContracts(true)

	return nil
}

func (k *Info) getKpMapRegular(ctx context.Context) {
	utils.MLogger.Info("Get kpMap from chain start!")

	peerID := k.localID
	role.SaveKpMap(peerID)
	ticker := time.NewTicker(kpMapTime)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go func() {
				role.SaveKpMap(peerID)
			}()
		}
	}
}
