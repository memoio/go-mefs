package main

import (
	plugin "github.com/memoio/go-mefs/source/metb-plugins/browser"
	testbedi "github.com/memoio/go-mefs/source/metb-plugins/metb-cli/testbed/interfaces"
)

var PluginName string
var NewNode testbedi.NewNodeFunc

func init() {
	PluginName = plugin.PluginName
	NewNode = plugin.NewNode
}
