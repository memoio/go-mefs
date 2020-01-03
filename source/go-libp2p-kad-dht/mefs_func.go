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
	"log"
	"strings"

	proto "github.com/gogo/protobuf/proto"
	peer "github.com/libp2p/go-libp2p-core/peer"
	routing "github.com/libp2p/go-libp2p-routing"
	ds "github.com/memoio/go-mefs/source/go-datastore"
	dsq "github.com/memoio/go-mefs/source/go-datastore/query"
	pb "github.com/memoio/go-mefs/source/go-libp2p-kad-dht/pb"
	"github.com/memoio/go-mefs/source/instance"
)

var (
	errmetahandlerNotAssign = errors.New("MetaMessageHandler not assign")
)

//AssignmetahandlerV2 根据角色为当前节点挂载回调函数
func (dht *KadDHT) AssignmetahandlerV2(metahandler instance.Service) error {
	if metahandler == nil {
		return errmetahandlerNotAssign
	}
	dht.metahandler = metahandler
	return nil
}

//将KV保存在本地，若key已经存在，则将value添加在后面而不是覆盖
func (dht *KadDHT) appendLocal(key string, rec *pb.Record) error {
	logger.Debugf("appendLocal: %v %v", key, rec)
	reclocal, _ := dht.getLocal(key) //检查本地数据

	if reclocal != nil { //本地数据不为空
		appendValue := rec.GetValue()
		oldValue := reclocal.GetValue()
		index := bytes.Index(oldValue, appendValue) //检查本地中是否已包含要添加的
		if index >= 0 {
			var buffer bytes.Buffer
			buffer.Write(oldValue)
			buffer.Write(appendValue)
			rec2 := MakePutRecord(key, buffer.Bytes())
			return dht.putLocal(key, rec2)
		}
	}
	return dht.putLocal(key, rec)
}

//SendRequest 向其他节点发送meta信息，需要等待对方节点的回复,当回复信息为空的时候，报错
// 传入参数为规定格式的KV对，调用该函数的上层函数名，和对方节点的peerid
func (dht *KadDHT) SendRequest(ctx context.Context, typ int32, metaKey string, metaValue, sig []byte, p peer.ID) ([]byte, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rec := MakePutRecord(metaKey, metaValue)

	pmes := pb.NewMessage(pb.Message_MetaInfo, []byte(metaKey), 0)
	pmes.OpType = typ

	if sig != nil {
		rec.Signature = sig
		pmes.Record = rec
	} else {
		if metaValue != nil {
			pmes.Record = rec
		}
	}

	rpmes, err := dht.sendRequest(ctx, p, pmes) //得到返回信息
	if err != nil {
		log.Println("Send metainfo error: ", err)
		return nil, err
	}
	response := rpmes.Record.GetValue()
	if response != nil {
		return response, nil
	}
	return nil, errInvalidRecord
}

//SendMessage need no reply
func (dht *KadDHT) SendMessage(ctx context.Context, typ int32, metaKey string, metaValue, sig []byte, p peer.ID) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rec := MakePutRecord(metaKey, metaValue)

	pmes := pb.NewMessage(pb.Message_MetaInfo, rec.GetKey(), 0)
	pmes.OpType = typ

	if sig != nil {
		rec.Signature = sig
		pmes.Record = rec
	} else {
		if metaValue != nil {
			pmes.Record = rec
		}
	}

	return dht.sendMessage(ctx, p, pmes)
}

//============dht层的直接调用函数==============

//PutTo 将kv对发送到指定节点，id为local时，保存在本地
func (dht *KadDHT) PutTo(ctx context.Context, key string, value []byte, id string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rec := MakePutRecord(key, value)

	if id == "local" {
		dht.putLocal(key, rec)
	}

	pmes := pb.NewMessage(pb.Message_PUT_VALUE, rec.Key, 0)
	pmes.Record = rec
	p, err := peer.IDB58Decode(id)
	if err != nil {
		return err
	}
	rpmes, err := dht.sendRequest(ctx, p, pmes)
	if err != nil {
		log.Println("sendRequest to: ", p.Pretty(), ", err: ", err)
		return err
	}

	if !bytes.Equal(rpmes.GetRecord().GetValue(), rec.GetValue()) {
		log.Println("PutTo: value not put correctly. (%v != %v)", pmes, rpmes)
		return errors.New("value not put correctly")
	}
	return nil
}

//GetFrom 从指定节点获得kv对 id为local时，在本地查找
func (dht *KadDHT) GetFrom(ctx context.Context, key string, to string) ([]byte, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	//全网查找
	if to == "" {
		return dht.GetValue(ctx, key)
	}
	//本地查找
	if strings.Compare(to, "local") == 0 {
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
	p, err := peer.IDB58Decode(to)
	if err != nil {
		return nil, err
	}
	pmes := pb.NewMessage(pb.Message_GET_VALUE, []byte(key), 0)
	rpmes, err := dht.sendRequest(ctx, p, pmes) //发送请求得到结果
	if err != nil {
		log.Println("dht.sendRequest-err", err)
		return nil, err
	}

	if !bytes.Equal(rpmes.GetRecord().GetKey(), []byte(key)) {
		return nil, routing.ErrNotFound
	}

	return rpmes.GetRecord().GetValue(), nil
}

//IterFrom 指定节点前缀查找
func (dht *KadDHT) IterFrom(ctx context.Context, prefix, id string) ([]byte, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var out string
	var err error
	recout := new(pb.Record)
	//本地查找
	if strings.Compare(id, "local") == 0 {
		recout, err = dht.iterLocal(prefix)
		if err != nil {
			out = err.Error()
		}
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

//从本地对给定前缀进行迭代查询,查询的结果用‘ ’连接保存在record结构中返回
func (dht *KadDHT) iterLocal(prefix string) (*pb.Record, error) {
	var keylist [][]byte
	var valuelist [][]byte
	recout := new(pb.Record)
	key := ds.NewKey(prefix).String() //去掉base32编码
	q := dsq.Query{Prefix: key}
	qr, _ := dht.datastore.Query(q) //进行查询操作
	es, _ := qr.Rest()

	for _, e := range es {
		rec := new(pb.Record)
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

//==================角色层metakey回调函数==================
// Meta操作的回调，解出kv对，携带发送方peerid一起传入角色层进行操作，操作完成后，返回信息,由对方决定是否等待返回信息
func (dht *KadDHT) handleMetaInfo(ctx context.Context, p peer.ID, pmes *pb.Message) (_ *pb.Message, err error) {
	ctx = logger.Start(ctx, "handleMetaRequest")
	logger.SetTag(ctx, "peer", p)
	defer func() { logger.FinishWithErr(ctx, err) }() //各项记录操作

	rpmes := pmes
	rec := pmes.GetRecord() //获得record信息
	metaKey := string(pmes.GetKey())
	metaValue := rec.GetValue()

	// append/iter can be done here

	if dht.metahandler == nil {
		log.Println("No MetaHandler")
		return nil, instance.ErrMetaHandlerNotAssign
	}

	log.Println("hanle metakey:", metaKey)

	res, err := dht.metahandler.HandleMetaMessage(int(rpmes.GetOpType()), metaKey, metaValue, p.Pretty())
	if err != nil {
		log.Printf("handleMetaInfo()err:%s\nmetakey:%s\nfrom:%s\n", err, metaKey, p.Pretty())
	}

	if rec == nil {
		rec = MakePutRecord(metaKey, res)
		rpmes.Record = rec
	} else {
		rpmes.Record.Value = []byte(res)
	}

	return rpmes, err
}

//广播的Meta操作的回调，在操作最后，会在返回信息中添加与本节点相连的最近节点
func (dht *KadDHT) handleMetaBroadcast(ctx context.Context, key string, p peer.ID) (_ *pb.Record, err error) {
	ctx = logger.Start(ctx, "handleMetaRequest")
	logger.SetTag(ctx, "peer", p)
	defer func() { logger.FinishWithErr(ctx, err) }() //各项记录操作

	if dht.metahandler == nil {
		return nil, instance.ErrMetaHandlerNotAssign
	}

	res, err := dht.metahandler.HandleMetaMessage(1, key, nil, p.Pretty())
	if err != nil {
		log.Println("handleMetaBroadcast()err:%s\nmetakey:%s\nfrom:%s\n", err, key, p.Pretty())
	}

	rec := MakePutRecord(key, res)

	return rec, err
}

func (dht *KadDHT) handleIter(ctx context.Context, p peer.ID, pmes *pb.Message) (_ *pb.Message, err error) {
	//接受信息节点的前缀查询回调函数
	ctx = logger.Start(ctx, "handleIter")
	logger.SetTag(ctx, "peer", p)
	defer func() { logger.FinishWithErr(ctx, err) }() //各项记录操作

	resp := pb.NewMessage(pmes.GetType(), pmes.GetKey(), 0)

	prefix := pmes.GetKey() //从pmes中解出数据
	if prefix == nil {
		logger.Infof("Got nil record from: %s", p.Pretty())
		return nil, errors.New("nil record")
	}

	resp.Record, _ = dht.iterLocal(string(prefix))

	return resp, err
}

func (dht *KadDHT) handleAppendValue(ctx context.Context, p peer.ID, pmes *pb.Message) (_ *pb.Message, err error) {
	//append操作的回调函数，收到信息，调用dht层appendLocal函数，将value添加在原value之后
	ctx = logger.Start(ctx, "handleAppendValue")
	logger.SetTag(ctx, "peer", p)
	defer func() { logger.FinishWithErr(ctx, err) }() //各项记录操作

	rec := pmes.GetRecord()
	if rec == nil {
		logger.Infof("Got nil record from: %s", p.Pretty())
		return nil, errors.New("nil record")
	}

	err = dht.appendLocal(string(rec.GetKey()), rec) //这里直接调用接受节点的appendlocal函数

	return pmes, err
}
