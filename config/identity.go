package config

const IdentityTag = "Identity"
const PrivKeyTag = "PrivKey"
const PrivKeySelector = IdentityTag + "." + PrivKeyTag

// Identity tracks the configuration of the local node's identity.
type Identity struct {
	PeerID  string
	PrivKey string `json:",omitempty"`
}
