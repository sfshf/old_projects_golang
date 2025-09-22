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

func TestListSecurityQuestions(t *testing.T) {
	var (
		ctx                 = context.Background()
		err                 error
		testPlainText       = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("TestListSecurityQuestions-plaintext_%s", Random(PasswordLength))))
		testCipherTextBytes []byte
		testTitle           = fmt.Sprintf("TestListSecurityQuestions-title_%s", Random(PasswordLength))
		testDescription     = fmt.Sprintf("TestListSecurityQuestions-description_%s", Random(PasswordLength))
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

	// TestListSecurityQuestions ------------------------------------
	respDataListSecurityQuestions := struct {
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
	respListSecurityQuestions, err := postJsonRequest(_kongDNS+"/doom/listSecurityQuestions/v1", nil, _testCookie, &respDataListSecurityQuestions, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respListSecurityQuestions.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataListSecurityQuestions.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	if len(respDataListSecurityQuestions.Data.List) < 1 {
		t.Error("not prospective response data")
		return
	}
}

func TestListSecurityQuestions_EmptySession(t *testing.T) {
	respDataListSecurityQuestions := struct {
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
	respListSecurityQuestions, err := postJsonRequest(_kongDNS+"/doom/listSecurityQuestions/v1", nil, nil, &respDataListSecurityQuestions, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respListSecurityQuestions.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataListSecurityQuestions.Code != slark_response.StatusCodeEmptySession {
		t.Error("not prospective response data code")
		return
	}
}
