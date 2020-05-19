package commands

import (
	"bytes"
	"context"
	"fmt"
	"io"

	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/memoio/go-mefs/core/commands/cmdenv"
	"github.com/memoio/go-mefs/manageNode/keeper"
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

var KeeperCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Interact with keeper.",
		ShortDescription: `
'mefs-keeper info' is a plumbing command used to manipulate keeper service.
`,
	},

	Subcommands: map[string]*cmds.Command{
		"list_users":     KeeperListUsersCmd,
		"list_providers": KeeperListProvidersCmd,
		"list_keepers":   KeeperListKeepersCmd,
		"flush":          KeeperFlushCmd,
	},
}

var KeeperListUsersCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "List Users.",
		ShortDescription: `
'mefs-keeper info list_users' is a plumbing command for printing users for a keeper.
`,
	},

	Arguments: []cmds.Argument{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}

		keeperIns, ok := node.Inst.(*keeper.Info)
		if !ok {
			return ErrNotReady
		}

		users, err := keeperIns.GetUsers()
		if err != nil {
			return err
		}
		list := &StringList{
			ChildLists: users,
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

var KeeperListProvidersCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "List Providers.",
		ShortDescription: `
'mefs-keeper info list_providers' is a plumbing command for printing providers for a keeper.
`,
	},

	Arguments: []cmds.Argument{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}
		keeperIns, ok := node.Inst.(*keeper.Info)
		if !ok {
			return ErrNotReady
		}
		providers, err := keeperIns.GetProviders()

		if err != nil {
			return err
		}
		list := &StringList{
			ChildLists: providers,
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

var KeeperListKeepersCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "List keepers.",
		ShortDescription: `
'mefs-keeper info list_keepers' is a plumbing command for printing providers for a keeper.
`,
	},

	Arguments: []cmds.Argument{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}

		keeperIns, ok := node.Inst.(*keeper.Info)
		if !ok {
			return ErrNotReady
		}

		keepers, err := keeperIns.GetKeepers()
		if err != nil {
			return err
		}
		list := &StringList{
			ChildLists: keepers,
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

var KeeperFlushCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Flush keepers and providers.",
		ShortDescription: `
'mefs-keeper info flush' is a plumbing command for printing providers for a keeper.
`,
	},

	Arguments: []cmds.Argument{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}
		keeperIns, ok := node.Inst.(*keeper.Info)
		if !ok {
			return ErrNotReady
		}
		return keeperIns.FlushPeers(context.Background())
	},
}
