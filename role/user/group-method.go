package user

import (
	"context"
	"log"
	"strconv"
	"strings"

	inet "github.com/libp2p/go-libp2p-core/network"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/memoio/go-mefs/utils/metainfo"
	sc "github.com/memoio/go-mefs/utils/swarmconnect"
)

//GetKeepers 返回本节点拥有的keeper信息，并且检查连接状况 传入参数为返回信息的数量，-1为全部
func GetKeepers(userID string, count int) ([]string, []string, error) {
	gp := getGroup(userID)
	return gp.getKeepers(count)
}

func (gp *groupInfo) getKeepers(count int) ([]string, []string, error) {
	var unconKeepers, conKeepers []string
	state, err := getState(gp.userid)
	if err != nil {
		return nil, nil, err
	}
	if state < groupStarted {
		return nil, nil, ErrGroupServiceNotReady
	}

	// count为-1则输出本地的所有keepers
	if count < 0 {
		count = len(gp.keepers)
	}
	for i, kp := range gp.keepers {
		if i >= count {
			break
		}
		kid, err := peer.IDB58Decode(kp.keeperID)
		if err != nil {
			continue
		}
		if localNode.PeerHost.Network().Connectedness(kid) == inet.Connected {
			conKeepers = append(conKeepers, kp.keeperID)
		} else {
			if !sc.ConnectTo(context.Background(), localNode, kp.keeperID) { //连接不上此keeper
				unconKeepers = append(unconKeepers, kp.keeperID)
				continue
			}
			conKeepers = append(conKeepers, kp.keeperID)
		}
	}
	if len(gp.keepers) < count {
		return unconKeepers, conKeepers, ErrNoEnoughKeeper
	}
	return unconKeepers, conKeepers, nil
}

//GetLocalProviders 从本地PeersInfo中 返回本节点provider信息,并且检查连接状况
func GetLocalProviders(userID string) ([]string, []string, error) {
	gp := getGroup(userID)
	return gp.getLocalProviders()
}

func (gp *groupInfo) getLocalProviders() ([]string, []string, error) {
	if gp == nil {
		return nil, nil, ErrGroupServiceNotReady
	}
	state, err := getState(gp.userid)
	if err != nil {
		return nil, nil, err
	}
	if state < groupStarted {
		return nil, nil, ErrGroupServiceNotReady
	}
	count := len(gp.providers)
	unconPro := make([]string, 0, count)
	conPro := make([]string, 0, count)
	for _, provider := range gp.providers {
		pid, err := peer.IDB58Decode(provider.providerID)
		if err != nil {
			continue
		}
		if localNode.PeerHost.Network().Connectedness(pid) == inet.Connected {
			conPro = append(conPro, provider.providerID)
		} else {
			if !sc.ConnectTo(context.Background(), localNode, provider.providerID) { //连接不上此provider
				unconPro = append(unconPro, provider.providerID)
				continue
			}
			conPro = append(conPro, provider.providerID)
		}
	}
	return unconPro, conPro, nil
}

//GetProviders 返回provider信息，
func GetProviders(userID string, count int) ([]string, error) {
	gp := getGroup(userID)
	return gp.getProviders(count)
}

func (gp *groupInfo) getProviders(count int) ([]string, error) {
	if gp == nil {
		return nil, ErrGroupServiceNotReady
	}
	state, err := getState(gp.userid)
	if err != nil {
		return nil, err
	}
	if state < groupStarted {
		return nil, ErrGroupServiceNotReady
	}
	var providers []string
	//小于零，则返回内存暂存所有即可
	if count < 0 {
		for _, provider := range gp.providers {
			_, err := peer.IDB58Decode(provider.providerID)
			if err != nil {
				continue
			}
			providers = append(providers, provider.providerID)
		}
		return providers, nil
	}

	for _, provider := range gp.providers {
		pid, err := peer.IDB58Decode(provider.providerID)
		if err != nil {
			continue
		}
		if localNode.PeerHost.Network().Connectedness(pid) != inet.Connected {
			if !sc.ConnectTo(context.Background(), localNode, provider.providerID) { //连接不上此provider
				log.Println("Cannot connect this provider:", provider.providerID)
				continue
			}
		}
		providers = append(providers, provider.providerID)
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
func GetBlockProviders(userID, blockID string) (string, int, error) {
	gp := getGroup(userID)
	return gp.getBlockProviders(blockID)
}

func (gp *groupInfo) getBlockProviders(blockID string) (string, int, error) {
	if gp == nil {
		return "", 0, ErrGroupServiceNotReady
	}
	var pidstr string
	var offset int
	state, err := getState(gp.userid)
	if err != nil {
		return "", 0, err
	}
	if state < groupStarted {
		return "", 0, ErrGroupServiceNotReady
	}
	kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.Local, metainfo.SyncTypeBlock)
	if err != nil {
		return "", 0, err
	}
	blockMeta := kmBlock.ToString()
	for i, kp := range gp.keepers {
		if sc.ConnectTo(context.Background(), localNode, kp.keeperID) {
			pidAndOffset, err := getKeyFrom(blockMeta, kp.keeperID)
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
			} else if err != nil && i >= len(gp.keepers)-1 {
				return "", 0, ErrNoProviders
			}
		} else if i >= len(gp.keepers)-1 {
			return "", 0, ErrNoProviders
		}
	}
	return pidstr, offset, nil
}

func (gp *groupInfo) putDataMetaToKeepers(blockID string, provider string, offset int) error {
	if gp == nil {
		return ErrGroupServiceNotReady
	}
	state, err := getState(gp.userid)
	if err != nil {
		return err
	}
	if state < groupStarted {
		return ErrGroupServiceNotReady
	}
	kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.BlockMetaInfo, metainfo.SyncTypeBlock)
	if err != nil {
		log.Println("construct put blockMeta KV error :", err)
		return err
	}
	metaValue := provider + metainfo.DELIMITER + strconv.Itoa(offset)
	var count int
	for _, keeper := range gp.keepers {
		_, err = sendMetaRequest(kmBlock, metaValue, keeper.keeperID)
		if err != nil {
			log.Println("send metaMessage to ", keeper.keeperID, " error :", err)
			count++
		}
	}
	if count == len(gp.keepers) {
		return ErrNoKeepers
	}
	return nil
}

//删除块
func (gp *groupInfo) deleteBlocksFromProvider(blockID string, updateMeta bool) error {
	if gp == nil {
		return ErrGroupServiceNotReady
	}
	state, err := getState(gp.userid)
	if err != nil {
		return err
	}
	if state < groupStarted {
		return ErrGroupServiceNotReady
	}
	provider, _, err := gp.getBlockProviders(blockID)
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

	for _, kp := range gp.keepers { //从keeper上删除blockMeta
		go sendMetaRequest(km, "", kp.keeperID)
	}

	return nil
}
