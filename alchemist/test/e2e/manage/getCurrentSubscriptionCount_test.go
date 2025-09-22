package manage_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/alchemist/api/response"
)

func TestGetCurrentSubscriptionCount(t *testing.T) {
	// mock data
	testAppID := "alchemist"

	reqData := struct {
		Password string `json:"password"`
		AppID    string `json:"appID"`
	}{
		Password: _alchemistApiKey,
		AppID:    testAppID,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Count int64 `json:"count"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/console/getCurrentSubscriptionCount/v1", &reqData, nil, &respData, nil)
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

func TestGetCurrentSubscriptionCount_EmptyParameter(t *testing.T) {
	reqData := struct {
		Password string `json:"password"`
		AppID    string `json:"appID"`
	}{}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Count int64 `json:"count"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/console/getCurrentSubscriptionCount/v1", &reqData, nil, &respData, nil)
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
