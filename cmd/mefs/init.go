package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	cmds "github.com/ipfs/go-ipfs-cmds"
	oldcmds "github.com/memoio/go-mefs/commands"
	config "github.com/memoio/go-mefs/config"
	cmdenv "github.com/memoio/go-mefs/core/commands/cmdenv"
	fsrepo "github.com/memoio/go-mefs/repo/fsrepo"
	"github.com/memoio/go-mefs/utils"
)

const (
	nBitsForKeypairDefault = 2048
)

var initCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Initializes mefs config file.",
		ShortDescription: `
Initializes mefs configuration files and generates a new keypair.

mefs uses a repository in the local file system. By default, the repo is
located at ~/.mefs. To change the repo location, set the $MEFS_PATH
environment variable:

    export MEFS_PATH=/path/to/mefsrepo
`,
	},
	Arguments: []cmds.Argument{
		cmds.FileArg("default-config", false, false, "Initialize with the given configuration.").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.StringOption(passwordKwd, "pwd", "the password is used to encrypt the privateKey").WithDefault(utils.DefaultPassword),
		cmds.StringOption(secretKeyKwd, "sk", "the stored privateKey").WithDefault(""),
		cmds.StringOption(netKeyKwd, "the netKey is used to setup private network").WithDefault("dev"),
	},
	PreRun: func(req *cmds.Request, env cmds.Environment) error {
		cctx := env.(*oldcmds.Context)
		daemonLocked, err := fsrepo.LockedByOtherProcess(cctx.ConfigRoot)
		if err != nil {
			return err
		}

		log.Info("checking if daemon is running...")
		if daemonLocked {
			log.Debug("mefs daemon is running")
			e := "mefs daemon is running. please stop it to run this command"
			return cmds.ClientError(e)
		}

		return nil
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cctx := env.(*oldcmds.Context)
		if cctx.Online {
			return cmds.Error{Message: "init must be run offline only"}
		}

		hexsk, _ := req.Options[secretKeyKwd].(string)
		password, _ := req.Options[passwordKwd].(string)
		netKey, _ := req.Options[netKeyKwd].(string)

		var conf *config.Config

		f := req.Files
		if f != nil {
			confFile, err := cmdenv.GetFileArg(req.Files.Entries())
			if err != nil {
				return err
			}

			conf = &config.Config{}
			if err := json.NewDecoder(confFile).Decode(conf); err != nil {
				return err
			}
		}

		return doInit(os.Stdout, cctx.ConfigRoot, nBitsForKeypairDefault, password, conf, hexsk, netKey)
	},
}

func doInit(out io.Writer, repoRoot string, nBitsForKeypair int, password string, conf *config.Config, prikey, netKey string) error {
	if _, err := fmt.Fprintf(out, "initializing MEFS node at %s\n", repoRoot); err != nil {
		return err
	}

	if err := checkWritable(repoRoot); err != nil {
		return err
	}

	if fsrepo.IsInitialized(repoRoot) {
		return errRepoExists
	}

	if conf == nil {
		var err error
		switch netKey {
		case "testnet":
			conf, prikey, err = config.InitTestnet(out, nBitsForKeypair, prikey)
			if err != nil {
				return err
			}
		default:
			conf, prikey, err = config.Init(out, nBitsForKeypair, prikey)
			if err != nil {
				return err
			}
		}

	}

	if err := fsrepo.Init(repoRoot, conf, prikey, password); err != nil {
		return err
	}

	_, err := fsrepo.Open(repoRoot)
	if err != nil {
		fmt.Println("fsrepo.Open falied: ", err)
	}
	return nil
}

func checkWritable(dir string) error {
	_, err := os.Stat(dir)
	if err == nil {
		// dir exists, make sure we can write to it
		testfile := path.Join(dir, "test")
		fi, err := os.Create(testfile)
		if err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("%s is not writeable by the current user", dir)
			}
			return fmt.Errorf("unexpected error while checking writeablility of repo root: %s", err)
		}
		err = fi.Close()
		if err != nil {
			fmt.Println("fi.Close() falied: ", err)
		}
		return os.Remove(testfile)
	}

	if os.IsNotExist(err) {
		// dir doesn't exist, check that we can create it
		return os.Mkdir(dir, 0775)
	}

	if os.IsPermission(err) {
		return fmt.Errorf("cannot write to %s, incorrect permissions", err)
	}

	return err
}
