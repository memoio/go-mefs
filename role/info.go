package role

import (
	"math/big"

	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/repo/fsrepo"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

// GetAllKeepers gets all keepers from keeper-contract
func GetAllKeepers(localID string) ([]*KeeperItem, *big.Int, error) {
	totalMoney := new(big.Int)
	localAddr, err := address.GetAddressFromID(localID)
	if err != nil {
		return nil, totalMoney, err
	}

	u := contracts.NewCR(localAddr, "")
	kaddrs, err := u.GetAllKeepersAddr()
	if err != nil {
		return nil, totalMoney, err
	}

	kItems := make([]*KeeperItem, 0, len(kaddrs))
	for _, kaddr := range kaddrs {
		keeperID, err := address.GetIDFromAddress(kaddr.Hex())
		if err != nil {
			return nil, totalMoney, err
		}
		item, err := GetKeeperInfo(localID, keeperID)
		if err != nil {
			return nil, totalMoney, err
		}
		kItems = append(kItems, &item)
		totalMoney.Add(totalMoney, item.PledgeMoney)
	}

	if len(kItems) > 0 {
		return kItems, totalMoney, nil
	}
	return kItems, totalMoney, ErrEmptyData
}

// GetAllProviders gets all providers from provider-contract and total storage
func GetAllProviders(localID string) ([]*ProviderItem, *big.Int, error) {
	totalMoney := new(big.Int)

	localAddr, err := address.GetAddressFromID(localID)
	if err != nil {
		return nil, totalMoney, err
	}

	u := contracts.NewCR(localAddr, "")
	paddrs, err := u.GetAllProvidersAddr()
	if err != nil {
		return nil, totalMoney, err
	}

	price, err := u.GetProviderPrice()
	if err != nil {
		return nil, totalMoney, err
	}

	weiPrice := new(big.Float).SetInt(price)
	weiPrice.Quo(weiPrice, contracts.GetMemoPrice())
	weiPrice.Int(price)
	if price.Sign() <= 0 {
		return nil, nil, ErrInvalidInput
	}

	pItems := make([]*ProviderItem, 0, len(paddrs))
	for _, paddr := range paddrs {
		providerID, err := address.GetIDFromAddress(paddr.Hex())
		if err != nil {
			return nil, totalMoney, err
		}
		item, err := GetProviderInfo(localID, providerID)
		if err != nil {
			continue
		}
		pItems = append(pItems, &item)
		totalMoney.Add(totalMoney, item.PledgeMoney)
	}

	if len(pItems) > 0 && price.Cmp(big.NewInt(0)) > 0 {
		totalMoney.Div(totalMoney, price)
		return pItems, totalMoney, nil
	}
	return pItems, totalMoney, ErrEmptyData
}

// GetDiskSpaceInfo gets local storage info
func GetDiskSpaceInfo() (*utils.DiskStats, error) {
	rootpath, err := fsrepo.BestKnownPath()
	if err != nil {
		return nil, err
	}

	dinfo, err := utils.DiskStatus(rootpath)
	if err != nil {
		return nil, err
	}

	return dinfo, nil
}
