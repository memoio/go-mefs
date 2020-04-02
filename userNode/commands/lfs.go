package commands

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	cmds "github.com/ipfs/go-ipfs-cmds"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/memoio/go-mefs/core/commands/cmdenv"
	"github.com/memoio/go-mefs/core/commands/e"
	id "github.com/memoio/go-mefs/crypto/identity"
	dataformat "github.com/memoio/go-mefs/data-format"
	"github.com/memoio/go-mefs/repo/fsrepo"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/userNode/user"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/mgutz/ansi"
)

const (
	SecreteKey = "secreteKey"
	PassWord   = "password"
)

type ObjectStat struct {
	Name           string
	Size           int64
	MD5            string
	Ctime          string
	Dir            bool
	LatestChalTime string
}

type Objects struct {
	Method  string
	Objects []ObjectStat
}

var (
	errLfsServiceNotReady       = errors.New("lfs service not ready")
	errGroupServiceNotReady     = errors.New("group service not ready")
	errNoFileToUpload           = errors.New("No file to upload")
	errSumCountsBeyondProviders = errors.New("The sum counts(datacount+paritycount) has beyond the providers counts")
	errWrongInput               = errors.New("The input option is wrong.")
)

func (ob ObjectStat) String() string {
	FloatStorage := float64(ob.Size)
	var OutStorage string
	if FloatStorage < 1024 && FloatStorage > 0 {
		OutStorage = fmt.Sprintf("%.2f", FloatStorage) + "B"
	} else if FloatStorage < 1048576 && FloatStorage >= 1024 {
		OutStorage = fmt.Sprintf("%.2f", FloatStorage/1024) + "KB"
	} else if FloatStorage < 1073741824 && FloatStorage >= 1048576 {
		OutStorage = fmt.Sprintf("%.2f", FloatStorage/1048576) + "MB"
	} else {
		OutStorage = fmt.Sprintf("%.2f", FloatStorage/1073741824) + "GB"
	}
	return fmt.Sprintf(
		"ObjectName: %s\n--ObjectSize: %s\n--MD5: %s\n--Ctime: %s\n--Dir: %t\n--LatestChalTime: %s\n",
		ansi.Color(ob.Name, "green"),
		OutStorage,
		ob.MD5,
		ob.Ctime,
		ob.Dir,
		ob.LatestChalTime,
	)
}

func (obs Objects) String() string {
	var str bytes.Buffer
	str.WriteString("Method: " + ansi.Color(obs.Method, "green") + "\n")
	for _, obStat := range obs.Objects {
		str.WriteString(obStat.String())
	}
	return str.String()
}

type BucketStat struct {
	Name        string
	BucketID    int64
	Ctime       string
	Policy      int32
	DataCount   int32
	ParityCount int32
	Encryption  int32
}

type Buckets struct {
	Method  string
	Buckets []BucketStat
}

func (bk BucketStat) String() string {
	return fmt.Sprintf(
		"Name: %s\n--BucketID: %d\n--Ctime: %s\n--Policy: %d\n--DataCount: %d\n--ParityCount: %d\n--Encryption:%d\n",
		ansi.Color(bk.Name, "green"),
		bk.BucketID,
		bk.Ctime,
		bk.Policy,
		bk.DataCount,
		bk.ParityCount,
		bk.Encryption,
	)
}

func (bus Buckets) String() string {
	var str bytes.Buffer
	str.WriteString("Method: " + ansi.Color(bus.Method, "green") + "\n")
	for _, buStat := range bus.Buckets {
		str.WriteString(buStat.String())
	}
	return str.String()
}

//PeerState 目前只做了最简单的状态记录
type PeerState struct {
	PeerID    string
	Connected bool
}

func (ps PeerState) String() string {
	if ps.Connected {
		return ps.PeerID + " connected"
	}
	return ps.PeerID + " unconnected"
}

type PeerList struct {
	Peers []PeerState
}

func (pl PeerList) String() string {
	var res string
	for _, ps := range pl.Peers {
		res += ps.String() + "\n"
	}
	return res
}

type StringList struct {
	ChildLists []string
}

type IntList struct {
	ChildLists []int
}

func (fl StringList) String() string {
	var buffer bytes.Buffer
	for i := 0; i < len(fl.ChildLists); i++ {
		buffer.WriteString(fl.ChildLists[i])
		buffer.WriteString("\n")
	}
	return buffer.String()
}

func (fl IntList) String() string {
	var buffer bytes.Buffer
	for i := 0; i < len(fl.ChildLists); i++ {
		buffer.WriteString(strconv.Itoa(fl.ChildLists[i]))
		buffer.WriteString("\n")
	}
	return buffer.String()
}

var LfsCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Interact with Lfs buckets and objects.",
		ShortDescription: `
'mefs lfs' is a plumbing command used to manipulate Lfs buckets and objects.
`,
	},

	Subcommands: map[string]*cmds.Command{
		"head_object":    lfsHeadObjectCmd,
		"put_object":     lfsPutObjectCmd,
		"get_object":     lfsGetObjectCmd,
		"list_objects":   lfsListObjectsCmd,
		"delete_object":  lfsDeleteObjectCmd,
		"head_bucket":    lfsHeadBucketCmd,
		"list_buckets":   lfsListBucketsCmd,
		"create_bucket":  lfsCreateBucketCmd,
		"delete_bucket":  lfsDeleteBucketCmd,
		"list_keepers":   lfsListKeepersCmd,
		"list_providers": lfsListProviderrsCmd,
		"list_users":     lfsListUsersCmd,
		"fsync":          lfsFsyncCmd,
		"start":          lfsStartUserCmd,
		"kill":           lfsKillUserCmd,
		"show_storage":   lfsShowStorageCmd,
		"get_share":      lfsGetShareCmd,
		"gen_share":      lfsGenShareCmd,
	},
}

const (
	BucketName = "bucketname"
	//BucketID    = "bucketid"
	Policy       = "policy"
	DataCount    = "datacount"
	ParityCount  = "paritycount"
	ObjectName   = "objectname"
	AddressID    = "address"
	Encryption   = "encryption"
	PrefixFilter = "prefix"
	AvailTime    = "Avail"
	OutputPath   = "output"
	ForceFlush   = "force" //设置这个选项，会强制刷新给Provider，无论是否表示为脏
)

// 关闭代理节点上user的服务
var lfsKillUserCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "End groupService and lfsService by user's address",
		ShortDescription: ``,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("addr", false, false, "The practice user's addressid that you want to kill"),
	},
	Options: []cmds.Option{
		cmds.StringOption(PassWord, "pwd", "The practice user's password that you want to exec").WithDefault(utils.DefaultPassword),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}
		userIns, ok := node.Inst.(*user.Info)
		if !ok {
			return ErrNotReady
		}
		var uid string
		if len(req.Arguments) > 0 {
			addr := req.Arguments[0]
			uid, err = address.GetIDFromAddress(addr)
			if err != nil {
				return err
			}
		} else {
			uid = node.Identity.Pretty()
		}

		// 查看pwd是否能获取sk，确定是user发起的kill命令
		pwd := req.Options[PassWord].(string)
		_, err = fsrepo.GetPrivateKeyFromKeystore(uid, pwd)
		if err != nil {
			return err
		}
		err = userIns.KillUser(uid)
		if err != nil {
			return err
		}

		addr, err := address.GetAddressFromID(uid)
		if err != nil {
			return err
		}
		// lfsService
		list := &StringList{
			ChildLists: []string{"Kill User: " + addr.String()},
		}
		return cmds.EmitOnce(res, list)
	},
	Type: StringList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, fl *StringList) error {
			_, err := fmt.Fprintf(w, "%s", fl)
			return err
		}),
	},
}

var errTimeOut = errors.New("Time Out")
var lfsStartUserCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "Start lfsService for user",
		ShortDescription: ``,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("addr", false, false, "Initialize user service with the given address."),
	},
	Options: []cmds.Option{
		cmds.StringOption(SecreteKey, "sk", "The practice user's private key that you want to exec").WithDefault(""),
		cmds.StringOption(PassWord, "pwd", "The practice user's password that you want to exec").WithDefault(utils.DefaultPassword),
		cmds.Int64Option("duration", "dur", "Time user wants to store data in deploying contracts, unit is day").WithDefault(utils.DefaultDuration),
		cmds.Int64Option("capacity", "cap", "Size user wants to store data in deploying contracts, unit is MB").WithDefault(utils.DefaultCapacity),
		cmds.Int64Option("storedPrice", "price", "Price user wants to store data in deploying contracts, unit is wei").WithDefault(utils.STOREPRICEPEDOLLAR),
		cmds.IntOption("keeperSla", "ks", "implement user needs how many keepers").WithDefault(utils.KeeperSLA),
		cmds.IntOption("providerSla", "ps", "implement user needs how many providers").WithDefault(utils.ProviderSLA),
		cmds.BoolOption("reDeployQuery", "rdo", "reDeploy query contract if user has not deploy upkeeping contract").WithDefault(false),
		cmds.BoolOption("force", "f", "force user to write mode").WithDefault(false),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}
		userIns, ok := node.Inst.(*user.Info)
		if !ok {
			return ErrNotReady
		}
		var addr = ""
		var uid = ""
		sk := req.Options[SecreteKey].(string)
		pwd := req.Options[PassWord].(string)
		if sk != "" {
			addrCommon, err := id.GetAdressFromSk(sk)
			if err != nil {
				return err
			}
			addr = addrCommon.Hex()
			uid, err = address.GetIDFromAddress(addr)
			if err != nil {
				return err
			}

			err = fsrepo.PutPrivateKeyToKeystore(sk, uid, pwd)
			if err != nil {
				return err
			}
		}

		if addr == "" {
			if len(req.Arguments) > 0 {
				addr = req.Arguments[0]
			}
			if addr != "" {
				uid, err = address.GetIDFromAddress(addr)
				if err != nil {
					return err
				}
			} else {
				uid = node.Identity.Pretty()
				addre, err := address.GetAddressFromID(uid)
				if err != nil {
					return err
				}
				addr = addre.Hex()
			}
		}

		capacity, ok := req.Options["capacity"].(int64) //user签署合约时指定的需求存储空间
		if !ok || capacity <= 0 {
			fmt.Println("input wrong capacity.")
			return errWrongInput
		}
		duration, ok := req.Options["duration"].(int64) //user签署合约时指定的需求存储时长
		if !ok || duration <= 0 {
			fmt.Println("input wrong duration.")
			return errWrongInput
		}
		price, ok := req.Options["storedPrice"].(int64) //user签署合约时指定的需求存储价格
		if !ok || price <= 0 {
			fmt.Println("input wrong price.")
			return errWrongInput
		}
		ks, ok := req.Options["keeperSla"].(int) //user签署合约时指定的需求keeper数量
		if !ok || ks <= 0 {
			fmt.Println("input wrong keeper nums.")
			return errWrongInput
		}
		ps, ok := req.Options["providerSla"].(int) //user签署合约时指定的需求provider数量
		if !ok || ps <= 1 {
			fmt.Println("input wrong provider nums.")
			return errWrongInput
		}

		rdo, ok := req.Options["reDeployQuery"].(bool)
		if !ok {
			fmt.Println("input wrong value for redeploy.")
			return errWrongInput
		}

		force, ok := req.Options["force"].(bool)
		if !ok {
			fmt.Println("input wrong value for force.")
			return errWrongInput
		}

		// 读keystore下uid文件
		hexSk, err := fsrepo.GetPrivateKeyFromKeystore(uid, pwd)
		if err != nil {
			return err
		}

		cfg, err := node.Repo.Config()
		if err != nil {
			return err
		}

		var qid string

		if cfg.Test {
			qid = uid
		} else {
			qitem, err := role.GetLatestQuery(uid)
			if err != nil {
				rdo = true
				qid = ""
			} else {
				qid = qitem.QueryID
			}
		}

		lfs, err := userIns.NewFS(uid, uid, qid, hexSk, capacity, duration, price, ks, ps, rdo, force)
		if err != nil {
			userIns.KillUser(uid)
			return err
		}

		err = lfs.Start(req.Context)
		if err != nil {
			userIns.KillUser(uid)
			return err
		}

		list := &StringList{
			ChildLists: []string{"User is started, the address is : " + addr},
		}
		return cmds.EmitOnce(res, list)
	},
	Type: StringList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, fl *StringList) error {
			_, err := fmt.Fprintf(w, "%s", fl)
			return err
		}),
	},
}

var lfsHeadObjectCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Print information of a lfs object.",
		ShortDescription: `
'mefs lfs head_object' is a plumbing command to print information of a lfs file.
 It outputs the following to stdout:

 	ObjectName	The Object to Head
 	ObjectSize	The Object Size(not include tag data)
 	Ctime       The Create time
 	Dir         Directory or Not
`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("BucketName", true, false, "The Bucket's name that object in."),
		cmds.StringArg("ObjectName", true, false, "The Object's Name"),
	},
	Options: []cmds.Option{
		cmds.StringOption(AddressID, "addr", "The practice user's addressid that you want to exec").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}

		userIns, ok := node.Inst.(*user.Info)
		if !ok {
			return ErrNotReady
		}

		var userid string
		addressid, found := req.Options[AddressID].(string)
		if addressid == "" || !found {
			userid = node.Identity.Pretty()
		} else {
			userid, err = address.GetIDFromAddress(addressid)
			if err != nil {
				return err
			}
		}
		lfs := userIns.GetUser(userid)
		if lfs == nil {
			return errLfsServiceNotReady
		}

		object, err := lfs.HeadObject(req.Context, req.Arguments[0], req.Arguments[1], user.DefaultOption())
		if err != nil {
			return err
		}
		lfsInfo, ok := lfs.(*user.LfsInfo)
		if !ok {
			return errLfsServiceNotReady
		}
		avail, err := lfsInfo.GetObjectAvailTime(object)
		if err != nil {
			return err
		}

		availTim, err := time.Parse(utils.BASETIME, avail)
		if err != nil {
			availTim = time.Unix(0, 0)
		}
		ctime := time.Unix(object.GetCTime(), 0).In(time.Local)
		objectStat := ObjectStat{
			Name:           object.GetInfo().GetName(),
			Size:           object.GetLength(),
			MD5:            object.GetETag(), //需要检查Parts是否为空
			Ctime:          ctime.Format(utils.SHOWTIME),
			Dir:            false,
			LatestChalTime: availTim.Format(utils.SHOWTIME),
		}

		return cmds.EmitOnce(res, &Objects{
			Method:  "Head Object",
			Objects: []ObjectStat{objectStat},
		})
	},
	Type: Objects{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, obs *Objects) error {
			_, err := fmt.Fprintf(w, "%s", obs)
			return err
		}),
	},
}

var lfsPutObjectCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Put a file as object to the specified bucket in lfs.",
		ShortDescription: `
'mefs lfs put_object' is a plumbing command for add file to lfs.
 It outputs the following to stdout:

	Method      Put Object	
	ObjectName	The Object Name
 	ObjectSize	The Object Size(not include tag data)
 	Ctime       The Create time
 	Dir         Directory or Not
`,
	},

	Arguments: []cmds.Argument{
		cmds.FileArg("data", true, false, "The data to be stored to LFS."),
		cmds.StringArg(BucketName, true, false, "BucketName you want to put object to"),
	},
	Options: []cmds.Option{
		cmds.StringOption(ObjectName, "obn", "The name of the file or Bucket that you want to put").WithDefault(""),
		cmds.StringOption(AddressID, "addr", "The practice user's addressid that you want to exec").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		//暂时不做成api层的接口
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}

		userIns, ok := node.Inst.(*user.Info)
		if !ok {
			return ErrNotReady
		}

		var userid string
		addressid, found := req.Options[AddressID].(string)
		if addressid == "" || !found {
			userid = node.Identity.Pretty()
		} else {
			userid, err = address.GetIDFromAddress(addressid)
			if err != nil {
				return err
			}
		}
		lfs := userIns.GetUser(userid)
		if lfs == nil || !lfs.Online() {
			return errLfsServiceNotReady
		}

		bucketName := req.Arguments[0]
		objectName := req.Options[ObjectName].(string)
		f := req.Files.Entries()
		//目前只上传第一个文件
		if !f.Next() {
			return errNoFileToUpload
		}

		if objectName == "" {
			objectName = f.Name()
		}
		file := f.Node()
		var fileNext files.File
		switch fileType := file.(type) {
		case files.Directory:
			return errors.New("unsupported now")
		case files.File:
			fileNext = fileType
		}
		object, err := lfs.PutObject(req.Context, bucketName, objectName, fileNext, user.DefaultOption())
		if err != nil {
			return err
		}

		ctime := time.Unix(object.GetCTime(), 0).In(time.Local)
		objectStat := ObjectStat{
			Name:  object.GetInfo().GetName(),
			Size:  object.GetLength(),
			MD5:   object.GetETag(),
			Ctime: ctime.Format(utils.SHOWTIME),
			Dir:   false,
		}
		return cmds.EmitOnce(res, &Objects{
			Method:  "Put Object Success",
			Objects: []ObjectStat{objectStat},
		})
	},
	Type: Objects{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, obs *Objects) error {
			_, err := fmt.Fprintf(w, "%s", obs)
			return err
		}),
	},
}

var lfsGetObjectCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Get a object to specified outputpath or current work directory",
		ShortDescription: `
'mefs lfs get_object' is a plumbing command for download a file or dir from lfs
 It outputs the following to stdout:

 	"GetObject success，Size: objectSize" or "error"

`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("BucketName", true, false, "The Group the file in"),
		cmds.StringArg("ObjectName", true, false, "The file in lfs you want to get."),
	},
	Options: []cmds.Option{
		cmds.StringOption(AddressID, "addr", "The practice user's addressid that you want to exec").WithDefault(""),
		cmds.StringOption(OutputPath, "o", "The path where the output should be stored."),
	},
	PreRun: func(req *cmds.Request, env cmds.Environment) error {
		outPath := getOutPath(req)
		fpath := filepath.Join(outPath, req.Arguments[1])
		_, err := os.Stat(fpath)
		if !os.IsNotExist(err) {
			return errors.New("The outpath already has file: " + req.Arguments[1])
		}
		return nil
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}
		userIns, ok := node.Inst.(*user.Info)
		if !ok {
			return ErrNotReady
		}
		var userid string
		addressid, found := req.Options[AddressID].(string)
		if addressid == "" || !found {
			userid = node.Identity.Pretty()
		} else {
			userid, err = address.GetIDFromAddress(addressid)
			if err != nil {
				return err
			}
		}

		lfs := userIns.GetUser(userid)
		if lfs == nil || !lfs.Online() {
			return errLfsServiceNotReady
		}

		piper, pipew := io.Pipe()
		bufw := bufio.NewWriterSize(pipew, user.DefaultBufSize)
		checkErrAndClosePipe := func(err error) error {
			if err != nil {
				err = pipew.CloseWithError(err)
				return err
			}
			err = pipew.Close()
			return err
		}
		var complete []user.CompleteFunc
		complete = append(complete, checkErrAndClosePipe)
		go lfs.GetObject(req.Context, req.Arguments[0], req.Arguments[1], bufw, complete, user.DefaultOption())

		return res.Emit(piper)
	},
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(res cmds.Response, re cmds.ResponseEmitter) error {
			req := res.Request()
			v, err := res.Next()
			if err != nil {
				return err
			}

			outReader, ok := v.(io.Reader)
			if !ok {
				return e.New(e.TypeErr(outReader, v))
			}
			outPath := getOutPath(req)
			rootExists := true
			rootIsDir := false
			var fpath string
			if stat, err := os.Stat(outPath); err != nil && os.IsNotExist(err) {
				rootExists = false
			} else if err != nil {
				return err
			} else if stat.IsDir() {
				rootIsDir = true
			}
			if rootIsDir {
				fpath = path.Join(outPath, req.Arguments[1])
			} else if !rootExists {
				fpath = outPath
			} else {
				return errors.New("The outpath already has file: " + req.Arguments[1])
			}
			var file *os.File
			if _, err := os.Stat(fpath); err != nil && os.IsNotExist(err) {
				file, err = os.Create(fpath)
				if err != nil {
					file.Close()
					return err
				}
			} else {
				return errors.New("The outpath already has file: " + req.Arguments[1])
			}
			//close这里不会报错么？
			defer file.Close()
			n, err := io.Copy(file, outReader)
			if err != nil {
				fmt.Println("Download failed - ", err)
				return err
			}
			fmt.Printf("GetObject to %s success，Size: %d\n", fpath, n)
			return nil
		},
	},
}

func getOutPath(req *cmds.Request) string {
	outPath, _ := req.Options[OutputPath].(string)
	if outPath == "" {
		outPath = "."
	}
	outPath = path.Clean(outPath)
	return outPath
}

var lfsListObjectsCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "List objects in the specified bucket.",
		ShortDescription: `
'mefs lfs list_objects' is a plumbing command for list objects in the specified bucket:

	ObjectName--BucketID--ObjectSize--IsDir

`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("BucketName", true, false, "The dirName you want to list").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.StringOption(AddressID, "addr", "The practice user's addressid that you want to exec").WithDefault(""),
		cmds.StringOption(PrefixFilter, "Prefix can filter result").WithDefault(""),
		cmds.BoolOption(AvailTime, "a", "The option determine wheather show available time.").WithDefault(true),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}
		userIns, ok := node.Inst.(*user.Info)
		if !ok {
			return ErrNotReady
		}
		var userid string
		addressid, found := req.Options[AddressID].(string)
		if addressid == "" || !found {
			userid = node.Identity.Pretty()
		} else {
			userid, err = address.GetIDFromAddress(addressid)
			if err != nil {
				return err
			}
		}
		prefix, found := req.Options[PrefixFilter].(string)
		if !found {
			prefix = ""
		}
		avail, found := req.Options[AvailTime].(bool)
		if !found {
			avail = true
		}

		lfs := userIns.GetUser(userid)
		if lfs == nil || !lfs.Online() {
			return errLfsServiceNotReady
		}

		bucketName := req.Arguments[0]
		objects, err := lfs.ListObjects(req.Context, bucketName, prefix, user.DefaultOption())
		if err != nil {
			return err
		}

		objectsInfo := &Objects{
			Method: "List Objects",
		}
		for _, object := range objects {
			ctime := time.Unix(object.GetCTime(), 0).In(time.Local)
			// init with creation time
			avaTime := ctime.Format(utils.SHOWTIME)
			if avail {
				at, err := lfs.(*user.LfsInfo).GetObjectAvailTime(object)
				if err == nil {
					availTim, err := time.Parse(utils.BASETIME, at)
					if err == nil {
						avaTime = availTim.Format(utils.SHOWTIME)
					}
				}
			}
			tempObState := ObjectStat{
				Name:           object.GetInfo().GetName(),
				Size:           object.GetLength(),
				MD5:            object.GetETag(),
				Ctime:          ctime.Format(utils.SHOWTIME),
				Dir:            false,
				LatestChalTime: avaTime,
			}
			objectsInfo.Objects = append(objectsInfo.Objects, tempObState)
		}
		return cmds.EmitOnce(res, objectsInfo)
	},
	Type: Objects{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, obs *Objects) error {
			_, err := fmt.Fprintf(w, "%s", obs)
			return err
		}),
	},
}

var lfsDeleteObjectCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Delete a  object.",
		ShortDescription: `
'mefs lfs delete_objects' is a plumbing command to delete a object,.
 now it only set the deletion flag, don't delete the data.
 It outputs the following to stdout:

    Method      Delete Object
 	ObjectName	The Object to Head
 	ObjectSize	The Object Size(not include tag data)
 	Ctime       The Create time
 	Dir         Directory or Not

`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("BucketName", true, false, "The Bucket's name that object in."),
		cmds.StringArg("ObjectName", true, false, "The Object's Name"),
	},
	Options: []cmds.Option{
		cmds.StringOption(AddressID, "addr", "The practice user's addressid that you want to exec").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}
		userIns, ok := node.Inst.(*user.Info)
		if !ok {
			return ErrNotReady
		}
		var userid string
		addressid, found := req.Options[AddressID].(string)
		if addressid == "" || !found {
			userid = node.Identity.Pretty()
		} else {
			userid, err = address.GetIDFromAddress(addressid)
			if err != nil {
				return err
			}
		}
		lfs := userIns.GetUser(userid)
		if lfs == nil || !lfs.Online() {
			return errLfsServiceNotReady
		}

		object, err := lfs.DeleteObject(req.Context, req.Arguments[0], req.Arguments[1])
		if err != nil {
			return err
		}

		ctime := time.Unix(object.GetCTime(), 0).In(time.Local)
		objectStat := ObjectStat{
			Name:  object.GetInfo().GetName(),
			Size:  object.GetLength(),
			MD5:   object.GetETag(),
			Ctime: ctime.Format(utils.SHOWTIME),
			Dir:   false,
		}
		return cmds.EmitOnce(res, &Objects{
			Method:  "Delete Object",
			Objects: []ObjectStat{objectStat},
		})
	},
	Type: Objects{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, obs *Objects) error {
			_, err := fmt.Fprintf(w, "%s", obs)
			return err
		}),
	},
}

var lfsHeadBucketCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Print a Bucket MetaData.",
		ShortDescription: `
'mefs lfs head_bucket' is a plumbing command for printing a bucket metadata:

	Method: 		Head Bucket
	BucketName: 	bucket's name
	BucketID: 		bucket's ID
	Ctime: 			Creation Time
	Policy:			Erasure code or MultiReplication
	DataCount: 		Data count
	ParityCount: 	Parity count

`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("BucketName", true, false, "The dirName you want to list").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.StringOption(AddressID, "addr", "The practice user's addressid that you want to exec").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}
		userIns, ok := node.Inst.(*user.Info)
		if !ok {
			return ErrNotReady
		}
		var userid string
		addressid, found := req.Options[AddressID].(string)
		if addressid == "" || !found {
			userid = node.Identity.Pretty()
		} else {
			userid, err = address.GetIDFromAddress(addressid)
			if err != nil {
				return err
			}
		}
		lfs := userIns.GetUser(userid)
		if lfs == nil || !lfs.Online() {
			return errLfsServiceNotReady
		}

		bucketName := req.Arguments[0]
		bucket, err := lfs.HeadBucket(req.Context, bucketName)
		if err != nil {
			return err
		}
		ctime := time.Unix(bucket.GetCTime(), 0).In(time.Local)
		bucketStat := BucketStat{
			Name:        bucket.Name,
			BucketID:    bucket.BucketID,
			Ctime:       ctime.Format(utils.SHOWTIME),
			Policy:      bucket.BOpts.Policy,
			DataCount:   bucket.BOpts.DataCount,
			ParityCount: bucket.BOpts.ParityCount,
			Encryption:  bucket.BOpts.Encryption,
		}
		return cmds.EmitOnce(res, &Buckets{
			Method:  "Head Bucket",
			Buckets: []BucketStat{bucketStat},
		})
	},
	Type: Buckets{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, bus *Buckets) error {
			_, err := fmt.Fprintf(w, "%s", bus)
			return err
		}),
	},
}

var lfsCreateBucketCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "create a bucket in lfs.",
		ShortDescription: `
'mefs lfs create_bucket' is a plumbing command for putting a bucket to lfs:

	
	Method: 		Create Bucket
	BucketName: 	bucket's name
	BucketID: 		bucket's ID
	Ctime: 			Creation Time
	Policy:			Erasure code or MultiReplication
	DataCount: 		Data count
	ParityCount: 	Parity count

`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("BucketName", true, false, "The dirName you want to list").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.StringOption(AddressID, "addr", "The practice user's addressid that you want to exec").WithDefault(""),
		cmds.IntOption(Policy, "pl", "Storage policy").WithDefault(dataformat.RsPolicy),
		cmds.BoolOption(Encryption, "encryp", "Encrytion or not").WithDefault(true),
		cmds.IntOption(DataCount, "dc", "data count").WithDefault(3),
		cmds.IntOption(ParityCount, "pc", "parity count, we suggest parity_count >= 2").WithDefault(2),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}
		userIns, ok := node.Inst.(*user.Info)
		if !ok {
			return ErrNotReady
		}
		var userid string
		addressid, found := req.Options[AddressID].(string)
		if addressid == "" || !found {
			userid = node.Identity.Pretty()
		} else {
			userid, err = address.GetIDFromAddress(addressid)
			if err != nil {
				return err
			}
		}
		policy, ok := req.Options[Policy].(int)
		if !ok || (policy != dataformat.MulPolicy && policy != dataformat.RsPolicy) {
			fmt.Println("input wrong policy.")
			return errWrongInput
		}
		dataCount, ok := req.Options[DataCount].(int)
		if !ok || dataCount <= 0 {
			fmt.Println("input wrong dataCount.")
			return errWrongInput
		}
		parityCount, ok := req.Options[ParityCount].(int)
		if !ok || parityCount <= 0 {
			fmt.Println("input wrong parityCount.")
			return errWrongInput
		}
		encrytion, ok := req.Options[Encryption].(bool)
		if !ok {
			fmt.Println("input wrong encrytion.")
			return errWrongInput
		}

		lfs := userIns.GetUser(userid)
		if lfs == nil || !lfs.Online() {
			return errLfsServiceNotReady
		}

		bucketOptions := dataformat.DefaultBucketOptions()
		bucketOptions.Policy = int32(policy)
		bucketOptions.DataCount = int32(dataCount)
		bucketOptions.ParityCount = int32(parityCount)
		if encrytion {
			bucketOptions.Encryption = 1
		} else {
			bucketOptions.Encryption = 0
		}

		bucket, err := lfs.CreateBucket(req.Context, req.Arguments[0], bucketOptions)
		if err != nil {
			return err
		}

		ctime := time.Unix(bucket.GetCTime(), 0).In(time.Local)
		bucketStat := BucketStat{
			Name:        bucket.Name,
			BucketID:    bucket.BucketID,
			Ctime:       ctime.Format(utils.SHOWTIME),
			Policy:      bucket.BOpts.Policy,
			DataCount:   bucket.BOpts.DataCount,
			ParityCount: bucket.BOpts.ParityCount,
			Encryption:  bucket.BOpts.Encryption,
		}
		return cmds.EmitOnce(res, &Buckets{
			Method:  "Create Bucket",
			Buckets: []BucketStat{bucketStat},
		})
	},
	Type: Buckets{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, bus *Buckets) error {
			_, err := fmt.Fprintf(w, "%s", bus)
			return err
		}),
	},
}

var lfsListBucketsCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "List buckets in lfs.",
		ShortDescription: `
'mefs lfs listbk' is a plumbing command for printing buckets in lfs.
It outputs the following to stdout:

	BucketName--BucketID--Policy-DataCount-ParityCount--Ctime

`,
	},

	Arguments: []cmds.Argument{},
	Options: []cmds.Option{
		cmds.StringOption(AddressID, "addr", "The practice user's addressid that you want to exec").WithDefault(""),
		cmds.StringOption(PrefixFilter, "Prefix can filter result").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}
		userIns, ok := node.Inst.(*user.Info)
		if !ok {
			return ErrNotReady
		}
		var userid string
		addressid, found := req.Options[AddressID].(string)
		if addressid == "" || !found {
			userid = node.Identity.Pretty()
		} else {
			userid, err = address.GetIDFromAddress(addressid)
			if err != nil {
				return err
			}
		}

		lfs := userIns.GetUser(userid)
		if lfs == nil || !lfs.Online() {
			return errLfsServiceNotReady
		}

		prefix := req.Options[PrefixFilter].(string)
		buckets, err := lfs.ListBuckets(req.Context, prefix)
		if err != nil {
			return err
		}
		bucketStats := &Buckets{
			Method: "List Buckets",
		}
		for _, bucket := range buckets {
			ctime := time.Unix(bucket.GetCTime(), 0).In(time.Local)
			bucketStat := BucketStat{
				Name:        bucket.Name,
				BucketID:    bucket.BucketID,
				Ctime:       ctime.Format(utils.SHOWTIME),
				Policy:      bucket.BOpts.Policy,
				DataCount:   bucket.BOpts.DataCount,
				ParityCount: bucket.BOpts.ParityCount,
				Encryption:  bucket.BOpts.Encryption,
			}
			bucketStats.Buckets = append(bucketStats.Buckets, bucketStat)
		}
		return cmds.EmitOnce(res, bucketStats)
	},
	Type: Buckets{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, bus *Buckets) error {
			_, err := fmt.Fprintf(w, "%s", bus)
			return err
		}),
	},
}

var lfsDeleteBucketCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Delete a bucket in lfs.",
		ShortDescription: `
'mefs lfs deletebk' is a plumbing command for deleting a bucket in lfs.(Not implement now)
It outputs the following to stdout:

	Method: 		Delete Bucket
	BucketName: 	bucket's name
	BucketID: 		bucket's ID
	Ctime: 			Creation Time
	Policy:			Erasure code or MultiReplication
	DataCount: 		Data count
	ParityCount: 	Parity count

`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("BucketName", true, false, "The dirName you want to list").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.StringOption(AddressID, "addr", "The practice user's addressid that you want to exec").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}
		userIns, ok := node.Inst.(*user.Info)
		if !ok {
			return ErrNotReady
		}
		var userid string
		addressid, found := req.Options[AddressID].(string)
		if addressid == "" || !found {
			userid = node.Identity.Pretty()
		} else {
			userid, err = address.GetIDFromAddress(addressid)
			if err != nil {
				return err
			}
		}

		lfs := userIns.GetUser(userid)
		if lfs == nil || !lfs.Online() {
			return errLfsServiceNotReady
		}

		bucket, err := lfs.DeleteBucket(req.Context, req.Arguments[0])
		if err != nil {
			return err
		}

		ctime := time.Unix(bucket.GetCTime(), 0).In(time.Local)
		bucketStat := BucketStat{
			Name:        bucket.Name,
			BucketID:    bucket.BucketID,
			Ctime:       ctime.Format(utils.SHOWTIME),
			Policy:      bucket.BOpts.Policy,
			DataCount:   bucket.BOpts.DataCount,
			ParityCount: bucket.BOpts.ParityCount,
			Encryption:  bucket.BOpts.Encryption,
		}
		return cmds.EmitOnce(res, &Buckets{
			Method:  "Delete Bucket",
			Buckets: []BucketStat{bucketStat},
		})
	},
	Type: Buckets{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, bus *Buckets) error {
			_, err := fmt.Fprintf(w, "%s", bus)
			return err
		}),
	},
}

var lfsListKeepersCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "List keepers in lfs.",
		ShortDescription: `
'mefs lfs list_keepers' is a plumbing command for printing buckets in lfs.
`,
	},

	Arguments: []cmds.Argument{},
	Options: []cmds.Option{
		cmds.StringOption(AddressID, "addr", "The practice user's addressid that you want to exec").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}
		userIns, ok := node.Inst.(*user.Info)
		if !ok {
			return ErrNotReady
		}
		var userid string
		addressid, found := req.Options[AddressID].(string)
		if addressid == "" || !found {
			userid = node.Identity.Pretty()
		} else {
			userid, err = address.GetIDFromAddress(addressid)
			if err != nil {
				return err
			}
		}

		lfs := userIns.GetUser(userid)
		lfsIns, ok := lfs.(*user.LfsInfo)
		if !ok {
			return errWrongInput
		}
		conkeepers, unconkeepers, _ := lfsIns.GetGroup().GetKeepers(req.Context, -1)
		keepers := make([]PeerState, len(unconkeepers)+len(conkeepers))
		for i := 0; i < len(conkeepers); i++ {
			keepers[i].PeerID = conkeepers[i]
			keepers[i].Connected = true
		}
		for i := 0; i < len(unconkeepers); i++ {
			keepers[i+len(conkeepers)].PeerID = unconkeepers[i]
			keepers[i+len(conkeepers)].Connected = false
		}
		list := &PeerList{
			Peers: keepers,
		}
		return cmds.EmitOnce(res, list)
	},
	Type: PeerList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, pl *PeerList) error {
			_, err := fmt.Fprintf(w, "%s", pl)
			return err
		}),
	},
}

var lfsListProviderrsCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "List keepers in lfs.",
		ShortDescription: `
'mefs lfs list_keepers' is a plumbing command for printing buckets in lfs.
`,
	},

	Arguments: []cmds.Argument{},
	Options: []cmds.Option{
		cmds.StringOption(AddressID, "addr", "The practice user's addressid that you want to exec").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}
		userIns, ok := node.Inst.(*user.Info)
		if !ok {
			return ErrNotReady
		}
		var userid string
		addressid, found := req.Options[AddressID].(string)
		if addressid == "" || !found {
			userid = node.Identity.Pretty()
		} else {
			userid, err = address.GetIDFromAddress(addressid)
			if err != nil {
				return err
			}
		}

		lfs := userIns.GetUser(userid)
		lfsIns, ok := lfs.(*user.LfsInfo)
		if !ok {
			return errWrongInput
		}
		conpro, unconpro, _ := lfsIns.GetGroup().GetProviders(req.Context, -1)
		providers := make([]PeerState, len(unconpro)+len(conpro))
		for i := 0; i < len(conpro); i++ {
			providers[i].PeerID = conpro[i]
			providers[i].Connected = true
		}
		for i := 0; i < len(unconpro); i++ {
			providers[i+len(conpro)].PeerID = unconpro[i]
			providers[i+len(conpro)].Connected = false
		}
		list := &PeerList{
			Peers: providers,
		}
		return cmds.EmitOnce(res, list)
	},
	Type: PeerList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, pl *PeerList) error {
			_, err := fmt.Fprintf(w, "%s", pl)
			return err
		}),
	},
}

var lfsListUsersCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "list users in this node.",
		ShortDescription: `
'mefs lfs list_users' is a plumbing command to list users in node.

`,
	},

	Arguments: []cmds.Argument{
		//暂时不需要输入
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if !node.OnlineMode() {
			return ErrNotOnline
		}
		userIns, ok := node.Inst.(*user.Info)
		if !ok {
			return ErrNotReady
		}
		users := userIns.GetAllUser()
		userAddrs := make([]string, len(users))
		for i, user := range users {
			addr, err := address.GetAddressFromID(user)
			if err != nil {
				continue
			}

			userAddrs[i] = user + "(" + addr.String() + ")"
		}
		list := &StringList{
			ChildLists: userAddrs,
		}
		return cmds.EmitOnce(res, list)
	},
	Type: StringList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, fl *StringList) error {
			_, err := fmt.Fprintf(w, "%s", fl)
			return err
		}),
	},
}

var lfsFsyncCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "flush lfs metablock to maintain consistency.",
		ShortDescription: `
'mefs lfs fsync' is a plumbing command to flush lfs.
 It outputs the following to stdout:

	error of Flush Success)

`,
	},

	Arguments: []cmds.Argument{
		//暂时不需要输入
	},
	Options: []cmds.Option{
		cmds.BoolOption(ForceFlush, "f", "Prefix can filter result").WithDefault(false),
		cmds.StringOption(AddressID, "addr", "The practice user's addressid that you want to exec").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}
		userIns, ok := node.Inst.(*user.Info)
		if !ok {
			return ErrNotReady
		}
		var userid string
		addressid, found := req.Options[AddressID].(string)
		if addressid == "" || !found {
			userid = node.Identity.Pretty()
		} else {
			userid, err = address.GetIDFromAddress(addressid)
			if err != nil {
				return err
			}
		}
		lfs := userIns.GetUser(userid)
		if lfs == nil || !lfs.Online() {
			return errLfsServiceNotReady
		}

		IsForce, ok := req.Options[ForceFlush].(bool)
		if !ok {
			fmt.Println("input wrong isForce.")
			return errWrongInput
		}
		err = lfs.Fsync(IsForce)
		if err != nil {
			return err
		}
		return cmds.EmitOnce(res, &StringList{
			ChildLists: []string{"Flush Success"},
		})
	},
	Type: StringList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, fl *StringList) error {
			_, err := fmt.Fprintf(w, "%s", fl)
			return err
		}),
	},
}

var lfsShowStorageCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "show the storage space used",
		ShortDescription: `
'
mefs lfs show_storage show the storage space used(kb)
`,
	},

	Arguments: []cmds.Argument{},
	Options: []cmds.Option{
		cmds.StringOption(AddressID, "addr", "The practice user's addressid that you want to exec").WithDefault(""),
		cmds.StringOption(PrefixFilter, "Prefix can filter result").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}
		userIns, ok := node.Inst.(*user.Info)
		if !ok {
			return ErrNotReady
		}
		var userid string
		addressid, found := req.Options[AddressID].(string)
		if addressid == "" || !found {
			userid = node.Identity.Pretty()
		} else {
			userid, err = address.GetIDFromAddress(addressid)
			if err != nil {
				return err
			}
		}
		prefix, found := req.Options[PrefixFilter].(string)
		if !found {
			prefix = ""
		}
		lfs := userIns.GetUser(userid)
		if lfs == nil || !lfs.Online() {
			return errLfsServiceNotReady
		}

		buckets, err := lfs.ListBuckets(req.Context, prefix)
		if err != nil {
			return err
		}
		var storageSize uint64
		for _, bucket := range buckets {
			storageSpace, err := lfs.ShowBucketStorage(req.Context, bucket.Name)
			if err != nil {
				return err
			}
			storageSize += storageSpace
		}
		FloatStorage := float64(storageSize)
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
		return cmds.EmitOnce(res, OutStorage)
	},
}

var lfsGetShareCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Get a share object to specified outputpath or current work directory",
		ShortDescription: `
'mefs lfs get_share' is a plumbing command for download a file or dir from lfs
 It outputs the following to stdout:

 	"GetShareObject success，Size: objectSize" or "error"

`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("ShareLink", true, false, "The share link"),
		cmds.StringArg("OutputName", true, false, "The file name you want to save."),
	},
	Options: []cmds.Option{
		cmds.StringOption(AddressID, "addr", "The practice user's addressid that you want to exec").WithDefault(""),
		cmds.StringOption(PassWord, "pwd", "The practice user's password that you want to exec").WithDefault(utils.DefaultPassword),
		cmds.StringOption(OutputPath, "o", "The path where the output should be stored."),
	},
	PreRun: func(req *cmds.Request, env cmds.Environment) error {
		outPath := getOutPath(req)
		fpath := filepath.Join(outPath, req.Arguments[1])
		_, err := os.Stat(fpath)
		if !os.IsNotExist(err) {
			return errors.New("The outpath already has file: " + req.Arguments[1])
		}
		return nil
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}

		userIns, ok := node.Inst.(*user.Info)
		if !ok {
			return ErrNotReady
		}
		var userid string
		addressid, found := req.Options[AddressID].(string)
		if addressid == "" || !found {
			userid = node.Identity.Pretty()
		} else {
			userid, err = address.GetIDFromAddress(addressid)
			if err != nil {
				return err
			}
		}

		pwd := req.Options[PassWord].(string)
		sk, err := fsrepo.GetPrivateKeyFromKeystore(userid, pwd)
		if err != nil {
			return err
		}

		piper, pipew := io.Pipe()
		bufw := bufio.NewWriterSize(pipew, user.DefaultBufSize)
		checkErrAndClosePipe := func(err error) error {
			if err != nil {
				err = pipew.CloseWithError(err)
				return err
			}
			err = pipew.Close()
			return err
		}
		var complete []user.CompleteFunc
		complete = append(complete, checkErrAndClosePipe)
		go userIns.GetShareObject(req.Context, bufw, complete, userid, sk, req.Arguments[0], user.DefaultOption())

		return res.Emit(piper)
	},
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(res cmds.Response, re cmds.ResponseEmitter) error {
			req := res.Request()
			v, err := res.Next()
			if err != nil {
				return err
			}

			outReader, ok := v.(io.Reader)
			if !ok {
				return e.New(e.TypeErr(outReader, v))
			}
			outPath := getOutPath(req)
			rootExists := true
			rootIsDir := false
			var fpath string
			if stat, err := os.Stat(outPath); err != nil && os.IsNotExist(err) {
				rootExists = false
			} else if err != nil {
				return err
			} else if stat.IsDir() {
				rootIsDir = true
			}
			if rootIsDir {
				fpath = path.Join(outPath, req.Arguments[1])
			} else if !rootExists {
				fpath = outPath
			} else {
				return errors.New("The outpath already has file: " + req.Arguments[1])
			}
			var file *os.File
			if _, err := os.Stat(fpath); err != nil && os.IsNotExist(err) {
				file, err = os.Create(fpath)
				if err != nil {
					return err
				}
			} else {
				return errors.New("The outpath already has file: " + req.Arguments[1])
			}
			//close这里不会报错么？
			defer file.Close()
			n, err := io.Copy(file, outReader)
			if err != nil {
				fmt.Println("Download failed - ", err)
				return err
			}
			fmt.Printf("GetObject to %s success，Size: %d\n", fpath, n)
			return nil
		},
	},
}

var lfsGenShareCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Gennerate share link of a lfs object.",
		ShortDescription: `
'mefs lfs gen_share' is a plumbing command to print information of a lfs file.
 It outputs the following to stdout:

	ShareLink   The ShareLink info of a 
`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("BucketName", true, false, "The Bucket's name that object in."),
		cmds.StringArg("ObjectName", true, false, "The Object's Name"),
	},
	Options: []cmds.Option{
		cmds.StringOption(AddressID, "addr", "The practice user's addressid that you want to exec").WithDefault(""),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		if !node.OnlineMode() {
			return ErrNotOnline
		}
		userIns, ok := node.Inst.(*user.Info)
		if !ok {
			return ErrNotReady
		}
		var userid string
		addressid, found := req.Options[AddressID].(string)
		if addressid == "" || !found {
			userid = node.Identity.Pretty()
		} else {
			userid, err = address.GetIDFromAddress(addressid)
			if err != nil {
				return err
			}
		}
		lfs := userIns.GetUser(userid)
		if lfs == nil || !lfs.Online() {
			return errLfsServiceNotReady
		}

		slink, err := lfs.(*user.LfsInfo).GenShareObject(req.Context, req.Arguments[0], req.Arguments[1], user.DefaultOption())
		if err != nil {
			return err
		}

		return cmds.EmitOnce(res, slink)
	},
}
