package commands

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/memoio/go-mefs/role"

	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/memoio/go-mefs/core/commands/cmdenv"
	datastore "github.com/memoio/go-mefs/source/go-datastore"
	"github.com/memoio/go-mefs/storageNode/provider"
)

type pInfoOutput struct {
	DepositCapacity uint64
	UsedCapacity    uint64
}

var InfoCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Interact with provider.",
		ShortDescription: `
'mefs-provider info' is a plumbing command used to manipulate provider service.
`,
	},

	Arguments: []cmds.Argument{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		var depositCapacity int64
		var usedCapacity uint64

		providerIns, ok := node.Inst.(*provider.Info)

		if !ok || !providerIns.Online() { //service is not ready, 从链上获取depositCapacity
			providerItem, err := role.GetProviderInfo(node.Identity.Pretty(), node.Identity.Pretty())
			if err != nil {
				return err
			}
			depositCapacity = providerItem.Capacity

			usedCapacity, err = datastore.DiskUsage(node.Data.DataStore())
			if err != nil {
				return err
			}
		} else {
			depositCapacity, usedCapacity = providerIns.GetStorageInfo()
		}

		output := &pInfoOutput{
			DepositCapacity: uint64(depositCapacity),
			UsedCapacity:    usedCapacity,
		}
		return cmds.EmitOnce(res, output)
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, out *pInfoOutput) error {
			marshaled, err := json.MarshalIndent(out, "", "\t")
			if err != nil {
				return err
			}
			marshaled = append(marshaled, byte('\n'))
			fmt.Fprintln(w, string(marshaled))
			return nil
		}),
	},
	Type: pInfoOutput{},
}
