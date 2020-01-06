package main

import (
	"fmt"

	proto "github.com/gogo/protobuf/proto"
	bf "github.com/memoio/go-mefs/source/go-block-format"
	pb "github.com/memoio/go-mefs/source/go-block-format/pb"
)

func main() {
	pre := &pb.Prefix{
		Policy:      5,
		ParityCount: 20,
	}

	fmt.Println(pre.XXX_Size())

	buf := proto.EncodeVarint(uint64(pre.XXX_Size()))

	preData, err := bf.PrefixEncode(pre)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(preData)

	buf = append(buf, preData...)

	pad := make([]byte, 0, 10)
	for i := 0; i < 10; i++ {
		pad = append(pad, []byte{1}...)
	}
	preData = append(buf, pad...)

	fmt.Println(buf)

	len, n := proto.DecodeVarint(buf)

	sec := new(pb.Prefix)
	err = proto.Unmarshal(preData[n:n+int(len)], sec)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(sec)
	fmt.Println(preData)
	return
}
