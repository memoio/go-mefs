package commands

import (
	"errors"

	cmds "github.com/ipfs/go-ipfs-cmds"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("core/commands")

var ErrNotOnline = errors.New("this command must be run in online mode. Try running 'mefs daemon' first")

const (
	ConfigOption = "config"
	DebugOption  = "debug"
	LocalOption  = "local"
	ApiOption    = "api"
)

var Root = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:  "Global p2p filesystem.",
		Synopsis: "mefs [--config=<config> | -c] [--debug=<debug> | -D] [--help=<help>] [-h=<h>] [--local=<local> | -L] [--api=<api>] <command> ...",
		Subcommands: `
BASIC COMMANDS
  init          Initialize mefs local configuration

DATA STRUCTURE COMMANDS
  block         Interact with raw blocks in the datastore

ADVANCED COMMANDS
  daemon        Start a long-running daemon process
  repo          Manipulate the MEFS repository
  stats         Various operational stats
  p2p           Libp2p stream mounting

NETWORK COMMANDS
  id            Show info about MEFS peers
  bootstrap     Add or remove bootstrap peers
  swarm         Manage connections to the p2p network
  dht           Query the DHT for values or peers
  ping          Measure the latency of a connection
  diag          Print diagnostics

TOOL COMMANDS
  config        Manage configuration
  version       Show MEfs version information
  commands      List all available commands

Use 'mefs <command> --help' to learn more about each command.

mefs uses a repository in the local file system. By default, the repo is
located at ~/.mefs. To change the repo location, set the $MEFS_PATH
environment variable:

  export MEFS_PATH=/path/to/mefsrepo

EXIT STATUS

The CLI will exit with one of the following values:

0     Successful execution.
1     Failed executions.
`,
	},
	Options: []cmds.Option{
		cmds.StringOption(ConfigOption, "c", "Path to the configuration file to use."),
		cmds.BoolOption(DebugOption, "D", "Operate in debug mode."),
		cmds.BoolOption(cmds.OptLongHelp, "Show the full command help text."),
		cmds.BoolOption(cmds.OptShortHelp, "Show a short version of the command help text."),
		cmds.BoolOption(LocalOption, "L", "Run the command locally, instead of using the daemon."),
		cmds.StringOption(ApiOption, "Use a specific API instance (defaults to /ip4/127.0.0.1/tcp/5001)"),

		// global options, added to every command
		cmds.OptionEncodingType,
		cmds.OptionStreamChannels,
		cmds.OptionTimeout,
	},
}

// commandsDaemonCmd is the "mefs commands" command for daemon
var CommandsDaemonCmd = CommandsCmd(Root)

//保存mefs一级命令的结构体
var rootSubcommands = map[string]*cmds.Command{
	"commands":  CommandsDaemonCmd,
	"block":     BlockCmd,
	"lfs":       LfsCmd,
	"keeper":    KeeperCmd,
	"stats":     StatsCmd,
	"bootstrap": BootstrapCmd,
	"config":    ConfigCmd,
	"dht":       DhtCmd,
	"id":        IDCmd,
	"log":       LogCmd,
	"swarm":     SwarmCmd,
	"version":   VersionCmd,
	"shutdown":  daemonShutdownCmd,
	"create":    createCmd,
	"diag":      sysDiagCmd,
	"test":      TestCmd,
	"contract":  ContractCmd,
	"gateway":   GatewayCmd,
}

// RootRO is the readonly version of Root
var RootRO = &cmds.Command{}

var CommandsDaemonROCmd = CommandsCmd(RootRO)

// RefsROCmd is `mefs refs` command
var RefsROCmd = &cmds.Command{}

var rootROSubcommands = map[string]*cmds.Command{
	"commands": CommandsDaemonROCmd,
	"version":  VersionCmd,
}

func init() {
	Root.ProcessHelp()
	*RootRO = *Root

	// sanitize readonly refs command
	// *RefsROCmd = *RefsCmd
	RefsROCmd.Subcommands = map[string]*cmds.Command{}

	// this was in the big map definition above before,
	// but if we leave it there lgc.NewCommand will be executed
	// before the value is updated (:/sanitize readonly refs command/)
	rootROSubcommands["refs"] = RefsROCmd

	Root.Subcommands = rootSubcommands

	RootRO.Subcommands = rootROSubcommands
}

type MessageOutput struct {
	Message string
}
