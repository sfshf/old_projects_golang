package cipher

import (
	ecies "github.com/ecies/go/v2"
	"github.com/klaytn/klaytn/crypto/sha3"
)

func Keccak256(src []byte) ([]byte, error) {
	h := sha3.NewKeccak256()
	if _, err := h.Write(src); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func EncryptByEcies(pubKeyHex string, plaintext []byte) ([]byte, error) {
	pubKey, err := ecies.NewPublicKeyFromHex(pubKeyHex)
	if err != nil {
		return nil, err
	}
	return ecies.Encrypt(pubKey, plaintext)
}

func DecryptByEcies(privKeyBytes, ciphertext []byte) ([]byte, error) {
	privKey := ecies.NewPrivateKeyFromBytes(privKeyBytes)
	return ecies.Decrypt(privKey, ciphertext)
}
