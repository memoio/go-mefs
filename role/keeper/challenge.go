package keeper

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/contracts"
	df "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/role/pb"
	b58 "github.com/mr-tron/base58/base58"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
)

//LedgerInfo 存放挑战信息的内存结构体
//key为结构体PU，value为结构体*chalinfo
var LedgerInfo sync.Map //key为PU结构体，value为chalinfo结构体

//PU 作为其他结构体的key，
//目前用PU做key的结构体有LedgerInfo，PayInfo,用pid和uid共同索引信息
type PU struct {
	pid string
	uid string
}

//chalinfo 作为LedgerInfo的value
type chalinfo struct {
	Time        sync.Map //某provider下user数据在某时刻发起挑战的结果，key为挑战发起时间的时间戳，格式为int64,value为chalresult
	Cid         sync.Map
	tmpCid      sync.Map
	inChallenge int
	maxlength   uint32
}

type cidInfo struct {
	res       bool
	repair    int32
	availtime int64
	offset    int
}

//chalresult 挑战结果在内存中的结构
//作为chalinfo.Time的value 记录单次挑战的各项信息
type chalresult struct {
	kid            string //挑战发起者
	pid            string //挑战对象
	uid            string //挑战的数据所属对象
	challenge_time int64  //挑战发起时间 使用unix时间戳
	sum            uint32 //挑战总空间
	length         uint32 //挑战成功空间
	h              int    //挑战的随机数
	res            bool   //挑战是否成功
	proof          string //挑战结果的证据
	//allproof       accproof  // 每个挑战结果的证据保存，当前先忽略
}

//==========================LegerInfo数据结构操作============================
//getChalresult 传入各层key，获取对应的chalresult结构指针，若无法取到，可能结构体还没被创建，返回nil
func getChalresult(thisPU PU, time int64) (*chalresult, bool) {
	thischalinfo, ok := getChalinfo(thisPU)
	if thischalinfo == nil {
		return nil, false
	}
	if !ok {
		return nil, false
	}
	thischalresult, ok := thischalinfo.Time.Load(time)
	if !ok {
		return nil, false
	}
	return thischalresult.(*chalresult), true
}

//getChalinfo 输入pid和uid 获取LedgerInfo中的chalinfo指针
func getChalinfo(thisPU PU) (*chalinfo, bool) {
	thischalinfo, ok := LedgerInfo.Load(thisPU)
	if !ok {
		return nil, false
	}
	return thischalinfo.(*chalinfo), true
}

func doAddBlocktoLedger(pid string, uid string, blockid string, offset int) error {
	pu := PU{
		pid: pid,
		uid: uid,
	}

	newcidinfo := &cidInfo{
		availtime: utils.GetUnixNow(),
		offset:    offset,
		repair:    0,
	}

	if thischalinfo, ok := getChalinfo(pu); ok {
		if thischalinfo.inChallenge == 1 {
			thischalinfo.tmpCid.Store(blockid, newcidinfo)
		} else if thischalinfo.inChallenge == 0 {
			thischalinfo.Cid.Store(blockid, newcidinfo)
		}
		return nil
	}

	var Cid sync.Map
	var Time sync.Map
	Cid.Store(blockid, newcidinfo)
	newchalinfo := &chalinfo{
		Cid:  Cid,
		Time: Time,
	}
	LedgerInfo.Store(pu, newchalinfo)
	return nil

}

//==========================================================================

// 挑战过程的起始函数 定时对本节点连接的provider发起挑战
func challengeRegular(ctx context.Context) { //定期挑战
	fmt.Println("ChallengeRegular() start!")
	ticker := time.NewTicker(CHALTIME)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go func() {
				ChallengeProviderBLS12()
			}()
		}
	}
}

//ChallengeProviderBLS12 对链接的所有provider发起一次挑战
func ChallengeProviderBLS12() {
	LedgerInfo.Range(func(k, v interface{}) bool { //对PU对进行循环
		pu := k.(PU)
		isTestUser := false
		//没有部署合约的不进行挑战修复
		addr, err := address.GetAddressFromID(pu.uid)
		if err != nil {
			isTestUser = true
		}
		config, _ := localNode.Repo.Config()
		endPoint := config.Eth //获取endPoint
		_, _, err = contracts.GetUKFromResolver(endPoint, addr)
		if err != nil {
			isTestUser = true
		}

		thischalinfo := v.(*chalinfo)
		challengetime := utils.GetUnixNow()
		var ret []string
		var sum, maxlength uint32
		if !isTestUser {
			thischalinfo.inChallenge = 1
		}
		thischalinfo.Cid.Range(func(key, value interface{}) bool { //对该PU对中provider保存的块循环
			cInfo := value.(*cidInfo)
			//测试User不进行挑战修复
			if isTestUser {
				cInfo.availtime = utils.GetUnixNow()
				return true
			}
			str := key.(string) + "_" + strconv.Itoa(cInfo.offset)
			ret = append(ret, str)
			maxlength += uint32((df.MAXOFFSET + 1) * df.DefaultSegmentSize)
			sum += uint32((cInfo.offset + 1) * df.DefaultSegmentSize)
			return true
		})
		//测试User不进行挑战修复
		if isTestUser {
			return true
		}
		thischalresult := &chalresult{ //对provider发起挑战之前，先构造好本次挑战信息的结构体
			challenge_time: challengetime,
			sum:            sum,
		}
		thischalinfo.Time.Store(challengetime, thischalresult)
		if len(ret) != 0 {
			err := doChallengeBLS12(pu, ret, challengetime) //对该provider关于该user发起一次挑战
			if err != nil {
				thischalinfo.inChallenge = 0
			}
		}
		return true
	})
}

//doChallengeBLS12 对某个PU对 进行一次挑战，传入时间和挑战的块信息
func doChallengeBLS12(pu PU, blocks []string, chaltime int64) error {
	chal := mcl.GenChallenge(blocks)

	if thischalresult, ok := getChalresult(pu, chaltime); ok {
		thischalresult.h = chal.C

		hProto := &pb.Chalnum{
			PubC:    int64(chal.C),
			Indices: chal.Indices,
		}
		hByte, err := proto.Marshal(hProto)
		if err != nil {
			log.Println("marshal h failed")
			return err
		}

		km, err := metainfo.NewKeyMeta(pu.uid, metainfo.Challenge, utils.UnixToString(chaltime))
		if err != nil {
			log.Println("construct challenge KV error :", err)
			return err
		}
		metaValue := b58.Encode(hByte)
		_, err = sendMetaRequest(km, metaValue, pu.pid)
		if err != nil {
			log.Println("DoChallengeBLS12 error :", err)
			return err
		}
	}
	return nil

}

func checkLedger(ctx context.Context) {
	fmt.Println("CheckLedger() start!")
	time.Sleep(2 * CHALTIME)
	ticker := time.NewTicker(CHECKTIME)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go func() {
				LedgerInfo.Range(func(k, v interface{}) bool {
					thischalinfo := v.(*chalinfo)
					thischalinfo.Cid.Range(func(key, value interface{}) bool {
						//fmt.Println("avaltime :", value.(*cidInfo).availtime)
						if EXPIRETIME < (utils.GetUnixNow()-value.(*cidInfo).availtime) && value.(*cidInfo).repair < 3 {
							fmt.Println("need repair cid :", key.(string))
							value.(*cidInfo).repair += 1
							repch <- key.(string)
						}
						return true
					})
					return true
				})
			}()
		}
	}
}

//获得用于证明的user的公用参数
func getUserBLS12Config(userID string) (*UserBLS12Config, error) {
	if !IsKeeperServiceRunning() {
		return nil, ErrKeeperServiceNotReady
	}
	userPubKey := new(mcl.PublicKey)
	userConfig := &UserBLS12Config{}

	kmBls12, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypeCfg, metainfo.CfgTypeBls12)
	if err != nil {
		return nil, err
	}
	userconfigkey := kmBls12.ToString()
	userconfigbyte, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(userconfigkey, "local")
	if err != nil {
		return userConfig, err
	}

	userconfigProto := &pb.UserBLS12Config{}
	err = proto.Unmarshal(userconfigbyte, userconfigProto) //反序列化
	if err != nil {
		return userConfig, err
	}
	err = userPubKey.BlsPK.Deserialize(userconfigProto.PubkeyBls)
	if err != nil {
		return userConfig, err
	}
	err = userPubKey.G.Deserialize(userconfigProto.PubkeyG)
	if err != nil {
		return userConfig, err
	}
	userPubKey.U = make([]mcl.G1, mcl.N)
	for i, u := range userconfigProto.PubkeyU {
		if i >= mcl.N {
			break
		}
		err = userPubKey.U[i].Deserialize(u)
		if err != nil {
			return userConfig, err
		}
	}
	userPubKey.W = make([]mcl.G2, mcl.N)
	for i, w := range userconfigProto.PubkeyW {
		if i >= mcl.N {
			break
		}
		err = userPubKey.W[i].Deserialize(w)
		if err != nil {
			return userConfig, err
		}
	}

	userConfig = &UserBLS12Config{
		PubKey: userPubKey,
	}
	return userConfig, nil
}

func getUserBLS12ConfigByte(userID string) ([]byte, error) {
	if !IsKeeperServiceRunning() {
		return nil, ErrKeeperServiceNotReady
	}
	kmBls12, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypeCfg, metainfo.CfgTypeBls12)
	if err != nil {
		return nil, err
	}
	userconfigkey := kmBls12.ToString()
	userconfigbyte, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(userconfigkey, "local")
	if err != nil {
		return nil, err
	}
	return userconfigbyte, nil
}
