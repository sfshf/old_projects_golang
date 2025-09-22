package pswds_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/pswds_backend/api/response"
)

func TestGetFamilyInfo(t *testing.T) {
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
		Data         struct {
			HasFamily     bool   `json:"hasFamily"`
			Description   string `json:"description"`
			FamilyMembers []struct {
				Id       string `json:"id"`
				UserID   int64  `json:"userID"`
				Email    string `json:"email"`
				FamilyID string `json:"familyID"`
				JoinedAt int64  `json:"joinedAt"`
				IsAdmin  bool   `json:"isAdmin"`
			} `json:"familyMembers"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/getFamilyInfo/v1", nil, _testCookie, &respData, nil)
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
	if respData.Data.HasFamily {
		t.Error("not prospective response data")
		return
	}
}

func TestGetFamilyInfo_EmptySession(t *testing.T) {
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
		Data         struct {
			HasFamily     bool   `json:"hasFamily"`
			Description   string `json:"description"`
			FamilyMembers []struct {
				Id       string `json:"id"`
				UserID   int64  `json:"userID"`
				Email    string `json:"email"`
				FamilyID string `json:"familyID"`
				JoinedAt int64  `json:"joinedAt"`
				IsAdmin  bool   `json:"isAdmin"`
			} `json:"familyMembers"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/getFamilyInfo/v1", nil, nil, &respData, nil)
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
