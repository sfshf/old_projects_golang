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

func TestGetData(t *testing.T) {
	var (
		ctx                 = context.Background()
		err                 error
		testPlainText       = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("TestGetData-plaintext_%s", Random(PasswordLength))))
		testCipherTextBytes []byte
		testTitle           = fmt.Sprintf("TestGetData-title_%s", Random(PasswordLength))
	)
	testCipherTextBytes, err = ecies.Encrypt(_publicKey, []byte(testPlainText))
	if err != nil {
		t.Error(err)
		return
	}
	rpcCtx := rpc.NewContext(ctx, _localizerManager)
	dataID := fmt.Sprintf("%d-data-%s", _testAccount.ID, testTitle)
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
	newDatum := &Datum{
		CreatedAt: ts,
		UpdatedAt: ts,
		UserID:    _testAccount.ID,
		DataID:    dataID,
		Title:     testTitle,
	}
	coll := _mongoDB.Collection(CollectionName_Datum)
	result, err := coll.InsertOne(ctx, newDatum)
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

	// TestGetData ------------------------------------
	reqDataGetData := struct {
		Title string `json:"title"`
	}{
		Title: testTitle,
	}
	respDataGetData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			PlainText string `json:"plainText"`
		} `json:"data"`
	}{}
	// send request
	respGetData, err := postJsonRequest(_kongDNS+"/doom/getData/v1", &reqDataGetData, _testCookie, &respDataGetData, nil)
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
	if respDataGetData.Data.PlainText != testPlainText {
		t.Error("not prospective response data")
		return
	}
}

func TestGetData_EmptySession(t *testing.T) {
	var (
		err       error
		testTitle = fmt.Sprintf("TestGetData_EmptySession-title_%s", Random(PasswordLength))
	)

	reqDataGetData := struct {
		Title string `json:"title"`
	}{
		Title: testTitle,
	}
	respDataGetData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			PlainText string `json:"plainText"`
		} `json:"data"`
	}{}
	// send request
	respGetData, err := postJsonRequest(_kongDNS+"/doom/getData/v1", &reqDataGetData, nil, &respDataGetData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respGetData.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataGetData.Code != slark_response.StatusCodeEmptySession {
		t.Error("not prospective response data code")
		return
	}
}

func TestGetData_EmptyParameter(t *testing.T) {
	var (
		err       error
		testTitle = ""
	)

	reqDataGetData := struct {
		Title string `json:"title"`
	}{
		Title: testTitle,
	}
	respDataGetData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			PlainText string `json:"plainText"`
		} `json:"data"`
	}{}
	// send request
	respGetData, err := postJsonRequest(_kongDNS+"/doom/getData/v1", &reqDataGetData, _testCookie, &respDataGetData, nil)
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
}
