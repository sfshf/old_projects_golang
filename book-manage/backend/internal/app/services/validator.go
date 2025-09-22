package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/nextsurfer/book-manage-api/internal/app/model"
)

func CheckEmpty(txt string) error {
	if txt == "" {
		return errors.New("string is empty")
	}
	return nil
}

func CheckInvalidCharacters(txt string) error {
	if strings.ContainsAny(txt, "()/") {
		return errors.New("string contains illegal characters")
	}
	return nil
}

func CheckPartOfSpeech(stringType, partOfSpeech string) error {
	var err error
	switch stringType {
	case "phrase":
		partOfSpeech = "phrase"
	case "word":
		parts := strings.Split(partOfSpeech, ",")
		for _, part := range parts {
			if strings.TrimSpace(part) != part {
				err = errors.New("partition of partOfSpeech contains illegal characters")
				return err
			}
			var valid bool
			for _, predefine := range []string{"noun", "pronoun", "verb", "adjective",
				"adverb", "preposition", "conjunction", "interjection", "article",
				"determiner", "predeterminer"} {
				if part == predefine {
					valid = true
				}
			}
			if !valid {
				err = errors.New("partOfSpeech contains illegal partition")
				return err
			}
		}
	default:
		err = errors.New("unsupported string type")
		return err
	}
	return nil
}

func CheckPronunciationIpa(pronunciationIpa string, str *model.String, oldDefinition *model.Definition) error {
	if str.Type == "word" {
		// 且string中不包含空格， 则Pronunciation ipa可以为空
		// 如果string中不包含空格，为单个单词，则Pronunciation ipa 不可为空
		if !strings.Contains(str.String, " ") && !strings.ContainsAny(str.String, "./_") {
			if pronunciationIpa == "" {
				return errors.New("pronunciation ipa not be empty when string is a single word")
			}
		}
		// 如果 Pronunciation ipa weak 或 Pronunciation ipa other 不为空，则Pronunciation ipa 不可为空
		if oldDefinition.PronunciationIpaWeak != "" || oldDefinition.PronunciationIpaOther != "" {
			if pronunciationIpa == "" {
				return errors.New("pronunciation ipa not be empty when pronunciation ipa weak or other is not empty")
			}
		}
	} else if str.Type == "phrase" {
		if pronunciationIpa != "" {
			return errors.New("pronunciation ipa must be empty when string type is phrase")
		}
	}
	return nil
}

func CheckPronunciationIpaWeak(pronunciationIpaWeak string, str *model.String, oldDefinition *model.Definition) error {
	if str.Type == "word" {
	} else if str.Type == "phrase" {
		if pronunciationIpaWeak != "" {
			return errors.New("pronunciation ipa weak must be empty when string type is phrase")
		}
	}
	return nil
}

func CheckPronunciationIpaOther(pronunciationIpaOther string, str *model.String, oldDefinition *model.Definition) error {
	if str.Type == "word" {
	} else if str.Type == "phrase" {
		if pronunciationIpaOther != "" {
			return errors.New("pronunciation ipa other must be empty when string type is phrase")
		}
	}
	return nil
}

func CheckPosition(position string, contentLength int64) error {
	splits := strings.Split(position, ",")
	if len(splits) == 0 || len(splits)%2 != 0 {
		return errors.New("invalid position")
	}
	for i := 0; i < len(splits); i += 2 {
		offset, err := strconv.ParseInt(splits[i], 10, 64)
		if err != nil {
			return err
		}
		limit, err := strconv.ParseInt(splits[i+1], 10, 64)
		if err != nil {
			return err
		}
		if offset < 0 || limit <= 0 || offset+limit > contentLength {
			err = fmt.Errorf("invalid position parameter: [%d,%d]", offset, limit)
			return err
		}
	}
	return nil
}

func CheckForm(partOfSpeech, form string) error {
	var valid bool
	parts := strings.Split(partOfSpeech, ",")
outer:
	for _, part := range parts {
		switch part {
		case "verb":
			for _, val := range []string{
				"present simple",
				"present participle",
				"past simple",
				"past participle",
				"also",
			} {
				if form == val {
					valid = true
					break outer
				}
			}
		case "noun":
			for _, val := range []string{
				"plural",
				"also",
			} {
				if form == val {
					valid = true
					break outer
				}
			}
		case "adjective", "adverb":
			for _, val := range []string{
				"comparative",
				"superlative",
				"also",
			} {
				if form == val {
					valid = true
					break outer
				}
			}
		default:
			if form == "also" {
				valid = true
			}
		}
	}

	if !valid {
		return errors.New("invalid form")
	}
	return nil
}
