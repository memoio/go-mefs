package user

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"path"
	"sync"

	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/core"
	"github.com/memoio/go-mefs/repo/fsrepo"
	config "github.com/memoio/go-mefs/config"
	ad "github.com/memoio/go-mefs/utils/address"
)

type UserState int32

const (
	Starting UserState = iota
	Collecting
	CollectCompleted
	GroupStarted
	BothStarted
)

var StateList = []string{
	"starting", "Collecting", "CollectComplited", "GroupStarted", "BothStarted",
}

type UsersInfo struct {
	UserBook     map[string]*UserService
	sync.RWMutex //防止映射并发写入的问题
	Count        int
}

var allUsers *UsersInfo

type UserService struct {
	UserID          string
	localNode       *core.MefsNode
	GroupService    *GroupService
	LfsService      *LfsService
	ContractService *ContractService
	Context         context.Context
	CancelFunc      context.CancelFunc
	state           UserState
}

func (us *UserService) StartUserService(ctx context.Context, node *core.MefsNode, isInit bool, pwd string, capacity int64, duration int64, price int64, ks int, ps int) error {
	err := SetUserState(us.UserID, Starting)
	if err != nil {
		return err
	}
	us.localNode = node
	// 读keystore下uid文件
	keypath, err := config.Path("", path.Join("keystore", us.UserID))
	if err != nil {
		return ErrDirNotExist
	}
	_, err = os.Stat(keypath)
	if os.IsNotExist(err) {
		return ErrDirNotExist
	}
	userkey, err := fsrepo.GetPrivateKeyFromKeystore(us.UserID, keypath, pwd)
	if err != nil {
		return ErrGetSecreteKey
	}
	gp := ConstructGroupService(us.UserID, userkey.PrivateKey, node, duration, capacity, price, ks, ps)
	if !isInit {
		//在这里先尝试获取一次Bls config，如果失败，在启动完Groupservice的时候会再试一次
		err = gp.loadBLS12ConfigMeta()
		if err != nil {
			log.Println("Load BLS12 Config error:", err)
		}
	}
	err = SetGroupService(gp)
	if err != nil {
		fmt.Println("SetGroupService()err")
		return err
	}
	// user联网
	err = gp.StartGroupService(ctx, pwd, isInit)
	if err != nil {
		fmt.Println("StartGroupService(()err")
		return err
	}

	cs := ConstructContractService(us.UserID)
	err = SetContractService(us.UserID, cs)
	if err != nil {
		fmt.Println("SetContractService()err")
		return err
	}
	err = cs.SaveContracts()
	if err != nil {
		fmt.Println("SaveContracts err:", err)
	}

	lfs := ConstructLfsService(us.UserID, userkey.PrivateKey)

	err = SetLfsService(lfs)
	if err != nil {
		fmt.Println("SetLfsService()err")
		return err
	}
	err = lfs.StartLfsService(ctx, node)
	if err != nil {
		fmt.Println("StartLfsService()err")
		return err
	}

	return nil
}

func InitUserBook() {
	allUsers = &UsersInfo{
		UserBook: make(map[string]*UserService),
	}
}

// TODO:增加userbook的count判断，若大于上限则删除一些记录
func AddUserBook(us *UserService) error {
	if us == nil {
		return nil
	}
	// 初始化流程放init里
	if allUsers == nil || allUsers.UserBook == nil {
		return ErrUserBookIsNil
	}
	allUsers.Lock()
	defer allUsers.Unlock()
	if _, ok := allUsers.UserBook[us.UserID]; !ok {
		allUsers.UserBook[us.UserID] = us
		allUsers.Count++
	}
	return nil
}

func SetUserState(userID string, state UserState) error {
	if allUsers == nil || allUsers.UserBook == nil {
		return ErrUserBookIsNil
	}
	allUsers.RLock()
	defer allUsers.RUnlock()
	if us, ok := allUsers.UserBook[userID]; ok && us != nil {
		us.state = state
	}
	return nil
}

func GetUserServiceState(userID string) (UserState, error) {
	if allUsers == nil || allUsers.UserBook == nil {
		return 0, ErrUserBookIsNil
	}
	allUsers.RLock()
	defer allUsers.RUnlock()
	if us, ok := allUsers.UserBook[userID]; ok && us != nil {
		return us.state, nil
	}
	return 0, ErrUserNotExist
}

func GetUsers() ([]string, []string, error) {
	if allUsers == nil || allUsers.UserBook == nil {
		return nil, nil, ErrUserBookIsNil
	}
	allUsers.RLock()
	defer allUsers.RUnlock()
	var users []string
	var states []string
	for id, user := range allUsers.UserBook {
		users = append(users, id)
		state, err := GetUserServiceState(user.UserID)
		if err != nil {
			return nil, nil, err
		}
		if int(state)+1 > len(StateList) {
			return nil, nil, ErrWrongState
		}
		states = append(states, StateList[int(state)])
	}
	return users, states, nil
}

func KillUser(userID string) error {
	if allUsers == nil || allUsers.UserBook == nil {
		return ErrUserBookIsNil
	}
	allUsers.Lock()
	defer allUsers.Unlock()
	if user, ok := allUsers.UserBook[userID]; ok {
		user.GroupService = nil
		user.LfsService = nil
		user.ContractService = nil
		//用于通知资源释放
		user.CancelFunc()
		delete(allUsers.UserBook, userID)
		return nil
	}

	return ErrUserNotExist
}

func SetLfsService(lfs *LfsService) error {
	if allUsers == nil || allUsers.UserBook == nil {
		return ErrUserBookIsNil
	}
	allUsers.RLock()
	defer allUsers.RUnlock()
	if us, ok := allUsers.UserBook[lfs.UserID]; ok && us != nil {
		if us.LfsService != nil {
			return ErrLfsServiceAlreadySet
		}
		us.LfsService = lfs
		return nil
	}
	return ErrCannotFindUserInUserBook
}

func GetLfsService(userID string) *LfsService {
	if allUsers == nil || allUsers.UserBook == nil {
		return nil
	}
	allUsers.RLock()
	defer allUsers.RUnlock()
	if us, ok := allUsers.UserBook[userID]; ok && us != nil {
		if us.LfsService != nil {
			return us.LfsService
		}
		return nil
	}
	return nil
}

// 把当前user加入代理的userbook中
func SetGroupService(group *GroupService) error {
	if allUsers == nil || allUsers.UserBook == nil {
		return ErrUserBookIsNil
	}
	allUsers.RLock()
	defer allUsers.RUnlock()
	if us, ok := allUsers.UserBook[group.Userid]; ok && us != nil {
		if us.GroupService != nil {
			return ErrGroupServiceAlreadySet
		}
		us.GroupService = group
		return nil
	}
	return ErrCannotFindUserInUserBook
}

//GetGroupService 根据uid获取该user的groupservice实例
func GetGroupService(userID string) *GroupService {
	if allUsers == nil || allUsers.UserBook == nil {
		return nil
	}
	allUsers.RLock()
	defer allUsers.RUnlock()
	if us, ok := allUsers.UserBook[userID]; ok && us != nil {
		if us.GroupService != nil {
			return us.GroupService
		}
		return nil
	}
	return nil
}

func SetContractService(userID string, contract *ContractService) error {
	if allUsers == nil || allUsers.UserBook == nil {
		return ErrUserBookIsNil
	}
	allUsers.RLock()
	defer allUsers.RUnlock()
	if us, ok := allUsers.UserBook[userID]; ok && us != nil {
		if us.ContractService != nil {
			return ErrContractServiceAlreadySet
		}
		us.ContractService = contract
		return nil
	}
	return ErrCannotFindUserInUserBook
}

func GetContractService(userID string) *ContractService {
	if allUsers == nil || allUsers.UserBook == nil {
		return nil
	}
	allUsers.RLock()
	defer allUsers.RUnlock()
	if us, ok := allUsers.UserBook[userID]; ok && us != nil {
		if us.ContractService != nil {
			return us.ContractService
		}
		return nil
	}
	return nil
}

func (gp *GroupService) GetKeyset() *mcl.KeySet {
	if gp == nil {
		return nil
	}
	return gp.KeySet
}

func ConstructUserService(uid string) *UserService {
	ctx, cancel := context.WithCancel(context.Background())
	return &UserService{
		UserID:     uid,
		Context:    ctx,
		CancelFunc: cancel,
	}
}

func PersistBeforeExit() error {
	if allUsers == nil || allUsers.UserBook == nil {
		return ErrUserBookIsNil
	}
	for UserID, UserService := range allUsers.UserBook {
		if UserService != nil {
			state, err := GetUserServiceState(UserService.UserID)
			if err != nil {
				return err
			}
			if state != BothStarted {
				continue
			}
			err = UserService.LfsService.Fsync(false)
			if err != nil {
				fmt.Printf("Sorry, something wrong in persisting for %s: %v\n", UserID, err)
			} else {
				fmt.Printf("User %s Persist completed\n", UserID)
			}
			UserService.CancelFunc() //释放资源
		}
	}
	return nil
}

//输出本节点的信息
func ShowInfo(userID string) map[string]string {
	outmap := map[string]string{}
	fmt.Println(">>>>>>>>>>>>>>ShowInfo>>>>>>>>>>>>>>")
	defer fmt.Println("================================")
	gp := GetGroupService(userID)
	lfs := GetLfsService(userID)
	if lfs == nil {
		outmap["error"] = "lfsService==nil"
		return outmap
	}
	if gp == nil {
		outmap["error"] = "groupService==nil"
		return outmap
	}

	//查keeper
	unconKeepers, conKeepers, err := gp.GetKeepers(-1)
	if err != nil {
		outmap["error"] = "GetKeepers(-1)err:" + err.Error()
		return outmap
	}
	outmap["unconKeepers:"] = ""
	outmap["conKeepers:"] = ""
	for _, keeper := range unconKeepers {
		outmap["unconKeepers:"] += "," + keeper
	}
	for _, keeper := range conKeepers {
		outmap["conKeepers:"] += "," + keeper
	}

	//查provider

	unconProviders, conProviders, err := gp.GetLocalProviders()
	if err != nil {
		outmap["error"] = "GetLocalProviders()err:" + err.Error()
		return outmap
	}
	outmap["conProviders:"] = ""
	outmap["unconProviders:"] = ""
	for _, provider := range unconProviders {
		outmap["unconProviders:"] += "," + provider
	}
	for _, provider := range conProviders {
		outmap["conProviders:"] += "," + provider
	}

	//查本节点余额
	addrLocal, err := ad.GetAddressFromID(userID)
	if err != nil {
		outmap["error"] = "GetAddressFromID() err:" + err.Error()
		return outmap
	}
	cfg, err := gp.localNode.Repo.Config()
	if err != nil {
		outmap["error"] = "Config() err:" + err.Error()
		return outmap
	}
	amountLocal, err := contracts.QueryBalance(cfg.Eth, addrLocal.Hex())
	if err != nil {
		outmap["error"] = "QueryBalance() err:" + err.Error()
	}
	outmap["localBalance:"] = amountLocal.String()

	//查upkeeping合约信息
	cs := GetContractService(userID)
	if cs == nil {
		outmap["error"] = "ContractService==nil"
		return outmap
	}
	//计算当前合约的花费(合约总金额-当前余额)
	upkeeping := cs.GetUpkeepingItem()
	outmap["upkeeping.UpKeepingAddr:"] = upkeeping.UpKeepingAddr
	amountUpkeeping, err := contracts.QueryBalance(cfg.Eth, upkeeping.UpKeepingAddr)
	if err != nil {
		outmap["error"] = "QueryBalance()err:" + err.Error()
		return outmap
	}
	d := gp.storeDays
	s := gp.storeSize
	price := gp.storePrice
	var moneyPerDay = new(big.Int)
	var moneyAccount = new(big.Int)
	moneyPerDay = moneyPerDay.Mul(big.NewInt(price), big.NewInt(s))
	moneyAccount = moneyAccount.Mul(moneyPerDay, big.NewInt(d))
	fmt.Printf("%v", upkeeping)
	outmap["upkeeping cost:"] = big.NewInt(0).Sub(moneyAccount, amountUpkeeping).String()
	outmap["upkeeping.Duration:"] = big.NewInt(upkeeping.Duration).String()
	outmap["upkeeping.Capacity"] = big.NewInt(upkeeping.Capacity).String()

	//查询当前使用的存储空间
	buckets, err := lfs.ListBucket("")
	if err != nil {
		outmap["error"] = "ListBucket() err:" + err.Error()
		return outmap
	}
	var StorageSize int
	for _, bucket := range buckets {
		storageSpace, err := lfs.ShowStorageSpace(bucket.BucketName, "")
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
