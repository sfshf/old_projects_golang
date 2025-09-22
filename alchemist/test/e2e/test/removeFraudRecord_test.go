package test_test

import (
	"net/http"
	"testing"

	"github.com/nextsurfer/alchemist/api/response"
	. "github.com/nextsurfer/alchemist/internal/pkg/model"
	"github.com/nextsurfer/alchemist/internal/pkg/util"
)

func TestRemoveFraudRecord(t *testing.T) {
	// mock data
	testAppID := "alchemist"
	testReferralCode := util.GenerateReferralCode()

	mockUserRegisteredOnOldDevice := UserRegisteredOnOldDevice{
		App:          testAppID,
		UserID:       _testAccount.ID,
		ReferralCode: testReferralCode,
	}
	if err := _alchemistGormDB.Create(&mockUserRegisteredOnOldDevice).Error; err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _alchemistGormDB.Delete(&UserRegisteredOnOldDevice{ID: mockUserRegisteredOnOldDevice.ID}).Error; err != nil {
			t.Error(err)
			return
		}
	}()

	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/alchemist/test/removeFraudRecord/v1", nil, _testCookie, &respData, nil)
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
	// check data
	var record UserRegisteredOnOldDevice
	if err := _alchemistGormDB.Table(TableNameUserRegisteredOnOldDevice).
		Where("id=? AND user_id=? AND deleted_at>0", mockUserRegisteredOnOldDevice.ID, _testAccount.ID).
		First(&record).Error; err != nil {
		t.Error(err)
		return
	}
}
