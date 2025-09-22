package handler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nextsurfer/book-manage-api/api"
	"github.com/nextsurfer/book-manage-api/internal/tools"
)

func init() {
	tools.InitConfig("../../../conf.json")
	tools.InitCommonTools()
}

func TestDepositAddressHandler_FetchAddress(t *testing.T) {
	type args struct {
		Request *http.Request
	}

	tests := []struct {
		name         string
		h            *BookHandler
		args         args
		expectedCode int32
	}{
		{
			"1",
			NewBookHandler(),
			args{
				Request: &http.Request{
					Method:     "POST",
					URL:        nil,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"name":"test","description":"test description"}`))),
					RemoteAddr: "",
				},
			},
			int32(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.h
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = tt.args.Request
			h.AddBook(c)
			body, err := ioutil.ReadAll(w.Result().Body)
			if err != nil {
				t.Errorf("DepositAddressHandler.FetchAddress() test read response body failed, statusCode: %d, err: %s",
					w.Result().StatusCode, err.Error())
				return
			}
			resp := &api.Response{}
			if err := json.Unmarshal(body, resp); err != nil {
				t.Errorf("DepositAddressHandler.FetchAddress() test read decode body failed, statusCode: %d, err: %s, body: %s",
					w.Result().StatusCode, err.Error(), string(body))
				return
			}
			if w.Result().StatusCode != 200 {
				t.Errorf("DepositAddressHandler.FetchAddress() test Fail, statusCode: %d, resp = %+v",
					w.Result().StatusCode, resp)
			}
			if resp.Code != tt.expectedCode {
				t.Errorf("DepositAddressHandler.FetchAddress() test failed, code: %d, expectedCode: %d, resp: %+v",
					resp.Code, tt.expectedCode, resp)
			} else {
				t.Logf("DepositAddressHandler.FetchAddress()test OK, name = %s, code = %d, expectedCode = %d, resp = %+v",
					tt.name, resp.Code, tt.expectedCode, resp)
			}
		})
	}
}
