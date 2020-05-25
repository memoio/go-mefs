package commands

import (
	"os"
	"path"
	"runtime"

	cmds "github.com/ipfs/go-ipfs-cmds"
	version "github.com/memoio/go-mefs"
	config "github.com/memoio/go-mefs/config"
	cmdenv "github.com/memoio/go-mefs/core/commands/cmdenv"
	manet "github.com/multiformats/go-multiaddr-net"
	sysi "github.com/whyrusleeping/go-sysinfo"
)

var SysDiagCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Print system diagnostic information.",
		ShortDescription: `
Prints out information about your computer to aid in easier debugging.
`,
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		info := make(map[string]interface{})
		err := runtimeInfo(info)
		if err != nil {
			return err
		}

		err = envVarInfo(info)
		if err != nil {
			return err
		}

		err = diskSpaceInfo(info)
		if err != nil {
			return err
		}

		err = memInfo(info)
		if err != nil {
			return err
		}
		nd, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		err = netInfo(nd.OnlineMode(), info)
		if err != nil {
			return err
		}

		info["mefs_version"] = version.CurrentVersionNumber
		info["mefs_commit"] = version.CurrentCommit
		return cmds.EmitOnce(res, info)
	},
}

func runtimeInfo(out map[string]interface{}) error {
	rt := make(map[string]interface{})
	rt["os"] = runtime.GOOS
	rt["arch"] = runtime.GOARCH
	rt["compiler"] = runtime.Compiler
	rt["version"] = runtime.Version()
	rt["numcpu"] = runtime.NumCPU()
	rt["gomaxprocs"] = runtime.GOMAXPROCS(0)
	rt["numgoroutines"] = runtime.NumGoroutine()

	out["runtime"] = rt
	return nil
}

func envVarInfo(out map[string]interface{}) error {
	ev := make(map[string]interface{})
	ev["GOPATH"] = os.Getenv("GOPATH")
	ev["MEFS_PATH"] = os.Getenv("MEFS_PATH")

	out["environment"] = ev
	return nil
}

func mefsPath() string {
	p := os.Getenv("MEFS_PATH")
	if p == "" {
		p = path.Join(os.Getenv("HOME"), config.DefaultPathName)
	}
	return p
}

func diskSpaceInfo(out map[string]interface{}) error {
	di := make(map[string]interface{})
	dinfo, err := sysi.DiskUsage(mefsPath())
	if err != nil {
		return err
	}

	di["fstype"] = dinfo.FsType
	di["total_space"] = dinfo.Total
	di["free_space"] = dinfo.Free

	out["diskinfo"] = di
	return nil
}

func memInfo(out map[string]interface{}) error {
	m := make(map[string]interface{})

	meminf, err := sysi.MemoryInfo()
	if err != nil {
		return err
	}

	m["swap"] = meminf.Swap
	m["virt"] = meminf.Used
	out["memory"] = m
	return nil
}

func netInfo(online bool, out map[string]interface{}) error {
	n := make(map[string]interface{})
	addrs, err := manet.InterfaceMultiaddrs()
	if err != nil {
		return err
	}

	var straddrs []string
	for _, a := range addrs {
		straddrs = append(straddrs, a.String())
	}

	n["interface_addresses"] = straddrs
	n["online"] = online
	out["net"] = n
	return nil
}
