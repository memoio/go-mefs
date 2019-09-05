/*
命令行 mefs test命令的操作，用于对系统做出各样的测试
包括测试特定的函数，显示当前节点的各项参数等
*/

package commands

import (
	"errors"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/contracts/indexer"
	"github.com/memoio/go-mefs/core/commands/cmdenv"
	fr "github.com/memoio/go-mefs/repo/fsrepo"
	"github.com/memoio/go-mefs/role/keeper"
	"github.com/memoio/go-mefs/role/user"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	ad "github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
)

var TestCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "",
		ShortDescription: `
`,
	},

	Subcommands: map[string]*cmds.Command{
		"keeper":         KeeperCmd,
		"helloworld":     helloWorldCmd, //命令行操作写法示例
		"localinfo":      infoCmd,
		"resultsummary":  resultSummaryCmd,
		"sc":             smartContractCmd,  //合约全流程测试
		"deployKeeper":   deployKeeperCmd,   //deploy keeper contract
		"setKeeper":      setKeeperCmd,      //将传入的账户设为keeper
		"deployProvider": deployProviderCmd, //deploy keeper contract
		"setProvider":    setProviderCmd,    //将传入的账户设为provider
		"channelTimeout": channelTimeoutCmd, //测试channel合约的部署以及channelTimeout()
		"closeChannel":   closeChannelCmd,   //测试channel合约的部署以及closeChannel()
		"showBalance":    showBalanceCmd,    //用于测试，查看自己的余额或者指定账户的余额
		"savePay":        savePayCmd,
	},
}

var helloWorldCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "the example of command",
		ShortDescription: `
		命令的示例，输入 mefs test helloword "str" 输出 str
	`,
	},

	Arguments: []cmds.Argument{ //参数列表
		cmds.StringArg("peerID", true, true, "The peerID to run the query against."),
	},
	Options: []cmds.Option{ //选项列表
		cmds.BoolOption("verbose", dhtVerboseOptionName, "Print extra information."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {

		list := &StringList{
			ChildLists: []string{"hello world!", "hello", "world"},
		}
		return cmds.EmitOnce(res, list)
	},
	Type: StringList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, fl *StringList) error {
			_, err := fmt.Fprintf(w, "%s", fl)
			return err
		}),
	},
}

//当前本节点运行时相关的数据，包括节点id，转换后的只能合约id
//TODO：添加节点角色，根据不同角色显示节点当前的关联节点（keeper管理的user，user雇佣的keeper等）
var infoCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "show info of this node",
		ShortDescription: `显示节点相关数据， 节点id，智能合约id，节点角色等信息`,
	},

	Arguments: []cmds.Argument{},
	Options:   []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {

		n, _ := cmdenv.GetNode(env) //获取当前ipfsnode
		id := n.Identity.Pretty()
		localAddress, _ := ad.GetAddressFromID(id)
		cfg, _ := n.Repo.Config()
		stringList := []string{"id:" + id, "address:" + localAddress.String(), "Role:" + cfg.Role}
		switch cfg.Role {
		case metainfo.RoleUser:
			outmap := user.ShowInfo(id)
			for key, value := range outmap {
				stringList = append(stringList, key+value)
			}
		default:
		}
		list := &StringList{
			ChildLists: stringList,
		}
		return cmds.EmitOnce(res, list)
	},
	Type: StringList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, fl *StringList) error {
			_, err := fmt.Fprintf(w, "%s", fl)
			return err
		}),
	},
}

var resultSummaryCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test resultSummary of keeper",
		ShortDescription: "测试时空值的计算，对某个provider的挑战数据进行计算，返回算好的时空值",
	},

	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {

		actual := keeper.ResultSummaryTest()
		list := &StringList{
			ChildLists: []string{actual},
		}
		return cmds.EmitOnce(res, list)
	},
	Type: StringList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, fl *IntList) error {
			_, err := fmt.Fprintf(w, "%s", fl)
			return err
		}),
	},
}

var smartContractCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test deploy->getbalance->spacetimepay",
		ShortDescription: "测试合约流程。单个节点进行合约部署，余额查询，时空支付操作",
	},
	Arguments: []cmds.Argument{ //参数列表
		cmds.StringArg("keeperCount", true, false, "the keeperCount needed in this test"),
		cmds.StringArg("providerCount", true, false, "the providerCount needed in this test"),
		cmds.StringArg("amount", true, false, "the pay amount "),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		keeperCount, err := strconv.Atoi(req.Arguments[0])
		if err != nil {
			return err
		}
		providerCount, err := strconv.Atoi(req.Arguments[1])
		if err != nil {
			return err
		}
		amount, ok := big.NewInt(0).SetString(req.Arguments[2], 0)
		if !ok {
			return errors.New("SetString()error")
		}
		list := &StringList{}
		err = keeper.SmartContractTest(keeperCount, providerCount, amount)
		if err != nil {
			list.ChildLists = []string{err.Error()}
		} else {
			list.ChildLists = []string{"Complete!"}
		}
		return cmds.EmitOnce(res, list)
	},
	Type: StringList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, fl *StringList) error {
			_, err := fmt.Fprintf(w, "%s", fl)
			return err
		}),
	},
}

var savePayCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "checkLastPay->saveChalPay->checkLastPay",
		ShortDescription: "测试支付信息的存取",
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		list := &StringList{}
		err := keeper.SaveChalPayTest()
		if err != nil {
			list.ChildLists = []string{err.Error()}
		} else {
			list.ChildLists = []string{"Complete!"}
		}
		return cmds.EmitOnce(res, list)
	},
	Type: StringList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, fl *StringList) error {
			_, err := fmt.Fprintf(w, "%s", fl)
			return err
		}),
	},
}

var deployKeeperCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test deployKeeper",
		ShortDescription: "deploy keeper contract，we need remember the hexPk for testing setKeeper",
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		keeper.DeployKeeperContractTest()
		list := &StringList{}
		return cmds.EmitOnce(res, list)
	},
	Type: StringList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, fl *StringList) error {
			_, err := fmt.Fprintf(w, "%s", fl)
			return err
		}),
	},
}

var setKeeperCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test setKeeper",
		ShortDescription: "set the account'Role keeper",
	},

	Arguments: []cmds.Argument{ //参数列表
		cmds.StringArg("address", false, true, "The address to set its Role keeper."),
	},
	Options: []cmds.Option{ //选项列表
		cmds.BoolOption("isKeeper", "isk", "set the address is keeper when it is true").WithDefault(true),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		args := req.Arguments
		var address string
		n, _ := cmdenv.GetNode(env) //获取当前ipfsnode
		if len(args) > 0 {
			address = args[0]
		} else {
			id := n.Identity.Pretty()
			localAddress, _ := ad.GetAddressFromID(id)
			address = localAddress.String()
		}

		isKeeper, _ := req.Options["isKeeper"].(bool)
		return keeper.SetKeeperTest(address, n, isKeeper)
	},
}

var deployProviderCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test deployProvider",
		ShortDescription: "deploy provider contract，we need remember the hexPk for testing setProvider",
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		keeper.DeployProviderContractTest()
		list := &StringList{}
		return cmds.EmitOnce(res, list)
	},
	Type: StringList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, fl *StringList) error {
			_, err := fmt.Fprintf(w, "%s", fl)
			return err
		}),
	},
}

var setProviderCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test deployProvider",
		ShortDescription: "set the account'Role provider",
	},

	Arguments: []cmds.Argument{ //参数列表
		cmds.StringArg("address", false, true, "The address to set its Role provider."),
	},
	Options: []cmds.Option{ //选项列表
		cmds.BoolOption("isProvider", "isp", "set the address is provider when it is true").WithDefault(true),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		args := req.Arguments
		var address string
		n, _ := cmdenv.GetNode(env) //获取当前ipfsnode
		if len(args) > 0 {
			address = args[0]
		} else {
			id := n.Identity.Pretty()
			localAddress, _ := ad.GetAddressFromID(id)
			address = localAddress.String()
		}
		isProvider, _ := req.Options["isProvider"].(bool)
		return keeper.SetProviderTest(address, n, isProvider)
	},
}

var channelTimeoutCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test channel-contract",
		ShortDescription: "deploy channel-contract between two accounts, then one account call the channelTimeout()",
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		n, err := cmdenv.GetNode(env) //获取当前ipfsnode
		if err != nil {
			return err
		}
		id := n.Identity.Pretty()
		localAddress, err := ad.GetAddressFromID(id)
		if err != nil {
			return err
		}
		hexKey, err := fr.GetHexPrivKeyFromKS(n.Identity, n.Password)
		if err != nil {
			fmt.Println("getHexPKErr", err)
			return err
		}
		config, err := n.Repo.Config()
		if err != nil {
			return err
		}
		ethEndPoint := config.Eth
		err = testChannelTimeout(localAddress, hexKey, ethEndPoint)
		if err != nil {
			fmt.Println("channelTimeout failed")
			return err
		}
		list := &StringList{}
		return cmds.EmitOnce(res, list)
	},
	Type: StringList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, fl *StringList) error {
			_, err := fmt.Fprintf(w, "%s", fl)
			return err
		}),
	},
}

var closeChannelCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test channel-contract",
		ShortDescription: "deploy channel-contract between two accounts, then one account call the channelTimeout()",
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		n, err := cmdenv.GetNode(env) //获取当前ipfsnode
		if err != nil {
			return err
		}
		id := n.Identity.Pretty()
		localAddress, err := ad.GetAddressFromID(id)
		if err != nil {
			return err
		}
		hexKey, err := fr.GetHexPrivKeyFromKS(n.Identity, n.Password)
		if err != nil {
			fmt.Println("getHexPKErr", err)
			return err
		}
		config, err := n.Repo.Config()
		if err != nil {
			return err
		}
		ethEndPoint := config.Eth
		err = testCloseChannel(localAddress, hexKey, ethEndPoint)
		if err != nil {
			return err
		}
		list := &StringList{}
		return cmds.EmitOnce(res, list)
	},
	Type: StringList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, fl *StringList) error {
			_, err := fmt.Fprintf(w, "%s", fl)
			return err
		}),
	},
}

var showBalanceCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "show balance in the account",
		ShortDescription: `
	'
	mefs lfs show_storage show balance in the account
	`,
	},

	Arguments: []cmds.Argument{},
	Options: []cmds.Option{
		cmds.StringOption(AddressID, "addr", "The practice user's addressid that you want to exec").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		var userid string
		addressid, found := req.Options[AddressID].(string)
		if addressid == "" || !found {
			userid = node.Identity.Pretty()
			address, err := address.GetAddressFromID(userid)
			addressid = address.String()
			if err != nil {
				return err
			}
		} else {
			userid, err = address.GetIDFromAddress(addressid)
			if err != nil {
				return err
			}
		}
		config, err := node.Repo.Config()
		if err != nil {
			fmt.Println(err)
			return err
		}
		endpoint := config.Eth
		balances, err := contracts.QueryBalance(endpoint, addressid)
		if err != nil {
			return err
		}
		return cmds.EmitOnce(res, balances)
	},
}

//TestChannelTimeout test channelTimeout()
func testChannelTimeout(localAddr common.Address, hexKey string, ethEndPoint string) (err error) {
	fmt.Println("==========开始测试channelTimeout=========")
	balance, err := contracts.QueryBalance(ethEndPoint, localAddr.String()) //查看账户余额
	if err != nil {
		fmt.Println("contracts.QueryBalanceErr:", err)
		return err
	}
	fmt.Println("部署前balance:", balance)

	//部署channel合约，测试中这个账户与自己部署channel合约
	indexerAddr := common.HexToAddress(contracts.IndexerHex)
	indexer, err := indexer.NewIndexer(indexerAddr, contracts.GetClient(ethEndPoint))
	if err != nil {
		fmt.Println("newIndexerErr:", err)
		return err
	}
	err = contracts.DeployResolver(ethEndPoint, hexKey, localAddr, indexer)
	if err != nil {
		fmt.Println("deployResolverErr:", err)
		return err
	}
	timeout := big.NewInt(60)
	moneyToChannel := big.NewInt(1000000)
	channelAddr, err := contracts.DeployChannelContract(ethEndPoint, hexKey, localAddr, localAddr, timeout, moneyToChannel)
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

	_, err = contracts.GetChannelStartDate(ethEndPoint, localAddr, localAddr, localAddr)
	if err != nil {
		fmt.Println("GetChannelStartDateErr:", err)
		return err
	}

	//触发channelTimeout()
	err = contracts.ChannelTimeout(ethEndPoint, hexKey, localAddr, localAddr)
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
func testCloseChannel(localAddr common.Address, hexKey string, ethEndPoint string) (err error) {
	fmt.Println("==========开始测试closeChannel=========")
	balance, err := contracts.QueryBalance(ethEndPoint, localAddr.String()) //查看账户余额
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
	err = contracts.DeployResolver(ethEndPoint, hexKey, localAddr, indexer)
	if err != nil {
		fmt.Println("deployResolverErr:", err)
		return err
	}
	timeout := big.NewInt(120)
	moneyToChannel := big.NewInt(1000000)
	channelAddr, err := contracts.DeployChannelContract(ethEndPoint, hexKey, localAddr, localAddr, timeout, moneyToChannel)
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
	balance, err = contracts.QueryBalance(ethEndPoint, localAddr.String()) //查看部署channel合约后的账户余额
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
	err = contracts.CloseChannel(ethEndPoint, hexKey, localAddr, localAddr, sig, value)
	if err != nil {
		fmt.Println("CloseChannelErr:", err)
		return err
	}

	time.Sleep(120 * time.Second)
	balance, err = contracts.QueryBalance(ethEndPoint, localAddr.String()) //查看触发Closechannel合约后的账户余额
	fmt.Println("触发后balance:", balance)
	balance, err = contracts.QueryBalance(ethEndPoint, channelAddr.String()) //查看channel合约的账户余额
	fmt.Println("channel合约的balance:", balance)
	if balance.Cmp(big.NewInt(0)) != 0 {
		return errors.New("close channel failed")
	}
	return nil
}
