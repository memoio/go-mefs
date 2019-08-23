# TendermintTx处理（暂定）

## 1. 数据结构定义

这里是用gogoprotobuf定义的，Tendermint官方还有一个amnio支持interface的序列化与反序列化，可能更方便。

首先定义如下类型：
在`$GOPATH/src/github.com/memoio/go-mefs/abci/memoriae/pb`下
```protobuf
message Tx {   //交易
    bytes Payload = 1; //将Payload先序列化,类型可以新订
    string Typ = 2; //根据不同类型Payload进行反序列化，解析出Tx
	bytes Signature = 3;  //发起Tx的Validator的签名
    bytes PubKey  = 4;  //公钥
    string KeyType = 5; // 公钥的类型，用于将公钥转成crypto.Pubkey类型（是否必要？）
}

message KVPayload {  //只同步KV的Tx
    bytes Key = 1;
    bytes Value = 2;
    string Typ = 3;
    repeated bytes Sign = 4;
}

message ChallengerRequest{  //发起挑战
    string ChallengerPrivateKey =1;
	string AcceptChallengerAddress  =2;
	string DataPath         =3;       
}

message ChallengeResult {  //挑战结果
    bool Success = 1;
    google.protobuf.Timestamp Time = 2 [(gogoproto.nullable)=false, (gogoproto.stdtime)=true];
    string ChallengerAccess=3;
    string AcceptChallengerAddress=4;
    string DataPath =5;
    Payment Payment =6;
}

message Payment {
    google.protobuf.Timestamp StartTime = 1  [(gogoproto.nullable)=false, (gogoproto.stdtime)=true];
    google.protobuf.Timestamp EndTime = 2 [(gogoproto.nullable)=false, (gogoproto.stdtime)=true];
    uint64 sum =3;
}
```

首先定义了Tx，做成全部的，所有的交易都会有一种类型，将其作为Payload序列化成[]byte，然后再作为交易发出去，收到交易再根据Typ反序列化，然后做相应的处理。

## 2.运行流程

关于KV的同步，是否先需要Keeper的签名集？如果没有签名怎么验证KV的正确性，如果要签名集的话，本身又是一个拜占庭问题，能解决的话就不需要Tendermint了，这个流程需要好好想想。

## 3.接口调用
`import "github.com/memoio/go-mefs/abci"`

+ `func NewNode(nodePath string, groupId string, time time.Time) Node`

  新建一个节点，注意，目前同一个Group的多个节点，应该传入相同的time，目的是保证创世状态文件里的时间是一样的，目前思路，先选一个节点确定时间，再传给别人，其他几个用一样的。

+ `func (n *Node) InitHome() error`

  这个命令将会在`nodePath/config`下创建一个私钥文件(priv_validator.json)和一个创世状态文件(genesis.json)等等。

+ `func (n *Node) Run(p2pPort int, rpcPort int, peers []string) error`

  传入的是端口以及此节点启动后需要持久化连接的节点，格式：`ID@host:port`

+ `func (n *Node) GetID() (string, error)`

  获取此节点的ID

+ `func (n *Node) GetPublicKey() (string, error)`

  获取节点公钥等信息，会构建一个字符串，用于创世时加入到genesis.json里去，组成初始验证节点。

+ `func (n *Node) AlterValidator(validators []string)`

  用上面那个函数获取所有创世验证节点的公钥等信息（字符串），组成一个切片（保证顺序），在每个节点上运行此函数。在保证时间一样的情况下，他们就可以组成一样的genesis.json，开始进行共识。


## 4.测试方法

目前的简略测试方法见：`$GOPATH/src/github.com/memoio/go-mefs/abci/node_test.go`