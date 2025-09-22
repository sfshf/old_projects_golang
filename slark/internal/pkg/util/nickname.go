package util

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	mrand "math/rand"
	"os"
	"regexp"
	"strings"
	"sync"
)

var (
	_adjectives   []string
	_nouns        []string
	_nicknameOnce sync.Once
)

func nickname_delay_init() error {
	var err error
	_nicknameOnce.Do(func() {
		adjFilepath := strings.TrimSpace(os.Getenv("ADJECTIVES_JSON_FILE"))
		if adjFilepath == "" {
			adjFilepath = "/etc/application/json/adjectives.json"
		}
		var adjFile *os.File
		adjFile, err = os.Open(adjFilepath)
		if err != nil {
			return
		}
		defer adjFile.Close()
		var adjFileContent []byte
		adjFileContent, err = io.ReadAll(adjFile)
		if err != nil {
			return
		}
		if err = json.Unmarshal(adjFileContent, &_adjectives); err != nil {
			return
		}
		nounFilepath := strings.TrimSpace(os.Getenv("NOUNS_JSON_FILE"))
		if nounFilepath == "" {
			nounFilepath = "/etc/application/json/nouns.json"
		}
		var nounFile *os.File
		nounFile, err = os.Open(nounFilepath)
		if err != nil {
			return
		}
		defer nounFile.Close()
		var nounFileContent []byte
		nounFileContent, err = io.ReadAll(nounFile)
		if err != nil {
			return
		}
		if err = json.Unmarshal(nounFileContent, &_nouns); err != nil {
			return
		}
	})
	return err
}

func GetRandomInt(min, max int32) (int32, error) {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		return -1, err
	}
	randomInt := int32(b[0]<<24 + b[1]<<16 + b[2]<<8 + b[3])
	return min + (randomInt % (max - min + 1)), nil
}

func randomNumber(maxNumber int) (string, error) {
	var err error
	var randomInt int32
	switch maxNumber {
	case 1:
		randomInt, err = GetRandomInt(1, 9)
		if err != nil {
			return "", err
		}
	case 2:
		randomInt, err = GetRandomInt(10, 90)
		if err != nil {
			return "", err
		}
	case 3:
		randomInt, err = GetRandomInt(100, 900)
		if err != nil {
			return "", err
		}
	case 4:
		randomInt, err = GetRandomInt(1000, 9000)
		if err != nil {
			return "", err
		}
	case 5:
		randomInt, err = GetRandomInt(10000, 90000)
		if err != nil {
			return "", err
		}
	case 6:
		randomInt, err = GetRandomInt(100000, 900000)
		if err != nil {
			return "", err
		}
	}
	if randomInt != 0 {
		return fmt.Sprintf("%d", randomInt), nil
	}
	return "", nil
}

func GenerateNickname(separator string, randomDigits, length int, prefix string) (string, error) {
	if err := nickname_delay_init(); err != nil {
		return "", err
	}
	noun := _nouns[mrand.Intn(len(_nouns))]
	var adjective string
	if prefix != "" {
		reg1 := regexp.MustCompile(`\s{2,}`)
		reg2 := regexp.MustCompile(`\s`)
		adjective = reg2.ReplaceAllLiteralString(reg1.ReplaceAllLiteralString(prefix, " "), "")
	} else {
		adjective = _adjectives[mrand.Intn(len(_adjectives))]
	}
	num, err := randomNumber(randomDigits)
	if err != nil {
		return "", err
	}
	var nickname string
	if separator != "" {
		nickname = adjective + strings.TrimSpace(separator) + noun + num
	} else {
		nickname = adjective + strings.Title(noun) + num
	}
	if length > 0 {
		return nickname[:length], nil
	}
	return nickname, nil
}
