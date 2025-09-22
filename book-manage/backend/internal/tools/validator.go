package tools

import (
	"errors"
	"fmt"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/golang/glog"
	"reflect"
	"strings"
)

var (
	// Validate gloabal validator. use a single instance of Validate, it caches struct info
	Validate *validator.Validate
	trans    ut.Translator
)

// RegisterCustomValidation register custom validation
func RegisterCustomValidation() {
	Validate = validator.New()

	overwriteTranslation()
}

func overwriteTranslation() {
	// validate
	uni := ut.New(en.New())
	trans, _ = uni.GetTranslator("en")
	if err := en_translations.RegisterDefaultTranslations(Validate, trans); err != nil {
		glog.Errorln("ValidateStruct: validation translate init failed:", err)
	}
}

func ValidateStruct(r interface{}) error {
	if r == nil {
		return errors.New("empty params")
	}

	validateErr := Validate.Struct(r)

	if validateErr != nil {
		return GetOneErr(validateErr)
	}

	return nil
}

func GetOneErr(validateErr error) error {
	if reflect.TypeOf(validateErr).String() != "validator.ValidationErrors" {
		return validateErr
	}

	errs := validateErr.(validator.ValidationErrors)

	for _, e := range errs {
		var field string

		nameSpace := e.StructNamespace()
		nameSpaceArr := strings.Split(nameSpace, ".")
		if len(nameSpaceArr) > 1 {
			for k, nameSpacePart := range nameSpaceArr {
				if k == 0 {
					continue
				}

				nameSpacePart = strings.ToLower(nameSpacePart[0:1]) + nameSpacePart[1:] // 首字母转为小写
				field += nameSpacePart + " "
			}
			field = strings.TrimRight(field, " ")
		} else {
			field = strings.ToLower(e.Field()[0:1]) + e.Field()[1:] // 首字母转为小写
		}

		if e.Param() != "" {
			return fmt.Errorf(field + " validate failed with rule: " + e.ActualTag() + " " + e.Param())
		} else {
			return fmt.Errorf(field + " validate failed with rule: " + e.ActualTag())
		}
	}

	return nil
}

func GetOneErrStr(validateErr error) string {
	err := GetOneErr(validateErr)
	if err != nil {
		return err.Error()
	} else {
		return ""
	}
}
