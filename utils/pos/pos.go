package pos

import (
	"encoding/hex"

	"github.com/memoio/go-mefs/utils/address"
)

const (
	PosSkStr = "6924bdb57177f7ee6ab2a56a4b0dada921635376d8e69a3224855c997427ce85"
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
