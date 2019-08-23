package node

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/memoio/go-mefs/consensus/memoriae"
	abci "github.com/tendermint/tendermint/abci/types"
	tcfg "github.com/tendermint/tendermint/config"
	tlog "github.com/tendermint/tendermint/libs/log"
	tnode "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
)

var (
	// config = tcfg.DefaultConfig()
	logger = tlog.NewTMLogger(tlog.NewSyncWriter(os.Stdout))
)

type Node struct {
	//文件存放目录
	NodePath string
	//组ID
	GroupID string

	//此节点的Tendermint cofig
	Conf *tcfg.Config

	Node *tnode.Node

	App abci.Application

	//组创建时间
	Time time.Time

	IsRunning bool

	closers []interface {
		Close()
	}
}

func NewNode(nodePath string, groupID string, time time.Time) *Node {
	conf := tcfg.DefaultConfig()
	conf.RootDir = nodePath
	conf.SetRoot(nodePath) //此处不必建立文件
	return &Node{
		NodePath:  nodePath,
		GroupID:   groupID,
		Time:      time,
		IsRunning: false,
		Conf:      conf,
	}
}

// ConstructNode returns a Tendermint node
func (n *Node) constructNode(config *tcfg.Config, logger tlog.Logger) (*tnode.Node, error) {
	n.Conf.SetRoot(n.Conf.RootDir)
	tcfg.EnsureRoot(n.Conf.RootDir)
	// Generate node PrivKey
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return nil, err
	}

	// Convert old PrivValidator if it exists.
	oldPrivVal := config.OldPrivValidatorFile()
	newPrivValKey := config.PrivValidatorKeyFile()
	newPrivValState := config.PrivValidatorStateFile()
	if _, err := os.Stat(oldPrivVal); !os.IsNotExist(err) {
		oldPV, err := privval.LoadOldFilePV(oldPrivVal)
		if err != nil {
			return nil, fmt.Errorf("Error reading OldPrivValidator from %v: %v\n", oldPrivVal, err)
		}
		logger.Info("Upgrading PrivValidator file",
			"old", oldPrivVal,
			"newKey", newPrivValKey,
			"newState", newPrivValState,
		)
		oldPV.Upgrade(newPrivValKey, newPrivValState)
	}

	app := memoriae.NewMemoriaeApplication(n.Conf.DBDir())
	n.closers = append(n.closers, app)
	n.App = app
	//App := memoriae.NewMockMemoriaeApplication()

	return tnode.NewNode(config,
		privval.LoadOrGenFilePV(newPrivValKey, newPrivValState),
		nodeKey,
		proxy.NewLocalClientCreator(n.App),
		tnode.DefaultGenesisDocProviderFunc(config),
		tnode.DefaultDBProvider,
		tnode.DefaultMetricsProvider(config.Instrumentation),
		logger,
	)
}

func (n *Node) Run(p2pPort int, rpcPort int, peers []string) error {
	n.IsRunning = true
	n.constuctCfg(p2pPort, rpcPort, peers)
	// private validator
	n.Conf.SetRoot(n.Conf.RootDir)
	tcfg.EnsureRoot(n.Conf.RootDir)
	logPath := path.Join(n.NodePath, n.GroupID+"-"+strconv.Itoa(p2pPort)+"-"+strconv.Itoa(rpcPort)+".log")
	f, err := os.OpenFile(logPath, os.O_RDWR, 666)
	if os.IsNotExist(err) {
		f, err = os.Create(logPath)
	}
	if err != nil {
		return fmt.Errorf("Failed to start node: %v", err)
	}

	//设置日志的过滤选项，避免日志过多
	logFilterOption, err := tlog.AllowLevel(n.Conf.BaseConfig.LogLevel)
	if err != nil {
		logFilterOption = tlog.AllowInfo()
	}
	logger := tlog.NewFilter(tlog.NewTMLogger(tlog.NewSyncWriter(f)), logFilterOption)

	n.Node, err = n.constructNode(n.Conf, logger)
	if err != nil {
		return fmt.Errorf("Failed to start node: %v", err)
	}
	// Stop upon receiving SIGTERM or CTRL-C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			logger.Error(fmt.Sprintf("captured %v, exiting...", sig))
			if n.Node.IsRunning() {
				n.Node.Stop()
				n.IsRunning = false
			}
			return
		}
	}()

	if err := n.Node.Start(); err != nil {
		return fmt.Errorf("Failed to start node: %v", err)
	}
	logger.Info("Started Tendermint node", "nodeInfo", n.Node.Switch().NodeInfo())

	// Run forever
	//select {}

	return nil
}

func (n *Node) Stop() error {
	err := n.Node.Stop()
	if err != nil {
		return err
	}

	n.IsRunning = false
	return nil
}

//修改本地tendermint节点中的对等节点信息
func (n *Node) ChangePeers(peers []string) {
	if n.Conf == nil {
		n.Conf = tcfg.DefaultConfig()
	}
	n.Conf.P2P.PersistentPeers = strings.Join(peers, ",")
}
