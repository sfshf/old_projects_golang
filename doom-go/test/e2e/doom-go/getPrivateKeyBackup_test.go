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

func TestGetPrivateKeyBackup(t *testing.T) {
	var (
		ctx                 = context.Background()
		err                 error
		testPlainText       = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("TestGetPrivateKeyBackup-plaintext_%s", Random(PasswordLength))))
		testCipherTextBytes []byte
		testTitle           = fmt.Sprintf("TestGetPrivateKeyBackup-title_%s", Random(PasswordLength))
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
	defer func() {
		// delete connector mock data
		if err := connector_grpc.DeleteData(ctx, rpcCtx, _connectorApiKey, _connectorKeyID, dataID); err != nil {
			t.Error(err)
			return
		}
	}()
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

	// TestGetPrivateKeyBackup ------------------------------------
	reqDataGetPrivateKeyBackup := struct {
		Title string `json:"title"`
	}{
		Title: testTitle,
	}
	respDataGetPrivateKeyBackup := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			PlainText string `json:"plainText"`
		} `json:"data"`
	}{}
	// send request
	respGetPrivateKeyBackup, err := postJsonRequest(_kongDNS+"/doom/getPrivateKeyBackup/v1", &reqDataGetPrivateKeyBackup, _testCookie, &respDataGetPrivateKeyBackup, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respGetPrivateKeyBackup.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataGetPrivateKeyBackup.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	if respDataGetPrivateKeyBackup.Data.PlainText != testPlainText {
		t.Error("not prospective response data")
		return
	}
}

func TestGetPrivateKeyBackup_EmptySession(t *testing.T) {
	var (
		err       error
		testTitle = fmt.Sprintf("TestGetPrivateKeyBackup_EmptySession-title_%s", Random(PasswordLength))
	)

	reqDataGetPrivateKeyBackup := struct {
		Title string `json:"title"`
	}{
		Title: testTitle,
	}
	respDataGetPrivateKeyBackup := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			PlainText string `json:"plainText"`
		} `json:"data"`
	}{}
	// send request
	respGetPrivateKeyBackup, err := postJsonRequest(_kongDNS+"/doom/getPrivateKeyBackup/v1", &reqDataGetPrivateKeyBackup, nil, &respDataGetPrivateKeyBackup, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respGetPrivateKeyBackup.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataGetPrivateKeyBackup.Code != slark_response.StatusCodeEmptySession {
		t.Error("not prospective response data code")
		return
	}
}

func TestGetPrivateKeyBackup_EmptyParameter(t *testing.T) {
	var (
		err       error
		testTitle = ""
	)

	reqDataGetPrivateKeyBackup := struct {
		Title string `json:"title"`
	}{
		Title: testTitle,
	}
	respDataGetPrivateKeyBackup := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			PlainText string `json:"plainText"`
		} `json:"data"`
	}{}
	// send request
	respGetPrivateKeyBackup, err := postJsonRequest(_kongDNS+"/doom/getPrivateKeyBackup/v1", &reqDataGetPrivateKeyBackup, _testCookie, &respDataGetPrivateKeyBackup, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respGetPrivateKeyBackup.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataGetPrivateKeyBackup.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
}
