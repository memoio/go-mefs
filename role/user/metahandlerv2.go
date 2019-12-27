package user

import (
	"context"
	"log"
	"strings"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

// HandlerV2 User角色回调接口的实现，
type HandlerV2 struct {
	Role string
}

// HandleMetaMessage User角色层metainfo的回调函数,传入对方节点发来的kv，和对方节点的peerid
//没有返回值时，返回complete，或者返回规定信息
func (user *HandlerV2) HandleMetaMessage(dt int, metaKey string, metaValue []byte, from string) ([]byte, error) {
	km, err := metainfo.GetKeyMeta(metaKey)
	if err != nil {
		return nil, err
	}
	keytype := km.GetDType()

	gService := getGroup(km.GetMid())
	if gService == nil {
		return nil, ErrLfsServiceNotReady
	}

	switch keytype {
	case metainfo.UserInit: //handle init response from keeper
		if dt == metainfo.Put {
			go handleUserInit(km, metaValue, from)
		}
	default: //没有匹配的信息，报错
		return nil, metainfo.ErrWrongType
	}
	return []byte(metainfo.MetaHandlerComplete), nil
}

// handleUserInitRes 收到keeper回应的初始化信息，将value中的keeper provider信息整理到备选信息中
// key: userID/"User_Init_Res"/keepercount/providercount,
// value: kid1kid2..../pid1pid2
func handleUserInit(km *metainfo.KeyMeta, metaValue []byte, from string) {
	gp := getGroup(km.GetMid())
	if gp == nil {
		return
	}

	gp.initResMutex.Lock()
	defer gp.initResMutex.Unlock()

	if gp.state == collecting { //收集信息阶段，才继续
		log.Println("Receive: InitResponse，from：", from, ", value is：", metaValue)
		splitedMeta := strings.Split(string(metaValue), metainfo.DELIMITER)
		if len(splitedMeta) != 2 {
			return
		}
		//把keeper信息和provider信息加入到备选中
		keepers := splitedMeta[0]
		for i := 0; i < len(keepers)/utils.IDLength; i++ {
			kid := keepers[i*utils.IDLength : (i+1)*utils.IDLength]
			_, err := peer.IDB58Decode(kid)
			if err != nil {
				continue
			}
			if !utils.CheckDup(gp.tempKeepers, kid) {
				continue
			}
			if localNode.Data.Connect(context.Background(), kid) {
				gp.tempKeepers = append(gp.tempKeepers, kid)
			}
		}
		providers := splitedMeta[1]
		for i := 0; i < len(providers)/utils.IDLength; i++ {
			pid := providers[i*utils.IDLength : (i+1)*utils.IDLength]
			if !utils.CheckDup(gp.tempProviders, pid) {
				continue
			}
			if localNode.Data.Connect(context.Background(), pid) {
				gp.tempProviders = append(gp.tempProviders, pid)
			}
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
