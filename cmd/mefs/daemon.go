package main

import (
	"errors"
	_ "expvar"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"sort"
	"sync"

	cmds "github.com/ipfs/go-ipfs-cmds"
	mprome "github.com/ipfs/go-metrics-prometheus"
	version "github.com/memoio/go-mefs"
	mcl "github.com/memoio/go-mefs/bls12"
	utilmain "github.com/memoio/go-mefs/cmd/mefs/util"
	oldcmds "github.com/memoio/go-mefs/commands"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/core"
	"github.com/memoio/go-mefs/core/commands"
	"github.com/memoio/go-mefs/core/corehttp"
	"github.com/memoio/go-mefs/repo/fsrepo"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/role/keeper"
	"github.com/memoio/go-mefs/role/provider"
	"github.com/memoio/go-mefs/role/user"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	adjustFDLimitKwd          = "manage-fdlimit"
	initOptionKwd             = "init"
	offlineKwd                = "offline"
	routingOptionKwd          = "routing"
	routingOptionSupernodeKwd = "supernode"
	routingOptionDHTClientKwd = "dhtclient"
	routingOptionDHTKwd       = "dht"
	routingOptionNoneKwd      = "none"
	routingOptionDefaultKwd   = "default"
	unencryptTransportKwd     = "disable-transport-encryption"
	unrestrictedAPIAccessKwd  = "unrestricted-api"
	writableKwd               = "writable"
	enableMultiplexKwd        = "enable-mplex-experiment"
	enableTendermintKwd       = "tendermint"
	capacityKwd               = "storageCapacity"
	durationKwd               = "storageDuration"
	priceKwd                  = "storagePrice"
	ksKwd                     = "keeperSla"
	psKwd                     = "providerSla"
	passwordKwd               = "password"
	secretKeyKwd              = "secretKey"
	reDeploy                  = "reDeployContract"
	netKeyKwd                 = "netKey"
	posKwd                    = "pos"
	gcKwd                     = "cleanPos"
)

var (
	errWrongInput = errors.New("The input option is wrong")
	errRepoExists = errors.New("mefs configuration file already exists, reinitializing would overwrite your keys")
)

var daemonCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Run a network-connected MEFS node.",
		ShortDescription: `
'mefs daemon' runs a persistent mefs daemon that can serve commands
over the network. Most applications that use MEFS will do so by
communicating with a daemon over the HTTP API. While the daemon is
running, calls to 'mefs' commands will be sent over the network to
the daemon.
`,
		LongDescription: `
The daemon will start listening on ports on the network, which are
documented in (and can be modified through) 'mefs config Addresses'.
For example, to change the 'API' port:

  mefs config Addresses.API /ip4/127.0.0.1/tcp/5001

Make sure to restart the daemon after changing addresses.

By default, the API is only accessible locally. To expose it to
other computers in the network, use 0.0.0.0 as the ip address:

  mefs config Addresses.API /ip4/0.0.0.0/tcp/5001

Be careful if you expose the API. It is a security risk, as anyone could
control your node remotely. If you need to control the node remotely,
make sure to protect the port as you would other services or database
(firewall, authenticated proxy, etc).

HTTP Headers

mefs supports passing arbitrary headers to the API and Gateway. You can
do this by setting headers on the API.HTTPHeaders key:

  mefs config --json API.HTTPHeaders.X-Special-Header '["so special :)"]'

Note that the value of the keys is an _array_ of strings. This is because
headers can have more than one value, and it is convenient to pass through
to other libraries.

CORS Headers (for API)

You can setup CORS headers the same way:

  mefs config --json API.HTTPHeaders.Access-Control-Allow-Origin '["example.com"]'
  mefs config --json API.HTTPHeaders.Access-Control-Allow-Methods '["PUT", "GET", "POST"]'
  mefs config --json API.HTTPHeaders.Access-Control-Allow-Credentials '["true"]'

Shutdown

To shutdown the daemon, send a SIGINT signal to it (e.g. by pressing 'Ctrl-C')
or send a SIGTERM signal to it (e.g. with 'kill'). It may take a while for the
daemon to shutdown gracefully, but it can be killed forcibly by sending a
second signal.

MEFS_PATH environment variable

mefs uses a repository in the local file system. By default, the repo is
located at ~/.mefs. To change the repo location, set the $MEFS_PATH
environment variable:

  export MEFS_PATH=/path/to/mefsrepo
`,
	},

	Options: []cmds.Option{
		cmds.BoolOption(initOptionKwd, "Initialize mefs with default settings if not already initialized"),
		cmds.BoolOption(posKwd, "Pos feature for provider").WithDefault(false),
		cmds.BoolOption(unencryptTransportKwd, "Disable transport encryption (for debugging protocols)"),
		cmds.BoolOption(adjustFDLimitKwd, "Check and raise file descriptor limits if needed").WithDefault(true),
		cmds.BoolOption(enableMultiplexKwd, "Add the experimental 'go-multiplex' stream muxer to libp2p on construction.").WithDefault(true),
		cmds.StringOption(netKeyKwd, "the netKey is used to setup private network").WithDefault("dev"),
		cmds.StringOption(passwordKwd, "pwd", "the password is used to decrypt the PrivateKey").WithDefault(utils.DefaultPassword),
		cmds.StringOption(secretKeyKwd, "sk", "the stored PrivateKey").WithDefault(""),
		cmds.BoolOption(enableTendermintKwd, "If true, use Tendermint Core").WithDefault(false),
		cmds.BoolOption(reDeploy, "rdo", "used for reDeploying contract").WithDefault(false),
		cmds.Int64Option(capacityKwd, "cap", "implement user needs or provider offers how many capacity of storage").WithDefault(provider.DefaultCapacity),
		cmds.Int64Option(durationKwd, "dur", "implement user needs or provider offers how much time of storage").WithDefault(provider.DefaultDuration),
		cmds.Int64Option(priceKwd, "price", "implement user needs or provider offers how much price of storage").WithDefault(utils.STOREPRICEPEDOLLAR),
		cmds.BoolOption(gcKwd, "gc", "used for provider to clean pos data").WithDefault(false),
	},
	Subcommands: map[string]*cmds.Command{},
	Run:         daemonFunc,
}

// defaultMux tells mux to serve path using the default muxer. This is
// mostly useful to hook up things that register in the default muxer,
// and don't provide a convenient http.Handler entry point, such as
// expvar and http/pprof.
func defaultMux(path string) corehttp.ServeOption {
	return func(node *core.MefsNode, _ net.Listener, mux *http.ServeMux) (*http.ServeMux, error) {
		mux.Handle(path, http.DefaultServeMux)
		return mux, nil
	}
}

func daemonFunc(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {
	// Inject metrics before we do anything
	err := mprome.Inject()
	if err != nil {
		log.Errorf("Injecting prometheus handler for metrics failed with message: %s\n", err.Error())
	}

	// let the user know we're going.
	fmt.Printf("Initializing daemon...\n")

	// print the mefs version
	printVersion()

	managefd, _ := req.Options[adjustFDLimitKwd].(bool)
	if managefd {
		if changedFds, newFdsLimit, err := utilmain.ManageFdLimit(); err != nil {
			log.Errorf("setting file descriptor limit: %s", err)
		} else if changedFds {
			fmt.Printf("Successfully raised file descriptor limit to %d.\n", newFdsLimit)
		}
	}

	cctx := env.(*oldcmds.Context)

	// 开协程是否req结束
	go func() {
		<-req.Context.Done()
		fmt.Println("Received interrupt signal, shutting down...")
		fmt.Println("(Hit ctrl-c again to force-shutdown the daemon.)")
	}()

	// check transport encryption flag.
	unencrypted, _ := req.Options[unencryptTransportKwd].(bool)
	if unencrypted {
		log.Warningf(`Running with --%s: All connections are UNENCRYPTED.
		You will not be able to connect to regular encrypted networks.`, unencryptTransportKwd)
	}

	hexsk, _ := req.Options[secretKeyKwd].(string)
	password, _ := req.Options[passwordKwd].(string)
	nKey, _ := req.Options[netKeyKwd].(string)

	// first, whether user has provided the initialization flag. we may be
	// running in an uninitialized state.
	initialize, _ := req.Options[initOptionKwd].(bool)
	if initialize {
		// cfg 为配置根路径
		cfg := cctx.ConfigRoot
		if !fsrepo.IsInitialized(cfg) {
			err := doInit(os.Stdout, cfg, nBitsForKeypairDefault, password, nil, hexsk, nKey)
			if err != nil {
				return err
			}
		}
	}

	// acquire the repo lock _before_ constructing a node. we need to make
	// sure we are permitted to access the resources (datastore, etc.)
	repo, err := fsrepo.Open(cctx.ConfigRoot)

	if err != nil {
		return err
	}

	offline, _ := req.Options[offlineKwd].(bool)
	mplex, _ := req.Options[enableMultiplexKwd].(bool)

	// Start assembling node config
	ncfg := &core.BuildCfg{
		Repo:                        repo,
		Permanent:                   true, // It is temporary way to signify that node is permanent
		Online:                      !offline,
		DisableEncryptedConnections: unencrypted,
		ExtraOpts: map[string]bool{
			"mplex": mplex,
		},
		//TODO(Kubuxu): refactor Online vs Offline by adding Permanent vs Ephemeral
	}

	cfg, err := repo.Config()
	if err != nil {
		return err
	}

	contracts.EndPoint = cfg.Eth

	// start logger
	utils.StartLogger()

	routingOption := cfg.Routing.Type
	if routingOption == "" {
		routingOption = routingOptionDHTKwd
	}
	switch routingOption {
	case routingOptionSupernodeKwd:
		return errors.New("supernode routing was never fully implemented and has been removed")
	case routingOptionDHTClientKwd:
		ncfg.Routing = core.DHTClientOption
	case routingOptionDHTKwd:
		ncfg.Routing = core.DHTOption
	default:
		return fmt.Errorf("unrecognized routing option: %s", routingOption)
	}

	node, err := core.NewNode(req.Context, ncfg, password, nKey) //根据配置信息获得本地mefs节点实例
	if err != nil {
		log.Error("error from node construction: ", err)
		return err
	}

	node.SetLocal(false)

	if node.PNetFingerprint != nil {
		fmt.Println("Swarm is limited to private network of peers with the swarm key")
		fmt.Printf("Swarm key fingerprint: %x\n", node.PNetFingerprint)
	}

	printSwarmAddrs(node)

	nid := node.Identity.Pretty()

	if !cfg.Test {
		//从合约中获取账户角色
		isKeeper, err := role.IsKeeper(nid)
		if err != nil {
			utils.MLogger.Error("Got Keeper err: ", err)
			return err
		}
		if isKeeper {
			cfg.Role = metainfo.RoleKeeper
		} else {
			isProvider, err := role.IsProvider(nid)
			if err != nil {
				utils.MLogger.Error("Got Provider role: ", err)
				return err
			}
			if isProvider {
				cfg.Role = metainfo.RoleProvider
			} else {
				cfg.Role = metainfo.RoleUser
			}
		}
	}

	kmRole, err := metainfo.NewKeyMeta(nid, metainfo.Role)
	if err != nil {
		return err
	}

	err = node.Data.PutKey(node.Context(), kmRole.ToString(), []byte(cfg.Role), "local")
	if err != nil {
		utils.MLogger.Error("Put role key falied: ", err)
	}

	defer func() { //关闭daemon时进行的操作
		// We wait for the node to close first, as the node has children
		// that it will wait for before closing, such as the API server.

		err = node.Inst.Stop()
		if err != nil {
			utils.MLogger.Error("Persist before exist falied: ", err)
		}

		err = node.Close()
		if err != nil {
			utils.MLogger.Error("Mefs node close falied: ", err)
		}

		select {
		case <-req.Context.Done():
			utils.MLogger.Info("Gracefully shut down daemon")
		default:
		}
	}()

	cctx.ConstructNode = func() (*core.MefsNode, error) {
		return node, nil
	}

	// construct api endpoint - every time
	apiErrc, err := serveHTTPApi(req, cctx)
	if err != nil {
		return err
	}

	// just for minio
	core.LocalNode = node

	// initialize metrics collector
	prometheus.MustRegister(&corehttp.MefsNodeCollector{Node: node})

	fmt.Printf("Daemon is ready\n")

	err = mcl.Init(mcl.BLS12_381)
	if err != nil {
		utils.MLogger.Error("Init BLS12_381 curve failed: ", err)
		<-req.Context.Done()
	} else {
		utils.MLogger.Info("Init BLS12_381 curve success")
	}

	capacity, ok := req.Options[capacityKwd].(int64)
	if !ok || capacity <= 0 {
		fmt.Println("input wrong capacity.")
		return errRepoExists
	}
	duration, ok := req.Options[durationKwd].(int64)
	if !ok || duration <= 0 {
		fmt.Println("input wrong duration.")
		return errRepoExists
	}
	price, ok := req.Options[priceKwd].(int64)
	if !ok || price <= 0 {
		fmt.Println("input wrong price.")
		return errRepoExists
	}

	rdo, ok := req.Options[reDeploy].(bool)
	if !ok {
		fmt.Println("input wrong value for redeploy.")
		return errRepoExists
	}

	switch cfg.Role {
	case metainfo.RoleKeeper:
		ins, err := keeper.New(node.Context(), node.Identity.Pretty(), node.PrivateKey, node.Data, node.Routing)
		if err != nil {
			fmt.Println("Start keeper service fails; please restart")
		}
		node.Inst = ins

		fmt.Println("Keeper daemon is ready")

	case metainfo.RoleUser:
		ins, err := user.New(node.Identity.Pretty(), node.Data, node.Routing)
		if err != nil {
			fmt.Println("Start user daemon fails; please restart")
		}

		node.Inst = ins

		fmt.Println("User daemon is ready; run `mefs lfs start` to start lfs service")
	case metainfo.RoleProvider: //provider和keeper同样
		fmt.Println("started as a provider")
		ins, err := provider.New(req.Context, node.Identity.Pretty(), node.PrivateKey, node.Data, node.Routing, capacity, duration, price, rdo)
		if err != nil {
			fmt.Println("Start providerService failed:", err)
			return err
		}

		node.Inst = ins

		fmt.Println("Provider daemon is ready")

		pos, _ := req.Options[posKwd].(bool)
		gc, _ := req.Options[gcKwd].(bool)

		if pos {
			utils.MLogger.Info("Start pos Service")
			go ins.(*provider.Info).PosService(req.Context, gc)
		}
	default:
	}

	// collect long-running errors and block for shutdown
	// TODO(cryptix): our fuse currently doesnt follow this pattern for graceful shutdown
	for err := range merge(apiErrc) {
		if err != nil {
			return err
		}
	}

	return nil
}

// serveHTTPApi collects options, creates listener, prints status message and starts serving requests
func serveHTTPApi(req *cmds.Request, cctx *oldcmds.Context) (<-chan error, error) {
	cfg, err := cctx.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("serveHTTPApi: GetConfig() failed: %s", err)
	}

	apiAddrs := make([]string, 0, 2)
	apiAddr, _ := req.Options[commands.ApiOption].(string)
	// apiAddr在metb与mefs都为空，apiAddrs是根据config的API确定的
	if apiAddr == "" {
		apiAddrs = cfg.Addresses.API
	} else {
		apiAddrs = append(apiAddrs, apiAddr)
	}

	listeners := make([]manet.Listener, 0, len(apiAddrs))
	for _, addr := range apiAddrs {
		apiMaddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, fmt.Errorf("serveHTTPApi: invalid API address: %q (err: %s)", apiAddr, err)
		}
		// 此时的端口号为0，传入的端口号为0或空则自动选择一个可用端口号
		apiLis, err := manet.Listen(apiMaddr)
		if err != nil {
			return nil, fmt.Errorf("serveHTTPApi: manet.Listen(%s) failed: %s", apiMaddr, err)
		}

		// we might have listened to /tcp/0 - lets see what we are listing on
		apiMaddr = apiLis.Multiaddr()
		fmt.Printf("API server listening on %s\n", apiMaddr)
		listeners = append(listeners, apiLis)
	}

	// by default, we don't let you load arbitrary mefs objects through the api,
	// because this would open up the api to scripting vulnerabilities.
	// only the webui objects are allowed.
	// if you know what you're doing, go ahead and pass --unrestricted-api.

	var opts = []corehttp.ServeOption{
		corehttp.MetricsCollectionOption("api"),
		corehttp.CheckVersionOption(),
		corehttp.CommandsOption(*cctx),
		defaultMux("/debug/vars"),
		defaultMux("/debug/pprof/"),
		corehttp.MutexFractionOption("/debug/pprof-mutex/"),
		corehttp.MetricsScrapingOption("/debug/metrics/prometheus"),
		corehttp.LogOption(),
		corehttp.MLog(),
	}

	if len(cfg.Gateway.RootRedirect) > 0 {
		opts = append(opts, corehttp.RedirectOption("", cfg.Gateway.RootRedirect))
	}

	node, err := cctx.ConstructNode()
	if err != nil {
		return nil, fmt.Errorf("serveHTTPApi: ConstructNode() failed: %s", err)
	}

	if err := node.Repo.SetAPIAddr(listeners[0].Multiaddr()); err != nil {
		return nil, fmt.Errorf("serveHTTPApi: SetAPIAddr() failed: %s", err)
	}

	errc := make(chan error)
	var wg sync.WaitGroup
	for _, apiLis := range listeners {
		wg.Add(1)
		go func(lis manet.Listener) {
			defer wg.Done()
			errc <- corehttp.Serve(node, manet.NetListener(lis), opts...)
		}(apiLis)
	}

	go func() {
		wg.Wait()
		close(errc)
	}()

	return errc, nil
}

// printSwarmAddrs prints the addresses of the host
func printSwarmAddrs(node *core.MefsNode) {
	if !node.OnlineMode() {
		fmt.Println("Swarm not listening, running in offline mode.")
		return
	}

	var lisAddrs []string
	ifaceAddrs, err := node.PeerHost.Network().InterfaceListenAddresses()
	if err != nil {
		log.Errorf("failed to read listening addresses: %s", err)
	}
	for _, addr := range ifaceAddrs {
		lisAddrs = append(lisAddrs, addr.String())
	}
	sort.Strings(lisAddrs)
	for _, addr := range lisAddrs {
		fmt.Printf("Swarm listening on %s\n", addr)
	}

	var addrs []string
	for _, addr := range node.PeerHost.Addrs() {
		addrs = append(addrs, addr.String())
	}
	sort.Strings(addrs)
	for _, addr := range addrs {
		fmt.Printf("Swarm announcing %s\n", addr)
	}

}

// serveHTTPGateway collects options, creates listener, prints status message and starts serving requests

// merge does fan-in of multiple read-only error channels
// taken from http://blog.golang.org/pipelines
func merge(cs ...<-chan error) <-chan error {
	var wg sync.WaitGroup
	out := make(chan error)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan error) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	for _, c := range cs {
		if c != nil {
			wg.Add(1)
			go output(c)
		}
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func YesNoPrompt(prompt string) bool {
	var s string
	for i := 0; i < 3; i++ {
		fmt.Printf("%s ", prompt)
		_, err := fmt.Scanf("%s", &s)
		if err != nil {
			fmt.Println("fmt.Scanf falied: ", err)
		}
		switch s {
		case "y", "Y":
			return true
		case "n", "N":
			return false
		case "":
			return false
		}
		fmt.Println("Please press either 'y' or 'n'")
	}

	return false
}

func printVersion() {
	fmt.Printf("go-mefs version: %s-%s\n", version.CurrentVersionNumber, version.CurrentCommit)
	fmt.Printf("Repo version: %d\n", fsrepo.RepoVersion)
	fmt.Printf("System version: %s\n", runtime.GOARCH+"/"+runtime.GOOS)
	fmt.Printf("Golang version: %s\n", runtime.Version())
}
