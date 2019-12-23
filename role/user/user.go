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

type userState int32

var localNode *core.MefsNode

const (
	errState userState = iota
	starting
	collecting
	collectCompleted
	onDeploy
	groupStarted
	bothStarted
)

var stateList = []string{
	"starting", "collecting", "CollectComplited", "onDeploy", "groupStarted", "bothStarted",
}

var allUsers sync.Map

type userInfo struct {
	userID     string
	gInfo      *groupInfo
	lInfo      *lfsInfo
	privateKey []byte
	keySet     *mcl.KeySet
	context    context.Context
	cancelFunc context.CancelFunc
	state      userState
}

// Service defines user's function
type Service interface {
	Start() error
	CreateBucket()
	Upload()
	Download()
	GetBucketInfo()
	GetObjectInfo()
	SetUserState(state userState) error
	GetUserState() (userState, error)
}

// NewUser add a new user
func NewUser(uid string, isInit bool, pwd string, capacity int64, duration int64, price int64, ks int, ps int, rdo bool) Service {
	ctx, cancel := context.WithCancel(context.Background())

	// 读keystore下uid文件
	keypath, err := config.Path("", path.Join("keystore", uid))
	if err != nil {
		return ErrDirNotExist
	}

	_, err = os.Stat(keypath)
	if os.IsNotExist(err) {
		return ErrDirNotExist
	}

	userkey, err := fsrepo.GetPrivateKeyFromKeystore(uid, keypath, pwd)
	if err != nil {
		return err
	}

	uInfo := &userInfo{
		userID:     uid,
		context:    ctx,
		cancelFunc: cancel,
		privateKey: userkey.PrivateKey,
		state:      errState,
	}
	thisInfo, ok := allUsers.LoadOrStore(userID, uInfo)
	if ok {
		return thisInfo.(*userInfo)
	}
	return uInfo
}

// Start starts user's info
func (u *userInfo) Start() error {
	// 证明该user已经启动
	if st, err := getState(uid); err == nil && st >= starting {
		return errors.New("The user is running")
	}

	u.state = starting

	u.gInfo = newGroup(u.userID, duration, capacity, price, ks, ps, rdo)

	has, err = u.gInfo.start(u.context)
	if err != nil {
		log.Println("start group err: ", err)
		return err
	}

	if has {
		// init or send bls config
		mkey, err := loadBLS12Config(userID, u.gInfo.tempKeepers, u.privateKey)

		if err != nil || mkey == nil {
			log.Println("no bls config err: ", err)
			return err
		}
		u.keySet = mkey
	} else {
		mkey, err := userBLS12ConfigInit()
		if err != nil {
			log.Println("init bls config err: ", err)
			return err
		}

		putUserConfig(userID, u.gInfo.tempKeepers, u.privateKey, mkey)

		u.keySet = mkey
	}

	lfs := constructLfsService(us.userid, userkey.PrivateKey)

	err = l.StartLfsService(us.context)
	if err != nil {
		log.Println("StartLfsService()err")
		return err
	}

}

func (u *userInfo) SetUserState(s userState) error {
	u.state = s
	return nil
}

func (u *userInfo) GetUserState(userID string) (userState, error) {
	return u.state
}

func getState(userID string) (userState, error) {
	if user, ok := allUsers.Load(userID); ok && user != nil {
		return value.(*userInfo).state, nil
	}
	return errState, errors.New("no such user")
}

func setState(userID string, s userState) error {
	if user, ok := allUsers.Load(userID); ok && user != nil {
		value.(*userInfo).state = s
	}
	return nil
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

func getLfs(userID string) *lfsInfo {
	if user, ok := allUsers.Load(userID); ok && user != nil {
		return value.(*userInfo).lInfo
	}
	return nil
}

// GetUsers gets
func GetUsers() ([]string, []string, error) {

	var users []string
	var states []string
	allUsers.Range(func(key, value interface{}) bool {
		id := key.(string)
		us := value.(*userInfo)
		users = append(users, id)
		state, err := us.GetUserState(user.userid)
		if err != nil {
			return true
		}

		if int(state)+1 > len(stateList) {
			return true
		}
		states = append(states, stateList[int(state)])

		return true
	})
	return users, states, nil
}

// IsOnline judges
func IsOnline(userID string) bool {
	if user, ok := allUsers.Load(userID); ok && user != nil {
		state, err := user.GetUserState(userID)
		if err != nil {
			return false
		}
		return state == bothStarted
	}

	return false
}

// KillUser kills
func KillUser(userID string) error {
	if user, ok := allUsers.Load(userID); ok && user != nil {
		user.groupInfo = nil
		user.lfsInfo = nil
		//用于通知资源释放
		user.cancelFunc()
		allUsers.Delete(userID)
		return nil
	}

	return ErrUserNotExist
}

// GetUser gets userInfo
func GetUser(userID string) Service {
	if user, ok := allUsers.Load(userID); ok && user != nil {
		return user
	}
	return nil
}

// PersistBeforeExit is
func PersistBeforeExit() error {
	if allUsers == nil || allUsers.userBook == nil {
		return ErrUserBookIsNil
	}
	allUsers.Range(func(key, value interface{}) bool {
		uInfo := value.(*userInfo)
		state, err := uInfo.GetUserState(uInfo.userID)
		if err != nil {
			return err
		}
		if state != bothStarted {
			return true
		}
		err = userInfo.lfsInfo.Fsync(false)
		if err != nil {
			log.Printf("Sorry, something wrong in persisting for %s: %v\n", userid, err)
		} else {
			log.Printf("User %s Persist completed\n", userid)
		}
		userInfo.cancelFunc() //释放资源
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
