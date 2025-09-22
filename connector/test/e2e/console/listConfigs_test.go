package console_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/connector/api/response"
	. "github.com/nextsurfer/connector/internal/pkg/model"
)

func TestListConfigs(t *testing.T) {
	// mock data
	testApp := "connector-e2e-app-TestListConfigs"

	newAppConfig := AppConfig{
		App: testApp,
		Config: `{
	"appName": "connector-e2e-app-TestListConfigs"
}`,
	}
	if err := _connectorGormDB.Create(&newAppConfig).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		// clear test data
		if err := _connectorGormDB.Delete(&AppConfig{ID: newAppConfig.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()

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
			ConfigList []struct {
				App    string `json:"app"`
				Config string `json:"config"`
			} `json:"configList"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/riki/console/listConfigs/v1", &reqData, nil, &respData, nil)
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
	// check data
	if len(respData.Data.ConfigList) < 1 {
		t.Error("not prospective response data")
		return
	}
}

func TestListConfigs_EmptyApiKey(t *testing.T) {
	reqData := struct {
		ApiKey string `json:"apiKey"`
	}{}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/riki/console/listConfigs/v1", &reqData, nil, &respData, nil)
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

func TestListConfigs_WrongApiKey(t *testing.T) {
	reqData := struct {
		ApiKey string `json:"apiKey"`
	}{
		ApiKey: "TestNotExistApiKey",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			ConfigList []struct {
				App    string `json:"app"`
				Config string `json:"config"`
			} `json:"configList"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/riki/console/listConfigs/v1", &reqData, nil, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeInvalidApiKey {
		t.Error("not prospective response data code")
		return
	}
	if respData.Message != "ApiKey is invalid. Please input apiKey correctly." {
		t.Error("not prospective response data")
		return
	}
}
