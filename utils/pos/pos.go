package pos

import (
	"encoding/hex"
	"fmt"

	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

func GetPosAddr() string {
	// sk:7a81e548e9f62f01e7eadb5af37d3ea15689602b470be9812752177464661c5c
	// groupID: 8MG5oUXz9vMgpuG8fvHGWu37rG1vqF
	// upKeeping: 8MG6NxZkxKE5gzGdm1bx5uWnTaPjLV
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
	seed, err := hex.DecodeString("adcd3318b3a31e74bf5f42fc837a1874155a396b9df04736ab19b23dcb7e2fd5")
	if err != nil {
		return nil
	}
	fmt.Println("pos seed is: ", hex.EncodeToString(seed[:]))
	return seed
}
