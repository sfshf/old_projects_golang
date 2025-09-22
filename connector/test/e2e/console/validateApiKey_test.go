package console_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/connector/api/response"
	. "github.com/nextsurfer/connector/internal/pkg/model"
	"github.com/nextsurfer/connector/internal/pkg/util"
)

func TestValidateApiKey(t *testing.T) {
	var (
		testApp    = "app_TestValidateApiKey"
		testKeyID  = "pswd_TestValidateApiKey"
		testApiKey = "testApiKey_TestValidateApiKey"
	)
	newRelationAppKey := RelationAppKey{
		App:          testApp,
		KeyID:        testKeyID,
		PasswordHash: testApiKey,
	}
	if err := _connectorGormDB.Create(&newRelationAppKey).Error; err != nil {
		t.Error(err)
		return
	}
	newAPIKey := APIKey{
		App:        testApp,
		KeyID:      testKeyID,
		Name:       testApp,
		Permission: util.PermWrite,
	}
	if err := _connectorGormDB.Create(&newAPIKey).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		// clear test data
		if err := _connectorGormDB.Delete(&RelationAppKey{ID: newRelationAppKey.ID}).Error; err != nil {
			t.Error(err)
			return
		}
		if err := _connectorGormDB.Delete(&APIKey{ID: newAPIKey.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()

	reqData := struct {
		ApiKey string `json:"apiKey"`
		App    string `json:"app"`
		Role   string `json:"role"`
	}{
		ApiKey: testApiKey,
		App:    testApp,
		Role:   util.PermWrite,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Valid bool `json:"valid"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/riki/console/validateApiKey/v1", &reqData, nil, &respData, nil)
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
	if !respData.Data.Valid {
		t.Error("not prospective response data")
		return
	}
}

func TestValidateApiKey_WrongApiKey(t *testing.T) {
	var (
		testApp    = "app_TestValidateApiKey_WrongApiKey"
		testKeyID  = "pswd_TestValidateApiKey_WrongApiKey"
		testApiKey = "testApiKey_TestValidateApiKey_WrongApiKey"
	)
	newRelationAppKey := RelationAppKey{
		App:          testApp,
		KeyID:        testKeyID,
		PasswordHash: testApiKey,
	}
	if err := _connectorGormDB.Create(&newRelationAppKey).Error; err != nil {
		t.Error(err)
		return
	}
	newAPIKey := APIKey{
		App:        testApp,
		KeyID:      testKeyID,
		Name:       testApp,
		Permission: util.PermWrite,
	}
	if err := _connectorGormDB.Create(&newAPIKey).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		// clear test data
		if err := _connectorGormDB.Delete(&RelationAppKey{ID: newRelationAppKey.ID}).Error; err != nil {
			t.Error(err)
			return
		}
		if err := _connectorGormDB.Delete(&APIKey{ID: newAPIKey.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()

	reqData := struct {
		ApiKey string `json:"apiKey"`
		App    string `json:"app"`
		Role   string `json:"role"`
	}{
		ApiKey: testApiKey + "_fault",
		App:    testApp,
		Role:   util.PermWrite,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Valid bool `json:"valid"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/riki/console/validateApiKey/v1", &reqData, nil, &respData, nil)
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
	if respData.Data.Valid {
		t.Error("not prospective response data")
		return
	}
}

func TestValidateApiKey_ReadAccessWrite(t *testing.T) {
	var (
		testApp    = "app_TestValidateApiKey_ReadAccessWrite"
		testKeyID  = "pswd_TestValidateApiKey_ReadAccessWrite"
		testApiKey = "testApiKey_TestValidateApiKey_ReadAccessWrite"
	)
	newRelationAppKey := RelationAppKey{
		App:          testApp,
		KeyID:        testKeyID,
		PasswordHash: testApiKey,
	}
	if err := _connectorGormDB.Create(&newRelationAppKey).Error; err != nil {
		t.Error(err)
		return
	}
	newAPIKey := APIKey{
		App:        testApp,
		KeyID:      testKeyID,
		Name:       testApp,
		Permission: util.PermRead,
	}
	if err := _connectorGormDB.Create(&newAPIKey).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		// clear test data
		if err := _connectorGormDB.Delete(&RelationAppKey{ID: newRelationAppKey.ID}).Error; err != nil {
			t.Error(err)
			return
		}
		if err := _connectorGormDB.Delete(&APIKey{ID: newAPIKey.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()

	reqData := struct {
		ApiKey string `json:"apiKey"`
		App    string `json:"app"`
		Role   string `json:"role"`
	}{
		ApiKey: testApiKey,
		App:    testApp,
		Role:   util.PermWrite,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Valid bool `json:"valid"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/riki/console/validateApiKey/v1", &reqData, nil, &respData, nil)
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
	if respData.Data.Valid {
		t.Error("not prospective response data")
		return
	}
}

func TestValidateApiKey_App1AccessApp2(t *testing.T) {
	var (
		testApp1   = "app_TestValidateApiKey_App1AccessApp2_App1"
		testApp2   = "app_TestValidateApiKey_App1AccessApp2_App2"
		testKeyID  = "pswd_TestValidateApiKey_App1AccessApp2_App1"
		testApiKey = "testApiKey_TestValidateApiKey_App1AccessApp2"
	)
	newRelationAppKey := RelationAppKey{
		App:          testApp1,
		KeyID:        testKeyID,
		PasswordHash: testApiKey,
	}
	if err := _connectorGormDB.Create(&newRelationAppKey).Error; err != nil {
		t.Error(err)
		return
	}
	newAPIKey := APIKey{
		App:        testApp1,
		KeyID:      testKeyID,
		Name:       testApp1,
		Permission: util.PermWrite,
	}
	if err := _connectorGormDB.Create(&newAPIKey).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		// clear test data
		if err := _connectorGormDB.Delete(&RelationAppKey{ID: newRelationAppKey.ID}).Error; err != nil {
			t.Error(err)
			return
		}
		if err := _connectorGormDB.Delete(&APIKey{ID: newAPIKey.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()

	reqData := struct {
		ApiKey string `json:"apiKey"`
		App    string `json:"app"`
		Role   string `json:"role"`
	}{
		ApiKey: testApiKey,
		App:    testApp2,
		Role:   util.PermWrite,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Valid bool `json:"valid"`
		} `json:"data"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/riki/console/validateApiKey/v1", &reqData, nil, &respData, nil)
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
	if respData.Data.Valid {
		t.Error("not prospective response data")
		return
	}
}
