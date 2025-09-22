package crypto

import (
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"

	"github.com/klaytn/klaytn/crypto/sha3"
	"golang.org/x/crypto/chacha20poly1305"
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

func NonceZeroX() []byte {
	return make([]byte, chacha20poly1305.NonceSizeX)
}

func EncryptByX(plaintext []byte, aead cipher.AEAD, dst, nonce []byte) []byte {
	encryptedBytes := aead.Seal(dst, nonce, plaintext, nil)
	cipherbytes := make([]byte, base64.StdEncoding.EncodedLen(len(encryptedBytes)))
	base64.StdEncoding.Encode(cipherbytes, encryptedBytes)
	return cipherbytes
}

func DecryptByX(base64bytes []byte, aead cipher.AEAD, nonce []byte) ([]byte, error) {
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(base64bytes)))
	n, err := base64.StdEncoding.Decode(decoded, base64bytes)
	if err != nil {
		return nil, err
	}
	cipherbytes := decoded[:n]
	return aead.Open(nil, nonce, cipherbytes, nil)
}
