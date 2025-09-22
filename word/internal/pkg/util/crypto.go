package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

func padding(plaintext []byte, size int) []byte {
	mod := len(plaintext) % size
	if mod > 0 {
		pad := size - mod
		pads := bytes.Repeat([]byte{byte(pad)}, pad)
		plaintext = append(plaintext, pads...)
	}
	return plaintext
}

func unpadding(plaintext []byte) []byte {
	l := len(plaintext)
	pad := plaintext[l-1]
	return plaintext[:l-int(pad)]
}

func secretKey(key []byte) ([]byte, error) {
	return MD5(key, nil)
}

func AES16CBCEncrypt(plaintext, key []byte) ([]byte, error) {
	plaintext = padding(plaintext, aes.BlockSize)
	key, err := secretKey(key)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	return ciphertext, nil
}

func AES16CBCDecrypt(ciphertext, key []byte) ([]byte, error) {
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("cipher text too short")
	}
	key, err := secretKey(key)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("cipher text is not a multiple of the block size")
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(ciphertext, ciphertext)
	plaintext := unpadding(ciphertext)
	return plaintext, nil
}
