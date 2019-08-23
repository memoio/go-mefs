package dataformat

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
)

func TestDataEncodeToMul(t *testing.T) {
	DataCount, ParityCount, SegmentSize := 3, 4, 20
	tmpData := make([]byte, 150)
	rand.Seed(0)
	fillRandom(tmpData)
	stripe, offset, err := DataEncodeToMul(tmpData, "17nCkDaiM7C9gLZWUBg9c4mCswP_1_0", int32(DataCount), int32(ParityCount), uint64(SegmentSize), CRC32, nil)
	if err != nil {
		t.Error("error")
	}
	fmt.Println(len(stripe))
	fmt.Println(offset)

	newData, err := GetDataFromRawData(stripe[0])
	if err != nil {
		t.Error("error")
	}
	if !bytes.Equal(newData[:150], tmpData) {
		t.Error("error")
	}

	newStripe, err := RecoverMul(stripe[2], 2)
	if err != nil {
		t.Error("error")
	}
	fmt.Println(len(newStripe))
	if !bytes.Equal(newStripe[0], stripe[0]) {
		t.Error("error")
	}
	if !bytes.Equal(newStripe[4], stripe[4]) {
		t.Error("error")
	}
}

func TestDataEncodeToMulForAppend(t *testing.T) {
	DataCount, ParityCount, SegmentSize := 3, 2, 20
	tmpData := make([]byte, 150)
	rand.Seed(0)
	fillRandom(tmpData)
	stripe, offset, err := DataEncodeToMulForAppend(tmpData, "17nCkDaiM7C9gLZWUBg9c4mCswP_1_0", int32(DataCount), int32(ParityCount), uint64(SegmentSize), CRC32, 8, nil)
	if err != nil {
		t.Error("error")
	}
	fmt.Println(len(stripe))
	fmt.Println(offset)
}

func TestPrefixEncode(t *testing.T) {
	pre, _ := PrefixEncode(1, 3, 2, CRC32, DefaultSegmentSize, DefaultLengths[CRC32])
	fmt.Println(pre)
	prefix, _, err := PrefixDecode(pre)
	fmt.Println(prefix, err)
	if bytes.Equal(pre, []byte{1, 3, 2, 0, 128, 32, 4}) {

	} else {
		t.Error("error")
	}
}
