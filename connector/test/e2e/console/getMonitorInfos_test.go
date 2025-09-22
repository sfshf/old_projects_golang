package console_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/connector/api/response"
)

func TestGetMonitorInfos(t *testing.T) {
	reqData := struct {
		ApiKey string `json:"apiKey"`
	}{
		ApiKey: _adminApiKey,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Infos []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"infos"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/riki/console/getMonitorInfos/v1", &reqData, nil, &respData, nil)
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
}

func TestGetMonitorInfosUnmatchedPermission(t *testing.T) {
	reqData := struct {
		ApiKey string `json:"apiKey"`
	}{
		ApiKey: _test2ApiKey,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Infos []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"infos"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/riki/console/getMonitorInfos/v1", &reqData, nil, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeUnmatchedPermission {
		t.Error("not prospective response data code")
		return
	}
}
