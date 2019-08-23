// Package dshelp provides utilities for parsing and creating
// datastore keys used by go-ipfs
package key

import (
	cid "github.com/memoio/go-mefs/source/go-cid"
	datastore "github.com/memoio/go-mefs/source/go-datastore"
	b58 "github.com/mr-tron/base58/base58"
)

// NewKeyFromBinary creates a new key from a byte slice.
func NewKeyFromBinary(rawKey []byte) datastore.Key {
	kbtc58 := string('/') + b58.Encode(rawKey)
	return datastore.RawKey(kbtc58)
}

// BinaryFromDsKey returns the byte slice corresponding to the given Key.
func BinaryFromDsKey(k datastore.Key) []byte {
	//return base32.RawStdEncoding.DecodeString(k.String()[1:])
	return []byte(k.String()[1:])
}

// CidToDsKey creates a Key from the given Cid.
func CidToDsKey(k cid.Cid) datastore.Key {
	if k.Version() == 2 {
		return datastore.RawKey(string('/') + k.String())
	}
	return NewKeyFromBinary(k.Bytes())
}

// DsKeyToCid converts the given Key to its corresponding Cid.
func DsKeyToCid(dsKey datastore.Key) (cid.Cid, error) {
	kb := BinaryFromDsKey(dsKey)
	return cid.Cast(kb)
}
