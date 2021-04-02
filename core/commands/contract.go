/*
命令行 mefs contract命令的操作，用于操作合约，主要功能是查询合约信息
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
	adminSk   = "aca26228a9ed5ca4da2dd08d225b1b1e049d80e1b126c0d7e644d04d0fb910a3"
	codeName  = "whoisyourdaddy"
	adminAddr = "0x1a249DB4cc739BD53b05E2082D3724b7e033F74F"
	//Dev链和testnet链上的AdminOwned合约地址
	adminOwnedContractAddr = "0x8026796Fd7cE63EAe824314AA5bacF55643e893d"
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
		"setBannedVersion":         setBannedCmd,
		"getBannedVersion":         getBannedCmd,
		"deployRecover":            deployRecoverCmd, //部署recover合约
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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

		hexSk := adminSk
		localAddr, err := id.GetAdressFromSk(hexSk)
		if err != nil {
			return err
		}

		cManage := contracts.NewCManage(localAddr, hexSk)
		_, _, err = cManage.GetResolverFromAdmin(req.Arguments[0], true)
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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
		cRole := contracts.NewCR("", hexPk)
		err := cRole.DeployKeeperAdmin()
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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

		hexSk := adminSk

		isKeeper, _ := req.Options["isKeeper"].(bool)

		localAddr := common.HexToAddress(req.Arguments[0][2:])
		cRole := contracts.NewCR("", hexSk)
		err := cRole.SetKeeper(localAddr, isKeeper)
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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

		localID, _ := address.GetIDFromAddress(req.Arguments[0])
		cRole := contracts.NewCR(localID, "")
		isKeeper, err := cRole.IsKeeper(localID)
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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

		localID, _ := address.GetIDFromAddress(req.Arguments[0])
		cRole := contracts.NewCR(localID, "")
		isProvider, err := cRole.IsProvider(localID)
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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

		hexSk := adminSk

		price, ok := req.Options["depositPrice"].(int64)
		if !ok || price <= 0 {
			price = utils.KeeperDeposit
		}

		localID, _ := address.GetIDFromAddress(adminAddr)
		cRole := contracts.NewCR(localID, hexSk)
		oldPrice, err := cRole.GetKeeperPrice()
		if err != nil {
			fmt.Println("get Keeper price err:", err)
			return err
		}

		fmt.Println("old deposit price is:", oldPrice)

		err = cRole.SetKeeperPrice(big.NewInt(price))
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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

		hexSk := adminSk

		cRole := contracts.NewCR("", hexSk)
		err := cRole.DeployProviderAdmin()
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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

		hexSk := adminSk
		localAddr := common.HexToAddress(req.Arguments[0])
		localID, _ := address.GetIDFromAddress(req.Arguments[0])

		isProvider, _ := req.Options["isProvider"].(bool)

		cRole := contracts.NewCR(localID, hexSk)
		err := cRole.SetProvider(localAddr, isProvider)
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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

		hexSk := adminSk

		dprice, ok := req.Options["depositPrice"].(int64)
		if !ok || dprice <= 0 {
			dprice = utils.ProviderDeposit
		}

		localID, _ := address.GetIDFromAddress(adminAddr)
		cRole := contracts.NewCR(localID, hexSk)
		oldPrice, err := cRole.GetProviderPrice()
		if err != nil {
			fmt.Println("get Provider price err:", err)
			return err
		}

		fmt.Println("old deposit price is:", oldPrice)

		err = cRole.SetProviderPrice(big.NewInt(dprice))
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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

		hexSk := adminSk
		cRole := contracts.NewCR("", hexSk)
		err := cRole.DeployKPMap()
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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

		hexSk := adminSk
		localAddr, err := id.GetAdressFromSk(hexSk)
		if err != nil {
			return err
		}

		kaddr := req.Arguments[0]
		paddr := req.Arguments[1]

		keeperAddr := common.HexToAddress(kaddr[2:])
		providerAddr := common.HexToAddress(paddr[2:])

		localID, _ := address.GetIDFromAddress(localAddr.Hex())
		cRole := contracts.NewCR(localID, hexSk)
		err = cRole.AddKeeperProvidersToKPMap(keeperAddr, []common.Address{providerAddr})
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

		localID, _ := address.GetIDFromAddress(localAddr.Hex())
		cRole := contracts.NewCR(localID, hexSk)
		err = cRole.AddKeeperProvidersToKPMap(keeperAddr, []common.Address{localAddr})
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

		localID, _ := address.GetIDFromAddress(localAddr.Hex())
		cRole := contracts.NewCR(localID, hexSk)
		err = cRole.AddKeeperProvidersToKPMap(localAddr, []common.Address{providerAddr})
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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

		hexSk := adminSk
		localAddr, err := id.GetAdressFromSk(hexSk)
		if err != nil {
			return err
		}

		kaddr := req.Arguments[0]
		paddr := req.Arguments[1]

		keeperAddr := common.HexToAddress(kaddr[2:])
		providerAddr := common.HexToAddress(paddr[2:])

		localID, _ := address.GetIDFromAddress(localAddr.Hex())
		cRole := contracts.NewCR(localID, hexSk)
		//删除KeeperProviderMap合约中指定keeper下的一个provider
		err = cRole.DeleteProviderFromKPMap(keeperAddr, providerAddr)
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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

		hexSk := adminSk
		localAddr, err := id.GetAdressFromSk(hexSk)
		if err != nil {
			return err
		}

		kaddr := req.Arguments[0]

		keeperAddr := common.HexToAddress(kaddr[2:])

		localID, _ := address.GetIDFromAddress(localAddr.Hex())
		cRole := contracts.NewCR(localID, hexSk)
		//删除KeeperProviderMap合约中指定的keeper以及与keeper关联的所有provider
		err = cRole.DeleteKeeperFromKPMap(keeperAddr)
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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

		localID, _ := address.GetIDFromAddress(localAddr.Hex())
		cRole := contracts.NewCR(localID, "")
		//获得KeeperProviderMap合约中与指定的keeper关联的所有provider
		providerAddrsGetted, err := cRole.GetProviderInKPMap(keeperAddr)
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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

		localID, _ := address.GetIDFromAddress(localAddr.Hex())
		cRole := contracts.NewCR(localID, "")
		//获得KeeperProviderMap合约中与指定的keeper关联的所有provider
		keeperAddrsGetted, err := cRole.GetAllKeeperInKPMap()
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
		cmds.StringOption("CodeName", "cn", "The CodeName this net used").WithDefault(""),
		cmds.UintOption("BannedVersion", "bv", "Set the bannedVersion").WithDefault(0),
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
		bannedVersion, ok := req.Options["BannedVersion"].(uint)
		if !ok {
			fmt.Println("ParamBanned is wrong")
			return nil
		}
		key, ok := req.Options["ParamKey"].(string)
		if !ok {
			fmt.Println("ParamKey is wrong")
			return nil
		}

		err := contracts.SetBannedVersion(hexPk, key, common.HexToAddress(adminOwnedContractAddr), uint16(bannedVersion))
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
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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

		bannedVersion, err := contracts.GetBannedVersion(key, common.HexToAddress(adminOwnedContractAddr), localAddress)
		if err != nil {
			return err
		}

		list := &StringList{
			ChildLists: []string{key + "BannedVersion is:", strconv.Itoa(int(bannedVersion))},
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

var deployRecoverCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "test deployRecover",
		ShortDescription: "deploy recover contract，we need remember the address of recover contract, and then write it in other contract",
	},
	Options: []cmds.Option{
		cmds.StringOption("EndPoint", "eth", "The Endpoint this net used").WithDefault("http://119.147.213.219:8101"),
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

		hexSk := adminSk
		recoverAddr, err := contracts.DeployRecover(hexSk)
		if err != nil {
			fmt.Println("recover合约部署错误:", err)
			return err
		}
		fmt.Println("recover合约部署合约成功")

		list := &StringList{
			ChildLists: []string{recoverAddr.String()},
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
