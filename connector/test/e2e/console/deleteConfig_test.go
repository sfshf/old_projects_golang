package console_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/connector/api/response"
	. "github.com/nextsurfer/connector/internal/pkg/model"
)

func TestDeleteConfig(t *testing.T) {
	// mock data
	testApp := "connector-e2e-app-TestDeleteConfig"

	newAppConfig := AppConfig{
		App: testApp,
		Config: `{
	"appName": "connector-e2e-app-TestDeleteConfig"
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
		App    string `json:"app"`
	}{
		ApiKey: _adminApiKey,
		App:    testApp,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/riki/console/deleteConfig/v1", &reqData, nil, &respData, nil)
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
	var appConfig AppConfig
	if err := _connectorGormDB.Table(TableNameAppConfig).
		Where(`app = ? AND deleted_at > 0`, testApp).
		First(&appConfig).Error; err != nil {
		t.Error(err)
		return
	}
}

func TestDeleteConfig_EmptyApiKey(t *testing.T) {
	reqData := struct {
		ApiKey string `json:"apiKey"`
		App    string `json:"app"`
	}{}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/riki/console/deleteConfig/v1", &reqData, nil, &respData, nil)
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
