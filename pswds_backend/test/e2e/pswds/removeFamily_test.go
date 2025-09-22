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

func TestRemoveFamily(t *testing.T) {
	var (
		ctx = context.Background()
	)
	// mock data
	daoManager := dao.ManagerWithDB(_pswdsGormDB)
	familyID := random.NewUUIDString()
	daoManager.FamilyDAO.Create(ctx, &Family{
		CreatedBy:   _testAccount.ID,
		FamilyID:    familyID,
		Description: "TestRemoveFamily",
	})
	daoManager.FamilyMemberDAO.Create(ctx, &FamilyMember{
		UserID:   _testAccount.ID,
		FamilyID: familyID,
		IsAdmin:  dao.FamilyMemberIsAdmin,
	})
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/removeFamily/v1", nil, _testCookie, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _pswdsGormDB.Delete(&Family{}, "created_by=?", _testAccount.ID).Error; err != nil {
			t.Error(err)
			return
		}
		if err := _pswdsGormDB.Delete(&FamilyMember{}, "user_id=?", _testAccount.ID).Error; err != nil {
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
	family, err := daoManager.FamilyDAO.GetByCreator(ctx, _testAccount.ID)
	if err != nil {
		t.Error(err)
		return
	}
	if family != nil {
		t.Error("not prospective response data")
		return
	}
	count, err := daoManager.FamilyMemberDAO.CountByFamilyID(ctx, familyID)
	if err != nil {
		t.Error(err)
		return
	}
	if count != 0 {
		t.Error("not prospective response data")
		return
	}
}

func TestRemoveFamily_EmptySession(t *testing.T) {
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/removeFamily/v1", nil, nil, &respData, nil)
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
