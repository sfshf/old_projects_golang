package admin_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/oracle/api/response"
)

func TestConsolePing(t *testing.T) {
	respDataPing := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         string `json:"data"`
	}{}
	respPing, err := postJsonRequest(_oracleConsoleDNS+"/console/ping/v1", nil, nil, &respDataPing, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if respPing.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respDataPing.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	if respDataPing.Data != "pong" {
		t.Error("not prospective response data")
		return
	}
}
