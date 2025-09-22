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

func TestDeleteSecurityQuestions(t *testing.T) {
	var (
		ctx                 = context.Background()
		err                 error
		testPlainText       = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("TestDeleteSecurityQuestions-plaintext_%s", Random(PasswordLength))))
		testCipherTextBytes []byte
		testTitle           = fmt.Sprintf("TestDeleteSecurityQuestions-title_%s", Random(PasswordLength))
		testDescription     = fmt.Sprintf("TestDeleteSecurityQuestions-description_%s", Random(PasswordLength))
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

	// TestDeleteSecurityQuestions ------------------------------------
	reqDataDeleteSecurityQuestions := struct {
		Title string `json:"title"`
	}{
		Title: testTitle,
	}
	respDataDeleteSecurityQuestions := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	respDeleteSecurityQuestions, err := postJsonRequest(_kongDNS+"/doom/deleteSecurityQuestions/v1", &reqDataDeleteSecurityQuestions, _testCookie, &respDataDeleteSecurityQuestions, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respDeleteSecurityQuestions.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataDeleteSecurityQuestions.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	// check data
	var securityQuestion SecurityQuestion
	if err := coll.FindOne(ctx, bson.D{
		{Key: "userID", Value: _testAccount.ID},
		{Key: "dataID", Value: dataID},
		{Key: "deletedAt", Value: bson.D{{Key: "$gt", Value: 0}}},
	}).Decode(&securityQuestion); err != nil {
		t.Error(err)
		return
	}
}

func TestDeleteSecurityQuestions_EmptySession(t *testing.T) {
	var (
		err       error
		testTitle = fmt.Sprintf("TestDeleteSecurityQuestions_EmptySession-title_%s", Random(PasswordLength))
	)

	reqDataDeleteSecurityQuestions := struct {
		Title string `json:"title"`
	}{
		Title: testTitle,
	}
	respDataDeleteSecurityQuestions := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	respDeleteSecurityQuestions, err := postJsonRequest(_kongDNS+"/doom/deleteSecurityQuestions/v1", &reqDataDeleteSecurityQuestions, nil, &respDataDeleteSecurityQuestions, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respDeleteSecurityQuestions.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataDeleteSecurityQuestions.Code != slark_response.StatusCodeEmptySession {
		t.Error("not prospective response data code")
		return
	}
}

func TestDeleteSecurityQuestions_EmptyParameter(t *testing.T) {
	var (
		err       error
		testTitle = ""
	)

	reqDataDeleteSecurityQuestions := struct {
		Title string `json:"title"`
	}{
		Title: testTitle,
	}
	respDataDeleteSecurityQuestions := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	respDeleteSecurityQuestions, err := postJsonRequest(_kongDNS+"/doom/deleteSecurityQuestions/v1", &reqDataDeleteSecurityQuestions, _testCookie, &respDataDeleteSecurityQuestions, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respDeleteSecurityQuestions.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataDeleteSecurityQuestions.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
}
