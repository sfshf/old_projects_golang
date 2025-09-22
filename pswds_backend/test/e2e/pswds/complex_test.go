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

func TestI18nEnglish(t *testing.T) {
	reqData := struct {
		UpdatedAt int64 `json:"updatedAt"`
	}{
		UpdatedAt: -1,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/pswds/checkUpdates/v1", &reqData, nil, &respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
	if respData.Message != "Request has wrong parameters, please inspect parameters" {
		t.Error("not prospective response data")
		return
	}
}

func TestI18nChinese(t *testing.T) {
	reqData := struct {
		UpdatedAt int64 `json:"updatedAt"`
	}{
		UpdatedAt: -1,
	}
	respData := struct {
		Code         int32  `json:"code"`
		Message      string `json:"message"`
		DebugMessage string `json:"debugMessage"`
	}{}
	// send request
	resp, err := postJsonRequest(_kongDNS+"/pswds/checkUpdates/v1", &reqData, nil, &respData, func(req *http.Request) {
		req.Header.Set("Accept-Language", "zh")
	})
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("not prospective response code")
		return
	}
	if respData.Code != response.StatusCodeWrongParameters {
		t.Error("not prospective response data code")
		return
	}
	if respData.Message != "请求参数有误，请您排错后重试" {
		t.Error("not prospective response data")
		return
	}
}

type RespData struct {
	Code         int32       `json:"code"`
	Message      string      `json:"message"`
	DebugMessage string      `json:"debugMessage"`
	Oracle       string      `json:"oracle"`
	Data         interface{} `json:"data,omitempty"`
}

func TestFamilyShareDatas(t *testing.T) {
	var (
		ctx        = context.Background()
		daoManager = dao.ManagerWithDB(_pswdsGormDB)
		reqData    interface{}
		respData   *RespData
	)
	updatedAt := time.Now().Unix()
	respData = &RespData{}
	// 给两个账号分别添加backup record
	if err := daoManager.BackupDAO.Create(ctx, &Backup{
		UpdatedAt:     updatedAt,
		UserID:        _testAccount.ID,
		PasswordHash:  "xxxxxx",
		UserPublicKey: "xxxxxx_pk",
	}); err != nil {
		t.Error(err)
		return
	}
	if err := daoManager.BackupDAO.Create(ctx, &Backup{
		UpdatedAt:     updatedAt,
		UserID:        _testAccount2.ID,
		PasswordHash:  "yyyyyy",
		UserPublicKey: "yyyyyy_pk",
	}); err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _pswdsGormDB.Delete(&Backup{}, "user_id=?", _testAccount.ID).Error; err != nil {
			t.Error(err)
			return
		}
		if err := _pswdsGormDB.Delete(&Backup{}, "user_id=?", _testAccount2.ID).Error; err != nil {
			t.Error(err)
			return
		}
	}()
	// 添加password record
	dataID := random.NewUUIDString()
	updatedAt = time.Now().Unix()
	reqData = &struct {
		UpdatedAt int64  `json:"updatedAt"`
		DataID    string `json:"dataID"`
		Content   string `json:"content"`
	}{
		UpdatedAt: updatedAt,
		DataID:    dataID,
		Content:   "{}",
	}
	respData = &RespData{}
	resp, err := postJsonRequest(_kongDNS+"/pswds/createPasswordRecord/v1", reqData, _testCookie, respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _pswdsGormDB.Delete(&PasswordRecord{}, "data_id=?", dataID).Error; err != nil {
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
	// 添加non password record
	dataID2 := random.NewUUIDString()
	updatedAt = time.Now().Unix()
	reqData = &struct {
		UpdatedAt int64  `json:"updatedAt"`
		DataID    string `json:"dataID"`
		Type      string `json:"type"`
		Content   string `json:"content"`
	}{
		UpdatedAt: updatedAt,
		DataID:    dataID2,
		Type:      "identity",
		Content:   "{}",
	}
	respData = &RespData{}
	resp, err = postJsonRequest(_kongDNS+"/pswds/createNonPasswordRecord/v1", reqData, _testCookie, respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _pswdsGormDB.Delete(&NonPasswordRecord{}, "data_id=?", dataID2).Error; err != nil {
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
	// 创建家庭
	reqData = &struct {
		Description        string `json:"description"`
		EncryptedFamilyKey string `json:"encryptedFamilyKey"`
	}{
		Description:        "testfamily",
		EncryptedFamilyKey: "ffffff",
	}
	respData = &RespData{}
	resp, err = postJsonRequest(_kongDNS+"/pswds/createFamily/v1", reqData, _testCookie, respData, nil)
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
	// 检查另一个账号是否可被邀请
	type CheckUserAvailableData struct {
		State         string `json:"state"`
		UserPublicKey string `json:"userPublicKey"`
	}
	const (
		InvitationState_NoUser     = "no_user"
		InvitationState_HasFamily  = "has_family"
		InvitationState_HasInvited = "has_invited"
		InvitationState_Invitable  = "invitable"
	)
	reqData = &struct {
		Email string `json:"email"`
	}{
		Email: _testEmail2,
	}
	respData = &RespData{
		Data: CheckUserAvailableData{},
	}
	resp, err = postJsonRequest(_kongDNS+"/pswds/checkUserAvailable/v1", reqData, _testCookie, respData, nil)
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
	if respData.Data.(map[string]any)["state"].(string) != InvitationState_Invitable {
		t.Error("not prospective response data")
		return
	}
	// 邀请另一个账号为家庭成员
	reqData = &struct {
		Email              string `json:"email"`
		EncryptedFamilyKey string `json:"encryptedFamilyKey"`
	}{
		Email:              _testEmail2,
		EncryptedFamilyKey: "ffffff",
	}
	respData = &RespData{}
	resp, err = postJsonRequest(_kongDNS+"/pswds/inviteFamilyMember/v1", reqData, _testCookie, respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _pswdsGormDB.Delete(&FamilyInvitation{}, "invited_by=?", _testAccount.ID).Error; err != nil {
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
	// 另一个账号检查家庭邀请
	type CheckFamilyInvitationData struct {
		HasInvitation bool   `json:"hasInvitation"`
		ID            int64  `json:"id"`
		InvitedBy     string `json:"invitedBy"`
		InvitedAt     int64  `json:"invitedAt"`
	}
	respData = &RespData{
		Data: CheckFamilyInvitationData{},
	}
	resp, err = postJsonRequest(_kongDNS+"/pswds/checkFamilyInvitation/v1", nil, _testCookie2, respData, nil)
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
	if !respData.Data.(map[string]any)["hasInvitation"].(bool) {
		t.Error("not prospective response data")
		return
	}
	// 另一个账号拒绝家庭邀请
	reqData = &struct {
		ID     int64 `json:"id"`
		Accept bool  `json:"accept"`
	}{
		ID:     int64(respData.Data.(map[string]any)["id"].(float64)),
		Accept: false,
	}
	respData = &RespData{}
	resp, err = postJsonRequest(_kongDNS+"/pswds/processFamilyInvitation/v1", reqData, _testCookie2, respData, nil)
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
	familyMember, err := daoManager.FamilyMemberDAO.GetByUserID(ctx, _testAccount2.ID)
	if err != nil {
		t.Error(err)
		return
	}
	if familyMember != nil {
		t.Error("not prospective response data code")
		return
	}
	invitation, err := daoManager.FamilyInvitationDAO.GetByEmail(ctx, _testEmail2)
	if err != nil {
		t.Error(err)
		return
	}
	if invitation != nil {
		t.Error("not prospective response data code")
		return
	}
	// 再次邀请另一个账号为家庭成员
	reqData = &struct {
		Email              string `json:"email"`
		EncryptedFamilyKey string `json:"encryptedFamilyKey"`
	}{
		Email:              _testEmail2,
		EncryptedFamilyKey: "ffffff",
	}
	respData = &RespData{}
	resp, err = postJsonRequest(_kongDNS+"/pswds/inviteFamilyMember/v1", reqData, _testCookie, respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _pswdsGormDB.Delete(&FamilyInvitation{}, "invited_by=?", _testAccount.ID).Error; err != nil {
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
	// 另一个账号再次检查家庭邀请
	respData = &RespData{
		Data: CheckFamilyInvitationData{},
	}
	resp, err = postJsonRequest(_kongDNS+"/pswds/checkFamilyInvitation/v1", nil, _testCookie2, respData, nil)
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
	if !respData.Data.(map[string]any)["hasInvitation"].(bool) {
		t.Error("not prospective response data")
		return
	}
	// 另一个账号同意家庭邀请
	reqData = &struct {
		ID     int64 `json:"id"`
		Accept bool  `json:"accept"`
	}{
		ID:     int64(respData.Data.(map[string]any)["id"].(float64)),
		Accept: true,
	}
	respData = &RespData{}
	resp, err = postJsonRequest(_kongDNS+"/pswds/processFamilyInvitation/v1", reqData, _testCookie2, respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _pswdsGormDB.Delete(&FamilyMember{}, "user_id=?", _testAccount2.ID).Error; err != nil {
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
	familyMember, err = daoManager.FamilyMemberDAO.GetByUserID(ctx, _testAccount2.ID)
	if err != nil {
		t.Error(err)
		return
	}
	if familyMember == nil {
		t.Error("not prospective response data code")
		return
	}
	invitation, err = daoManager.FamilyInvitationDAO.GetByEmail(ctx, _testEmail2)
	if err != nil {
		t.Error(err)
		return
	}
	if invitation != nil {
		t.Error("not prospective response data code")
		return
	}
	// 检查家庭成员
	type GetFamilyInfoData struct {
		HasFamily     bool   `json:"hasFamily"`
		Description   string `json:"description"`
		FamilyMembers []struct {
			ID       int64  `json:"id"`
			UserID   int64  `json:"userID"`
			Email    string `json:"email"`
			FamilyID string `json:"familyID"`
			JoinedAt int64  `json:"joinedAt"`
			IsAdmin  bool   `json:"isAdmin"`
		}
		SharedNumbers int64 `json:"sharedNumbers"`
	}
	respData = &RespData{
		Data: GetFamilyInfoData{},
	}
	resp, err = postJsonRequest(_kongDNS+"/pswds/getFamilyInfo/v1", nil, _testCookie2, respData, nil)
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
	if !respData.Data.(map[string]any)["hasFamily"].(bool) ||
		len(respData.Data.(map[string]any)["familyMembers"].([]interface{})) != 2 ||
		respData.Data.(map[string]any)["sharedNumbers"].(float64) != 0 {
		t.Error("not prospective response data")
		return
	}
	// 一个账号分享其数据
	reqData = &struct {
		SharingMembers []int64 `json:"sharingMembers"`
		DataID         string  `json:"dataID"`
		Type           string  `json:"type"`
		Content        string  `json:"content"`
	}{
		DataID:  dataID,
		Type:    "password",
		Content: "{}",
	}
	respData = &RespData{}
	resp, err = postJsonRequest(_kongDNS+"/pswds/shareDataToFamily/v1", reqData, _testCookie, respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _pswdsGormDB.Delete(&FamilySharedRecord{}, "shared_by=?", _testAccount.ID).Error; err != nil {
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
		t.Error("not prospective response data")
		return
	}
	reqData = &struct {
		SharingMembers []int64 `json:"sharingMembers"`
		DataID         string  `json:"dataID"`
		Type           string  `json:"type"`
		Content        string  `json:"content"`
	}{
		DataID:  dataID2,
		Type:    "identity",
		Content: "{}",
	}
	respData = &RespData{}
	resp, err = postJsonRequest(_kongDNS+"/pswds/shareDataToFamily/v1", reqData, _testCookie, respData, nil)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := _pswdsGormDB.Delete(&FamilySharedRecord{}, "shared_by=?", _testAccount.ID).Error; err != nil {
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
	sharedRecord2, err := daoManager.FamilySharedRecordDAO.GetByUserIDAndDataID(ctx, _testAccount.ID, dataID2)
	if err != nil {
		t.Error(err)
		return
	}
	if sharedRecord2 == nil {
		t.Error("not prospective response data")
		return
	}
	// 再次检查家庭成员
	respData = &RespData{
		Data: GetFamilyInfoData{},
	}
	resp, err = postJsonRequest(_kongDNS+"/pswds/getFamilyInfo/v1", nil, _testCookie2, respData, nil)
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
	if !respData.Data.(map[string]any)["hasFamily"].(bool) ||
		len(respData.Data.(map[string]any)["familyMembers"].([]interface{})) != 2 ||
		respData.Data.(map[string]any)["sharedNumbers"].(float64) != 2 {
		t.Error("not prospective response data")
		return
	}
	// 另一个账号检查共享数据
	reqData = &struct {
		UpdatedAt int64 `json:"updatedAt"`
	}{
		UpdatedAt: updatedAt,
	}
	type CheckBackupData struct {
		State             string `json:"state"`
		UpdatedAt         int64  `json:"updatedAt"`
		HasFamily         bool   `json:"hasFamily"`
		SharedDataUpdated bool   `json:"sharedDataUpdated"`
	}
	respData = &RespData{
		Data: CheckBackupData{},
	}
	resp, err = postJsonRequest(_kongDNS+"/pswds/checkUpdates/v1", reqData, _testCookie2, respData, nil)
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
	if !respData.Data.(map[string]any)["sharedDataUpdated"].(bool) {
		t.Error("not prospective response data")
		return
	}
	// 另一个账号下载分享数据
	type DownloadSharedDataData struct {
		SharingList []struct {
			DataID    string `json:"dataID"`
			UpdatedAt int64  `json:"updatedAt"`
			FamilyID  string `json:"familyID"`
			SharedBy  int64  `json:"sharedBy"`
			Type      string `json:"type"`
			Content   string `json:"content"`
			Version   int64  `json:"version"`
		} `json:"sharingList"`
	}
	respData = &RespData{
		Data: DownloadSharedDataData{},
	}
	resp, err = postJsonRequest(_kongDNS+"/pswds/downloadSharedData/v1", nil, _testCookie2, respData, nil)
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
	if len(respData.Data.(map[string]any)["sharingList"].([]interface{})) != 2 {
		t.Error("not prospective response data")
		return
	}
	// 一个账号停止分享数据
	reqData = &struct {
		DataID         string  `json:"dataID"`
		SharingMembers []int64 `json:"sharingMembers"`
		Stop           bool    `json:"stop"`
	}{
		DataID: dataID,
		Stop:   true,
	}
	respData = &RespData{}
	resp, err = postJsonRequest(_kongDNS+"/pswds/manageSharingData/v1", reqData, _testCookie, respData, nil)
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
	sharedRecord, err = daoManager.FamilySharedRecordDAO.GetByUserIDAndDataID(ctx, _testAccount.ID, dataID)
	if err != nil {
		t.Error(err)
		return
	}
	if sharedRecord != nil {
		t.Error("not prospective response data")
		return
	}
	reqData = &struct {
		DataID         string  `json:"dataID"`
		SharingMembers []int64 `json:"sharingMembers"`
		Stop           bool    `json:"stop"`
	}{
		DataID: dataID2,
		Stop:   true,
	}
	respData = &RespData{}
	resp, err = postJsonRequest(_kongDNS+"/pswds/manageSharingData/v1", reqData, _testCookie, respData, nil)
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
	sharedRecord2, err = daoManager.FamilySharedRecordDAO.GetByUserIDAndDataID(ctx, _testAccount2.ID, dataID2)
	if err != nil {
		t.Error(err)
		return
	}
	if sharedRecord2 != nil {
		t.Error("not prospective response data")
		return
	}
	// 另一个账号再次检查共享数据
	reqData = &struct {
		UpdatedAt int64 `json:"updatedAt"`
	}{
		UpdatedAt: updatedAt,
	}
	respData = &RespData{
		Data: CheckBackupData{},
	}
	resp, err = postJsonRequest(_kongDNS+"/pswds/checkUpdates/v1", reqData, _testCookie2, respData, nil)
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
	if !respData.Data.(map[string]any)["sharedDataUpdated"].(bool) {
		t.Error("not prospective response data")
		return
	}
	// 管理员删除家庭成员
	reqData = &struct {
		UserID int64 `json:"userID"`
	}{
		UserID: _testAccount2.ID,
	}
	respData = &RespData{}
	resp, err = postJsonRequest(_kongDNS+"/pswds/removeFamilyMember/v1", reqData, _testCookie, respData, nil)
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
	familyMember2, err := daoManager.FamilyMemberDAO.GetByUserID(ctx, _testAccount2.ID)
	if err != nil {
		t.Error(err)
		return
	}
	if familyMember2 != nil {
		t.Error("not prospective response data")
		return
	}
}
