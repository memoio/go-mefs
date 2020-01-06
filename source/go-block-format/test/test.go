package main

import (
	"fmt"

	proto "github.com/gogo/protobuf/proto"
	bf "github.com/memoio/go-mefs/source/go-block-format"
	pb "github.com/memoio/go-mefs/source/go-block-format/pb"
)

func main() {
	pre := &pb.Prefix{
		Policy:      567677,
		ParityCount: 200000000,
	}

	fmt.Println(proto.Size(pre))
	preData, err := bf.PrefixEncode(pre)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(preData)

	pad := make([]byte, 0, 10)
	for i := 0; i < 10; i++ {
		pad = append(pad, []byte{1}...)
	}
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
