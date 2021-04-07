package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strings"
	"time"

	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/memoio/go-mefs/core/commands/cmdenv"
	"github.com/memoio/go-mefs/repo/fsrepo"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/storageNode/provider"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

type pInfoOutput struct {
	Wallet          string
	StartTime       string
	UpTime          string
	ReadyForService bool
	PublicNetwork   string
	PublicReachable bool
	Balance         string
	PledgeBytes     string
	UsedBytes       string
	PosBytes        string
	LocalFreeBytes  string
	OfferAddress    string
	OfferCapacity   string
	OfferPrice      string
	OfferDuration   string
	OfferStartTime  string
	TotalIncome     string
	StorageIncome   string
	DownloadIncome  string
	PosIncome       string
}

type StringList struct {
	ChildLists []string
}

func (list StringList) String() string {
	var buffer bytes.Buffer
	for i := 0; i < len(list.ChildLists); i++ {
		buffer.WriteString(list.ChildLists[i])
		buffer.WriteString("\n")
	}
	return buffer.String()
}

var InfoCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "list infomations",
		ShortDescription: `
`,
	},

	Subcommands: map[string]*cmds.Command{
		"self":  SelfCmd, //命令行操作写法示例
		"users": userCmd,
		"group": gpInfoCmd,
	},
}

var SelfCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Interact with provider.",
		ShortDescription: `
'mefs-provider info' is a plumbing command used to manipulate provider service.
`,
	},

	Arguments: []cmds.Argument{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if !node.OnlineMode() {
			return ErrNotOnline
		}

		localID := node.Identity.Pretty()
		localAddr, err := address.GetAddressFromID(localID)
		if err != nil {
			return err
		}

		var depositCapacity, usedCapacity, posCapacity uint64
		balance, err := role.QueryBalance(localID)
		if err != nil {
			return err
		}

		oItem, err := role.GetLatestOffer(localID, localID)
		if err != nil {
			return err
		}

		ti := new(big.Int)
		di := new(big.Int)
		si := new(big.Int)
		pi := new(big.Int)

		var eAddr string
		ready := false
		reachable := false
		stime := time.Unix(0, 0)
		uTime := int64(0)
		providerIns, ok := node.Inst.(*provider.Info)
		if !ok { //service is not ready, 从链上获取depositCapacity
			providerItem, err := role.GetProviderInfo(node.Identity.Pretty(), node.Identity.Pretty())
			if err != nil {
				return err
			}
			depositCapacity = uint64(providerItem.Capacity) * 1024 * 1024

			rootpath, err := fsrepo.BestKnownPath()
			if err != nil {
				return err
			}

			usedCapacity, err = utils.GetDirSize(rootpath)
			if err != nil {
				return err
			}

			fmt.Println("if you want to see the real income, please run 'mefs-provider daemon' first")
		} else {
			depositCapacity = providerIns.StorageTotal
			usedCapacity = providerIns.StorageUsed
			posCapacity = providerIns.StoragePosUsed

			ti = providerIns.TotalIncome
			si = providerIns.StorageIncome
			di = providerIns.ReadIncome
			pi = providerIns.PosIncome
			eAddr, _ = providerIns.GetPublicAddress()
			ea := strings.Split(eAddr, "/")
			if len(ea) >= 5 {
				ipa := ea[2] + ":" + ea[4]
				reachable = utils.IsReachable(ipa)
			}
			ready = providerIns.Online()
			stime = providerIns.StartTime
			uTime = time.Now().Unix() - stime.Unix()
		}

		offerAddr, err := address.GetAddressFromID(oItem.OfferID)
		if err != nil {
			return err
		}

		lsinfo, err := role.GetDiskSpaceInfo()
		if err != nil {
			return err
		}

		output := &pInfoOutput{
			Wallet:          localAddr.String(),
			StartTime:       stime.In(time.Local).Format(utils.SHOWTIME),
			UpTime:          utils.FormatSecond(uTime),
			ReadyForService: ready,
			PublicNetwork:   eAddr,
			PublicReachable: reachable,
			PledgeBytes:     utils.FormatBytes(int64(depositCapacity)),
			UsedBytes:       utils.FormatBytes(int64(usedCapacity)),
			PosBytes:        utils.FormatBytes(int64(posCapacity)),
			LocalFreeBytes:  utils.FormatBytes(int64(lsinfo.Free)),
			OfferAddress:    offerAddr.String(),
			OfferDuration:   utils.FormatSecond(oItem.Duration),
			OfferStartTime:  time.Unix(oItem.CreateDate, 0).In(time.Local).Format(utils.SHOWTIME),
			OfferCapacity:   utils.FormatBytes(int64(oItem.Capacity * 1024 * 1024)),
			OfferPrice:      utils.FormatStorePrice(oItem.Price),
			Balance:         utils.FormatWei(balance),
			TotalIncome:     utils.FormatWei(ti),
			DownloadIncome:  utils.FormatWei(di),
			StorageIncome:   utils.FormatWei(si),
			PosIncome:       utils.FormatWei(pi),
		}
		return cmds.EmitOnce(res, output)
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, out *pInfoOutput) error {
			marshaled, err := json.MarshalIndent(out, "", "\t")
			if err != nil {
				return err
			}
			marshaled = append(marshaled, byte('\n'))
			fmt.Fprintln(w, string(marshaled))
			return nil
		}),
	},
	Type: pInfoOutput{},
}

var gpInfoCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "show group information.",
		ShortDescription: `
		'mefs-provider groupInfo' is a plumbing command used to show information of group which provider participate.
		`,
	},
	Arguments: []cmds.Argument{
		cmds.StringArg("uid", true, false, "The user's id"),
		cmds.StringArg("qid", true, false, "The user's query id"),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if !node.OnlineMode() {
			return ErrNotOnline
		}

		providerIns, ok := node.Inst.(*provider.Info)
		if !ok || providerIns == nil {
			return role.ErrServiceNotReady
		}

		uid := req.Arguments[0]
		qid := req.Arguments[1]

		output, err := providerIns.GetGroupInfoOutput(uid, qid)
		if err != nil {
			return err
		}

		return cmds.EmitOnce(res, output)
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, out *provider.GInfoOutput) error {
			marshaled, err := json.MarshalIndent(out, "", "\t")
			if err != nil {
				return err
			}
			marshaled = append(marshaled, byte('\n'))
			fmt.Fprintln(w, string(marshaled))
			return nil
		}),
	},
	Type: provider.GInfoOutput{},
}

var userCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "list all users this provider served",
		ShortDescription: `list all users that this provider has offered storage resource to`,
	},

	Arguments: []cmds.Argument{},
	Options:   []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		n, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}
		providerIns, ok := n.Inst.(*provider.Info)
		if !ok {
			return role.ErrServiceNotReady
		}

		output := providerIns.ShowUserInfo()

		list := &StringList{
			ChildLists: output,
		}

		return cmds.EmitOnce(res, list)
	},
	Type: StringList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, list *StringList) error {
			_, err := fmt.Fprintf(w, "%s", list)
			return err
		}),
	},
}
