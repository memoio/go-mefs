package loader

import (
	"github.com/memoio/go-mefs/plugin"
	"github.com/memoio/go-mefs/repo/fsrepo"

	"github.com/opentracing/opentracing-go"
)

func initialize(plugins []plugin.Plugin) error {
	for _, p := range plugins {
		err := p.Init()
		if err != nil {
			return err
		}
	}

	return nil
}

func run(plugins []plugin.Plugin) error {
	for _, pl := range plugins {
		switch pl := pl.(type) {
		case plugin.PluginTracer:
			err := runTracerPlugin(pl)
			if err != nil {
				return err
			}
		case plugin.PluginDatastore:
			err := fsrepo.AddDatastoreConfigHandler(pl.DatastoreTypeName(), pl.DatastoreConfigParser())
			if err != nil {
				return err
			}
		default:
			panic(pl)
		}
	}
	return nil
}

func runTracerPlugin(pl plugin.PluginTracer) error {
	tracer, err := pl.InitTracer()
	if err != nil {
		return err
	}
	opentracing.SetGlobalTracer(tracer)
	return nil
}
