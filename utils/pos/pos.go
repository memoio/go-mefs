package pos

import (
	"encoding/hex"
	"log"
	"math/big"

	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

const (
	DLen     = 100 * 1024 * 1024
	SegCount = 3200
	SegSize  = 32768
	Reps     = 10 // 10备份
)

//sk:3aef4a160e08ffcd27f4ca6d49ec0a0dfe39f8ae2f3f880b32a405ade9eb0eea
//ukAddr: 0x180956aC2979c424e481689B484186B3fe184b40
func GetPostAddr() string {
	return "0x2F34Aae01b7A66502d114EbcC50b16C78E645C32"
}

func GetPostId() string {
	id, _ := address.GetIDFromAddress(GetPostAddr())
	return id
}

func GetPostPrice() *big.Int {
	return big.NewInt(utils.STOREPRICE / 10)
}

func GetPostSeed() []byte {
	seed, err := hex.DecodeString("c7463268d6b2ad969b3457d9e89c5f34e9debe445417fbf50299fe471a28f2c7")
	if err != nil {
		return nil
	}
	log.Println("post seed is: ", hex.EncodeToString(seed[:]))
	return seed
}
