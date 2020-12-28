/*
命令行 mefs test命令的操作，用于对系统做出各样的测试
包括测试特定的函数，显示当前节点的各项参数等
*/

package commands

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/core/commands/cmdenv"
	id "github.com/memoio/go-mefs/crypto/identity"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

const (
	adminSk   = "928969b4eb7fbca964a41024412702af827cbc950dbe9268eae9f5df668c85b4"
	codeName  = "whoisyourdaddy"
	adminAddr = "0x0eb5b66c31b3c5a12aae81a9d629540b6433cac6"
	//Dev链和testnet链上的AdminOwned合约地址
	adminOwnedContractAddr = "0x8391984e2F1cC8F6b916F566C1D0a6bb8a15C73A"
)

type StringList struct {
	ChildLists []string
}

func (fl StringList) String() string {
	var buffer bytes.Buffer
	for i := 0; i < len(fl.ChildLists); i++ {
		buffer.WriteString(fl.ChildLists[i])
		buffer.WriteString("\n")
	}
	return buffer.String()
}

var ContractCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "",
		ShortDescription: `contract comamnd, CodeName is required`,
	},

	Subcommands: map[string]*cmds.Command{
		"deployResolver":           deployResolverCmd, //deploy resolver contract
		"deployKeeper":             deployKeeperCmd,   //deploy keeper contract
		"setKeeper":                setKeeperCmd,      //将传入的账户设为keeper
		"setKeeperPrice":           setKeeperPriceCmd,
		"deployProvider":           deployProviderCmd, //deploy keeper contract
		"setProvider":              setProviderCmd,    //将传入的账户设为provider
		"setProviderPrice":         setProviderPriceCmd,
		"deployKeeperProviderMap":  deployKeeperProviderMapCmd, //部署 KeeperProviderMap 合约
		"addMasterKeeper":          addMasterKeeperCmd,
		"addMyProvider":            addMyProviderCmd,
		"addKeeperProviderToKPMap": addKeeperProviderToKPMapCmd, //往KeeperProviderMap里添加keeper和provider
		"deleteProviderInKPMap":    deleteProviderInKPMapCmd,    //删除KeeperProviderMap里的指定provider
		"deleteKeeperInKPMap":      deleteKeeperInKPMapCmd,      //删除keeperProviderMap里的指定keeper
		"getProviderInKPMap":       getProviderInKPMapCmd,       //获得keeperProviderMap合约中与指定keeper关联的所有provider
		"getAllKeeperInKPMap":      getAllKeeperInKPMapCmd,      //获得keeperProviderMap合约中的所有keeper
		"isKeeper":                 isKeeperCmd,
		"isProvider":               isProviderCmd,
		"deployAdminOwned":         deployAdminOwnedCmd, //部署adminOwned合约
		"getAdminOwner":            getAdminOwnerCmd,
		"alterAdminOwner":          alterAdminOwnerCmd,
		"setBanned":                setBannedCmd,
		"getBanned":                getBannedCmd,
	},
}

var deployResolverCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test deployReover",
		ShortDescription: "deploy keeper contract，we need remember the hexPk for testing setKeeper",
	},
	Arguments: []cmds.Argument{ //参数列表
		cmds.StringArg("key", true, true, "The resolver key."),
	},
	Options: []cmds.Option{
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		hexPk := adminSk
		localAddr, err := id.GetAdressFromSk(hexPk)
		if err != nil {
			return err
		}

		_, _, err = contracts.GetResolverFromAdmin(localAddr, localAddr, req.Arguments[0], hexPk, true)
		if err != nil {
			fmt.Println("deploy resolver contract err:", err)
			return err
		}
		fmt.Println("deploy resolver contract success")

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

var deployKeeperCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test deployKeeper",
		ShortDescription: "deploy keeper contract，we need remember the hexPk for testing setKeeper",
	},
	Options: []cmds.Option{
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		hexPk := adminSk
		err := contracts.DeployKeeperAdmin(hexPk)
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
		cmds.StringArg("address", true, true, "The address to set its Role keeper."),
	},
	Options: []cmds.Option{ //选项列表
		cmds.BoolOption("isKeeper", "isk", "set the address is keeper when it is true").WithDefault(true),
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		hexPk := adminSk

		isKeeper, _ := req.Options["isKeeper"].(bool)

		localAddr := common.HexToAddress(req.Arguments[0][2:])

		err := contracts.SetKeeper(localAddr, hexPk, isKeeper)
		if err != nil {
			fmt.Println("setKeeper err:", err)
			return err
		}
		fmt.Println("setKeeper success")

		return nil
	},
}

var isKeeperCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test isKeeper",
		ShortDescription: "the account'Role is keeper?",
	},

	Arguments: []cmds.Argument{ //参数列表
		cmds.StringArg("address", true, true, "The address to test if it is Role keeper."),
	},
	Options: []cmds.Option{ //选项列表
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		localAddr := common.HexToAddress(req.Arguments[0][2:])

		isKeeper, err := contracts.IsKeeper(localAddr)
		if err != nil {
			fmt.Println("isKeeper err:", err)
			return err
		}

		return cmds.EmitOnce(res, isKeeper)
	},
}

var isProviderCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test isProvider",
		ShortDescription: "the account'Role is provider?",
	},

	Arguments: []cmds.Argument{ //参数列表
		cmds.StringArg("address", true, true, "The address to test if it is Role provider."),
	},
	Options: []cmds.Option{ //选项列表
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		localAddr := common.HexToAddress(req.Arguments[0][2:])

		isProvider, err := contracts.IsProvider(localAddr)
		if err != nil {
			fmt.Println("isProvider err:", err)
			return err
		}

		return cmds.EmitOnce(res, isProvider)
	},
}

var setKeeperPriceCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test set Keeper price",
		ShortDescription: "set the keeper price",
	},

	Options: []cmds.Option{ //选项列表
		cmds.Int64Option("depositPrice", "price", "deposit price").WithDefault(utils.KeeperDeposit),
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		hexPk := adminSk

		price, ok := req.Options["depositPrice"].(int64)
		if !ok || price <= 0 {
			price = utils.KeeperDeposit
		}

		localAddr := common.HexToAddress(adminAddr[2:])

		oldPrice, err := contracts.GetProviderPrice(localAddr)
		if err != nil {
			fmt.Println("get Keeper price err:", err)
			return err
		}

		fmt.Println("old deposit price is:", oldPrice)

		err = contracts.SetKeeperPrice(localAddr, hexPk, big.NewInt(price))
		if err != nil {
			fmt.Println("setKeeper price err:", err)
			return err
		}
		fmt.Println("setKeeper price success")

		return nil
	},
}

var deployProviderCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test deployProvider",
		ShortDescription: "deploy provider contract，we need remember the hexPk for testing setProvider",
	},
	Options: []cmds.Option{
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		hexPk := adminSk

		err := contracts.DeployProviderAdmin(hexPk)
		if err != nil {
			fmt.Println("provider合约部署错误:", err)
			return err
		}
		fmt.Println("provider部署合约成功")

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
		cmds.StringArg("address", true, true, "The address to set its Role provider."),
	},
	Options: []cmds.Option{ //选项列表
		cmds.BoolOption("isProvider", "isp", "set the address is provider when it is true").WithDefault(true),
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		hexPk := adminSk
		localAddr := common.HexToAddress(req.Arguments[0][2:])

		isProvider, _ := req.Options["isProvider"].(bool)

		err := contracts.SetProvider(localAddr, hexPk, isProvider)
		if err != nil {
			fmt.Println("setProvider err:", err)
			return err
		}
		fmt.Println("setProvider success")

		return nil
	},
}

var setProviderPriceCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test setKeeper",
		ShortDescription: "set the account'Role keeper",
	},

	Options: []cmds.Option{ //选项列表
		cmds.Int64Option("depositPrice", "price", "deposit price").WithDefault(utils.ProviderDeposit),
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		hexPk := adminSk

		dprice, ok := req.Options["depositPrice"].(int64)
		if !ok || dprice <= 0 {
			dprice = utils.ProviderDeposit
		}
		localAddr := common.HexToAddress(adminAddr[2:])

		oldPrice, err := contracts.GetProviderPrice(localAddr)
		if err != nil {
			fmt.Println("get Provider price err:", err)
			return err
		}

		fmt.Println("old deposit price is:", oldPrice)

		err = contracts.SetProviderPrice(localAddr, hexPk, big.NewInt(dprice))
		if err != nil {
			fmt.Println("set Provider price err:", err)
			return err
		}
		fmt.Println("set Provider price success")

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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		hexPk := adminSk
		err := contracts.DeployKPMap(hexPk)
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
		cmds.StringArg("kaddress", true, false, "keeper address, 0x...").EnableStdin(),
		cmds.StringArg("paddress", true, false, "provider address, 0x...").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		hexPk := adminSk
		localAddr, err := id.GetAdressFromSk(hexPk)
		if err != nil {
			return err
		}

		kaddr := req.Arguments[0]
		paddr := req.Arguments[1]

		keeperAddr := common.HexToAddress(kaddr[2:])
		providerAddr := common.HexToAddress(paddr[2:])

		err = contracts.AddKeeperProvidersToKPMap(localAddr, hexPk, keeperAddr, []common.Address{providerAddr})
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

var addMasterKeeperCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "addMasterKeeperCmd",
		ShortDescription: "add master Keeper to KeeperProviderMap contract",
	},
	Arguments: []cmds.Argument{ //参数列表
		cmds.StringArg("kaddress", true, false, "keeper address, 0x...").EnableStdin(),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if !node.OnlineMode() {
			return ErrNotOnline
		}

		cfg, err := node.Repo.Config()
		if err != nil {
			return err
		}

		contracts.EndPoint = cfg.Eth

		hexSk := node.PrivateKey
		localAddr, _ := id.GetAdressFromSk(hexSk)

		kaddr := req.Arguments[0]

		keeperAddr := common.HexToAddress(kaddr[2:])

		err = contracts.AddKeeperProvidersToKPMap(localAddr, hexSk, keeperAddr, []common.Address{localAddr})
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

var addMyProviderCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "addMyProviderCmd",
		ShortDescription: "add my provider to KeeperProviderMap contract",
	},
	Arguments: []cmds.Argument{ //参数列表
		cmds.StringArg("paddress", true, false, "provider address, 0x...").EnableStdin(),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if !node.OnlineMode() {
			return ErrNotOnline
		}

		cfg, err := node.Repo.Config()
		if err != nil {
			return err
		}

		contracts.EndPoint = cfg.Eth

		hexSk := node.PrivateKey
		localAddr, _ := id.GetAdressFromSk(hexSk)

		paddr := req.Arguments[0]

		providerAddr := common.HexToAddress(paddr[2:])

		err = contracts.AddKeeperProvidersToKPMap(localAddr, hexSk, localAddr, []common.Address{providerAddr})
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
		cmds.StringArg("kaddress", true, false, "keeper address, 0x...").EnableStdin(),
		cmds.StringArg("paddress", true, false, "provider address, 0x...").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		hexPk := adminSk
		localAddr, err := id.GetAdressFromSk(hexPk)
		if err != nil {
			return err
		}

		kaddr := req.Arguments[0]
		paddr := req.Arguments[1]

		keeperAddr := common.HexToAddress(kaddr[2:])
		providerAddr := common.HexToAddress(paddr[2:])

		//删除KeeperProviderMap合约中指定keeper下的一个provider
		err = contracts.DeleteProviderFromKPMap(localAddr, hexPk, keeperAddr, providerAddr)
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
		cmds.StringArg("kaddress", true, false, "keeper address, 0x...").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		hexPk := adminSk
		localAddr, err := id.GetAdressFromSk(hexPk)
		if err != nil {
			return err
		}

		kaddr := req.Arguments[0]

		keeperAddr := common.HexToAddress(kaddr[2:])

		//删除KeeperProviderMap合约中指定的keeper以及与keeper关联的所有provider
		err = contracts.DeleteKeeperFromKPMap(localAddr, hexPk, keeperAddr)
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
		cmds.StringArg("kaddress", true, false, "keeper address, 0x...").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		hexPk := adminSk
		localAddr, err := id.GetAdressFromSk(hexPk)
		if err != nil {
			return err
		}

		kaddr := req.Arguments[0]

		keeperAddr := common.HexToAddress(kaddr[2:])

		//获得KeeperProviderMap合约中与指定的keeper关联的所有provider
		providerAddrsGetted, err := contracts.GetProviderInKPMap(localAddr, keeperAddr)
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		hexPk := adminSk
		localAddr, err := id.GetAdressFromSk(hexPk)
		if err != nil {
			return err
		}

		//获得KeeperProviderMap合约中与指定的keeper关联的所有provider
		keeperAddrsGetted, err := contracts.GetAllKeeperInKPMap(localAddr)
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

var deployAdminOwnedCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test deployAdminOwned",
		ShortDescription: "deploy adminOwned contract，we need remember the address of adminOwned contract",
	},
	Options: []cmds.Option{
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		hexPk := adminSk
		adminOwnedAddr, err := contracts.DeployAdminOwned(hexPk)
		if err != nil {
			fmt.Println("AdminOwned合约部署错误:", err)
			return err
		}
		fmt.Println("AdminOwned合约部署合约成功")

		list := &StringList{
			ChildLists: []string{adminOwnedAddr.String()},
		}
		return cmds.EmitOnce(res, list)
	},
	Type: StringList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, fl *StringList) error {
			_, err := fmt.Fprintf(w, "contract address: %s", fl)
			return err
		}),
	},
}

var getAdminOwnerCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test get owner of AdminOwned",
		ShortDescription: "get the owner of AdminOwned-contract",
	},
	Options: []cmds.Option{
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		n, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		peerID := n.Identity.Pretty()
		localAddress, _ := address.GetAddressFromID(peerID)

		adminOwner, err := contracts.GetAdminOwner(common.HexToAddress(adminOwnedContractAddr), localAddress)
		if err != nil {
			return err
		}

		list := &StringList{
			ChildLists: []string{adminOwner.String()},
		}
		return cmds.EmitOnce(res, list)
	},
	Type: StringList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, fl *StringList) error {
			_, err := fmt.Fprintf(w, "owner of adminOwned address: %s", fl)
			return err
		}),
	},
}

var alterAdminOwnerCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test alter owner of AdminOwned",
		ShortDescription: "alter the owner of AdminOwned-contract",
	},
	Arguments: []cmds.Argument{
		cmds.StringArg("addr", true, false, "The new owner of AdminOwner-contract."),
	},
	Options: []cmds.Option{
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		newOwner := req.Arguments[0]
		hexPk := adminSk

		err := contracts.AlterOwner(hexPk, common.HexToAddress(adminOwnedContractAddr), common.HexToAddress(newOwner))
		if err != nil {
			return err
		}

		list := &StringList{
			ChildLists: []string{},
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

var setBannedCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test set param 'banned' of AdminOwned",
		ShortDescription: "set the param 'banned' of AdminOwned-contract",
	},
	Options: []cmds.Option{
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
		cmds.BoolOption("ParamBanned", "banned", "Set the param 'banned'").WithDefault(false),
		cmds.StringOption("ParamKey", "k", "Specify parameter index, can be: mapper、offer、query、channel、upkeeping、root、keeper、provider、kpMap").WithDefault("root"),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		hexPk := adminSk
		banned, ok := req.Options["ParamBanned"].(bool)
		if !ok {
			fmt.Println("ParamBanned is wrong")
			return nil
		}
		key, ok := req.Options["ParamKey"].(string)
		if !ok {
			fmt.Println("ParamKey is wrong")
			return nil
		}

		err := contracts.SetBanned(hexPk, key, common.HexToAddress(adminOwnedContractAddr), banned)
		if err != nil {
			return err
		}

		list := &StringList{
			ChildLists: []string{},
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

var getBannedCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test get param 'banned' of AdminOwned",
		ShortDescription: "get the param 'banned' of AdminOwned-contract",
	},
	Options: []cmds.Option{
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://212.64.28.207:8101"),
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
		cmds.StringOption("ParamKey", "k", "Specify parameter index, can be: mapper、offer、query、channel、upkeeping、root、keeper、provider、kpMap").WithDefault("root"),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cn := req.Options["CodeName"].(string)
		if cn != codeName {
			fmt.Println("CodeName is wrong")
			return nil
		}

		eth, ok := req.Options["EndPoint"].(string)
		if !ok {
			fmt.Println("Endpoint is wrong")
			return nil
		}

		contracts.EndPoint = eth

		n, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		peerID := n.Identity.Pretty()
		localAddress, _ := address.GetAddressFromID(peerID)

		key, ok := req.Options["ParamKey"].(string)
		if !ok {
			fmt.Println("ParamKey is wrong")
			return nil
		}

		banned, err := contracts.GetBanned(key, common.HexToAddress(adminOwnedContractAddr), localAddress)
		if err != nil {
			return err
		}

		list := &StringList{
			ChildLists: []string{key + "Banned is:", strconv.FormatBool(banned)},
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
