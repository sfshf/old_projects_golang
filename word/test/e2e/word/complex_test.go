package word_test

import (
	"net/http"
	"testing"

	slark_response "github.com/nextsurfer/slark/api/response"
	"github.com/nextsurfer/word/api/response"
)

func TestI18nEnglish(t *testing.T) {
	reqData := struct {
		DefinitionID int64 `json:"definitionID"`
	}{}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/word/user/definition/favorite/v1", &reqData, nil, &respData, nil)
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
	if respData.Message != "Request has wrong parameters, please inspect parameters" {
		t.Error("not prospective response data")
		return
	}
}

func TestI18nChinese(t *testing.T) {
	reqData := struct {
		DefinitionID int64 `json:"definitionID"`
	}{}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/word/user/definition/favorite/v1", &reqData, nil, &respData, func(req *http.Request) {
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
	if respData.Message != "请求参数有误，请您排错后重试" {
		t.Error("not prospective response data")
		return
	}
}

func TestAcceptLanguageEnglish(t *testing.T) {
	reqData := struct {
		DefinitionID int64 `json:"definitionID"`
	}{
		DefinitionID: 100000000,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/word/user/definition/favorite/v1", &reqData, nil, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != slark_response.StatusCodeEmptySession {
		t.Error("not prospective response data code")
		return
	}
	if respData.Message != "Empty session. You need login." {
		t.Error("not prospective response data")
		return
	}
}

func TestAcceptLanguageChinese(t *testing.T) {
	reqData := struct {
		DefinitionID int64 `json:"definitionID"`
	}{
		DefinitionID: 100000000,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/word/user/definition/favorite/v1", &reqData, nil, &respData, func(req *http.Request) {
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
	if respData.Code != slark_response.StatusCodeEmptySession {
		t.Error("not prospective response data code")
		return
	}
	if respData.Message != "空会话，请先登录" {
		t.Error("not prospective response data")
		return
	}
}
