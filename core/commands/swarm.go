package commands

import (
	"errors"
	"fmt"
	"io"
	"path"
	"sort"

	iaddr "github.com/ipfs/go-ipfs-addr"
	cmds "github.com/ipfs/go-ipfs-cmds"
	inet "github.com/libp2p/go-libp2p-core/network"
	peer "github.com/libp2p/go-libp2p-core/peer"
	swarm "github.com/libp2p/go-libp2p-swarm"
	commands "github.com/memoio/go-mefs/commands"
	config "github.com/memoio/go-mefs/config"
	cmdenv "github.com/memoio/go-mefs/core/commands/cmdenv"
	repo "github.com/memoio/go-mefs/repo"
	fsrepo "github.com/memoio/go-mefs/repo/fsrepo"
	ma "github.com/multiformats/go-multiaddr"
	mafilter "github.com/whyrusleeping/multiaddr-filter"
)

type stringList struct {
	Strings []string
}

type addrMap struct {
	Addrs map[string][]string
}

var SwarmCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Interact with the swarm.",
		ShortDescription: `
'mefs swarm' is a tool to manipulate the network swarm. The swarm is the
component that opens, listens for, and maintains connections to other
mefs peers in the internet.
`,
	},
	Subcommands: map[string]*cmds.Command{
		"addrs":      swarmAddrsCmd,
		"connect":    swarmConnectCmd,
		"disconnect": swarmDisconnectCmd,
		"filters":    swarmFiltersCmd,
		"peers":      swarmPeersCmd,
	},
}

const (
	swarmVerboseOptionName   = "verbose"
	swarmStreamsOptionName   = "streams"
	swarmLatencyOptionName   = "latency"
	swarmDirectionOptionName = "direction"
)

var swarmPeersCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "List peers with open connections.",
		ShortDescription: `
'mefs swarm peers' lists the set of peers this node is connected to.
`,
	},
	Options: []cmds.Option{
		cmds.BoolOption(swarmVerboseOptionName, "v", "display all extra information"),
		cmds.BoolOption(swarmStreamsOptionName, "Also list information about open streams for each peer"),
		cmds.BoolOption(swarmLatencyOptionName, "Also list information about latency to each peer"),
		cmds.BoolOption(swarmDirectionOptionName, "Also list information about the direction of connection"),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env)
		if err != nil {
			return err
		}

		verbose, _ := req.Options[swarmVerboseOptionName].(bool)
		latency, _ := req.Options[swarmLatencyOptionName].(bool)
		streams, _ := req.Options[swarmStreamsOptionName].(bool)
		direction, _ := req.Options[swarmDirectionOptionName].(bool)

		conns, err := api.Swarm().Peers(req.Context)
		if err != nil {
			return err
		}

		var out connInfos
		for _, c := range conns {
			ci := connInfo{
				Addr: c.Address().String(),
				Peer: c.ID().Pretty(),
			}

			if verbose || direction {
				// set direction
				ci.Direction = c.Direction()
			}

			if verbose || latency {
				lat, err := c.Latency()
				if err != nil {
					return err
				}

				if lat == 0 {
					ci.Latency = "n/a"
				} else {
					ci.Latency = lat.String()
				}
			}
			if verbose || streams {
				strs, err := c.Streams()
				if err != nil {
					return err
				}

				for _, s := range strs {
					ci.Streams = append(ci.Streams, streamInfo{Protocol: string(s)})
				}
			}
			sort.Sort(&ci)
			out.Peers = append(out.Peers, ci)
		}

		sort.Sort(&out)
		return cmds.EmitOnce(res, &out)
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, ci *connInfos) error {
			pipfs := ma.ProtocolWithCode(ma.P_IPFS).Name
			for _, info := range ci.Peers {
				fmt.Fprintf(w, "%s/%s/%s", info.Addr, pipfs, info.Peer)
				if info.Latency != "" {
					fmt.Fprintf(w, " %s", info.Latency)
				}

				if info.Direction != inet.DirUnknown {
					fmt.Fprintf(w, " %s", directionString(info.Direction))
				}
				fmt.Fprintln(w)

				for _, s := range info.Streams {
					if s.Protocol == "" {
						s.Protocol = "<no protocol name>"
					}

					fmt.Fprintf(w, "  %s\n", s.Protocol)
				}
			}

			return nil
		}),
	},
	Type: connInfos{},
}

type streamInfo struct {
	Protocol string
}

type connInfo struct {
	Addr      string
	Peer      string
	Latency   string
	Muxer     string
	Direction inet.Direction
	Streams   []streamInfo
}

func (ci *connInfo) Less(i, j int) bool {
	return ci.Streams[i].Protocol < ci.Streams[j].Protocol
}

func (ci *connInfo) Len() int {
	return len(ci.Streams)
}

func (ci *connInfo) Swap(i, j int) {
	ci.Streams[i], ci.Streams[j] = ci.Streams[j], ci.Streams[i]
}

type connInfos struct {
	Peers []connInfo
}

func (ci connInfos) Less(i, j int) bool {
	return ci.Peers[i].Addr < ci.Peers[j].Addr
}

func (ci connInfos) Len() int {
	return len(ci.Peers)
}

func (ci connInfos) Swap(i, j int) {
	ci.Peers[i], ci.Peers[j] = ci.Peers[j], ci.Peers[i]
}

// directionString transfers to string
func directionString(d inet.Direction) string {
	switch d {
	case inet.DirInbound:
		return "inbound"
	case inet.DirOutbound:
		return "outbound"
	default:
		return ""
	}
}

var swarmAddrsCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "List known addresses. Useful for debugging.",
		ShortDescription: `
'mefs swarm addrs' lists all addresses this node is aware of.
`,
	},
	Subcommands: map[string]*cmds.Command{
		"local":  swarmAddrsLocalCmd,
		"listen": swarmAddrsListenCmd,
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env)
		if err != nil {
			return err
		}

		addrs, err := api.Swarm().KnownAddrs(req.Context)
		if err != nil {
			return err
		}

		out := make(map[string][]string)
		for p, paddrs := range addrs {
			s := p.Pretty()
			for _, a := range paddrs {
				out[s] = append(out[s], a.String())
			}
		}

		return cmds.EmitOnce(res, &addrMap{Addrs: out})
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, am *addrMap) error {
			// sort the ids first
			ids := make([]string, 0, len(am.Addrs))
			for p := range am.Addrs {
				ids = append(ids, p)
			}
			sort.Sort(sort.StringSlice(ids))

			for _, p := range ids {
				paddrs := am.Addrs[p]
				fmt.Fprintf(w, "%s (%d)\n", p, len(paddrs))
				for _, addr := range paddrs {
					fmt.Fprintf(w, "\t"+addr+"\n")
				}
			}

			return nil
		}),
	},
	Type: addrMap{},
}

var swarmAddrsLocalCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "List local addresses.",
		ShortDescription: `
'mefs swarm addrs local' lists all local listening addresses announced to the network.
`,
	},
	Options: []cmds.Option{
		cmds.BoolOption("id", "Show peer ID in addresses."),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env)
		if err != nil {
			return err
		}

		showid, _ := req.Options["id"].(bool)
		//self, err := api.Key().Self(req.Context)
		node, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		maddrs, err := api.Swarm().LocalAddrs(req.Context)
		if err != nil {
			return err
		}

		var addrs []string
		for _, addr := range maddrs {
			saddr := addr.String()
			if showid {
				saddr = path.Join(saddr, "p2p", node.Identity.Pretty())
			}
			addrs = append(addrs, saddr)
		}
		sort.Sort(sort.StringSlice(addrs))
		return cmds.EmitOnce(res, &stringList{addrs})
	},
	Type: stringList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(stringListEncoder),
	},
}

var swarmAddrsListenCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "List interface listening addresses.",
		ShortDescription: `
'mefs swarm addrs listen' lists all interface addresses the node is listening on.
`,
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env)
		if err != nil {
			return err
		}

		var addrs []string
		maddrs, err := api.Swarm().ListenAddrs(req.Context)
		if err != nil {
			return err
		}

		for _, addr := range maddrs {
			addrs = append(addrs, addr.String())
		}
		sort.Sort(sort.StringSlice(addrs))

		return cmds.EmitOnce(res, &stringList{addrs})
	},
	Type: stringList{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(stringListEncoder),
	},
}

var swarmConnectCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Open connection to a given address.",
		ShortDescription: `
'mefs swarm connect' opens a new direct connection to a peer address.

The address format is an MEFS multiaddr:

mefs swarm connect "/ip4/103.39.231.220/tcp/14069/p2p/8MHSTe6UWXsBYLgFKdEU3VnrnVPRUp"
`,
	},
	Arguments: []cmds.Argument{
		cmds.StringArg("address", true, true, "Address of peer to connect to.").EnableStdin(),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env)
		if err != nil {
			return err
		}

		addrs := req.Arguments

		pis, err := peersWithAddresses(addrs)
		if err != nil {
			return err
		}

		output := make([]string, len(pis))
		for i, pi := range pis {
			output[i] = "connect " + pi.ID.Pretty()

			err := api.Swarm().Connect(req.Context, pi)
			if err != nil {
				return fmt.Errorf("%s failure: %s", output[i], err)
			}
			output[i] += " success"
		}

		return cmds.EmitOnce(res, &stringList{output})
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(stringListEncoder),
	},
	Type: stringList{},
}

var swarmDisconnectCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Close connection to a given address.",
		ShortDescription: `
'mefs swarm disconnect' closes a connection to a peer address. The address
format is an MEFS multiaddr:

mefs swarm disconnect "/ip4/103.39.231.220/tcp/14069/p2p/8MHSTe6UWXsBYLgFKdEU3VnrnVPRUp"

The disconnect is not permanent; if mefs needs to talk to that address later,
it will reconnect.
`,
	},
	Arguments: []cmds.Argument{
		cmds.StringArg("address", true, true, "Address of peer to disconnect from.").EnableStdin(),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env)
		if err != nil {
			return err
		}

		iaddrs, err := parseAddresses(req.Arguments)
		if err != nil {
			return err
		}

		output := make([]string, len(iaddrs))
		for i, addr := range iaddrs {
			output[i] = "disconnect " + addr.ID().Pretty()

			if err := api.Swarm().Disconnect(req.Context, addr.Multiaddr()); err != nil {
				output[i] += " failure: " + err.Error()
			} else {
				output[i] += " success"
			}
		}
		return cmds.EmitOnce(res, &stringList{output})
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(stringListEncoder),
	},
	Type: stringList{},
}

// parseAddresses is a function that takes in a slice of string peer addresses
// (multiaddr + peerid) and returns slices of multiaddrs and peerids.
func parseAddresses(addrs []string) (iaddrs []iaddr.IPFSAddr, err error) {
	iaddrs = make([]iaddr.IPFSAddr, len(addrs))
	for i, saddr := range addrs {
		iaddrs[i], err = iaddr.ParseString(saddr)
		if err != nil {
			return nil, cmds.ClientError("invalid peer address: " + err.Error())
		}
	}
	return
}

// peersWithAddresses is a function that takes in a slice of string peer addresses
// (multiaddr + peerid) and returns a slice of properly constructed peers
func peersWithAddresses(addrs []string) ([]peer.AddrInfo, error) {
	iaddrs, err := parseAddresses(addrs)
	if err != nil {
		return nil, err
	}

	peers := make(map[peer.ID][]ma.Multiaddr, len(iaddrs))
	for _, iaddr := range iaddrs {
		id := iaddr.ID()
		current, ok := peers[id]
		if tpt := iaddr.Transport(); tpt != nil {
			peers[id] = append(current, tpt)
		} else if !ok {
			peers[id] = nil
		}
	}
	pis := make([]peer.AddrInfo, 0, len(peers))
	for id, maddrs := range peers {
		pis = append(pis, peer.AddrInfo{
			ID:    id,
			Addrs: maddrs,
		})
	}
	return pis, nil
}

var swarmFiltersCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Manipulate address filters.",
		ShortDescription: `
'mefs swarm filters' will list out currently applied filters. Its subcommands
can be used to add or remove said filters. Filters are specified using the
multiaddr-filter format:

Example:

    /ip4/192.168.0.0/ipcidr/16

Where the above is equivalent to the standard CIDR:

    192.168.0.0/16

Filters default to those specified under the "Swarm.AddrFilters" config key.
`,
	},
	Subcommands: map[string]*cmds.Command{
		"add": swarmFiltersAddCmd,
		"rm":  swarmFiltersRmCmd,
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		n, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if n.PeerHost == nil {
			return ErrNotOnline
		}

		// FIXME(steb)
		swrm, ok := n.PeerHost.Network().(*swarm.Swarm)
		if !ok {
			return errors.New("failed to cast network to swarm network")
		}

		var output []string
		for _, f := range swrm.Filters.Filters() {
			s, err := mafilter.ConvertIPNet(f)
			if err != nil {
				return err
			}
			output = append(output, s)
		}
		return cmds.EmitOnce(res, &stringList{output})
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(stringListEncoder),
	},
	Type: stringList{},
}

var swarmFiltersAddCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Add an address filter.",
		ShortDescription: `
'mefs swarm filters add' will add an address filter to the daemons swarm.
Filters applied this way will not persist daemon reboots, to achieve that,
add your filters to the mefs config file.
`,
	},
	Arguments: []cmds.Argument{
		cmds.StringArg("address", true, true, "Multiaddr to filter.").EnableStdin(),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		n, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if n.PeerHost == nil {
			return ErrNotOnline
		}

		// FIXME(steb)
		swrm, ok := n.PeerHost.Network().(*swarm.Swarm)
		if !ok {
			return errors.New("failed to cast network to swarm network")
		}

		if len(req.Arguments) == 0 {
			return errors.New("no filters to add")
		}

		r, err := fsrepo.Open(env.(*commands.Context).ConfigRoot)
		if err != nil {
			return err
		}
		defer r.Close()
		cfg, err := r.Config()
		if err != nil {
			return err
		}

		for _, arg := range req.Arguments {
			mask, err := mafilter.NewMask(arg)
			if err != nil {
				return err
			}

			swrm.Filters.AddDialFilter(mask)
		}

		added, err := filtersAdd(r, cfg, req.Arguments)
		if err != nil {
			return err
		}

		return cmds.EmitOnce(res, &stringList{added})
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(stringListEncoder),
	},
	Type: stringList{},
}

var swarmFiltersRmCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Remove an address filter.",
		ShortDescription: `
'mefs swarm filters rm' will remove an address filter from the daemons swarm.
Filters removed this way will not persist daemon reboots, to achieve that,
remove your filters from the mefs config file.
`,
	},
	Arguments: []cmds.Argument{
		cmds.StringArg("address", true, true, "Multiaddr filter to remove.").EnableStdin(),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		n, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		if n.PeerHost == nil {
			return ErrNotOnline
		}

		swrm, ok := n.PeerHost.Network().(*swarm.Swarm)
		if !ok {
			return errors.New("failed to cast network to swarm network")
		}

		r, err := fsrepo.Open(env.(*commands.Context).ConfigRoot)
		if err != nil {
			return err
		}
		defer r.Close()
		cfg, err := r.Config()
		if err != nil {
			return err
		}

		if req.Arguments[0] == "all" || req.Arguments[0] == "*" {
			fs := swrm.Filters.Filters()
			for _, f := range fs {
				swrm.Filters.Remove(f)
			}

			removed, err := filtersRemoveAll(r, cfg)
			if err != nil {
				return err
			}

			return cmds.EmitOnce(res, &stringList{removed})
		}

		for _, arg := range req.Arguments {
			mask, err := mafilter.NewMask(arg)
			if err != nil {
				return err
			}

			swrm.Filters.Remove(mask)
		}

		removed, err := filtersRemove(r, cfg, req.Arguments)
		if err != nil {
			return err
		}

		return cmds.EmitOnce(res, &stringList{removed})
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(stringListEncoder),
	},
	Type: stringList{},
}

func filtersAdd(r repo.Repo, cfg *config.Config, filters []string) ([]string, error) {
	addedMap := map[string]struct{}{}
	addedList := make([]string, 0, len(filters))

	// re-add cfg swarm filters to rm dupes
	oldFilters := cfg.Swarm.AddrFilters
	cfg.Swarm.AddrFilters = nil

	// add new filters
	for _, filter := range filters {
		if _, found := addedMap[filter]; found {
			continue
		}

		cfg.Swarm.AddrFilters = append(cfg.Swarm.AddrFilters, filter)
		addedList = append(addedList, filter)
		addedMap[filter] = struct{}{}
	}

	// add back original filters. in this order so that we output them.
	for _, filter := range oldFilters {
		if _, found := addedMap[filter]; found {
			continue
		}

		cfg.Swarm.AddrFilters = append(cfg.Swarm.AddrFilters, filter)
		addedMap[filter] = struct{}{}
	}

	if err := r.SetConfig(cfg); err != nil {
		return nil, err
	}

	return addedList, nil
}

func filtersRemoveAll(r repo.Repo, cfg *config.Config) ([]string, error) {
	removed := cfg.Swarm.AddrFilters
	cfg.Swarm.AddrFilters = nil

	if err := r.SetConfig(cfg); err != nil {
		return nil, err
	}

	return removed, nil
}

func filtersRemove(r repo.Repo, cfg *config.Config, toRemoveFilters []string) ([]string, error) {
	removed := make([]string, 0, len(toRemoveFilters))
	keep := make([]string, 0, len(cfg.Swarm.AddrFilters))

	oldFilters := cfg.Swarm.AddrFilters

	for _, oldFilter := range oldFilters {
		found := false
		for _, toRemoveFilter := range toRemoveFilters {
			if oldFilter == toRemoveFilter {
				found = true
				removed = append(removed, toRemoveFilter)
				break
			}
		}

		if !found {
			keep = append(keep, oldFilter)
		}
	}
	cfg.Swarm.AddrFilters = keep

	if err := r.SetConfig(cfg); err != nil {
		return nil, err
	}

	return removed, nil
}

func stringListEncoder(req *cmds.Request, w io.Writer, list *stringList) error {
	for _, str := range list.Strings {
		_, err := fmt.Fprintf(w, "%s\n", str)
		if err != nil {
			return err
		}
	}
	return nil
}
