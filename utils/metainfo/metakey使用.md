# MetaKey使用
## 节点之间信息交互的方法


### 操作码opcode：
  + 标识这段信息处理哪一种操作，放在dht pmes中随KV对一起发送。对方节点接到信息之后，根据操作码来调用不同的回调函数完成操作。


### 数据格式
  + `注意！` 这里分隔符用‘/’表示，在实际应用中，可能不使用'/'
  + key有固定的格式： `/mainID/keytype/arg1/arg2......`
    + `mainID`是本次操作主要节点的ID号，可以是UKP任何一个角色,也可能是blockID
    + `keytype`是这段信息的类型，回调函数根据不同的keytype做不同的操作
    + `arg`是操作数，具体信息不同，数量内容不一样
  + value记录本次数据传输中的主要数据，比如keeper的ID 挑战中的结果和签名等


### 具体操作以及信息汇总
  +  初始化流程
     +  keytype:`"init_req"`/`"init_res"`/`"init_notif"`/`user_init_notif_res`
     +  user： key:`UserID/"init_req"/keepercount/providercount` value:无
     +  keeper： key:`UserID/"init_res"/keepercount/providercount` value:`kid1kid2.../pid1pid2....`
     +  user： key:`UserID/"init_notif"/keepercount/providercount` value:`kid1kid2.../pid1pid2....`
     +  keeper: key:`PeerID/"user_init_notif_res"/"bft"` value:`"simple"`或`IP:p2pport/IP:rpcport`
  + 元数据同步
    + keytype:`"sync"`，第一个操作数表示是哪一类数据
    + block位置信息 key:`blockID/"sync"/"block"` value:`pid/offset`
    + 挑战结果信息: key:`uid/"sync"/"chalres"/pid/kid/time` value: `length/result/proof/sum/h`
    + 挑战汇总信息(支付): key:`uid/"sync"/"chalpay"/pid/beginTime/endTime` value:`spacetime/signature/proof`
    + tendermint信息：key:`uid/"sync"/"tendermintinfo"` value:`id/ip/pubkey/p2pport/rpcport`
  + tendermint重新启动
    + keytype:`"trestart"`
    + key:`uid/"trestart"` value:`id/ip/pubkey/p2pport/rpcport`
  + 挑战流程
    + keytype:`"challenge"`/`"Proof"`
    + keeper： key:`uid/"challenge"/chaltime` value:`userconfigByte`
    + provider: key:`uid/"proof"/FaultBlock/chaltime` value:`proof`
  + 修复流程`Repair`
    + keeper: key：`blockID/"repair"` value:`rpid`
    + provider: key:`blockID/"repair_res"`
  + 数据块删除
    + keytype:`"delete_block"`
    + key：`blockID/"delete_block"` value:无
  + User的各项查询
    + keytype:`"query"`
    + key:value:`uid/"query"/"lastchal"`
  + User申请新的provider
    + keytype：`NewKPReq`
    + key:`uid/"UserNewKP"/count` value:`pid1pid2...`(无分隔符，当前拥有的provider 的id)
  + provider冷启动相关操作
    + 添加pos块 keytype：`"PosAdd"`  key:`pid/"PosAdd"`  value:`blockid1/blockid2/blockid3.........`
    + 删除pos块 keytype:`PosDelete`  key:`pid/"PosDelete"` value:`blockid1/blockid2/blockid3.........`
  + 其他保存在本地的KV
    + keytype: `"local"`
    + block位置信息 key:`blockID/"local"/"block"` value:`pid/offset`
    + config信息(目前特指bls12) key:`PeerID/"local"/"config"/"bls12"` value:具体数据
    + 角色信息 key:`PeerID/"local"/"roleinfo"` value:`user`/`keeper`/`provider`
    + legerinfo信息 key:`PeerID/"local"/legerinfo` value:具体数据
    + 最近一次支付的信息 key:`uid/"local"/"lastpay"/pid` value:`beginTime/endTime/spacetime/signature/proof`
    + 与角色相连的其他节点信息 
      + user key:`PeerID/"local"/"uids"` value:`uid1uid2uid3....`
      + keeper key:`PeerID/"local"/"kids"` value:`kid1kid2kid3....`
      + provider key:`PeerID/"local"/"pids"` value:`pid1pid2pid3....`
    + 节点的tendermint信息 key:`PeerID/"local"/"bft"` value:`"simple"`或`IP:p2pport/IP:rpcport`
    + ledger信息和credit信息
      + key:`PeerID/"local"/"ledgerinfo"`
      + key:`PeerID/"local"/"credit"`
    + chanelAddr信息
      + key`uid/"local"/"chaneladdr"/pid` value:`chanelAddr`
    + channelValue信息
      + key `channeladdr/"local"/"channelvalue"` value:`channel.value`


###使用逻辑
  + 数据结构
    ```go
    type KeyMeta struct {
        mid     string 
        keytype string
        options []interface{}
    }
    ```
  + 收到pmes，根据opcode不同，进入不同的回调函数，key整理成KeyMeta结构，进行之后的操作
  + 数据发送
    dht层`SendMetaMessage()`和`SendMetaRequest()`这两个函数用于发送数据，不同之处是`SendMetaMessage()`不需要返回值
  + 数据接受
    dht层 handlers.go `handleMetaInfo()`用于数据接受，收到的数据传入角色层，根据不同角色的回调函数进行处理
  + 数据结构的构造
    ```go
    km.err := metainfo.NewKeyMeta(mainID string, keyType MetaKeyType, listOptions ...string)
    ```
    这个函数会检查输入的参数是否满足keytype的要求，不满足会报错，想要构造字符串信息，先构造keymeta结构，然后`km.ToString()`
    ```go
    km,err := metainfo.GetKeyMeta(key string)
    ```
    这个函数从字符串构造keymeta，同样会根据keytype查错
  + 信息类别的添加
    keymeta.go中，添加到const中，并且在keytypemap中添加这个信息类别key的长度