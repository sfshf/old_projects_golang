package console_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/connector/api/response"
)

func TestGetManagePlatformLogs(t *testing.T) {
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
			Total int `json:"total"`
			List  []struct {
				CreatedAt  int64  `json:"createdAt"`
				Ip         string `json:"ip"`
				Status     string `json:"status"`
				Object     string `json:"object"`
				Operation  string `json:"operation"`
				KeyID      string `json:"keyID"`
				App        string `json:"app"`
				ApiKeyName string `json:"apiKeyName"`
			} `json:"list"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/riki/console/getManagePlatformLogs/v1", &reqData, nil, &respData, nil)
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

func TestGetManagePlatformLogs_EmptyApiKey(t *testing.T) {
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
	resp, err := postJsonRequest(_kongDNS+"/riki/console/getManagePlatformLogs/v1", &reqData, nil, &respData, nil)
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
