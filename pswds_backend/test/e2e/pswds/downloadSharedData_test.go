package pswds_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/pswds_backend/api/response"
)

func TestDownloadSharedData(t *testing.T) {
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
		Data         struct {
			SharingList []struct {
				Id        string `json:"id"`
				UpdatedAt int64  `json:"updatedAt"`
				FamilyID  string `json:"familyID"`
				SharedBy  int64  `json:"sharedBy"`
				Type      string `json:"type"`
				Content   string `json:"content"`
				Version   int64  `json:"version"`
			} `json:"sharingList"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/downloadSharedData/v1", nil, _testCookie, &respData, nil)
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
}

func TestDownloadSharedData_EmptySession(t *testing.T) {
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
		Data         struct {
			SharingList []struct {
				Id        string `json:"id"`
				UpdatedAt int64  `json:"updatedAt"`
				FamilyID  string `json:"familyID"`
				SharedBy  int64  `json:"sharedBy"`
				Type      string `json:"type"`
				Content   string `json:"content"`
				Version   int64  `json:"version"`
			} `json:"sharingList"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/downloadSharedData/v1", nil, nil, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeUnauthorized {
		t.Error("not prospective response data code")
		return
	}
}
