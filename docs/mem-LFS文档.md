# LFS文档(对象存储兼容)_2019.3.16

[TOC]

## 1.数据结构概述

SuperBLock包含最基本的元数据，Bucket和object是对象存储里的基本概念，

+ Bucket（桶）

  桶是用户用于存储对象（Object）的容器，所有的对象都属于某个Bucket，同一个User的Bucket不允许重名。可以在创建时设置Bucket的某些属性，例如存储策略，纠删方式等等。Bucket内所有对象都按照这些属性进行存取，即Bucket是LFS设置策略的最基本单元，用户可以建立多个不同属性的桶，灵活控制不同对象冗余度、修复难度，耐用性以及价格等等。

+ Object（对象）

  对象是LFS中存储数据的基本单元，也可以认为是文件，只是没有文件存储的层次命名空间，在对象存储中，只能通过文件名前缀来模拟文件夹，故如果上传一个文件名为"a/b.txt"，将自动创建一个a的文件夹项。对象由元信息（Object Info），数据（Data）和文件名（Key）组成，在LFS现在的设计中，元信息（Object Info）和数据（Data）是分开存储维护的，元信息本身也被当成数据对待。对象有Bucket内唯一的Key来标识。对象元信息是一组键值对，表示了对象的属性，比如创建时间、大小、获取方式等信息，同时用户也可以在元信息中自定义一些信息便于检索。

+ Storage Policy（存储策略） 

  LFS提供对多副本、以及纠删的调整，可调整的属性为多副本副本数

其具体的定义如下

```protobuf
syntax = "proto3";
package lfs.pb;

message SuperBlock {
  map<int32, string> Buckets = 1; // User现有的Buckets
  int32 MetaBackupCount = 2; //对于所有的元数据块，副本数
  int32 NextBucketID = 3;
  int32 MagicNumber = 4;
  int32 Version = 5; //版本号
}

message BucketInfo {
  string BucketName = 1;
  int32 BucketID = 2;
  bool Policy = 3;     //存储策略-纠删(True)or多备份(False)
  int32 DataCount = 4; //若为纠删，则为数据块数量，否则为为主块数量（目前为1）
  int32 ParityCount = 5; //纠删下为校验块数量，多备份下为备份块数量
  int32 CurStripe = 6;        //当前写到哪一个Stripe
  int32 NextOffset = 7;        //当前写到哪一个Offset
  string Ctime = 8;           //本Bucket的创建时间
  string LastModified = 9;    //本Bucket内数据上次修改时间
  int32 ObjectsBlockSize = 10; //本Bucket的ObjectsBlock序列化后占多少字节
  uint64 SegmentSize = 11;   //本Bucket的segment size
  uint64 TagFlag = 12;       //本Bucket的TagFLag
}

message ObjectInfo {
  string ObjectName = 1;            //对象名称
  int32 ObjectSize = 2;             //对象长度
  string ContentType = 3;           //对象的类型，如文本、图片
  string ETag = 4;                  //生成的校验码，S3使用MD5
  int32 BucketID = 5;               //本对象所属的Bucket
  map<string, string> Metadata = 6; // User可以对文件自定义一些元信息
  int32 StripeStart = 7;            //本对象的起始Stripe
  int32 offsetStart = 8;            //本对象在起始Stripe的起始offset
  string Ctime = 9;                 //对象创建时间
  bool Deletion = 10;               //对象是否已删除
  bool Dir = 11;                    //是否为目录
}
```

在存储文件是，每一个文件将先切分成一个一个的原数据block（目前为1M），然后再根据冗余策略将原数据块Encode（切分成Segments，并生成相应的Tag），按顺序填塞Bucket下的Stripe，直到文件完全上传完毕。新加入的Block则从上一次写到的Stripe_Offset，顺着往下写（对小文件顺序追加写比较友好）。目前同一Bucket的文件写入只能串行顺序追加，不能同时写多个文件，可以同时读取，但不同Group的写入支持有限的同时写入。（如果有迁移功能的话，并行写入，可以先写在Bucket的随机temp_stripe下，再迁移合并）

对于一个普通的Block，其CID格式为 UID_BkID_SID_BID(UserID_BucketID_StripeID_BlockID)。比如17nCkDaiLmPCu4qxsQSxRP6Afy4_1_32_5，即表明PeerID为17nCkDaiLmPCu4qxsQSxRP6Afy4的Bucket_1，第32个stripe的第5个块。

对于元数据块，元数据块默认为多副本模式,其规定如下：

+ superBlock，占据UID_0_0_0...UID_0_m_n这个CID（故用户可获取的BucketID是从1开始的），不过目前SuperBlock存的东西略少，是否可以存一些User自己的config放里面。
+ BucketInfo，占据UID_ -BucketID_ 0_0...UID_ -BucketID_ 0_n（副本数），BucketInfo大小受限，不会超过一个Block大小（目前为1M），故只留了一个Block的位置。
+ ObjectInfo，占据UID_ -BucketID_ 1_0...UID_ -BucketID_ m_n，即负数的BucketID，以及从1开始的Stripe号，object可以随意累加。

LFS全局维护一个Log信息，即LFS的全局元数据。 

```protobuf
type Logs struct {
	node         *core.MefsNode
	sb           *pb.SuperBlock
	dirty   bool                                //看看superBlock是否需要更新（仅在新创建Bucket时需要）
	BucketByName map[string]*pb.BucketInfo           //通过BucketName找到Bucket信息
	bucketByID   map[int32]*pb.BucketInfo            //通过BucketID知道到Bucket信息
	Entries      map[int32]map[string]*pb.ObjectInfo //通过BucketID检索Bucket下文件
	dirty        map[int32]bool                      //通过BucketID确定一个Bucket是不是脏
}
```



## 2.可用接口

## 3.CLi Commands

```
  mefs lfs create_bucket <BucketName>              - Put a bucket to lfs.
  mefs lfs delete_bucket <BucketName>              - Delete a bucket in lfs.
  mefs lfs delete_object <BucketName> <ObjectName> - Delete a  object.
  mefs lfs fsync                                   - flush lfs metablock to maintain consistency.
  mefs lfs get_object <BucketName> <ObjectName>    - Get a object to specified outputpath or current work directory
  mefs lfs head_bucket <BucketName>                - Print a Bucket MetaData.
  mefs lfs head_object <BucketName> <ObjectName>   - Print information of a lfs object.
  mefs lfs list_buckets                             - List buckets in lfs.
  mefs lfs list_objects <BucketName>                - List objects in the specified bucket.
  mefs lfs put_object <data> <bucketname>          - Put a file as object to the specified bucket in lfs.

```

+ `mefs lfs create_bucket [--policy=false] [--datacount=<datacount> | --dc] [--paritycount=<paritycount> | --pc] [--] <BucketName>`

  可选选项： 

    --pl     --policy          bool - Storage policy. Default: true.
    --dc,   --datacount   int  - data count. Default: 3.
    --pc,   --paritycount int  - parity count. Default: 2.

+ `mefs lfs put_object [--objectname=<objectname> | --obn] [--] <data> <bucketname>`

  可选选项： 

  --obn, --objectname string - The name of the file or Bucket that you want to put. Default: .

  如果不指定对象名，则直接用上传的文件名。

+ `mefs lfs get_object [--output=<output> | -o] [--] <BucketName> <ObjectName>`

  可选选项：

   -o, --output string - The path where the output should be stored.

  不指定则直接输出到当前文件夹

+ 还有list_objects 和list_buckets，可以指定--pre=xxx，用于过滤list的结果

## 4.S3 API兼容列表

S3为RESTful风格的API设计，而IPFS为RPC风格，体现在，S3的action全体现在HTTP方法上，如（GET, PUT, POST, HEAD, DELETE...），而IPFS的action体现在URL中，如 POST /xxxx/api/v0/add....。如果要兼容S3 API，需要修改IPFS的api Handler的func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request)函数，改成RESTful风格。

目前来看，可以兼容的API可以有

```
GET Service（获取桶列表）

DELETE Bucket
HEAD Bucket
PUT Bucket
GET Bucket(List Object)

GET Object
PUT Object
HEAD Object
POST Object
DELETE Object
```

上传超大文件（大于5G），需要加入MultiPart分块上传下载功能（每个Part都占据一个ObjectInfo？）。

## 5.测试方法

### 5.1 启动测试节点

下面的命令可以直接复制到命令行批量运行

安装metb

cd $GOPATH/github.com/memoio/go-mefs/source/metb-plugins/metb
go install .

1. 首先Init节点

```shell

metb auto -type localmefs -count 13

y

```

2. 然后启动节点修改角色，`0 1 - user, 2 3 4 5- keeper, 6 7 8 9 10 11 12 - provider`


```shell

metb start

metb run -- mefs config Test true --json

metb run 2 mefs config Role keeper

metb run 3 mefs config Role keeper

metb run 4 mefs config Role keeper

metb run 5 mefs config Role keeper

metb run 6 mefs config Role provider

metb run 7 mefs config Role provider

metb run 8 mefs config Role provider

metb run 9 mefs config Role provider

metb run 10 mefs config Role provider

metb run 11 mefs config Role provider

metb run 12 mefs config Role provider

```

3. 修改完角色，重启节点连接节点

```shell

metb restart

metb connect 2 3
metb connect 2 4
metb connect 2 5
metb connect 2 6
metb connect 2 7
metb connect 2 8
metb connect 2 9
metb connect 2 10
metb connect 2 11
metb connect 2 12

metb connect 3 4
metb connect 3 5
metb connect 3 6
metb connect 3 7
metb connect 3 8
metb connect 3 9
metb connect 3 10
metb connect 3 11
metb connect 3 12

metb connect 4 5
metb connect 4 6
metb connect 4 7
metb connect 4 8
metb connect 4 9
metb connect 4 10
metb connect 4 11
metb connect 4 12

metb connect 5 6
metb connect 5 7
metb connect 5 8
metb connect 5 9
metb connect 5 10
metb connect 5 11
metb connect 5 12

metb connect 0 1
metb connect 0 2
metb connect 0 3
metb connect 0 4
metb connect 0 5
metb connect 0 6
metb connect 0 7
metb connect 0 8
metb connect 0 9
metb connect 0 10
metb connect 0 11
metb connect 0 12

metb connect 1 2
metb connect 1 3
metb connect 1 4
metb connect 1 5
metb connect 1 6
metb connect 1 7
metb connect 1 8
metb connect 1 9
metb connect 1 10
metb connect 1 11
metb connect 1 12

```

### 5.2 命令行测试

1. 进入一个user，测试lfs的命令行命令

```

metb shell 0

mefs lfs create_bucket bucket01

mefs lfs put_object  /home/xxx/xxx/1.txt bucket01

mefs lfs get_object bucket01  1.txt

mefs lfs list_objects


```

测试完要退出的话，最好先退出User，再退出其他的

```shell

metb stop 0

metb stop 1

metb stop
```

### 5.3 HTTP curl模拟测试

在启动好节点后，也可以使用HTTP来对User进行操作，举例如下

+ 首先，进入一个User的IPFS_PATH，比如~/testbed/0，然后vim daemon.stout，找到`API server listening on /ip4/127.0.0.1/tcp/41047`这一行，即为该User的API端口。
+ 然后运行，`curl "http://localhost:41047/api/v0/lfs/create_bucket?arg=bucket03&policy=false&dc=3&pc=2" `即可put bucket。
+ 运行`curl -F file=@Typora-linux-x64.tar.gz "http://localhost:43363/api/v0/lfs/put_object?arg=bucket01"`即可上传文件，file=@这里即可是相对路径也可是绝对路径，不指定文件名则取默认。
+ 运行`curl "http://localhost:43363/api/v0/lfs/get_object?arg=bucket01&arg=阿里云OSSAPI.pdf" -o 阿里云oss.pdf`即可下载文件，-o是指定curl输出二进制的路径，这里必须指定。
+ 其他的HTTP测试都大同小异。
