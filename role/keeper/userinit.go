package keeper

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils/metainfo"
)

// userInitInMem 在PInfo中进行查找，找到user节点信息，则看是否是需要添加K/P。否则加入WaitingList等待确认
//返回keeper和provider id组成的字符串，格式为 kid1kid2../pid1pid2..
func userInitInMem(userID string, keeperCount, providerCount int) (string, error) {
	thisGroupsInfo, ok := getGroupsInfo(userID)
	if !ok { //内存中没有该节点信息
		localPeerInfo.UserCache.Add(userID, time.Now().Unix()) //先只加入到WaitingList，等待其确认
		return "", nil
	}

	//如果内存中直接有该节点信息,说明该user属于本keeper,并不需要初始化，看是否需要添加K和P
	//localID := localNode.Identity.Pretty() //本地id

	var responseExisted bytes.Buffer //此变量暂存返回的peerIDs
	kmKid, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypeKid)
	if err != nil {
		return "", err
	}
	kmPid, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypePid)
	if err != nil {
		return "", err
	}

	//本节点记录的该user拥有的keeper数量小于要求的数量
	//对本节点连接的所有keeper进行循环，找到一个不属于该user的keeper，分给该user 加入PInfo
	if len(thisGroupsInfo.Keepers) < keeperCount {
		for i := 0; i < len(localPeerInfo.Keepers); i++ {
			var j int
			for j = 0; j < len(thisGroupsInfo.Keepers); j++ { //找到user
				if strings.Compare(localPeerInfo.Keepers[i], thisGroupsInfo.Keepers[j].KID) == 0 {
					break
				}
			}
			if j == len(thisGroupsInfo.Keepers) {
				keeper := &KeeperInGroup{
					KID: localPeerInfo.Keepers[i],
				}
				thisGroupsInfo.Keepers = append(thisGroupsInfo.Keepers, keeper)
			}
			if len(thisGroupsInfo.Keepers) == keeperCount {
				break
			}
		}
	}

	//本节点记录的该user拥有的provider数量小于要求的数量
	//处理方法同上
	if len(thisGroupsInfo.Providers) < providerCount {
		for i := 0; i < len(localPeerInfo.Providers); i++ {
			var j int
			for j = 0; j < len(thisGroupsInfo.Providers); j++ {
				if strings.Compare(localPeerInfo.Providers[i], thisGroupsInfo.Providers[j]) == 0 {
					break
				}
			}
			if j == len(thisGroupsInfo.Providers) {
				thisGroupsInfo.Providers = append(thisGroupsInfo.Providers, localPeerInfo.Providers[i])
			}
			if len(thisGroupsInfo.Providers) == providerCount {
				break
			}
		}
	}

	//如果无法找到足够的Keeper和provider(说明User想要扩大Keeper数量？这种情况需要后续处理）
	//目前只要内存中有就分配，毕竟已经证明它是本Keeper的user

	//将本地保存的这个user的K-P信息整理，更新，并进行返回
	for i := 0; i < len(thisGroupsInfo.Keepers); i++ {
		responseExisted.WriteString(thisGroupsInfo.Keepers[i].KID)
		err := localNode.Routing.(*dht.IpfsDHT).CmdAppendTo(kmKid.ToString(), thisGroupsInfo.Keepers[i].KID, "local")
		if err != nil {
			return "", err
		}
	}
	responseExisted.WriteString(metainfo.DELIMITER) //keeper和provider之间以斜杠区分
	for i := 0; i < len(thisGroupsInfo.Providers); i++ {
		responseExisted.WriteString(thisGroupsInfo.Providers[i])
		err := localNode.Routing.(*dht.IpfsDHT).CmdAppendTo(kmPid.ToString(), thisGroupsInfo.Providers[i], "local")
		if err != nil {
			return "", err
		}
	}
	return responseExisted.String(), nil
}

func userInitInLocal(userID string, keeperCount, providerCount int) (string, error) {
	_, ok := localPeerInfo.UserCache.Get(userID)
	if !ok { //看看再WaitingList里是不是有记录
		localPeerInfo.UserCache.Add(userID, time.Now().Unix())
	}
	var flag int
	kmKid, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypeKid)
	if err != nil {
		return "", err
	}
	kmPid, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypePid)
	if err != nil {
		return "", err
	}
	tempKeepers := make([]string, 0, keeperCount)
	tempProvider := make([]string, 0, providerCount)
	localID := localNode.Identity.Pretty()
	var responseExisted bytes.Buffer //此变量暂存返回的peerIDs

	if kids, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(kmKid.ToString(), ""); kids != nil && err == nil { //如果DHT中有这个节点的信息，user的init是输错的
		if remain := len(kids) % IDLength; remain != 0 {
			kids = kids[:len(kids)-remain]
		}
		for i := 0; i < len(kids)/IDLength; i++ { //在dht返回结果中，看该user是否属于本节点
			keeperID := string(kids[i*IDLength : (i+1)*IDLength])
			if keeperID == localID { //此User是本节点的
				flag = 1
			}
		}
		if flag == 1 { //只处理keeper列表里有自己的情况
			responseExisted.WriteString(localID)
			tempKeepers = append(tempKeepers, localID)
			for i := 0; i < len(kids)/IDLength; i++ {
				if len(tempKeepers) == keeperCount {
					break
				}
				keeperID := string(kids[i*IDLength : (i+1)*IDLength])
				kmRole, err := metainfo.NewKeyMeta(keeperID, metainfo.Local, metainfo.SyncTypeRole)
				if err != nil {
					return "", err
				}
				if keeperID == localID { //本节点已加入
					continue
				}
				if result, _ := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(kmRole.ToString(), keeperID); string(result) == metainfo.RoleKeeper { //表示此结点仍然连在网上，且角色未变
					responseExisted.WriteString(keeperID)
					tempKeepers = append(tempKeepers, keeperID)
				}
			}

			if len(tempKeepers) < keeperCount { //如果已连接的Keeper数目不够，用内存中的补足，即重新分配
				for j := 0; j < len(localPeerInfo.Keepers); j++ {
					var k int
					for k = 0; k < len(tempKeepers); k++ {
						if localPeerInfo.Keepers[j] == tempKeepers[k] {
							break
						}

					}

					if k == len(tempKeepers) { //新的keeper，可以加入
						keeper := localPeerInfo.Keepers[j]
						responseExisted.WriteString(keeper)
						tempKeepers = append(tempKeepers, keeper)
						err = localNode.Routing.(*dht.IpfsDHT).CmdAppendTo(kmKid.ToString(), keeper, "local")
						if err != nil {
							return "", err
						}
					}
				}

			}

			//补不足也就这样处理了，毕竟已确认是本节点User

			responseExisted.WriteString(metainfo.DELIMITER)

			if pids, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(kmPid.ToString(), ""); pids != nil && err == nil {
				if remain := len(pids) % IDLength; remain != 0 {
					pids = pids[:len(pids)-remain]
				}

				for i := 0; i < len(pids)/IDLength; i++ {
					providerID := string(pids[i*IDLength : (i+1)*IDLength])
					kmRole, err := metainfo.NewKeyMeta(providerID, metainfo.Local, metainfo.SyncTypeRole)
					if err != nil {
						return "", err
					}
					result, _ := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(kmRole.ToString(), providerID) //表示此结点仍然连在网上，且角色未变
					if string(result) == metainfo.RoleProvider {
						tempProvider = append(tempProvider, providerID)
						responseExisted.WriteString(providerID)
					}
				}
			}

			if len(tempProvider) < providerCount { //如果已连接的Provider数目不够，补足
				for j := 0; j < len(localPeerInfo.Providers); j++ {
					var k int
					for k = 0; k < len(tempProvider); k++ {
						if localPeerInfo.Providers[j] == tempProvider[k] {
							break
						}
					}

					if k == len(tempProvider) { //不重复的provider
						providerID := localPeerInfo.Providers[j]
						responseExisted.WriteString(providerID)
						err = localNode.Routing.(*dht.IpfsDHT).CmdAppendTo(kmPid.ToString(), providerID, "local")
						if err != nil {
							return "", err
						}
						tempProvider = append(tempProvider, providerID)
					}
				}
			}
			//同理，无法补足也没办法
			return responseExisted.String(), nil
		}
		//这种情况暂时不处理
		return "", ErrUnmatchedPeerID
	}
	return "", nil

}

// newUserInit 为新的user进行初始化操作，返回keeper和provider的信息
func newUserInit(userID string, keeperCount, providerCount int) (string, error) {
	localID := localNode.Identity.Pretty()
	kmKid, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypeKid)
	if err != nil {
		return "", err
	}
	kmPid, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypePid)
	if err != nil {
		return "", err
	}
	_, ok := localPeerInfo.UserCache.Get(userID)
	if !ok {
		localPeerInfo.UserCache.Add(userID, time.Now().Unix())
	}

	var newResponse bytes.Buffer
	newResponse.WriteString(localID)                                                    //先把自己加入进去
	err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(kmKid.ToString(), localID, "local") //先将本节点加入Keeper列表
	if err != nil {
		return "", err
	}
	keeperCount--
	//填写keeper信息
	for _, keeperID := range localPeerInfo.Keepers {
		if keeperCount == 0 {
			break
		}
		kmRole, err := metainfo.NewKeyMeta(keeperID, metainfo.Local, metainfo.SyncTypeRole)
		if err != nil {
			return "", err
		}
		result, _ := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(kmRole.ToString(), keeperID) //确认这个节点的角色信息
		if string(result) == metainfo.RoleKeeper {
			newResponse.WriteString(keeperID) //加入返回信息中
			err = localNode.Routing.(*dht.IpfsDHT).CmdAppendTo(kmKid.ToString(), keeperID, "local")
			if err != nil {
				return "", err
			}
			keeperCount--
		}
	}

	newResponse.WriteString(metainfo.DELIMITER) //分隔符

	//填写provider信息
	for _, providerID := range localPeerInfo.Providers {
		if providerCount == 0 {
			break
		}
		kmRole, err := metainfo.NewKeyMeta(providerID, metainfo.Local, metainfo.SyncTypeRole)
		if err != nil {
			return "", err
		}
		result, _ := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(kmRole.ToString(), providerID) //确认这个节点的角色信息
		if string(result) == metainfo.RoleProvider {
			newResponse.WriteString(providerID)
			err = localNode.Routing.(*dht.IpfsDHT).CmdAppendTo(kmPid.ToString(), providerID, "local")
			if err != nil {
				return "", err
			}
			providerCount--
		}
	}

	return newResponse.String(), nil
}

//fillPinfo user初始化时填充Pinfo的keeper和provider信息
//检查本地节点是否在这个组中，然后进行填充。填充完成后，会为本组进行tendermint信息的初始化，并且将信息同步到同组其他节点
func fillPinfo(groupid string, keepers []*KeeperInGroup, providers []string, from string) {
	var localkeeper *KeeperInGroup
	for _, keeper := range keepers { //获取本地keeper的组内信息
		if strings.Compare(keeper.KID, localNode.Identity.Pretty()) == 0 {
			localkeeper = keeper
		}
	}
	if localkeeper == nil { //本地节点可能不在这个组中，则直接返回
		fmt.Println(ErrNotKeeperInThisGroup)
		return
	}
	tempInfo := &GroupsInfo{
		Keepers:   keepers,
		Providers: providers,
		User:      groupid,
		GroupID:   groupid,
	}
	PInfo.Store(groupid, tempInfo)
	if !localPeerInfo.enableTendermint { //初始化tendermint之前进行判断本节点是否使用tendermint
		kmRes, err := metainfo.NewKeyMeta(groupid, metainfo.Local, metainfo.SyncTypeBft)
		if err != nil {
			fmt.Println(err)
			return
		}
		resValue := "simple"
		err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(kmRes.ToString(), resValue, "local") //放在本地供User或Provider启动的时候查询是否为拜占庭容错节点
		if err != nil {
			fmt.Println(err)
		}
		kmRes.SetKeyType(metainfo.UserInitNotifRes)
		_, err = sendMetaRequest(kmRes, resValue, from)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("本节点不使用Tendermint，GroupID:", groupid)
	}
	return
}
