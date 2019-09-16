/*
命令行 mefs test命令的操作，用于对系统做出各样的测试
包括测试特定的函数，显示当前节点的各项参数等
*/

package commands

import (
	"fmt"
	"io"

	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/core/commands/cmdenv"
	"github.com/memoio/go-mefs/role/keeper"
	"github.com/memoio/go-mefs/role/user"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
)

var TestCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "",
		ShortDescription: `
`,
	},

	Subcommands: map[string]*cmds.Command{
		"helloworld":    helloWorldCmd, //命令行操作写法示例
		"localinfo":     infoCmd,
		"resultsummary": resultSummaryCmd,
		"savePay":       savePayCmd,
		"showBalance":   showBalanceCmd, //用于测试，查看自己的余额或者指定账户的余额
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
		stringList := []string{"id: " + id, "address: " + localAddress.String(), "Role: " + cfg.Role}
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
