package contracts

import (
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts/recover"
)

//DeployRecover deploy a recover contract
func (a *AdminOwnedInfo) DeployRecover() (common.Address, error) {
	var recoverAddr common.Address
	client := getClient(EndPoint)
	auth, err := makeAuth(a.hexSk, nil, nil, big.NewInt(defaultGasPrice), 0)
	if err != nil {
		return recoverAddr, err
	}

	recoverAddr, _, _, err = recover.DeployRecover(auth, client)
	if err != nil {
		log.Println("deployRecoverErr:", err)
		return recoverAddr, err
	}
	log.Println("recoverContractAddr:", recoverAddr.String())
	return recoverAddr, nil
}
