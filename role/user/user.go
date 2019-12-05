package user

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"sync"

	mcl "github.com/memoio/go-mefs/bls12"
	config "github.com/memoio/go-mefs/config"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/core"
	"github.com/memoio/go-mefs/repo/fsrepo"
	ad "github.com/memoio/go-mefs/utils/address"
)

type userState int32

var localNode *core.MefsNode

const (
	starting userState = iota
	collecting
	collectCompleted
	onDeploy
	groupStarted
	bothStarted
)

var stateList = []string{
	"starting", "collecting", "CollectComplited", "onDeploy", "groupStarted", "bothStarted",
}

type usersInfo struct {
	userBook     map[string]*userService
	sync.RWMutex //防止映射并发写入的问题
	count        int
}

var allUsers *usersInfo

type userService struct {
	userid       string
	groupService *groupService
	LfsService   *LfsService
	context      context.Context
	cancelFunc   context.CancelFunc
	state        userState
}

// StartUser starts user
func StartUser(uid string, isInit bool, pwd string, capacity int64, duration int64, price int64, ks int, ps int, rdo bool) error {
	// 证明该user已经启动
	if st, err := getUserState(uid); err == nil && st >= starting {
		return errors.New("The user is running")
	}

	us := constructUserService(uid)
	err := addUserBook(us)
	if err != nil {
		return err
	}
	err = setUserState(us.userid, starting)
	if err != nil {
		return err
	}

	// 读keystore下uid文件
	keypath, err := config.Path("", path.Join("keystore", us.userid))
	if err != nil {
		return ErrDirNotExist
	}
	_, err = os.Stat(keypath)
	if os.IsNotExist(err) {
		return ErrDirNotExist
	}
	userkey, err := fsrepo.GetPrivateKeyFromKeystore(us.userid, keypath, pwd)
	if err != nil {
		return err
	}
	gp := constructGroupService(us.userid, userkey.PrivateKey, duration, capacity, price, ks, ps, rdo)
	err = setGroupService(gp)
	if err != nil {
		log.Println("setGroupService()err")
		return err
	}
	// user联网
	err = gp.startGroupService(us.context, isInit)
	if err != nil {
		log.Println("startGroupService(()err")
		return err
	}

	lfs := constructLfsService(us.userid, userkey.PrivateKey)

	err = setLfsService(lfs)
	if err != nil {
		log.Println("setLfsService()err")
		return err
	}
	err = lfs.StartLfsService(us.context)
	if err != nil {
		log.Println("StartLfsService()err")
		return err
	}

	return nil
}

// GetUsers gets
func GetUsers() ([]string, []string, error) {
	if allUsers == nil || allUsers.userBook == nil {
		return nil, nil, ErrUserBookIsNil
	}
	allUsers.RLock()
	defer allUsers.RUnlock()
	var users []string
	var states []string
	for id, user := range allUsers.userBook {
		users = append(users, id)
		state, err := getUserState(user.userid)
		if err != nil {
			return nil, nil, err
		}
		if int(state)+1 > len(stateList) {
			return nil, nil, ErrWrongState
		}
		states = append(states, stateList[int(state)])
	}
	return users, states, nil
}

// KillUser kills
func KillUser(userID string) error {
	if allUsers == nil || allUsers.userBook == nil {
		return ErrUserBookIsNil
	}
	allUsers.Lock()
	defer allUsers.Unlock()
	if user, ok := allUsers.userBook[userID]; ok {
		user.groupService = nil
		user.LfsService = nil
		//用于通知资源释放
		user.cancelFunc()
		delete(allUsers.userBook, userID)
		return nil
	}

	return ErrUserNotExist
}

// IsUserOnline judges
func IsUserOnline(userID string) bool {
	state, err := getUserState(userID)
	if err != nil {
		return false
	}
	return state == bothStarted
}

// InitUserBook inits
func InitUserBook(node *core.MefsNode) {
	localNode = node
	allUsers = &usersInfo{
		userBook: make(map[string]*userService),
	}
}

// addUserBook adds
// TODO:增加userbook的count判断，若大于上限则删除一些记录
func addUserBook(us *userService) error {
	if us == nil {
		return nil
	}
	// 初始化流程放init里
	if allUsers == nil || allUsers.userBook == nil {
		return ErrUserBookIsNil
	}
	allUsers.Lock()
	defer allUsers.Unlock()
	if _, ok := allUsers.userBook[us.userid]; !ok {
		allUsers.userBook[us.userid] = us
		allUsers.count++
	}
	return nil
}

func setUserState(userID string, state userState) error {
	if allUsers == nil || allUsers.userBook == nil {
		return ErrUserBookIsNil
	}
	allUsers.RLock()
	defer allUsers.RUnlock()
	if us, ok := allUsers.userBook[userID]; ok && us != nil {
		us.state = state
	}
	return nil
}

func getUserState(userID string) (userState, error) {
	if allUsers == nil || allUsers.userBook == nil {
		return 0, ErrUserBookIsNil
	}
	allUsers.RLock()
	defer allUsers.RUnlock()
	if us, ok := allUsers.userBook[userID]; ok && us != nil {
		return us.state, nil
	}
	return 0, ErrUserNotExist
}

func setLfsService(lfs *LfsService) error {
	if allUsers == nil || allUsers.userBook == nil {
		return ErrUserBookIsNil
	}
	allUsers.RLock()
	defer allUsers.RUnlock()
	if us, ok := allUsers.userBook[lfs.userid]; ok && us != nil {
		if us.LfsService != nil {
			return ErrLfsServiceAlreadySet
		}
		us.LfsService = lfs
		return nil
	}
	return ErrCannotFindUserInUserBook
}

// GetLfsService gets
func GetLfsService(userID string) *LfsService {
	if allUsers == nil || allUsers.userBook == nil {
		return nil
	}
	allUsers.RLock()
	defer allUsers.RUnlock()
	if us, ok := allUsers.userBook[userID]; ok && us != nil {
		if us.LfsService != nil {
			return us.LfsService
		}
		return nil
	}
	return nil
}

// 把当前user加入代理的userbook中
func setGroupService(group *groupService) error {
	if allUsers == nil || allUsers.userBook == nil {
		return ErrUserBookIsNil
	}
	allUsers.RLock()
	defer allUsers.RUnlock()
	if us, ok := allUsers.userBook[group.userid]; ok && us != nil {
		if us.groupService != nil {
			return ErrGroupServiceAlreadySet
		}
		us.groupService = group
		return nil
	}
	return ErrCannotFindUserInUserBook
}

//getGroupService 根据uid获取该user的groupservice实例
func getGroupService(userID string) *groupService {
	if allUsers == nil || allUsers.userBook == nil {
		return nil
	}
	allUsers.RLock()
	defer allUsers.RUnlock()
	if us, ok := allUsers.userBook[userID]; ok && us != nil {
		if us.groupService != nil {
			return us.groupService
		}
		return nil
	}
	return nil
}

func (gp *groupService) getKeyset() *mcl.KeySet {
	if gp == nil {
		return nil
	}
	return gp.keySet
}

// constructUserService new
func constructUserService(uid string) *userService {
	ctx, cancel := context.WithCancel(context.Background())
	return &userService{
		userid:     uid,
		context:    ctx,
		cancelFunc: cancel,
	}
}

// PersistBeforeExit is
func PersistBeforeExit() error {
	if allUsers == nil || allUsers.userBook == nil {
		return ErrUserBookIsNil
	}
	for userid, userService := range allUsers.userBook {
		if userService != nil {
			state, err := getUserState(userService.userid)
			if err != nil {
				return err
			}
			if state != bothStarted {
				continue
			}
			err = userService.LfsService.Fsync(false)
			if err != nil {
				log.Printf("Sorry, something wrong in persisting for %s: %v\n", userid, err)
			} else {
				log.Printf("User %s Persist completed\n", userid)
			}
			userService.cancelFunc() //释放资源
		}
	}
	return nil
}

//ShowInfo 输出本节点的信息
func ShowInfo(userID string) map[string]string {
	outmap := map[string]string{}
	log.Println(">>>>>>>>>>>>>>ShowInfo>>>>>>>>>>>>>>")
	defer log.Println("================================")
	gp := getGroupService(userID)
	if gp == nil {
		outmap["error: "] = "groupService==nil"
		return outmap
	}

	lfs := GetLfsService(userID)
	if lfs == nil {
		outmap["error: "] = "lfsService==nil"
		return outmap
	}

	//查keeper
	unconKeepers, conKeepers, err := gp.getKeepers(-1)
	if err != nil {
		outmap["error: "] = "GetKeepers(-1)err:" + err.Error()
		return outmap
	}
	outmap["unconKeepers: "] = ""
	outmap["conKeepers: "] = ""
	for _, keeper := range unconKeepers {
		outmap["unconKeepers: "] += "," + keeper
	}
	for _, keeper := range conKeepers {
		outmap["conKeepers: "] += "," + keeper
	}

	//查provider

	unconProviders, conProviders, err := gp.getLocalProviders()
	if err != nil {
		outmap["error: "] = "GetLocalProviders()err:" + err.Error()
		return outmap
	}
	outmap["conProviders: "] = ""
	outmap["unconProviders: "] = ""
	for _, provider := range unconProviders {
		outmap["unconProviders: "] += "," + provider
	}
	for _, provider := range conProviders {
		outmap["conProviders: "] += "," + provider
	}

	//查本节点余额
	addrLocal, err := ad.GetAddressFromID(userID)
	if err != nil {
		outmap["error: "] = "GetAddressFromID() err:" + err.Error()
		return outmap
	}
	amountLocal, err := contracts.QueryBalance(addrLocal.Hex())
	if err != nil {
		outmap["error: "] = "QueryBalance() err:" + err.Error()
	}
	outmap["user balance: "] = amountLocal.String()

	/*
		//查upkeeping合约信息
		cs := getContractService(userID)
		if cs == nil {
			outmap["error: "] = "contractService==nil"
			return outmap
		}
		//计算当前合约的花费(合约总金额-当前余额)
		upkeeping, err := cs.getUpkeepingItem()
		if err != nil {
			outmap["error: "] = "getUpkeepingItem() err:" + err.Error()
			return outmap
		}
		outmap["upkeeping.UpKeepingAddr:"] = upkeeping.UpKeepingAddr
		amountUpkeeping, err := contracts.QueryBalance(upkeeping.UpKeepingAddr)
		if err != nil {
			outmap["error: "] = "QueryBalance()err:" + err.Error()
			return outmap
		}
		outmap["ukeeping balance: "] = amountUpkeeping.String()

		d := upkeeping.Duration
		s := upkeeping.Capacity
		price := upkeeping.Price
		var moneyPerDay = new(big.Int)
		var moneyAccount = new(big.Int)
		moneyPerDay = moneyPerDay.Mul(big.NewInt(price), big.NewInt(s))
		moneyAccount = moneyAccount.Mul(moneyPerDay, big.NewInt(d))
		outmap["upkeeping cost:"] = big.NewInt(0).Sub(moneyAccount, amountUpkeeping).String()
		outmap["upkeeping.Duration:"] = big.NewInt(upkeeping.Duration).String()
		outmap["upkeeping.Capacity:"] = big.NewInt(upkeeping.Capacity).String()
	*/
	//查询当前使用的存储空间
	superBucket, err := lfs.ListBucket("")
	if err != nil {
		outmap["error"] = "ListBucket() err:" + err.Error()
		return outmap
	}
	var StorageSize int
	for _, bucket := range superBucket {
		storageSpace, err := lfs.ShowStorageSpace(bucket.Name, "")
		if err != nil {
			outmap["error"] = "ShowStorageSpace() err" + err.Error()
			return outmap
		}
		StorageSize += storageSpace
	}
	FloatStorage := float64(StorageSize)
	var outstorage string
	if FloatStorage < 1024 && FloatStorage >= 0 {
		outstorage = fmt.Sprintf("%.2f", FloatStorage) + "B"
	} else if FloatStorage < 1048576 && FloatStorage >= 1024 {
		outstorage = fmt.Sprintf("%.2f", FloatStorage/1024) + "KB"
	} else if FloatStorage < 1073741824 && FloatStorage >= 1048576 {
		outstorage = fmt.Sprintf("%.2f", FloatStorage/1048576) + "MB"
	} else {
		outstorage = fmt.Sprintf("%.2f", FloatStorage/1073741824) + "GB"
	}
	outmap["storage"] = outstorage
	return outmap

}
