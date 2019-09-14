package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	cmds "github.com/ipfs/go-ipfs-cmds"
	peer "github.com/libp2p/go-libp2p-core/peer"
	notif "github.com/libp2p/go-libp2p-routing/notifications"
	cmdenv "github.com/memoio/go-mefs/core/commands/cmdenv"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
)

var ErrNotDHT = errors.New("routing service is not a DHT")

// TODO: Factor into `mefs dht` and `mefs routing`.
// Everything *except `query` goes into `mefs routing`.

var DhtCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "Issue commands directly through the DHT.",
		ShortDescription: ``,
	},

	Subcommands: map[string]*cmds.Command{
		"query":      queryDhtCmd,
		"findpeer":   findPeerDhtCmd,
		"get":        getValueDhtCmd,
		"put":        putValueDhtCmd,
		"putto":      putValuetoDhtCmd,
		"getfrom":    getValuefromDhtCmd,
		"liter":      literDhtCmd,
		"literfrom":  literFromDhtCmd,
		"append":     appendValueDhtCmd,
		"deletefrom": deleteFromDhtCmd,
	},
}

const (
	dhtVerboseOptionName = "v"
)

type queryEvent struct {
	ID    string
	Extra string
}

var queryDhtCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "Find the closest Peer IDs to a given Peer ID by querying the DHT.",
		ShortDescription: "Outputs a list of newline-delimited Peer IDs.",
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("peerID", true, true, "The peerID to run the query against."),
	},
	Options: []cmds.Option{
		cmds.BoolOption("verbose", dhtVerboseOptionName, "Print extra information."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		nd, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if nd.Routing == nil {
			return ErrNotDHT
		}

		id, err := peer.IDB58Decode(req.Arguments[0])
		if err != nil {
			return cmds.ClientError("invalid peer ID")
		}

		ctx, cancel := context.WithCancel(req.Context)
		ctx, events := notif.RegisterForQueryEvents(ctx)

		closestPeers, err := nd.Routing.(*dht.IpfsDHT).GetClosestPeers(ctx, string(id))
		if err != nil {
			cancel()
			return err
		}

		go func() {
			defer cancel()
			for p := range closestPeers {
				notif.PublishQueryEvent(ctx, &notif.QueryEvent{
					ID:   p,
					Type: notif.FinalPeer,
				})
			}
		}()

		for e := range events {
			if err := res.Emit(e); err != nil {
				return err
			}
		}

		return nil
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, out *notif.QueryEvent) error {
			pfm := pfuncMap{
				notif.PeerResponse: func(obj *notif.QueryEvent, out io.Writer, verbose bool) {
					for _, p := range obj.Responses {
						fmt.Fprintf(out, "%s\n", p.ID.Pretty())
					}
				},
			}
			verbose, _ := req.Options[dhtVerboseOptionName].(bool)
			printEvent(out, w, verbose, pfm)
			return nil
		}),
	},
	Type: notif.QueryEvent{},
}

var findPeerDhtCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "Find the multiaddresses associated with a Peer ID.",
		ShortDescription: "Outputs a list of newline-delimited multiaddresses.",
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("peerID", true, true, "The ID of the peer to search for."),
	},
	Options: []cmds.Option{
		cmds.BoolOption("verbose", dhtVerboseOptionName, "Print extra information."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		nd, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if nd.Routing == nil {
			return ErrNotOnline
		}

		pid, err := peer.IDB58Decode(req.Arguments[0])
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(req.Context)
		ctx, events := notif.RegisterForQueryEvents(ctx)

		go func() {
			defer cancel()
			pi, err := nd.Routing.FindPeer(ctx, pid)
			if err != nil {
				notif.PublishQueryEvent(ctx, &notif.QueryEvent{
					Type:  notif.QueryError,
					Extra: err.Error(),
				})
				return
			}

			notif.PublishQueryEvent(ctx, &notif.QueryEvent{
				Type:      notif.FinalPeer,
				Responses: []*peer.AddrInfo{&pi},
			})
		}()

		for e := range events {
			if err := res.Emit(e); err != nil {
				return err
			}
		}

		return nil
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, out *notif.QueryEvent) error {
			pfm := pfuncMap{
				notif.FinalPeer: func(obj *notif.QueryEvent, out io.Writer, verbose bool) {
					pi := obj.Responses[0]
					for _, a := range pi.Addrs {
						fmt.Fprintf(out, "%s\n", a)
					}
				},
			}

			verbose, _ := req.Options[dhtVerboseOptionName].(bool)
			printEvent(out, w, verbose, pfm)
			return nil
		}),
	},
	Type: notif.QueryEvent{},
}

var getValueDhtCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Given a key, query the routing system for its best value.",
		ShortDescription: `
Outputs the best value for the given key.

There may be several different values for a given key stored in the routing
system; in this context 'best' means the record that is most desirable. There is
no one metric for 'best': it depends entirely on the key type. For IPNS, 'best'
is the record that is both valid and has the highest sequence number (freshest).
Different key types can specify other 'best' rules.
`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("key", true, true, "The key to find a value for."),
	},
	Options: []cmds.Option{
		cmds.BoolOption("verbose", dhtVerboseOptionName, "Print extra information."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		nd, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if nd.Routing == nil {
			return ErrNotOnline
		}

		dhtkey := req.Arguments[0] //这里修改key的格式

		ctx, cancel := context.WithCancel(req.Context)
		ctx, events := notif.RegisterForQueryEvents(ctx)

		go func() {
			defer cancel()
			val, err := nd.Routing.GetValue(ctx, dhtkey)
			if err != nil {
				notif.PublishQueryEvent(ctx, &notif.QueryEvent{
					Type:  notif.QueryError,
					Extra: err.Error(),
				})
			} else {
				notif.PublishQueryEvent(ctx, &notif.QueryEvent{
					Type:  notif.Value,
					Extra: string(val),
				})
			}
		}()

		for e := range events {
			if err := res.Emit(e); err != nil {
				return err
			}
		}

		return nil
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, out *notif.QueryEvent) error {
			pfm := pfuncMap{
				notif.Value: func(obj *notif.QueryEvent, out io.Writer, verbose bool) {
					if verbose {
						fmt.Fprintf(out, "got value: '%s'\n", obj.Extra)
					} else {
						fmt.Fprintln(out, obj.Extra)
					}
				},
			}

			verbose, _ := req.Options[dhtVerboseOptionName].(bool)
			printEvent(out, w, verbose, pfm)

			return nil
		}),
	},
	Type: notif.QueryEvent{},
}

var getValuefromDhtCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "从指定节点get kv对",
		ShortDescription: `新加功能`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("key", true, false, "The key to find a value for."),
		cmds.StringArg("id", true, false, "The id to find a value from."),
	},
	Options: []cmds.Option{
		cmds.BoolOption("verbose", dhtVerboseOptionName, "Print extra information."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		nd, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if nd.Routing == nil {
			return ErrNotOnline
		}

		dhtkey := req.Arguments[0] //这里修改key的格式
		id := req.Arguments[1]     //这里取出id号

		ctx, cancel := context.WithCancel(req.Context)
		ctx, events := notif.RegisterForQueryEvents(ctx)

		go func() {
			defer cancel()
			val, err := nd.Routing.(*dht.IpfsDHT).CmdGetFrom(dhtkey, id)
			if err != nil {
				notif.PublishQueryEvent(ctx, &notif.QueryEvent{
					Type:  notif.QueryError,
					Extra: err.Error(),
				})
			} else {
				notif.PublishQueryEvent(ctx, &notif.QueryEvent{
					Type:  notif.Value,
					Extra: string(val),
				})
			}
		}()

		for e := range events {
			ne := &queryEvent{
				ID:    e.ID.Pretty(),
				Extra: e.Extra,
			}

			if e.Type != notif.Value {
				continue
			}

			if err := res.Emit(ne); err != nil {
				return err
			}
		}

		return nil
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, out *queryEvent) error {
			verbose, _ := req.Options[dhtVerboseOptionName].(bool)
			if verbose {
				fmt.Fprintf(w, "got value: '%s'\n", out.Extra)
			} else {
				fmt.Fprint(w, out.Extra)
			}

			return nil
		}),
	},
	Type: queryEvent{},
}

var literFromDhtCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "给一个前缀，从指定节点找到有该前缀的所有KV对",
		ShortDescription: `新加功能`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("key", true, false, "The prefix to find a value for."),
		cmds.StringArg("id", true, false, "The id to find a value from."),
	},
	Options: []cmds.Option{
		cmds.BoolOption("verbose", dhtVerboseOptionName, "Print extra information."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		nd, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if nd.Routing == nil {
			return ErrNotOnline
		}

		prefix := req.Arguments[0] //这里修改key的格式
		id := req.Arguments[1]     //取id

		ctx, cancel := context.WithCancel(req.Context)
		ctx, events := notif.RegisterForQueryEvents(ctx)
		go func() {
			defer cancel()
			val, err := nd.Routing.(*dht.IpfsDHT).CmdLiterFrom(prefix, id)
			if err != nil {
				notif.PublishQueryEvent(ctx, &notif.QueryEvent{
					Type:  notif.QueryError,
					Extra: err.Error(),
				})
			} else {
				notif.PublishQueryEvent(ctx, &notif.QueryEvent{
					Type:  notif.Value,
					Extra: string(val),
				})
			}
		}()

		for e := range events {
			if err := res.Emit(e); err != nil {
				return err
			}
		}

		return nil
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, out *notif.QueryEvent) error {
			pfm := pfuncMap{
				notif.Value: func(obj *notif.QueryEvent, out io.Writer, verbose bool) {
					if verbose {
						fmt.Fprintf(out, "got value: '%s'\n", obj.Extra)
					} else {
						fmt.Fprintln(out, obj.Extra)
					}
				},
			}

			verbose, _ := req.Options[dhtVerboseOptionName].(bool)
			printEvent(out, w, verbose, pfm)

			return nil
		}),
	},
	Type: notif.QueryEvent{},
}

var literDhtCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "给一个前缀，从本地找到有该前缀的所有KV对",
		ShortDescription: `新加功能`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("key", true, false, "The prefix to find a value by."),
	},
	Options: []cmds.Option{
		cmds.BoolOption("verbose", dhtVerboseOptionName, "Print extra information."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		nd, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if nd.Routing == nil {
			return ErrNotOnline
		}

		prefix := req.Arguments[0] //这里修改key的格式

		ctx, cancel := context.WithCancel(req.Context)
		ctx, events := notif.RegisterForQueryEvents(ctx)
		ctx = context.WithValue(ctx, "prefix", true) //在上下文中记录前缀查询标志
		go func() {
			defer cancel()
			val, err := nd.Routing.(*dht.IpfsDHT).CmdLiterFrom(prefix, "local")
			if err != nil {
				notif.PublishQueryEvent(ctx, &notif.QueryEvent{
					Type:  notif.QueryError,
					Extra: err.Error(),
				})
			} else {
				notif.PublishQueryEvent(ctx, &notif.QueryEvent{
					Type:  notif.Value,
					Extra: string(val),
				})
			}
		}()

		for e := range events {
			if err := res.Emit(e); err != nil {
				return err
			}
		}

		return nil
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, out *notif.QueryEvent) error {
			pfm := pfuncMap{
				notif.Value: func(obj *notif.QueryEvent, out io.Writer, verbose bool) {
					if verbose {
						fmt.Fprintf(out, "got value: '%s'\n", obj.Extra)
					} else {
						fmt.Fprintln(out, obj.Extra)
					}
				},
			}

			verbose, _ := req.Options[dhtVerboseOptionName].(bool)
			printEvent(out, w, verbose, pfm)

			return nil
		}),
	},
	Type: notif.QueryEvent{},
}

var putValueDhtCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Write a key/value pair to the routing system.",
		ShortDescription: `
Given a key of the form /foo/bar and a value of any form, this will write that
value to the routing system with that key.

Keys have two parts: a keytype (foo) and the key name (bar). 

Value is arbitrary text. Standard input can be used to provide value.

NOTE: A value may not exceed 2048 bytes.
`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("key", true, false, "The key to store the value at."),
		cmds.StringArg("value", true, false, "The value to store.").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.BoolOption("verbose", dhtVerboseOptionName, "Print extra information."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		nd, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if nd.Routing == nil {
			return ErrNotOnline
		}

		key := req.Arguments[0] //这里获得key

		data := req.Arguments[1]

		ctx, cancel := context.WithCancel(req.Context)
		ctx, events := notif.RegisterForQueryEvents(ctx)

		go func() {
			defer cancel()
			err := nd.Routing.PutValue(ctx, key, []byte(data))
			if err != nil {
				notif.PublishQueryEvent(ctx, &notif.QueryEvent{
					Type:  notif.QueryError,
					Extra: err.Error(),
				})
			}
		}()

		for e := range events {
			if err := res.Emit(e); err != nil {
				return err
			}
		}

		return nil
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, out *notif.QueryEvent) error {
			pfm := pfuncMap{
				notif.FinalPeer: func(obj *notif.QueryEvent, out io.Writer, verbose bool) {
					if verbose {
						fmt.Fprintf(out, "* closest peer %s\n", obj.ID)
					}
				},
				notif.Value: func(obj *notif.QueryEvent, out io.Writer, verbose bool) {
					fmt.Fprintf(out, "%s\n", obj.ID.Pretty())
				},
			}

			verbose, _ := req.Options[dhtVerboseOptionName].(bool)

			printEvent(out, w, verbose, pfm)

			return nil
		}),
	},
	Type: notif.QueryEvent{},
}

var putValuetoDhtCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "将k/v对put到指定节点上",
		ShortDescription: `新加功能`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("key", true, false, "The key to store the value at."),
		cmds.StringArg("value", true, false, "The value to store.").EnableStdin(),
		cmds.StringArg("id", true, false, "The value to store the value to."),
	},
	Options: []cmds.Option{
		cmds.BoolOption("verbose", dhtVerboseOptionName, "Print extra information."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		nd, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if nd.Routing == nil {
			return ErrNotOnline
		}

		key := req.Arguments[0] //这里对key的格式进行修改

		data := req.Arguments[1]
		id := req.Arguments[2]

		ctx, cancel := context.WithCancel(req.Context)
		ctx, events := notif.RegisterForQueryEvents(ctx)
		ctx = context.WithValue(ctx, "id", id) //上下文中添加id信息

		go func() {
			defer cancel()
			err := nd.Routing.(*dht.IpfsDHT).CmdPutTo(key, data, id)
			if err != nil {
				notif.PublishQueryEvent(ctx, &notif.QueryEvent{
					Type:  notif.QueryError,
					Extra: err.Error(),
				})
			}
		}()

		for e := range events {
			if err := res.Emit(e); err != nil {
				return err
			}
		}

		return nil
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, out *notif.QueryEvent) error {
			pfm := pfuncMap{
				notif.FinalPeer: func(obj *notif.QueryEvent, out io.Writer, verbose bool) {
					if verbose {
						fmt.Fprintf(out, "* closest peer %s\n", obj.ID)
					}
				},
				notif.Value: func(obj *notif.QueryEvent, out io.Writer, verbose bool) {
					fmt.Fprintf(out, "%s\n", obj.ID.Pretty())
				},
			}

			verbose, _ := req.Options[dhtVerboseOptionName].(bool)

			printEvent(out, w, verbose, pfm)

			return nil
		}),
	},
	Type: notif.QueryEvent{},
}

var appendValueDhtCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "将k/v对put到指定节点上",
		ShortDescription: `新加功能`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("key", true, false, "The key to store the value at."),
		cmds.StringArg("value", true, false, "The value to store.").EnableStdin(),
	},
	Options: []cmds.Option{
		cmds.BoolOption("verbose", dhtVerboseOptionName, "Print extra information."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		nd, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if nd.Routing == nil {
			return ErrNotOnline
		}

		key := req.Arguments[0] //这里对key的格式进行修改

		data := req.Arguments[1]

		ctx, cancel := context.WithCancel(req.Context)
		ctx, events := notif.RegisterForQueryEvents(ctx)
		ctx = context.WithValue(ctx, "append", true) //上下文中append标志

		go func() {
			defer cancel()
			err := nd.Routing.PutValue(ctx, key, []byte(data))
			if err != nil {
				notif.PublishQueryEvent(ctx, &notif.QueryEvent{
					Type:  notif.QueryError,
					Extra: err.Error(),
				})
			}
		}()

		for e := range events {
			if err := res.Emit(e); err != nil {
				return err
			}
		}

		return nil
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, out *notif.QueryEvent) error {
			pfm := pfuncMap{
				notif.FinalPeer: func(obj *notif.QueryEvent, out io.Writer, verbose bool) {
					if verbose {
						fmt.Fprintf(out, "* closest peer %s\n", obj.ID)
					}
				},
				notif.Value: func(obj *notif.QueryEvent, out io.Writer, verbose bool) {
					fmt.Fprintf(out, "%s\n", obj.ID.Pretty())
				},
			}

			verbose, _ := req.Options[dhtVerboseOptionName].(bool)

			printEvent(out, w, verbose, pfm)

			return nil
		}),
	},
	Type: notif.QueryEvent{},
}

type printFunc func(obj *notif.QueryEvent, out io.Writer, verbose bool)
type pfuncMap map[notif.QueryEventType]printFunc

func printEvent(obj *notif.QueryEvent, out io.Writer, verbose bool, override pfuncMap) {
	if verbose {
		fmt.Fprintf(out, "%s: ", time.Now().Format("15:04:05.000"))
	}

	if override != nil {
		if pf, ok := override[obj.Type]; ok {
			pf(obj, out, verbose)
			return
		}
	}

	switch obj.Type {
	case notif.SendingQuery:
		if verbose {
			fmt.Fprintf(out, "* querying %s\n", obj.ID)
		}
	case notif.Value:
		if verbose {
			fmt.Fprintf(out, "got value: '%s'\n", obj.Extra)
		} else {
			fmt.Fprint(out, obj.Extra)
		}
	case notif.PeerResponse:
		if verbose {
			fmt.Fprintf(out, "* %s says use ", obj.ID)
			for _, p := range obj.Responses {
				fmt.Fprintf(out, "%s ", p.ID)
			}
			fmt.Fprintln(out)
		}
	case notif.QueryError:
		if verbose {
			fmt.Fprintf(out, "error: %s\n", obj.Extra)
		}
	case notif.DialingPeer:
		if verbose {
			fmt.Fprintf(out, "dialing peer: %s\n", obj.ID)
		}
	case notif.AddingPeer:
		if verbose {
			fmt.Fprintf(out, "adding peer to query: %s\n", obj.ID)
		}
	case notif.FinalPeer:
	default:
		if verbose {
			fmt.Fprintf(out, "unrecognized event type: %d\n", obj.Type)
		}
	}
}

var deleteFromDhtCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "在provider上删除块，用在测试挑战修复的时候",
		ShortDescription: `新加功能`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("key", true, false, "The value to store.").EnableStdin(),
		cmds.StringArg("to", true, false, "The value to store.").EnableStdin(),
	},
	Options: []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		nd, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if nd.Routing == nil {
			return ErrNotOnline
		}

		key := req.Arguments[0]
		to := req.Arguments[1]

		go func() {
			_, err = nd.Routing.(*dht.IpfsDHT).SendMetaRequest(key, "", to, "deletefrom")
			if err != nil {
				fmt.Println("delete block error :", err)
				return
			}
		}()

		return res.Emit("ok")
	},
}
