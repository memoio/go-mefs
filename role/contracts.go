package role

import (
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gogo/protobuf/proto"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/contracts/upKeeping"
	id "github.com/memoio/go-mefs/crypto/identity"
	mpb "github.com/memoio/go-mefs/pb"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"golang.org/x/crypto/sha3"
)

// ProviderItem has provider's info
type ProviderItem struct {
	ProviderID  string   // providerid
	Capacity    int64    // Bytes, pledge capacity
	PledgeMoney *big.Int // pledge money
	StartTime   int64    // start time
}

// KeeperItem has provider's info
type KeeperItem struct {
	KeeperID    string   // providerid
	PledgeMoney *big.Int // pledge money
	StartTime   int64    // start time; not in contract
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
	Duration    int64    //存储时间，单位second
	Capacity    int64    // MB
	Price       *big.Int // 部署的价格: wei/(MB*h)
	StartTime   int64    // 部署的时间: second
	Money       *big.Int
	EndTime     int64 // second
	Cycle       int64 // 设置周期
	NeedPay     *big.Int
	Proofs      []upKeeping.UpKeepingProof
}

// OfferItem has offer information
type OfferItem struct {
	ProviderID string // 部署Offer的providerid
	OfferID    string // offer address : id format
	Capacity   int64
	Duration   int64
	Price      *big.Int //合约给出的单价
	CreateDate int64    //合约创建时间
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
	Dirty     bool   // value is change?
}

// QueryItem has query information
type QueryItem struct {
	UserID       string // 部署Query的userid
	QueryID      string
	Capacity     int64
	Duration     int64
	Price        *big.Int // 合约给出的单价
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

const retryGetInfoSleepTime = time.Minute

func QueryBalance(localID string) (*big.Int, error) {
	localAddress, err := address.GetAddressFromID(localID)
	if err != nil {
		return nil, err
	}

	//获得用户的账户余额
	a := contracts.NewCA(localAddress, "")
	return a.QueryBalance(localAddress.Hex())
}

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

	r := contracts.NewCR(localAddress, "")
	isKeeper, isBanned, money, ptime, err := r.GetKeeperInfo(keeperAddress)
	if err != nil {
		return item, err
	}

	if isKeeper && !isBanned {
		item = KeeperItem{
			KeeperID:    keeperID,
			PledgeMoney: money,
			StartTime:   ptime,
		}
		return item, nil
	}

	return item, ErrNotKeeper
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

	r := contracts.NewCR(localAddress, "")
	isProvider, isBanned, money, stime, err := r.GetProviderInfo(proAddress)
	if err != nil {
		return item, err
	}

	price, err := r.GetProviderPrice()
	if err != nil {
		return item, err
	}

	cap := new(big.Int)
	weiPrice := new(big.Float).SetInt(price)
	weiPrice.Quo(weiPrice, contracts.GetMemoPrice())
	weiPrice.Int(price)
	if price.Sign() > 0 {
		cap.Quo(money, price)
	}

	if isProvider && !isBanned {
		item = ProviderItem{
			ProviderID:  proID,
			PledgeMoney: money,
			StartTime:   stime,
			Capacity:    cap.Int64(),
		}
		return item, nil
	}

	return item, ErrNotProvider
}

//GetOfferInfo get provider's offer-info
func GetOfferInfo(localID, offerID string) (OfferItem, error) {
	var item OfferItem

	localAddress, err := address.GetAddressFromID(localID)
	if err != nil {
		return item, err
	}

	offerAddress, err := address.GetAddressFromID(offerID)
	if err != nil {
		return item, err
	}

	m := contracts.NewCM(localAddress, "")
	capacity, duration, price, createDate, err := m.GetOfferInfo(offerAddress)
	if err != nil {
		return item, err
	}

	item = OfferItem{
		Capacity:   capacity,
		Duration:   duration,
		Price:      price,
		OfferID:    offerID,
		CreateDate: createDate,
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

	m := contracts.NewCM(userAddr, "")
	oAddrs, err := m.GetOfferAddrs(proAddr)
	if err != nil {
		utils.MLogger.Info("get ", proID, " 's offer address err: ", err)
		return item, err
	}

	if len(oAddrs) < 1 {
		utils.MLogger.Info("get ", proID, " 's offer address is empty")
		return item, ErrNoContract
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

// GetLatestQuery gets
func GetLatestQuery(userID string) (QueryItem, error) {
	var item QueryItem
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return item, err
	}

	m := contracts.NewCM(userAddr, "")
	qAddrs, err := m.GetQueryAddrs(userAddr)
	if err != nil {
		utils.MLogger.Info("get ", userID, " 's query address err: ", err)
		return item, err
	}

	if len(qAddrs) < 1 {
		utils.MLogger.Info("get ", userID, " 's query address is empty")
		return item, ErrNoContract
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

// GetAllQuerys gets all query IDs
func GetAllQuerys(userID string) ([]string, error) {
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return nil, err
	}

	m := contracts.NewCM(userAddr, "")
	qAddrs, err := m.GetQueryAddrs(userAddr)
	if err != nil {
		return nil, err
	}
	res := make([]string, len(qAddrs))
	for i := 0; i < len(qAddrs); i++ {
		queryID, err := address.GetIDFromAddress(qAddrs[i].String())
		if err != nil {
			return nil, err
		}
		res[i] = queryID
	}

	return res, nil
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

	m := contracts.NewCM(localAddr, "")
	capacity, duration, price, ks, ps, completed, err := m.GetQueryInfo(queryAddr)
	if err != nil {
		return item, err
	}

	item = QueryItem{
		Capacity:     capacity,
		Duration:     duration,
		Price:        price,
		KeeperNums:   int32(ks),
		ProviderNums: int32(ps),
		Completed:    completed,
		QueryID:      queryID,
	}
	return item, nil
}

// DeployUpKeeping is
func DeployUpKeeping(userID, queryID, hexSk string, ks, ps []string, storeDays, storeSize int64, storePrice *big.Int, stPayCycle int64, redo bool) (ukID string, err error) {
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

	weiPrice := new(big.Float).SetInt(storePrice)
	weiPrice.Quo(weiPrice, contracts.GetMemoPrice())
	newPrice := big.NewInt(0)
	weiPrice.Int(newPrice)

	moneyAccount := big.NewInt(24)
	moneyAccount.Mul(moneyAccount, newPrice)
	moneyAccount.Mul(moneyAccount, big.NewInt(storeSize))
	moneyAccount.Mul(moneyAccount, big.NewInt(storeDays))

	// getbalance
	balance, err := QueryBalance(userID)
	if err != nil {
		return ukID, err
	}

	utils.MLogger.Infof("%s (%s) has balance: %s", userID, localAddress.String(), balance)

	if moneyAccount.Cmp(balance) > 0 {
		utils.MLogger.Errorf("%s (%s) has balance: %s, need %d to deploy upkeeping", userID, localAddress.String(), balance, moneyAccount)
		return ukID, ErrNotEnoughBalance
	}

	utils.MLogger.Info("Begin to deploy upkeeping contract...")

	duration := storeDays * 24 * 60 * 60
	u := contracts.NewCU(localAddress, hexSk)
	ukAddr, err := u.DeployUpkeeping(queryAddress, keepers, providers, duration, storeSize, storePrice, stPayCycle, moneyAccount, redo)
	if err != nil {
		utils.MLogger.Error("Deploy upkeeping contract failed: ", err)
		return ukID, err
	}

	m := contracts.NewCM(localAddress, hexSk)
	err = m.SetQueryCompleted(queryAddress)
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

// GetUpkeepingInfo get Upkeeping-contract's params by ukID
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

	ukMoney, err := QueryBalance(ukID)
	if err != nil {
		return item, err
	}

	u := contracts.NewCU(localAddr, "")
	queryAddr, keepers, providers, duration, capacity, price, startTime, endDate, cycle, needPay, proofs, err := u.GetOrder(ukAddr)
	if err != nil {
		return item, err
	}

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
		Price:       price,
		StartTime:   startTime.Int64(),
		EndTime:     endDate.Int64(),
		Cycle:       cycle.Int64(),
		NeedPay:     needPay,
		Proofs:      proofs,
		Money:       ukMoney,
	}
	return item, nil
}

// GetUpKeeping get Upkeeping-contract's params by queryID
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

	u := contracts.NewCU(userAddr, "")

	ukAddr, _, err := u.GetUpkeeping(userAddr, queryAddr.String())
	if err != nil {
		return item, err
	}

	qid, err := address.GetIDFromAddress(queryAddr.String())
	if err != nil {
		return item, err
	}

	ukID, err := address.GetIDFromAddress(ukAddr.String())
	if err != nil {
		return item, err
	}

	ukMoney, err := QueryBalance(ukID)
	if err != nil {
		return item, err
	}

	queryAddr, keepers, providers, duration, capacity, price, startTime, endDate, cycle, needPay, proofs, err := u.GetOrder(ukAddr)
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
		Price:       price,
		StartTime:   startTime.Int64(),
		EndTime:     endDate.Int64(),
		Cycle:       cycle.Int64(),
		NeedPay:     needPay,
		Proofs:      proofs,
		Money:       ukMoney,
	}
	return item, nil
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
	balance, err := QueryBalance(userID)
	if err != nil {
		return rootID, err
	}

	utils.MLogger.Infof("%s (%s) has balance: %s", userID, uaddr.String(), balance)

	// deploy root
	r := contracts.NewCRoot(uaddr, sk)
	rootAddr, err := r.DeployRoot(queryAddr, rdo)
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

// GetRoot gets rootID
func GetRoot(userID, queryID string) (string, error) {
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return "", err
	}

	queryAddr, err := address.GetAddressFromID(queryID)
	if err != nil {
		return "", err
	}

	r := contracts.NewCRoot(userAddr, "")
	rootAddr, _, err := r.GetRoot(userAddr, queryAddr.String())
	if err != nil {
		return "", err
	}

	rootID, err := address.GetIDFromAddress(rootAddr.String())
	if err != nil {
		return "", err
	}

	return rootID, nil
}

func GetLatestMerkleRoot(rootID string) (int64, [32]byte, error) {
	var val [32]byte
	rootAddr, err := address.GetAddressFromID(rootID)
	if err != nil {
		return 0, val, err
	}

	r := contracts.NewCRoot(rootAddr, "")
	return r.GetLatestMerkleRoot(rootAddr)
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
	// read ps times
	moneyToChannel := new(big.Int).Mul(big.NewInt(utils.READPRICE), big.NewInt(storeSize))
	moneyToChannel.Mul(moneyToChannel, big.NewInt(3))

	weiPrice := new(big.Float).SetInt(moneyToChannel)
	weiPrice.Quo(weiPrice, contracts.GetMemoPrice())
	weiPrice.Int(moneyToChannel)

	balance, err := QueryBalance(userID)
	if err != nil {
		return chanAddr, err
	}

	utils.MLogger.Infof("%s (%s) has balance: %s", userID, localAddress.Hex(), balance)

	if moneyToChannel.Cmp(balance) > 0 {
		return chanAddr, ErrNotEnoughBalance
	}

	proAddress, err := address.GetAddressFromID(proID)
	if err != nil {
		return chanAddr, err
	}

	ch := contracts.NewCH(localAddress, hexSk)
	cAddr, err := ch.DeployChannelContract(queryAddress, proAddress, timeOut, moneyToChannel, redo)
	if err != nil {
		utils.MLogger.Error("Deploy channel contract failed: ", err)
		return chanAddr, err
	}

	chanID, err := address.GetIDFromAddress(cAddr.String())
	if err != nil {
		return chanAddr, err
	}

	utils.MLogger.Info("Finish deploy channel contract: ", chanID)

	return chanID, nil
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

	ch := contracts.NewCH(localAddress, "")
	startDate, timeOut, sender, receiver, err := ch.GetChannelInfo(chanAddress)
	if err != nil {
		return item, err
	}

	uid, err := address.GetIDFromAddress(sender.String())
	if err != nil {
		return item, err
	}

	pid, err := address.GetIDFromAddress(receiver.String())
	if err != nil {
		return item, err
	}

	ba, err := QueryBalance(channelID)
	if err != nil {
		return item, err
	}

	item = ChannelItem{
		StartTime: startDate,
		Duration:  timeOut,
		ChannelID: channelID,
		UserID:    uid,
		ProID:     pid,
		Value:     big.NewInt(0),
		Money:     ba,
	}
	return item, nil

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

	cChannel := contracts.NewCH(userAddr, "")
	channelAddr, _, err := cChannel.GetLatestChannel(userAddr, proAddr, queryAddr)
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

func GetAllChannels(userID, queryID, proID string) ([]string, error) {
	utils.MLogger.Debugf("get channel for user %s, provider %s, and query %s", userID, proID, queryID)
	userAddr, err := address.GetAddressFromID(userID)
	if err != nil {
		return nil, err
	}

	queryAddr, err := address.GetAddressFromID(queryID)
	if err != nil {
		return nil, err
	}

	proAddr, err := address.GetAddressFromID(proID)
	if err != nil {
		return nil, err
	}

	ch := contracts.NewCH(userAddr, "")
	channelAddrs, err := ch.GetChannelAddrs(userAddr, proAddr, queryAddr)
	if err != nil {
		return nil, err
	}

	var res []string
	for _, cAddr := range channelAddrs {
		channelID, err := address.GetIDFromAddress(cAddr.String())
		if err != nil {
			return nil, err
		}
		res = append(res, channelID)
	}

	return res, nil
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
	skECDSA, err := id.ECDSAStringToSk(hexKey)
	if err != nil {
		return sig, err
	}

	//私钥对上述哈希值签名
	sig, err = crypto.Sign(hash, skECDSA)
	if err != nil {
		return sig, err
	}

	pubKey, err := id.GetCompressPubByte(hexKey)
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

//VerifyChannelSign provider used to verify user's signature for channel-contract
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

	ch := contracts.NewCH(chanAddress, sk)
	return ch.CloseChannel(chanAddress, sign, value)
}

// KillChannel closes chnannel by users
func KillChannel(channelID, sk string) error {
	chanAddress, err := address.GetAddressFromID(channelID)
	if err != nil {
		return err
	}

	ch := contracts.NewCH(chanAddress, sk)
	return ch.ChannelTimeout(chanAddress)
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
	peerAddr, err := address.GetAddressFromID(peerID)
	if err != nil {
		return err
	}

	r := contracts.NewCR(peerAddr, "")
	kps, err := r.GetAllKeeperInKPMap()
	if err != nil {
		return err
	}

	for _, kpaddr := range kps {
		pids, err := r.GetProviderInKPMap(kpaddr)
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
					kpMap.Store(pid, kidres)
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

func GetHashForST(ukID, providerID string, stStart, stLength, stValue *big.Int, merkleRoot []byte, share []int64) ([]byte, error) {
	upKeepingAddr, err := address.GetAddressFromID(ukID)
	if err != nil {
		return nil, err
	}

	providerAddr, err := address.GetAddressFromID(providerID)
	if err != nil {
		return nil, err
	}

	//keccak256(upKeepingAddr, providerAddr, stStart, stLength, stValue, merkleRoot, share)
	d := sha3.NewLegacyKeccak256()
	d.Write(upKeepingAddr.Bytes())
	d.Write(providerAddr.Bytes())
	d.Write(common.LeftPadBytes(stStart.Bytes(), 32))
	d.Write(common.LeftPadBytes(stLength.Bytes(), 32))
	d.Write(common.LeftPadBytes(stValue.Bytes(), 32))
	d.Write(merkleRoot[:])
	for i := 0; i < len(share); i++ {
		d.Write(common.LeftPadBytes(big.NewInt(int64(share[i])).Bytes(), 32))
	}

	return d.Sum(nil), nil
}

//SignForStPay keeper signature
func SignForStPay(upKeepingAddr, providerAddr common.Address, hexKey string, stStart, stLength, stValue *big.Int, merkleRoot [32]byte, share []int64) ([]byte, error) {
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

	//私钥对上述哈希值签名
	sig, err := id.Sign(hexKey, hash)
	if err != nil {
		return sig, err
	}

	return sig, nil
}

//GetHashForAddProvider keeper signature
func GetHashForAddProvider(upKeepingAddr common.Address, providerAddr []common.Address) ([]byte, error) {
	//(upKeepingAddr, []providerAddr)的哈希值

	//keccak256内部实现
	d := sha3.NewLegacyKeccak256()
	d.Write(upKeepingAddr.Bytes())

	for i := 0; i < len(providerAddr); i++ {
		d.Write(common.LeftPadBytes(providerAddr[i].Bytes(), 32))
	}

	return d.Sum(nil), nil
}

//SignForAddProvider keeper signature
func SignForAddProvider(upKeepingAddr common.Address, providerAddr []common.Address, hexKey string) ([]byte, error) {
	//(upKeepingAddr, []providerAddr)的哈希值

	//keccak256内部实现
	d := sha3.NewLegacyKeccak256()
	d.Write(upKeepingAddr.Bytes())

	for i := 0; i < len(providerAddr); i++ {
		d.Write(common.LeftPadBytes(providerAddr[i].Bytes(), 32))
	}
	hash := d.Sum(nil)

	//私钥对上述哈希值签名
	sig, err := id.Sign(hexKey, hash)
	if err != nil {
		return sig, err
	}

	return sig, nil
}

//SignForSetStop keeper signature to set provider or keeper stop in upkeeping contract
func SignForSetStop(upKeepingAddr, providerAddr common.Address, hexKey string) ([]byte, error) {
	hash := crypto.Keccak256(upKeepingAddr.Bytes(), providerAddr.Bytes())

	sig, err := id.Sign(hexKey, hash)
	if err != nil {
		return nil, err
	}

	return sig, nil
}
