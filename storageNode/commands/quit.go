package commands

import (
	"fmt"

	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/memoio/go-mefs/core/commands/cmdenv"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/storageNode/provider"
)

//QuitCmd provider quit group
var QuitCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "provider quit group",
		ShortDescription: "provider run this to exit the specified storage service",
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("groupID", true, false, "group's ID that provider want to exit"),
	},

	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		groupID := req.Arguments[0]

		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if !node.OnlineMode() {
			return ErrNotOnline
		}

		providerIns, ok := node.Inst.(*provider.Info)
		if !ok {
			return role.ErrServiceNotReady
		}

		//moveData: 1. keepers send metavalue and another proID to provider
		//2. provider move its data to another provider;
		fmt.Println("begin move data to other provider")
		response, err := providerIns.MoveData(req.Context, groupID)
		if err != nil {
			return err
		}
		fmt.Println("data has been moved successfully")

		//setProviderStop:
		//1. provider send 'quit' message to keepers, so keepers set provider stop in upkeeping-contract.
		//2. keepers update their groupInfo, delete this provider specifically.
		fmt.Println("begin send 'quit' message to keepers")
		err = providerIns.SetProviderStop(req.Context, groupID, response)
		if err != nil {
			return err
		}

		fmt.Println("quit successfully")
		return nil
	},
}
