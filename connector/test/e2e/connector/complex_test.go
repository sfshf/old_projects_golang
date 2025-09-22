package connector_test

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"testing"

	ecies "github.com/ecies/go/v2"
	"github.com/nextsurfer/connector/api/response"
	. "github.com/nextsurfer/connector/internal/pkg/model"
	"github.com/nextsurfer/connector/internal/pkg/util"
)

func TestConnectorPassword(t *testing.T) {
	var (
		testKeyID        = fmt.Sprintf("pswd_TestConnectorPassword_%s", util.Random(util.PasswordLength))
		testPassword     string
		testPasswordHash []byte
	)
	// TestCreatePassword ------------------------------------
	reqDataCreatePassword := struct {
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
	}{
		ApiKey: _adminApiKey,
		KeyID:  testKeyID,
	}
	respDataCreatePassword := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Password string `json:"password"`
		} `json:"data"`
	}{}
	respCreatePassword, err := postJsonRequest(_kongDNS+"/riki/createPassword/v1", &reqDataCreatePassword, nil, &respDataCreatePassword, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respCreatePassword.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataCreatePassword.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	testPassword = respDataCreatePassword.Data.Password
	testPasswordHash, _ = util.Keccak256Hex([]byte(testPassword))
	var relationAppKey RelationAppKey
	if err := _connectorGormDB.Where("app = ? AND key_id = ? AND deleted_at = 0", "admin", testKeyID).
		First(&relationAppKey).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		// TestDeletePassword ------------------------------------
		reqDataDeletePassword := struct {
			ApiKey string `json:"apiKey"`
			KeyID  string `json:"keyID"`
		}{
			ApiKey: _adminApiKey,
			KeyID:  testKeyID,
		}
		respDataDeletePassword := struct {
			Code         int32  `json:"code"`
			Message      string `json:"message"`
			DebugMessage string `json:"debugMessage"`
		}{}
		// send request
		respDeletePassword, err := postJsonRequest(_kongDNS+"/riki/deletePassword/v1", &reqDataDeletePassword, nil, &respDataDeletePassword, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if respDeletePassword.StatusCode != http.StatusOK {
			t.Error("not prospective response code")
			return
		}
		if respDataDeletePassword.Code != response.StatusCodeOK {
			t.Error("not prospective response data code")
			return
		}
		if err := _connectorGormDB.Delete(&RelationAppKey{},
			"app=? AND key_id=? AND deleted_at>0", "admin", testKeyID).Error; err != nil {
			t.Error(err)
			return
		}
	}()

	// TestCheckPassword ------------------------------------
	reqDataCheckPassword := struct {
		ApiKey       string `json:"apiKey"`
		KeyID        string `json:"keyID"`
		PasswordHash string `json:"passwordHash"`
	}{
		ApiKey:       _adminApiKey,
		KeyID:        testKeyID,
		PasswordHash: string(testPasswordHash),
	}
	respDataCheckPassword := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Valid bool `json:"valid"`
		} `json:"data"`
	}{}
	// send request
	respCheckPassword, err := postJsonRequest(_kongDNS+"/riki/checkPassword/v1", &reqDataCheckPassword, nil, &respDataCheckPassword, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respCheckPassword.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataCheckPassword.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	if !respDataCheckPassword.Data.Valid {
		t.Error("not prospective response data")
		return
	}
}

func TestConnectorPassword_EmptyApiKey(t *testing.T) {
	var (
		testKeyID        = fmt.Sprintf("pswd_TestConnectorPasswordEmptyApiKey_%s", util.Random(util.PasswordLength))
		testPasswordHash []byte
	)

	// TestCreatePassword ------------------------------------
	reqDataCreatePassword := struct {
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
	}{
		KeyID: testKeyID,
	}
	respDataCreatePassword := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Password string `json:"password"`
		} `json:"data"`
	}{}
	respCreatePassword, err := postJsonRequest(_kongDNS+"/riki/createPassword/v1", &reqDataCreatePassword, nil, &respDataCreatePassword, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respCreatePassword.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataCreatePassword.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
	defer func() {
		// TestDeletePassword ------------------------------------
		reqDataDeletePassword := struct {
			ApiKey string `json:"apiKey"`
			KeyID  string `json:"keyID"`
		}{
			KeyID: testKeyID,
		}
		respDataDeletePassword := struct {
			Code         int32  `json:"code"`
			Message      string `json:"message"`
			DebugMessage string `json:"debugMessage"`
		}{}
		// send request
		respDeletePassword, err := postJsonRequest(_kongDNS+"/riki/deletePassword/v1", &reqDataDeletePassword, nil, &respDataDeletePassword, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if respDeletePassword.StatusCode != http.StatusOK {
			t.Error("not prospective response code")
			return
		}
		if respDataDeletePassword.Code != response.StatusCodeWrongParameters {
			t.Error("not prospective response data code")
			return
		}
	}()
	// TestCheckPasswordEmptyApiKey ------------------------------------
	reqDataCheckPassword := struct {
		ApiKey       string `json:"apiKey"`
		KeyID        string `json:"keyID"`
		PasswordHash string `json:"passwordHash"`
	}{
		KeyID:        testKeyID,
		PasswordHash: string(testPasswordHash),
	}
	respDataCheckPassword := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Valid bool `json:"valid"`
		} `json:"data"`
	}{}
	// send request
	respCheckPassword, err := postJsonRequest(_kongDNS+"/riki/checkPassword/v1", &reqDataCheckPassword, nil, &respDataCheckPassword, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respCheckPassword.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataCheckPassword.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
}

func TestConnectorPrivateKey(t *testing.T) {
	var (
		testKeyID     = fmt.Sprintf("TestConnectorPrivateKey_%s", util.Random(6))
		testPublicKey string
	)

	// TestCreatePrivateKey ------------------------------------
	reqDataCreatePrivateKey := struct {
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
	}{
		ApiKey: _adminApiKey,
		KeyID:  testKeyID,
	}
	respDataCreatePrivateKey := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			PublicKey string `json:"publicKey"`
		} `json:"data"`
	}{}
	respCreatePrivateKey, err := postJsonRequest(_kongDNS+"/riki/createPrivateKey/v1", &reqDataCreatePrivateKey, nil, &respDataCreatePrivateKey, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respCreatePrivateKey.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataCreatePrivateKey.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	testPublicKey = respDataCreatePrivateKey.Data.PublicKey
	var relationAppKey RelationAppKey
	if err := _connectorGormDB.Table(TableNameRelationAppKey).
		Where("app = ? AND key_id = ? AND deleted_at = 0", "admin", testKeyID).
		First(&relationAppKey).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		// TestDeletePrivateKey ------------------------------------
		reqDataDeletePrivateKey := struct {
			ApiKey string `json:"apiKey"`
			KeyID  string `json:"keyID"`
		}{
			ApiKey: _adminApiKey,
			KeyID:  testKeyID,
		}
		respDataDeletePrivateKey := struct {
			Code         int32  `json:"code"`
			Message      string `json:"message"`
			DebugMessage string `json:"debugMessage"`
		}{}
		// send request
		respDeletePrivateKey, err := postJsonRequest(_kongDNS+"/riki/deletePrivateKey/v1", &reqDataDeletePrivateKey, nil, &respDataDeletePrivateKey, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if respDeletePrivateKey.StatusCode != http.StatusOK {
			t.Error("not prospective response code")
			return
		}
		if respDataDeletePrivateKey.Code != response.StatusCodeOK {
			t.Error("not prospective response data code")
			return
		}
		if err := _connectorGormDB.Where("app = ? AND key_id = ? AND deleted_at > 0", "admin", testKeyID).
			Delete(&RelationAppKey{}).Error; err != nil {
			t.Error(err)
			return
		}
	}()

	// TestCheckKeyExisting ------------------------------------
	reqDataCheckKeyExisting := struct {
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
	}{
		ApiKey: _adminApiKey,
		KeyID:  testKeyID,
	}
	respDataCheckKeyExisting := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Existing bool `json:"existing"`
		} `json:"data"`
	}{}
	// send request
	respCheckKeyExisting, err := postJsonRequest(_kongDNS+"/riki/checkKeyExisting/v1", &reqDataCheckKeyExisting, nil, &respDataCheckKeyExisting, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respCheckKeyExisting.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataCheckKeyExisting.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	if !respDataCheckKeyExisting.Data.Existing {
		t.Error("not prospective response data")
		return
	}
	// TestGetPublicKey ------------------------------------
	reqDataGetPublicKey := struct {
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
	}{
		ApiKey: _adminApiKey,
		KeyID:  testKeyID,
	}
	respDataGetPublicKey := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			PublicKey string `json:"publicKey"`
		} `json:"data"`
	}{}
	// send request
	respGetPublicKey, err := postJsonRequest(_kongDNS+"/riki/getPublicKey/v1", &reqDataGetPublicKey, nil, &respDataGetPublicKey, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respGetPublicKey.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataGetPublicKey.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	if respDataGetPublicKey.Data.PublicKey != testPublicKey {
		t.Error("not prospective response data")
		return
	}
}

func TestConnectorPrivateKey_EmptyApiKey(t *testing.T) {
	var (
		testKeyID = fmt.Sprintf("TestConnectorPrivateKeyEmptyApiKey_%s", util.Random(6))
	)

	// TestCreatePrivateKey ------------------------------------
	reqDataCreatePrivateKey := struct {
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
	}{
		KeyID: testKeyID,
	}
	respDataCreatePrivateKey := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			PublicKey string `json:"publicKey"`
		} `json:"data"`
	}{}
	respCreatePrivateKey, err := postJsonRequest(_kongDNS+"/riki/createPrivateKey/v1", &reqDataCreatePrivateKey, nil, &respDataCreatePrivateKey, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respCreatePrivateKey.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataCreatePrivateKey.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
	defer func() {
		// TestDeletePrivateKey ------------------------------------
		reqDataDeletePrivateKey := struct {
			ApiKey string `json:"apiKey"`
			KeyID  string `json:"keyID"`
		}{
			KeyID: testKeyID,
		}
		respDataDeletePrivateKey := struct {
			Code         int32  `json:"code"`
			Message      string `json:"message"`
			DebugMessage string `json:"debugMessage"`
		}{}
		// send request
		respDeletePrivateKey, err := postJsonRequest(_kongDNS+"/riki/deletePrivateKey/v1", &reqDataDeletePrivateKey, nil, &respDataDeletePrivateKey, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if respDeletePrivateKey.StatusCode != http.StatusOK {
			t.Error("not prospective response code")
			return
		}
		if respDataDeletePrivateKey.Code != response.StatusCodeWrongParameters {
			t.Error("not prospective response data code")
			return
		}
	}()

	// TestCheckKeyExisting ------------------------------------
	reqDataCheckKeyExisting := struct {
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
	}{
		KeyID: testKeyID,
	}
	respDataCheckKeyExisting := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Existing bool `json:"existing"`
		} `json:"data"`
	}{}
	// send request
	respCheckKeyExisting, err := postJsonRequest(_kongDNS+"/riki/checkKeyExisting/v1", &reqDataCheckKeyExisting, nil, &respDataCheckKeyExisting, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respCheckKeyExisting.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataCheckKeyExisting.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
	// TestGetPublicKeyEmptyApiKey ------------------------------------
	reqDataGetPublicKey := struct {
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
	}{
		KeyID: testKeyID,
	}
	respDataGetPublicKey := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			PublicKey string `json:"publicKey"`
		} `json:"data"`
	}{}
	// send request
	respGetPublicKey, err := postJsonRequest(_kongDNS+"/riki/getPublicKey/v1", &reqDataGetPublicKey, nil, &respDataGetPublicKey, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respGetPublicKey.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataGetPublicKey.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
	// TestDeletePrivateKeyEmptyApiKey ------------------------------------
	reqDataDeletePrivateKey := struct {
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
	}{
		KeyID: testKeyID,
	}
	respDataDeletePrivateKey := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	respDeletePrivateKey, err := postJsonRequest(_kongDNS+"/riki/deletePrivateKey/v1", &reqDataDeletePrivateKey, nil, &respDataDeletePrivateKey, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respDeletePrivateKey.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataDeletePrivateKey.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
}

func TestConnectorStoredData(t *testing.T) {
	var (
		testKeyID     = fmt.Sprintf("TestConnectorStoredData_%s", util.Random(6))
		testPublicKey string
		testDataID    = fmt.Sprintf("TestConnectorStoredData_%s", util.Random(6))
	)
	// create a private key
	reqDataCreatePrivateKey := struct {
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
	}{
		ApiKey: _adminApiKey,
		KeyID:  testKeyID,
	}
	respDataCreatePrivateKey := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			PublicKey string `json:"publicKey"`
		} `json:"data"`
	}{}
	respCreatePrivateKey, err := postJsonRequest(_kongDNS+"/riki/createPrivateKey/v1", &reqDataCreatePrivateKey, nil, &respDataCreatePrivateKey, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respCreatePrivateKey.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataCreatePrivateKey.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	testPublicKey = respDataCreatePrivateKey.Data.PublicKey

	// defer delete the private key
	defer func() {
		reqDataDeletePrivateKey := struct {
			ApiKey string `json:"apiKey"`
			KeyID  string `json:"keyID"`
		}{
			ApiKey: _adminApiKey,
			KeyID:  testKeyID,
		}
		respDataDeletePrivateKey := struct {
			Code         int32  `json:"code"`
			Message      string `json:"message"`
			DebugMessage string `json:"debugMessage"`
		}{}
		// send request
		respDeletePrivateKey, err := postJsonRequest(_kongDNS+"/riki/deletePrivateKey/v1", &reqDataDeletePrivateKey, nil, &respDataDeletePrivateKey, nil)
		if err != nil {
			log.Println(err)
			return
		}
		if respDeletePrivateKey.StatusCode != http.StatusOK {
			t.Error("not prospective response code")
			return
		}
		if respDataDeletePrivateKey.Code != response.StatusCodeOK {
			t.Error("not prospective response data code")
			return
		}
		// remove relation record
		if err := _connectorGormDB.Where("app = ? AND key_id = ? AND deleted_at > 0", "admin", testKeyID).
			Delete(&RelationAppKey{}).Error; err != nil {
			t.Error(err)
			return
		}
	}()

	// mock test data, and encrypt the data
	testData := "TestConnectorStoredData"
	pubKeyBytes, err := hex.DecodeString(testPublicKey)
	if err != nil {
		t.Errorf("base64 decode public key fail: %v", err)
		return
	}
	pubKey, err := ecies.NewPublicKeyFromBytes(pubKeyBytes)
	if err != nil {
		t.Errorf("parse public key fail: %v", err)
		return
	}
	encryptedTestData, err := ecies.Encrypt(pubKey, []byte(testData))
	if err != nil {
		t.Errorf("encrypt test data fail: %v", err)
		return
	}
	plaintextHash, err := util.Keccak256Hex([]byte(testData))
	if err != nil {
		t.Errorf("hash test data fail: %v", err)
		return
	}
	base64EncryptedTestData := base64.StdEncoding.EncodeToString(encryptedTestData)

	// TestSaveData ------------------------------------
	reqDataSaveData := struct {
		ApiKey             string `json:"apiKey"`
		KeyID              string `json:"keyID"`
		DataID             string `json:"dataID"`
		ReplaceCurrentItem bool   `json:"replaceCurrentItem"`
		Data               string `json:"data"`
		PlaintextHash      string `json:"plaintextHash"`
	}{
		ApiKey:             _adminApiKey,
		KeyID:              testKeyID,
		DataID:             testDataID,
		ReplaceCurrentItem: false,
		Data:               base64EncryptedTestData,
		PlaintextHash:      string(plaintextHash),
	}
	respDataSaveData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	respSaveData, err := postJsonRequest(_kongDNS+"/riki/saveData/v1", &reqDataSaveData, nil, &respDataSaveData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respSaveData.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataSaveData.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	// check relation record
	var relationAppDatum RelationAppDatum
	if err := _connectorGormDB.Table(TableNameRelationAppDatum).
		Where("app = ? AND key_id = ? AND data_id = ? AND deleted_at = 0", "admin", testKeyID, testDataID).
		First(&relationAppDatum).Error; err != nil {
		t.Error(err)
		return
	}

	// TestGetData ------------------------------------
	reqDataGetData := struct {
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
		DataID string `json:"dataID"`
	}{
		ApiKey: _adminApiKey,
		KeyID:  testKeyID,
		DataID: testDataID,
	}
	respDataGetData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Data string `json:"data"`
		} `json:"data"`
	}{}
	respGetData, err := postJsonRequest(_kongDNS+"/riki/getData/v1", &reqDataGetData, nil, &respDataGetData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respGetData.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataGetData.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	plaintext, err := base64.StdEncoding.DecodeString(respDataGetData.Data.Data)
	if err != nil {
		t.Error(err)
		return
	}
	if testData != string(plaintext) {
		t.Error("not prospective response data")
		return
	}

	// TestDeleteData ------------------------------------
	reqDataDeleteData := struct {
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
		DataID string `json:"dataID"`
	}{
		ApiKey: _adminApiKey,
		KeyID:  testKeyID,
		DataID: testDataID,
	}
	respDataDeleteData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	respDeleteData, err := postJsonRequest(_kongDNS+"/riki/deleteData/v1", &reqDataDeleteData, nil, &respDataDeleteData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respDeleteData.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataDeleteData.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	// remove relation record
	if err := _connectorGormDB.Where("app = ? AND key_id = ? AND data_id = ? AND deleted_at > 0", "admin", testKeyID, testDataID).
		Delete(&RelationAppDatum{}).Error; err != nil {
		t.Error(err)
		return
	}

	// TestDecryptData ------------------------------------
	reqDataDecryptData := struct {
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
		Data   string `json:"data"`
	}{
		ApiKey: _adminApiKey,
		KeyID:  testKeyID,
		Data:   base64EncryptedTestData,
	}
	respDataDecryptData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			DecrypedData string `json:"decrypedData"`
		} `json:"data"`
	}{}
	respDecryptData, err := postJsonRequest(_kongDNS+"/riki/decryptData/v1", &reqDataDecryptData, nil, &respDataDecryptData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respDecryptData.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataDecryptData.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	decryptedData, err := base64.StdEncoding.DecodeString(respDataDecryptData.Data.DecrypedData)
	if err != nil {
		t.Error(err)
		return
	}
	if string(decryptedData) != testData {
		t.Error("not prospective response data")
		return
	}
}

func TestConnectorStoredData_EmptyApiKey(t *testing.T) {
	var (
		testKeyID  = fmt.Sprintf("TestConnectorStoredDataEmptyApiKey_%s", util.Random(6))
		testDataID = fmt.Sprintf("TestConnectorStoredDataEmptyApiKey_%s", util.Random(6))
	)

	// TestSaveData ------------------------------------
	reqDataSaveData := struct {
		ApiKey             string `json:"apiKey"`
		KeyID              string `json:"keyID"`
		DataID             string `json:"dataID"`
		ReplaceCurrentItem bool   `json:"replaceCurrentItem"`
		Data               string `json:"data"`
		PlaintextHash      string `json:"plaintextHash"`
	}{
		KeyID:              testKeyID,
		DataID:             testDataID,
		ReplaceCurrentItem: false,
	}
	respDataSaveData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	respSaveData, err := postJsonRequest(_kongDNS+"/riki/saveData/v1", &reqDataSaveData, nil, &respDataSaveData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respSaveData.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataSaveData.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}

	// TestGetData ------------------------------------
	reqDataGetData := struct {
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
		DataID string `json:"dataID"`
	}{
		KeyID:  testKeyID,
		DataID: testDataID,
	}
	respDataGetData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Data string `json:"data"`
		} `json:"data"`
	}{}
	respGetData, err := postJsonRequest(_kongDNS+"/riki/getData/v1", &reqDataGetData, nil, &respDataGetData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respGetData.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataGetData.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}

	// TestDeleteData ------------------------------------
	reqDataDeleteData := struct {
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
		DataID string `json:"dataID"`
	}{
		KeyID:  testKeyID,
		DataID: testDataID,
	}
	respDataDeleteData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	respDeleteData, err := postJsonRequest(_kongDNS+"/riki/deleteData/v1", &reqDataDeleteData, nil, &respDataDeleteData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respDeleteData.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataDeleteData.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}

	// TestDecryptData ------------------------------------
	reqDataDecryptData := struct {
		ApiKey string `json:"apiKey"`
		KeyID  string `json:"keyID"`
		Data   string `json:"data"`
	}{
		KeyID: testKeyID,
	}
	respDataDecryptData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			DecrypedData string `json:"decrypedData"`
		} `json:"data"`
	}{}
	respDecryptData, err := postJsonRequest(_kongDNS+"/riki/decryptData/v1", &reqDataDecryptData, nil, &respDataDecryptData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respDecryptData.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataDecryptData.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
}

func TestI18nEnglish(t *testing.T) {
	reqData := struct {
		ApiKey string `json:"apiKey"`
	}{}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/riki/console/getMonitorInfos/v1", &reqData, nil, &respData, nil)
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
	if respData.Message != "Request has wrong parameters, please inspect parameters" {
		t.Error("not prospective response data")
		return
	}
}

func TestI18nChinese(t *testing.T) {
	reqData := struct {
		ApiKey string `json:"apiKey"`
	}{}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/riki/console/getMonitorInfos/v1", &reqData, nil, &respData, func(req *http.Request) {
		req.Header.Set("Accept-Language", "zh")
	})
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
	if respData.Message != "请求参数有误，请您排错后重试" {
		t.Error("not prospective response data")
		return
	}
}
