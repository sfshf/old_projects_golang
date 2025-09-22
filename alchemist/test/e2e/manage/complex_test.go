package manage_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/alchemist/api/response"
)

func TestI18nEnglish(t *testing.T) {
	reqData := struct {
		Password string `json:"password"`
		Config   string `json:"config"`
	}{}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/console/createConfig/v1", &reqData, nil, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
	// check data
	if respData.Message != "Request has wrong parameters, please inspect parameters" {
		t.Error("not prospective response data")
		return
	}
}

func TestI18nChinese(t *testing.T) {
	reqData := struct {
		Password string `json:"password"`
		Config   string `json:"config"`
	}{}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/console/createConfig/v1", &reqData, nil, &respData, func(req *http.Request) {
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
	if respData.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
	// check data
	if respData.Message != "请求参数有误，请您排错后重试" {
		t.Error("not prospective response data")
		return
	}
}
