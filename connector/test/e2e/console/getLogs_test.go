package console_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/connector/api/response"
)

func TestGetLogs(t *testing.T) {
	reqData := struct {
		ApiKey     string `json:"apiKey"`
		PageNumber int    `json:"pageNumber"`
		PageSize   int    `json:"pageSize"`
	}{
		ApiKey:     _adminApiKey,
		PageNumber: 1,
		PageSize:   1,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Total int      `json:"total"`
			List  []string `json:"list"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/riki/console/getLogs/v1", &reqData, nil, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	if respData.Data.Total <= 0 {
		t.Error("not prospective response data")
		return
	}
}

func TestGetLogs_EmptyApiKey(t *testing.T) {
	reqData := struct {
		ApiKey     string `json:"apiKey"`
		PageNumber int    `json:"pageNumber"`
		PageSize   int    `json:"pageSize"`
	}{
		PageNumber: 1,
		PageSize:   1,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Total int      `json:"total"`
			List  []string `json:"list"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/riki/console/getLogs/v1", &reqData, nil, &respData, nil)
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
}
