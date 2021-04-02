package contracts

import (
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

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

	for retryTimes := 0; retryTimes < sendTransactionRetryCount; retryTimes++ {
		_, err = adminOwnedContract.AlterAdminOwner(auth, newAdminOwner)
		if err != nil {
			log.Println("AlterAdminOwnerErr:", err)
			return err
		}

		//check if the tx has been completed by inquiring contract state variables
		for checkTimes := 0; checkTimes < checkTxRetryCount; checkTimes++ {
			adminOwner, err := GetAdminOwner(adminOwnedAddress, newAdminOwner)
			if err != nil || adminOwner.Hex() != newAdminOwner.Hex() {
				time.Sleep(retryGetInfoSleepTime)
				continue
			}
			return nil
		}

		time.Sleep(retryTxSleepTime)
	}
	log.Println("AlterAdminOwnerErr: ", ErrNotRight)

	return ErrNotRight
}

//SetBannedVersion set bannedVersion represented by key
func SetBannedVersion(hexKey, key string, adminOwnedAddress common.Address, version uint16) error {
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
		tx, err = adminOwnedContract.SetChannelBannedVersion(auth, version)
	case "mapper":
		tx, err = adminOwnedContract.SetMapperBannedVersion(auth, version)
	case "query":
		tx, err = adminOwnedContract.SetQueryBannedVersion(auth, version)
	case "offer":
		tx, err = adminOwnedContract.SetOfferBannedVersion(auth, version)
	case "upkeeping":
		tx, err = adminOwnedContract.SetUpkeepingBannedVersion(auth, version)
	case "root":
		tx, err = adminOwnedContract.SetRootBannedVersion(auth, version)
	case "keeper":
		tx, err = adminOwnedContract.SetKeeperBannedVersion(auth, version)
	case "provider":
		tx, err = adminOwnedContract.SetProviderBannedVersion(auth, version)
	case "kpMap":
		tx, err = adminOwnedContract.SetKPMapBannedVersion(auth, version)
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
		Key     string
		From    common.Address
		Version uint16
	}{}

	err = contractABI.Unpack(&event, "SetBanned", receipt.Logs[0].Data)
	if err != nil {
		log.Println("setBanned tx err:", err)
		return err
	}
	fmt.Println("Log.key:", event.Key, "Log.from:", event.From.String(), "Log.version:", event.Version)

	if event.Version != version {
		fmt.Println("tx failed, the version is not right")
	}
	return nil
}

//GetBannedVersion get bannedVersion represented by key
func GetBannedVersion(key string, adminOwnedAddress, localAddress common.Address) (uint16, error) {
	client := GetClient(EndPoint)
	adminOwnedContract, err := adminOwned.NewAdminOwned(adminOwnedAddress, client)
	if err != nil {
		log.Println("getAdminOwnedErr:", err)
		return 0, err
	}

	var bannedVersion uint16

	switch key {
	case "channel":
		bannedVersion, err = adminOwnedContract.GetChannelBannedVersion(&bind.CallOpts{
			From: localAddress,
		})
	case "mapper":
		bannedVersion, err = adminOwnedContract.GetMapperBannedVersion(&bind.CallOpts{
			From: localAddress,
		})
	case "query":
		bannedVersion, err = adminOwnedContract.GetQueryBannedVersion(&bind.CallOpts{
			From: localAddress,
		})
	case "offer":
		bannedVersion, err = adminOwnedContract.GetOfferBannedVersion(&bind.CallOpts{
			From: localAddress,
		})
	case "upkeeping":
		bannedVersion, err = adminOwnedContract.GetUpkeepingBannedVersion(&bind.CallOpts{
			From: localAddress,
		})
	case "root":
		bannedVersion, err = adminOwnedContract.GetRootBannedVersion(&bind.CallOpts{
			From: localAddress,
		})
	case "keeper":
		bannedVersion, err = adminOwnedContract.GetKeeperBannedVersion(&bind.CallOpts{
			From: localAddress,
		})
	case "provider":
		bannedVersion, err = adminOwnedContract.GetProviderBannedVersion(&bind.CallOpts{
			From: localAddress,
		})
	case "kpMap":
		bannedVersion, err = adminOwnedContract.GetKPMapBannedVersion(&bind.CallOpts{
			From: localAddress,
		})
	default:
		log.Println("unsupported key")
	}

	if err != nil {
		log.Println("get Banned error:", err)
		return 0, err
	}
	return bannedVersion, nil
}
