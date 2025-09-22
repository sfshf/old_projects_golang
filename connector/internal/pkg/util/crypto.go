package util

import (
	"encoding/hex"

	"github.com/klaytn/klaytn/crypto/sha3"
)

func Keccak256(src []byte) ([]byte, error) {
	h := sha3.NewKeccak256()
	if _, err := h.Write(src); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func Keccak256Hex(src []byte) ([]byte, error) {
	sum, err := Keccak256(src)
	if err != nil {
		return nil, err
	}
	dst := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(dst, sum)
	return dst, nil
}
