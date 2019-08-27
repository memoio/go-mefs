package provider

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/contracts/upKeeping"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/memoio/go-mefs/core"
	fr "github.com/memoio/go-mefs/repo/fsrepo"
	ad "github.com/memoio/go-mefs/utils/address"
)

var errBalance = errors.New("your account's balance is insufficient, we will not deploy resolver")

func getParamsForDeployContract(node *core.MefsNode) (endPoint string, hexPK string, localAddress common.Address, err error) {
	//得到endPoint
	config, err := node.Repo.Config()
	if err != nil {
		fmt.Println("getConfigErr", err)
		return "", "", localAddress, err
	}
	endPoint = config.Eth

	//得到部署resolver的provider的地址
	id := peer.IDB58Encode(node.Identity)
	localAddress, err = ad.GetAddressFromID(id)
	if err != nil {
		fmt.Println("getLocalAddrErr", err)
		return "", "", localAddress, err
	}

	//得到部署resolver的provider的私钥
	hexPK, err = fr.GetHexPrivKeyFromKS(node.Identity, node.Password)
	if err != nil {
		fmt.Println("getHexPKErr", err)
		return "", "", localAddress, err
	}

	return endPoint, hexPK, localAddress, nil
}

func providerDeployResolverAndOffer(node *core.MefsNode, capacity int64, duration int64, price int64, reDeployOffer bool) (err error) {
	//得到部署resolver所需的参数
	endPoint, hexPK, localAddress, err := getParamsForDeployContract(node)
	if err != nil {
		return err
	}

	//得到indexer实例
	indexerAddr := common.HexToAddress(contracts.IndexerHex)
	indexer, err := upKeeping.NewIndexer(indexerAddr, contracts.GetClient(endPoint))
	if err != nil {
		fmt.Println("newIndexerErr:", err)
		return err
	}

	//检查是否已经部署过resolver
	_, resAddr, err := indexer.Get(&bind.CallOpts{
		From: localAddress,
	}, localAddress.String())
	if err != nil {
		fmt.Println("Geterr:", err)
		return err
	}
	//如果部署过resolver，那么也部署过offer，目前可以先只用部署一次offer，后面可以加一个选项来选择是否要再部署新的offer
	if resAddr.String() != "" && resAddr.String() != contracts.InvalidAddr { //部署过
		fmt.Println("you have deployed resolver and offer, resolverAddr is: ", resAddr.String())

		if reDeployOffer { //用户想要重新部署offer合约
			//部署offer
			fmt.Println("provider wants to redeploy offer-contract")
			err = contracts.DeployOffer(endPoint, localAddress, hexPK, capacity, duration, price)
			if err != nil {
				return err
			}
		}

		return nil
	}

	//获得用户的账户余额
	balance, _ := contracts.QueryBalance(endPoint, localAddress.Hex())
	fmt.Println("balance:", balance)

	//判断余额是否能够部署resolver以及offer
	deployPrice := 931369000000000
	leastMoney := big.NewInt(int64(deployPrice))
	if balance.Cmp(leastMoney) < 0 { //余额不足
		return errBalance
	}

	//部署resolver
	err = contracts.DeployResolver(endPoint, hexPK, localAddress, indexer)
	if err != nil {
		return err
	}

	//部署offer
	err = contracts.DeployOffer(endPoint, localAddress, hexPK, capacity, duration, price)
	if err != nil {
		return err
	}

	return nil
}
