package pos

import (
	"encoding/hex"
	"log"
	"math/big"

	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

func GetPosAddr() string {
	return "0x46EF41A6c60912173b50dAD714Ca2a6f82c32aE8"
}

func GetPosId() string {
	id, _ := address.GetIDFromAddress(GetPosAddr())
	return id
}

func GetPosPrice() *big.Int {
	return big.NewInt(utils.STOREPRICE / 10)
}

func GetPosSeed() []byte {
	seed, err := hex.DecodeString("c7463268d6b2ad969b3457d9e89c5f34e9debe445417fbf50299fe471a28f2c7")
	if err != nil {
		return nil
	}
	log.Println("pos seed is: ", hex.EncodeToString(seed[:]))
	return seed
}
