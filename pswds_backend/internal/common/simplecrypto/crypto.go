package simplecrypto

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"

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

func NewNonceAndX(secretKey []byte) ([]byte, cipher.AEAD, error) {
	aead, err := chacha20poly1305.NewX(secretKey)
	if err != nil {
		return nil, nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}
	return nonce, aead, nil
}

func EncryptByX(plaintext []byte, aead cipher.AEAD, nonce []byte) []byte {
	encryptedBytes := aead.Seal(nil, nonce, plaintext, nil)
	cipherbytes := make([]byte, base64.StdEncoding.EncodedLen(len(encryptedBytes)))
	base64.StdEncoding.Encode(cipherbytes, encryptedBytes)
	return cipherbytes
}

func DecryptByX(base64bytes []byte, aead cipher.AEAD) ([]byte, error) {
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(base64bytes)))
	n, err := base64.StdEncoding.Decode(decoded, base64bytes)
	if err != nil {
		return nil, err
	}
	cipherbytes := decoded[:n]
	if len(cipherbytes) < aead.NonceSize() {
		return nil, errors.New("length of cipherbytes too short")
	}
	nonce, ciphertext := cipherbytes[:aead.NonceSize()], cipherbytes[aead.NonceSize():]
	return aead.Open(nil, nonce, ciphertext, nil)
}
