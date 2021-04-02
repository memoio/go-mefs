/*
命令行 mefs test命令的操作，用于对系统做出各样的测试
包括测试特定的函数，显示当前节点的各项参数等
*/

package commands

import (
	"fmt"
	"io"
	"math/big"

	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/core/commands/cmdenv"
	mcl "github.com/memoio/go-mefs/crypto/bls12"
	"github.com/memoio/go-mefs/test"
	"github.com/memoio/go-mefs/utils/address"
)

const (
	moneyTo  = 1100000000000000000
	multiple = 1000
)

var TestCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "",
		ShortDescription: `
`,
	},

	Subcommands: map[string]*cmds.Command{
		"localinfo":     infoCmd,
		"showBalance":   showBalanceCmd, //用于测试，查看自己的余额或者指定账户的余额
		"mcl":           mclCmd,
		"transferMoney": transferCmd, //用于给指定账户转账
	},
}

var transferCmd = &cmds.Command{

	Helptext: cmds.HelpText{
		Tagline: "transfer money to the account",
		ShortDescription: `
		'
		mefs test transferMoney transfer money to the account
		`,
	},

	Arguments: []cmds.Argument{},
	Options: []cmds.Option{
		cmds.StringOption("address", "addr", "The practice user's addressid that you want to transfer to").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		toAddr, _ := req.Options["address"].(string)
		test.TransferTo(new(big.Int).Mul(big.NewInt(moneyTo), big.NewInt(multiple)), toAddr, "http://119.147.213.219:8101", "http://119.147.213.219:8101")

		balances, err := contracts.QueryBalance(toAddr)
		if err != nil {
			return err
		}
		return cmds.EmitOnce(res, balances)
	},
}

var mclCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "test mcl lib",
		ShortDescription: `test mcl lib
	`,
	},

	Arguments: []cmds.Argument{},
	Options:   []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {

		err := mcl.Init(mcl.BLS12_381)
		if err != nil {
			panic(err)
		}

		list := &StringList{
			ChildLists: []string{"mcl is ok!"},
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

//当前本节点运行时相关的数据，包括节点id，转换后的智能合约id
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
		stringList := []string{"id: " + id, "address: " + localAddress.String(), "Role: " + cfg.Role}
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

var showBalanceCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "show balance in the account",
		ShortDescription: `
	'
	mefs test showBalance show balance in the account
	`,
	},

	Arguments: []cmds.Argument{},
	Options: []cmds.Option{
		cmds.StringOption("address", "addr", "The practice user's addressid that you want to exec").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		var userid string
		addressid, found := req.Options["address"].(string)
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
