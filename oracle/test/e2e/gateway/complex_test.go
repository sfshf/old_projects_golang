package gateway_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"testing"
	"time"

	console_api "github.com/nextsurfer/oracle/api/console"
	"github.com/nextsurfer/oracle/api/response"
	. "github.com/nextsurfer/oracle/internal/model"
	console_grpc "github.com/nextsurfer/oracle/pkg/console/grpc"
)

var e2eProto = `syntax = "proto3";

package oracle_e2e_api;

import "google/api/annotations.proto";
import "google/protobuf/struct.proto";

service OracleE2eService {
    rpc Ping(Empty) returns (EmptyResponse) {
        option (google.api.http) = {
            post: "/oracle-e2e/ping/v1"
            body: "*"
        };
    };
}

message Empty {
}

message EmptyResponse {
    int32 code = 1;
    string message = 2;
    google.protobuf.Value debugMessage = 3;
}
`

func TestService(t *testing.T) {
	ctx := context.Background()
	// TestUpsertService
	respData, err := console_grpc.UpsertService(ctx, &console_api.UpsertServiceRequest{
		Name:        "oracle-e2e",
		Application: "oracle-e2e",
		PathPrefix:  "/oracle-e2e",
		ProtoFile:   base64.StdEncoding.EncodeToString([]byte(e2eProto)),
	})
	if err != nil {
		t.Error(err)
		return
	}
	if respData.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	// check data
	var application Application
	if err := _oracleGormDB.Model(&Application{}).
		Where(`name=? AND deleted_at=0`, "oracle-e2e").
		First(&application).Error; err != nil {
		t.Error("no prospective data")
		return
	}
	var service Service
	if err := _oracleGormDB.Model(&Service{}).
		Where(`name=? AND application_id=? AND deleted_at=0`, "oracle-e2e", application.ID).
		First(&service).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _oracleGormDB.Delete(&Application{ID: application.ID}).Error; err != nil {
			t.Error(err)
			return
		}
		if err := _oracleGormDB.Delete(&Service{ID: service.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	time.Sleep(5 * time.Second)
	// TestOracleE2ePing_Failed
	respDataOracleE2ePing := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	respOracleE2ePing, err := postJsonRequest(_oracleGwDNS+"/oracle-e2e/ping/v1", nil, nil, &respDataOracleE2ePing, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respOracleE2ePing.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataOracleE2ePing.Code != response.StatusCodeInternalServerError {
		t.Error("not prospective response data code")
		return
	}
	// check data
	if !strings.Contains(respDataOracleE2ePing.DebugMessage, "consul service not found") {
		t.Error("not prospective response data")
		return
	}
	// TestDeleteService
	reqDataDeleteService := struct {
		ApiKey string `json:"apiKey"`
		Name   string `json:"name"`
	}{
		ApiKey: _oracleConsoleApiKey,
		Name:   "oracle-e2e",
	}
	respDataDeleteService := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	respDeleteService, err := postJsonRequest(_oracleConsoleDNS+"/console/deleteService/v1", &reqDataDeleteService, nil, &respDataDeleteService, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respDeleteService.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataDeleteService.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	// check data
	service = Service{}
	if err := _oracleGormDB.Model(&Service{}).
		Where(`name=? AND application_id=? AND deleted_at>0`, "oracle-e2e", application.ID).
		First(&service).Error; err != nil {
		t.Error(err)
		return
	}
	// TestDeleteApplication
	reqDataDeleteApplication := struct {
		ApiKey string `json:"apiKey"`
		Name   string `json:"name"`
	}{
		ApiKey: _oracleConsoleApiKey,
		Name:   "oracle-e2e",
	}
	respDataDeleteApplication := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	respDeleteApplication, err := postJsonRequest(_oracleConsoleDNS+"/console/deleteApplication/v1", &reqDataDeleteApplication, nil, &respDataDeleteApplication, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respDeleteApplication.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataDeleteApplication.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	// check data
	application = Application{}
	if err := _oracleGormDB.Model(&Application{}).
		Where(`name=? AND deleted_at>0`, "oracle-e2e").
		First(&application).Error; err != nil {
		t.Error(err)
		return
	}
}
