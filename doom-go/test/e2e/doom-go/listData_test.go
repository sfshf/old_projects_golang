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

func TestListData(t *testing.T) {
	var (
		ctx                 = context.Background()
		err                 error
		testPlainText       = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("TestListData-plaintext_%s", Random(PasswordLength))))
		testCipherTextBytes []byte
		testTitle           = fmt.Sprintf("TestListData-title_%s", Random(PasswordLength))
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

	// TestListData ------------------------------------
	respDataListData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []struct {
				Title       string `json:"title"`
				DataID      string `json:"dataID"`
				Description string `json:"description"`
				Date        string `json:"date"`
			} `json:"list"`
		} `json:"data"`
	}{}
	// send request
	respListData, err := postJsonRequest(_kongDNS+"/doom/listData/v1", nil, _testCookie, &respDataListData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respListData.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataListData.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	if len(respDataListData.Data.List) < 1 {
		t.Error("not prospective response data")
		return
	}
}

func TestListData_EmptySession(t *testing.T) {
	respDataListData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			List []struct {
				Title       string `json:"title"`
				DataID      string `json:"dataID"`
				Description string `json:"description"`
				Date        string `json:"date"`
			} `json:"list"`
		} `json:"data"`
	}{}
	// send request
	respListData, err := postJsonRequest(_kongDNS+"/doom/listData/v1", nil, nil, &respDataListData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respListData.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataListData.Code != slark_response.StatusCodeEmptySession {
		t.Error("not prospective response data code")
		return
	}
}
