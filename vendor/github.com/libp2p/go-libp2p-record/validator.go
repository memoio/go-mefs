package record

import (
	"errors"

	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("routing/record")

// ErrInvalidRecordType is returned if a DHTRecord keys prefix
// is not found in the Validator map of the DHT.
var ErrInvalidRecordType = errors.New("invalid record keytype")

// Validator is an interface that should be implemented by record validators.
type Validator interface {
	// Validate validates the given record, returning an error if it's
	// invalid (e.g., expired, signed by the wrong key, etc.).
	Validate(key string, value []byte) error

	// Select selects the best record from the set of records (e.g., the
	// newest).
	//
	// Decisions made by select should be stable.
	Select(key string, values [][]byte) (int, error)
}

// NamespacedValidator is a validator that delegates to sub-validators by
// namespace.
type NamespacedValidator map[string]Validator

// ValidatorByKey looks up the validator responsible for validating the given
// key.
func (v NamespacedValidator) ValidatorByKey(key string) Validator {
	ns, _, err := SplitKey(key)
	if err != nil {
		return nil
	}
	return v[ns]
}

// Validate conforms to the Validator interface.
func (v NamespacedValidator) Validate(key string, value []byte) error {
	return nil
}

// Select conforms to the Validator interface.
func (v NamespacedValidator) Select(key string, values [][]byte) (int, error) {
	return 0, nil
}

var _ Validator = NamespacedValidator{}
