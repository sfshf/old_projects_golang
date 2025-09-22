package util_test

import (
	"context"
	"testing"

	"github.com/nextsurfer/slark/internal/pkg/util"
)

func TestGenerateDigitCaptchaWithStoreFuncs(t *testing.T) {
	ctx := context.TODO()
	testEmail := "test@example.com"
	store := make(map[string]string)
	storeKey := "UnitTest-DeviceID" + "/" + testEmail
	captcha, err := util.GenerateDigitCaptchaWithStoreFuncs(ctx, func(captcha string) error {
		store[storeKey] = captcha
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if store[storeKey] != captcha {
		t.Error("captcha in the store is invalid")
	}
}
