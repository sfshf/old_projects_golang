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

func TestInviteFamilyMember(t *testing.T) {
	var (
		ctx = context.Background()
	)
	// mock data
	daoManager := dao.ManagerWithDB(_pswdsGormDB)
	familyID := random.NewUUIDString()
	daoManager.FamilyMemberDAO.Create(ctx, &FamilyMember{
		UserID:   _testAccount.ID,
		FamilyID: familyID,
		IsAdmin:  dao.FamilyMemberIsAdmin,
	})
	defer func() {
		if err := _pswdsGormDB.Delete(&FamilyMember{}, "family_id=?", familyID).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	reqData := struct {
		Email              string `json:"email"`
		EncryptedFamilyKey string `json:"encryptedFamilyKey"`
	}{
		Email:              _testEmail2,
		EncryptedFamilyKey: "ffffff",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/inviteFamilyMember/v1", &reqData, _testCookie, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _pswdsGormDB.Delete(&FamilyInvitation{}, "family_id=?", familyID).Error; err != nil {
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
}

func TestInviteFamilyMember_EmptySession(t *testing.T) {
	reqData := struct {
		Email              string `json:"email"`
		EncryptedFamilyKey string `json:"encryptedFamilyKey"`
	}{
		Email:              _testEmail2,
		EncryptedFamilyKey: "ffffff",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/inviteFamilyMember/v1", &reqData, nil, &respData, nil)
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
