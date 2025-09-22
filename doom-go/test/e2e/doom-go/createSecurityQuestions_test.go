package doom_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"

	ecies "github.com/ecies/go/v2"
	connector_grpc "github.com/nextsurfer/connector/pkg/grpc"
	"github.com/nextsurfer/doom-go/api/response"
	. "github.com/nextsurfer/doom-go/internal/model"
	"github.com/nextsurfer/ground/pkg/rpc"
	slark_response "github.com/nextsurfer/slark/api/response"
	"go.mongodb.org/mongo-driver/bson"
)

func TestCreateSecurityQuestions(t *testing.T) {
	var (
		ctx                 = context.Background()
		err                 error
		testPlainText       = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("TestCreateSecurityQuestions-plaintext_%s", Random(PasswordLength))))
		testCipherTextBytes []byte
		testTitle           = fmt.Sprintf("TestCreateSecurityQuestions-title_%s", Random(PasswordLength))
		testDescription     = fmt.Sprintf("TestCreateSecurityQuestions-description_%s", Random(PasswordLength))
	)
	testCipherTextBytes, err = ecies.Encrypt(_publicKey, []byte(testPlainText))
	if err != nil {
		t.Error(err)
		return
	}
	rpcCtx := rpc.NewContext(ctx, _localizerManager)
	dataID := fmt.Sprintf("%d-sq-%s", _testAccount.ID, testTitle)

	// TestCreateSecurityQuestions ------------------------------------
	reqDataCreateSecurityQuestions := struct {
		PlainText   string `json:"plainText"`
		CipherText  string `json:"cipherText"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}{
		PlainText:   testPlainText,
		CipherText:  base64.StdEncoding.EncodeToString(testCipherTextBytes),
		Title:       testTitle,
		Description: testDescription,
	}
	respDataCreateSecurityQuestions := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	respCreateSecurityQuestions, err := postJsonRequest(_kongDNS+"/doom/createSecurityQuestions/v1", &reqDataCreateSecurityQuestions, _testCookie, &respDataCreateSecurityQuestions, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respCreateSecurityQuestions.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataCreateSecurityQuestions.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	defer func() {
		// delete doom data
		if _, err := _mongoDB.Collection(CollectionName_SecurityQuestion).
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
		// delete connector data ------------------------------------
		if err := connector_grpc.DeleteData(ctx, rpcCtx, _connectorApiKey, _connectorKeyID, dataID); err != nil {
			t.Error(err)
			return
		}
	}()
}

func TestCreateSecurityQuestions_EmptySession(t *testing.T) {
	var (
		err                 error
		testPlainText       = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("TestCreateSecurityQuestions_EmptySession-plaintext_%s", Random(PasswordLength))))
		testCipherTextBytes []byte
		testTitle           = fmt.Sprintf("TestCreateSecurityQuestions_EmptySession-title_%s", Random(PasswordLength))
		testDescription     = fmt.Sprintf("TestCreateSecurityQuestions_EmptySession-description_%s", Random(PasswordLength))
	)
	testCipherTextBytes, err = ecies.Encrypt(_publicKey, []byte(testPlainText))
	if err != nil {
		t.Error(err)
		return
	}

	reqDataCreateSecurityQuestions := struct {
		PlainText   string `json:"plainText"`
		CipherText  string `json:"cipherText"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}{
		PlainText:   testPlainText,
		CipherText:  base64.StdEncoding.EncodeToString(testCipherTextBytes),
		Title:       testTitle,
		Description: testDescription,
	}
	respDataCreateSecurityQuestions := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	respCreateSecurityQuestions, err := postJsonRequest(_kongDNS+"/doom/createSecurityQuestions/v1", &reqDataCreateSecurityQuestions, nil, &respDataCreateSecurityQuestions, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respCreateSecurityQuestions.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataCreateSecurityQuestions.Code != slark_response.StatusCodeEmptySession {
		t.Error("not prospective response data code")
		return
	}
}

func TestCreateSecurityQuestions_EmptyParameter(t *testing.T) {
	var (
		err                 error
		testPlainText       = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("TestCreateSecurityQuestions-plaintext_%s", Random(PasswordLength))))
		testCipherTextBytes []byte
		testTitle           = ""
		testDescription     = fmt.Sprintf("TestCreateSecurityQuestions-description_%s", Random(PasswordLength))
	)
	testCipherTextBytes, err = ecies.Encrypt(_publicKey, []byte(testPlainText))
	if err != nil {
		t.Error(err)
		return
	}

	reqDataCreateSecurityQuestions := struct {
		PlainText   string `json:"plainText"`
		CipherText  string `json:"cipherText"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}{
		PlainText:   testPlainText,
		CipherText:  base64.StdEncoding.EncodeToString(testCipherTextBytes),
		Title:       testTitle,
		Description: testDescription,
	}
	respDataCreateSecurityQuestions := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	respCreateSecurityQuestions, err := postJsonRequest(_kongDNS+"/doom/createSecurityQuestions/v1", &reqDataCreateSecurityQuestions, _testCookie, &respDataCreateSecurityQuestions, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respCreateSecurityQuestions.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataCreateSecurityQuestions.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
}
