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

// GetBlockProviders 从provider获取数据块的元数据，传入数据块id号
//返回值为保存数据块的pid，offset，错误信息
func GetBlockProviders(userID, blockID string) (string, int, error) {
	u := GetUser(userID)
	if !u.online {
		return "", 0, ErrLfsServiceNotReady
	}
	return u.gInfo.getBlockProviders(blockID)
}

func (g *groupInfo) getBlockProviders(blockID string) (string, int, error) {
	var pidstr string
	var offset int

	kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.Local, metainfo.SyncTypeBlock)
	if err != nil {
		return "", 0, err
	}
	blockMeta := kmBlock.ToString()
	for _, kp := range g.keepers {
		pidAndOffset, err := getKeyFrom(blockMeta, kp.keeperID)
		if err != nil || pidAndOffset == nil {
			continue
		}
		//成功收到
		splitedValue := strings.Split(string(pidAndOffset), metainfo.DELIMITER)
		if len(splitedValue) < 2 {
			continue
		}
		pidstr = splitedValue[0]
		offset, err = strconv.Atoi(splitedValue[1])
		if err != nil {
			log.Println("Offset decode error-", pidstr, err)
			continue
		}
		pid, err := peer.IDB58Decode(pidstr)
		if err != nil {
			log.Println("Wrong format providerID-", pidstr, err)
			continue
		}
		if localNode.PeerHost.Network().Connectedness(pid) != inet.Connected {
			if !sc.ConnectTo(context.Background(), localNode, pidstr) { //连接不上此provider
				log.Println("Cannot connect to blockprovider-", pidstr)
				return pidstr, offset, ErrNoProviders
			}
		}
		return pidstr, offset, nil
	}
	return "", 0, ErrNoProviders
}

//GetKeepers 返回本节点拥有的keeper信息，并且检查连接状况 传入参数为返回信息的数量，-1为全部
func GetKeepers(userID string, count int) ([]string, []string, error) {
	u := GetUser(userID)
	if !u.online {
		return nil, nil, ErrLfsServiceNotReady
	}
	return u.gInfo.getKeepers(count)
}

func (g *groupInfo) getKeepers(count int) ([]string, []string, error) {
	num := count
	if count < 0 {
		num = len(g.keepers)
	}

	unconKeepers := make([]string, 0, num)
	conKeepers := make([]string, 0, num)

	i := 0
	for _, kp := range g.keepers {
		if i >= num {
			break
		}
		kid, err := peer.IDB58Decode(kp.keeperID)
		if err != nil {
			continue
		}
		if localNode.PeerHost.Network().Connectedness(kid) == inet.Connected {
			conKeepers = append(conKeepers, kp.keeperID)
			i++
		} else {
			if !sc.ConnectTo(context.Background(), localNode, kp.keeperID) { //连接不上此keeper
				unconKeepers = append(unconKeepers, kp.keeperID)
				continue
			}
			conKeepers = append(conKeepers, kp.keeperID)
			i++
		}
	}

	if len(conKeepers) < num && count > 0 {
		return conKeepers, unconKeepers, ErrNoEnoughKeeper
	}

	return conKeepers, unconKeepers, nil
}

//GetProviders 返回provider信息，
func GetProviders(userID string, count int) ([]string, []string, error) {
	u := GetUser(userID)
	if !u.online {
		return nil, nil, ErrLfsServiceNotReady
	}
	return u.gInfo.getProviders(count)
}

func (g *groupInfo) getProviders(count int) ([]string, []string, error) {
	num := count
	if count < 0 {
		num = len(g.providers)
	}

	i := 0

	unconPro := make([]string, 0, num)
	conPro := make([]string, 0, num)

	for _, provider := range g.providers {
		if i >= num {
			break
		}
		pid, err := peer.IDB58Decode(provider.providerID)
		if err != nil {
			continue
		}
		if localNode.PeerHost.Network().Connectedness(pid) == inet.Connected {
			conPro = append(conPro, provider.providerID)
			i++
		} else {
			if !sc.ConnectTo(context.Background(), localNode, provider.providerID) { //连接不上此provider
				unconPro = append(unconPro, provider.providerID)
				continue
			} else {
				conPro = append(conPro, provider.providerID)
				i++
			}

		}
	}

	if len(conPro) < num && count > 0 {
		return conPro, unconPro, ErrNoEnoughProvider
	}

	return conPro, unconPro, nil
}

func (g *groupInfo) putDataToKeepers(key *metainfo.KeyMeta, value string) error {
	if g.state < groupStarted {
		return ErrLfsServiceNotReady
	}

	var count int
	for _, keeper := range g.keepers {
		_, err := sendMetaRequest(key, value, keeper.keeperID)
		if err != nil {
			log.Println("send metaMessage to ", keeper.keeperID, " error :", err)
			count++
		}
	}
	if count == len(g.keepers) {
		return ErrNoKeepers
	}
	return nil
}

func (g *groupInfo) putDataToProviders(key *metainfo.KeyMeta, value string) error {
	if g.state < groupStarted {
		return ErrLfsServiceNotReady
	}

	var count int
	for _, provider := range g.providers {
		_, err := sendMetaRequest(key, value, provider.providerID)
		if err != nil {
			log.Println("send metaMessage to ", provider.providerID, " error :", err)
			count++
		}
	}
	if count == len(g.providers) {
		return ErrNoProviders
	}
	return nil
}

func (g *groupInfo) putDataMetaToKeepers(blockID string, provider string, offset int) error {
	if g.state < groupStarted {
		return ErrLfsServiceNotReady
	}
	kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.BlockMetaInfo, metainfo.SyncTypeBlock)
	if err != nil {
		log.Println("construct put blockMeta KV error :", err)
		return err
	}
	metaValue := provider + metainfo.DELIMITER + strconv.Itoa(offset)
	return g.putDataToKeepers(kmBlock, metaValue)
}

//删除块
func (g *groupInfo) deleteBlocksFromProvider(blockID string, updateMeta bool) error {
	if g.state < groupStarted {
		return ErrLfsServiceNotReady
	}
	provider, _, err := g.getBlockProviders(blockID)
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
			log.Println("Cannot delete Block-", blockID, ErrCannotConnectNetwork)
			return ErrCannotConnectNetwork
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

	for _, kp := range g.keepers { //从keeper上删除blockMeta
		go sendMetaRequest(km, "", kp.keeperID)
	}

	return nil
}
