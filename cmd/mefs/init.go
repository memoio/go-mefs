package mefs

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"

	cmds "github.com/ipfs/go-ipfs-cmds"
	oldcmds "github.com/memoio/go-mefs/commands"
	config "github.com/memoio/go-mefs/config"
	cmdenv "github.com/memoio/go-mefs/core/commands/cmdenv"
	id "github.com/memoio/go-mefs/crypto/identity"
	fsrepo "github.com/memoio/go-mefs/repo/fsrepo"
	"github.com/memoio/go-mefs/utils"
)

const (
	passwordKwd  = "password"
	secretKeyKwd = "secretKey"
	netKeyKwd    = "netKey"
)

var (
	errRepoExists = errors.New("mefs configuration file already exists, reinitializing would overwrite your keys")
)

// InitCmd inits
var InitCmd = &cmds.Command{
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
		cmds.FileArg("config", false, false, "Initialize with the given configuration.").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.StringOption(passwordKwd, "pwd", "the password is used to load/store the PrivateKey from keyfile").WithDefault(""),
		cmds.StringOption(secretKeyKwd, "sk", "the privateKey").WithDefault(""),
		cmds.StringOption("keyfile", "kf", "the absolute path of keyfile").WithDefault(""),
		cmds.StringOption(netKeyKwd, "the netKey is used to setup private network").WithDefault("dev"),
	},
	PreRun: func(req *cmds.Request, env cmds.Environment) error {
		cctx := env.(*oldcmds.Context)
		daemonLocked, err := fsrepo.LockedByOtherProcess(cctx.ConfigRoot)
		if err != nil {
			return err
		}

		log.Println("checking if daemon is running...")
		if daemonLocked {
			log.Println("mefs-user daemon is running")
			e := "mefs-user daemon is running. please stop it to run this command"
			return cmds.ClientError(e)
		}

		return nil
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		cctx := env.(*oldcmds.Context)
		if cctx.Online {
			return cmds.Error{Message: "init must be run offline only"}
		}

		err := CheckRepo(cctx.ConfigRoot)
		if err != nil {
			return err
		}

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

		kf, ok := req.Options["keyfile"].(string)
		if ok && kf != "" {
			password, ok := req.Options[passwordKwd].(string)
			if !ok || password == "" {
				password, _ = utils.GetPassWord()
			}
			hexsk, err := id.GetPrivateKey("", password, kf)
			if err == nil {
				return DoInit(os.Stdout, cctx.ConfigRoot, password, conf, hexsk, netKey)
			}

			fmt.Println("load keyfile fails:", err)
			fmt.Println("====== manually input privatekey and password =====")
		}

		hexsk, ok := req.Options[secretKeyKwd].(string)
		if !ok || hexsk == "" {
			fmt.Printf("input your private key: ")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			go func() {
				defer cancel()
				input := bufio.NewScanner(os.Stdin)
				ok := input.Scan()
				if ok {
					hexsk = input.Text()
				}
			}()

			select {
			case <-ctx.Done():
			}
		}
		fmt.Printf("\n")

		password, ok := req.Options[passwordKwd].(string)
		if !ok || len(password) < 8 {
			fmt.Println("Password is too short, length should be at least 8")
			retry := 0
			for {
				gotpwd, err := utils.GetPassWord()
				if err != nil {
					if retry > 2 {
						return err
					}
					retry++
					continue
				}
				password = gotpwd
				break
			}
		}

		return DoInit(os.Stdout, cctx.ConfigRoot, password, conf, hexsk, netKey)
	},
}

func CheckRepo(repoRoot string) error {
	if err := checkWritable(repoRoot); err != nil {
		return err
	}

	if fsrepo.IsInitialized(repoRoot) {
		fmt.Println(repoRoot, "already exists, reinitializing need to delete files in this directory")
		fmt.Printf("delete (y/N): ")
		var deleteCmd string
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		go func() {
			defer cancel()
			input := bufio.NewScanner(os.Stdin)
			ok := input.Scan()
			if ok {
				deleteCmd = input.Text()
			}
		}()

		select {
		case <-ctx.Done():
		}

		if deleteCmd == "y" {
			return os.Rename(repoRoot, repoRoot+".bak")
			//return os.RemoveAll(repoRoot)
		}
		return errRepoExists
	}

	return nil
}

func DoInit(out io.Writer, repoRoot string, password string, conf *config.Config, prikey, netKey string) error {
	if _, err := fmt.Fprintf(out, "initializing MEFS node at %s\n", repoRoot); err != nil {
		return err
	}

	if err := checkWritable(repoRoot); err != nil {
		return err
	}

	if fsrepo.IsInitialized(repoRoot) {
		pub, err := id.GetPubByte(prikey)
		if err != nil {
			return err
		}
		pid, err := id.GetIDFromPubKey(pub)
		if err != nil {
			return err
		}

		err = fsrepo.PutPrivateKeyToKeystore(prikey, pid, password)
		if err != nil {
			return err
		}
		return errRepoExists
	}

	if conf == nil {
		var err error
		switch netKey {
		case "testnet":
			conf, prikey, err = config.InitTestnet(out, prikey)
			if err != nil {
				return err
			}
		default:
			conf, prikey, err = config.Init(out, prikey)
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
