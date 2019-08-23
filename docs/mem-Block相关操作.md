# Block相关操作

## 基本概念

Block定义（go/src/github.com/memoio/go-mefs/source/go-block-format/blocks.go）

```go
type Block interface {
	RawData() []byte
	Cid() cid.Cid
	String() string
	Loggable() map[string]interface{}
}
```

每一个block对应一个Cid（Content Identity），Cid唯一标志一个block。

原ipfs里cid共有两个版本，v0和v1，其中v1可以兼容多种编码。（go/src/github.com/memoio/go-mefs/source/go-cid/cid.go）

现在在cid里加了一个cidv2，近似为一个cid包装的string类型。

假设有一个string类型的值，想转化成cid格式，可以使用以下代码：

```go
import(
    b58 "github.com/mr-tron/base58/base58"
    cid "github.com/memoio/go-mefs/source/go-cid"
)

var ncid cid.Cid //ncid表示返回的cid
if len(cid) == 46 {  //长度=46,为普通cid
    pidb, _ := b58.Decode(p) //p为输入的string
	ncid = cid.NewCidV0(pidb)
} else if len(p) > NEWCIDLENGTH { //长度>30，为自定义（peerid_chunkid形式的cid,peerid长度为30）
    ncid = cid.NewCidV2([]byte(p))
}
```



## 命令行操作

+ `$mefs block put 文件名`

  这个命令会读取文件，并根据文件哈希得一个v0的cid，然后构造一个block存在本地的.ipfs/data

+ `$mefs block putn 文件名 chunkid`

  这个命令会读取文件，使用本节点的peerid+chunkid构造一个v2的cid，然后构造一个block存在本地的.ipfs/data

  如下：

	```shell
	$mefs block putn 1.txt 001
	QmR3sEELV73vy1YiMAkayib1sQ1XrbE1mDRZrCqtRdNzjA_001
	```

+ `$mefs block putto cid peerid`

  这个命令会从本地寻找输入cid标志的block，并将其put到peerid指定的节点上。

+ `$mefs block get cid`

  这个命令会首先在本地查找cid标志的block，查找失败会在全网继续查找该block，最后将其存在本地。

+ `$mefs block getfrom cid peerid `

  这个命令只给指定的节点发送消息，要求cid标志的块，不会在本地查找，也不会全网查找。

## 接口

+ `func (s *blockService) GetFrom(peerid string, ncid string, tim time.Duration) (blocks.Block, error)` 

  这个函数定义在（go/src/github.com/memoio/go-mefs/source/go-blockservice/blockservice.go）里，输入为peerid，字符串形式的cid以及超时时间，返回一个block以及错误信息。

  由于方法实现在blockservice里，所以可以**获取当前的IpfsNode实例，假设获取为n，那么使用n.Blocks即可获取所有blockservice层的函数**，在core包里（go/src/github.com/memoio/go-mefs/core/core.go）我**定义了一个全局变量LocalNode，里面保存了当前节点的IpfsNode实例，以及Config，可以获取Node以及角色Role**。如下：

  ```go
  n := core.LocalNode.Node
  n.Blocks.GetBlockFrom(...)
  n.Blocks.PutBlockTo(...)
  role := core.LocalNode.Cfg.Role
  if role == "keeper" {
      ...
  }
  ```

  使用举例如下：

  ```go
  import(
      "github.com/memoio/go-mefs/core"
      "fmt"
      "time"
  )
  func Challenge(provider string, cids []string)(uint64){
      n := core.LocalNode.Node
      var credit uint64
      credit = 0
      for _, ncid := range cids{
          b,err := n.Blocks.GetBlockFrom(provider,ncid,20*time.Second)
          if err != nil {
              return 0
          }
          credit++
      }
      return credit
  }
  ```


+ `func (s *blockService) GetBlock(ncid cid.Cid)(blocks.Block, error)`

  此函数用于直接从本地获取Block，使用方法类似上面所说，首先获取当前的Node，再调用Node的blockservice即可调用到此函数。

+ `func (s *blockService) PutTo(ncid string, peerid string)(error)`

  此函数用于将ncid标识的block发送到peerid指定的节点上,该函数会自行在本地查找相应的block，若本地无该block则返回错误，使用方法基本同上。 