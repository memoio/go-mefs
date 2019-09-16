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
	"github.com/memoio/go-mefs/utils/address"
)

const (
	adminSk  = "928969b4eb7fbca964a41024412702af827cbc950dbe9268eae9f5df668c85b4"
	codeName = "whoisyourdaddy"
)

var ContractCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "",
		ShortDescription: `contract comamnd, CodeName is required`,
	},

	Subcommands: map[string]*cmds.Command{
		"deployKeeper":             deployKeeperCmd,             //deploy keeper contract
		"setKeeper":                setKeeperCmd,                //将传入的账户设为keeper
		"deployProvider":           deployProviderCmd,           //deploy keeper contract
		"setProvider":              setProviderCmd,              //将传入的账户设为provider
		"deployKeeperProviderMap":  deployKeeperProviderMapCmd,  //部署 KeeperProviderMap 合约
		"addKeeperProviderToKPMap": addKeeperProviderToKPMapCmd, //往KeeperProviderMap里添加keeper和provider
		"deleteProviderInKPMap":    deleteProviderInKPMapCmd,    //删除KeeperProviderMap里的指定provider
		"deleteKeeperInKPMap":      deleteKeeperInKPMapCmd,      //删除keeperProviderMap里的指定keeper
		"getProviderInKPMap":       getProviderInKPMapCmd,       //获得keeperProviderMap合约中与指定keeper关联的所有provider
		"getAllKeeperInKPMap":      getAllKeeperInKPMapCmd,      //获得keeperProviderMap合约中的所有keeper
	},
}

var deployKeeperCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test deployKeeper",
		ShortDescription: "deploy keeper contract，we need remember the hexPk for testing setKeeper",
	},
	Options: []cmds.Option{
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}
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
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		args := req.Arguments
		var addr string
		node, _ := cmdenv.GetNode(env) //获取当前mefsnode
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
	Options: []cmds.Option{
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}
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
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}
		args := req.Arguments
		var addr string
		n, _ := cmdenv.GetNode(env) //获取当前mefsnode
		if len(args) > 0 {
			addr = args[0]
		} else {
			id := n.Identity.Pretty()
			localAddress, _ := address.GetAddressFromID(id)
			addr = localAddress.String()
		}
		isProvider, _ := req.Options["isProvider"].(bool)

		hexPk := adminSk //此私钥部署过keeper合约

		err := contracts.SetProvider(common.HexToAddress(addr), hexPk, isProvider)
		if err != nil {
			fmt.Println("setProvider err:", err)
			return err
		}
		fmt.Println("setProvider success")

		return nil
	},
}

var deployKeeperProviderMapCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test deployKeeperProviderMap",
		ShortDescription: "deploy KeeperProviderMap contract",
	},
	Options: []cmds.Option{
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}
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
		cmds.StringArg("kaddress", true, true, "keeper address, 0x...").EnableStdin(),
		cmds.StringArg("paddress", true, true, "provider address, 0x...").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}
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
		cmds.StringArg("kaddress", true, true, "keeper address, 0x...").EnableStdin(),
		cmds.StringArg("paddress", true, true, "provider address, 0x...").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}
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
		cmds.StringArg("kaddress", true, true, "keeper address, 0x...").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}
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
		cmds.StringArg("kaddress", true, true, "keeper address, 0x...").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
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
				providerIDsList = append(providerIDsList, tmpProviderAddr.String())
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
	Options: []cmds.Option{
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
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
				keeperIDsList = append(keeperIDsList, tmpKeeperAddr.String())
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
