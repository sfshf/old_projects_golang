package console_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/connector/api/response"
	. "github.com/nextsurfer/connector/internal/pkg/model"
)

func TestCreateConfig(t *testing.T) {
	// mock data
	testApp := "connector-e2e-app-TestCreateConfig"
	testConfig := `{
	"appName": "connector-e2e-app-TestCreateConfig"
}`

	reqData := struct {
		ApiKey string `json:"apiKey"`
		App    string `json:"app"`
		Config string `json:"config"`
	}{
		ApiKey: _adminApiKey,
		App:    testApp,
		Config: testConfig,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/riki/console/createConfig/v1", &reqData, nil, &respData, nil)
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
		Where(`app = ? AND deleted_at = 0`, testApp).
		First(&appConfig).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		// clear test data
		if err := _connectorGormDB.Delete(&AppConfig{ID: appConfig.ID}).Error; err != nil {
			t.Error(err)
		}
	}()
}

func TestCreateConfig_EmptyApiKey(t *testing.T) {
	reqData := struct {
		ApiKey string `json:"apiKey"`
		App    string `json:"app"`
		Config string `json:"config"`
	}{}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/riki/console/createConfig/v1", &reqData, nil, &respData, nil)
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

func TestCreateConfig_ReadAccessWrite(t *testing.T) {
	reqData := struct {
		ApiKey string `json:"apiKey"`
		App    string `json:"app"`
		Config string `json:"config"`
	}{
		ApiKey: _test3ApiKey,
		App:    "test3",
		Config: `{"appName":"test3"}`,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/riki/console/createConfig/v1", &reqData, nil, &respData, nil)
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
}
