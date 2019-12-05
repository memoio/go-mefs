# METB

`metb` is a program used to create and manage a cluster of sandboxed MEFS nodes locally on your computer. Spin up 1000s of nodes! It exposes various options, such as different bootstrapping patterns. `metb` makes testing MEFS networks easy!

### Example

```
$ metb init -n 5

$ metb start
Started daemon 0, pid = 12396
Started daemon 1, pid = 12406
Started daemon 2, pid = 12415
Started daemon 3, pid = 12424
Started daemon 4, pid = 12434

$ metb shell 0
$ echo $IPFS_PATH
/home/noffle/testbed/0

$ echo 'hey!' | mefs add -q
QmNqugRcYjwh9pEQUK7MLuxvLjxDNZL1DH8PJJgWtQXxuF

$ exit

$ metb connect 0 4

$ metb shell 4
$ mefs cat QmNqugRcYjwh9pEQUK7MLuxvLjxDNZL1DH8PJJgWtQXxuF
hey!
```

### Usage
```
$ metb --help

NAME:
	metb - The MEFS TestBed

USAGE:
	metb [global options] command [command options] [arguments...]

COMMANDS:
	init			create and initialize testbed configuration
	start			start up all testbed nodes
	kill, stop		kill a specific node (or all nodes, if none specified)
	restart			kill all nodes, then restart
	shell			spawn a subshell with certain MEFS environment variables set
	get			get an attribute of the given node
	connect			connect two nodes together
	dump-stack		get a stack dump from the given daemon
	help, h			show a list of subcommands, or help for a specific subcommand

GLOBAL OPTIONS:
	--help, -h		show help
	--version, -v		print the version
```

### Install

```
go get github.com/memoio/go-mefs/source/metb
```

### Configuration

By default, `metb` uses `$HOME/testbed` to store created nodes. This path is configurable via the environment variables `METB_ROOT`.



### License

MIT
