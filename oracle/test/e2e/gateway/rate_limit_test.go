package gateway_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"sync"
	"testing"
	"time"

	console_api "github.com/nextsurfer/oracle/api/console"
	"github.com/nextsurfer/oracle/api/response"
	. "github.com/nextsurfer/oracle/internal/model"
	console_grpc "github.com/nextsurfer/oracle/pkg/console/grpc"
)

var e2eRateLimitProto = `syntax = "proto3";

package oracle_e2e_rate_limit_api;

import "google/api/annotations.proto";
import "google/protobuf/struct.proto";

service OracleE2eRateLimitService {
    rpc Ping(Empty) returns (EmptyResponse) {
        option (google.api.http) = {
            post: "/oracle-e2e-rate-limit/ping/v1"
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

func TestRateLimit(t *testing.T) {
	var (
		err error
		ctx = context.Background()
	)
	// TestUpsertService
	respData, err := console_grpc.UpsertService(ctx, &console_api.UpsertServiceRequest{
		Name:        "oracle-e2e-rate-limit",
		Application: "oracle-e2e-rate-limit",
		PathPrefix:  "/oracle-e2e-rate-limit",
		ProtoFile:   base64.StdEncoding.EncodeToString([]byte(e2eRateLimitProto)),
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
		Where(`name=? AND deleted_at=0`, "oracle-e2e-rate-limit").
		First(&application).Error; err != nil {
		t.Error(err)
		return
	}
	var service Service
	if err := _oracleGormDB.Model(&Service{}).
		Where(`name=? AND application_id=? AND deleted_at=0`, "oracle-e2e-rate-limit", application.ID).
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
	// TestOracleE2eRateLimitPing_Failed
	// send 10 requests once, and check response of the last request
	respDataOracleE2eRateLimitPing := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	rate := time.Second / 10
	ticker := time.NewTicker(rate)
	var wg sync.WaitGroup
	var respOracleE2eRateLimitPing *http.Response
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int, wg *sync.WaitGroup) {
			defer wg.Done()
			if i != 9 {
				if _, err := postJsonRequest(_oracleGwDNS+"/oracle-e2e-rate-limit/ping/v1", nil, nil, nil, nil); err != nil {
					t.Error(err)
					return
				}
			} else {
				respOracleE2eRateLimitPing, err = postJsonRequest(_oracleGwDNS+"/oracle-e2e-rate-limit/ping/v1", nil, nil, &respDataOracleE2eRateLimitPing, nil)
				if err != nil {
					t.Error(err)
					return
				}
			}
		}(i, &wg)
		<-ticker.C
	}
	wg.Wait()
	ticker.Stop()
	if respOracleE2eRateLimitPing == nil {
		t.Error("not prospective response data")
		return
	}
	if respOracleE2eRateLimitPing.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataOracleE2eRateLimitPing.Code != response.StatusCodeTooManyRequests {
		t.Error("not prospective response data code")
		return
	}
	// TestDeleteService
	reqDataDeleteService := struct {
		ApiKey string `json:"apiKey"`
		Name   string `json:"name"`
	}{
		ApiKey: _oracleConsoleApiKey,
		Name:   "oracle-e2e-rate-limit",
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
		Where(`name=? AND application_id=? AND deleted_at>0`, "oracle-e2e-rate-limit", application.ID).
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
		Name:   "oracle-e2e-rate-limit",
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
		Where(`name=? AND deleted_at>0`, "oracle-e2e-rate-limit").
		First(&application).Error; err != nil {
		t.Error(err)
		return
	}
}
