package console_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/nextsurfer/connector/api/response"
	. "github.com/nextsurfer/connector/internal/pkg/model"
	"github.com/nextsurfer/connector/internal/pkg/util"
)

func TestConnectorConsolePassword(t *testing.T) {
	var (
		test2App  = "test2"
		testKeyID = fmt.Sprintf("pswd_TestConnectorConsolePassword_%s", util.Random(util.PasswordLength))
	)

	// TestAddPassword ------------------------------------
	reqDataAddPassword := struct {
		App    string `json:"app"`
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
	}{
		App:    test2App,
		ApiKey: _adminApiKey,
		KeyID:  testKeyID,
	}
	respDataAddPassword := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			KeyID        string `json:"keyID"`
			PasswordHash string `json:"passwordHash"`
			CreatedAt    int64  `json:"createdAt"`
		} `json:"data"`
	}{}
	respAddPassword, err := postJsonRequest(_kongDNS+"/riki/console/addPassword/v1", &reqDataAddPassword, nil, &respDataAddPassword, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respAddPassword.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataAddPassword.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	var relationAppKey RelationAppKey
	if err := _connectorGormDB.Table(TableNameRelationAppKey).
		Where("app = ? AND key_id = ? AND deleted_at = 0", test2App, testKeyID).
		First(&relationAppKey).Error; err != nil {
		t.Error(err)
		return
	}

	defer func() {
		// TestRemovePassword ------------------------------------
		reqDataRemovePassword := struct {
			App    string `json:"app"`
			ApiKey string `json:"apiKey"`
			KeyID  string `json:"keyID"`
		}{
			App:    test2App,
			ApiKey: _adminApiKey,
			KeyID:  testKeyID,
		}
		respDataRemovePassword := struct {
			Code         int32  `json:"code"`
			Message      string `json:"message"`
			DebugMessage string `json:"debugMessage"`
		}{}
		// send request
		respRemovePassword, err := postJsonRequest(_kongDNS+"/riki/console/removePassword/v1", &reqDataRemovePassword, nil, &respDataRemovePassword, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if respRemovePassword.StatusCode != http.StatusOK {
			t.Error("not prospective response code")
			return
		}
		if respDataRemovePassword.Code != response.StatusCodeOK {
			t.Error("not prospective response data code")
			return
		}
		if err := _connectorGormDB.Where("app = ? AND key_id = ? AND deleted_at > 0", test2App, testKeyID).
			Delete(&RelationAppKey{}).Error; err != nil {
			t.Error(err)
			return
		}
	}()

	// TestFetchPassword ------------------------------------
	reqDataFetchPassword := struct {
		App    string `json:"app"`
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
	}{
		App:    test2App,
		ApiKey: _adminApiKey,
		KeyID:  testKeyID,
	}
	respDataFetchPassword := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Password string `json:"password"`
		} `json:"data"`
	}{}
	// send request
	respFetchPassword, err := postJsonRequest(_kongDNS+"/riki/console/fetchPassword/v1", &reqDataFetchPassword, nil, &respDataFetchPassword, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respFetchPassword.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataFetchPassword.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	passwordHash, err := util.Keccak256Hex([]byte(respDataFetchPassword.Data.Password))
	if err != nil {
		t.Error(err)
		return
	}
	if string(passwordHash) != respDataAddPassword.Data.PasswordHash {
		t.Error("not prospective response data")
		return
	}

	// TestListPassword ------------------------------------
	reqDataListPassword := struct {
		App    string `json:"app"`
		ApiKey string `json:"apiKey"`
	}{
		App:    test2App,
		ApiKey: _adminApiKey,
	}
	respDataListPassword := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []struct {
				KeyID        string `json:"keyID"`
				PasswordHash string `json:"passwordHash"`
				CreatedAt    int64  `json:"createdAt"`
			} `json:"list"`
		} `json:"data"`
	}{}
	// send request
	respListPassword, err := postJsonRequest(_kongDNS+"/riki/console/listPassword/v1", &reqDataListPassword, nil, &respDataListPassword, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respListPassword.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataListPassword.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	if respDataListPassword.Data.List[0].KeyID != testKeyID {
		t.Error("not prospective response data")
		return
	}
}

func TestConnectorConsolePassword_EmptyApiKey(t *testing.T) {
	var (
		test2App  = "test2"
		testKeyID = fmt.Sprintf("pswd_TestConnectorConsolePassword_EmptyApiKey_%s", util.Random(util.PasswordLength))
	)
	// TestAddPassword ------------------------------------
	reqDataAddPassword := struct {
		App    string `json:"app"`
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
	}{
		App:   test2App,
		KeyID: testKeyID,
	}
	respDataAddPassword := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	respAddPassword, err := postJsonRequest(_kongDNS+"/riki/console/addPassword/v1", &reqDataAddPassword, nil, &respDataAddPassword, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respAddPassword.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataAddPassword.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}

	defer func() {
		// TestRemovePassword ------------------------------------
		reqDataRemovePassword := struct {
			App    string `json:"app"`
			ApiKey string `json:"apiKey"`
			KeyID  string `json:"keyID"`
		}{}
		respDataRemovePassword := struct {
			Code         int32  `json:"code"`
			Message      string `json:"message"`
			DebugMessage string `json:"debugMessage"`
		}{}
		// send request
		respRemovePassword, err := postJsonRequest(_kongDNS+"/riki/console/removePassword/v1", &reqDataRemovePassword, nil, &respDataRemovePassword, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if respRemovePassword.StatusCode != http.StatusOK {
			t.Error("not prospective response code")
			return
		}
		if respDataRemovePassword.Code != response.StatusCodeWrongParameters {
			t.Error("not prospective response data code")
			return
		}
	}()

	// TestFetchPassword ------------------------------------
	reqDataFetchPassword := struct {
		App    string `json:"app"`
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
	}{}
	respDataFetchPassword := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	respFetchPassword, err := postJsonRequest(_kongDNS+"/riki/console/fetchPassword/v1", &reqDataFetchPassword, nil, &respDataFetchPassword, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respFetchPassword.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataFetchPassword.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}

	// TestListPassword ------------------------------------
	reqDataListPassword := struct {
		App    string `json:"app"`
		ApiKey string `json:"apiKey"`
	}{}
	respDataListPassword := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	respListPassword, err := postJsonRequest(_kongDNS+"/riki/console/listPassword/v1", &reqDataListPassword, nil, &respDataListPassword, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respListPassword.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataListPassword.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
}

func TestConnectorConsolePrivateKey(t *testing.T) {
	var (
		test2App  = "test2"
		testKeyID = fmt.Sprintf("TestConnectorConsolePrivateKey_%s", util.Random(util.PasswordLength))
	)

	// TestAddPrivateKey ------------------------------------
	reqDataAddPrivateKey := struct {
		App    string `json:"app"`
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
	}{
		App:    test2App,
		ApiKey: _adminApiKey,
		KeyID:  testKeyID,
	}
	respDataAddPrivateKey := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			KeyID     string `json:"keyID"`
			PublicKey string `json:"publicKey"`
			CreatedAt int64  `json:"createdAt"`
		} `json:"data"`
	}{}
	respAddPrivateKey, err := postJsonRequest(_kongDNS+"/riki/console/addPrivateKey/v1", &reqDataAddPrivateKey, nil, &respDataAddPrivateKey, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respAddPrivateKey.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataAddPrivateKey.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	var relationAppKey RelationAppKey
	if err := _connectorGormDB.Table(TableNameRelationAppKey).
		Where("app = ? AND key_id = ? AND deleted_at = 0", test2App, testKeyID).
		First(&relationAppKey).Error; err != nil {
		t.Error(err)
		return
	}

	defer func() {
		// TestRemovePrivateKey ------------------------------------
		reqDataRemovePrivateKey := struct {
			App    string `json:"app"`
			ApiKey string `json:"apiKey"`
			KeyID  string `json:"keyID"`
		}{
			App:    test2App,
			ApiKey: _adminApiKey,
			KeyID:  testKeyID,
		}
		respDataRemovePrivateKey := struct {
			Code         int32  `json:"code"`
			Message      string `json:"message"`
			DebugMessage string `json:"debugMessage"`
		}{}
		// send request
		respRemovePrivateKey, err := postJsonRequest(_kongDNS+"/riki/console/removePrivateKey/v1", &reqDataRemovePrivateKey, nil, &respDataRemovePrivateKey, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if respRemovePrivateKey.StatusCode != http.StatusOK {
			t.Error("not prospective response code")
			return
		}
		if respDataRemovePrivateKey.Code != response.StatusCodeOK {
			t.Error("not prospective response data code")
			return
		}
		if err := _connectorGormDB.Where("app = ? AND key_id = ? AND deleted_at > 0", test2App, testKeyID).
			Delete(&RelationAppKey{}).Error; err != nil {
			t.Error(err)
			return
		}
	}()

	// TestListPrivateKey ------------------------------------
	reqDataListPrivateKey := struct {
		App          string `json:"app"`
		ApiKey       string `json:"apiKey"`
		KeyID        string `json:"keyID"`
		PasswordHash string `json:"passwordHash"`
	}{
		App:    test2App,
		ApiKey: _adminApiKey,
		KeyID:  testKeyID,
	}
	respDataListPrivateKey := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []struct {
				KeyID     string `json:"keyID"`
				PublicKey string `json:"publicKey"`
				CreatedAt int64  `json:"createdAt"`
			} `json:"list"`
		} `json:"data"`
	}{}
	// send request
	respListPrivateKey, err := postJsonRequest(_kongDNS+"/riki/console/listPrivateKey/v1", &reqDataListPrivateKey, nil, &respDataListPrivateKey, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respListPrivateKey.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataListPrivateKey.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	if respDataListPrivateKey.Data.List[0].KeyID != testKeyID {
		t.Error("not prospective response data")
		return
	}
}

func TestConnectorConsolePrivateKey_EmptyApiKey(t *testing.T) {
	var (
		test2App  = "test2"
		testKeyID = fmt.Sprintf("TestConnectorConsolePrivateKey_EmptyApiKey_%s", util.Random(util.PasswordLength))
	)
	// TestAddPrivateKey ------------------------------------
	reqDataAddPrivateKey := struct {
		App    string `json:"app"`
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
	}{
		App:   test2App,
		KeyID: testKeyID,
	}
	respDataAddPrivateKey := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	respAddPrivateKey, err := postJsonRequest(_kongDNS+"/riki/console/addPrivateKey/v1", &reqDataAddPrivateKey, nil, &respDataAddPrivateKey, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respAddPrivateKey.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataAddPrivateKey.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}

	defer func() {
		// TestRemovePrivateKey ------------------------------------
		reqDataRemovePrivateKey := struct {
			App    string `json:"app"`
			ApiKey string `json:"apiKey"`
			KeyID  string `json:"keyID"`
		}{}
		respDataRemovePrivateKey := struct {
			Code         int32  `json:"code"`
			Message      string `json:"message"`
			DebugMessage string `json:"debugMessage"`
		}{}
		// send request
		respRemovePrivateKey, err := postJsonRequest(_kongDNS+"/riki/console/removePrivateKey/v1", &reqDataRemovePrivateKey, nil, &respDataRemovePrivateKey, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if respRemovePrivateKey.StatusCode != http.StatusOK {
			t.Error("not prospective response code")
			return
		}
		if respDataRemovePrivateKey.Code != response.StatusCodeWrongParameters {
			t.Error("not prospective response data code")
			return
		}
	}()

	// TestListPrivateKey ------------------------------------
	reqDataListPrivateKey := struct {
		App          string `json:"app"`
		ApiKey       string `json:"apiKey"`
		KeyID        string `json:"keyID"`
		PasswordHash string `json:"passwordHash"`
	}{}
	respDataListPrivateKey := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	respListPrivateKey, err := postJsonRequest(_kongDNS+"/riki/console/listPrivateKey/v1", &reqDataListPrivateKey, nil, &respDataListPrivateKey, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respListPrivateKey.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataListPrivateKey.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
}

func TestConnectorConsoleApiKey(t *testing.T) {
	var (
		testApp   = "app_TestConnectorConsoleApiKey"
		testKeyID = fmt.Sprintf("pswd_TestConnectorConsoleApiKey_%s", util.Random(util.PasswordLength))
	)
	// mock data
	newRelationAppKey := RelationAppKey{
		App:   testApp,
		KeyID: testKeyID,
	}
	if err := _connectorGormDB.Create(&newRelationAppKey).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		// clear test data
		if err := _connectorGormDB.Delete(&RelationAppKey{ID: newRelationAppKey.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()

	// TestAddApiKey ------------------------------------
	reqDataAddApiKey := struct {
		ApiKey     string `json:"apiKey"`
		App        string `json:"app"`
		KeyID      string `json:"keyID"`
		Name       string `json:"name"`
		Permission string `json:"permission"`
	}{
		App:        testApp,
		ApiKey:     _adminApiKey,
		KeyID:      testKeyID,
		Name:       testApp,
		Permission: util.PermWrite,
	}
	respDataAddApiKey := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	respAddApiKey, err := postJsonRequest(_kongDNS+"/riki/console/addApiKey/v1", &reqDataAddApiKey, nil, &respDataAddApiKey, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respAddApiKey.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataAddApiKey.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	var aPIKey APIKey
	if err := _connectorGormDB.Table(TableNameAPIKey).
		Where("app = ? AND key_id = ? AND deleted_at = 0", testApp, testKeyID).
		First(&aPIKey).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _connectorGormDB.Delete(&APIKey{ID: aPIKey.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()

	defer func() {
		// TestRemoveApiKey ------------------------------------
		reqDataRemoveApiKey := struct {
			ApiKey string `json:"apiKey"`
			ID     int64  `json:"id"`
		}{
			ApiKey: _adminApiKey,
			ID:     aPIKey.ID,
		}
		respDataRemoveApiKey := struct {
			Code         int32  `json:"code"`
			Message      string `json:"message"`
			DebugMessage string `json:"debugMessage"`
		}{}
		// send request
		respRemoveApiKey, err := postJsonRequest(_kongDNS+"/riki/console/removeApiKey/v1", &reqDataRemoveApiKey, nil, &respDataRemoveApiKey, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if respRemoveApiKey.StatusCode != http.StatusOK {
			t.Error("not prospective response code")
			return
		}
		if respDataRemoveApiKey.Code != response.StatusCodeOK {
			t.Error("not prospective response data code")
			return
		}
		var apiKey APIKey
		if err := _connectorGormDB.Where("app = ? AND key_id = ? AND deleted_at > 0", testApp, testKeyID).
			First(&apiKey).Error; err != nil {
			t.Error(err)
			return
		}
	}()

	// TestListApiKey ------------------------------------
	reqDataListApiKey := struct {
		ApiKey string `json:"apiKey"`
	}{
		ApiKey: _adminApiKey,
	}
	respDataListApiKey := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []struct {
				ID         int64  `json:"id"`
				CreatedAt  int64  `json:"createdAt"`
				KeyID      string `json:"keyID"`
				Permission string `json:"permission"`
				App        string `json:"app"`
				Name       string `json:"name"`
			} `json:"list"`
		} `json:"data"`
	}{}
	// send request
	respListApiKey, err := postJsonRequest(_kongDNS+"/riki/console/listApiKey/v1", &reqDataListApiKey, nil, &respDataListApiKey, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respListApiKey.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataListApiKey.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	if respDataListApiKey.Data.List[0].KeyID != testKeyID {
		t.Error("not prospective response data")
		return
	}
}

func TestConnectorConsoleApiKey_EmptyApiKey(t *testing.T) {
	var (
		test2App  = "test2"
		testKeyID = fmt.Sprintf("pswd_TestConnectorConsoleApiKey_%s", util.Random(util.PasswordLength))
	)

	// TestAddApiKey ------------------------------------
	reqDataAddApiKey := struct {
		ApiKey     string `json:"apiKey"`
		App        string `json:"app"`
		KeyID      string `json:"keyID"`
		Name       string `json:"name"`
		Permission string `json:"permission"`
	}{
		App:        test2App,
		KeyID:      testKeyID,
		Name:       test2App,
		Permission: util.PermWrite,
	}
	respDataAddApiKey := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	respAddApiKey, err := postJsonRequest(_kongDNS+"/riki/console/addApiKey/v1", &reqDataAddApiKey, nil, &respDataAddApiKey, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respAddApiKey.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataAddApiKey.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}

	defer func() {
		// TestRemoveApiKey ------------------------------------
		reqDataRemoveApiKey := struct {
			ApiKey string `json:"apiKey"`
			ID     int64  `json:"id"`
		}{}
		respDataRemoveApiKey := struct {
			Code         int32  `json:"code"`
			Message      string `json:"message"`
			DebugMessage string `json:"debugMessage"`
		}{}
		// send request
		respRemoveApiKey, err := postJsonRequest(_kongDNS+"/riki/console/removeApiKey/v1", &reqDataRemoveApiKey, nil, &respDataRemoveApiKey, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if respRemoveApiKey.StatusCode != http.StatusOK {
			t.Error("not prospective response code")
			return
		}
		if respDataRemoveApiKey.Code != response.StatusCodeWrongParameters {
			t.Error("not prospective response data code")
			return
		}
	}()

	// TestListApiKey ------------------------------------
	reqDataListApiKey := struct {
		ApiKey string `json:"apiKey"`
	}{}
	respDataListApiKey := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []struct {
				ID         int64  `json:"id"`
				CreatedAt  int64  `json:"createdAt"`
				KeyID      string `json:"keyID"`
				Permission string `json:"permission"`
				App        string `json:"app"`
				Name       string `json:"name"`
			} `json:"list"`
		} `json:"data"`
	}{}
	// send request
	respListApiKey, err := postJsonRequest(_kongDNS+"/riki/console/listApiKey/v1", &reqDataListApiKey, nil, &respDataListApiKey, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respListApiKey.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataListApiKey.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
}
