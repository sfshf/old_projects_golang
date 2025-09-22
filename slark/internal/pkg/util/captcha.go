package util

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dchest/captcha"
	"github.com/go-redis/redis/v8"
	"github.com/nextsurfer/ground/pkg/rpc"
)

const (
	PrefixLoginCode        = "LoginCode:"
	PrefixRegistrationCode = "RegistrationCode:"
	PrefixQRLoginToken     = "QRLoginToken:"
)

type emailCaptcha struct {
	Email   string
	Captcha string
}

func GetRegistrationEmailCaptchasInRedis(ctx context.Context, rpcCtx *rpc.Context, client *redis.Client) ([]emailCaptcha, error) {
	return GetEmailCaptchasInRedis(ctx, rpcCtx, client, PrefixRegistrationCode+"*")
}

func GetLoginEmailCaptchasInRedis(ctx context.Context, rpcCtx *rpc.Context, client *redis.Client) ([]emailCaptcha, error) {
	return GetEmailCaptchasInRedis(ctx, rpcCtx, client, PrefixLoginCode+"*")
}

func GetEmailCaptchasInRedis(ctx context.Context, rpcCtx *rpc.Context, client *redis.Client, match string) ([]emailCaptcha, error) {
	var res []emailCaptcha
	if err := ScanKeys(ctx, rpcCtx, client, match, func(key string) error {
		one, err := GetEmailCaptcha(ctx, rpcCtx, client, key)
		if err != nil {
			return err
		}
		res = append(res, one)
		return nil
	}); err != nil {
		return nil, err
	}
	return res, nil
}

func GetEmailCaptcha(ctx context.Context, rpcCtx *rpc.Context, client *redis.Client, key string) (res emailCaptcha, err error) {
	captcha, err := client.Get(ctx, key).Result()
	if err != nil {
		return
	}
	splits := strings.Split(key, ":")
	if len(splits) != 3 {
		err = errors.New("invalid redis key of login email captcha")
		return
	}
	res.Email = splits[2]
	res.Captcha = captcha
	return
}

func ScanKeys(ctx context.Context, rpcCtx *rpc.Context, client *redis.Client, match string, handleKey func(key string) error) error {
	var keys []string
	var cursor uint64
	var err error
	for {
		keys, cursor, err = client.Scan(ctx, cursor, match, 0).Result()
		if err != nil {
			return err
		}
		for _, key := range keys {
			if err := handleKey(key); err != nil {
				return err
			}
		}
		if cursor == 0 {
			break
		}
	}
	return nil
}

func StoreQRLoginTokenInRedis(ctx context.Context, rpcCtx *rpc.Context, client *redis.Client, token string) error {
	return client.Set(ctx, PrefixQRLoginToken+token, rpcCtx.SessionID, time.Minute*5).Err()
}

func QRLoginTokenExistsInRedis(ctx context.Context, rpcCtx *rpc.Context, client *redis.Client, token string) error {
	exists := client.Exists(ctx, PrefixQRLoginToken+token)
	if err := exists.Err(); err != nil {
		return err
	}
	if exists.Val() == 0 {
		err := errors.New("invalid login token")
		return err
	}
	return nil
}

func ValidateRegistrationCodeInRedis(ctx context.Context, rpcCtx *rpc.Context, client *redis.Client, email, captcha string) error {
	return ValidateEmailCaptchaInRedis(ctx, rpcCtx, client, PrefixRegistrationCode, email, captcha)
}

func ValidateLoginCodeInRedis(ctx context.Context, rpcCtx *rpc.Context, client *redis.Client, email, captcha string) error {
	return ValidateEmailCaptchaInRedis(ctx, rpcCtx, client, PrefixLoginCode, email, captcha)
}

func ValidateEmailCaptchaInRedis(ctx context.Context, rpcCtx *rpc.Context, client *redis.Client, prefix, email, captcha string) error {
	key := prefix + rpcCtx.DeviceID + ":" + email
	captchaInRedis, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("captcha %s has been expired", captcha)
	} else if err != nil {
		return err
	}
	if captchaInRedis != captcha {
		return errors.New("the input captcha code is fault")
	} else {
		if err := client.Del(ctx, key).Err(); err != nil {
			return err
		}
	}
	return nil
}

func RandomEmail() string {
	nickname, _ := GenerateNickname("", 0, 0, "")
	return nickname + "@nextsurfer.com"
}

func RandomCaptcha() string {
	randomDigits := captcha.RandomDigits(6)
	return fmt.Sprintf("%d%d%d%d%d%d",
		randomDigits[0],
		randomDigits[1],
		randomDigits[2],
		randomDigits[3],
		randomDigits[4],
		randomDigits[5],
	)
}

func StoreRegistrationCodeInRedis(ctx context.Context, rpcCtx *rpc.Context, client *redis.Client, email, captcha string) error {
	return client.Set(ctx, PrefixRegistrationCode+rpcCtx.DeviceID+":"+email, captcha, time.Minute*5).Err()
}

func StoreLoginCodeInRedis(ctx context.Context, rpcCtx *rpc.Context, client *redis.Client, email, captcha string) error {
	return client.Set(ctx, PrefixLoginCode+rpcCtx.DeviceID+":"+email, captcha, time.Minute*5).Err()
}

func GenerateDigitCaptchaWithStoreFuncs(ctx context.Context, storeFuncs ...func(captcha string) error) (string, error) {
	captcha := RandomCaptcha()
	for _, storeFunc := range storeFuncs {
		if err := storeFunc(captcha); err != nil {
			return "", err
		}
	}
	return captcha, nil
}
