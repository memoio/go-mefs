package contracts

import (
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts/recover"
)

//DeployRecover deploy a recover contract
func DeployRecover(hexKey string) (common.Address, error) {
	var recoverAddr common.Address
	client := GetClient(EndPoint)
	auth, err := MakeAuth(hexKey, nil, nil, big.NewInt(defaultGasPrice), 0)
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
