package util_test

import (
	"context"
	"testing"

	"github.com/nextsurfer/slark/internal/pkg/util"
)

func TestSendEmailCaptcha(t *testing.T) {
	ctx := context.TODO()
	testEmail := "gavin@n1xt.net"
	captcha, err := util.GenerateDigitCaptchaWithStoreFuncs(ctx, func(captcha string) error {
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if err := util.SendCaptchaEmail(ctx, testEmail, captcha); err != nil {
		t.Error(err)
	}
}
