package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"time"

	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/memoio/go-mefs/core/commands/cmdenv"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

type allKeepers struct {
	KeeperCount int
	PledgeMoney *big.Int
	KeeperInfos []keeperInfo
}

type keeperInfo struct {
	Address     string
	Online      bool
	PledgeMoney *big.Int
	PledgeTime  string
}

type allProviders struct {
	ProviderCount int
	PledgeMoney   *big.Int
	ProInfos      []proInfo
}

type proInfo struct {
	Address     string
	Online      bool
	Storage     int64
	PledgeMoney *big.Int
	PledgeTime  string
}

var ListCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "list infomations",
		ShortDescription: `
`,
	},

	Subcommands: map[string]*cmds.Command{
		"keepers":   keeperCmd, //命令行操作写法示例
		"providers": proCmd,
	},
}

var keeperCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "list all keepers",
		ShortDescription: `list all keepers on chain
	`,
	},

	Arguments: []cmds.Argument{},
	Options:   []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		n, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		kItems, pledge, err := role.GetAllKeepers(n.Identity.Pretty())
		if err != nil {
			return err
		}

		var aks []keeperInfo

		for _, ki := range kItems {
			if ki.PledgeMoney.Sign() <= 0 {
				continue
			}

			kaddr, err := address.GetAddressFromID(ki.KeeperID)
			if err != nil {
				continue
			}

			kinfo := keeperInfo{
				Address:     kaddr.String(),
				PledgeMoney: ki.PledgeMoney,
				PledgeTime:  time.Unix(ki.StartTime, 0).In(time.Local).Format(utils.SHOWTIME),
				Online:      n.Data.FastConnect(req.Context, ki.KeeperID),
			}
			aks = append(aks, kinfo)
		}

		output := &allKeepers{
			KeeperCount: len(aks),
			PledgeMoney: pledge,
			KeeperInfos: aks,
		}

		return cmds.EmitOnce(res, output)
	},
	Type: allKeepers{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, output *allKeepers) error {
			marshaled, err := json.MarshalIndent(output, "", "\t")
			if err != nil {
				return err
			}
			fmt.Fprintln(w, string(marshaled))
			return nil
		}),
	},
}

var proCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "list all providers",
		ShortDescription: `list all providers on chain
	`,
	},

	Arguments: []cmds.Argument{},
	Options:   []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		n, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		pItems, pledge, err := role.GetAllProviders(n.Identity.Pretty())
		if err != nil {
			return err
		}

		var aks []proInfo
		for _, ki := range pItems {
			if ki.PledgeMoney.Sign() <= 0 {
				continue
			}

			kaddr, err := address.GetAddressFromID(ki.ProviderID)
			if err != nil {
				continue
			}

			kinfo := proInfo{
				Address:     kaddr.String(),
				PledgeMoney: ki.PledgeMoney,
				PledgeTime:  time.Unix(ki.StartTime, 0).In(time.Local).Format(utils.SHOWTIME),
				Online:      n.Data.Connect(req.Context, ki.ProviderID),
			}
			aks = append(aks, kinfo)
		}

		output := &allProviders{
			ProviderCount: len(aks),
			PledgeMoney:   pledge,
			ProInfos:      aks,
		}

		return cmds.EmitOnce(res, output)
	},
	Type: allProviders{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, output *allProviders) error {
			marshaled, err := json.MarshalIndent(output, "", "\t")
			if err != nil {
				return err
			}
			fmt.Fprintln(w, string(marshaled))
			return nil
		}),
	},
}
