package util

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strings"
)

func GenerateReferralCode() string {
	rands, err := randomBytesMod(8, 36)
	if err != nil {
		return ""
	}
	var res strings.Builder
	for _, rand := range rands {
		if rand < 10 {
			res.WriteRune(rune(rand + 48))
		} else {
			res.WriteRune(rune(rand + 87))
		}
	}
	return res.String()
}

func randomBytesMod(length int, mod byte) ([]byte, error) {
	if length <= 0 {
		return nil, errors.New("length must be greater than zero")
	}
	if mod <= 0 {
		return nil, errors.New("captcha: bad mod argument for randomBytesMod")
	}
	maxrb := 255 - byte(256%int(mod))
	b := make([]byte, length)
	i := 0
	for {
		r, err := randomBytes(length + (length / 4))
		if err != nil {
			return nil, err
		}
		for _, c := range r {
			if c > maxrb {
				// Skip this number to avoid modulo bias.
				continue
			}
			b[i] = c % mod
			i++
			if i == length {
				return b, nil
			}
		}
	}
}

func randomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, fmt.Errorf("captcha: error reading random source: %w", err)
	}
	return b, nil
}
