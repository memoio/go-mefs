package memoriae

import (
	"crypto/sha256"
	"encoding/hex"
)

func bytesToSha256(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return "0x" + hex.EncodeToString(h.Sum(nil))
}
