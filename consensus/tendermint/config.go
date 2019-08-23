package node

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	mjson "github.com/memoio/go-mefs/consensus/util/json"
	tcfg "github.com/tendermint/tendermint/config"
	cmn "github.com/tendermint/tendermint/libs/common"
	tp2p "github.com/tendermint/tendermint/p2p"
)

type Config struct {
	NodePath string
	//组ID
	GroupID string

	//此节点的Tendermint cofig
	Conf *tcfg.Config

	//组创建时间
	Time time.Time
}

func (n *Node) constuctCfg(p2pPort int, rpcPort int, peers []string) *tcfg.Config {
	if n.Conf == nil {
		n.Conf = tcfg.DefaultConfig()
	}
	n.Conf.RootDir = n.NodePath
	n.Conf.ProxyApp = "memoriae"
	n.Conf.RPC.ListenAddress = "tcp://0.0.0.0:" + strconv.Itoa(rpcPort)
	n.Conf.P2P.ListenAddress = "tcp://0.0.0.0:" + strconv.Itoa(p2pPort)
	n.Conf.P2P.PersistentPeers = strings.Join(peers, ",")
	n.Conf.P2P.AllowDuplicateIP = true
	n.Conf.P2P.AddrBookStrict = false
	n.Conf.Consensus.CreateEmptyBlocks = false
	n.Conf.Moniker = n.GroupID + "-" + strconv.Itoa(p2pPort) + "-" + strconv.Itoa(rpcPort)
	return n.Conf
}

func (n *Node) GetID() (string, error) {
	nodeKey, err := tp2p.LoadNodeKey(n.Conf.NodeKeyFile())
	if err != nil {
		return "", err
	}
	return string(nodeKey.ID()), nil
}

// 获取节点公钥
func (n *Node) GetPublicKey() (string, error) {
	keyPath := path.Join(n.NodePath, "config", "priv_validator_key.json")
	byteValue, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return "", err
	}

	var data map[string]interface{}
	err = json.Unmarshal(byteValue, &data)
	if err != nil {
		return "", err
	}

	data["power"] = "10"
	data["name"] = ""
	delete(data, "priv_key")
	byteValue, err = json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(byteValue), nil
}

// 修改验证信息
func (n *Node) AlterValidator(validators []string) {
	//修改之前先检查是否修改过。
	flagpath := path.Join(n.NodePath, "config", "Validatorlock.flag")
	if cmn.FileExists(flagpath) {
		fmt.Println("创世块已被修改过，这次不修改")
		return
	}

	genesis := path.Join(n.NodePath, "config", "genesis.json")
	vd := "["
	for _, v := range validators {
		vd += v + ","
	}
	vd = vd[0 : len(vd)-1]
	vd += "]"
	mjson.HandleJson(genesis, genesis, func(data map[string]interface{}) {
		var result []interface{}
		err := json.Unmarshal([]byte(vd), &result)
		if err != nil {
			log.Fatal(err)
		}
		data["validators"] = result
	})

	file, _ := os.Create(flagpath)
	defer file.Close() //修改完成后，创建flag文件，之后的重启操作不执行创世块修改
}
