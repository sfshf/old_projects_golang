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

func TestLeaveFamily(t *testing.T) {
	var (
		ctx = context.Background()
	)
	// mock data
	daoManager := dao.ManagerWithDB(_pswdsGormDB)
	familyID := random.NewUUIDString()
	daoManager.FamilyMemberDAO.Create(ctx, &FamilyMember{
		UserID:   _testAccount.ID,
		FamilyID: familyID,
		IsAdmin:  dao.FamilyMemberIsNotAdmin,
	})
	defer func() {
		if err := _pswdsGormDB.Delete(&FamilyMember{}, "family_id=?", familyID).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/leaveFamily/v1", nil, _testCookie, &respData, nil)
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
	record, err := daoManager.FamilyMemberDAO.GetByUserID(ctx, _testAccount.ID)
	if err != nil {
		t.Error(err)
		return
	}
	if record != nil {
		t.Error("not prospective db data")
		return
	}
}

func TestLeaveFamily_EmptySession(t *testing.T) {
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/leaveFamily/v1", nil, nil, &respData, nil)
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
