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

func TestDeletePrivateKeyBackup(t *testing.T) {
	var (
		ctx                 = context.Background()
		err                 error
		testPlainText       = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("TestDeletePrivateKeyBackup-plaintext_%s", Random(PasswordLength))))
		testCipherTextBytes []byte
		testTitle           = fmt.Sprintf("TestDeletePrivateKeyBackup-title_%s", Random(PasswordLength))
	)
	testCipherTextBytes, err = ecies.Encrypt(_publicKey, []byte(testPlainText))
	if err != nil {
		t.Error(err)
		return
	}
	rpcCtx := rpc.NewContext(ctx, _localizerManager)
	dataID := fmt.Sprintf("%d-pkbp-%s", _testAccount.ID, testTitle)
	// mock data
	if err := connector_grpc.SaveData(ctx, rpcCtx, _connectorApiKey, _connectorKeyID, dataID, testPlainText, base64.StdEncoding.EncodeToString(testCipherTextBytes), true); err != nil {
		t.Error(err)
		return
	}
	ts := time.Now().UnixMilli()
	newPrivateKeyBackup := &PrivateKeyBackup{
		CreatedAt: ts,
		UpdatedAt: ts,
		UserID:    _testAccount.ID,
		DataID:    dataID,
		Title:     testTitle,
	}
	coll := _mongoDB.Collection(CollectionName_PrivateKeyBackup)
	result, err := coll.InsertOne(ctx, newPrivateKeyBackup)
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

	// TestDeletePrivateKeyBackup ------------------------------------
	reqDataDeletePrivateKeyBackup := struct {
		Title string `json:"title"`
	}{
		Title: testTitle,
	}
	respDataDeletePrivateKeyBackup := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	respDeletePrivateKeyBackup, err := postJsonRequest(_kongDNS+"/doom/deletePrivateKeyBackup/v1", &reqDataDeletePrivateKeyBackup, _testCookie, &respDataDeletePrivateKeyBackup, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respDeletePrivateKeyBackup.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataDeletePrivateKeyBackup.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	// check data
	var privateKeyBackup PrivateKeyBackup
	if err := coll.FindOne(ctx, bson.D{
		{Key: "userID", Value: _testAccount.ID},
		{Key: "dataID", Value: dataID},
		{Key: "deletedAt", Value: bson.D{{Key: "$gt", Value: 0}}},
	}).Decode(&privateKeyBackup); err != nil {
		t.Error(err)
		return
	}
}

func TestDeletePrivateKeyBackup_EmptySession(t *testing.T) {
	var (
		err       error
		testTitle = fmt.Sprintf("TestDeletePrivateKeyBackup_EmptySession-title_%s", Random(PasswordLength))
	)

	reqDataDeletePrivateKeyBackup := struct {
		Title string `json:"title"`
	}{
		Title: testTitle,
	}
	respDataDeletePrivateKeyBackup := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	respDeletePrivateKeyBackup, err := postJsonRequest(_kongDNS+"/doom/deletePrivateKeyBackup/v1", &reqDataDeletePrivateKeyBackup, nil, &respDataDeletePrivateKeyBackup, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respDeletePrivateKeyBackup.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataDeletePrivateKeyBackup.Code != slark_response.StatusCodeEmptySession {
		t.Error("not prospective response data code")
		return
	}
}

func TestDeletePrivateKeyBackup_EmptyParameter(t *testing.T) {
	var (
		err       error
		testTitle = ""
	)

	reqDataDeletePrivateKeyBackup := struct {
		Title string `json:"title"`
	}{
		Title: testTitle,
	}
	respDataDeletePrivateKeyBackup := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	respDeletePrivateKeyBackup, err := postJsonRequest(_kongDNS+"/doom/deletePrivateKeyBackup/v1", &reqDataDeletePrivateKeyBackup, _testCookie, &respDataDeletePrivateKeyBackup, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respDeletePrivateKeyBackup.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataDeletePrivateKeyBackup.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
}
