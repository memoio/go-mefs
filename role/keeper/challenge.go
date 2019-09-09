package keeper

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	mcl "github.com/memoio/go-mefs/bls12"
	df "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/role/pb"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	b58 "github.com/mr-tron/base58/base58"
)

//chalresult 挑战结果在内存中的结构
//作为chalinfo.Time的value 记录单次挑战的各项信息
type chalresult struct {
	kid           string //挑战发起者
	pid           string //挑战对象
	uid           string //挑战的数据所属对象
	challengeTime int64  //挑战发起时间 使用unix时间戳
	sum           uint32 //挑战总空间
	length        uint32 //挑战成功空间
	h             int    //挑战的随机数
	res           bool   //挑战是否成功
	proof         string //挑战结果的证据
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

// 挑战过程的起始函数 定时对本节点连接的provider发起挑战
func challengeRegular(ctx context.Context) { //定期挑战
	log.Println("ChallengeRegular() start!")
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
		thischalinfo := v.(*chalinfo)

		isTestUser := thischalinfo.testuser
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
			maxlength += uint32((utils.MAXOFFSET + 1) * df.DefaultSegmentSize)
			sum += uint32((cInfo.offset + 1) * df.DefaultSegmentSize)
			return true
		})
		//测试User不进行挑战修复
		if isTestUser {
			return true
		}
		thischalresult := &chalresult{ //对provider发起挑战之前，先构造好本次挑战信息的结构体
			challengeTime: challengetime,
			sum:           sum,
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
			log.Println("marshal h failed, err: ", err)
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

//handleProofResultBls12 收到provider返回的挑战结果的回调
//kv格式(uid/"proof"/FaultBlock/chaltime,proof)
func handleProofResultBls12(km *metainfo.KeyMeta, proof, pid string) {
	ops := km.GetOptions()
	Indicesstr := ops[0]
	chaltime := ops[1]
	uid := km.GetMid()
	var h mcl.Challenge
	indices, _ := b58.Decode(Indicesstr)
	splitedindex := strings.Split(string(indices), metainfo.DELIMITER)
	var blocks []string //存放挑战失败的blockid

	for _, index := range splitedindex {
		if index != "" {
			block, _, err := utils.SplitIndex(index)
			if err != nil {
				log.Println("SplitIndex err:", err)
				return
			}
			blocks = append(blocks, block)
		}
	}
	if len(blocks) != 0 {
		log.Println("Fault or NotFound blocks :", blocks)
		reduceCredit(pid)
	}
	pu := PU{
		pid: pid,
		uid: uid,
	}
	challengetime := utils.StringToUnix(chaltime)
	thischalinfo, ok := getChalinfo(pu)
	if !ok {
		log.Println("getChalinfo error!pu: ", pu)
		return
	}
	thischalresult, ok := thischalinfo.Time.Load(challengetime) //获取之前建立好的挑战信息结构
	if !ok {
		log.Println("thischalinfo.Time.Load error!challengetime:", challengetime)
		return
	}
	h.C = thischalresult.(*chalresult).h

	var length uint32
	var offset, electedOffset int
	thischalinfo.Cid.Range(func(k, v interface{}) bool { //记录每个块的挑战结果
		var flag int
		if len(blocks) != 0 { //存在挑战失败的块
			for _, block := range blocks {
				if strings.Compare(k.(string), block) != 0 {
					flag++
					if flag == len(blocks) {
						off := v.(*cidInfo).offset
						if off < 0 {
							return false
						} else if off > 0 {
							electedOffset = h.C % off
						} else {
							electedOffset = 0
						}
						h.Indices = append(h.Indices, k.(string)+metainfo.BLOCK_DELIMITER+strconv.Itoa(electedOffset))
					}
				}
			}
		} else {
			off := v.(*cidInfo).offset
			if off < 0 {
				return false
			} else if off > 0 {
				electedOffset = h.C % off
			} else {
				electedOffset = 0
			}
			h.Indices = append(h.Indices, k.(string)+metainfo.BLOCK_DELIMITER+strconv.Itoa(electedOffset))
		}
		return true
	})
	if len(h.Indices) == 0 {
		return
	}
	pubs, err := getUserBLS12Config(uid)
	if err != nil {
		log.Println("getUserBLS12Config error! uid:", uid)
		return
	}
	res, err := mcl.VerifyProof(pubs.PubKey, h, proof)
	if err != nil {
		log.Println("mcl.VerifyProof err: ", err)
		return
	}
	if res { //验证proof通过后,循环记录每一个块的挑战信息（用于修复），和此次对provider的挑战信息
		//log.Println("verify success cid :", h.Indices)
		//更新thischalinfo.Cid的信息
		for _, tmpindex := range h.Indices {
			blockID, _, _ := utils.SplitIndex(tmpindex)
			if thiscidinfo, ok := thischalinfo.Cid.Load(blockID); ok { //获取当前blockid的offset,若内存中有则直接用，没有就在硬盘中查
				offset = thiscidinfo.(*cidInfo).offset
			} else {
				kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.Local, metainfo.SyncTypeBlock)
				if err != nil {
					log.Println("NewKeyMeta err:", err)
					return
				}
				pidoff, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(kmBlock.ToString(), "")
				if pidoff != nil && err == nil {
					offset, _ = strconv.Atoi((strings.Split(string(pidoff), metainfo.DELIMITER))[1]) //*格式修改
				}
			}
			newcidinfo := &cidInfo{
				res:       true,
				repair:    0,
				availtime: challengetime,
				offset:    offset,
			}
			length += uint32((newcidinfo.offset + 1) * df.DefaultSegmentSize)
			thischalinfo.Cid.Store(blockID, newcidinfo)
		}

		//更新thischalinfo.Time的信息
		thisSum := thischalresult.(*chalresult).sum
		thisH := thischalresult.(*chalresult).h
		newchalresult := &chalresult{
			kid:           localNode.Identity.Pretty(),
			pid:           pid,
			uid:           uid,
			challengeTime: challengetime,
			sum:           thisSum,
			h:             thisH,
			res:           true,
			proof:         proof,
			length:        length,
		}
		//挑战信息验证通过后，同步给组内的其他keeper
		kmChal, err := metainfo.NewKeyMeta(uid, metainfo.Sync, metainfo.SyncTypeChalRes, pid, localNode.Identity.Pretty(), chaltime)
		if err != nil {
			log.Println("metainfo.NewKeyMeta err: ", err)
			return
		}
		metavalue := strings.Join([]string{strconv.Itoa(int(length)), "1", strconv.Itoa(int(thisSum)), strconv.Itoa(int(thisH)), proof}, metainfo.DELIMITER)
		metaSyncTo(kmChal, metavalue)
		thischalinfo.Time.Store(challengetime, newchalresult)
		addCredit(pid)
	} else {
		log.Println("verify failed cid: ", h.Indices)
		reduceCredit(pid)
	}

	thischalinfo.inChallenge = 0

	thischalinfo.tmpCid.Range(func(k, v interface{}) bool {
		act, loaded := thischalinfo.Cid.LoadOrStore(k, v)
		if loaded && act.(*cidInfo).offset < v.(*cidInfo).offset {
			act.(*cidInfo).offset = v.(*cidInfo).offset
			thischalinfo.tmpCid.Delete(k)
			return true
		}
		thischalinfo.maxlength += uint32((utils.MAXOFFSET + 1) * df.DefaultSegmentSize)
		thischalinfo.tmpCid.Delete(k)
		return true
	})

	return
}

//获得用于证明的user的公用参数
func getUserBLS12Config(userID string) (*UserBLS12Config, error) {
	userPubKey := new(mcl.PublicKey)
	userConfig := &UserBLS12Config{}

	userconfigbyte, err := getUserBLS12ConfigByte(userID)
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
