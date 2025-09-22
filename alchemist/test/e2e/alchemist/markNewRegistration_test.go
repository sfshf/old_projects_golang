package alchemist_test

import (
	"net/http"
	"testing"

	slark_response "github.com/nextsurfer/slark/api/response"
)

func TestMarkNewRegistration_EmptySession(t *testing.T) {
	// mock data
	reqData := struct {
		AppID        string `json:"appID"`
		DeviceToken0 string `json:"deviceToken0"`
		DeviceToken1 string `json:"deviceToken1"`
	}{}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/markNewRegistration/v1", &reqData, nil, &respData, nil)
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
}
