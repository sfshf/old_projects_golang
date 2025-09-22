package pswds_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/pswds_backend/api/response"
)

func TestCheckFamilyInvitation(t *testing.T) {
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
		Data         struct {
			Id            int64  `json:"id"`
			InvitedBy     string `json:"invitedBy"`
			InvitedAt     int64  `json:"invitedAt"`
			HasInvitation bool   `json:"hasInvitation"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/checkFamilyInvitation/v1", nil, _testCookie, &respData, nil)
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
	if respData.Data.HasInvitation {
		t.Error("not prospective response data")
		return
	}
}

func TestCheckFamilyInvitation_EmptySession(t *testing.T) {
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
		Data         struct {
			Id            int64  `json:"id"`
			InvitedBy     string `json:"invitedBy"`
			InvitedAt     int64  `json:"invitedAt"`
			HasInvitation bool   `json:"hasInvitation"`
		} `json:"data"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/checkFamilyInvitation/v1", nil, nil, &respData, nil)
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
