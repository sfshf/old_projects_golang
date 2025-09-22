package pswds_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/nextsurfer/pswds_backend/api/response"
	"github.com/nextsurfer/pswds_backend/internal/common/random"
	"github.com/nextsurfer/pswds_backend/internal/dao"
	. "github.com/nextsurfer/pswds_backend/internal/model"
)

func TestShareDataToFamily(t *testing.T) {
	var (
		ctx = context.Background()
	)
	// mock data
	dataID := random.NewUUIDString()
	daoManager := dao.ManagerWithDB(_pswdsGormDB)
	daoManager.PasswordRecordDAO.Create(ctx, &PasswordRecord{
		DataID:    dataID,
		UpdatedAt: time.Now().Unix(),
		UserID:    _testAccount.ID,
		Content:   "{}",
		Version:   1,
	})
	defer func() {
		if err := _pswdsGormDB.Delete(&PasswordRecord{}, "data_id=?", dataID).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	familyID := random.NewUUIDString()
	daoManager.FamilyMemberDAO.Create(ctx, []*FamilyMember{{
		UserID:   _testAccount2.ID,
		FamilyID: familyID,
		IsAdmin:  dao.FamilyMemberIsNotAdmin,
	}, {
		UserID:   _testAccount.ID,
		FamilyID: familyID,
		IsAdmin:  dao.FamilyMemberIsNotAdmin,
	}})
	defer func() {
		if err := _pswdsGormDB.Delete(&FamilyMember{}, "family_id=?", familyID).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	reqData := struct {
		SharingMembers []int64 `json:"sharingMembers"`
		DataID         string  `json:"dataID"`
		Type           string  `json:"type"`
		Content        string  `json:"content"`
	}{
		DataID:  dataID,
		Type:    "password",
		Content: "{}",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/shareDataToFamily/v1", &reqData, _testCookie, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _pswdsGormDB.Delete(&FamilySharedRecord{}, "data_id=?", dataID).Error; err != nil {
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
	sharedRecord, err := daoManager.FamilySharedRecordDAO.GetByUserIDAndDataID(ctx, _testAccount.ID, dataID)
	if err != nil {
		t.Error(err)
		return
	}
	if sharedRecord == nil {
		t.Error("not prospective db data")
		return
	}
	sharingMembers := make([]int64, 0)
	if err := json.Unmarshal([]byte(sharedRecord.SharingMembers), &sharingMembers); err != nil {
		t.Error(err)
		return
	}
	if len(sharingMembers) != 1 {
		t.Error("not prospective db data")
		return
	}
}

func TestShareDataToFamily_EmptySession(t *testing.T) {
	dataID := random.NewUUIDString()
	reqData := struct {
		SharingMembers []int64 `json:"sharingMembers"`
		DataID         string  `json:"dataID"`
		Type           string  `json:"type"`
		Content        string  `json:"content"`
	}{
		DataID:  dataID,
		Type:    "password",
		Content: "{}",
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/shareDataToFamily/v1", &reqData, nil, &respData, nil)
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
