package node

import (
	"fmt"
	"path"

	_ "net/http/pprof"

	mjson "github.com/memoio/go-mefs/consensus/util/json"
	tcfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"

	cmn "github.com/tendermint/tendermint/libs/common"

	tmtime "github.com/tendermint/tendermint/types/time"
)

//初始化目录
func (n *Node) InitHome() error {
	if n.Conf == nil {
		n.Conf = tcfg.DefaultConfig()
	}
	//创建相关目录
	n.Conf.RootDir = n.NodePath
	//初始化目录
	err := initFilesWithConfig(n.Conf)
	if err != nil {
		return err
	}
	//修改文件
	genesis := path.Join(n.NodePath, "config", "genesis.json") //修改创世状态文件
	mjson.HandleJson(genesis, genesis, func(data map[string]interface{}) {
		data["chain_id"] = "Group-" + n.GroupID
		data["genesis_time"] = n.Time.Round(0).UTC()
	})
	return nil
}

func initFilesWithConfig(config *tcfg.Config) error {
	// private validator
	config.SetRoot(config.RootDir)
	tcfg.EnsureRoot(config.RootDir)
	privValKeyFile := config.PrivValidatorKeyFile()
	privValStateFile := config.PrivValidatorStateFile()
	var pv *privval.FilePV
	if cmn.FileExists(privValKeyFile) {
		pv = privval.LoadFilePV(privValKeyFile, privValStateFile)
		logger.Info("Found private validator", "keyFile", privValKeyFile,
			"stateFile", privValStateFile)
	} else {
		pv = privval.GenFilePV(privValKeyFile, privValStateFile)
		pv.Save()
		logger.Info("Generated private validator", "keyFile", privValKeyFile,
			"stateFile", privValStateFile)
	}

	nodeKeyFile := config.NodeKeyFile()
	if cmn.FileExists(nodeKeyFile) {
		logger.Info("Found node key", "path", nodeKeyFile)
	} else {
		if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
			return err
		}
		logger.Info("Generated node key", "path", nodeKeyFile)
	}

	// genesis file
	genFile := config.GenesisFile()
	if cmn.FileExists(genFile) {
		logger.Info("Found genesis file", "path", genFile)
	} else {
		genDoc := types.GenesisDoc{
			ChainID:         fmt.Sprintf("test-chain-%v", cmn.RandStr(6)),
			GenesisTime:     tmtime.Now(),
			ConsensusParams: types.DefaultConsensusParams(),
		}
		key := pv.GetPubKey()
		genDoc.Validators = []types.GenesisValidator{{
			Address: key.Address(),
			PubKey:  key,
			Power:   10,
		}}

		if err := genDoc.SaveAs(genFile); err != nil {
			return err
		}
		logger.Info("Generated genesis file", "path", genFile)
	}

	return nil
}
