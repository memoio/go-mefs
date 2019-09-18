/*
dht层中我们项目自己添加的代码和改动，其中对于函数的改动记录如下
+ 删除validator相关的操作
+ put到本地时key的编码方法 ds.newkey()和mkDskey()
+ 角色层回调函数 liter append metainfo metainfobroadcast
+ putValue getValue中添加 根据ctx携带信息做不同的操作，putto getfrom append liter 以及初始化的广播操作
*/

package dht

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	proto "github.com/gogo/protobuf/proto"
	u "github.com/ipfs/go-ipfs-util"
	peer "github.com/libp2p/go-libp2p-core/peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	record "github.com/libp2p/go-libp2p-record"
	recpb "github.com/libp2p/go-libp2p-record/pb"
	routing "github.com/libp2p/go-libp2p-routing"
	ds "github.com/memoio/go-mefs/source/go-datastore"
	dsq "github.com/memoio/go-mefs/source/go-datastore/query"
	pb "github.com/memoio/go-mefs/source/go-libp2p-kad-dht/pb"
	"github.com/memoio/go-mefs/utils/metainfo"
)

var (
	errmetahandlerNotAssign = errors.New("MetaMessageHandler not assign")
)

//AssignmetahandlerV2 根据角色为当前节点挂载回调函数
func (dht *IpfsDHT) AssignmetahandlerV2(metahandler metainfo.MetaMessageHandlerV2) error {
	if metahandler == nil {
		return errmetahandlerNotAssign
	}
	dht.metahandlerv2 = metahandler
	return nil
}

//将KV保存在本地，若key已经存在，则将value添加在后面而不是覆盖
func (dht *IpfsDHT) appendLocal(key string, rec *recpb.Record) error {
	logger.Debugf("appendLocal: %v %v", key, rec)
	reclocal, _ := dht.getLocal(key) //检查本地数据
	if reclocal == nil {             //本地数据为空  直接添加
		return dht.putLocal(key, rec)
	}
	if reclocal != nil { //本地数据不为空
		appendValue := rec.GetValue()
		oldValue := reclocal.GetValue()
		isExisted := false
		index := bytes.Index(oldValue, appendValue) //检查本地中是否已包含要添加的
		if index != -1 {
			isExisted = true
		}
		if isExisted == false { //若输入数据和本地数据不一样，则将数据添加在原数据之后
			var buffer bytes.Buffer
			buffer.Write(oldValue)
			buffer.Write(appendValue)
			rec2 := record.MakePutRecord(key, buffer.Bytes())
			rec2.TimeReceived = rec.TimeReceived //这里注意记录数据时间
			return dht.putLocal(key, rec2)
		}
	}
	return nil
}

//从本地对给定前缀进行迭代查询,查询的结果用‘ ’连接保存在record结构中返回
func (dht *IpfsDHT) literLocal(prefix string) (*recpb.Record, error) {
	var keylist [][]byte
	var valuelist [][]byte
	recout := new(recpb.Record)
	key := ds.NewKey(prefix).String() //去掉base32编码
	q := dsq.Query{Prefix: key}
	qr, _ := dht.datastore.Query(q) //进行查询操作
	es, _ := qr.Rest()

	for _, e := range es {
		rec := new(recpb.Record)
		proto.Unmarshal(e.Value, rec) //解出rec 一条记录的kv
		keylist = append(keylist, rec.Key)
		valuelist = append(valuelist, rec.Value) //将这条记录的kv添加进kv表中
	}
	if es != nil { //将kv表合并放进record中
		recout.Key = bytes.Join(keylist, []byte(" "))
		recout.Value = bytes.Join(valuelist, []byte(" "))
	} else {
		recout = nil
	}
	return recout, nil
}

//putLocal操作的便利版本，传入参数为字符串KV
func (dht *IpfsDHT) putKVLocal(key string, value string) error {
	recTemp := new(recpb.Record)
	recTemp.Key = []byte(key)
	recTemp.Value = []byte(value)
	recTemp.TimeReceived = u.FormatRFC3339(time.Now())
	return dht.putLocal(key, recTemp)
}

//SendMetaRequest 向其他节点发送meta信息，需要等待对方节点的回复,当回复信息为空的时候，报错
// 传入参数为规定格式的KV对，调用该函数的上层函数名，和对方节点的peerid
func (dht *IpfsDHT) SendMetaRequest(metaKey, metaValue, peerID, caller string) (string, error) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	rec := record.MakePutRecord(metaKey, []byte(metaValue))
	pmes := pb.NewMessage(pb.Message_MetaInfo, []byte(caller), 0)
	pmes.Record = rec //构建用于发送的pmes
	p, _ := peer.IDB58Decode(peerID)
	rpmes, err := dht.sendRequest(ctx, p, pmes) //得到返回信息
	if err != nil {
		return "", err
	}
	response := rpmes.Record.GetValue()
	if response != nil {
		return string(response), nil
	}
	return "", errInvalidRecord
}

//SendMetaMessage 向其他节点发送meta信息，与MetaRequest不同的是，不需要对方节点回复
func (dht *IpfsDHT) SendMetaMessage(metaKey, metaValue, peerID, caller string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	rec := record.MakePutRecord(metaKey, []byte(metaValue))
	pmes := pb.NewMessage(pb.Message_MetaInfo, []byte(caller), 0)
	pmes.Record = rec //构建用于发送的pmes
	p, _ := peer.IDB58Decode(peerID)
	return dht.sendMessage(ctx, p, pmes)
}

//BroadcastMetaMessage 广播meta信息
func (dht *IpfsDHT) BroadcastMetaMessage(metaKey, metaValue, caller string, count int) {
	rec := record.MakePutRecord(metaKey, []byte(metaValue))
	pmes := pb.NewMessage(pb.Message_MetaInfo, []byte(caller), 0)
	pmes.Record = rec //构建用于发送的pmes
}

//============dht层的直接调用函数==============

//在本地的DB里删除Key
func (dht *IpfsDHT) DeleteLocal(key string) error {
	return dht.datastore.Delete(ds.NewKey(key))
}

//CmdPutTo 将kv对发送到指定节点，id为local时，保存在本地
func (dht *IpfsDHT) CmdPutTo(key, value, id string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	valueByte := []byte(value)
	rec := record.MakePutRecord(key, valueByte)
	rec.TimeReceived = u.FormatRFC3339(time.Now())

	if strings.Compare(id, "local") == 0 {
		return dht.putLocal(key, rec)
	}
	pmes := pb.NewMessage(pb.Message_PUT_VALUE, rec.Key, 0)
	pmes.Record = rec
	p, _ := peer.IDB58Decode(id)
	_, err := dht.sendRequest(ctx, p, pmes)
	if err != nil {
		fmt.Println("sendRequest to: ", p.Pretty(), ", err: ", err)
	}
	return err
}

//CmdAppendTo 将传入value添加到现有value之后，和put不同的是对于相同的key，put为覆盖操作，append为添加
func (dht *IpfsDHT) CmdAppendTo(key, value, id string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	valueByte := []byte(value)
	rec := record.MakePutRecord(key, valueByte)
	rec.TimeReceived = u.FormatRFC3339(time.Now())

	if strings.Compare(id, "local") == 0 {
		return dht.appendLocal(key, rec)
	}
	pmes := pb.NewMessage(pb.Message_APPEND_VALUE, rec.Key, 0)
	pmes.Record = rec
	p, _ := peer.IDB58Decode(id)
	_, err := dht.sendRequest(ctx, p, pmes)

	return err
}

//CmdGetFrom 从指定节点获得kv对 id为local时，在本地查找
func (dht *IpfsDHT) CmdGetFrom(key, id string) ([]byte, error) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	//全网查找
	if id == "" {
		return dht.GetValue(ctx, key)
	}
	//本地查找
	if strings.Compare(id, "local") == 0 {
		rec, err := dht.getLocal(key)
		if err != nil {
			return nil, err
		}
		keyrec := string(rec.GetKey())
		if strings.Compare(keyrec, key) == 0 {
			return rec.GetValue(), nil
		}
		return nil, routing.ErrNotFound
	}
	//指定节点查找
	p, err := peer.IDB58Decode(id)
	if err != nil {
		return nil, err
	}
	pmes := pb.NewMessage(pb.Message_GET_VALUE, []byte(key), 0)
	resp, err := dht.sendRequest(ctx, p, pmes) //发送请求得到结果
	if err != nil {
		fmt.Println("dht.sendRequest-err", err)
	}
	rec := resp.GetRecord()
	keyrec := string(rec.GetKey())
	if strings.Compare(keyrec, key) == 0 {
		return rec.GetValue(), nil
	}
	return nil, routing.ErrNotFound

}

//CmdLiterFrom 指定节点前缀查找
func (dht *IpfsDHT) CmdLiterFrom(prefix, id string) ([]byte, error) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var out string
	var err error
	recout := new(recpb.Record)
	//本地查找
	if strings.Compare(id, "local") == 0 {
		recout, err = dht.literLocal(prefix)
		if err != nil {
			out = err.Error()
		}
	} else {
		p, _ := peer.IDB58Decode(id)
		pmes := pb.NewMessage(pb.Message_GET_PREFIX, []byte(prefix), 0)
		resp, _ := dht.sendRequest(ctx, p, pmes) //发送请求得到结果
		recout = resp.GetRecord()                //解出查询结果
	}
	if recout != nil { //从查询返回的record中解出信息
		keylist := bytes.Split(recout.GetKey(), []byte(" "))
		valuelist := bytes.Split(recout.GetValue(), []byte(" "))
		for i, k := range keylist {
			out += fmt.Sprintf("key:%s\tvalue:%s\n", string(k), string(valuelist[i])) //构造输出
		}
	} else {
		return nil, errors.New("Not found")
	}
	return []byte(out), nil
}

//==================角色层metakey回调函数==================
// Meta操作的回调，解出kv对，携带发送方peerid一起传入角色层进行操作，操作完成后，返回信息,由对方决定是否等待返回信息
func (dht *IpfsDHT) handleMetaInfo(ctx context.Context, p peer.ID, pmes *pb.Message) (_ *pb.Message, err error) {
	ctx = logger.Start(ctx, "handleMetaRequest")
	logger.SetTag(ctx, "peer", p)
	defer func() { logger.FinishWithErr(ctx, err) }() //各项记录操作

	rpmes := pmes
	rec := pmes.GetRecord() //获得record信息
	if rec == nil {
		logger.Infof("Got nil record from: %s", p.Pretty())
		return nil, errors.New("nil record")
	}
	metaKey := string(rec.GetKey())
	metaValue := string(rec.GetValue())

	if dht.metahandlerv2 == nil {
		return nil, metainfo.ErrMetaHandlerNotAssign
	}
	res, err := dht.metahandlerv2.HandleMetaMessage(metaKey, metaValue, p.Pretty())
	if err != nil {
		fmt.Printf("handleMetaInfo()err:%s\nmetakey:%s\nfrom:%s\ncaller:%s\n", err, metaKey, p.Pretty(), string(pmes.GetKey()))
	}
	rpmes.Record.Value = []byte(res) //role层回调函数的返回值放在返回信息中，一般会返回"complete"
	return rpmes, err
}

//广播的Meta操作的回调，在操作最后，会在返回信息中添加与本节点相连的最近节点
func (dht *IpfsDHT) handleMetaBroadcast(ctx context.Context, p peer.ID, pmes *pb.Message) (_ *pb.Message, err error) {
	ctx = logger.Start(ctx, "handleMetaRequest")
	logger.SetTag(ctx, "peer", p)
	defer func() { logger.FinishWithErr(ctx, err) }() //各项记录操作

	rpmes := pmes
	rec := pmes.GetRecord() //获得record信息
	if rec == nil {
		logger.Infof("Got nil record from: %s", p.Pretty())
		return nil, errors.New("nil record")
	}
	metaKey := string(rec.GetKey())
	metaValue := string(rec.GetValue())

	if dht.metahandlerv2 == nil {
		return nil, metainfo.ErrMetaHandlerNotAssign
	}
	res, err := dht.metahandlerv2.HandleMetaMessage(metaKey, metaValue, p.Pretty())
	if err != nil {
		fmt.Printf("handleMetaBroadcast()err:%s\nmetakey:%s\nfrom:%s\ncaller:%s\n", err, metaKey, p.Pretty(), string(pmes.GetKey()))
	}
	rpmes.Record.Value = []byte(res)

	// Find closest peer on given cluster to desired key and reply with that info
	closer := dht.betterPeersToQuery(pmes, p, CloserPeerCount)
	if len(closer) > 0 {
		closerinfos := pstore.PeerInfos(dht.peerstore, closer)
		for _, pi := range closerinfos {
			logger.Debugf("handleGetValue returning closer peer: '%s'", pi.ID)
			if len(pi.Addrs) < 1 {
				logger.Warningf(`no addresses on peer being sent!
						[local:%s]
						[sending:%s]
						[remote:%s]`, dht.self, pi.ID, p)
			}
		}

		rpmes.CloserPeers = pb.PeerInfosToPBPeers(dht.host.Network(), closerinfos)
	}

	return rpmes, err
}

func (dht *IpfsDHT) handleLiter(ctx context.Context, p peer.ID, pmes *pb.Message) (_ *pb.Message, err error) {
	//接受信息节点的前缀查询回调函数
	ctx = logger.Start(ctx, "handleLiter")
	logger.SetTag(ctx, "peer", p)
	defer func() { logger.FinishWithErr(ctx, err) }() //各项记录操作

	resp := pb.NewMessage(pmes.GetType(), pmes.GetKey(), pmes.GetClusterLevel()) //做一个答复信息

	prefix := pmes.GetKey() //从pmes中解出数据
	if prefix == nil {
		logger.Infof("Got nil record from: %s", p.Pretty())
		return nil, errors.New("nil record")
	}

	resp.Record, _ = dht.literLocal(string(prefix))

	return resp, err
}

func (dht *IpfsDHT) handleAppendValue(ctx context.Context, p peer.ID, pmes *pb.Message) (_ *pb.Message, err error) {
	//append操作的回调函数，收到信息，调用dht层appendLocal函数，将value添加在原value之后
	ctx = logger.Start(ctx, "handleAppendValue")
	logger.SetTag(ctx, "peer", p)
	defer func() { logger.FinishWithErr(ctx, err) }() //各项记录操作

	rec := pmes.GetRecord()
	if rec == nil {
		logger.Infof("Got nil record from: %s", p.Pretty())
		return nil, errors.New("nil record")
	}
	rec.TimeReceived = u.FormatRFC3339(time.Now())   //记录时间
	err = dht.appendLocal(string(rec.GetKey()), rec) //这里直接调用接受节点的appendlocal函数

	return pmes, err
}

func (dht *IpfsDHT) bootstrapRegular() {
	fmt.Println("dht bootstrap reqular start!")

	ctx := dht.ctx

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go dht.Bootstrap(ctx)
		}
	}

}
