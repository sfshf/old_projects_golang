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

func TestGetSecurityQuestions(t *testing.T) {
	var (
		ctx                 = context.Background()
		err                 error
		testPlainText       = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("TestGetSecurityQuestions-plaintext_%s", Random(PasswordLength))))
		testCipherTextBytes []byte
		testTitle           = fmt.Sprintf("TestGetSecurityQuestions-title_%s", Random(PasswordLength))
		testDescription     = fmt.Sprintf("TestGetSecurityQuestions-description_%s", Random(PasswordLength))
	)
	testCipherTextBytes, err = ecies.Encrypt(_publicKey, []byte(testPlainText))
	if err != nil {
		t.Error(err)
		return
	}
	rpcCtx := rpc.NewContext(ctx, _localizerManager)
	dataID := fmt.Sprintf("%d-sq-%s", _testAccount.ID, testTitle)
	// mock data
	if err := connector_grpc.SaveData(ctx, rpcCtx, _connectorApiKey, _connectorKeyID, dataID, testPlainText, base64.StdEncoding.EncodeToString(testCipherTextBytes), true); err != nil {
		t.Error(err)
		return
	}
	defer func() {
		// delete connector mock data
		if err := connector_grpc.DeleteData(ctx, rpcCtx, _connectorApiKey, _connectorKeyID, dataID); err != nil {
			t.Error(err)
			return
		}
	}()
	ts := time.Now().UnixMilli()
	newSecurityQuestion := &SecurityQuestion{
		CreatedAt:   ts,
		UpdatedAt:   ts,
		UserID:      _testAccount.ID,
		DataID:      dataID,
		Title:       testTitle,
		Description: testDescription,
	}
	coll := _mongoDB.Collection(CollectionName_SecurityQuestion)
	result, err := coll.InsertOne(ctx, newSecurityQuestion)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		// delete doom mock data
		if _, err := coll.DeleteOne(ctx, bson.D{{Key: "_id", Value: result.InsertedID}}); err != nil {
			t.Error(err)
			return
		}
	}()

	// TestGetSecurityQuestions ------------------------------------
	reqDataGetSecurityQuestions := struct {
		Title string `json:"title"`
	}{
		Title: testTitle,
	}
	respDataGetSecurityQuestions := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			PlainText string `json:"plainText"`
		} `json:"data"`
	}{}
	// send request
	respGetSecurityQuestions, err := postJsonRequest(_kongDNS+"/doom/getSecurityQuestions/v1", &reqDataGetSecurityQuestions, _testCookie, &respDataGetSecurityQuestions, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respGetSecurityQuestions.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataGetSecurityQuestions.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	if respDataGetSecurityQuestions.Data.PlainText != testPlainText {
		t.Error("not prospective response data")
		return
	}
}

func TestGetSecurityQuestions_EmptySession(t *testing.T) {
	var (
		err       error
		testTitle = fmt.Sprintf("TestGetSecurityQuestions_EmptySession-title_%s", Random(PasswordLength))
	)

	reqDataGetSecurityQuestions := struct {
		Title string `json:"title"`
	}{
		Title: testTitle,
	}
	respDataGetSecurityQuestions := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			PlainText string `json:"plainText"`
		} `json:"data"`
	}{}
	// send request
	respGetSecurityQuestions, err := postJsonRequest(_kongDNS+"/doom/getSecurityQuestions/v1", &reqDataGetSecurityQuestions, nil, &respDataGetSecurityQuestions, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respGetSecurityQuestions.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataGetSecurityQuestions.Code != slark_response.StatusCodeEmptySession {
		t.Error("not prospective response data code")
		return
	}
}

func TestGetSecurityQuestions_EmptyParameter(t *testing.T) {
	var (
		err       error
		testTitle = ""
	)

	reqDataGetSecurityQuestions := struct {
		Title string `json:"title"`
	}{
		Title: testTitle,
	}
	respDataGetSecurityQuestions := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			PlainText string `json:"plainText"`
		} `json:"data"`
	}{}
	// send request
	respGetSecurityQuestions, err := postJsonRequest(_kongDNS+"/doom/getSecurityQuestions/v1", &reqDataGetSecurityQuestions, _testCookie, &respDataGetSecurityQuestions, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respGetSecurityQuestions.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataGetSecurityQuestions.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
}
