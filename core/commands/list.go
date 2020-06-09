package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"sync"
	"time"

	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/memoio/go-mefs/core/commands/cmdenv"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

type allKeepers struct {
	KeeperCount    int
	PledgeMoney    string
	OnlineCount    int
	OnlineKepepers []keeperInfo
	OfflineCount   int
	OfflineKeepers []keeperInfo
}

type keeperInfo struct {
	Address     string
	Online      bool
	PledgeMoney string
	PledgeTime  string
}

type allProviders struct {
	ProviderCount    int
	PledgeBytes      string
	OnlineCount      int
	OnlineProviders  []proInfo
	OfflineCount     int
	OfflineProviders []proInfo
}

type proInfo struct {
	Address     string
	Online      bool
	PledgeBytes string
	PledgeMoney string
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

		fmt.Println("has keepers:", len(kItems))
		var wg sync.WaitGroup
		for _, ki := range kItems {
			if ki.PledgeMoney.Sign() <= 0 {
				continue
			}

			wg.Add(1)
			go func(kid string) {
				defer wg.Done()
				n.Data.Connect(req.Context, kid)
			}(ki.KeeperID)
		}

		wg.Wait()

		var ons, offs []keeperInfo

		price, err := role.GetKeeperPrice(n.Identity.Pretty())
		if err != nil {
			return err
		}

		for _, ki := range kItems {
			if ki.PledgeMoney.Cmp(price) < 0 {
				continue
			}

			kaddr, err := address.GetAddressFromID(ki.KeeperID)
			if err != nil {
				continue
			}

			if n.Data.FastConnect(req.Context, ki.KeeperID) {
				kinfo := keeperInfo{
					Address:     kaddr.String(),
					PledgeMoney: utils.FormatWei(ki.PledgeMoney),
					PledgeTime:  time.Unix(ki.StartTime, 0).In(time.Local).Format(utils.SHOWTIME),
					Online:      true,
				}
				ons = append(ons, kinfo)
			} else {
				kinfo := keeperInfo{
					Address:     kaddr.String(),
					PledgeMoney: utils.FormatWei(ki.PledgeMoney),
					PledgeTime:  time.Unix(ki.StartTime, 0).In(time.Local).Format(utils.SHOWTIME),
					Online:      false,
				}
				offs = append(offs, kinfo)
			}

		}

		output := &allKeepers{
			KeeperCount:    len(ons) + len(offs),
			PledgeMoney:    utils.FormatWei(pledge),
			OnlineCount:    len(ons),
			OnlineKepepers: ons,
			OfflineCount:   len(offs),
			OfflineKeepers: offs,
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

		fmt.Println("has providers:", len(pItems))

		var wg sync.WaitGroup
		for _, ki := range pItems {
			if ki.PledgeMoney.Sign() <= 0 {
				continue
			}

			wg.Add(1)
			go func(kid string) {
				defer wg.Done()
				n.Data.Connect(req.Context, kid)
			}(ki.ProviderID)
		}

		wg.Wait()
		var ons, offs []proInfo
		for _, ki := range pItems {
			if ki.PledgeMoney.Sign() <= 0 {
				continue
			}

			if ki.Capacity <= 0 {
				continue
			}

			kaddr, err := address.GetAddressFromID(ki.ProviderID)
			if err != nil {
				continue
			}

			if n.Data.FastConnect(req.Context, ki.ProviderID) {
				kinfo := proInfo{
					Address:     kaddr.String(),
					PledgeMoney: utils.FormatWei(ki.PledgeMoney),
					PledgeTime:  time.Unix(ki.StartTime, 0).In(time.Local).Format(utils.SHOWTIME),
					Online:      true,
					PledgeBytes: utils.FormatBytes(ki.Capacity * 1024 * 1024),
				}
				ons = append(ons, kinfo)
			} else {
				kinfo := proInfo{
					Address:     kaddr.String(),
					PledgeMoney: utils.FormatWei(ki.PledgeMoney),
					PledgeTime:  time.Unix(ki.StartTime, 0).In(time.Local).Format(utils.SHOWTIME),
					Online:      false,
					PledgeBytes: utils.FormatBytes(ki.Capacity * 1024 * 1024),
				}
				offs = append(offs, kinfo)
			}

		}

		pledge.Mul(pledge, big.NewInt(1024*1024))
		output := &allProviders{
			ProviderCount:    len(ons) + len(offs),
			PledgeBytes:      utils.FormatBytes(pledge.Int64()),
			OnlineCount:      len(ons),
			OnlineProviders:  ons,
			OfflineCount:     len(offs),
			OfflineProviders: offs,
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
