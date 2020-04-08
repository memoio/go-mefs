package commands

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	cmds "github.com/ipfs/go-ipfs-cmds"
	ic "github.com/libp2p/go-libp2p-core/crypto"
	peer "github.com/libp2p/go-libp2p-core/peer"
	pstore "github.com/libp2p/go-libp2p-core/peerstore"
	kb "github.com/libp2p/go-libp2p-kbucket"
	identify "github.com/libp2p/go-libp2p/p2p/protocol/identify"
	core "github.com/memoio/go-mefs/core"
	cmdenv "github.com/memoio/go-mefs/core/commands/cmdenv"
	id "github.com/memoio/go-mefs/crypto/identity"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

const offlineIdErrorMessage = `'mefs id' currently cannot query information on remote
peers without a running daemon; we are working to fix this.
In the meantime, if you want to query remote peers using 'mefs id',
please run the daemon:

	mefs daemon &
    mefs id
`

type IdOutput struct {
	NetworkAddr  string
	AccountAddr  string
	PublicKey    string
	Addresses    []string
	AgentVersion string
}

const (
	formatOptionName = "format"
)

var IDCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Show mefs node id info.",
		ShortDescription: `
Prints out information about the specified peer.
If no peer is specified, prints out information for local peers.

'mefs id' supports the format option for output with the following keys:
<id> : The peers id.
<peerAddr>: The peers address.
<aver>: Agent version.
<pubkey>: Public key.
<addrs>: Addresses (newline delimited).
`,
	},
	Arguments: []cmds.Argument{
		cmds.StringArg("peerid", false, false, "Peer.ID of node to look up."),
	},
	Options: []cmds.Option{
		cmds.StringOption("password", "pwd", "the password is used to encrypt the PrivateKey").WithDefault(utils.DefaultPassword),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		n, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		pwd, ok := req.Options["password"].(string)
		if ok {
			n.SetPassWord(pwd)
		}

		var id peer.ID
		if len(req.Arguments) > 0 {
			var err error
			id, err = peer.IDB58Decode(req.Arguments[0])
			if err != nil {
				return fmt.Errorf("invalid peer id")
			}
		} else {
			id = n.Identity
		}

		if id == n.Identity {
			output, err := printSelf(n)
			if err != nil {
				return err
			}
			return cmds.EmitOnce(res, output)
		}

		// TODO handle offline mode with polymorphism instead of conditionals
		if !n.OnlineMode() {
			return errors.New(offlineIdErrorMessage)
		}

		p, err := n.Routing.FindPeer(req.Context, id)
		if err == kb.ErrLookupFailure {
			return errors.New(offlineIdErrorMessage)
		}
		if err != nil {
			return err
		}

		output, err := printPeer(n.Peerstore, p.ID)
		if err != nil {
			return err
		}
		return cmds.EmitOnce(res, output)
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, out *IdOutput) error {
			marshaled, err := json.MarshalIndent(out, "", "\t")
			if err != nil {
				return err
			}
			marshaled = append(marshaled, byte('\n'))
			fmt.Fprintln(w, string(marshaled))
			return nil
		}),
	},
	Type: IdOutput{},
}

func printPeer(ps pstore.Peerstore, p peer.ID) (interface{}, error) {
	if p == "" {
		return nil, errors.New("attempted to print nil peer")
	}

	info := new(IdOutput)
	info.NetworkAddr = p.Pretty()
	tmpAddr, err := address.GetAddressFromID(p.Pretty())
	if err != nil {
		return nil, err
	}
	info.AccountAddr = tmpAddr.String()

	if pk := ps.PubKey(p); pk != nil {
		pkb, err := ic.MarshalPublicKey(pk)
		if err != nil {
			return nil, err
		}
		info.PublicKey = base64.StdEncoding.EncodeToString(pkb)
	}

	for _, a := range ps.Addrs(p) {
		info.Addresses = append(info.Addresses, a.String())
	}

	if v, err := ps.Get(p, "AgentVersion"); err == nil {
		if vs, ok := v.(string); ok {
			info.AgentVersion = vs
		}
	}

	return info, nil
}

// printing self is special cased as we get values differently.
func printSelf(node *core.MefsNode) (interface{}, error) {
	info := new(IdOutput)
	info.NetworkAddr = node.Identity.Pretty()
	tmpAddr, err := address.GetAddressFromID(node.Identity.Pretty())
	if err != nil {
		return nil, err
	}
	info.AccountAddr = tmpAddr.String()

	if node.PrivateKey == "" {
		if err := node.LoadPrivateKey(); err != nil {
			return nil, err
		}
	}

	pkb, err := id.GetCompressPubByte(node.PrivateKey)
	if err != nil {
		return nil, err
	}
	info.PublicKey = base64.StdEncoding.EncodeToString(pkb)

	if node.PeerHost != nil {
		for _, a := range node.PeerHost.Addrs() {
			s := a.String() + "/p2p/" + info.NetworkAddr
			info.Addresses = append(info.Addresses, s)
		}
	}
	info.AgentVersion = identify.ClientVersion
	return info, nil
}
