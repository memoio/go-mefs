package role

import (
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gogo/protobuf/proto"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/contracts/channel"
	"github.com/memoio/go-mefs/contracts/market"
	"github.com/memoio/go-mefs/contracts/upKeeping"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"golang.org/x/crypto/sha3"
)

// ProviderItem has provider's info
type ProviderItem struct {
	ProviderID string   // providerid
	Capacity   int64    // MB
	Money      *big.Int // pledge money
	StartTime  int64    // start time
}

// KeeperItem has provider's info
type KeeperItem struct {
	KeeperID  string   // providerid
	Money     *big.Int // pledge money
	StartTime int64    // start time; not in contract
}

// UpKeepingItem has upkeeping information
type UpKeepingItem struct {
	UserID      string // 部署upkeeping的userid
	QueryID     string // 部署upkeeping的queryID
	UpKeepingID string // 合约地址
	Keepers     []upKeeping.UpKeepingKPInfo
	Providers   []upKeeping.UpKeepingKPInfo
	KeeperSLA   int32
	ProviderSLA int32
	Duration    int64 //存储时间，单位s(部署合约时的单位是天，获得的参数单位是s)
	Capacity    int64
	Price       int64 // 部署的价格
	StartTime   int64 // 部署的时间
	Money       *big.Int
	EndDate     int64
	Cycle       int64
	NeedPay     int64
	Proofs      []upKeeping.UpKeepingProof
}

// OfferItem has offer information
type OfferItem struct {
	ProviderID string // 部署Offer的providerid
	OfferID    string // offer address : id format
	Capacity   int64
	Duration   int64
	Price      int64 // 合约给出的单价
	CreateDate int64 //合约创建时间
}

// ChannelItem has channel information
type ChannelItem struct {
	UserID    string
	ProID     string
	ChannelID string
	QueryID   string   //belongs to which
	Money     *big.Int // channel has balance
	Value     *big.Int
	Sig       []byte // pb.Channelsignature(channel addr, value)
	StartTime int64  // 部署的时间
	Duration  int64  // timeout
	Dirty     bool   //  value is change?
}

// QueryItem has query information
type QueryItem struct {
	UserID       string // 部署Query的userid
	QueryID      string
	Capacity     int64
	Duration     int64
	Price        int64 // 合约给出的单价
	KeeperNums   int32
	ProviderNums int32
	Completed    bool
}

// RootItem has root information
type RootItem struct {
	UserID  string // 部署Query的userid
	QueryID string
	Keys    []int64
}

type kpItem struct {
	keeperIDs []string
}

var kpMap sync.Map

// GetKeeperInfo returns keeper info
func GetKeeperInfo(localID, keeperID string) (KeeperItem, error) {
	var item KeeperItem
	localAddress, err := address.GetAddressFromID(localID)
	if err != nil {
		return item, err
	}

	keeperAddress, err := address.GetAddressFromID(keeperID)
	if err != nil {
		return item, err
	}

	keeperInstance, err := contracts.GetKeeperContractFromIndexer(localAddress)
	if err != nil {
		return item, nil
	}

	retryCount := 0
	for {
		retryCount++
		isKeeper, isBanned, money, _, err := keeperInstance.Info(&bind.CallOpts{From: localAddress}, keeperAddress)
		if err != nil {
			if retryCount > 10 {
				return item, nil
			}
			time.Sleep(30 * time.Second)
			continue
		}
		if isKeeper && !isBanned {
			keeperID, err := address.GetIDFromAddress(keeperAddress.String())
			if err != nil {
				return item, err
			}

			item = KeeperItem{
				KeeperID: keeperID,
				Money:    money,
			}
			return item, nil
		}
		break
	}

	return item, ErrNotKeeper
}

func IsKeeper(userID string) (bool, error) {
	localAddress, err := address.GetAddressFromID(userID)
	if err != nil {
		return false, err
	}
	return contracts.IsKeeper(localAddress)
}

// GetProviderInfo returns provider info
func GetProviderInfo(localID, proID string) (ProviderItem, error) {
	var item ProviderItem
	localAddress, err := address.GetAddressFromID(localID)
	if err != nil {
		return item, err
	}

	proAddress, err := address.GetAddressFromID(proID)
	if err != nil {
		return item, err
	}

	proInstance, err := contracts.GetProviderContractFromIndexer(localAddress)
	if err != nil {
		return item, nil
	}

	retryCount := 0
	for {
		retryCount++
		isProvider, isBanned, money, stime, err := proInstance.Info(&bind.CallOpts{From: localAddress}, proAddress)
		if err != nil {
			if retryCount > 10 {
				return item, nil
			}
			time.Sleep(30 * time.Second)
			continue
		}
		if isProvider && !isBanned {
			item = ProviderItem{
				ProviderID: proID,
				Money:      money,
				StartTime:  stime.Int64(),
				Capacity:   0,
			}
			return item, nil
		}
		break
	}

	return item, ErrNotProvider
}

func IsProvider(userID string) (bool, error) {
	localAddress, err := address.GetAddressFromID(userID)
	if err != nil {
		return false, err
	}
	return contracts.IsProvider(localAddress)
}

// DeployOffer is
func DeployOffer(localID, sk string, capacity, duration, price int64, redo bool) (offerID string, err error) {
	utils.MLogger.Info("Begin to deploy offer contract...")
	localAddress, err := address.GetAddressFromID(localID)
	if err != nil {
		return offerID, err
	}

	//获得用户的账户余额
	balance, err := contracts.QueryBalance(localAddress.Hex())
	if err == nil {
		utils.MLogger.Infof("%s (%s) has balance: %s", localID, localAddress.Hex(), balance)
	}

	offerAddr, err := contracts.DeployOffer(localAddress, sk, capacity, duration, price, redo)
	if err != nil {
		utils.MLogger.Error("Fail to deploy offer contract: ", err)
		return offerID, err
	}

	offerID, err = address.GetIDFromAddress(offerAddr.String())
	if err != nil {
		return offerID, err
	}
	utils.MLogger.Info("Finish deploy offer contract: ", offerID)

	return offerID, nil
}

//GetOfferInfo get provider's offer-info
func GetOfferInfo(localID, offerID string) (OfferItem, error) {
	var item OfferItem

	localAddress, err := address.GetAddressFromID(localID)
	if err != nil {
		return item, nil
	}

	offerAddress, err := address.GetAddressFromID(offerID)
	if err != nil {
		return item, nil
	}

	offerInstance, err := market.NewOffer(offerAddress, contracts.GetClient(contracts.EndPoint))
	if err != nil {
		return item, err
	}

	retryCount := 0
	for {
		retryCount++
		capacity, duration, price, createDate, err := offerInstance.Get(&bind.CallOpts{
			From: localAddress,
		})
		if err != nil {
			if retryCount > 10 {
				return item, err
			}
			time.Sleep(30 * time.Second)
			continue
		}

		item = OfferItem{
			Capacity:   capacity.Int64(),
			Duration:   duration.Int64(),
			Price:      price.Int64(),
			OfferID:    offerID,
			CreateDate: createDate.Int64(),
		}
		break
	}

	return item, nil
}

// GetLatestOffer gets
func GetLatestOffer(localID, proID string) (OfferItem, error) {
	var item OfferItem
	userAddr, err := address.GetAddressFromID(localID)
	if err != nil {
		return item, err
	}

	proAddr, err := address.GetAddressFromID(proID)
	if err != nil {
		return item, err
	}

	oAddrs, err := contracts.GetOfferAddrs(userAddr, proAddr)
	if err != nil {
		utils.MLogger.Info("get ", proID, " 's offer address err: ", err)
		return item, err
	}

	offerAddr := oAddrs[len(oAddrs)-1]
	offerID, err := address.GetIDFromAddress(offerAddr.String())
	if err != nil {
		return item, err
	}

	item, err = GetOfferInfo(localID, offerID)
	if err != nil {
		return item, err
	}
	return item, nil
}

// DeployQuery is
func DeployQuery(userID, sk string, storeDays, storeSize, storePrice int64, ks, ps int, rdo bool) (queryID string, err error) {
	utils.MLogger.Info("Begin to deploy query contract...")
	uaddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return queryID, err
	}

	// getbalance
	balance, err := contracts.QueryBalance(uaddr.String())
	if err == nil {
		utils.MLogger.Infof("%s (%s) has balance: %s", userID, uaddr.String(), balance)
	}

	//balance >? query + upKeeping + channel cost
	var moneyPerDay = new(big.Int)
	moneyPerDay = moneyPerDay.Mul(big.NewInt(storePrice), big.NewInt(storeSize))
	var moneyAccount = new(big.Int)
	moneyAccount = moneyAccount.Mul(moneyPerDay, big.NewInt(storeDays))

	deployPrice := big.NewInt(int64(740621000000000))
	deployPrice.Add(big.NewInt(1128277), big.NewInt(int64(652346*ps)))
	var leastMoney = new(big.Int)
	leastMoney = leastMoney.Add(moneyAccount, deployPrice)
	if balance.Cmp(leastMoney) < 0 { //余额不足
		utils.MLogger.Info(uaddr.String(), " need more balance to start")
		return queryID, ErrNotEnoughBalance
	}

	// deploy query
	queryAddr, err := contracts.DeployQuery(uaddr, sk, storeSize, storeDays, storePrice, ks, ps, rdo)
	if err != nil {
		utils.MLogger.Error("fail to deploy query contract: ", err)
		return queryID, err
	}

	utils.MLogger.Info(uaddr.String(), " has new query: ", queryAddr.String())

	queryID, err = address.GetIDFromAddress(queryAddr.String())
	if err != nil {
		return queryID, err
	}

	utils.MLogger.Info("Finish deploy query contract: ", queryID)

	return queryID, nil
}

// GetLatestQuery gets
func GetLatestQuery(userID string) (QueryItem, error) {
	var item QueryItem
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return item, err
	}

	qAddrs, err := contracts.GetQueryAddrs(userAddr, userAddr)
	if err != nil {
		return item, err
	}

	queryID, err := address.GetIDFromAddress(qAddrs[len(qAddrs)-1].String())
	if err != nil {
		return item, err
	}

	item, err = GetQueryInfo(userID, queryID)
	if err != nil {
		return item, err
	}

	return item, nil
}

//GetQueryInfo get user's query-info
// 分别返回申请的容量、持久化时间、价格、keeper数量、provider数量、是否成功放进upkeeping中
func GetQueryInfo(localID, queryID string) (QueryItem, error) {
	var item QueryItem
	localAddr, err := address.GetAddressFromID(localID)
	if err != nil {
		return item, err
	}

	queryAddr, err := address.GetAddressFromID(queryID)
	if err != nil {
		return item, err
	}

	queryInstance, err := market.NewQuery(queryAddr, contracts.GetClient(contracts.EndPoint))
	if err != nil {
		return item, err
	}
	retryCount := 0
	for {
		retryCount++
		capacity, duration, price, ks, ps, completed, err := queryInstance.Get(&bind.CallOpts{
			From: localAddr,
		})
		if err != nil {
			if retryCount > 10 {
				return item, err
			}
			time.Sleep(60 * time.Second)
			continue
		}
		item = QueryItem{
			Capacity:     capacity.Int64(),
			Duration:     duration.Int64(),
			Price:        price.Int64(),
			KeeperNums:   int32(ks.Int64()),
			ProviderNums: int32(ps.Int64()),
			Completed:    completed,
			QueryID:      queryID,
		}
		return item, nil
	}
}

// DeployUpKeeping is
func DeployUpKeeping(userID, queryID, hexSk string, ks, ps []string, storeDays, storeSize, storePrice, stPayCycle int64, redo bool) (ukID string, err error) {
	localAddress, err := address.GetAddressFromID(userID)
	if err != nil {
		return ukID, err
	}

	queryAddress, err := address.GetAddressFromID(queryID)
	if err != nil {
		return ukID, err
	}

	var keepers, providers []common.Address
	for _, keeper := range ks {
		keeperAddress, err := address.GetAddressFromID(keeper)
		if err != nil {
			return ukID, err
		}
		keepers = append(keepers, keeperAddress)
	}

	for _, provider := range ps {
		providerAddress, err := address.GetAddressFromID(provider)
		if err != nil {
			return ukID, err
		}
		providers = append(providers, providerAddress)
	}

	var moneyPerDay = new(big.Int)
	var moneyAccount = new(big.Int)
	moneyPerDay = moneyPerDay.Mul(big.NewInt(storePrice), big.NewInt(storeSize))
	moneyAccount = moneyAccount.Mul(moneyPerDay, big.NewInt(storeDays))

	utils.MLogger.Info("Begin to deploy upkeeping contract...")

	ukAddr, err := contracts.DeployUpkeeping(hexSk, localAddress, queryAddress, keepers, providers, storeDays, storeSize, storePrice, stPayCycle, moneyAccount, redo)
	if err != nil {
		utils.MLogger.Error("Deploy upkeeping contract failed: ", err)
		return ukID, err
	}

	err = contracts.SetQueryCompleted(hexSk, queryAddress)
	if err != nil {
		return ukID, err
	}

	ukID, err = address.GetIDFromAddress(ukAddr.String())
	if err != nil {
		return ukID, err
	}

	utils.MLogger.Info("Finish deploy upkeeping contract: ", ukID)

	return ukID, nil
}

// GetUpkeepingInfo get Upkeeping-contract's params
func GetUpkeepingInfo(localID, ukID string) (UpKeepingItem, error) {
	var item UpKeepingItem
	localAddr, err := address.GetAddressFromID(localID)
	if err != nil {
		return item, err
	}

	ukAddr, err := address.GetAddressFromID(ukID)
	if err != nil {
		return item, err
	}

	ukInstance, err := upKeeping.NewUpKeeping(ukAddr, contracts.GetClient(contracts.EndPoint))
	if err != nil {
		return item, nil
	}

	retryCount := 0
	for {
		retryCount++
		queryAddr, keepers, providers, duration, capacity, price, startTime, endDate, cycle, needPay, proofs, err := ukInstance.GetOrder(&bind.CallOpts{
			From: localAddr,
		})
		if err != nil {
			if retryCount > 10 {
				return item, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		// var keepers []string
		// var providers []string
		// for _, keeper := range keeperAddrs {
		// 	kid, err := address.GetIDFromAddress(keeper.String())
		// 	if err != nil {
		// 		return item, err
		// 	}
		// 	keepers = append(keepers, kid)
		// }

		// for _, provider := range providerAddrs {
		// 	pid, err := address.GetIDFromAddress(provider.String())
		// 	if err != nil {
		// 		return item, err
		// 	}
		// 	providers = append(providers, pid)
		// }

		qid, err := address.GetIDFromAddress(queryAddr.String())
		if err != nil {
			return item, err
		}

		item = UpKeepingItem{
			QueryID:     qid,
			UpKeepingID: ukID,
			Keepers:     keepers,
			KeeperSLA:   int32(len(keepers)),
			Providers:   providers,
			ProviderSLA: int32(len(providers)),
			Duration:    duration.Int64(),
			Capacity:    capacity.Int64(),
			Price:       price.Int64(),
			StartTime:   startTime.Int64(),
			EndDate:     endDate.Int64(),
			Cycle:       cycle.Int64(),
			NeedPay:     needPay.Int64(),
			Proofs:      proofs,
		}
		return item, nil
	}
}

// GetUpKeeping gets
func GetUpKeeping(userID, queryID string) (UpKeepingItem, error) {
	var item UpKeepingItem
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return item, err
	}

	queryAddr, err := address.GetAddressFromID(queryID)
	if err != nil {
		return item, err
	}

	ukAddr, ukInstance, err := contracts.GetUpkeeping(userAddr, userAddr, queryAddr.String())
	if err != nil {
		return item, err
	}
	retryCount := 0
	for {
		retryCount++
		queryAddr, keepers, providers, duration, capacity, price, startTime, endDate, cycle, needPay, proofs, err := ukInstance.GetOrder(&bind.CallOpts{
			From: userAddr,
		})
		if err != nil {
			if retryCount > 10 {
				return item, err
			}
			time.Sleep(30 * time.Second)
			continue
		}
		// var keepers []string
		// var providers []string
		// for _, keeper := range keeperAddrs {
		// 	kid, err := address.GetIDFromAddress(keeper.String())
		// 	if err != nil {
		// 		return item, err
		// 	}
		// 	keepers = append(keepers, kid)
		// }

		// for _, provider := range providerAddrs {
		// 	pid, err := address.GetIDFromAddress(provider.String())
		// 	if err != nil {
		// 		return item, err
		// 	}
		// 	providers = append(providers, pid)
		// }

		qid, err := address.GetIDFromAddress(queryAddr.String())
		if err != nil {
			return item, err
		}

		ukID, err := address.GetIDFromAddress(ukAddr.String())
		if err != nil {
			return item, err
		}

		item = UpKeepingItem{
			UserID:      userID,
			QueryID:     qid,
			UpKeepingID: ukID,
			Keepers:     keepers,
			KeeperSLA:   int32(len(keepers)),
			Providers:   providers,
			ProviderSLA: int32(len(providers)),
			Duration:    duration.Int64(),
			Capacity:    capacity.Int64(),
			Price:       price.Int64(),
			StartTime:   startTime.Int64(),
			EndDate:     endDate.Int64(),
			Cycle:       cycle.Int64(),
			NeedPay:     needPay.Int64(),
			Proofs:      proofs,
		}
		return item, nil
	}
}

// DeployRoot is
func DeployRoot(sk, userID, queryID string, rdo bool) (rootID string, err error) {
	utils.MLogger.Info("Begin to deploy root contract...")
	uaddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return rootID, err
	}

	queryAddr, err := address.GetAddressFromID(queryID)
	if err != nil {
		return rootID, err
	}

	// getbalance
	balance, err := contracts.QueryBalance(uaddr.String())
	if err == nil {
		utils.MLogger.Infof("%s (%s) has balance: %s", userID, uaddr.String(), balance)
	}

	// deploy root
	rootAddr, err := contracts.DeployRoot(sk, uaddr, queryAddr, rdo)
	if err != nil {
		utils.MLogger.Error("fail to deploy root contract: ", err)
		return queryID, err
	}

	utils.MLogger.Info(uaddr.String(), " has new root: ", rootAddr.String())

	rootID, err = address.GetIDFromAddress(rootAddr.String())
	if err != nil {
		return rootID, err
	}

	utils.MLogger.Info("Finish deploy root contract: ", rootID)

	return rootID, nil
}

// GetRoot gets
func GetRoot(userID, queryID string) (string, error) {
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return "", err
	}

	queryAddr, err := address.GetAddressFromID(queryID)
	if err != nil {
		return "", err
	}

	rootAddr, _, err := contracts.GetRoot(userAddr, userAddr, queryAddr.String())
	if err != nil {
		return "", err
	}

	rootID, err := address.GetIDFromAddress(rootAddr.String())
	if err != nil {
		return "", err
	}

	return rootID, nil
}

func SetMerkleRoot(sk, rootID string, key int64, val [32]byte) error {
	rootAddr, err := address.GetAddressFromID(rootID)
	if err != nil {
		return err
	}

	err = contracts.SetMerkleRoot(sk, rootAddr, key, val)
	if err != nil {
		return err
	}

	return nil
}

func GetLatestMerkleRoot(rootID string) (int64, [32]byte, error) {
	var val [32]byte
	rootAddr, err := address.GetAddressFromID(rootID)
	if err != nil {
		return 0, val, err
	}

	return contracts.GetLatestMerkleRoot(rootAddr, rootAddr)
}

// DeployChannel is
func DeployChannel(userID, queryID, proID, hexSk string, storeDays, storeSize int64, redo bool) (string, error) {
	utils.MLogger.Info("Begin to deploy channel contract...")
	var chanAddr string
	localAddress, err := address.GetAddressFromID(userID)
	if err != nil {
		return chanAddr, err
	}

	queryAddress, err := address.GetAddressFromID(queryID)
	if err != nil {
		return chanAddr, err
	}

	//依次与各provider签署channel合约，存储时间单位秒
	timeOut := big.NewInt(int64(storeDays * 24 * 60 * 60))
	moneyToChannel := big.NewInt(utils.READPRICEPERMB * int64(storeSize*100)) //暂定往每个channel合约中存储金额为：存储大小 x 每MB单价

	proAddress, err := address.GetAddressFromID(proID)
	if err != nil {
		return chanAddr, err
	}

	cAddr, err := contracts.DeployChannelContract(hexSk, localAddress, queryAddress, proAddress, timeOut, moneyToChannel, redo)
	if err != nil {
		utils.MLogger.Error("Deploy channel contract failed: ", err)
		return chanAddr, err
	}

	chanID, err := address.GetIDFromAddress(cAddr.String())
	if err != nil {
		return chanAddr, err
	}

	utils.MLogger.Info("Finish deploy channel contract: ", chanID)

	return chanID, err
}

//GetChannelInfo used to getchannel-contract item
func GetChannelInfo(localID, channelID string) (ChannelItem, error) {
	var item ChannelItem
	localAddress, err := address.GetAddressFromID(localID)
	if err != nil {
		return item, err
	}

	chanAddress, err := address.GetAddressFromID(channelID)
	if err != nil {
		return item, err
	}

	channelInstance, err := channel.NewChannel(chanAddress, contracts.GetClient(contracts.EndPoint))
	if err != nil {
		return item, err
	}
	retryCount := 0
	for {
		retryCount++
		startDate, timeOut, sender, receiver, err := channelInstance.GetInfo(&bind.CallOpts{
			From: localAddress,
		})
		if err != nil {
			if retryCount > 10 {
				return item, err
			}
			time.Sleep(60 * time.Second)
			continue
		}

		uid, err := address.GetIDFromAddress(sender.String())
		if err != nil {
			return item, err
		}

		pid, err := address.GetIDFromAddress(receiver.String())
		if err != nil {
			return item, err
		}

		item = ChannelItem{
			StartTime: startDate.Int64(),
			Duration:  timeOut.Int64(),
			ChannelID: channelID,
			UserID:    uid,
			ProID:     pid,
			Value:     big.NewInt(0),
		}
		break
	}

	retryCount = 0
	for {
		retryCount++
		balance, err := contracts.QueryBalance(chanAddress.String())
		if err != nil {
			if retryCount > 10 {
				return item, err
			}
			time.Sleep(30 * time.Second)
			continue
		}

		item.Money = balance
		return item, nil
	}
}

// GetBalance gets balance from
func GetBalance(pid string) *big.Int {
	paddr, err := address.GetAddressFromID(pid)
	if err != nil {
		return big.NewInt(0)
	}

	retryCount := 0
	for {
		retryCount++
		balance, err := contracts.QueryBalance(paddr.String())
		if err != nil {
			if retryCount > 10 {
				return big.NewInt(0)
			}
			time.Sleep(30 * time.Second)
			continue
		}

		return balance
	}
}

// GetLatestChannel gets
func GetLatestChannel(userID, queryID, proID string) (ChannelItem, error) {
	utils.MLogger.Debugf("get channel for user %s, provider %s, and query %s", userID, proID, queryID)
	var item ChannelItem
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return item, err
	}

	queryAddr, err := address.GetAddressFromID(queryID)
	if err != nil {
		return item, err
	}

	proAddr, err := address.GetAddressFromID(proID)
	if err != nil {
		return item, err
	}

	channelAddr, _, err := contracts.GetLatestChannel(userAddr, userAddr, proAddr, queryAddr)
	if err != nil {
		return item, err
	}

	channelID, err := address.GetIDFromAddress(channelAddr.String())
	if err != nil {
		return item, err
	}

	item, err = GetChannelInfo(userID, channelID)
	if err != nil {
		return item, err
	}

	utils.MLogger.Debugf("get channel %s for user %s, provider %s, and query %s", channelID, userID, proID, queryID)

	if item.UserID != userID || item.ProID != proID {
		utils.MLogger.Errorf("got queryID %s, sender %s and receiver %s are not compatabile: ", item.ChannelID, item.UserID, item.ProID)
		return item, ErrWrongContarctContent
	}

	return item, nil
}

//SignForChannel user sends a private key signature to the provider
func SignForChannel(channelID, hexKey string, value *big.Int) (sig []byte, err error) {
	channelAddr, err := address.GetAddressFromID(channelID)
	if err != nil {
		return nil, err
	}

	//(channelAddress, value)的哈希值
	valueNew := common.LeftPadBytes(value.Bytes(), 32)
	hash := crypto.Keccak256(channelAddr.Bytes(), valueNew) //32Byte

	//私钥格式转换
	skECDSA, err := utils.EthskToECDSAsk(hexKey)
	if err != nil {
		return sig, err
	}

	//私钥对上述哈希值签名
	sig, err = crypto.Sign(hash, skECDSA)
	if err != nil {
		return sig, err
	}

	pubKey, err := utils.GetPkFromEthSk(hexKey)
	if err != nil {
		utils.MLogger.Error("Get public key fail: ", err)
		return nil, err
	}

	message := &mpb.ChannelSign{
		Sig:       sig,
		PubKey:    pubKey,
		Value:     value.Bytes(),
		ChannelID: channelID,
	}

	mes, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}

	return mes, nil
}

//VerifyChannelSig provider used to verify user's signature for channel-contract
func VerifyChannelSign(cSign *mpb.ChannelSign) (verify bool) {
	channelAddr, err := address.GetAddressFromID(cSign.GetChannelID())
	if err != nil {
		return false
	}

	//(channelAddress, value)的哈希值
	valueNew := common.LeftPadBytes(cSign.GetValue(), 32)
	hash := crypto.Keccak256(channelAddr.Bytes(), valueNew)

	//验证签名
	return crypto.VerifySignature(cSign.GetPubKey(), hash, cSign.GetSig()[:64])
}

// CloseChannel closes channel by provider
func CloseChannel(channelID, sk string, sign []byte, value *big.Int) error {
	chanAddress, err := address.GetAddressFromID(channelID)
	if err != nil {
		return err
	}

	return contracts.CloseChannel(chanAddress, sk, sign, value)
}

// KillChannel closes chnannel by users
func KillChannel(channelID, sk string) error {
	chanAddress, err := address.GetAddressFromID(channelID)
	if err != nil {
		return err
	}

	return contracts.ChannelTimeout(chanAddress, sk)
}

// GetKeepersOfPro get keepers of some provider
func GetKeepersOfPro(peerID string) ([]string, bool) {
	res, ok := kpMap.Load(peerID)
	if !ok {
		return nil, false
	}
	return res.(*kpItem).keeperIDs, true
}

// SaveKpMap saves kpmap
func SaveKpMap(peerID string) error {
	localAddr, err := address.GetAddressFromID(peerID)
	if err != nil {
		return err
	}
	kps, err := contracts.GetAllKeeperInKPMap(localAddr)
	if err != nil {
		return err
	}

	for _, kpaddr := range kps {
		pids, err := contracts.GetProviderInKPMap(localAddr, kpaddr)
		if err != nil {
			continue
		}
		if len(pids) > 0 {
			keeperID, _ := address.GetIDFromAddress(kpaddr.String())
			kidList := []string{keeperID}
			for _, paddr := range pids {
				pid, _ := address.GetIDFromAddress(paddr.String())
				res, ok := kpMap.Load(pid)
				if ok {
					res.(*kpItem).keeperIDs = append(res.(*kpItem).keeperIDs, keeperID)
				} else {
					kidres := &kpItem{
						keeperIDs: kidList,
					}
					kpMap.Store(keeperID, kidres)
				}
			}
		}
	}
	return nil
}

// BuildSignMessage builds sign message for test or repair
func BuildSignMessage() ([]byte, error) {
	message := &mpb.ChannelSign{
		Value:     []byte("123"),
		ChannelID: "test",
	}

	mes, err := proto.Marshal(message)
	if err != nil {
		utils.MLogger.Error("protoMarshal failed: ", err)
		return nil, err
	}
	return mes, nil
}

//SignForStPay keeper signature
func SignForStPay(upKeepingAddr, providerAddr common.Address, hexKey string, stStart, stLength, stValue *big.Int, merkleRoot [32]byte, share []int) ([]byte, error) {
	var sig []byte
	//(upKeepingAddr, providerAddr, stStart, stLength, stValue, merkleRoot, share)的哈希值
	stStartNew := common.LeftPadBytes(stStart.Bytes(), 32)
	stLengthNew := common.LeftPadBytes(stLength.Bytes(), 32)
	stValueNew := common.LeftPadBytes(stValue.Bytes(), 32)
	var merkleRootNew = merkleRoot[:]
	//hash := crypto.Keccak256(upKeepingAddr.Bytes(), providerAddr.Bytes(), stStartNew, stLengthNew, stValueNew, merkleRootNew, share) //32Byte
	data := [][]byte{}
	for i := 0; i < len(share); i++ {
		data = append(data, common.LeftPadBytes(big.NewInt(int64(share[i])).Bytes(), 32))
	}
	//keccak256内部实现
	d := sha3.NewLegacyKeccak256()
	d.Write(upKeepingAddr.Bytes())
	d.Write(providerAddr.Bytes())
	d.Write(stStartNew)
	d.Write(stLengthNew)
	d.Write(stValueNew)
	d.Write(merkleRootNew)
	for i := 0; i < len(data); i++ {
		d.Write(data[i])
	}
	hash := d.Sum(nil)

	//私钥格式转换
	skECDSA, err := utils.EthskToECDSAsk(hexKey)
	if err != nil {
		return sig, err
	}

	//私钥对上述哈希值签名
	sig, err = crypto.Sign(hash, skECDSA)
	if err != nil {
		return sig, err
	}

	return sig, nil
}
