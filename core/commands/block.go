package commands

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	cmds "github.com/ipfs/go-ipfs-cmds"
	peer "github.com/libp2p/go-libp2p-core/peer"
	mh "github.com/multiformats/go-multihash"

	cmdenv "github.com/memoio/go-mefs/core/commands/cmdenv"
	e "github.com/memoio/go-mefs/core/commands/e"
	"github.com/memoio/go-mefs/core/coreapi/interface/options"
)

type BlockStat struct {
	Key  string
	Size int
}

func (bs BlockStat) String() string {
	return fmt.Sprintf("Key: %s\nSize: %d\n", bs.Key, bs.Size)
}

var BlockCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Interact with raw MEFS blocks.",
		ShortDescription: `
'mefs block' is a plumbing command used to manipulate raw MEFS blocks.
Reads from stdin or writes to stdout, and <key> is a base58 encoded
multihash.
`,
	},

	Subcommands: map[string]*cmds.Command{
		"stat":    blockStatCmd,
		"get":     blockGetCmd,
		"getfrom": blockGetFromCmd,
		"put":     blockPutCmd,
		"putto":   blockPutToCmd,
	},
}

var blockStatCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Print information of a raw MEFS block.",
		ShortDescription: `
'mefs block stat' is a plumbing command for retrieving information
on raw MEFS blocks. It outputs the following to stdout:

	Key  - the base58 encoded multihash
	Size - the size of the block in bytes

`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("key", true, false, "The base58 multihash of an existing block to stat.").EnableStdin(),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env)
		if err != nil {
			return err
		}

		p := req.Arguments[0]

		b, err := api.Block().Stat(req.Context, p)
		if err != nil {
			return err
		}

		return cmds.EmitOnce(res, &BlockStat{
			Key:  b.Cid().String(),
			Size: b.Size(),
		})
	},
	Type: BlockStat{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, bs *BlockStat) error {
			_, err := fmt.Fprintf(w, "%s", bs)
			return err
		}),
	},
}

var blockGetCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Get a raw MEFS block.",
		ShortDescription: `
'mefs block get' is a plumbing command for retrieving raw MEFS blocks.
It outputs to stdout, and <key> is a base58 encoded multihash.
`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("key", true, false, "The base58 multihash of an existing block to get.").EnableStdin(),
	},

	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env)
		if err != nil {
			return err
		}

		p := req.Arguments[0]

		r, err := api.Block().Get(req.Context, p)
		if err != nil {
			return err
		}

		return res.Emit(r)
	},
}

var blockGetFromCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Get a raw MEFS block from a given node.",
		ShortDescription: `
'mefs block getfrom' is a plumbing command for retrieving raw MEFS blocks from a given node.
It outputs to stdout, and <key> is a base58 encoded multihash, peer ID identifies a node.
`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("key", true, false, "The base58 multihash of an existing block to get.").EnableStdin(),
		cmds.StringArg("Peer ID", true, false, "the Peer ID of a given Node").EnableStdin(),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env)
		if err != nil {
			return err
		}

		p := req.Arguments[0]
		if err != nil {
			return err
		}
		peerid := req.Arguments[1]
		r, err := api.Block().GetFrom(req.Context, p, peerid)
		if err != nil {
			return err
		}
		h := md5.New()
		tmpByte, err := ioutil.ReadAll(r)
		if err != nil {
			fmt.Println("read reader error :", err)
			return err
		}
		h.Write(tmpByte)

		return res.Emit(hex.EncodeToString(h.Sum(nil)))
	},
}

const (
	blockFormatOptionName = "format"
	mhtypeOptionName      = "mhtype"
	mhlenOptionName       = "mhlen"
)

var blockPutCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Store input as an MEFS block by using new CID.",
		ShortDescription: `
'mefs block put' is a plumbing command for storing raw MEFS blocks.
It reads from stdin, and <key> is a base58 encoded multihash.
`,
	},

	Arguments: []cmds.Argument{
		cmds.FileArg("data", true, false, "The data to be stored as an MEFS block.").EnableStdin(),
		cmds.StringArg("New CID", true, false, "A symbol that you want to identify a block."),
	},
	Options: []cmds.Option{
		cmds.StringOption(blockFormatOptionName, "f", "cid format for blocks to be created with."),
		cmds.StringOption(mhtypeOptionName, "multihash hash function").WithDefault("sha2-256"),
		cmds.IntOption(mhlenOptionName, "multihash hash length").WithDefault(-1),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env)
		if err != nil {
			return err
		}
		node, err := cmdenv.GetNode(env)
		if err != nil {
			println("Error, can't find the node")
			return err
		}
		id := peer.IDB58Encode(node.Identity)

		file, err := cmdenv.GetFileArg(req.Files.Entries())
		if err != nil {
			return err
		}

		mhtype, _ := req.Options[mhtypeOptionName].(string)
		mhtval, ok := mh.Names[mhtype]
		if !ok {
			return fmt.Errorf("unrecognized multihash function: %s", mhtype)
		}

		mhlen, ok := req.Options[mhlenOptionName].(int)
		if !ok {
			return errors.New("missing option \"mhlen\"")
		}

		format, formatSet := req.Options[blockFormatOptionName].(string)
		if !formatSet {
			if mhtval != mh.SHA2_256 || (mhlen != -1 && mhlen != 32) {
				format = "protobuf"
			} else {
				format = "v0"
			}
		}
		ncid := req.Arguments[0]
		ncid = id + "_" + ncid

		p, err := api.Block().Put(req.Context, file, ncid, options.Block.Hash(mhtval, mhlen), options.Block.Format(format))
		if err != nil {
			return err
		}

		return cmds.EmitOnce(res, &BlockStat{
			Key:  p.Cid().String(),
			Size: p.Size(),
		})
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeEncoder(func(req *cmds.Request, w io.Writer, v interface{}) error {
			bs, ok := v.(*BlockStat)
			if !ok {
				return e.TypeErr(bs, v)
			}
			_, err := fmt.Fprintf(w, "%s\n", bs.Key)
			return err
		}),
	},
	Type: BlockStat{},
}

var blockPutToCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Send an MEFS block to peer partner.",
		ShortDescription: `
'mefs block putto' is a plumbing command for send an MEFS blocks to partner.
It reads from stdin, and <key> is a base58 encoded multihash.
`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("block's CID", true, false, "the block will be sent to peer.").EnableStdin(),
		cmds.StringArg("peer's PID", true, false, "the id is the received node's id.").EnableStdin(),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env)
		if err != nil {
			return err
		}

		var Cid string = req.Arguments[0]
		var Pid string = req.Arguments[1]
		p, err := api.Block().Putto(Cid, Pid, req.Context)
		if err != nil {
			return err
		}

		return cmds.EmitOnce(res, &BlockStat{
			Key:  p.Cid().String(),
			Size: p.Size(),
		})
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeEncoder(func(req *cmds.Request, w io.Writer, v interface{}) error {
			bs, ok := v.(*BlockStat)
			if !ok {
				return e.TypeErr(bs, v)
			}
			_, err := fmt.Fprintf(w, "%s\n", bs.Key)
			return err
		}),
	},
	Type: BlockStat{},
}

const (
	forceOptionName      = "force"
	blockQuietOptionName = "quiet"
)
