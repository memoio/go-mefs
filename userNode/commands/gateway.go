package commands

import (
	"fmt"
	"io"

	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/memoio/go-mefs/core/commands/cmdenv"
	"github.com/memoio/go-mefs/userNode/miniogw"
	"github.com/memoio/go-mefs/userNode/user"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

var GatewayCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "",
		ShortDescription: `
`,
	},

	Subcommands: map[string]*cmds.Command{
		"start": gwStartCmd, //命令行操作写法示例
	},
}

var gwStartCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "start gateway",
		ShortDescription: `gateway start
	`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("addr", false, false, "The user's account that you want to start gateway for"),
	},
	Options: []cmds.Option{
		cmds.StringOption(PassWord, "pwd", "The password for user").WithDefault(""),
		cmds.StringOption("EndPoint", "url", "The gateway endpoint: ip:port, default is: 127.0.0.1:5080").WithDefault("127.0.0.1:5080"),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if !node.OnlineMode() {
			return ErrNotOnline
		}
		userIns, ok := node.Inst.(*user.Info)
		if !ok {
			return ErrNotReady
		}

		var uid, addr string
		if len(req.Arguments) > 0 {
			addr = req.Arguments[0]
			uid, err = address.GetIDFromAddress(addr)
			if err != nil {
				return err
			}
		} else {
			uid = node.Identity.Pretty()
			gotaddr, err := address.GetAddressFromID(uid)
			if err != nil {
				return err
			}
			addr = gotaddr.String()
		}

		lfs := userIns.GetUser(uid)
		if lfs == nil || !lfs.Online() {
			return errLfsServiceNotReady
		}

		pwd, ok := req.Options[PassWord].(string)
		if !ok || pwd == "" {
			retry := 0
			for {
				gotpwd, err := utils.GetPassWord()
				if err != nil {
					if retry > 2 {
						return err
					}
					retry++
					continue
				}
				pwd = gotpwd
				break
			}
		}

		if len(pwd) < 8 {
			return errWrongInput
		}

		ep, ok := req.Options["EndPoint"].(string)
		if !ok {
			return errWrongInput
		}

		err = miniogw.Start(addr, pwd, ep)
		if err != nil {
			return err
		}
		list := &StringList{
			ChildLists: []string{"Gateway of " + addr + " started at: " + ep},
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
