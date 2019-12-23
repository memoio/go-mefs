package user

import (
	"context"
	"errors"
	"log"
	"os"
	"path"
	"sync"

	mcl "github.com/memoio/go-mefs/bls12"
	config "github.com/memoio/go-mefs/config"
	"github.com/memoio/go-mefs/core"
	"github.com/memoio/go-mefs/repo/fsrepo"
)

var localNode *core.MefsNode

var allUsers sync.Map

// NewUser add a new user
func NewUser(uid string, isInit bool, pwd string, capacity int64, duration int64, price int64, ks int, ps int, rdo bool) (Service, error) {
	if user, ok := allUsers.Load(userID); ok && user != nil {
		return value.(*userInfo), error.New("user is running")
	}

	ctx, cancel := context.WithCancel(context.Background())
	// 读keystore下uid文件
	keypath, err := config.Path("", path.Join("keystore", uid))
	if err != nil {
		return nil, ErrDirNotExist
	}

	_, err = os.Stat(keypath)
	if os.IsNotExist(err) {
		return nil, ErrDirNotExist
	}

	userkey, err := fsrepo.GetPrivateKeyFromKeystore(uid, keypath, pwd)
	if err != nil {
		return nil, err
	}

	ginfo := &groupInfo{
		userID:      uid,
		state:       errState,
		storeDays:   duration,
		storeSize:   capacity,
		storePrice:  price,
		keeperSLA:   ks,
		providerSLA: ps,
		reDeploy:    redeploy,
	}

	lInfo := &LfsInfo{
		userID:     uid,
		context:    ctx,
		cancelFunc: cancel,
		privateKey: userkey.PrivateKey,
		gInfo : ginfo,
		state:      errState,
	}

	allUsers.Store(userID, lInfo)

	return lInfo, nil
}

// IsOnline judges
func IsOnline(userID string) bool {
	if user, ok := allUsers.Load(userID); ok && user != nil {
		return value.(*userInfo).online
	}

	return false
}

func getSk(userID string) []byte {
	if user, ok := allUsers.Load(userID); ok && user != nil {
		return value.(*userInfo).privateKey
	}
	return nil
}

func getGroup(userID string) *groupInfo {
	if user, ok := allUsers.Load(userID); ok && user != nil {
		return value.(*userInfo).gInfo
	}
	return nil
}

// KillUser kills
func KillUser(userID string) error {
	if user, ok := allUsers.Load(userID); ok && user != nil {
		user.(*LfsInfo).Stop()
		allUsers.Delete(userID)
		return nil
	}

	return ErrUserNotExist
}

// GetUser gets userInfo
func GetUser(userID string) Service {
	if user, ok := allUsers.Load(userID); ok && user != nil {
		return user.(*LfsInfo)
	}
	return nil
}

// GetUsers gets
func GetUsers() ([]string, []bool, error) {

	var users []string
	var states []bool
	allUsers.Range(func(key, value interface{}) bool {
		id := key.(string)
		us := value.(*LfsInfo)
		users = append(users, id)

		states = append(states, us.)

		return true
	})
	return users, states, nil
}

// PersistBeforeExit is
func PersistBeforeExit() error {
	allUsers.Range(func(key, value interface{}) bool {
		uInfo := value.(*userInfo)
		if !uInfo.online {
			return true
		}
		err = uInfo.Fsync(false)
		if err != nil {
			log.Printf("Sorry, something wrong in persisting for %s: %v\n", uInfo.userID, err)
		} else {
			log.Printf("User %s Persist completed\n", uInfo.userID)
		}
		uInfo.cancelFunc() //释放资源
	})
	return nil
}

// ShowInfo 输出本节点的信息
func ShowInfo(userID string) map[string]string {
	outmap := map[string]string{}
	log.Println(">>>>>>>>>>>>>>ShowInfo>>>>>>>>>>>>>>")
	defer log.Println("================================")
	us := GetUser(userID)
	if us == nil {
		outmap["error: "] = "userService==nil"
		return outmap
	}

	return outmap
}
