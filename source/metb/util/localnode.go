package metbutil

import (
	"bytes"
	"errors"
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

	config "github.com/memoio/go-mefs/config"
	serial "github.com/memoio/go-mefs/config/serialize"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
)

var ErrTimeout = errors.New("timeout")

// LocalNode is a machine-local MEFS node controlled by METB
type LocalNode struct {
	Dir    string
	PeerID string
}

// assert LocalNode satisfies the MefsNode interface
var _ MefsNode = (*LocalNode)(nil)

func (n *LocalNode) Init() error {
	err := os.MkdirAll(n.Dir, 0777)
	if err != nil {
		return err
	}

	cmd := exec.Command("mefs", "init")
	cmd.Env, err = n.envForDaemon()
	if err != nil {
		return err
	}
	// 执行init
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(out))
	}

	return nil
}

func (n *LocalNode) GetPeerID() string {
	return n.PeerID
}

func (n *LocalNode) String() string {
	return n.PeerID
}

// Shell sets up environment variables for a new shell to more easily
// control the given daemon
func (n *LocalNode) Shell() error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return fmt.Errorf("couldnt find shell!")
	}

	nenvs := []string{"MEFS_PATH=" + n.Dir}

	nodes, err := LoadNodes()
	if err != nil {
		return err
	}

	for i, n := range nodes {
		peerid := n.GetPeerID()
		if peerid == "" {
			return fmt.Errorf("failed to check peerID")
		}

		nenvs = append(nenvs, fmt.Sprintf("NODE%d=%s", i, peerid))
	}
	nenvs = append(os.Environ(), nenvs...)

	return syscall.Exec(shell, []string{shell}, nenvs)
}

func (n *LocalNode) RunCmd(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)

	var err error
	cmd.Env, err = n.envForDaemon()
	if err != nil {
		return "", err
	}

	outbuf := new(bytes.Buffer)
	errbuf := new(bytes.Buffer)
	cmd.Stdout = outbuf
	cmd.Stderr = errbuf

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%s: %s %s", err, outbuf.String(), errbuf.String())
	}

	return outbuf.String(), nil
}

func (n *LocalNode) APIAddr() (string, error) {
	dir := n.Dir

	addrb, err := ioutil.ReadFile(filepath.Join(dir, "api"))
	if err != nil {
		return "", err
	}

	maddr, err := ma.NewMultiaddr(string(addrb))
	if err != nil {
		fmt.Println("error parsing multiaddr: ", err)
		return "", err
	}

	_, addr, err := manet.DialArgs(maddr)
	if err != nil {
		fmt.Println("error on multiaddr dialargs: ", err)
		return "", err
	}
	return addr, nil
}

func (n *LocalNode) envForDaemon() ([]string, error) {
	envs := os.Environ()
	dir := n.Dir
	npath := "MEFS_PATH=" + dir
	for i, e := range envs {
		p := strings.Split(e, "=")
		if p[0] == "MEFS_PATH" {
			envs[i] = npath
			return envs, nil
		}
	}

	return append(envs, npath), nil
}

func (n *LocalNode) Start(args []string) error {
	alive, err := n.isAlive()
	if err != nil {
		return err
	}

	if alive {
		return fmt.Errorf("node is already running")
	}
	// 在指定目录下配置命令，启动daemon
	dir := n.Dir
	dargs := append([]string{"daemon"}, args...)
	cmd := exec.Command("mefs", dargs...)
	cmd.Dir = dir

	cmd.Env, err = n.envForDaemon()
	if err != nil {
		return err
	}

	setupOpt(cmd)

	stdout, err := os.Create(filepath.Join(dir, "daemon.stdout"))
	if err != nil {
		return err
	}

	stderr, err := os.Create(filepath.Join(dir, "daemon.stderr"))
	if err != nil {
		return err
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Start()
	if err != nil {
		return err
	}
	pid := cmd.Process.Pid

	fmt.Printf("Started daemon %s, pid = %d\n", dir, pid)
	err = ioutil.WriteFile(filepath.Join(dir, "daemon.pid"), []byte(fmt.Sprint(pid)), 0666)
	if err != nil {
		return err
	}

	// Make sure node 0 is up before starting the rest so
	// bootstrapping works properly
	cfg, err := serial.Load(filepath.Join(dir, "config"))
	if err != nil {
		return err
	}

	n.PeerID = cfg.PeerID

	err = waitOnAPI(n)
	if err != nil {
		return err
	}

	return nil
}

func (n *LocalNode) getPID() (int, error) {
	b, err := ioutil.ReadFile(filepath.Join(n.Dir, "daemon.pid"))
	if err != nil {
		return -1, err
	}

	return strconv.Atoi(string(b))
}

func (n *LocalNode) isAlive() (bool, error) {
	pid, err := n.getPID()
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

func (n *LocalNode) Kill() error {
	pid, err := n.getPID()
	if err != nil {
		return fmt.Errorf("error killing daemon %s: %s", n.Dir, err)
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("error killing daemon %s: %s", n.Dir, err)
	}

	waitch := make(chan struct{}, 1)
	go func() {
		p.Wait() //TODO: pass return state
		waitch <- struct{}{}
	}()

	defer func() {
		err := os.Remove(filepath.Join(n.Dir, "daemon.pid"))
		if err != nil && !os.IsNotExist(err) {
			panic(fmt.Errorf("error removing pid file for daemon at %s: %s\n", n.Dir, err))
		}
	}()

	if err := n.signalAndWait(p, waitch, syscall.SIGTERM, 1*time.Second); err != ErrTimeout {
		return err
	}

	if err := n.signalAndWait(p, waitch, syscall.SIGTERM, 2*time.Second); err != ErrTimeout {
		return err
	}

	if err := n.signalAndWait(p, waitch, syscall.SIGQUIT, 5*time.Second); err != ErrTimeout {
		return err
	}

	if err := n.signalAndWait(p, waitch, syscall.SIGKILL, 5*time.Second); err != ErrTimeout {
		return err
	}

	for {
		err := p.Signal(syscall.Signal(0))
		if err != nil {
			break
		}
		time.Sleep(time.Millisecond * 10)
	}

	return nil
}

func (n *LocalNode) signalAndWait(p *os.Process, waitch <-chan struct{}, signal os.Signal, t time.Duration) error {
	err := p.Signal(signal)
	if err != nil {
		return fmt.Errorf("error killing daemon %s: %s\n", n.Dir, err)
	}

	select {
	case <-waitch:
		return nil
	case <-time.After(t):
		return ErrTimeout
	}
}

func (n *LocalNode) GetAttr(attr string) (string, error) {
	switch attr {
	case attrId:
		return n.GetPeerID(), nil
	case attrPath:
		return n.Dir, nil
	case attrBwIn:
		bw, err := GetBW(n)
		if err != nil {
			return "", err
		}
		return fmt.Sprint(bw.TotalIn), nil
	case attrBwOut:
		bw, err := GetBW(n)
		if err != nil {
			return "", err
		}
		return fmt.Sprint(bw.TotalOut), nil
	default:
		return "", errors.New("unrecognized attribute: " + attr)
	}
}

func (n *LocalNode) GetConfig() (*config.Config, error) {
	return serial.Load(filepath.Join(n.Dir, "config"))
}

func (n *LocalNode) WriteConfig(c *config.Config) error {
	return serial.WriteConfigFile(filepath.Join(n.Dir, "config"), c)
}

func (n *LocalNode) SetAttr(name, val string) error {
	return fmt.Errorf("no atttributes to set")
}

func (n *LocalNode) StdoutReader() (io.ReadCloser, error) {
	return n.readerFor("daemon.stdout")
}

func (n *LocalNode) StderrReader() (io.ReadCloser, error) {
	return n.readerFor("daemon.stderr")
}

func (n *LocalNode) readerFor(file string) (io.ReadCloser, error) {
	f, err := os.OpenFile(filepath.Join(n.Dir, file), os.O_RDONLY, 0)
	return f, err
}
