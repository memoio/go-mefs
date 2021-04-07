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

//AdminOwnedInfo  The basic information of node used for 'adminOwned' contract
type AdminOwnedInfo struct {
	addr  common.Address //local address
	hexSk string         //local privateKey
}

//NewCA new a instance of contractAdminOwned
func NewCA(addr common.Address, hexSk string) ContractAdminOwned {
	AInfo := &AdminOwnedInfo{
		addr:  addr,
		hexSk: hexSk,
	}

	return AInfo
}

//DeployAdminOwned deploy an AdminOwned contract
func (a *AdminOwnedInfo) DeployAdminOwned() (common.Address, error) {
	var adminOwnedAddr common.Address
	client := getClient(EndPoint)
	auth, err := makeAuth(a.hexSk, nil, nil, big.NewInt(defaultGasPrice), 0)
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
func (a *AdminOwnedInfo) GetAdminOwner(adminOwnedAddress common.Address) (common.Address, error) {
	client := getClient(EndPoint)
	adminOwnedContract, err := adminOwned.NewAdminOwned(adminOwnedAddress, client)
	if err != nil {
		log.Println("getAdminOwnedErr:", err)
		return adminOwnedAddress, err
	}

	adminOwner, err := adminOwnedContract.GetAdminOwner(&bind.CallOpts{
		From: a.addr,
	})
	if err != nil {
		return adminOwner, err
	}

	return adminOwner, nil
}

//AlterOwner alter the owner of AdminOwned-contract
func (a *AdminOwnedInfo) AlterOwner(adminOwnedAddress, newAdminOwner common.Address) error {
	client := getClient(EndPoint)
	adminOwnedContract, err := adminOwned.NewAdminOwned(adminOwnedAddress, client)
	if err != nil {
		log.Println("getAdminOwnedErr:", err)
		return err
	}

	auth, err := makeAuth(a.hexSk, nil, nil, big.NewInt(defaultGasPrice), 0)
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
			adminOwner, err := a.GetAdminOwner(adminOwnedAddress)
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
func (a *AdminOwnedInfo) SetBannedVersion(key string, adminOwnedAddress common.Address, version uint16) error {
	client := getClient(EndPoint)
	adminOwnedContract, err := adminOwned.NewAdminOwned(adminOwnedAddress, client)
	if err != nil {
		log.Println("getAdminOwnedErr:", err)
		return err
	}

	auth, err := makeAuth(a.hexSk, nil, nil, big.NewInt(defaultGasPrice), 0)
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

	err = checkTx(tx)

	//解析日志
	receipt := getTransactionReceipt(tx.Hash())

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
func (a *AdminOwnedInfo) GetBannedVersion(key string, adminOwnedAddress common.Address) (uint16, error) {
	client := getClient(EndPoint)
	adminOwnedContract, err := adminOwned.NewAdminOwned(adminOwnedAddress, client)
	if err != nil {
		log.Println("getAdminOwnedErr:", err)
		return 0, err
	}

	var bannedVersion uint16

	switch key {
	case "channel":
		bannedVersion, err = adminOwnedContract.GetChannelBannedVersion(&bind.CallOpts{
			From: a.addr,
		})
	case "mapper":
		bannedVersion, err = adminOwnedContract.GetMapperBannedVersion(&bind.CallOpts{
			From: a.addr,
		})
	case "query":
		bannedVersion, err = adminOwnedContract.GetQueryBannedVersion(&bind.CallOpts{
			From: a.addr,
		})
	case "offer":
		bannedVersion, err = adminOwnedContract.GetOfferBannedVersion(&bind.CallOpts{
			From: a.addr,
		})
	case "upkeeping":
		bannedVersion, err = adminOwnedContract.GetUpkeepingBannedVersion(&bind.CallOpts{
			From: a.addr,
		})
	case "root":
		bannedVersion, err = adminOwnedContract.GetRootBannedVersion(&bind.CallOpts{
			From: a.addr,
		})
	case "keeper":
		bannedVersion, err = adminOwnedContract.GetKeeperBannedVersion(&bind.CallOpts{
			From: a.addr,
		})
	case "provider":
		bannedVersion, err = adminOwnedContract.GetProviderBannedVersion(&bind.CallOpts{
			From: a.addr,
		})
	case "kpMap":
		bannedVersion, err = adminOwnedContract.GetKPMapBannedVersion(&bind.CallOpts{
			From: a.addr,
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
