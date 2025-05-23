package main

import (
	"fmt"

	cmds "github.com/ipfs/go-ipfs-cmds"
	minit "github.com/memoio/go-mefs/cmd/mefs"
	"github.com/memoio/go-mefs/core/commands"
	newcmd "github.com/memoio/go-mefs/userNode/commands"
)

// Root is the CLI root, used for executing commands accessible to CLI clients.
// Some subcommands (like 'mefs daemon' or 'mefs init') are only accessible here,
// and can't be called through the HTTP API.
var Root = &cmds.Command{
	Options:  newcmd.Root.Options,
	Helptext: newcmd.Root.Helptext,
}

// commandsClientCmd is the "mefs commands" command for local cli
var commandsClientCmd = commands.CommandsCmd(Root)

// Commands in localCommands should always be run locally (even if daemon is running).
// They can override subcommands in commands.Root by defining a subcommand with the same name.
var localCommands = map[string]*cmds.Command{
	"daemon":   daemonCmd,
	"init":     minit.InitCmd,
	"commands": commandsClientCmd,
}

func init() {
	// setting here instead of in literal to prevent initialization loop
	// (some commands make references to Root)
	Root.Subcommands = localCommands

	for k, v := range newcmd.Root.Subcommands {
		if _, found := Root.Subcommands[k]; !found {
			Root.Subcommands[k] = v
		}
	}
}

// NB: when necessary, properties are described using negatives in order to
// provide desirable defaults
type cmdDetails struct {
	cannotRunOnClient bool
	cannotRunOnDaemon bool
	doesNotUseRepo    bool

	// doesNotUseConfigAsInput describes commands that do not use the config as
	// input. These commands either initialize the config or perform operations
	// that don't require access to the config.
	//
	// pre-command hooks that require configs must not be run before these
	// commands.
	doesNotUseConfigAsInput bool

	// preemptsAutoUpdate describes commands that must be executed without the
	// auto-update pre-command hook
	preemptsAutoUpdate bool
}

func (d *cmdDetails) String() string {
	return fmt.Sprintf("on client? %t, on daemon? %t, uses repo? %t",
		d.canRunOnClient(), d.canRunOnDaemon(), d.usesRepo())
}

func (d *cmdDetails) Loggable() map[string]interface{} {
	return map[string]interface{}{
		"canRunOnClient":     d.canRunOnClient(),
		"canRunOnDaemon":     d.canRunOnDaemon(),
		"preemptsAutoUpdate": d.preemptsAutoUpdate,
		"usesConfigAsInput":  d.usesConfigAsInput(),
		"usesRepo":           d.usesRepo(),
	}
}

func (d *cmdDetails) usesConfigAsInput() bool { return !d.doesNotUseConfigAsInput }
func (d *cmdDetails) canRunOnClient() bool    { return !d.cannotRunOnClient }
func (d *cmdDetails) canRunOnDaemon() bool    { return !d.cannotRunOnDaemon }
func (d *cmdDetails) usesRepo() bool          { return !d.doesNotUseRepo }

// "What is this madness!?" you ask. Our commands have the unfortunate problem of
// not being able to run on all the same contexts. This map describes these
// properties so that other code can make decisions about whether to invoke a
// command or return an error to the user.
var cmdDetailsMap = map[string]cmdDetails{
	"init":        {doesNotUseConfigAsInput: true, cannotRunOnDaemon: true, doesNotUseRepo: true},
	"daemon":      {doesNotUseConfigAsInput: true, cannotRunOnDaemon: true},
	"commands":    {doesNotUseRepo: true},
	"version":     {doesNotUseConfigAsInput: true, doesNotUseRepo: true}, // must be permitted to run before init
	"log":         {cannotRunOnClient: true},
	"diag/cmds":   {cannotRunOnClient: true},
	"repo/fsck":   {cannotRunOnDaemon: true},
	"config/edit": {cannotRunOnDaemon: true, doesNotUseRepo: true},
	"cid":         {doesNotUseRepo: true},
}
