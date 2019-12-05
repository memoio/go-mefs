package user

import (
	"log"
	"strings"

	"github.com/memoio/go-mefs/utils/metainfo"
)

// HandlerV2 User角色回调接口的实现，
type HandlerV2 struct {
	Role string
}

// HandleMetaMessage User角色层metainfo的回调函数,传入对方节点发来的kv，和对方节点的peerid
//没有返回值时，返回complete，或者返回规定信息
func (user *HandlerV2) HandleMetaMessage(metaKey, metaValue, from string) (string, error) {
	km, err := metainfo.GetKeyMeta(metaKey)
	if err != nil {
		return "", err
	}
	keytype := km.GetKeyType()

	gService := getGroupService(km.GetMid())
	if gService == nil {
		return "", ErrGroupServiceNotReady
	}

	switch keytype {
	case metainfo.UserInitReq:
		log.Println("keytype：UserInitReq 不处理")
	case metainfo.UserInitRes: //handle init response from keeper
		go gService.handleUserInitRes(km, metaValue, from)
	case metainfo.UserInitNotifRes: //user init response
		go gService.handleUserInitNotifRes(metaValue, from)
	case metainfo.Test:
		handleTest(km)
	case metainfo.GetBlock:
		log.Println("getBlock: ", km.ToString())
	case metainfo.PutBlock:
		log.Println("putBlock: ", km.ToString())
	default: //没有匹配的信息，报错
		return "", metainfo.ErrWrongType
	}
	return metainfo.MetaHandlerComplete, nil
}

// handleUserInitRes 收到keeper回应的初始化信息，将value中的keeper provider信息整理到备选信息中
// key: userID/"User_Init_Res"/keepercount/providercount,
// value: kid1kid2..../pid1pid2
func (gp *groupService) handleUserInitRes(km *metainfo.KeyMeta, metaValue, from string) {
	gp.initResMutex.Lock()
	defer gp.initResMutex.Unlock()
	userState, err := getUserState(gp.userid)
	if err != nil {
		log.Println("handleUserInitRes()getUserState()err:", err, "userid:", gp.userid)
	}
	if userState == collecting { //收集信息阶段，才继续
		log.Println("Receive: InitResponse，from：", from, ", value is：", metaValue)
		splitedMeta := strings.Split(metaValue, metainfo.DELIMITER)
		if len(splitedMeta) == 2 {
			gp.addKeepersAndProviders(splitedMeta[0], splitedMeta[1]) //把keeper信息和provider信息加入到备选中
		}
	}
}

//handleUserInitNotifRes 初始化第四次握手的回调，收到的信息是keeper发来的bft信息
func (gp *groupService) handleUserInitNotifRes(metaValue, from string) {
	gp.initResMutex.Lock()
	defer gp.initResMutex.Unlock()
	userState, err := getUserState(gp.userid)
	if err != nil {
		log.Println("handleUserInitNotifRes()getUserState()err:", err, "userid:", gp.userid)
	}
	if userState == collectCompleted && len(gp.keepers) == gp.keeperSLA { //收集信息完成阶段，继续
		err := gp.keeperConfirm(from, metaValue)
		if err != nil {
			log.Println("handleUserInitNotifRes()error", err)
		}
	}
}

// GetRole gets role
// 获取这个节点的角色信息，返回错误说明User还没有启动好
func (user *HandlerV2) GetRole() (string, error) {
	return user.Role, nil
}

func handleTest(km *metainfo.KeyMeta) {
	log.Println("测试用回调函数")
	log.Println("km.mid:", km.GetMid())
	log.Println("km.options", km.GetOptions())
}
