package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"sync"
	"syscall"
	"time"

	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/ipfs/go-ipfs-cmds/cli"
	"github.com/ipfs/go-ipfs-cmds/http"
	u "github.com/ipfs/go-ipfs-util"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-core/transport"
	loggables "github.com/libp2p/go-libp2p-loggables"
	"github.com/libp2p/go-libp2p-swarm"
	oldcmds "github.com/memoio/go-mefs/commands"
	config "github.com/memoio/go-mefs/config"
	"github.com/memoio/go-mefs/core"
	"github.com/memoio/go-mefs/plugin/loader"
	"github.com/memoio/go-mefs/repo"
	"github.com/memoio/go-mefs/repo/fsrepo"
	newcmd "github.com/memoio/go-mefs/userNode/commands"
	"github.com/memoio/go-mefs/userNode/corehttp"
	"github.com/memoio/go-mefs/utils"
	ma "github.com/multiformats/go-multiaddr"
	madns "github.com/multiformats/go-multiaddr-dns"
	manet "github.com/multiformats/go-multiaddr-net"
)

func init() {
	swarm.DialTimeoutLocal = transport.DialTimeout
}

// log is the command logger
var log = logging.Logger("cmd/mefs-user")

// declared as a var for testing purposes
var dnsResolver = madns.DefaultResolver

const (
	EnvEnableProfiling = "MEFS_PROF"
	cpuProfile         = "mefs.cpuprof"
	heapProfile        = "mefs.memprof"
)

// main roadmap:
// - parse the commandline to get a cmdInvocation
// - if user requests help, print it and exit.
// - run the command invocation
// - output the response
// - if anything fails, print error, maybe with help
func main() {
	os.Exit(mainRet())
}

func mainRet() int {
	rand.Seed(time.Now().UnixNano())
	ctx := logging.ContextWithLoggable(context.Background(), loggables.Uuid("session"))
	var err error

	// we'll call this local helper to output errors.
	// this is so we control how to print errors in one place.
	printErr := func(err error) {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	}

	stopFunc, err := profileIfEnabled()
	if err != nil {
		printErr(err)
		return 1
	}
	defer stopFunc() // to be executed as late as possible

	intrh, ctx := setupInterruptHandler(ctx)
	defer intrh.Close()

	// Handle `mefs version` or `mefs help`
	if len(os.Args) > 1 {
		// Handle `mefs --version'
		if os.Args[1] == "--version" {
			os.Args[1] = "version"
		}

		//Handle `mefs help` and `mefs help <sub-command>`
		if os.Args[1] == "help" {
			if len(os.Args) > 2 {
				os.Args = append(os.Args[:1], os.Args[2:]...)
				// Handle `mefs help --help`
				// append `--help`,when the command is not `mefs help --help`
				if os.Args[1] != "--help" {
					os.Args = append(os.Args, "--help")
				}
			} else {
				os.Args[1] = "--help"
			}
		}
	}

	// output depends on executable name passed in os.Args
	// so we need to make sure it's stable
	os.Args[0] = "mefs-user"

	buildEnv := func(ctx context.Context, req *cmds.Request) (cmds.Environment, error) {
		checkDebug(req)
		repoPath, err := getRepoPath(req)
		if err != nil {
			return nil, err
		}
		log.Debugf("config path is %s", repoPath)

		// this sets up the function that will initialize the node
		// this is so that we can construct the node lazily.
		return &oldcmds.Context{
			ConfigRoot: repoPath,
			LoadConfig: loadConfig,
			ReqLog:     &oldcmds.ReqLog{},
			ConstructNode: func() (n *core.MefsNode, err error) {
				if req == nil {
					return nil, errors.New("constructing node without a request")
				}

				r, err := fsrepo.Open(repoPath)
				if err != nil { // repo is owned by the node
					return nil, err
				}

				// ok everything is good. set it on the invocation (for ownership)
				// and return it.
				n, err = core.NewNode(ctx, &core.BuildCfg{
					Repo: r,
				}, "", utils.DefaultPassword, "")
				if err != nil {
					return nil, err
				}

				n.SetLocal(true)
				return n, nil
			},
		}, nil
	}

	err = cli.Run(ctx, Root, os.Args, os.Stdin, os.Stdout, os.Stderr, buildEnv, makeExecutor)
	if err != nil {
		return 1
	}

	// everything went better than expected :)
	return 0
}

func checkDebug(req *cmds.Request) {
	// check if user wants to debug. option OR env var.
	debug, _ := req.Options["debug"].(bool)
	if debug || os.Getenv("MEFS_LOGGING") == "debug" {
		u.Debug = true
		logging.SetDebugLogging()
	}
	if u.GetenvBool("DEBUG") {
		u.Debug = true
	}
}

func apiAddrOption(req *cmds.Request) (ma.Multiaddr, error) {
	apiAddrStr, apiSpecified := req.Options[newcmd.ApiOption].(string)
	if !apiSpecified {
		return nil, nil
	}
	return ma.NewMultiaddr(apiAddrStr)
}

// loadPlugins init datastore
func loadPlugins(cctx *oldcmds.Context) error {
	pluginpath := filepath.Join(cctx.ConfigRoot, "plugins")

	// check if repo is accessible before loading plugins
	ok, err := checkPermissions(cctx.ConfigRoot)
	if err != nil {
		return err
	}
	if !ok {
		pluginpath = ""
	}
	if _, err := loader.LoadPlugins(pluginpath); err != nil {
		log.Error("error loading plugins: ", err)
	}
	return nil
}

func makeExecutor(req *cmds.Request, env interface{}) (cmds.Executor, error) {
	exe := cmds.NewExecutor(req.Root)
	cctx := env.(*oldcmds.Context)

	loadPlugins(cctx)

	details := commandDetails(req.Path)

	// Check if the command is disabled.
	if details.cannotRunOnClient && details.cannotRunOnDaemon {
		return nil, fmt.Errorf("command disabled: %v", req.Path)
	}

	// Can we just run this locally?
	if !details.cannotRunOnClient && details.doesNotUseRepo {
		return exe, nil
	}

	// Get the API option from the commandline.
	apiAddr, err := apiAddrOption(req)
	if err != nil {
		return nil, err
	}

	// Require that the command be run on the daemon when the API flag is
	// passed (unless we're trying to _run_ the daemon).
	daemonRequested := apiAddr != nil && req.Command != daemonCmd

	// Run this on the client if required.
	if details.cannotRunOnDaemon || req.Command.External {
		if daemonRequested {
			// User requested that the command be run on the daemon but we can't.
			// NOTE: We drop this check for the `mefs daemon` command.
			return nil, errors.New("api flag specified but command cannot be run on the daemon")
		}
		return exe, nil
	}

	// Finally, look in the repo for an API file.
	if apiAddr == nil {
		var err error
		apiAddr, err = fsrepo.APIAddr(cctx.ConfigRoot)
		switch err {
		case nil, repo.ErrApiNotRunning:
		default:
			return nil, err
		}
	}

	// Still no api specified? Run it on the client or fail.
	if apiAddr == nil {
		if details.cannotRunOnClient {
			return nil, fmt.Errorf("command must be run on the daemon: %v", req.Path)
		}
		return exe, nil
	}

	// Resolve the API addr.
	apiAddr, err = resolveAddr(req.Context, apiAddr)
	if err != nil {
		return nil, err
	}
	_, host, err := manet.DialArgs(apiAddr)
	if err != nil {
		return nil, err
	}

	// Construct the executor.
	opts := []http.ClientOpt{
		http.ClientWithAPIPrefix(corehttp.APIPath),
	}

	// Fallback on a local executor if we (a) have a repo and (b) aren't
	// forcing a daemon.
	if !daemonRequested && fsrepo.IsInitialized(cctx.ConfigRoot) {
		opts = append(opts, http.ClientWithFallback(exe))
	}

	return http.NewClient(host, opts...), nil
}

func checkPermissions(path string) (bool, error) {
	_, err := os.Open(path)
	if os.IsNotExist(err) {
		// repo does not exist yet - don't load plugins, but also don't fail
		return false, nil
	}
	if os.IsPermission(err) {
		// repo is not accessible. error out.
		return false, fmt.Errorf("error opening repository at %s: permission denied", path)
	}

	return true, nil
}

// commandDetails returns a command's details for the command given by |path|.
func commandDetails(path []string) cmdDetails {
	var details cmdDetails
	// find the last command in path that has a cmdDetailsMap entry
	for i := range path {
		if cmdDetails, found := cmdDetailsMap[strings.Join(path[:i+1], "/")]; found {
			details = cmdDetails
		}
	}
	return details
}

func getRepoPath(req *cmds.Request) (string, error) {
	repoOpt, found := req.Options["config"].(string)
	if found && repoOpt != "" {
		return repoOpt, nil
	}

	repoPath, err := fsrepo.BestKnownPath()
	if err != nil {
		return "", err
	}
	return repoPath, nil
}

func loadConfig(path string) (*config.Config, error) {
	return fsrepo.ConfigAt(path)
}

// startProfiling begins CPU profiling and returns a `stop` function to be
// executed as late as possible. The stop function captures the memprofile.
func startProfiling() (func(), error) {
	// start CPU profiling as early as possible
	ofi, err := os.Create(cpuProfile)
	if err != nil {
		return nil, err
	}
	err = pprof.StartCPUProfile(ofi)
	if err != nil {
		fmt.Println("pprof.StartCPUProfile(ofi) falied: ", err)
	}
	go func() {
		for range time.NewTicker(time.Second * 30).C {
			err := writeHeapProfileToFile()
			if err != nil {
				log.Error(err)
			}
		}
	}()

	stopProfiling := func() {
		pprof.StopCPUProfile()
		err = ofi.Close() // captured by the closure
		if err != nil {
			fmt.Println("ofi.Close() falied: ", err)
		}
	}
	return stopProfiling, nil
}

func writeHeapProfileToFile() error {
	mprof, err := os.Create(heapProfile)
	if err != nil {
		return err
	}
	defer mprof.Close() // _after_ writing the heap profile
	return pprof.WriteHeapProfile(mprof)
}

// IntrHandler helps set up an interrupt handler that can
// be cleanly shut down through the io.Closer interface.
type IntrHandler struct {
	sig chan os.Signal
	wg  sync.WaitGroup
}

func NewIntrHandler() *IntrHandler {
	ih := &IntrHandler{}
	ih.sig = make(chan os.Signal, 1)
	return ih
}

func (ih *IntrHandler) Close() error {
	close(ih.sig)
	ih.wg.Wait()
	return nil
}

// Handle starts handling the given signals, and will call the handler
// callback function each time a signal is catched. The function is passed
// the number of times the handler has been triggered in total, as
// well as the handler itself, so that the handling logic can use the
// handler's wait group to ensure clean shutdown when Close() is called.
func (ih *IntrHandler) Handle(handler func(count int, ih *IntrHandler), sigs ...os.Signal) {
	signal.Notify(ih.sig, sigs...)
	ih.wg.Add(1)
	go func() {
		defer ih.wg.Done()
		count := 0
		for s := range ih.sig {
			fmt.Println("exit signal:", s)
			count++
			handler(count, ih)
		}
		signal.Stop(ih.sig)
	}()
}

func setupInterruptHandler(ctx context.Context) (io.Closer, context.Context) {
	intrh := NewIntrHandler()
	ctx, cancelFunc := context.WithCancel(ctx)

	handlerFunc := func(count int, ih *IntrHandler) {
		switch count {
		case 1:
			fmt.Println() // Prevent un-terminated ^C character in terminal

			ih.wg.Add(1)
			go func() {
				defer ih.wg.Done()
				cancelFunc()
			}()

		default:
			fmt.Println("Received another interrupt before graceful shutdown, terminating...")
			os.Exit(-1)
		}
	}

	intrh.Handle(handlerFunc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	return intrh, ctx
}

func profileIfEnabled() (func(), error) {
	// FIXME this is a temporary hack so profiling of asynchronous operations
	// works as intended.
	if os.Getenv(EnvEnableProfiling) != "" {
		stopProfilingFunc, err := startProfiling() // TODO maybe change this to its own option... profiling makes it slower.
		if err != nil {
			return nil, err
		}
		return stopProfilingFunc, nil
	}
	return func() {}, nil
}

func resolveAddr(ctx context.Context, addr ma.Multiaddr) (ma.Multiaddr, error) {
	ctx, cancelFunc := context.WithTimeout(ctx, 10*time.Second)
	defer cancelFunc()

	addrs, err := dnsResolver.Resolve(ctx, addr)
	if err != nil {
		return nil, err
	}

	if len(addrs) == 0 {
		return nil, errors.New("non-resolvable API endpoint")
	}

	return addrs[0], nil
}
