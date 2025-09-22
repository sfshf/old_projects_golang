package utils_test

import (
	"bytes"
	"encoding/base64"
	"log"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	ecies "github.com/ecies/go/v2"
	"github.com/nextsurfer/keystore/internal/app/utils"
	"golang.org/x/crypto/chacha20poly1305"
)

func TestSystemKey(t *testing.T) {
	secretKey, err := utils.NewSecretKey()
	if err != nil {
		t.Error(err)
		return
	}
	nonce, _, err := utils.NewNonceAndX(secretKey)
	if err != nil {
		t.Error(err)
		return
	}
	systemKey, err := utils.SystemKey(secretKey, nonce)
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("SystemKey:", systemKey)
	parsedSecretKey, parsedNonce, err := utils.ParseSystemKey(systemKey)
	if err != nil {
		t.Error(err)
		return
	}

	if !bytes.Equal(secretKey, parsedSecretKey) {
		t.Error("parsed secret key not equal")
		return
	}
	if !bytes.Equal(nonce, parsedNonce) {
		t.Error("parsed nonce not equal")
		return
	}
}

func TestEncryptPassword(t *testing.T) {
	password, err := utils.NewPassword()
	if err != nil {
		t.Error(err)
		return
	}
	log.Printf("%s\n", password)
	secretKey, err := utils.NewSecretKey()
	if err != nil {
		t.Error(err)
		return
	}
	nonce, aead, err := utils.NewNonceAndX(secretKey)
	if err != nil {
		t.Error(err)
		return
	}
	encryptedPassword := utils.EncryptByX([]byte(password), aead, nonce)
	log.Printf("encryptedPassword: %s\n", encryptedPassword)
}

func TestSecp256k1(t *testing.T) {
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Error(err)
		return
	}
	compressedPublicKey := privKey.PubKey().SerializeUncompressed()
	log.Println(string(compressedPublicKey))
	log.Println(base64.StdEncoding.EncodeToString(compressedPublicKey))

	eciesPrivKey := ecies.NewPrivateKeyFromBytes(privKey.Serialize())
	ciphertext, err := ecies.Encrypt(eciesPrivKey.PublicKey, []byte("THIS IS THE TEST"))
	if err != nil {
		panic(err)
	}
	log.Printf("plaintext encrypted: %v\n", ciphertext)

	plaintext, err := ecies.Decrypt(eciesPrivKey, ciphertext)
	if err != nil {
		panic(err)
	}
	log.Printf("ciphertext decrypted: %s\n", string(plaintext))
}

func TestDeEncryptByX(t *testing.T) {
	secretKey, err := utils.NewSecretKey()
	if err != nil {
		t.Error(err)
		return
	}
	nonce, aead, err := utils.NewNonceAndX(secretKey)
	if err != nil {
		t.Error(err)
		return
	}
	systemKey, err := utils.SystemKey(secretKey, nonce)
	if err != nil {
		t.Error(err)
		return
	}
	log.Println("SystemKey:", systemKey)

	plaintext := "this is a plaintext"
	encryptedtext := utils.EncryptByX([]byte(plaintext), aead, nonce)

	oldSecretKey, _, err := utils.ParseSystemKey(systemKey)
	if err != nil {
		t.Error(err)
		return
	}

	oldAead, err := chacha20poly1305.NewX(oldSecretKey)
	if err != nil {
		t.Error(err)
		return
	}
	plaintext2, err := utils.DecryptByX(encryptedtext, oldAead)
	if err != nil {
		t.Error(err)
		return
	}

	log.Println(plaintext)
	log.Println(string(plaintext2))

}
