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
	"github.com/memoio/go-mefs/role"
	datastore "github.com/memoio/go-mefs/source/go-datastore"
	"github.com/memoio/go-mefs/storageNode/provider"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

type pInfoOutput struct {
	Address         string
	Balance         *big.Int
	DepositCapacity uint64
	UsedCapacity    uint64
	OfferAddress    string
	OfferCapacity   int64
	OfferPrice      *big.Int
	OfferDuration   int64
	OfferStartTime  string
	TotalIncome     *big.Int
	DownloadIncome  *big.Int
	StorageIncome   *big.Int
	LastDayIncome   *big.Int
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

		localID := node.Identity.Pretty()
		localAddr, err := address.GetAddressFromID(localID)
		if err != nil {
			return err
		}

		var depositCapacity int64
		var usedCapacity uint64
		balance, err := contracts.QueryBalance(localAddr.String())
		if err != nil {
			return err
		}

		oItem, err := role.GetLatestOffer(localID, localID)
		if err != nil {
			return err
		}
		providerIns, ok := node.Inst.(*provider.Info)

		if !ok || !providerIns.Online() { //service is not ready, 从链上获取depositCapacity
			providerItem, err := role.GetProviderInfo(node.Identity.Pretty(), node.Identity.Pretty())
			if err != nil {
				return err
			}
			depositCapacity = providerItem.Capacity

			usedCapacity, err = datastore.DiskUsage(node.Data.DataStore())
			if err != nil {
				return err
			}
			fmt.Println("if you want to see the real income, please run 'mefs-provider daemon' first")
		} else {
			depositCapacity, usedCapacity = providerIns.GetStorageInfo()

			/*
				// need test
				ukAddress, channelAddress := providerIns.GetIncomeInfo()

				totalIncome, dailyIncome, err = contracts.GetStorageIncome(ukAddress, localAddr) //income for storage
				if err != nil {
					return err
				}
				t, d, err := contracts.GetDownloadIncome(channelAddress, localAddr)
				if err != nil {
					return err
				}

				totalIncome.Add(totalIncome, t)
				dailyIncome.Add(dailyIncome, d)
			*/
		}

		output := &pInfoOutput{
			Address:         localAddr.String(),
			DepositCapacity: uint64(depositCapacity) * 1024 * 1024,
			UsedCapacity:    usedCapacity,
			OfferAddress:    oItem.OfferID,
			OfferDuration:   oItem.Duration,
			OfferStartTime:  time.Unix(oItem.CreateDate, 0).In(time.Local).Format(utils.SHOWTIME),
			OfferCapacity:   oItem.Capacity,
			OfferPrice:      oItem.Price,
			Balance:         balance,
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
