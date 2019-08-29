package provider

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/contracts/indexer"
	"github.com/memoio/go-mefs/core"
	"github.com/memoio/go-mefs/repo/fsrepo"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

//TestChannelTimeout test channelTimeout()
func TestChannelTimeout(providerAddr common.Address, hexKey string) (err error) {
	fmt.Println("==========开始测试channelTimeout=========")
	config, err := localNode.Repo.Config()
	if err != nil {
		return err
	}
	ethEndPoint := config.Eth
	balance, err := contracts.QueryBalance(ethEndPoint, providerAddr.String()) //查看账户余额
	if err != nil {
		fmt.Println("contracts.QueryBalanceErr:", err)
		return err
	}
	fmt.Println("部署前balance:", balance)

	//部署channel合约，测试中这个provider账户与自己部署channel合约
	indexerAddr := common.HexToAddress(contracts.IndexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, contracts.GetClient(ethEndPoint))
	if err != nil {
		fmt.Println("newIndexerErr:", err)
		return err
	}
	err = contracts.DeployResolver(ethEndPoint, hexKey, providerAddr, indexer)
	if err != nil {
		fmt.Println("deployResolverErr:", err)
		return err
	}
	timeout := big.NewInt(60)
	moneyToChannel := big.NewInt(1000000)
	channelAddr, err := contracts.DeployChannelContract(ethEndPoint, hexKey, providerAddr, providerAddr, timeout, moneyToChannel)
	if err != nil {
		fmt.Println("deployChannelErr:", err)
		return err
	}

	time.Sleep(120 * time.Second)
	balance, err = contracts.QueryBalance(ethEndPoint, channelAddr.String()) //查看部署的channel合约的账户余额
	if err != nil {
		fmt.Println("contracts.QueryBalanceErr:", err)
		return err
	}
	fmt.Println("channel合约的balance:", balance)

	//触发channelTimeout()
	err = contracts.ChannelTimeout(ethEndPoint, hexKey, providerAddr, providerAddr)
	if err != nil {
		fmt.Println("channelTimeoutErr:", err)
		return err
	}

	time.Sleep(120 * time.Second)
	balance, err = contracts.QueryBalance(ethEndPoint, channelAddr.String()) //查看触发channelTimeout后的合约余额
	if err != nil {
		fmt.Println("contracts.QueryBalanceErr:", err)
		return err
	}
	fmt.Println("触发channelTimeout后的合约balance:", balance)
	if balance.Cmp(big.NewInt(0)) != 0 {
		return errors.New("channel timeout failed")
	}
	return nil
}

//TestCloseChannel test CloseChannel()
//因为是同一个账户，所以只能保证流程是通的，但是无法确认金额是否转给provider
//看合约代码，应该是会转过去的，所以closeChannel应该是测通的，当然还需要能连上服务器上节点时再进行本地测试
func TestCloseChannel(n *core.MefsNode) (err error) {
	id := n.Identity.Pretty()
	providerAddr, err := address.GetAddressFromID(id)
	if err != nil {
		return err
	}
	hexKey, err := fsrepo.GetHexPrivKeyFromKS(n.Identity, n.Password)
	if err != nil {
		fmt.Println("getHexPKErr", err)
		return err
	}
	config, err := localNode.Repo.Config()
	if err != nil {
		return err
	}
	ethEndPoint := config.Eth
	fmt.Println("==========开始测试closeChannel=========")
	balance, err := contracts.QueryBalance(ethEndPoint, providerAddr.String()) //查看账户余额
	if err != nil {
		fmt.Println("QueryBalanceErr", err)
		return err
	}
	fmt.Println("部署前balance:", balance)

	//部署channel合约，测试中这个provider账户与自己部署channel合约
	indexerAddr := common.HexToAddress(contracts.IndexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, contracts.GetClient(ethEndPoint))
	if err != nil {
		fmt.Println("newIndexerErr:", err)
		return err
	}
	err = contracts.DeployResolver(ethEndPoint, hexKey, providerAddr, indexer)
	if err != nil {
		fmt.Println("deployResolverErr:", err)
		return err
	}
	timeout := big.NewInt(120)
	moneyToChannel := big.NewInt(1000000)
	channelAddr, err := contracts.DeployChannelContract(ethEndPoint, hexKey, providerAddr, providerAddr, timeout, moneyToChannel)
	if err != nil {
		fmt.Println("deployChannelErr:", err)
		return err
	}

	//签名
	value := big.NewInt(100)
	sig, err := contracts.SignForChannel(channelAddr, value, hexKey)
	if err != nil {
		fmt.Println("SignForChannelErr:", err)
		return err
	}

	//账户验证签名
	pubKey, err := utils.GetCompressedPkFromHexSk(hexKey)
	if err != nil {
		fmt.Println("GetCompressedPkFromHexSkErr:", err)
		return err
	}
	verify, err := contracts.VerifySig(pubKey, sig, channelAddr, value)
	if err != nil {
		fmt.Println("verifyErr:", err)
		return err
	}
	if !verify {
		fmt.Println("verify失败")
		return errors.New("verify失败")
	}

	time.Sleep(60 * time.Second)
	balance, err = contracts.QueryBalance(ethEndPoint, providerAddr.String()) //查看部署channel合约后的账户余额
	if err != nil {
		fmt.Println("QueryBalanceErr:", err)
		return err
	}
	fmt.Println("部署后balance:", balance)
	balance, err = contracts.QueryBalance(ethEndPoint, channelAddr.String()) //查看channel合约的账户余额
	if err != nil {
		fmt.Println("QueryBalanceErr:", err)
		return err
	}
	if balance.Cmp(moneyToChannel) != 0 {
		return errors.New("channel-balance is not true")
	}
	fmt.Println("channel合约的balance:", balance)

	//触发CloseChannel()
	err = contracts.CloseChannel(ethEndPoint, hexKey, providerAddr, providerAddr, sig, value)
	if err != nil {
		fmt.Println("CloseChannelErr:", err)
		return err
	}

	time.Sleep(120 * time.Second)
	balance, err = contracts.QueryBalance(ethEndPoint, providerAddr.String()) //查看触发Closechannel合约后的账户余额
	fmt.Println("触发后balance:", balance)
	balance, err = contracts.QueryBalance(ethEndPoint, channelAddr.String()) //查看channel合约的账户余额
	fmt.Println("channel合约的balance:", balance)
	if balance.Cmp(big.NewInt(0)) != 0 {
		return errors.New("close channel failed")
	}
	return nil
}
