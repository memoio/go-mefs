package provider

import (
	"errors"
	"log"

	"github.com/ethereum/go-ethereum/common"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/core"
	fr "github.com/memoio/go-mefs/repo/fsrepo"
	ad "github.com/memoio/go-mefs/utils/address"
)

var errBalance = errors.New("your account's balance is insufficient, we will not deploy resolver")

func getParamsForDeployContract(node *core.MefsNode) (hexPK string, localAddress common.Address, err error) {

	//得到部署resolver的provider的地址
	id := peer.IDB58Encode(node.Identity)
	localAddress, err = ad.GetAddressFromID(id)
	if err != nil {
		log.Println("getLocalAddr err: ", err)
		return "", localAddress, err
	}

	//得到部署resolver的provider的私钥
	hexPK, err = fr.GetHexPrivKeyFromKS(node.Identity, node.Password)
	if err != nil {
		log.Println("getHexPK err: ", err)
		return "", localAddress, err
	}

	return hexPK, localAddress, nil
}

func providerDeployResolverAndOffer(node *core.MefsNode, capacity int64, duration int64, price int64, reDeployOffer bool) error {
	//得到部署resolver所需的参数
	hexPK, localAddress, err := getParamsForDeployContract(node)
	if err != nil {
		return err
	}

	//获得用户的账户余额
	balance, _ := contracts.QueryBalance(localAddress.Hex())
	log.Println("balance is: ", balance)
	//先部署resolver-for-channel
	//如果部署过resolver-for-channel，那接下来就可以直接检查是否部署过offer合约，没有的话就部署
	//DeployResolver()函数内部会进行判断是否部署过
	_, err = contracts.DeployResolverForChannel(localAddress, hexPK)
	if err != nil {
		return err
	}

	//获得用户的账户余额
	balance, _ = contracts.QueryBalance(localAddress.Hex())
	log.Println("after deploying resolver for channel, balance is: ", balance)
	//再开始部署offer合约
	if reDeployOffer { //用户想要重新部署offer合约
		log.Println("provider wants to redeploy offer-contract")
	}
	_, err = contracts.DeployOffer(localAddress, hexPK, capacity, duration, price, reDeployOffer)
	if err != nil {
		return err
	}

	return nil
}
