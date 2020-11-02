package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/memoio/go-mefs/contracts"
	df "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/pb"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/test"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	shell "github.com/memoio/mefs-go-http-client"
)

//随机文件最大大小
const randomDataSize = 1024 * 1024 * 10
const bucketName = "Bucket01"
const dataCount = 3
const parityCount = 2

const moneyTo = 6000000000000000000

var ethEndPoint, qethEndPoint string

func main() {
	eth := flag.String("eth", "http://212.64.28.207:8101", "eth api address for set;")
	qeth := flag.String("qeth", "http://39.100.146.165:8101", "eth api address for query;")

	flag.Parse()
	ethEndPoint = *eth
	qethEndPoint = *qeth

	contracts.EndPoint = ethEndPoint

	if err := challengeTest(); err != nil {
		log.Fatal(err)
	}
}

func challengeTest() error {
	sh := shell.NewShell("localhost:5001")
	testuser, err := sh.CreateUser()
	if err != nil {
		log.Println("Create user failed :", err)
		return err
	}
	addr := testuser.Address
	uid, err := address.GetIDFromAddress(addr)
	if err != nil {
		log.Println("address to id failed")
		return err
	}

	flag := true
	if flag {
		err := test.TransferTo(big.NewInt(moneyTo), addr, ethEndPoint, qethEndPoint)
		if err != nil {
			log.Println("transfer fails: ", err)
			return err
		}
	}

	//set ks:3, ps:6
	var startOpts []func(*shell.RequestBuilder) error
	startOpts = append(startOpts, shell.SetOp("ks", "3"))
	startOpts = append(startOpts, shell.SetOp("ps", "6"))
	err = sh.StartUser(addr, startOpts...)
	if err != nil {
		log.Println("Start user failed :", err)
		return err
	}

	log.Println(addr, "started, begin to upload")

	_, err = sh.ShowStorage(shell.SetAddress(addr))
	if err != nil {
		log.Println(addr, " show storage fail: ", err)
		return err
	}

	var opts []func(*shell.RequestBuilder) error
	//设置某些选项
	opts = append(opts, shell.SetAddress(addr))
	opts = append(opts, shell.SetDataCount(dataCount))
	opts = append(opts, shell.SetParityCount(parityCount))
	opts = append(opts, shell.SetPolicy(df.RsPolicy))
	bk, err := sh.CreateBucket(bucketName, opts...)
	if err != nil {
		log.Println(addr, " create bucket fails: ", err)
		return err
	}

	log.Println(bk, "addr:", addr)

	//然后开始上传文件
	r := rand.Int63n(randomDataSize)
	data := make([]byte, r)
	fillRandom(data)
	buf := bytes.NewBuffer(data)
	objectName := addr + "_" + strconv.Itoa(int(r))
	log.Println("Begin to upload", objectName, "Size is", ToStorageSize(r), "addr", addr)
	beginTime := time.Now().Unix()
	ob, err := sh.PutObject(buf, objectName, bucketName, shell.SetAddress(addr))
	if err != nil {
		log.Println(addr, "Upload failed", err)
		return err
	}
	storagekb := float64(r) / 1024.0
	endTime := time.Now().Unix()
	speed := storagekb / float64(endTime-beginTime)
	log.Println("Upload", objectName, "Size is", ToStorageSize(r), "speed is", speed, "KB/s", " addr", addr)
	log.Println(ob.String() + "address: " + addr)

	getOb, err := sh.ListObjects(bucketName, shell.SetAddress(addr), shell.SetAvailTime(true))
	if err != nil {
		log.Println("List Objects failed :", err)
		return err
	}
	log.Println("Object Name :", getOb.Objects[0].Name, "\nObject LastChallenge Time :", getOb.Objects[0].LatestChalTime)
	lastChallengeTime := getOb.Objects[0].LatestChalTime

	keepers, err := sh.ListKeepers(shell.SetAddress(addr))
	if err != nil {
		log.Println("list keepers error :", err)
		return err
	}

	qid := uid
	if flag {
		log.Println("conatracts has endpoint: ", contracts.EndPoint)
		qItem, err := role.GetLatestQuery(uid)
		if err != nil {
			log.Println("got query from", uid, "fails: ", err)
			return err
		}
		qid = qItem.QueryID
	}

	cid := make([]string, 2)
	pro := make([]string, 2)
	for i := 0; i < 2; i++ {
		bm, err := metainfo.NewBlockMeta(qid, "1", "0", strconv.Itoa(i))
		if err != nil {
			log.Println(err)
			return err
		}

		cid[i] = bm.ToString()
		kmBlock, err := metainfo.NewKey(cid[i], mpb.KeyType_BlockPos)
		if err != nil {
			log.Println(err)
			return err
		}
		blockMeta := kmBlock.ToString()
		keeper := keepers.Peers[0].PeerID
		log.Println("got blockMeta: ", blockMeta, " from: ", keeper)
		var provider string
		retry := 0
		for retry < 5 {
			keeper = keepers.Peers[0].PeerID
			resPid, err := sh.GetFrom(blockMeta, keeper)
			if err == nil {
				provider = strings.Split(resPid.Extra, metainfo.DELIMITER)[0]
				pro[i] = provider
				log.Println("provider is: ", provider)
				break
			} else {
				keeper = keepers.Peers[1].PeerID
				resPid, err := sh.GetFrom(blockMeta, keeper)
				if err == nil {
					provider = strings.Split(resPid.Extra, metainfo.DELIMITER)[0]
					log.Println("provider is: ", provider)
					pro[i] = provider
					break
				}
			}
			retry++
		}
	}

	if len(pro) != 2 || pro[0] == "" || pro[1] == "" {
		log.Fatal("cannot get block pos: ", pro[0], " and ", pro[1])
	}

	ret, err := getBlock(sh, cid[0], pro[0]) //获取块的MD5
	if err != nil || ret == "" {
		log.Fatal("get block from old provider error: ", err)
		return err
	}
	log.Println("md5 of block`s rawdata :", ret)

	for i := 0; i < 2; i++ {
		//在provider上删除指定块
		km, err := metainfo.NewKey(cid[i], mpb.KeyType_Block)
		if err != nil {
			log.Fatal("construct del block KV error :", err)
			return err
		}

		_, err = sh.DeleteFrom(km.ToString(), pro[i])
		if err != nil {
			log.Fatal("run dht delete error :", err)
			return err
		}
	}

	time.Sleep(1 * time.Minute)

	nret, err := getBlock(sh, cid[0], pro[0]) //获取块的MD5
	if nret != "" && err == nil {
		log.Fatal("get block from provider: ", pro[0], " expcted not")
		return err
	}

	log.Println("successfully delete block :", cid[0], " in provider: ", pro[0])

	// read whole file again
	outer, err := sh.GetObject(objectName, bucketName, shell.SetAddress(addr))
	if err != nil {
		log.Fatal("download file ", objectName, " err:", err)
		return err
	}

	obuf := new(bytes.Buffer)
	obuf.ReadFrom(outer)
	if obuf.Len() != int(r) {
		log.Fatal("download file ", objectName, "failed, got: ", obuf.Len(), "expected: ", r)
	}

	obj, err := sh.HeadObject(objectName, bucketName, shell.SetAddress(addr))
	if err != nil {
		log.Fatal("head file ", objectName, " err:", err)
		return err
	}

	gotTag := md5.Sum(obuf.Bytes())
	if hex.EncodeToString(gotTag[:]) != obj.Objects[0].MD5 {
		log.Fatal("download file ", objectName, "failed, got md5: ", hex.EncodeToString(gotTag[:]), "expected: ", obj.Objects[0].MD5)
	}

	log.Println("successfully get object :", objectName, " in bucket:", bucketName)

	time.Sleep(60 * time.Minute)

	getOb, err = sh.ListObjects(bucketName, shell.SetAddress(addr), shell.SetAvailTime(true))
	if err != nil {
		log.Println("List Objects failed :", err)
		return err
	}
	log.Println("Object Name :", getOb.Objects[0].Name, "\nObject LastChallenge Time :", getOb.Objects[0].LatestChalTime)

	if strings.Compare(lastChallengeTime, getOb.Objects[0].LatestChalTime) == 0 {
		log.Println("Challenge time not change")
		return errors.New("ChallengeTest failed, Last challenge time not change")
	}

	//获取新的provider，从新的provider上获得块的MD5
	var newProvider string
	kmBlock, err := metainfo.NewKey(cid[0], mpb.KeyType_BlockPos)
	if err != nil {
		log.Println(err)
		return err
	}

	blockMeta := kmBlock.ToString()
	retry := 0
	for retry < 5 {
		keeper := keepers.Peers[0].PeerID
		resPid, err := sh.GetFrom(blockMeta, keeper)
		if err == nil {
			resPro := strings.Split(resPid.Extra, metainfo.DELIMITER)
			if len(resPro) == 2 {
				newProvider = resPro[0]
				break
			}
		}

		keeper = keepers.Peers[1].PeerID
		resPid, err = sh.GetFrom(blockMeta, keeper)
		if err == nil {
			resPro := strings.Split(resPid.Extra, metainfo.DELIMITER)
			if len(resPro) == 2 {
				newProvider = resPro[0]

				break
			}
		}

		retry++
	}

	if len(newProvider) == 0 {
		log.Fatal("cannot get block pos")
	}

	log.Println("new provider is: ", newProvider)
	newRet, err := getBlock(sh, cid[0], newProvider)
	if err != nil || newRet == "" {
		log.Fatal("get block from new provider error :", err)
		return err
	}

	log.Println("md5 of repaired block rawdata :", newRet)
	if ret == newRet {
		log.Println("Repair success")
	} else {
		log.Fatal("old and new block`s rawdata md5 not match")
		return errors.New("repair failed")
	}
	return nil
}

func ToStorageSize(r int64) string {
	FloatStorage := float64(r)
	var OutStorage string
	if FloatStorage < 1024 && FloatStorage >= 0 {
		OutStorage = fmt.Sprintf("%.2f", FloatStorage) + "B"
	} else if FloatStorage < 1048576 && FloatStorage >= 1024 {
		OutStorage = fmt.Sprintf("%.2f", FloatStorage/1024) + "KB"
	} else if FloatStorage < 1073741824 && FloatStorage >= 1048576 {
		OutStorage = fmt.Sprintf("%.2f", FloatStorage/1048576) + "MB"
	} else {
		OutStorage = fmt.Sprintf("%.2f", FloatStorage/1073741824) + "GB"
	}
	return OutStorage
}

func fillRandom(p []byte) {
	for i := 0; i < len(p); i += 7 {
		val := rand.Int63()
		for j := 0; i+j < len(p) && j < 7; j++ {
			p[i+j] = byte(val)
			val >>= 8
		}
	}
}

func getBlock(sh *shell.Shell, cid, provider string) (string, error) {
	i := 0
	for i < 10 {
		ret, err := sh.GetBlockFrom(cid, provider)
		if err == nil && ret != "" {
			log.Println("Getblock success in ", i+1, " try")
			return ret, nil
		}
		i++
	}
	return "", errors.New("Tried Too Many Times")
}
