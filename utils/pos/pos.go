package pos

import (
	"encoding/hex"

	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

const (
	PosSkStr = "724c1e75ba94dd0305cf532cd6db95df0721c33dcdc323502eba067409e4842b"
)

func GetPosSkByte() []byte {
	SkByte, _ := hex.DecodeString(PosSkStr)
	return SkByte
}

func GetPosAddr() string {
	addr, _ := address.GetAdressFromSk(PosSkStr)
	return addr
}

func GetPosId() string {
	id, _ := address.GetIDFromAddress(GetPosAddr())
	return id
}

func GetPosPrice() int64 {
	return utils.STOREPRICEPEDOLLAR / 10
}
