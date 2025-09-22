package doom_console_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/doom-console/api/response"
)

func TestI18nEnglish(t *testing.T) {
	reqData := struct {
		ApiKey string `json:"apiKey"`
	}{}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/doom/console/listReputableTokens/v1", &reqData, nil, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeUnauthorized {
		t.Error("not prospective response data code")
		return
	}
	if respData.Message != "ApiKey is invalid. Please input apiKey correctly." {
		t.Error("not prospective response data")
		return
	}
}

func TestI18nChinese(t *testing.T) {
	reqData := struct {
		ApiKey string `json:"apiKey"`
	}{}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/doom/console/listReputableTokens/v1", &reqData, nil, &respData, func(req *http.Request) {
		req.Header.Set("Accept-Language", "zh")
	})
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeUnauthorized {
		t.Error("not prospective response data code")
		return
	}
	if respData.Message != "ApiKey无效，请正确输入apiKey" {
		t.Error("not prospective response data")
		return
	}
}
