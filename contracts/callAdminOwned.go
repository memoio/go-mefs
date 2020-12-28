package contracts

import (
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/memoio/go-mefs/contracts/adminOwned"
)

//DeployAdminOwned deploy an AdminOwned contract
func DeployAdminOwned(hexKey string) (common.Address, error) {
	var adminOwnedAddr common.Address
	client := GetClient(EndPoint)
	auth, err := MakeAuth(hexKey, nil, nil, big.NewInt(defaultGasPrice), 0)
	if err != nil {
		return adminOwnedAddr, err
	}

	adminOwnedAddr, _, _, err = adminOwned.DeployAdminOwned(auth, client)
	if err != nil {
		log.Println("deployAdminOwnedErr:", err)
		return adminOwnedAddr, err
	}
	log.Println("adminOwnedContractAddr:", adminOwnedAddr.String())
	return adminOwnedAddr, nil
}

//GetAdminOwner get owner of the adminOwned-contract
func GetAdminOwner(adminOwnedAddress, localAddress common.Address) (common.Address, error) {
	client := GetClient(EndPoint)
	adminOwnedContract, err := adminOwned.NewAdminOwned(adminOwnedAddress, client)
	if err != nil {
		log.Println("getAdminOwnedErr:", err)
		return localAddress, err
	}

	adminOwner, err := adminOwnedContract.GetAdminOwner(&bind.CallOpts{
		From: localAddress,
	})
	if err != nil {
		return adminOwner, err
	}

	return adminOwner, nil
}

//AlterOwner alter the owner of AdminOwned-contract
func AlterOwner(hexKey string, adminOwnedAddress, newAdminOwner common.Address) error {
	client := GetClient(EndPoint)
	adminOwnedContract, err := adminOwned.NewAdminOwned(adminOwnedAddress, client)
	if err != nil {
		log.Println("getAdminOwnedErr:", err)
		return err
	}

	auth, err := MakeAuth(hexKey, nil, nil, big.NewInt(defaultGasPrice), 0)
	if err != nil {
		return err
	}

	_, err = adminOwnedContract.AlterAdminOwner(auth, newAdminOwner)
	if err != nil {
		log.Println("AlterAdminOwnerErr:", err)
		return err
	}

	return nil
}

//SetBanned set parameter represented by key
func SetBanned(hexKey, key string, adminOwnedAddress common.Address, banned bool) error {
	client := GetClient(EndPoint)
	adminOwnedContract, err := adminOwned.NewAdminOwned(adminOwnedAddress, client)
	if err != nil {
		log.Println("getAdminOwnedErr:", err)
		return err
	}

	auth, err := MakeAuth(hexKey, nil, nil, big.NewInt(defaultGasPrice), 0)
	if err != nil {
		return err
	}

	var tx *types.Transaction

	switch key {
	case "channel":
		tx, err = adminOwnedContract.SetChannelBanned(auth, banned)
	case "mapper":
		tx, err = adminOwnedContract.SetMapperBanned(auth, banned)
	case "query":
		tx, err = adminOwnedContract.SetQueryBanned(auth, banned)
	case "offer":
		tx, err = adminOwnedContract.SetOfferBanned(auth, banned)
	case "upkeeping":
		tx, err = adminOwnedContract.SetUpkeepingBanned(auth, banned)
	case "root":
		tx, err = adminOwnedContract.SetRootBanned(auth, banned)
	case "keeper":
		tx, err = adminOwnedContract.SetKeeperBanned(auth, banned)
	case "provider":
		tx, err = adminOwnedContract.SetProviderBanned(auth, banned)
	case "kpMap":
		tx, err = adminOwnedContract.SetKPMapBanned(auth, banned)
	default:
		log.Println("unsupported key")
		return nil
	}

	if err != nil {
		log.Println("set Banned error:", err)
		return err
	}

	err = CheckTx(tx)

	//解析日志
	receipt := GetTransactionReceipt(tx.Hash())

	contractABI, err := abi.JSON(strings.NewReader(string(adminOwned.AdminOwnedABI)))
	if err != nil {
		log.Println("get contract abi err:", err)
		return err
	}

	event := struct {
		Key   string
		From  common.Address
		Param bool
	}{}

	err = contractABI.Unpack(&event, "SetBanned", receipt.Logs[0].Data)
	if err != nil {
		log.Println("setBanned tx err:", err)
		return err
	}
	fmt.Println("Log.key:", event.Key, "Log.from:", event.From.String(), "Log.param:", event.Param)

	return nil
}

//GetBanned get parameter represented by key
func GetBanned(key string, adminOwnedAddress, localAddress common.Address) (bool, error) {
	client := GetClient(EndPoint)
	adminOwnedContract, err := adminOwned.NewAdminOwned(adminOwnedAddress, client)
	if err != nil {
		log.Println("getAdminOwnedErr:", err)
		return false, err
	}

	var banned bool

	switch key {
	case "channel":
		banned, err = adminOwnedContract.GetChannelBanned(&bind.CallOpts{
			From: localAddress,
		})
	case "mapper":
		banned, err = adminOwnedContract.GetMapperBanned(&bind.CallOpts{
			From: localAddress,
		})
	case "query":
		banned, err = adminOwnedContract.GetQueryBanned(&bind.CallOpts{
			From: localAddress,
		})
	case "offer":
		banned, err = adminOwnedContract.GetOfferBanned(&bind.CallOpts{
			From: localAddress,
		})
	case "upkeeping":
		banned, err = adminOwnedContract.GetUpkeepingBanned(&bind.CallOpts{
			From: localAddress,
		})
	case "root":
		banned, err = adminOwnedContract.GetRootBanned(&bind.CallOpts{
			From: localAddress,
		})
	case "keeper":
		banned, err = adminOwnedContract.GetKeeperBanned(&bind.CallOpts{
			From: localAddress,
		})
	case "provider":
		banned, err = adminOwnedContract.GetProviderBanned(&bind.CallOpts{
			From: localAddress,
		})
	case "kpMap":
		banned, err = adminOwnedContract.GetKPMapBanned(&bind.CallOpts{
			From: localAddress,
		})
	default:
		log.Println("unsupported key")
	}

	if err != nil {
		log.Println("get Banned error:", err)
		return false, err
	}
	return banned, nil
}
