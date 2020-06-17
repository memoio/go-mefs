package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"time"

	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/core/commands/cmdenv"
	"github.com/memoio/go-mefs/repo/fsrepo"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/storageNode/provider"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

type pInfoOutput struct {
	Address        string
	PublicNetwork  string
	Balance        string
	PledgeBytes    string
	UsedBytes      string
	PosBytes       string
	LocalFreeBytes string
	OfferAddress   string
	OfferCapacity  string
	OfferPrice     string
	OfferDuration  string
	OfferStartTime string
	TotalIncome    string
	StorageIncome  string
	DownloadIncome string
	PosIncome      string
}

var InfoCmd = &cmds.Command{
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
		balance, err := contracts.QueryBalance(localAddr.String())
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

		providerIns, ok := node.Inst.(*provider.Info)
		if !ok || !providerIns.Online() { //service is not ready, 从链上获取depositCapacity
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
			eAddr, _ = providerIns.GetIPAddress()
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
			Address:        localAddr.String(),
			PublicNetwork:  eAddr,
			PledgeBytes:    utils.FormatBytes(int64(depositCapacity)),
			UsedBytes:      utils.FormatBytes(int64(usedCapacity)),
			PosBytes:       utils.FormatBytes(int64(posCapacity)),
			LocalFreeBytes: utils.FormatBytes(int64(lsinfo.Free)),
			OfferAddress:   offerAddr.String(),
			OfferDuration:  utils.FormatSecond(oItem.Duration),
			OfferStartTime: time.Unix(oItem.CreateDate, 0).In(time.Local).Format(utils.SHOWTIME),
			OfferCapacity:  utils.FormatBytes(int64(oItem.Capacity * 1024 * 1024)),
			OfferPrice:     utils.FormatStorePrice(oItem.Price),
			Balance:        utils.FormatWei(balance),
			TotalIncome:    utils.FormatWei(ti),
			DownloadIncome: utils.FormatWei(di),
			StorageIncome:  utils.FormatWei(si),
			PosIncome:      utils.FormatWei(pi),
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
