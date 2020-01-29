package main

import (
	"bytes"
	"fmt"
	"io"

	ggio "github.com/gogo/protobuf/io"
	proto "github.com/gogo/protobuf/proto"
	bf "github.com/memoio/go-mefs/source/go-block-format"
	pb "github.com/memoio/go-mefs/source/go-block-format/pb"
)

func main() {
	pre := &pb.Prefix{
		Policy:      567677,
		ParityCount: 200000000,
	}

	sbBuffer := bytes.NewBuffer(nil)
	sbWriter := ggio.NewDelimitedWriter(sbBuffer)

	err := sbWriter.WriteMsg(pre)
	if err != nil {
		return
	}

	pad := make([]byte, 0, 60)
	for i := 0; i < 60; i++ {
		pad = append(pad, []byte{1}...)
	}

	data := sbBuffer.Bytes()
	data = append(data, pad...)

	fmt.Println(sbBuffer.Len())
	sbBuffer.Reset()
	sbBuffer = bytes.NewBuffer(data)
	fmt.Println(sbBuffer.Len())

	newpre := new(pb.Prefix)
	sbReader := ggio.NewDelimitedReader(sbBuffer, 60)
	err = sbReader.ReadMsg(newpre)
	if err != nil && err != io.EOF {
		return
	}

	fmt.Println(newpre)

	fmt.Println(proto.Size(pre))
	preData, _, err := bf.PrefixEncode(pre)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(preData)

	preData = append(preData, pad...)

	fmt.Println(preData)

	sec, len, err := bf.PrefixDecode(preData)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(sec)
	fmt.Println(len)
	return
}
