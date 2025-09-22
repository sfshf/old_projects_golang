package pswds_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/nextsurfer/pswds_backend/api/response"
	"github.com/nextsurfer/pswds_backend/internal/common/random"
	"github.com/nextsurfer/pswds_backend/internal/dao"
	. "github.com/nextsurfer/pswds_backend/internal/model"
)

func TestStopSharingData(t *testing.T) {
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
	daoManager.FamilySharedRecordDAO.Create(ctx, &FamilySharedRecord{
		DataID:    dataID,
		UpdatedAt: time.Now().Unix(),
		FamilyID:  familyID,
		SharedBy:  _testAccount.ID,
		Type:      "password",
		Content:   "{}",
		Version:   1,
	})
	defer func() {
		if err := _pswdsGormDB.Delete(&FamilySharedRecord{}, "data_id=?", dataID).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	daoManager.FamilyMemberDAO.Create(ctx, []*FamilyMember{{
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
		DataID         string  `json:"dataID"`
		SharingMembers []int64 `json:"sharingMembers"`
		Stop           bool    `json:"stop"`
	}{
		DataID: dataID,
		Stop:   true,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/manageSharingData/v1", &reqData, _testCookie, &respData, nil)
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
	sharedRecord, err := daoManager.FamilySharedRecordDAO.GetByUserIDAndDataID(ctx, _testAccount.ID, dataID)
	if err != nil {
		t.Error(err)
		return
	}
	if sharedRecord != nil {
		t.Error("not prospective db data")
		return
	}
}

func TestStopSharingData_EmptySession(t *testing.T) {
	dataID := random.NewUUIDString()
	reqData := struct {
		DataID         string  `json:"dataID"`
		SharingMembers []int64 `json:"sharingMembers"`
		Stop           bool    `json:"stop"`
	}{
		DataID: dataID,
		Stop:   true,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
		Oracle       string `json:"oracle"`
	}{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/manageSharingData/v1", &reqData, nil, &respData, nil)
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
