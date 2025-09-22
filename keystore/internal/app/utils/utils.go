package utils

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	ecies "github.com/ecies/go/v2"
	"github.com/klaytn/klaytn/crypto/sha3"
	"github.com/syndtr/goleveldb/leveldb"
	"golang.org/x/crypto/chacha20poly1305"
)

const (
	PasswordLength = 16
)

func ParseSystemKey(systemKey string) ([]byte, []byte, error) {
	key, err := base64.StdEncoding.DecodeString(systemKey)
	if err != nil {
		return nil, nil, err
	}
	buf := bytes.NewBuffer(key)
	return buf.Next(chacha20poly1305.KeySize), buf.Next(buf.Len()), nil
}

func SystemKey(secretKey, nonce []byte) (string, error) {
	buf := bytes.NewBuffer(secretKey)
	n, err := buf.Write(nonce)
	if err != nil {
		return "", err
	}
	if n != len(nonce) {
		return "", errors.New("secret key append nonce fail")
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func NewSecretKey() ([]byte, error) {
	key := make([]byte, chacha20poly1305.KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

func NewNonceAndX(secretKey []byte) ([]byte, cipher.AEAD, error) {
	aead, err := chacha20poly1305.NewX(secretKey)
	if err != nil {
		return nil, nil, err
	}
	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+PasswordLength+aead.Overhead())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}
	return nonce, aead, nil
}

func NewPassword() (string, error) {
	rands := randomBytesMod(8, 36)
	var buf bytes.Buffer
	for _, rand := range rands {
		if rand < 10 {
			buf.WriteRune(rune(rand + 48))
		} else {
			buf.WriteRune(rune(rand + 87))
		}
	}
	suffix := fmt.Sprintf("%d%s", time.Now().UnixMilli(), "9C9B913EB1B6254F4737CE947EFD16F16E916F9D6EE5C1102A2002E48D4C88BD")
	buf.WriteString(suffix)
	randomBytes, err := Keccak256(buf.Bytes())
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(randomBytes)[:24], nil
}

func randomBytesMod(length int, mod byte) (b []byte) {
	if length == 0 {
		return nil
	}
	if mod == 0 {
		panic("captcha: bad mod argument for randomBytesMod")
	}
	maxrb := 255 - byte(256%int(mod))
	b = make([]byte, length)
	i := 0
	for {
		r := randomBytes(length + (length / 4))
		for _, c := range r {
			if c > maxrb {
				// Skip this number to avoid modulo bias.
				continue
			}
			b[i] = c % mod
			i++
			if i == length {
				return
			}
		}
	}
}

func randomBytes(length int) (b []byte) {
	b = make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic("captcha: error reading random source: " + err.Error())
	}
	return
}

func EncryptByX(plaintext []byte, aead cipher.AEAD, nonce []byte) []byte {
	encryptedBytes := aead.Seal(nonce, nonce, plaintext, nil)
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

func DecryptByEcies(privKeyBytes, data []byte) ([]byte, error) {
	privKey := ecies.NewPrivateKeyFromBytes(privKeyBytes)
	return ecies.Decrypt(privKey, data)
}

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

type StoredKey interface {
	Key() []byte
	SetKey([]byte)
}

type StoredPasswordKey struct {
	Password     []byte `json:"password"`
	PasswordHash []byte `json:"passwordHash"`
}

func (a StoredPasswordKey) Key() []byte {
	return a.Password
}

func (a *StoredPasswordKey) SetKey(key []byte) {
	a.Password = key
}

func PutStoredPasswordKey(keyDB *leveldb.DB, keyID, password, passwordHash []byte) error {
	storedKey := StoredPasswordKey{
		Password:     password,
		PasswordHash: passwordHash,
	}
	storedKeyBytes, err := json.Marshal(storedKey)
	if err != nil {
		return err
	}
	if err := keyDB.Put(keyID, storedKeyBytes, nil); err != nil {
		return err
	}
	return nil
}

func GetStoredPasswordKey(keyDB *leveldb.DB, keyID []byte) (*StoredPasswordKey, error) {
	storedKeyBytes, err := keyDB.Get(keyID, nil)
	if err != nil {
		return nil, err
	}
	var res StoredPasswordKey
	if err := json.Unmarshal(storedKeyBytes, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

type StoredPrivateKey struct {
	Private []byte `json:"private"`
	Public  []byte `json:"public"`
}

func (a StoredPrivateKey) Key() []byte {
	return a.Private
}

func (a *StoredPrivateKey) SetKey(key []byte) {
	a.Private = key
}

func PutStoredPrivateKey(keyDB *leveldb.DB, keyID, private, public []byte) error {
	storedKey := StoredPrivateKey{
		Private: private,
		Public:  public,
	}
	storedKeyBytes, err := json.Marshal(storedKey)
	if err != nil {
		return err
	}
	if err := keyDB.Put(keyID, storedKeyBytes, nil); err != nil {
		return err
	}
	return nil
}

func GetStoredPrivateKey(keyDB *leveldb.DB, keyID []byte) (*StoredPrivateKey, error) {
	storedKeyBytes, err := keyDB.Get(keyID, nil)
	if err != nil {
		return nil, err
	}
	var res StoredPrivateKey
	if err := json.Unmarshal(storedKeyBytes, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func CopyDir(src string, dest string) (err error) {
	if err = os.MkdirAll(dest, os.ModePerm); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())
		if entry.IsDir() {
			if err = CopyDir(srcPath, destPath); err != nil {
				return err
			}
		} else {
			if err = CopyFile(srcPath, destPath); err != nil {
				return err
			}
		}
	}
	return
}

func CopyFile(src, dest string) (err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}
	return
}

func GeneratePrivateKey() (*secp256k1.PrivateKey, error) {
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}
	privKeyBytes := privKey.Serialize()
	suffix := fmt.Sprintf("%d%s", time.Now().UnixMilli(), "9C9B913EB1B6254F4737CE947EFD16F16E916F9D6EE5C1102A2002E48D4C88BD")
	privKeyBytes = append(privKeyBytes, []byte(suffix)...)
	privKeyBytes, err = Keccak256(privKeyBytes)
	if err != nil {
		return nil, err
	}
	return secp256k1.PrivKeyFromBytes(privKeyBytes), nil
}
