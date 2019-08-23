package address

import (
	"testing"
)

func TestGetIPFSAddressFromID(t *testing.T) {
	id := "8MJVWqs8fsJ823dRsbJdVak9HnPMEp"
	_, err := GetAddressFromID(id)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetIDFromIPFSAddress(t *testing.T) {
	address := "0xAF4F9731aBfd349BAbb48AA569EB22bE2e4C55B1"
	_, err := GetIDFromAddress(address)
	if err != nil {
		t.Fatal(err)
	}
}
