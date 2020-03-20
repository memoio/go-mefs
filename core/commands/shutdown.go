package commands

import (
	cmds "github.com/ipfs/go-ipfs-cmds"
	cmdenv "github.com/memoio/go-mefs/core/commands/cmdenv"
)

var DaemonShutdownCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Shut down the mefs daemon",
	},
	Run: func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {
		nd, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if nd.LocalMode() {
			return cmds.Errorf(cmds.ErrClient, "daemon not running")
		}

		if err := nd.Process().Close(); err != nil {
			log.Error("error while shutting down mefs daemon:", err)
		}

		return nil
	},
}
