package pluginLocalMefs

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	ipfs "github.com/memoio/go-mefs/source/metb-plugins"
	testbedi "github.com/memoio/go-mefs/source/metb-plugins/cli/testbed/interfaces"
	metbutil "github.com/memoio/go-mefs/source/metb-plugins/cli/util"

	config "github.com/memoio/go-mefs/config"
	serial "github.com/memoio/go-mefs/config/serialize"
	"github.com/memoio/go-mefs/source/go-cid"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

var errTimeout = errors.New("timeout")

var PluginName = "localmefs"

type LocalMefs struct {
	dir       string
	peerid    *cid.Cid
	apiaddr   multiaddr.Multiaddr
	swarmaddr multiaddr.Multiaddr
	binary    string
	mdns      bool
}

// NewNode creates a LocalMefs metb core node that runs ipfs on the local
// system by using random ports for both the api and the swarm.
// Attributes
// - binary: binary to use for Init, Start (defaults to ipfs in path)
// - apiaddr: multiaddr use for the api (defaults to /ip4/127.0.0.1/tcp/0)
// - swarmaddr: multiaddr used for swarm (defaults to /ip4/127.0.0.1/tcp/0)
// - mdns: if present, enables mdns (off by default)
func NewNode(dir string, attrs map[string]string) (testbedi.Core, error) {
	mdns := false
	binary := ""

	var ok bool

	if binary, ok = attrs["binary"]; !ok {
		var err error
		binary, err = exec.LookPath("mefs")
		if err != nil {
			return nil, err
		}
	}

	apiaddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
	if err != nil {
		return nil, err
	}

	swarmaddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
	if err != nil {
		return nil, err
	}

	if apiaddrstr, ok := attrs["apiaddr"]; ok {
		var err error
		apiaddr, err = multiaddr.NewMultiaddr(apiaddrstr)

		if err != nil {
			return nil, err
		}
	}

	if swarmaddrstr, ok := attrs["swarmaddr"]; ok {
		var err error
		swarmaddr, err = multiaddr.NewMultiaddr(swarmaddrstr)

		if err != nil {
			return nil, err
		}
	}

	if _, ok := attrs["mdns"]; ok {
		mdns = true
	}
	return &LocalMefs{
		dir:       dir,
		apiaddr:   apiaddr,
		swarmaddr: swarmaddr,
		binary:    binary,
		mdns:      mdns,
	}, nil

}

func GetAttrList() []string {
	return ipfs.GetAttrList()
}

func GetAttrDesc(attr string) (string, error) {
	return ipfs.GetAttrDesc(attr)
}

func GetMetricList() []string {
	return ipfs.GetMetricList()
}

func GetMetricDesc(attr string) (string, error) {
	return ipfs.GetMetricDesc(attr)
}

/// TestbedNode Interface

func (l *LocalMefs) Init(ctx context.Context, agrs ...string) (testbedi.Output, error) {
	agrs = append([]string{l.binary, "init"}, agrs...)
	output, oerr := l.RunCmd(ctx, nil, agrs...)
	if oerr != nil {
		return nil, oerr
	}

	icfg, err := l.Config()
	if err != nil {
		return nil, err
	}

	lcfg := icfg.(*config.Config)

	lcfg.Bootstrap = []string{}
	lcfg.Addresses.Swarm = []string{l.swarmaddr.String()}
	lcfg.Addresses.API = []string{l.apiaddr.String()}
	lcfg.Addresses.Gateway = []string{""}
	lcfg.Discovery.MDNS.Enabled = l.mdns

	err = l.WriteConfig(lcfg)
	if err != nil {
		return nil, err
	}

	return output, oerr
}

func (l *LocalMefs) Start(ctx context.Context, wait bool, args ...string) (testbedi.Output, error) {
	alive, err := l.isAlive()
	if err != nil {
		return nil, err
	}

	if alive {
		return nil, fmt.Errorf("node is already running")
	}

	dir := l.dir
	dargs := append([]string{"daemon"}, args...)
	cmd := exec.Command(l.binary, dargs...)
	cmd.Dir = dir

	cmd.Env, err = l.env()
	if err != nil {
		return nil, err
	}

	metbutil.SetupOpt(cmd)

	stdout, err := os.Create(filepath.Join(dir, "daemon.stdout"))
	if err != nil {
		return nil, err
	}

	stderr, err := os.Create(filepath.Join(dir, "daemon.stderr"))
	if err != nil {
		return nil, err
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	pid := cmd.Process.Pid

	err = ioutil.WriteFile(filepath.Join(dir, "daemon.pid"), []byte(fmt.Sprint(pid)), 0666)
	if err != nil {
		return nil, err
	}

	if wait {
		return nil, ipfs.WaitOnAPI(l)
	}

	return nil, nil
}

func (l *LocalMefs) Stop(ctx context.Context) error {
	pid, err := l.getPID()
	if err != nil {
		return fmt.Errorf("error killing daemon %s: %s", l.dir, err)
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("error killing daemon %s: %s", l.dir, err)
	}

	waitch := make(chan struct{}, 1)
	go func() {
		p.Wait()
		waitch <- struct{}{}
	}()

	defer func() {
		err := os.Remove(filepath.Join(l.dir, "daemon.pid"))
		if err != nil && !os.IsNotExist(err) {
			panic(fmt.Errorf("error removing pid file for daemon at %s: %s", l.dir, err))
		}
	}()

	if err := l.signalAndWait(p, waitch, syscall.SIGTERM, 1*time.Second); err != errTimeout {
		return err
	}

	if err := l.signalAndWait(p, waitch, syscall.SIGTERM, 2*time.Second); err != errTimeout {
		return err
	}

	if err := l.signalAndWait(p, waitch, syscall.SIGQUIT, 5*time.Second); err != errTimeout {
		return err
	}

	if err := l.signalAndWait(p, waitch, syscall.SIGKILL, 5*time.Second); err != errTimeout {
		return err
	}

	return fmt.Errorf("Could not stop LocalMefs node with pid %d", pid)
}

func (l *LocalMefs) RunCmd(ctx context.Context, stdin io.Reader, args ...string) (testbedi.Output, error) {
	env, err := l.env()

	if err != nil {
		return nil, fmt.Errorf("error getting env: %s", err)
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Env = env
	cmd.Stdin = stdin

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	stderrbytes, err := ioutil.ReadAll(stderr)
	if err != nil {
		return nil, err
	}

	stdoutbytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	exiterr := cmd.Wait()

	var exitcode = 0
	switch oerr := exiterr.(type) {
	case *exec.ExitError:
		if ctx.Err() == context.DeadlineExceeded {
			err = errors.Wrapf(oerr, "context deadline exceeded for command: %q", strings.Join(cmd.Args, " "))
		}

		exitcode = 1
	case nil:
		err = oerr
	}

	return metbutil.NewOutput(args, stdoutbytes, stderrbytes, exitcode, err), nil
}

func (l *LocalMefs) Connect(ctx context.Context, tbn testbedi.Core) error {
	swarmaddrs, err := tbn.SwarmAddrs()
	if err != nil {
		return err
	}

	fmt.Println(swarmaddrs)
	output, err := l.RunCmd(ctx, nil, "mefs", "swarm", "connect", swarmaddrs[0])

	if err != nil {
		return err
	}

	if output.ExitCode() != 0 {
		out, err := ioutil.ReadAll(output.Stderr())
		if err != nil {
			return err
		}

		return fmt.Errorf("%s", string(out))
	}

	return nil
}

func (l *LocalMefs) Shell(ctx context.Context, nodes []testbedi.Core) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return fmt.Errorf("no shell found")
	}

	if len(os.Getenv("IPFS_PATH")) != 0 {
		// If the users shell sets IPFS_PATH, it will just be overridden by the shell again
		return fmt.Errorf("shell has IPFS_PATH set, please unset before trying to use metb shell")
	}

	nenvs, err := l.env()
	if err != nil {
		return err
	}

	// TODO(tperson): It would be great if we could guarantee that the shell
	// is using the same binary. However, the users shell may prepend anything
	// we change in the PATH

	for i, n := range nodes {
		peerid, err := n.PeerID()

		if err != nil {
			return err
		}

		nenvs = append(nenvs, fmt.Sprintf("NODE%d=%s", i, peerid))
	}

	return syscall.Exec(shell, []string{shell}, nenvs)
}

func (l *LocalMefs) String() string {
	pcid, err := l.PeerID()
	if err != nil {
		return fmt.Sprintf("%s", l.Type())
	}
	return fmt.Sprintf("%s", pcid[0:12])
}

func (l *LocalMefs) APIAddr() (string, error) {
	return ipfs.GetAPIAddrFromRepo(l.dir)
}

func (l *LocalMefs) SwarmAddrs() ([]string, error) {
	return ipfs.SwarmAddrs(l)
}

func (l *LocalMefs) Dir() string {
	return l.dir
}

func (l *LocalMefs) PeerID() (string, error) {
	if l.peerid != nil {
		return l.peerid.String(), nil
	}

	var err error
	l.peerid, err = ipfs.GetPeerID(l)

	if err != nil {
		return "", err
	}

	return l.peerid.String(), nil
}

/// Metric Interface

func (l *LocalMefs) GetMetricList() []string {
	return GetMetricList()
}

func (l *LocalMefs) GetMetricDesc(attr string) (string, error) {
	return GetMetricDesc(attr)
}

func (l *LocalMefs) Metric(metric string) (string, error) {
	return ipfs.GetMetric(l, metric)
}

func (l *LocalMefs) Heartbeat() (map[string]string, error) {
	return nil, nil
}

func (l *LocalMefs) Events() (io.ReadCloser, error) {
	return ipfs.ReadLogs(l)
}

func (l *LocalMefs) Logs() (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

// Attribute Interface

func (l *LocalMefs) GetAttrList() []string {
	return GetAttrList()
}

func (l *LocalMefs) GetAttrDesc(attr string) (string, error) {
	return GetAttrDesc(attr)
}

func (l *LocalMefs) Attr(attr string) (string, error) {
	return ipfs.GetAttr(l, attr)
}

func (l *LocalMefs) SetAttr(string, string) error {
	return fmt.Errorf("no attribute to set")
}

func (l *LocalMefs) StderrReader() (io.ReadCloser, error) {
	return l.readerFor("daemon.stderr")
}

func (l *LocalMefs) StdoutReader() (io.ReadCloser, error) {
	return l.readerFor("daemon.stdout")
}

func (l *LocalMefs) Config() (interface{}, error) {
	return serial.Load(filepath.Join(l.dir, "config"))
}

func (l *LocalMefs) WriteConfig(cfg interface{}) error {
	return serial.WriteConfigFile(filepath.Join(l.dir, "config"), cfg)
}

func (l *LocalMefs) Type() string {
	return "mefs"
}

func (l *LocalMefs) Deployment() string {
	return "local"
}

func (l *LocalMefs) readerFor(file string) (io.ReadCloser, error) {
	return os.OpenFile(filepath.Join(l.dir, file), os.O_RDONLY, 0)
}

func (l *LocalMefs) signalAndWait(p *os.Process, waitch <-chan struct{}, signal os.Signal, t time.Duration) error {
	err := p.Signal(signal)
	if err != nil {
		return fmt.Errorf("error killing daemon %s: %s", l.dir, err)
	}

	select {
	case <-waitch:
		return nil
	case <-time.After(t):
		return errTimeout
	}
}

func (l *LocalMefs) getPID() (int, error) {
	b, err := ioutil.ReadFile(filepath.Join(l.dir, "daemon.pid"))
	if err != nil {
		return -1, err
	}

	return strconv.Atoi(string(b))
}

func (l *LocalMefs) isAlive() (bool, error) {
	pid, err := l.getPID()
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return false, nil
	}

	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return true, nil
	}

	return false, nil
}

func (l *LocalMefs) env() ([]string, error) {
	envs := os.Environ()
	ipfspath := "MEFS_PATH=" + l.dir

	for i, e := range envs {
		if strings.HasPrefix(e, "MEFS_PATH=") {
			envs[i] = ipfspath
			return envs, nil
		}
	}
	return append(envs, ipfspath), nil
}
