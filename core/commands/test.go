/*
命令行 mefs test命令的操作，用于对系统做出各样的测试
包括测试特定的函数，显示当前节点的各项参数等
*/

package commands

import (
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/common"
	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/core/commands/cmdenv"
	"github.com/memoio/go-mefs/role/keeper"
	"github.com/memoio/go-mefs/role/user"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
)

const (
	adminSk = "928969b4eb7fbca964a41024412702af827cbc950dbe9268eae9f5df668c85b4"
)

var TestCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "",
		ShortDescription: `
`,
	},

	Subcommands: map[string]*cmds.Command{
		"helloworld":               helloWorldCmd, //命令行操作写法示例
		"localinfo":                infoCmd,
		"resultsummary":            resultSummaryCmd,
		"savePay":                  savePayCmd,
		"deployKeeper":             deployKeeperCmd,             //deploy keeper contract
		"setKeeper":                setKeeperCmd,                //将传入的账户设为keeper
		"deployProvider":           deployProviderCmd,           //deploy keeper contract
		"setProvider":              setProviderCmd,              //将传入的账户设为provider
		"showBalance":              showBalanceCmd,              //用于测试，查看自己的余额或者指定账户的余额
		"deployKeeperProviderMap":  deployKeeperProviderMapCmd,  //部署 KeeperProviderMap 合约
		"addKeeperProviderToKPMap": addKeeperProviderToKPMapCmd, //往KeeperProviderMap里添加keeper和provider
		"deleteProviderInKPMap":    deleteProviderInKPMapCmd,    //删除KeeperProviderMap里的指定provider
		"deleteKeeperInKPMap":      deleteKeeperInKPMapCmd,      //删除keeperProviderMap里的指定keeper
		"getProviderInKPMap":       getProviderInKPMapCmd,       //获得keeperProviderMap合约中与指定keeper关联的所有provider
		"getAllKeeperInKPMap":      getAllKeeperInKPMapCmd,      //获得keeperProviderMap合约中的所有keeper
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
		localAddress, _ := address.GetAddressFromID(id)
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
		hexPk := adminSk

		err := contracts.KeeperContract(hexPk)
		if err != nil {
			fmt.Println("keeper合约部署错误:", err)
			return err
		}
		fmt.Println("keeper部署合约成功")

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
		var addr string
		node, _ := cmdenv.GetNode(env) //获取当前ipfsnode
		if len(args) > 0 {
			addr = args[0]
		} else {
			id := node.Identity.Pretty()
			localAddress, _ := address.GetAddressFromID(id)
			addr = localAddress.String()
		}

		isKeeper, _ := req.Options["isKeeper"].(bool)

		hexPk := adminSk //此私钥部署过keeper合约

		err := contracts.SetKeeper(common.HexToAddress(addr), hexPk, isKeeper)
		if err != nil {
			fmt.Println("setKeeper err:", err)
			return err
		}
		fmt.Println("setKeeper success")

		return nil
	},
}

var deployProviderCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test deployProvider",
		ShortDescription: "deploy provider contract，we need remember the hexPk for testing setProvider",
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		hexPk := adminSk

		err := contracts.ProviderContract(hexPk)
		if err != nil {
			fmt.Println("keeper合约部署错误:", err)
			return err
		}
		fmt.Println("keeper部署合约成功")

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
		var addr string
		n, _ := cmdenv.GetNode(env) //获取当前ipfsnode
		if len(args) > 0 {
			addr = args[0]
		} else {
			id := n.Identity.Pretty()
			localAddress, _ := address.GetAddressFromID(id)
			addr = localAddress.String()
		}
		isProvider, _ := req.Options["isProvider"].(bool)

		hexPk := adminSk //此私钥部署过keeper合约

		err := contracts.SetKeeper(common.HexToAddress(addr), hexPk, isProvider)
		if err != nil {
			fmt.Println("setProvider err:", err)
			return err
		}
		fmt.Println("setProvider success")

		return nil
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
		balances, err := contracts.QueryBalance(addressid)
		if err != nil {
			return err
		}
		return cmds.EmitOnce(res, balances)
	},
}

var deployKeeperProviderMapCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test deployKeeperProviderMap",
		ShortDescription: "deploy KeeperProviderMap contract",
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		hexSk := adminSk
		err := contracts.DeployKeeperProviderMap(hexSk)
		if err != nil {
			fmt.Println("deployKeeperProviderMapErr:", err)
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

var addKeeperProviderToKPMapCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test addKeeperProviderToKPMapCmd",
		ShortDescription: "add KeeperProvider to KeeperProviderMap contract",
	},
	Arguments: []cmds.Argument{ //参数列表
		cmds.StringArg("kaddress", false, true, "keeper address, 0x..."),
		cmds.StringArg("paddress", false, true, "provider address, 0x..."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		hexSk := adminSk
		localAddr, _ := address.GetAdressFromSk(hexSk)
		localAddress := common.HexToAddress(localAddr[2:])

		kaddr := req.Arguments[0]
		paddr := req.Arguments[1]

		keeperAddr := common.HexToAddress(kaddr[2:])
		providerAddr := common.HexToAddress(paddr[2:])

		err := contracts.AddKeeperProvidersToKPMap(localAddress, hexSk, keeperAddr, []common.Address{providerAddr})
		if err != nil {
			fmt.Println("addKeeperProviderToKPMapErr:", err)
			return err
		}

		fmt.Println("Add keeper and provider map success")

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

var deleteProviderInKPMapCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test delete Provider in KPMap",
		ShortDescription: "delete Provider in KeeperProviderMap contract",
	},
	Arguments: []cmds.Argument{ //参数列表
		cmds.StringArg("kaddress", false, true, "keeper address, 0x..."),
		cmds.StringArg("paddress", false, true, "provider address, 0x..."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		hexSk := adminSk
		localAddr, _ := address.GetAdressFromSk(hexSk)
		localAddress := common.HexToAddress(localAddr[2:])

		kaddr := req.Arguments[0]
		paddr := req.Arguments[1]

		keeperAddr := common.HexToAddress(kaddr[2:])
		providerAddr := common.HexToAddress(paddr[2:])

		//删除KeeperProviderMap合约中指定keeper下的一个provider
		err := contracts.DeleteProviderFromKPMap(localAddress, hexSk, keeperAddr, providerAddr)
		if err != nil {
			fmt.Println("DeleteProviderErr:", err)
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

var deleteKeeperInKPMapCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test delete keeper in KPMap",
		ShortDescription: "delete keeper in KeeperProviderMap contract",
	},
	Arguments: []cmds.Argument{ //参数列表
		cmds.StringArg("kaddress", false, true, "keeper address, 0x..."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		hexSk := adminSk
		localAddr, _ := address.GetAdressFromSk(hexSk)
		localAddress := common.HexToAddress(localAddr[2:])

		kaddr := req.Arguments[0]

		keeperAddr := common.HexToAddress(kaddr[2:])

		//删除KeeperProviderMap合约中指定的keeper以及与keeper关联的所有provider
		err := contracts.DeleteKeeperFromKPMap(localAddress, hexSk, keeperAddr)
		if err != nil {
			fmt.Println("DeleteKeeperErr:", err)
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

var getProviderInKPMapCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test get providers in KPMap",
		ShortDescription: "get providers with keeper-index in KeeperProviderMap contract",
	},
	Arguments: []cmds.Argument{ //参数列表
		cmds.StringArg("kaddress", false, true, "keeper address, 0x..."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		hexSk := adminSk
		localAddr, _ := address.GetAdressFromSk(hexSk)
		localAddress := common.HexToAddress(localAddr[2:])

		kaddr := req.Arguments[0]

		keeperAddr := common.HexToAddress(kaddr[2:])

		//获得KeeperProviderMap合约中与指定的keeper关联的所有provider
		providerAddrsGetted, err := contracts.GetProviderInKPMap(localAddress, keeperAddr)
		if err != nil {
			fmt.Println("GetProviderInKPMapErr:", err)
			return err
		}

		var providerIDsList []string
		if len(providerAddrsGetted) > 0 {
			for _, tmpProviderAddr := range providerAddrsGetted {
				tempID, _ := address.GetIDFromAddress(tmpProviderAddr.String())
				providerIDsList = append(providerIDsList, tempID)
			}
		}

		list := &StringList{
			ChildLists: providerIDsList,
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

var getAllKeeperInKPMapCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test get all keepers in KPMap",
		ShortDescription: "get all keepers in KeeperProviderMap contract",
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		hexSk := adminSk
		localAddr, _ := address.GetAdressFromSk(hexSk)
		localAddress := common.HexToAddress(localAddr[2:])

		//获得KeeperProviderMap合约中与指定的keeper关联的所有provider
		keeperAddrsGetted, err := contracts.GetAllKeeperInKPMap(localAddress)
		if err != nil {
			fmt.Println("GetAllKeeperInKPMapErr:", err)
			return err
		}

		var keeperIDsList []string
		if len(keeperAddrsGetted) > 0 {
			for _, tmpKeeperAddr := range keeperAddrsGetted {
				tempID, _ := address.GetIDFromAddress(tmpKeeperAddr.String())
				keeperIDsList = append(keeperIDsList, tempID)
			}
		}

		list := &StringList{
			ChildLists: keeperIDsList,
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
