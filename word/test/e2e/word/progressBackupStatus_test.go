package word_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/nextsurfer/word/api/response"
	. "github.com/nextsurfer/word/internal/pkg/model"
)

func TestProgressBackupStatus(t *testing.T) {
	// mock data
	testTS := time.Now()
	mockBackup := ProgressBackup{
		UserID:    _testAccount.ID,
		Version:   1,
		Timestamp: testTS.UnixMilli(),
		Resource:  "/path/to/resource",
	}
	if err := _wordGormDB.Create(&mockBackup).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _wordGormDB.Delete(&ProgressBackup{ID: mockBackup.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()

	// nothing to do
	reqData := struct {
		Timestamp int64 `json:"timestamp"`
		Version   int32 `json:"version"`
	}{
		Timestamp: testTS.UnixMilli(),
		Version:   1,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Code int32 `json:"code"`
		} `json:"data,omitempty"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/word/user/progress/backup/status/v1", &reqData, _testCookie, &respData, nil)
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
	// check data
	if respData.Data.Code != 0 {
		t.Error("not prospective response data")
		return
	}

	// need to upload
	reqData = struct {
		Timestamp int64 `json:"timestamp"`
		Version   int32 `json:"version"`
	}{
		Timestamp: testTS.Add(5 * time.Minute).UnixMilli(),
		Version:   1,
	}
	respData = struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Code int32 `json:"code"`
		} `json:"data,omitempty"`
	}{}
	resp, err = postJsonRequest(_kongDNS+"/word/user/progress/backup/status/v1", &reqData, _testCookie, &respData, nil)
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
	// check data
	if respData.Data.Code != 1 {
		t.Error("not prospective response data")
		return
	}

	// need to download
	reqData = struct {
		Timestamp int64 `json:"timestamp"`
		Version   int32 `json:"version"`
	}{
		Timestamp: testTS.Add(-5 * time.Minute).UnixMilli(),
		Version:   1,
	}
	respData = struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Data         struct {
			Code int32 `json:"code"`
		} `json:"data,omitempty"`
	}{}
	resp, err = postJsonRequest(_kongDNS+"/word/user/progress/backup/status/v1", &reqData, _testCookie, &respData, nil)
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
	// check data
	if respData.Data.Code != 2 {
		t.Error("not prospective response data")
		return
	}
}
