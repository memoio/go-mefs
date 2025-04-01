package dht

import (
	"bytes"
	"testing"

	proto "github.com/gogo/protobuf/proto"
	recpb "github.com/memoio/go-mefs/source/go-libp2p-kad-dht/pb"
)

func TestCleanRecordSigned(t *testing.T) {
	actual := new(recpb.Record)
	actual.Value = []byte("value")
	actual.Key = []byte("key")

	actualBytes, err := proto.Marshal(actual)
	if err != nil {
		t.Fatal(err)
	}

	expected := new(recpb.Record)
	expected.Value = []byte("value")
	expected.Key = []byte("key")
	expectedBytes, err := proto.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(actualBytes, expectedBytes) {
		t.Error("failed to clean record")
	}
}

func TestCleanRecord(t *testing.T) {
	actual := new(recpb.Record)
	actual.Key = []byte("key")
	actual.Value = []byte("value")

	actualBytes, err := proto.Marshal(actual)
	if err != nil {
		t.Fatal(err)
	}

	expected := new(recpb.Record)
	expected.Key = []byte("key")
	expected.Value = []byte("value")
	expectedBytes, err := proto.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(actualBytes, expectedBytes) {
		t.Error("failed to clean record")
	}
}
