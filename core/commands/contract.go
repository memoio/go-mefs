/*
命令行 mefs contract命令的操作，用于操作合约，主要功能是查询合约信息
*/

package commands

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/core/commands/cmdenv"
	id "github.com/memoio/go-mefs/crypto/identity"
	"github.com/memoio/go-mefs/utils/address"
)

const (
	codeName = "whoisyourdaddy"
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
		"addMasterKeeper":  addMasterKeeperCmd,
		"addMyProvider":    addMyProviderCmd,
		"isKeeper":         isKeeperCmd,
		"isProvider":       isProviderCmd,
		"getAdminOwner":    getAdminOwnerCmd,
		"getBannedVersion": getBannedCmd,
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

		localAddr := common.HexToAddress(req.Arguments[0][2:])

		isProvider, err := contracts.IsProvider(localAddr)
		if err != nil {
			fmt.Println("isProvider err:", err)
			return err
		}

		return cmds.EmitOnce(res, isProvider)
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
