// Package iface defines MEFS Core API which is a set of interfaces used to
// interact with MEFS nodes.
package iface

// CoreAPI defines an unified interface to MEFS for Go programs
type CoreAPI interface {
	// Block returns an implementation of Block API
	Block() BlockAPI

	// Swarm returns an implementation of Swarm API
	Swarm() SwarmAPI
}
