package simplehash

import (
	"crypto/md5"
	"encoding/hex"
)

func HexMd5ToString(data []byte) string {
	h := md5.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
