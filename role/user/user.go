package user

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"path"
	"sync"

	config "github.com/memoio/go-mefs/config"
	"github.com/memoio/go-mefs/core"
	"github.com/memoio/go-mefs/repo/fsrepo"
	pb "github.com/memoio/go-mefs/role/user/pb"
)

var localNode *core.MefsNode

var allUsers sync.Map

// Service defines user's function
type Service interface {
	Start() error
	Stop()
	Fsync(bool) error
	Online() bool

	ListBuckets(prefix string) ([]*pb.BucketInfo, error)
	CreateBucket(bucketName string, options *pb.BucketOptions) (*pb.BucketInfo, error)
	HeadBucket(bucketName string) (*pb.BucketInfo, error)
	DeleteBucket(bucketName string) (*pb.BucketInfo, error)

	ListObjects(bucketName, prefix string, opts ObjectOptions) ([]*pb.ObjectInfo, error)

	PutObject(bucketName, objectName string, reader io.Reader) (*pb.ObjectInfo, error)

	GetObject(bucketName, objectName string, writer io.Writer, completeFuncs []CompleteFunc, opts *DownloadOptions) error
	HeadObject(bucketName, objectName string, opts ObjectOptions) (*pb.ObjectInfo, error)
	DeleteObject(bucketName, objectName string) (*pb.ObjectInfo, error)

	ShowStorage() (uint64, error)
	ShowBucketStorage(bucketName string) (uint64, error)
}

// NewUser add a new user
func NewUser(uid string, isInit bool, pwd string, capacity int64, duration int64, price int64, ks int, ps int, rdo bool) (Service, error) {
	if user, ok := allUsers.Load(uid); ok && user != nil {
		return user.(*LfsInfo), errors.New("user is running")
	}

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
		reDeploy:    rdo,
	}

	ctx, cancel := context.WithCancel(context.Background())

	lInfo := &LfsInfo{
		userID:     uid,
		context:    ctx,
		cancelFunc: cancel,
		privateKey: userkey.PrivateKey,
		gInfo:      ginfo,
	}

	allUsers.Store(uid, lInfo)

	return lInfo, nil
}

// IsOnline judges
func IsOnline(userID string) bool {
	if user, ok := allUsers.Load(userID); ok && user != nil {
		return user.(*LfsInfo).online
	}

	return false
}

func getSk(userID string) []byte {
	if user, ok := allUsers.Load(userID); ok && user != nil {
		return user.(*LfsInfo).privateKey
	}
	return nil
}

func getGroup(userID string) *groupInfo {
	if user, ok := allUsers.Load(userID); ok && user != nil {
		return user.(*LfsInfo).gInfo
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
func GetUser(userID string) *LfsInfo {
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

		states = append(states, us.online)

		return true
	})
	return users, states, nil
}

// PersistBeforeExit is
func PersistBeforeExit() error {
	allUsers.Range(func(key, value interface{}) bool {
		uInfo := value.(*LfsInfo)
		if !uInfo.online {
			return true
		}
		err := uInfo.Fsync(false)
		if err != nil {
			log.Printf("Sorry, something wrong in persisting for %s: %v\n", uInfo.userID, err)
		} else {
			log.Printf("User %s Persist completed\n", uInfo.userID)
		}
		uInfo.cancelFunc() //释放资源
		return true
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
