package commands

import (
	"fmt"
	"io"

	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/memoio/go-mefs/role/keeper"
)

var KeeperCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Interact with keeper.",
		ShortDescription: `
'mefs keeper' is a plumbing command used to manipulate keeper service.
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
'mefs keeper list_users' is a plumbing command for printing users for a keeper.
`,
	},

	Arguments: []cmds.Argument{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		users, err := keeper.GetUsers()
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
'mefs keeper list_providers' is a plumbing command for printing providers for a keeper.
`,
	},

	Arguments: []cmds.Argument{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		providers, err := keeper.GetProviders()
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
'mefs keeper list_keepers' is a plumbing command for printing providers for a keeper.
`,
	},

	Arguments: []cmds.Argument{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		keepers, err := keeper.GetKeepers()
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
		Tagline: "List keepers.",
		ShortDescription: `
'mefs keeper list_keepers' is a plumbing command for printing providers for a keeper.
`,
	},

	Arguments: []cmds.Argument{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		return keeper.FlushKeepersAndProviders()
	},
}
