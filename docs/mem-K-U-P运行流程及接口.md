# K-U-P运行流程及接口V2019.1.28

## 一、元数据接发方式

定义了一个接口类型-MetaMessageHandler，其具体定义如下

```go
type MetaMessageHandler interface {
	HandleMetaMessage(MetaMessageType, string, string, string) (string, error)
	GetRole() (string, error)
}

//下面是用到的一些类型或变量的定义
type MetaMessageType int32

const MetaHandleCompleted = "Completed"
const MetaHandleError = "Error"

const (
	RoleKeeper   = "Keeper"
	RoleUser     = "User"
	RoleProvider = "Provider"
)
//只是将DHT protobuf生成的类型定义值转接到这里
const (
	User_Init_Req    MetaMessageType = MetaMessageType(dht_pb.Message_User_Init_Req)
	User_Init_Res    MetaMessageType = MetaMessageType(dht_pb.Message_User_Init_Res)
	New_User_Notif   MetaMessageType = MetaMessageType(dht_pb.Message_New_User_Notif)
	Block_Meta       MetaMessageType = MetaMessageType(dht_pb.Message_Block_Meta)
	Delete_Block     MetaMessageType = MetaMessageType(dht_pb.Message_Delete_Block)
)
```

在IpfsDHT层将这个接口类型当成一个属性，每个角色（keeper，user，provider)都有一个根据该类型的具体实现。

在节点启动后，自动将DHT的MetaMessageHandler挂接为具体角色的类型，举例如下：

```go
type keeperHandler struct {
	Role string
}

func (keeper *keeperHandler) HandleMetaMessage(typ meta.MetaMessageType, metaKey, metaValue, from string) (string, error) { //from即发来这个请求的节点ID，可以做一点最基本的检查
    if keeper == nil {
		return meta.MetaHandleError, ErrKeeperServiceNotReady
	}
    switch typ {
        //根据不同类型操作
    }
}

func (keeper *keeperHandler) GetRole() (string, error) {
	if keeper == nil {
		return "", ErrKeeperServiceNotReady
	}
	return keeper.Role, nil
}
```

角色可以通过DHT的下列操作发送元数据信息

```go
MetaInitRequest(metaKey string) ([]byte, error)                    //User请求分配
MetaInitResponse(metaKey, metaValue, to string) (string, error)    //Keeper回复User
MetaNewUserNotif(metaKey, metaValue, to string) (string, error)    //通知Keeper有新user
MetaPutBlockMeta(metaKey, metaValue, to string) (string, error)    //通知Keeper有新block
MetaDeleteBlock(metaKey, metaValue, to string) (string, error)     //通知provider和keeper删除块
```

接收方在DHT里收到发来的元数据信息后，调用相应的handle函数进行操作，而handle函数又会调用DHT的`HandleMetaMessage(typ meta.MetaMessageType, metaKey, metaValue, from string)`函数（已经挂载本节点角色），进而使得当前角色可以处理相应的元数据信息。

处理完以后，在回复中将Value设置为MetaHandleCompleted或者MetaHandleError，元数据发送方可以根据回复判断自己的请求有没有成功，进而进行处理。

**疑问：DHT是用stream来控制回话的，有的时候发现这个stream被reset了，这也就收不到回复了，应该是因为两个Peer之间只能维持一个stream，新建一个stream会reset上一个，不知道能不能多开stream会话。**

**Todo：Key Value可以在发送方这里做一个签名，然后接收方再验证一下。**

## 二、keeper

+ `func StartKeeperService(ctx context.Context, node *core.MefsNode)`

  启动keeper服务

+ `func searchAllKeepersAndProviders(ctx context.Context) error`

  不断搜索全网的keeper和provider，并将其加入PeersInfo，全局搜索是用下面的协程

+ `func NewConnPeerRole(kidsMeta, pidsMeta string, PeerIDch chan string) error`

  协程将新连接的Provider和keeper加入列表。PeerIDch，是一个通道，每当新连接上一个节点，就会将其PeerID发过来**（只有角色为Keeper才发）**。

  ***

  下面是metamessage的处理函数

+ `func HandleUserInitReq(userID string, keeperCount, providerCount int, from string) (string, error)`

  当收到user的init请求后，按照user需要的K-P数量分配好，然后用MetaInitResponse通知User。

+ `func HandleNewUserNotif(metaKey, kidsAndPids, from string) error`

  判断是User自己发的还是其他Keeper发的，并进行相应操作

+ `func HandleAddBlockMeta(blockID, metaValue, from string) error`

  收到user的块元信息，按照Opcode的说明进行操作。

+ `func HandleDeleteBlockMeta(blockID, from string) error`

  收到user的请求，删除某个块，此时将本地关于此块的元数据删除。

  是否应挑战Provider，检测其是否删除数据块？

## 三、User

+ `func StartGroupService(ctx context.Context, node *core.MefsNode, IsInit bool)`

  这个启动相关的User服务，`ResponseChan = make(chan string)`，User里的这个通道可以处理对某个发出消息的回复，有时候会用到。

+ `func findKeepersAndProvider(ctx context.Context, IsInit bool) error`

  初始化一个节点，状态IsInit为true，初始化一次后即设置成false。

  如果不是Init，User会根据`"/metainfo/" + localID + "/kids"`找到自己的keeper，找到Keeper即可算启动服务了。

  如果是Init，User会用MetaInitRequest-`"/metainfo/localID/init/keeperCount/providerCount"`广播，直到有指定数量的Keeper回复为止（暂定两个)，然后监听本地keeper数量，等待有Keeper回信。（也可以直接从CmdGet的返回值里获得返回，不过感觉用这种会出现问题，因为用新发送，可以做成类似TCP的握手）。

  收到回复后handleUserInitRes会处理，并加入keeper。

+ `func GetProviders(count int) []string `

  先从内存中找，找不到则去Keeper那里找，使用之前先检查连接。

+ `func GetBlockProviders(blockID string) []string`

  以`"/block/" + blockID + "/pids"`的格式去Keeper那里找到provider。

+ `func putDataMetaToKeepers(blockID string, providers []string, keepers []string) error`

  以`"/metainfo/block/" + blockID + "/pids"`的格式通知keeper，自己将这个块存放在哪些provider上了。

+ `func deleteBlocksFromProvider(blockStart, blockCount int32, updateMeta bool) error`

  通知provider删除某些块，通知keeper自己删除了某些块。如果设置了updateMeta，证明是要更新metaBlock，那要返回DHT消息的返回值，看看是不是completed，再进行操作。

  ***

  下面是metamessage的处理函数

+ `func handleUserInitRes(metaKey, metaValue, from string) error`

  处理Keeper对Init的回复

## 四、Provider

`func StartProviderService(node *core.MefsNode) `

启动Provider服务

***

`func HandleDeleteBlock(blockID, from string) error `

收到通知，删除User在上面存的块

## 五、OpCode

+ User:

  InitReq(向全网申请Keeper)

  :key::`"/metainfo/userID/init/keeperCount/providerCount"`

  :v: :

+ Keeper:

  handleUserInitReq(处理)

  发起InitRes(抢先回复User)

  :key::`"/metainfo/usereID/init/keeperCount/providerCount"`

  :v: :`"kids/pids"`

+ User:

  handleUserInitRes(处理)

  NewUserNotif(回复Keeper，已确认)

  :key::`"/metainfo/userID/init/keeperCount/providerCount"`

  :v: :`"kids/pids"`

+ Keeper:

  handleNewUserNotif(处理)

  NewUserNotif(通知此User的其他keeper)

  :key::`"/metainfo/userID/init/keeperCount/providerCount"`

  :v: :`"kids/pids"`

+ 其他keeper:

  handleNewUserNotif(处理)

**UserInit流程结束**

------

+ User:

  PutBlockMeta(通知Keeper块的位置)

  :key::`"/metainfo/block/" + blockID + "/pids"`

  :v: :`pids`

+ Keeper:

  handleBlockMeta(处理)

  往本地put下面的K-V供User查询以及挑战

  :key::`"/block/" + blockID + "/pids"`

  :v::`pids`

------

+ User:

  DeleteBlock

  :key::`"/metainfo/delete/" + blockID`

  :v: :`localID`

+ Keeper:

  handleDeleteBlock(从本地KV里删除相应元数据，对相应块的挑战停止)

+ Provider:

  从本地删除数据块

+ 如果是UpdateMetablock，删除完User再put新的metaBlock一遍

## 六、测试方法

使用LFS的Command进行测试。

测试挑战、修复和同步的时候，可以kill一个provider，然后触发挑战失败，进行修复，然后同步成功

