package tester_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/nextsurfer/tester/api/response"
	tester_mongo "github.com/nextsurfer/tester/internal/pkg/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestUpdateAPITestcase(t *testing.T) {
	apiTestcases := []tester_mongo.ApiTestcase{
		{Name: "tester-e2e", Path: "/path/to/tester-e2e", Body: "body"},
	}
	data, _ := json.Marshal(apiTestcases)
	testApp := "tester-e2e"
	reqData := struct {
		ApiKey string `json:"apiKey"`
		App    string `json:"app"`
		Data   string `json:"data"`
	}{
		ApiKey: _testerApiKey,
		App:    testApp,
		Data:   string(data),
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/tester/updateAPITestcase/v1", &reqData, nil, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	ctx := context.Background()
	coll := _mongoDB.Collection(tester_mongo.CollectionName_AppApiTestcases)
	result, err := coll.DeleteOne(ctx, bson.D{{Key: "app", Value: testApp}})
	if err != nil {
		t.Error(err)
		return
	}
	if result.DeletedCount != 1 {
		t.Error("not prospective response data")
		return
	}
}
