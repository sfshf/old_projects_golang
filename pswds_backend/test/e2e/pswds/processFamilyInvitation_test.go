package pswds_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/nextsurfer/pswds_backend/api/response"
	"github.com/nextsurfer/pswds_backend/internal/common/random"
	"github.com/nextsurfer/pswds_backend/internal/dao"
	. "github.com/nextsurfer/pswds_backend/internal/model"
)

func TestProcessFamilyInvitation(t *testing.T) {
	var (
		ctx = context.Background()
	)
	// mock data
	daoManager := dao.ManagerWithDB(_pswdsGormDB)
	familyID := random.NewUUIDString()
	mockInvitation := &FamilyInvitation{
		InvitedBy: _testAccount2.ID,
		FamilyID:  familyID,
		Email:     _testEmail,
	}
	daoManager.FamilyInvitationDAO.Create(ctx, mockInvitation)
	defer func() {
		if err := _pswdsGormDB.Delete(&FamilyInvitation{}, "family_id=?", familyID).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	reqData := struct {
		Id     int64 `json:"id"`
		Accept bool  `json:"accept"`
	}{
		Id:     mockInvitation.ID,
		Accept: true,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/processFamilyInvitation/v1", &reqData, _testCookie, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _pswdsGormDB.Delete(&FamilyMember{}, "family_id=?", familyID).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeOK {
		t.Error("not prospective response data code")
		return
	}
	record, err := daoManager.FamilyMemberDAO.GetByUserID(ctx, _testAccount.ID)
	if err != nil {
		t.Error(err)
		return
	}
	if record == nil {
		t.Error("not prospective db data")
		return
	}
}

func TestProcessFamilyInvitation_EmptySession(t *testing.T) {
	reqData := struct {
		Id     int64 `json:"id"`
		Accept bool  `json:"accept"`
	}{
		Id:     1,
		Accept: true,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/processFamilyInvitation/v1", &reqData, nil, &respData, nil)
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
