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
