package keeper

import (
	"context"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

// force update if mode is set true
func (g *groupInfo) getContracts(mode bool) error {
	if g.groupID == g.userID {
		return errors.New("not deploying contracts")
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
			utils.MLogger.Info(g.userID, " is not my user")
			return errors.New("Not my user")
		}

		g.providers = uItem.ProviderIDs
		g.keepers = uItem.KeeperIDs

		g.upkeeping = &uItem
	}

	return nil
}

// addProvider 将传入pid加入posuser的upkeeping合约
func (k *Info) ukAddProvider(uid, gid, pid, sk string) error {
	gp := k.getGroupInfo(uid, gid, false)
	if gp == nil || gp.upkeeping == nil {
		return errors.New("No upkeeping")
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

	userAddr, err := address.GetAddressFromID(uid)
	if err != nil {
		return err
	}

	queryAddr, err := address.GetAddressFromID(gid)
	if err != nil {
		return err
	}

	if gp.isMaster(pid) {
		utils.MLogger.Info("add provider to: ", userAddr)
		err = contracts.AddProvider(sk, userAddr, userAddr, []common.Address{providerAddr}, queryAddr.String())
		if err != nil {
			utils.MLogger.Error("ukAddProvider AddProvider error", err)
			return err
		}
	}

	// update uk info
	gp.getContracts(true)

	return nil
}

func (k *Info) getKpMapRegular(ctx context.Context) {
	utils.MLogger.Info("Get kpMap from chain start!")

	peerID := k.localID
	role.SaveKpMap(peerID)
	ticker := time.NewTicker(KPMAPTIME)
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
