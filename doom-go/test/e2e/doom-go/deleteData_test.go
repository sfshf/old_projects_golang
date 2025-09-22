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

func TestDeleteData(t *testing.T) {
	var (
		ctx                 = context.Background()
		err                 error
		testPlainText       = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("TestDeleteData-plaintext_%s", Random(PasswordLength))))
		testCipherTextBytes []byte
		testTitle           = fmt.Sprintf("TestDeleteData-title_%s", Random(PasswordLength))
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

	// TestDeleteData ------------------------------------
	reqDataDeleteData := struct {
		Title string `json:"title"`
	}{
		Title: testTitle,
	}
	respDataDeleteData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	respDeleteData, err := postJsonRequest(_kongDNS+"/doom/deleteData/v1", &reqDataDeleteData, _testCookie, &respDataDeleteData, nil)
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
	// check data
	var data Datum
	if err := coll.FindOne(ctx, bson.D{
		{Key: "userID", Value: _testAccount.ID},
		{Key: "dataID", Value: dataID},
		{Key: "deletedAt", Value: bson.D{{Key: "$gt", Value: 0}}},
	}).Decode(&data); err != nil {
		t.Error(err)
		return
	}
}

func TestDeleteData_EmptySession(t *testing.T) {
	var (
		err       error
		testTitle = fmt.Sprintf("TestDeleteData_EmptySession-title_%s", Random(PasswordLength))
	)

	reqDataDeleteData := struct {
		Title string `json:"title"`
	}{
		Title: testTitle,
	}
	respDataDeleteData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	respDeleteData, err := postJsonRequest(_kongDNS+"/doom/deleteData/v1", &reqDataDeleteData, nil, &respDataDeleteData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respDeleteData.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataDeleteData.Code != slark_response.StatusCodeEmptySession {
		t.Error("not prospective response data code")
		return
	}
}

func TestDeleteData_EmptyParameter(t *testing.T) {
	var (
		err       error
		testTitle = ""
	)

	reqDataDeleteData := struct {
		Title string `json:"title"`
	}{
		Title: testTitle,
	}
	respDataDeleteData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	respDeleteData, err := postJsonRequest(_kongDNS+"/doom/deleteData/v1", &reqDataDeleteData, _testCookie, &respDataDeleteData, nil)
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
}
