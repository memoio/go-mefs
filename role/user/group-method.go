package user

import (
	"context"
	"log"
	"strconv"
	"strings"

	inet "github.com/libp2p/go-libp2p-core/network"
	peer "github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils/metainfo"
	sc "github.com/memoio/go-mefs/utils/swarmconnect"
)

//GetKeepers 返回本节点拥有的keeper信息，并且检查连接状况 传入参数为返回信息的数量，-1为全部
func (gp *GroupService) GetKeepers(count int) ([]string, []string, error) {
	if gp == nil {
		return nil, nil, ErrGroupServiceNotReady
	}
	if len(gp.localPeersInfo.Keepers) == 0 {
		return nil, nil, ErrNoEnoughKeeper
	}
	var unconKeepers, conKeepers []string
	state, err := GetUserServiceState(gp.Userid)
	if err != nil {
		return nil, nil, err
	}
	if state < GroupStarted {
		return nil, nil, ErrGroupServiceNotReady
	}

	// count为-1则输出本地的所有keepers
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
		if localNode.PeerHost.Network().Connectedness(kid) == inet.Connected {
			conKeepers = append(conKeepers, keeperInfo.KeeperID)
		} else {
			if !sc.ConnectTo(context.Background(), localNode, keeperInfo.KeeperID) { //连接不上此keeper
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
	if gp == nil {
		return nil, nil, ErrGroupServiceNotReady
	}
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
		if localNode.PeerHost.Network().Connectedness(pid) == inet.Connected {
			conPro = append(conPro, provider)
		} else {
			if !sc.ConnectTo(context.Background(), localNode, provider) { //连接不上此provider
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
	if gp == nil {
		return nil, ErrGroupServiceNotReady
	}
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
		if localNode.PeerHost.Network().Connectedness(pid) != inet.Connected {
			if !sc.ConnectTo(context.Background(), localNode, provider) { //连接不上此provider
				log.Println("Cannot connect this provider:", provider)
				continue
			}
		}
		providers = append(providers, provider)
	}
	if count < 0 {
		return providers, nil
	}

	if len(providers) >= count {
		return providers[0:count], nil
	}

	return providers, ErrNoEnoughProvider
}

// GetBlockProviders 从provider获取数据块的元数据，传入数据块id号
//返回值为保存数据块的pid，offset，错误信息
func (gp *GroupService) GetBlockProviders(blockID string) (string, int, error) {
	if gp == nil {
		return "", 0, ErrGroupServiceNotReady
	}
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
		if sc.ConnectTo(context.Background(), localNode, gp.localPeersInfo.Keepers[i].KeeperID) {
			pidAndOffset, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(blockMeta, gp.localPeersInfo.Keepers[i].KeeperID)
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
				if localNode.PeerHost.Network().Connectedness(pid) != inet.Connected {
					if !sc.ConnectTo(context.Background(), localNode, pidstr) { //连接不上此provider
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

func (gp *GroupService) putDataMetaToKeepers(blockID string, provider string, offset int) error {
	if gp == nil {
		return ErrGroupServiceNotReady
	}
	state, err := GetUserServiceState(gp.Userid)
	if err != nil {
		return err
	}
	if state < GroupStarted {
		return ErrGroupServiceNotReady
	}
	kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.BlockMetaInfo, metainfo.SyncTypeBlock)
	if err != nil {
		log.Println("construct put blockMeta KV error :", err)
		return err
	}
	metaValue := provider + metainfo.DELIMITER + strconv.Itoa(offset)
	for _, keeper := range gp.localPeersInfo.Keepers {
		_, err = sendMetaRequest(kmBlock, metaValue, keeper.KeeperID)
		if err != nil {
			log.Println("send metaMessage to ", keeper.KeeperID, " error :", err)
		}
	}
	return nil
}

//删除块
func (gp *GroupService) deleteBlocksFromProvider(blockID string, updateMeta bool) error {
	if gp == nil {
		return ErrGroupServiceNotReady
	}
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
		log.Println("construct delete block KV error :", err)
		return err
	}
	pid, err := peer.IDB58Decode(provider)
	if err != nil {
		return err
	}
	if localNode.PeerHost.Network().Connectedness(pid) != inet.Connected {
		if !sc.ConnectTo(context.Background(), localNode, provider) { //连接不上此provider
			log.Println("Cannot delete Block-", blockID, ErrCannotConnectProvider)
			return ErrCannotConnectProvider
		}
	}
	if updateMeta { //这个需要等待返回
		res, err := sendMetaRequest(km, "", provider)
		if strings.Compare(res, metainfo.MetaHandlerComplete) != 0 || err != nil {
			log.Println("Cannot delete Block-", blockID, res, err)
			return ErrCannotDeleteMetaBlock
		}
	} else {
		go sendMetaRequest(km, "", provider)
	}

	for j := 0; j < len(gp.localPeersInfo.Keepers); j++ { //从keeper上删除blockMeta
		go sendMetaRequest(km, "", gp.localPeersInfo.Keepers[j].KeeperID)
	}

	return nil
}
