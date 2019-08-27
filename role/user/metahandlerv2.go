package user

import (
	"fmt"
	"strings"

	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils/metainfo"
)

// UserHandlerV2 User角色回调接口的实现，
type UserHandlerV2 struct {
	Role string
}

// HandleMetaMessage User角色层metainfo的回调函数,传入对方节点发来的kv，和对方节点的peerid
//没有返回值时，返回complete，或者返回规定信息
func (user *UserHandlerV2) HandleMetaMessage(metaKey, metaValue, from string) (string, error) {
	km, err := metainfo.GetKeyMeta(metaKey)
	if err != nil {
		return "", err
	}
	keytype := km.GetKeyType()

	gService := GetGroupService(km.GetMid())
	if gService == nil {
		return "", ErrGroupServiceNotReady
	}

	switch keytype {
	case metainfo.UserInitReq:
		fmt.Println("keytype：UserInitReq 不处理")
	case metainfo.UserInitRes: //keeper初始化的回复
		go gService.handleUserInitRes(km, metaValue, from)
	case metainfo.UserInitNotifRes: //keeper初始化确认的回复
		go gService.handleUserInitNotifRes(metaValue, from)
	case metainfo.Test:
		handleTest(km)
	case metainfo.GetBlock:
		fmt.Println("getBlock: ", km.ToString())
	case metainfo.PutBlock:
		fmt.Println("putBlock: ", km.ToString())
	default: //没有匹配的信息，报错
		return "", metainfo.ErrWrongType
	}
	return metainfo.MetaHandlerComplete, nil
}

// handleUserInitRes 收到keeper回应的初始化信息，将value中的keeper provider信息整理到备选信息中
// userID/"User_Init_Res"/keepercount/providercount,kid1kid2..../pid1pid2
func (gp *GroupService) handleUserInitRes(km *metainfo.KeyMeta, metaValue, from string) {
	gp.initResMutex.Lock()
	defer gp.initResMutex.Unlock()
	userState, err := GetUserServiceState(gp.Userid)
	if err != nil {
		fmt.Println("handleUserInitRes()GetUserServiceState()err:", err, "userid:", gp.Userid)
	}
	if userState == Collecting { //收集信息阶段，才继续
		//确认和Keeper连接了
		kids, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(km.ToString(), from) //尝试从该节点获取Kids
		if len(kids) == 0 && err != nil {
			fmt.Println("handleUserInitRes()error:", ErrCannotConnectKeeper)
		}
		fmt.Println("Receive: InitResponse，来源：", from, "值：", metaValue)
		splitedMeta := strings.Split(metaValue, metainfo.DELIMITER)
		if len(splitedMeta) == 2 {
			gp.addKeepersAndProviders(splitedMeta[0], splitedMeta[1]) //把keeper信息和provider信息加入到备选中
		}
	}
}

//handleUserInitNotifRes 初始化第四次握手的回调，收到的信息是keeper发来的bft信息
func (gp *GroupService) handleUserInitNotifRes(metaValue, from string) {
	gp.initResMutex.Lock()
	defer gp.initResMutex.Unlock()
	userState, err := GetUserServiceState(gp.Userid)
	if err != nil {
		fmt.Println("handleUserInitNotifRes()GetUserServiceState()err:", err, "userid:", gp.Userid)
	}
	if userState == CollectCompleted && len(gp.localPeersInfo.Keepers) != 0 { //收集信息完成阶段，继续
		err := gp.keeperConfirm(from, metaValue)
		if err != nil {
			fmt.Println("handleUserInitNotifRes()error", err)
		}
	}
}

// 获取这个节点的角色信息，返回错误说明User还没有启动好
func (user *UserHandlerV2) GetRole() (string, error) {
	return user.Role, nil
}

func handleTest(km *metainfo.KeyMeta) {
	fmt.Println("测试用回调函数")
	fmt.Println("km.mid:", km.GetMid())
	fmt.Println("km.options", km.GetOptions())
}
