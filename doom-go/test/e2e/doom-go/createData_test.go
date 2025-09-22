package doom_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"
	"time"

	ecies "github.com/ecies/go/v2"
	connector_grpc "github.com/nextsurfer/connector/pkg/grpc"
	"github.com/nextsurfer/doom-go/api/response"
	. "github.com/nextsurfer/doom-go/internal/model"
	"github.com/nextsurfer/ground/pkg/rpc"
	slark_response "github.com/nextsurfer/slark/api/response"
	"go.mongodb.org/mongo-driver/bson"
)

func TestCreateData(t *testing.T) {
	var (
		ctx                 = context.Background()
		err                 error
		testPlainText       = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("TestCreateData-plaintext_%s", Random(PasswordLength))))
		testCipherTextBytes []byte
		testTitle           = fmt.Sprintf("TestCreateData-title_%s", Random(PasswordLength))
	)
	testCipherTextBytes, err = ecies.Encrypt(_publicKey, []byte(testPlainText))
	if err != nil {
		t.Error(err)
		return
	}
	rpcCtx := rpc.NewContext(ctx, _localizerManager)
	dataID := fmt.Sprintf("%d-data-%s", _testAccount.ID, testTitle)
	// TestCreateData ------------------------------------
	reqDataCreateData := struct {
		PlainText  string `json:"plainText"`
		CipherText string `json:"cipherText"`
		Title      string `json:"title"`
	}{
		PlainText:  testPlainText,
		CipherText: base64.StdEncoding.EncodeToString(testCipherTextBytes),
		Title:      testTitle,
	}
	respDataCreateData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	time.Sleep(1 * time.Second) // bypass oracle rate limit
	respCreateData, err := postJsonRequest(_kongDNS+"/doom/createData/v1", &reqDataCreateData, _testCookie, &respDataCreateData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respCreateData.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataCreateData.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	defer func() {
		// delete doom data
		if _, err := _mongoDB.Collection(CollectionName_Datum).
			DeleteMany(ctx, bson.D{{Key: "userID", Value: _testAccount.ID}, {Key: "dataID", Value: dataID}}); err != nil {
			t.Error(err)
			return
		}
	}()
	// check data
	plaintext, err := connector_grpc.GetData(ctx, rpcCtx, _connectorApiKey, _connectorKeyID, dataID)
	if err != nil {
		t.Error(err)
		return
	}
	if plaintext != string(testPlainText) {
		t.Error("not prospective response data")
		return
	}
	defer func() {
		// delete connector data
		if err := connector_grpc.DeleteData(ctx, rpcCtx, _connectorApiKey, _connectorKeyID, dataID); err != nil {
			t.Error(err)
			return
		}
	}()
}

func TestCreateData_EmptySession(t *testing.T) {
	var (
		err                 error
		testPlainText       = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("TestCreateData_EmptySession-plaintext_%s", Random(PasswordLength))))
		testCipherTextBytes []byte
		testTitle           = fmt.Sprintf("TestCreateData_EmptySession-title_%s", Random(PasswordLength))
	)
	testCipherTextBytes, err = ecies.Encrypt(_publicKey, []byte(testPlainText))
	if err != nil {
		t.Error(err)
		return
	}

	reqDataCreateData := struct {
		PlainText  string `json:"plainText"`
		CipherText string `json:"cipherText"`
		Title      string `json:"title"`
	}{
		PlainText:  testPlainText,
		CipherText: base64.StdEncoding.EncodeToString(testCipherTextBytes),
		Title:      testTitle,
	}
	respDataCreateData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	time.Sleep(1 * time.Second) // bypass oracle rate limit
	respCreateData, err := postJsonRequest(_kongDNS+"/doom/createData/v1", &reqDataCreateData, nil, &respDataCreateData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respCreateData.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataCreateData.Code != slark_response.StatusCodeEmptySession {
		t.Error("not prospective response data code")
		return
	}
}

func TestCreateData_EmptyParameter(t *testing.T) {
	var (
		err                 error
		testPlainText       = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("TestCreateData-plaintext_%s", Random(PasswordLength))))
		testCipherTextBytes []byte
		testTitle           = ""
	)
	testCipherTextBytes, err = ecies.Encrypt(_publicKey, []byte(testPlainText))
	if err != nil {
		t.Error(err)
		return
	}

	// TestCreateData ------------------------------------
	reqDataCreateData := struct {
		PlainText  string `json:"plainText"`
		CipherText string `json:"cipherText"`
		Title      string `json:"title"`
	}{
		PlainText:  testPlainText,
		CipherText: base64.StdEncoding.EncodeToString(testCipherTextBytes),
		Title:      testTitle,
	}
	respDataCreateData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	time.Sleep(1 * time.Second) // bypass oracle rate limit
	respCreateData, err := postJsonRequest(_kongDNS+"/doom/createData/v1", &reqDataCreateData, _testCookie, &respDataCreateData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respCreateData.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataCreateData.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
}
