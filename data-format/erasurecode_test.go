package dataformat

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	"github.com/memoio/go-mefs/data-format/reedsolomon"
)

func TestEncodeData(t *testing.T) {
	DataCount, ParityCount, BlockSize := 20, 5, 1024*1024
	r, err := reedsolomon.New(DataCount, ParityCount)
	if err != nil {
		t.Fatal(err)
	}
	tmpData := make([][]byte, DataCount)
	for i := 0; i < DataCount; i++ {
		tmpData[i] = make([]byte, BlockSize)
	}
	rand.Seed(0)
	for i := 0; i < DataCount; i++ {
		fillRandom(tmpData[i])
	}
	newData, err := EncodeData(tmpData, DataCount, ParityCount)
	if err != nil {
		t.Fatal(err)
	}
	ok, err := r.Verify(newData)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Verification failed")
	}

	buf1, buf2 := new(bytes.Buffer), new(bytes.Buffer)
	err = r.Join(buf1, newData, 20)
	if err != nil {
		t.Fatal(err)
	}
	err = r.Join(buf2, tmpData, 20)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf1.Bytes(), buf2.Bytes()) {
		t.Fatal("recovered data does match original")
	}
}

func TestEncodeSegment(t *testing.T) {
	DataCount, ParityCount, SegmentSize := 3, 4, 20

	// 小数据装入且不含有prefix
	tmpData := make([]byte, 150)
	rand.Seed(0)
	fillRandom(tmpData)
	blocks, offset, err := EncodeDataToNoPreStripe(tmpData, "17nCkDaiM7C9gLZWUBg9c4mCswP_1_0", DataCount, ParityCount, CRC32, 0, uint64(SegmentSize), nil)
	// _, blocks, offset, datalength, err := EncodeDataToStripe(tmpData,[]byte{66,21,32,124,6}, nil, DataCount, ParityCount, 0, CRC32, uint64(SegmentSize))
	if err != nil {
		t.Fatal(err)
	}
	// if !bytes.Equal(blocks[0][3:9], tmpData[:6]) {
	// 	t.Fatal("recovered data does match original")
	// }
	// if !bytes.Equal(blocks[1][3:9], tmpData[20:26]) {
	// 	t.Fatal("recovered data does match original")
	// }
	// if !bytes.Equal(blocks[2][3:9], tmpData[40:46]) {
	// 	t.Fatal("recovered data does match original")
	// }
	// if !bytes.Equal(blocks[0][31:50], tmpData[60:79]) {
	// 	t.Fatal("recovered data does match original")
	// }
	// if !bytes.Equal(blocks[1][31:50], tmpData[80:99]) {
	// 	t.Fatal("recovered data does match original")
	// }
	// if !bytes.Equal(blocks[2][31:50], tmpData[100:119]) {
	// 	t.Fatal("recovered data does match original")
	// }

	if !bytes.Equal(blocks[0][:2], tmpData[:2]) {
		t.Fatal("recovered data does match original")
	}
	fmt.Println("offset:", offset)
	fmt.Println("len(blocks)", len(blocks))

	// 一个Stripe不能装入的数据
	// tmpData2 := make([]byte, 300*60)
	// rand.Seed(0)
	// fillRandom(tmpData2)
	// blocks2, offset, err := EncodeDataToStripe(tmpData2, "17nCkDaiM7C9gLZWUBg9c4mCswP_1_0", DataCount, ParityCount, BLS, true, uint64(SegmentSize))
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// fmt.Println("len(blocks2)", len(blocks2))
	// fmt.Println("offset:", offset)

	// 小数据装入含有prefix
	tmpData3 := make([]byte, 100)
	rand.Seed(0)
	fillRandom(tmpData3)
	_, offset, err = EncodeDataToPreStripe(tmpData3, "17nCkDaiM7C9gLZWUBg9c4mCswP_1_0", DataCount, ParityCount, CRC32, uint64(SegmentSize), nil)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("offset:", offset)
}

func TestRecoverData(t *testing.T) {
	DataCount, ParityCount, BlockSize := 20, 5, 20
	r, err := reedsolomon.New(DataCount, ParityCount)
	if err != nil {
		t.Fatal(err)
	}
	tmpData := make([][]byte, DataCount)
	for i := 0; i < DataCount; i++ {
		tmpData[i] = make([]byte, BlockSize)
	}
	rand.Seed(0)
	for i := 0; i < DataCount; i++ {
		fillRandom(tmpData[i])
	}
	newData, err := EncodeData(tmpData, DataCount, ParityCount)
	if err != nil {
		t.Fatal(err)
	}
	ok, err := r.Verify(newData)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Verification failed")
	}

	// RecoverData with all shards present
	recoverData, err := RecoverData(newData, DataCount, ParityCount, -1)
	if err != nil {
		t.Fatal(err)
	}
	ok, err = r.Verify(recoverData)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Verification failed")
	}

	//RecoverData with all DataBLock without ParityBlock
	for i := DataCount; i < DataCount+ParityCount; i++ {
		newData[i] = nil
	}
	recoverData, err = RecoverData(newData, DataCount, ParityCount, -1)
	if err != nil {
		t.Fatal(err)
	}
	ok, err = r.Verify(recoverData)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Verification failed")
	}

	//RecoverData with some DataBLocks and ParityBlock
	newData[0] = nil
	newData[4] = nil
	newData[8] = nil
	newData[19] = nil
	newData[22] = nil
	recoverData, err = RecoverData(newData, DataCount, ParityCount, 0, 4, 8, 19, 22)
	if err != nil {
		t.Fatal(err)
	}
	newData, err = RecoverData(newData, DataCount, ParityCount, -1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(newData[1])
	fmt.Println(recoverData[0])
	if !bytes.Equal(newData[0], recoverData[0]) {
		t.Fatal("recovered data does match original")
	}
	if !bytes.Equal(newData[4], recoverData[1]) {
		t.Fatal("recovered data does match original")
	}
	if !bytes.Equal(newData[8], recoverData[2]) {
		t.Fatal("recovered data does match original")
	}
	if !bytes.Equal(newData[19], recoverData[3]) {
		t.Fatal("recovered data does match original")
	}
	if !bytes.Equal(newData[22], recoverData[4]) {
		t.Fatal("recovered data does match original")
	}

	// 改变块内数据
	newData[0][12] = 27
	newData[1] = nil
	_, err = RecoverData(newData, DataCount, ParityCount, -1)
	if err != nil {
		t.Fatal(err)
	}

	// append块内数据，长度变化
	fmt.Println(newData[0][12])
	newData[0] = append(newData[0], 12)
	newData[1] = nil
	_, err = RecoverData(newData, DataCount, ParityCount, -1)
	if err != reedsolomon.ErrShardSize {
		t.Fatal(err)
	}
}

func TestGetFileFromStripe(t *testing.T) {
	DataCount, ParityCount, SegmentSize := 3, 4, 20
	tmpData := make([]byte, 150)
	rand.Seed(0)
	fillRandom(tmpData)

	//  CRC
	// _, blocks, offset, _, err := EncodeDataToStripe(tmpData, nil, nil, DataCount, ParityCount, 0, CRC32, uint64(SegmentSize))
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// newData, err := GetFileDataFromSripe(blocks, DataCount, 0, offset-1)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// if !bytes.Equal(tmpData[0:20], newData[0:20]) {
	// 	t.Fatal("recovered data does match original")
	// }
	// if !bytes.Equal(tmpData[0:40], newData[0:40]) {
	// 	t.Fatal("recovered data does match original")
	// }
	// if !bytes.Equal(tmpData[0:100], newData[0:100]) {
	// 	t.Fatal("recovered data does match original")
	// }

	// newData, err = GetFileDataFromSripe(blocks, DataCount, 0, -1)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// if !bytes.Equal(tmpData[0:20], newData[0:20]) {
	// 	t.Fatal("recovered data does match original")
	// }
	// if !bytes.Equal(tmpData[0:40], newData[0:40]) {
	// 	t.Fatal("recovered data does match original")
	// }
	// if !bytes.Equal(tmpData[0:100], newData[0:100]) {
	// 	t.Fatal("recovered data does match original")
	// }

	// newData, err = GetFileDataFromSripe(blocks, DataCount, 1, -1)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// fmt.Println(tmpData[60:79])
	// fmt.Println(newData[0:20])
	// if !bytes.Equal(tmpData[60:80], newData[0:20]) {
	// 	t.Fatal("recovered data does match original")
	// }
	// if !bytes.Equal(tmpData[60:100], newData[0:40]) {
	// 	t.Fatal("recovered data does match original")
	// }

	// BLS
	blocks, offset, err := EncodeDataToPreStripe(tmpData, "17nCkDaiM7C9gLZWUBg9c4mCswP_1_0", DataCount, ParityCount, CRC32, uint64(SegmentSize), nil)
	if err != nil {
		t.Fatal(err)
	}
	newData, err := GetFileDataFromSripe(blocks, DataCount, 0, offset-1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(tmpData[0:20], newData[0:20]) {
		t.Fatal("recovered data does match original")
	}
	if !bytes.Equal(tmpData[0:40], newData[0:40]) {
		t.Fatal("recovered data does match original")
	}
	if !bytes.Equal(tmpData[0:100], newData[0:100]) {
		t.Fatal("recovered data does match original")
	}

	newData, err = GetFileDataFromSripe(blocks, DataCount, 0, -1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(tmpData[0:20], newData[0:20]) {
		t.Fatal("recovered data does match original")
	}
	if !bytes.Equal(tmpData[0:40], newData[0:40]) {
		t.Fatal("recovered data does match original")
	}
	if !bytes.Equal(tmpData[0:100], newData[0:100]) {
		t.Fatal("recovered data does match original")
	}

	newData, err = GetFileDataFromSripe(blocks, DataCount, 1, -1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(tmpData[60:79])
	fmt.Println(newData[0:20])
	if !bytes.Equal(tmpData[60:80], newData[0:20]) {
		t.Fatal("recovered data does match original")
	}
	if !bytes.Equal(tmpData[60:100], newData[0:40]) {
		t.Fatal("recovered data does match original")
	}
}

func TestRecoverBlocks(t *testing.T) {
	DataCount, ParityCount, SegmentSize := 3, 4, 20
	tmpData := make([]byte, 150)
	rand.Seed(0)
	fillRandom(tmpData)
	blocks, _, err := EncodeDataToPreStripe(tmpData, "17nCkDaiM7C9gLZWUBg9c4mCswP_1_0", DataCount, ParityCount, CRC32, uint64(SegmentSize), nil)
	if err != nil {
		t.Fatal(err)
	}
	newBlocks, _ := RecoverStripe(blocks)
	newBlocks = blocks

	blocks, _, err = EncodeDataToPreStripe(tmpData, "17nCkDaiM7C9gLZWUBg9c4mCswP_1_0", DataCount, ParityCount, CRC32, uint64(SegmentSize), nil)
	if err != nil {
		t.Fatal(err)
	}
	blocks[1] = nil
	blocks[3] = nil
	blocks, _ = RecoverStripe(blocks)
	fmt.Println(len(blocks))
	if !bytes.Equal(newBlocks[0], blocks[0]) {
		t.Fatal("recovered data does match original")
	}
	fmt.Println(newBlocks[1][:50])
	fmt.Println(blocks[1][:50])
	if !bytes.Equal(newBlocks[1], blocks[1]) {
		t.Fatal("recovered data does match original")
	}
	if !bytes.Equal(newBlocks[2], blocks[2]) {
		t.Fatal("recovered data does match original")
	}
	if !bytes.Equal(newBlocks[3], blocks[3]) {
		t.Fatal("recovered data does match original")
	}
	if !bytes.Equal(newBlocks[6], blocks[6]) {
		t.Fatal("recovered data does match original")
	}
}

func fillRandom(p []byte) {
	for i := 0; i < len(p); i += 7 {
		val := rand.Int63()
		for j := 0; i+j < len(p) && j < 7; j++ {
			p[i+j] = byte(val)
			val >>= 8
		}
	}
}
