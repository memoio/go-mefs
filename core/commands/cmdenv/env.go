package cmdenv

import (
	"fmt"

	"github.com/memoio/go-mefs/commands"
	"github.com/memoio/go-mefs/core"
	coreiface "github.com/memoio/go-mefs/core/coreapi/interface"
	cmds "github.com/ipfs/go-ipfs-cmds"
	config "github.com/memoio/go-mefs/config"
	files "github.com/ipfs/go-ipfs-files"
)

// GetNode extracts the node from the environment.
func GetNode(env interface{}) (*core.MefsNode, error) {
	ctx, ok := env.(*commands.Context)
	if !ok {
		return nil, fmt.Errorf("expected env to be of type %T, got %T", ctx, env)
	}

	return ctx.GetNode()
}

// GetApi extracts CoreAPI instance from the environment.
func GetApi(env cmds.Environment) (coreiface.CoreAPI, error) {
	ctx, ok := env.(*commands.Context)
	if !ok {
		return nil, fmt.Errorf("expected env to be of type %T, got %T", ctx, env)
	}

	return ctx.GetApi()
}

// GetConfig extracts the config from the environment.
func GetConfig(env cmds.Environment) (*config.Config, error) {
	ctx, ok := env.(*commands.Context)
	if !ok {
		return nil, fmt.Errorf("expected env to be of type %T, got %T", ctx, env)
	}

	return ctx.GetConfig()
}

// GetConfigRoot extracts the config root from the environment
func GetConfigRoot(env cmds.Environment) (string, error) {
	ctx, ok := env.(*commands.Context)
	if !ok {
		return "", fmt.Errorf("expected env to be of type %T, got %T", ctx, env)
	}

	return ctx.ConfigRoot, nil
}

// GetFileArg returns the next file from the directory or an error
func GetFileArg(it files.DirIterator) (files.File, error) {
	if !it.Next() {
		err := it.Err()
		if err == nil {
			err = fmt.Errorf("expected a file argument")
		}
		return nil, err
	}
	file := files.FileFromEntry(it)
	if file == nil {
		return nil, fmt.Errorf("file argument was nil")
	}
	return file, nil
}
