package tester_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/tester/api/response"
)

func TestGetMysqlInfo(t *testing.T) {
	reqData := struct {
		ApiKey string `json:"apiKey"`
	}{
		ApiKey: _testerApiKey,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []struct {
				Database string  `json:"database"`
				Size     float64 `json:"size"`
			} `json:"list"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/tester/getMysqlInfo/v1", &reqData, nil, &respData, nil)
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
	if len(respData.Data.List) == 0 {
		t.Error("not prospective response data")
		return
	}
}
