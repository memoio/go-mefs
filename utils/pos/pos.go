package pos

import (
	"encoding/hex"

	"github.com/memoio/go-mefs/utils/address"
)

const (
	PosSkStr = "7b686dff4326fadf098ae156d57d6c917715a44fdb1655cc7472cd78915e2e03"
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
