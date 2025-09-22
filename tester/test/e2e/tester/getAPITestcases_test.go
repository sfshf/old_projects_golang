package tester_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/tester/api/response"
)

func TestGetAPITestcases(t *testing.T) {
	reqData := struct {
		ApiKey string `json:"apiKey"`
		App    string `json:"app"`
	}{
		ApiKey: _testerApiKey,
		App:    "doom",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []struct {
				Name string `json:"name"`
				Path string `json:"path"`
				Body string `json:"body"`
			} `json:"list"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/tester/getAPITestcases/v1", &reqData, nil, &respData, nil)
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
