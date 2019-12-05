package main

import (
	plugin "github.com/memoio/go-mefs/source/metb-plugins/local"
	testbedi "github.com/memoio/go-mefs/source/metb-plugins/metb-cli/testbed/interfaces"
)

var PluginName string
var NewNode testbedi.NewNodeFunc
var GetAttrList testbedi.GetAttrListFunc
var GetAttrDesc testbedi.GetAttrDescFunc

func init() {
	PluginName = plugin.PluginName
	NewNode = plugin.NewNode
	GetAttrList = plugin.GetAttrList
	GetAttrDesc = plugin.GetAttrDesc
}
