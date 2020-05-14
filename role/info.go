package role

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

// GetAllKeepers gets all
func GetAllKeepers(localID string) ([]*KeeperItem, *big.Int, error) {
	totalMoney := new(big.Int)
	localAddress, err := address.GetAddressFromID(localID)
	if err != nil {
		return nil, totalMoney, err
	}

	kaddrs, err := contracts.GetAllKeepers(localAddress)
	if err != nil {
		return nil, totalMoney, err
	}

	_, keeperInstance, err := contracts.GetKeeperContractFromIndexer(localAddress)
	if err != nil {
		return nil, totalMoney, err
	}

	kItems := make([]*KeeperItem, 0, len(kaddrs))
	for _, kaddr := range kaddrs {
		isKeeper, isBanned, money, _, err := keeperInstance.Info(&bind.CallOpts{From: localAddress}, kaddr)
		if err != nil {
			continue
		}
		if isKeeper && !isBanned {
			keeperID, err := address.GetIDFromAddress(kaddr.String())
			if err != nil {
				continue
			}

			item := &KeeperItem{
				KeeperID: keeperID,
				Money:    money,
			}
			kItems = append(kItems, item)
			totalMoney.Add(totalMoney, money)
		}
	}

	if len(kItems) > 0 {
		return kItems, totalMoney, nil
	}
	return kItems, totalMoney, ErrEmptyData
}

// GetAllProviders gets all providers and total storage
func GetAllProviders(localID string) ([]*ProviderItem, *big.Int, error) {
	totalMoney := new(big.Int)
	localAddress, err := address.GetAddressFromID(localID)
	if err != nil {
		return nil, totalMoney, err
	}

	paddrs, err := contracts.GetAllProviders(localAddress)
	if err != nil {
		return nil, totalMoney, err
	}

	_, proInstance, err := contracts.GetProviderContractFromIndexer(localAddress)
	if err != nil {
		return nil, totalMoney, err
	}

	price, err := proInstance.GetPrice(&bind.CallOpts{From: localAddress})
	if err != nil {
		return nil, totalMoney, err
	}

	pItems := make([]*ProviderItem, 0, len(paddrs))
	for _, paddr := range paddrs {
		isProvider, isBanned, money, _, err := proInstance.Info(&bind.CallOpts{From: localAddress}, paddr)
		if err != nil {
			continue
		}
		if isProvider && !isBanned {
			proID, err := address.GetIDFromAddress(paddr.String())
			if err != nil {
				continue
			}

			item := &ProviderItem{
				ProviderID: proID,
				Money:      money,
			}
			pItems = append(pItems, item)
			totalMoney.Add(totalMoney, money)
		}
	}

	if len(pItems) > 0 && price.Cmp(big.NewInt(0)) > 0 {
		totalMoney.Div(totalMoney, price)
		return pItems, totalMoney, nil
	}
	return pItems, totalMoney, ErrEmptyData
}

// GetMemoPrice gets memo price
func GetMemoPrice() *big.Float {
	return big.NewFloat(utils.Memo2Dollar)
}
