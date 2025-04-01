package commands

import (
	"errors"
	"fmt"
	"io"
	"runtime"

	cmds "github.com/ipfs/go-ipfs-cmds"
	logging "github.com/ipfs/go-log"
	version "github.com/memoio/go-mefs"
	fsrepo "github.com/memoio/go-mefs/repo/fsrepo"
)

var log = logging.Logger("core/commands")

var ErrNotOnline = errors.New("this command must be run in online mode. Try running 'mefs daemon' first")

type MessageOutput struct {
	Message string
}

type VersionOutput struct {
	Version string
	Commit  string
	Repo    string
	System  string
	Golang  string
}

const (
	versionNumberOptionName = "number"
	versionCommitOptionName = "commit"
	versionRepoOptionName   = "repo"
	versionAllOptionName    = "all"
)

var VersionCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "Show mefs version information.",
		ShortDescription: "Returns the current version of mefs and exits.",
	},

	Options: []cmds.Option{
		cmds.BoolOption(versionNumberOptionName, "n", "Only show the version number."),
		cmds.BoolOption(versionCommitOptionName, "Show the commit hash."),
		cmds.BoolOption(versionRepoOptionName, "Show repo version."),
		cmds.BoolOption(versionAllOptionName, "Show all version information"),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		return cmds.EmitOnce(res, &VersionOutput{
			Version: version.CurrentVersionNumber,
			Commit:  version.CurrentCommit,
			Repo:    fmt.Sprint(fsrepo.RepoVersion),
			System:  runtime.GOARCH + "/" + runtime.GOOS, //TODO: Precise version here
			Golang:  runtime.Version(),
		})
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, version *VersionOutput) error {
			commit, _ := req.Options[versionCommitOptionName].(bool)
			commitTxt := ""
			if commit {
				commitTxt = "-" + version.Commit
			}

			all, _ := req.Options[versionAllOptionName].(bool)
			if all {
				out := fmt.Sprintf("go-mefs version: %s-%s\n"+
					"Repo version: %s\nSystem version: %s\nGolang version: %s\n",
					version.Version, version.Commit, version.Repo, version.System, version.Golang)
				fmt.Fprint(w, out)
				return nil
			}

			repo, _ := req.Options[versionRepoOptionName].(bool)
			if repo {
				fmt.Fprintln(w, version.Repo)
				return nil
			}

			number, _ := req.Options[versionNumberOptionName].(bool)
			if number {
				fmt.Fprintln(w, version.Version+commitTxt)
				return nil
			}

			fmt.Fprint(w, fmt.Sprintf("mefs version %s%s\n", version.Version, commitTxt))
			return nil
		}),
	},
	Type: VersionOutput{},
}
