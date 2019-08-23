/*
该模块负责和tendermint接口交互
包含keeper组的定义，各种交互函数
*/

package keeper

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/memoio/go-mefs/utils/metainfo"

	"github.com/memoio/go-mefs/consensus/rpc"
	sc "github.com/memoio/go-mefs/utils/swarmconnect"

	tnode "github.com/memoio/go-mefs/consensus/tendermint"
	config "github.com/memoio/go-mefs/config"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	inet "github.com/libp2p/go-libp2p-core/network"
	peer "github.com/libp2p/go-libp2p-core/peer"
)

//扩展的U-K-P对应关系，加入组的概念，多个keeper组成一个组，组内进行数据同步
//key为组名，格式为hash（uid+时间戳） value为组信息结构体
var GInfo map[string]*GroupsInfo

//为本节点初始化tendermint节点启动需要的信息，放在节点信息的结构体中
//初始化目录、获取id、ip、P2PPort、RPCPort、节点公钥
//传入需要初始化节点的组名
func initTendermintInfo(groupid string) {
	localkeeper, err := getLocalKeeperInGroup(groupid)
	if err != nil {
		log.Printf("User: %s Tendermint localKeeper not exist", groupid)
		fmt.Println(err)
		return
	}

	thisGroupsInfo, ok := getGroupsInfo(groupid)
	if !ok {
		fmt.Println(ErrNoGroupsInfo)
		return
	}

	thisGroupsInfo.P2PListener = getFreePort()
	thisGroupsInfo.RPCListener = getFreePort()

	nodepath := getPath(groupid)                   //设置路径
	time := getTendermintTime()                    //所有节点要求时间一致
	node := tnode.NewNode(nodepath, groupid, time) //初始化节点
	if err := node.InitHome(); err != nil {        //初始化节点目录，若已经初始化过，则不做任何操作
		log.Fatal(err)
	}
	id, err := node.GetID() //获取id
	if err != nil {
		fmt.Println("get ID err")
		fmt.Println(err)
	}
	validator, err := node.GetPublicKey() //获取本节点公钥
	if err != nil {
		fmt.Println("get PubKey err")
		fmt.Println(err)
	}
	//填数据
	thisGroupsInfo.TendermintNode = node
	localkeeper.ID = id
	localkeeper.IP = "127.0.0.1"
	localkeeper.PubKey = validator
	localkeeper.P2PPort = thisGroupsInfo.P2PListener.Addr().(*net.TCPAddr).Port
	localkeeper.RpcPort = thisGroupsInfo.RPCListener.Addr().(*net.TCPAddr).Port

}

//为本节点启动tendermint实例 传入组名
//启动条件为同组节点的信息收集完全（检查公钥）
func startTendermintCore(groupid string) {
	localID := localNode.Identity.Pretty()
	var localkeeper *KeeperInGroup
	var validators []string
	var peers []string
	//判断是否满足启动条件,收集公钥信息，为接下来修改validator字段做准备

	thisGroupsInfo, ok := getGroupsInfo(groupid)
	if !ok {
		fmt.Println("本地PInfo还没构造好")
		return
	}
	for _, keeper := range thisGroupsInfo.Keepers { //检查同组所有keeper的信息
		if keeper.PubKey == "" {
			fmt.Printf("Group %s 没有收到 %s 的信息，不启动tendermint\n", groupid, keeper.KID)
			return
		}
		if strings.Compare(localID, keeper.KID) == 0 {
			localkeeper = keeper
		} else {
			peers = append(peers, keeper.ID+"@"+keeper.IP+":"+strconv.Itoa(keeper.P2PPort))
		}
		validators = append(validators, keeper.PubKey)
	}
	tnode := thisGroupsInfo.TendermintNode
	//修改validator字段，保证同组节点的创世块一样,注意顺序也要一样
	sort.Strings(validators)
	tnode.AlterValidator(validators)
	//tendermint core 跑起
	err := thisGroupsInfo.RPCListener.Close()
	if err != nil {
		fmt.Println(err)
	}
	err = thisGroupsInfo.P2PListener.Close() //释放初始化时占用的端口
	if err != nil {
		fmt.Println(err)
	}
	thisGroupsInfo.RunLock.Lock()
	defer thisGroupsInfo.RunLock.Unlock()
	if !tnode.IsRunning {
		fmt.Printf("Start Group -%s\tP2PPort:%d\tRPCPort:%d\n", tnode.GroupID, localkeeper.P2PPort, localkeeper.RpcPort)
		err := tnode.Run(localkeeper.P2PPort, localkeeper.RpcPort, peers)
		thisGroupsInfo.Client = rpc.GetLocalClient(tnode.Node)
		if err != nil { //当节点重复启动的时候会报错
			fmt.Println("node Run err:", err)
			log.Println(err)
		}
		//此时通知User，Tendermint core已经启动
		localkeeper, err := getLocalKeeperInGroup(groupid)
		if err != nil {
			fmt.Println("getLocalKeeperInGroup() error!", err)
			return
		}
		kmRes, err := metainfo.NewKeyMeta(groupid, metainfo.Local, metainfo.SyncTypeBft)
		if err != nil {
			fmt.Println(err)
			return
		}
		resValue := strings.Join([]string{"bft", localkeeper.IP + ":" + strconv.Itoa(localkeeper.P2PPort), localkeeper.IP + ":" + strconv.Itoa(localkeeper.RpcPort)}, metainfo.DELIMITER)
		err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(kmRes.ToString(), resValue, "local") //放在本地供User或Provider启动的时候查询
		if err != nil {
			fmt.Println(err)
		}
		kmRes.SetKeyType(metainfo.UserInitNotifRes)
		_, err = sendMetaRequest(kmRes, resValue, groupid)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Printf("The Group %s is already started\n", tnode.GroupID)
		return
	}
}

//重新启动tendermint节点，传入组名
//重新启动时 应该保证Pinfo构造完成，并且tendermint节点已初始化过
//重新找端口进行填充->向同组其他节点进行同步->等待其他节点发来的信息->启动
func restartTendermintCore(groupid string) {
	initTendermintInfo(groupid) //对节点进行初始化并且将相应的信息填到结构体中
	localkeeper, err := getLocalKeeperInGroup(groupid)
	if err != nil {
		fmt.Println("getLocalKeeperInGroup() error!", err)
		return
	}
	thisGroupsInfo, ok := getGroupsInfo(groupid)
	if !ok {
		fmt.Println("没有取到groupinfo")
		return
	}
	var failKeepers []string
	for _, keeper := range thisGroupsInfo.Keepers {
		if strings.Compare(keeper.KID, localNode.Identity.Pretty()) == 0 {
			continue
		}
		kid, _ := peer.IDB58Decode(keeper.KID)
		if localNode.PeerHost.Network().Connectedness(kid) != inet.Connected {
			if !sc.ConnectTo(context.Background(), localNode, keeper.KID) { //连接不上此keeper
				failKeepers = append(failKeepers, keeper.KID)
			}
		}
	}

	if len(failKeepers) > 0 { //若有连接不上的节点等待10s尝试重新链接
		time.Sleep(10 * time.Second)
		for _, kidstr := range failKeepers {
			kid, _ := peer.IDB58Decode(kidstr)
			if localNode.PeerHost.Network().Connectedness(kid) != inet.Connected {
				if !sc.ConnectTo(context.Background(), localNode, kidstr) { //连接不上此keeper
					fmt.Println("重启过程链接keeper失败，keeper：", kid)
				}
			}
		}
	}

	km, err := metainfo.NewKeyMeta(groupid, metainfo.TendermintRestart) //构造tendermint重启信息，发送给同组节点
	if err != nil {
		fmt.Println("restartTendermintCore()NewKeyMeta()err", err)
	}
	metavalue := strings.Join([]string{localkeeper.ID, localkeeper.IP, localkeeper.PubKey, strconv.Itoa(localkeeper.P2PPort), strconv.Itoa(localkeeper.RpcPort)}, metainfo.DELIMITER)
	for _, keeper := range thisGroupsInfo.Keepers {
		if strings.Compare(keeper.KID, localNode.Identity.Pretty()) != 0 { //去掉本地节点
			err = sendMetaMessage(km, metavalue, keeper.KID)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

//找到空闲端口并且占用,返回占用端口的监听器，之后使用的时候可直接close
func getFreePort() *net.TCPListener {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println("FreePort", l.Addr().String())
	return l
}

//获取工作路径
func getPath(groupid string) string {
	p := os.Getenv("MEFS_PATH")
	if p == "" {
		p = path.Join(os.Getenv("HOME"), config.DefaultPathName)
	}
	if p == "" {
		log.Print("mefsPath == ''")
		return ""
	}
	return path.Join(p, "tendermint", groupid)
}
