package keeper

import (
	"errors"
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

var (
	errLocalKeeper     = errors.New("Not send to self")
	errorWrongSyncType = errors.New("miss match synctype")
)

// metaSyncTo 元数据同步调用函数
//传入metainfo结构体keyMeta和metavalue，查出需要同步的keeperid 进行发送，targets为可变参数，若传入target，则只向设置的target同步
func metaSyncTo(keyMeta *metainfo.KeyMeta, metaValue string, targets ...string) {
	var err error
	mainID := keyMeta.GetMid()
	options := keyMeta.GetOptions()
	if len(options) < 1 {
		return
	}

	// sync block meta seems to too heavily
	if options[0] == metainfo.SyncTypeBlock {
		return
	}

	if len(targets) == 0 { //获取同步对象
		targets, err = getTarget(mainID, options[0])
		if err != nil {
			log.Println(err)
			return
		}
	}

	keyMeta.SetKeyType(metainfo.Sync)

	for _, p := range targets {
		_, err := sendMetaRequest(keyMeta, metaValue, p)
		if err != nil {
			log.Println(err)
			return
		}
	}

}

//根据传入的key 得到user信息，找到相关的keeper返回
func getTarget(mainID, syncType string) ([]string, error) {
	var target []string
	var user string
	//从不同的元数据中取出userid用于寻找同步的keeper
	switch syncType {
	case metainfo.SyncTypeBlock: //(uid_gid_sid_bid
		bm, err := metainfo.GetBlockMeta(mainID)
		if err != nil {
			return nil, err
		}
		user = bm.GetUid()
	case metainfo.SyncTypeChalRes, metainfo.SyncTypeChalPay, metainfo.SyncTypeTInfo: //uid
		user = mainID
	case metainfo.SyncTypePid:
		user = mainID
	default:
		return nil, errorWrongSyncType
	}
	//target中去掉本节点
	if strings.Compare(user, localNode.Identity.Pretty()) == 0 {
		return nil, errLocalKeeper
	}
	thisgroupInfo, ok := getGroupsInfo(user)
	if !ok {
		return nil, errNoGroupsInfo
	}
	for _, keeperID := range thisgroupInfo.keepers {
		if strings.Compare(keeperID, localNode.Identity.Pretty()) != 0 {
			target = append(target, keeperID)
		}
	}
	return target, nil
}

//该函数用于将master发来的支付信息，构造两份信息，一份为最近一次支付结果保存在本地，一份为支付信息，保存在内存中和本地
//(groupid/"sync"/"chalpay"/pid/beginTime/endTime, spacetime/signature/proof)
func handleSyncChalPay(km *metainfo.KeyMeta, metaValue string) error {
	groupid := km.GetMid()
	options := km.GetOptions()
	if len(options) < 4 {
		return metainfo.ErrIllegalKey
	}
	splitedMetaValue := strings.Split(metaValue, metainfo.DELIMITER)
	if len(splitedMetaValue) < 3 {
		return metainfo.ErrIllegalValue
	}
	pidString := options[1]
	beginTime := utils.StringToUnix(options[2])
	endTime := utils.StringToUnix(options[3])
	st, ok := big.NewInt(0).SetString(splitedMetaValue[0], 0)
	if !ok {
		log.Println("SetString()err!value: ", splitedMetaValue[0])
		return metainfo.ErrIllegalValue
	}

	thisPU := puKey{
		uid: groupid,
		pid: pidString,
	}

	//保存此次的支付结果
	_, _, err := saveLastPay(thisPU, "signature", "proof", beginTime, endTime, st)
	if err != nil {
		return err
	}
	return nil
}

// syncProof 收到单次挑战信息同步的操作，保存在内存和硬盘中
// uid/"sync"/"chalres"/pid/kid/time,length/result/proof/sum/h
func handleSyncChalres(km *metainfo.KeyMeta, metaValue string) error {
	groupid := km.GetMid()
	options := km.GetOptions()
	if len(options) < 4 {
		return metainfo.ErrIllegalKey
	}
	splitedMetaValue := strings.Split(metaValue, metainfo.DELIMITER)
	if len(splitedMetaValue) < 7 {
		return metainfo.ErrIllegalValue
	}
	timerec := utils.StringToUnix(options[3])   //转换收到的时间信息格式
	l, err := strconv.Atoi(splitedMetaValue[0]) //转换长度信息格式
	if err != nil {
		return err
	}
	res := splitedMetaValue[1]
	chalres := true //转换挑战结果信息格式
	if res == "0" {
		chalres = false
	}
	thisSum, err := strconv.Atoi(splitedMetaValue[2])
	if err != nil {
		return err
	}
	thisH, err := strconv.Atoi(splitedMetaValue[3])
	if err != nil {
		return err
	}
	proofStr := strings.Join(splitedMetaValue[4:], metainfo.DELIMITER)
	thischalresult := &chalresult{ //构建挑战结果
		kid:           options[2],
		pid:           options[1],
		uid:           groupid,
		challengeTime: timerec,
		sum:           int64(thisSum),
		h:             thisH,
		res:           chalres,
		proof:         proofStr,
		length:        int64(l),
	}

	pu := puKey{
		pid: thischalresult.pid,
		uid: thischalresult.uid,
	}
	thischalinfo, ok := getChalinfo(pu)
	if !ok {
		return errNoChalInfo
	}
	thischalinfo.chalMap.Store(timerec, thischalresult) //放到LedgerInfo里
	return nil
}

//syncBlock 收到数据块信息的同步操作
// blockID/"sync"/"block",pid/length
func handleSyncBlock(km *metainfo.KeyMeta, metaValue string) error {
	km.SetKeyType(metainfo.Local)
	splitedMetaValue := strings.Split(metaValue, metainfo.DELIMITER)
	if len(splitedMetaValue) < 2 {
		return metainfo.ErrIllegalValue
	}
	localValueByte, _ := getKeyFrom(km.ToString(), "local")
	localValueString := string(localValueByte) //从本地取数据
	if localValueString != "" {                //如果本地有同样数据，判断是否覆盖原有数据
		splitedLocalValue := strings.Split(localValueString, metainfo.DELIMITER)
		newLength, err := strconv.Atoi(splitedMetaValue[1])
		if err != nil { //如果新数据有问题，直接返回
			return err
		}
		localLength, err := strconv.Atoi(splitedLocalValue[1])
		if err == nil { //本地数据没问题
			if localLength > newLength { //对比传入数据和本地数据 value中的长度项 若传入长度短，则说明数据较旧，直接返回
				return nil
			}
		}

	}
	blockID := km.GetMid()
	uid := blockID[:utils.IDLength]
	pid := splitedMetaValue[0]
	offset, err := strconv.Atoi(splitedMetaValue[1])
	if err != nil {
		return err
	}

	addBlocktoMem(uid, pid, blockID, offset)

	err = putKeyTo(km.ToString(), metaValue, "local") //最后，保存数据到本地
	if err != nil {
		return err
	}
	return nil
}

// syncIDs 收到节点的 U-K-P 信息的同步操作 这个就直接保存在本地
// 以provider信息为例：peerID/"sync"/"pids"，pid1pid2pid3......
func syncKUPIDs(km *metainfo.KeyMeta, pids string) error {
	km.SetKeyType(metainfo.Local)
	metaKey := km.ToString()
	err := putKeyTo(metaKey, pids, "local")
	if err != nil {
		return err
	}
	return nil
}
