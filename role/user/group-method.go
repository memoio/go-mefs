package user

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/memoio/go-mefs/consensus/rpc"
	"github.com/memoio/go-mefs/consensus/util/code"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	inet "github.com/libp2p/go-libp2p-core/network"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	sc "github.com/memoio/go-mefs/utils/swarmconnect"
)

//GetKeepers 返回本节点拥有的keeper信息，并且检查连接状况 传入参数为返回信息的数量，-1为全部
func (gp *GroupService) GetKeepers(count int) ([]string, []string, error) {
	if len(gp.localPeersInfo.Keepers) == 0 {
		return nil, nil, ErrNoEnoughKeeper
	}
	var unconKeepers, conKeepers []string
	// 离线操作，count为-1则输出本地的所有keepers
	state, err := GetUserServiceState(gp.Userid)
	if err != nil {
		return nil, nil, err
	}
	if state < GroupStarted {
		if count < 0 {
			for _, keeperInfo := range gp.localPeersInfo.Keepers {
				unconKeepers = append(unconKeepers, keeperInfo.KeeperID)
			}
			return unconKeepers, nil, ErrGroupServiceNotReady
		}
		return nil, nil, ErrGroupServiceNotReady
	}
	// 线上操作
	if count < 0 {
		count = len(gp.localPeersInfo.Keepers)
	}
	for i, keeperInfo := range gp.localPeersInfo.Keepers {
		if i >= count {
			break
		}
		kid, err := peer.IDB58Decode(keeperInfo.KeeperID)
		if err != nil {
			continue
		}
		if gp.localNode.PeerHost.Network().Connectedness(kid) == inet.Connected {
			conKeepers = append(conKeepers, keeperInfo.KeeperID)
		} else {
			if !sc.ConnectTo(context.Background(), gp.localNode, keeperInfo.KeeperID) { //连接不上此keeper
				unconKeepers = append(unconKeepers, keeperInfo.KeeperID)
				continue
			}
			conKeepers = append(conKeepers, keeperInfo.KeeperID)
		}
	}
	if len(gp.localPeersInfo.Keepers) < count {
		return unconKeepers, conKeepers, ErrNoEnoughKeeper
	}
	return unconKeepers, conKeepers, nil
}

//GetLocalProviders 从本地PeersInfo中 返回本节点provider信息,并且检查连接状况
func (gp *GroupService) GetLocalProviders() ([]string, []string, error) {
	state, err := GetUserServiceState(gp.Userid)
	if err != nil {
		return nil, nil, err
	}
	if state < GroupStarted {
		return nil, nil, ErrGroupServiceNotReady
	}
	count := len(gp.localPeersInfo.Providers)
	unconPro := make([]string, 0, count)
	conPro := make([]string, 0, count)
	for _, provider := range gp.localPeersInfo.Providers {
		pid, err := peer.IDB58Decode(provider)
		if err != nil {
			continue
		}
		if gp.localNode.PeerHost.Network().Connectedness(pid) == inet.Connected {
			conPro = append(conPro, provider)
		} else {
			if !sc.ConnectTo(context.Background(), gp.localNode, provider) { //连接不上此provider
				unconPro = append(unconPro, provider)
				continue
			}
			conPro = append(conPro, provider)
		}
	}
	return unconPro, conPro, nil
}

//GetProviders 返回provider信息，
func (gp *GroupService) GetProviders(count int) ([]string, error) {
	state, err := GetUserServiceState(gp.Userid)
	if err != nil {
		return nil, err
	}
	if state < GroupStarted {
		return nil, ErrGroupServiceNotReady
	}
	var providers []string
	//小于零，则返回内存暂存所有即可
	if count < 0 {
		for _, provider := range gp.localPeersInfo.Providers {
			_, err := peer.IDB58Decode(provider)
			if err != nil {
				continue
			}
			providers = append(providers, provider)
		}
		return providers, nil
	}

	for _, provider := range gp.localPeersInfo.Providers {
		pid, err := peer.IDB58Decode(provider)
		if err != nil {
			continue
		}
		if gp.localNode.PeerHost.Network().Connectedness(pid) != inet.Connected {
			if !sc.ConnectTo(context.Background(), gp.localNode, provider) { //连接不上此provider
				log.Println("Cannot connect this provider:", provider)
				continue
			}
		}
		providers = append(providers, provider)
	}
	if len(providers) >= count {
		return providers[0:count], nil
	}

	//还不够，去找Keeper要
	km, err := metainfo.NewKeyMeta(gp.Userid, metainfo.Local, metainfo.SyncTypePid)
	if err != nil {
		return nil, err
	}
	for _, keeper := range gp.localPeersInfo.Keepers {
		pids, err := gp.localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(km.ToString(), keeper.KeeperID)
		if err == nil && pids != nil { //成功收到
			for i := 0; i < len(pids)/IDLength; i++ {
				pidstr := string(pids[i*IDLength : (i+1)*IDLength])
				if !utils.CheckDup(gp.localPeersInfo.Providers, pidstr) {
					continue
				}
				pid, _ := peer.IDB58Decode(pidstr)
				if gp.localNode.PeerHost.Network().Connectedness(pid) != inet.Connected {
					if !sc.ConnectTo(context.Background(), gp.localNode, pidstr) { //连接不上此provider
						continue
					}
				}
				//加入内存中
				gp.localPeersInfo.Providers = append(gp.localPeersInfo.Providers, pidstr)
				//添加到返回值
				providers = append(providers, pidstr)
			}
			break
		}
	}

	if len(providers) < count {
		for _, keeper := range gp.localPeersInfo.Keepers {
			if len(providers) >= count {
				return providers[0:count], nil
			}
			newProviders, err := gp.GetNewProvider(count-len(providers), providers, keeper.KeeperID)
			if err != nil {
				return providers, ErrNoEnoughProvider
			}
			for _, newProvider := range newProviders {
				pid, err := peer.IDB58Decode(newProvider)
				if err != nil {
					continue
				}
				if gp.localNode.PeerHost.Network().Connectedness(pid) != inet.Connected {
					if !sc.ConnectTo(context.Background(), gp.localNode, newProvider) { //连接不上此provider
						continue
					}
				}
				providers = append(providers, newProvider)
				if utils.CheckDup(gp.localPeersInfo.Providers, newProvider) {
					km, err := metainfo.NewKeyMeta(gp.Userid, metainfo.Local, metainfo.SyncTypePid)
					if err != nil {
						return nil, err
					}
					gp.localPeersInfo.Providers = append(gp.localPeersInfo.Providers, newProvider)
					for _, keeper := range gp.localPeersInfo.Keepers {
						err = gp.localNode.Routing.(*dht.IpfsDHT).CmdAppendTo(km.ToString(), newProvider, keeper.KeeperID)
						if err != nil {
							fmt.Println("gp.localNode.Routing.CmdAppend failed :", err)
						}
					}
				}
			}
		}
	}
	if len(providers) >= count {
		return providers[0:count], nil
	}
	return providers, ErrNoEnoughProvider
}

//TODO: 若User需要更多的Provider，可向Keeper申请
func (gp *GroupService) GetNewProvider(count int, providers []string, keeper string) ([]string, error) {
	state, err := GetUserServiceState(gp.Userid)
	if err != nil {
		return nil, err
	}
	if state < GroupStarted {
		return nil, ErrGroupServiceNotReady
	}
	var metaValue string
	for _, provider := range providers {
		metaValue += provider
	}

	km, err := metainfo.NewKeyMeta(gp.Userid, metainfo.NewKPReq, strconv.Itoa(count))
	if err != nil {
		return nil, err
	}
	pids, err := gp.sendMetaRequest(km, metaValue, keeper)

	if err != nil {
		return nil, err
	}
	if remain := len(pids) % IDLength; remain != 0 {
		pids = pids[:len(pids)-remain]
	}
	var NewProviders []string
	for i := 0; i < len(pids)/IDLength; i++ {
		pidstr := string(pids[i*IDLength : (i+1)*IDLength])
		if !utils.CheckDup(providers, pidstr) {
			continue
		}
		//添加到返回值
		NewProviders = append(NewProviders, pidstr)
	}
	return NewProviders, nil
}

//从provider获取数据块的元数据，传入数据块id号
//返回值为保存数据块的pid，offset，错误信息
func (gp *GroupService) GetBlockProviders(blockID string) (string, int, error) {
	var pidstr string
	var offset int
	state, err := GetUserServiceState(gp.Userid)
	if err != nil {
		return "", 0, err
	}
	if state < GroupStarted {
		return "", 0, ErrGroupServiceNotReady
	}
	kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.Local, metainfo.SyncTypeBlock)
	if err != nil {
		return "", 0, err
	}
	blockMeta := kmBlock.ToString()
	for i := 0; i < len(gp.localPeersInfo.Keepers); i++ {
		if sc.ConnectTo(context.Background(), gp.localNode, gp.localPeersInfo.Keepers[i].KeeperID) {
			pidAndOffset, err := gp.localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(blockMeta, gp.localPeersInfo.Keepers[i].KeeperID)
			if err == nil && pidAndOffset != nil { //成功收到
				splitedValue := strings.Split(string(pidAndOffset), metainfo.DELIMITER)
				if len(splitedValue) < 2 {
					return "", 0, ErrNoProviders
				}
				pidstr = splitedValue[0]
				offset, err = strconv.Atoi(splitedValue[1])
				if err != nil {
					log.Println("Offset decode error-", pidstr, err)
					return "", 0, err
				}
				pid, err := peer.IDB58Decode(pidstr)
				if err != nil {
					log.Println("Wrong format providerID-", pidstr, err)
					return "", 0, err
				}
				if gp.localNode.PeerHost.Network().Connectedness(pid) != inet.Connected {
					if !sc.ConnectTo(context.Background(), gp.localNode, pidstr) { //连接不上此provider
						log.Println("Cannot connect to blockprovider-", pidstr)
						return pidstr, offset, ErrCannotConnectProvider
					}

				}
				break
			} else if err != nil && i >= len(gp.localPeersInfo.Keepers)-1 {
				return "", 0, ErrNoProviders
			}
		} else if i >= len(gp.localPeersInfo.Keepers)-1 {
			return "", 0, ErrNoProviders
		}
	}
	return pidstr, offset, nil
}

func (gp *GroupService) PutDataMetaToKeepers(blockID string, provider string, offset int) error {
	state, err := GetUserServiceState(gp.Userid)
	if err != nil {
		return err
	}
	if state < GroupStarted {
		return ErrGroupServiceNotReady
	}
	kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.BlockMetaInfo, metainfo.SyncTypeBlock)
	if err != nil {
		fmt.Println("construct put blockMeta KV error :", err)
		return err
	}
	metaValue := provider + metainfo.DELIMITER + strconv.Itoa(offset)
	for _, keeper := range gp.localPeersInfo.Keepers {
		_, err = gp.sendMetaRequest(kmBlock, metaValue, keeper.KeeperID)
		if err != nil {
			fmt.Println("send metaMessage to ", keeper.KeeperID, " error :", err)
		}
	}
	return nil
}

//删除块
func (gp *GroupService) DeleteBlocksFromProvider(blockID string, updateMeta bool) error {
	state, err := GetUserServiceState(gp.Userid)
	if err != nil {
		return err
	}
	if state < GroupStarted {
		return ErrGroupServiceNotReady
	}
	provider, _, err := gp.GetBlockProviders(blockID)
	if err == ErrNoProviders { //Noprovider说明此块还不存在，不用删除
		log.Printf("Get block:%s's location error, no exist or keepers lost it.\n", blockID)
		return nil
	} else if err != nil {
		return err
	}

	km, err := metainfo.NewKeyMeta(blockID, metainfo.DeleteBlock)
	if err != nil {
		fmt.Println("construct delete block KV error :", err)
		return err
	}
	pid, err := peer.IDB58Decode(provider)
	if err != nil {
		return err
	}
	if gp.localNode.PeerHost.Network().Connectedness(pid) != inet.Connected {
		if !sc.ConnectTo(context.Background(), gp.localNode, provider) { //连接不上此provider
			log.Println("Cannot delete Block-", blockID, ErrCannotConnectProvider)
			return ErrCannotConnectProvider
		}
	}
	if updateMeta { //这个需要等待返回
		res, err := gp.sendMetaRequest(km, "", provider)
		if strings.Compare(res, metainfo.MetaHandlerComplete) != 0 || err != nil {
			log.Println("Cannot delete Block-", blockID, res, err)
			return ErrCannotDeleteMetaBlock
		}
	} else {
		go gp.sendMetaRequest(km, "", provider)
	}

	for j := 0; j < len(gp.localPeersInfo.Keepers); j++ { //从keeper上删除blockMeta
		go gp.sendMetaRequest(km, "", gp.localPeersInfo.Keepers[j].KeeperID)
	}

	return nil
}

func (gp *GroupService) GetBlockProvidersFromChain(blockID string) (string, int, error) {
	var pidstr string
	var offset int
	km, err := metainfo.NewKeyMeta(blockID, metainfo.Local, metainfo.SyncTypeBlock)
	if err != nil {
		return "", 0, err
	}
	key := km.ToString()
	c := rpc.GetHTTPClient("tcp://0.0.0.0:30201")
	res, err := c.ABCIQuery(string(code.BlockMetaPrefix), []byte(key))
	if err != nil {
		return "", 0, err
	} else if res.Response.Code != code.CodeTypeOK {
		return "", 0, errors.New(res.Response.Log)
	}
	pidAndOffset := res.Response.GetKey()
	if len(pidAndOffset) == 0 {
		return "", 0, ErrNoProviders
	}
	splitedValue := strings.Split(string(pidAndOffset), "/")
	if len(splitedValue) < 2 {
		return "", 0, ErrNoProviders
	}
	pidstr = splitedValue[0]
	offset, err = strconv.Atoi(splitedValue[1])
	if err != nil {
		return "", 0, err
	}
	pid, err := peer.IDB58Decode(pidstr)
	if err != nil {
		return "", 0, err
	}
	if gp.localNode.PeerHost.Network().Connectedness(pid) != inet.Connected {
		if !sc.ConnectTo(context.Background(), gp.localNode, pidstr) { //连接不上此provider
			return pidstr, offset, ErrCannotConnectProvider
		}

	}

	return pidstr, offset, nil
}
