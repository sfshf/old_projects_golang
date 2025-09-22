package pswds_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/nextsurfer/pswds_backend/api/response"
	"github.com/nextsurfer/pswds_backend/internal/dao"
	. "github.com/nextsurfer/pswds_backend/internal/model"
)

func TestCreateFamily(t *testing.T) {
	var (
		ctx        = context.Background()
		daoManager = dao.ManagerWithDB(_pswdsGormDB)
	)
	reqData := struct {
		Description        string `json:"description"`
		EncryptedFamilyKey string `json:"encryptedFamilyKey"`
	}{
		Description:        "testfamily",
		EncryptedFamilyKey: "ffffff",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/createFamily/v1", &reqData, _testCookie, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		family, err := daoManager.FamilyDAO.GetByCreator(ctx, _testAccount.ID)
		if err != nil {
			t.Error(err)
			return
		}
		if family == nil {
			t.Error("not prospective data")
			return
		}
		if err := _pswdsGormDB.Delete(&Family{}, "created_by=?", _testAccount.ID).Error; err != nil {
			t.Error(err)
			return
		}
		if err := _pswdsGormDB.Delete(&FamilyMember{}, "user_id=?", _testAccount.ID).Error; err != nil {
			t.Error(err)
			return
		}
		if err := _pswdsGormDB.Delete(&FamilyMessage{}, "family_id=?", family.FamilyID).Error; err != nil {
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

func TestCreateFamily_EmptySession(t *testing.T) {
	reqData := struct {
		Description        string `json:"description"`
		EncryptedFamilyKey string `json:"encryptedFamilyKey"`
	}{
		Description:        "testfamily",
		EncryptedFamilyKey: "ffffff",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/createFamily/v1", &reqData, nil, &respData, nil)
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
