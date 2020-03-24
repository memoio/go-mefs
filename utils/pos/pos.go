package pos

import (
	"encoding/hex"
	"fmt"

	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

func GetPosAddr() string {
	// sk:7a81e548e9f62f01e7eadb5af37d3ea15689602b470be9812752177464661c5c
	// groupID: 8MKM7it7TwjdMN8S8tfEbFMvjqoYBL
	// upKeeping: 8MH8MpDqsHGST7CJ8EDA3JbmW2AD4p
	return "0x6C93F7D1437CF44048849657853f10F9802f3364"
}

func GetPosId() string {
	id, _ := address.GetIDFromAddress(GetPosAddr())
	return id
}

func GetPosPrice() int64 {
	return utils.STOREPRICEPEDOLLAR / 10
}

func GetPosSeed() []byte {
	seed, err := hex.DecodeString("8bee9763fcf2740b0cb6ad3351addfb9f8ee48efc1b7b4204fa6fd06b25cde89")
	if err != nil {
		return nil
	}
	fmt.Println("pos seed is: ", hex.EncodeToString(seed[:]))
	return seed
}
