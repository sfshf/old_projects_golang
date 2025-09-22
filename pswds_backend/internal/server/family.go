package server

import (
	"context"
	"time"

	"github.com/nextsurfer/ground/pkg/rpc"
	pswds_api "github.com/nextsurfer/pswds_backend/api"
	"github.com/nextsurfer/pswds_backend/api/response"
)

func (s *PswdsServer) DownloadSharedData(ctx context.Context, req *pswds_api.DownloadSharedDataRequest) (*pswds_api.DownloadSharedDataResponse, error) {
	startTS := time.Now()
	var resp pswds_api.DownloadSharedDataResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "DownloadSharedData", req)
	defer func() { s.deferLogResponseData(rpcCtx, "DownloadSharedData", &resp) }()
	data, appError := s.FamilyService.DownloadSharedData(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	resp.Data = data
	return &resp, nil
}

func (s *PswdsServer) CreateFamily(ctx context.Context, req *pswds_api.CreateFamilyRequest) (*pswds_api.CreateFamilyResponse, error) {
	startTS := time.Now()
	var resp pswds_api.CreateFamilyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CreateFamily", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CreateFamily", &resp) }()
	appError := s.FamilyService.CreateFamily(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *PswdsServer) GetFamilyInfo(ctx context.Context, req *pswds_api.GetFamilyInfoRequest) (*pswds_api.GetFamilyInfoResponse, error) {
	startTS := time.Now()
	var resp pswds_api.GetFamilyInfoResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetFamilyInfo", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetFamilyInfo", &resp) }()
	data, appError := s.FamilyService.GetFamilyInfo(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	resp.Data = data
	return &resp, nil
}

func (s *PswdsServer) CheckUserAvailable(ctx context.Context, req *pswds_api.CheckUserAvailableRequest) (*pswds_api.CheckUserAvailableResponse, error) {
	startTS := time.Now()
	var resp pswds_api.CheckUserAvailableResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CheckUserAvailable", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CheckUserAvailable", &resp) }()
	data, appError := s.FamilyService.CheckUserAvailable(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	resp.Data = data
	return &resp, nil
}

func (s *PswdsServer) InviteFamilyMember(ctx context.Context, req *pswds_api.InviteFamilyMembersRequest) (*pswds_api.InviteFamilyMemberResponse, error) {
	startTS := time.Now()
	var resp pswds_api.InviteFamilyMemberResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "InviteFamilyMember", req)
	defer func() { s.deferLogResponseData(rpcCtx, "InviteFamilyMember", &resp) }()
	appError := s.FamilyService.InviteFamilyMember(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *PswdsServer) CheckFamilyInvitation(ctx context.Context, req *pswds_api.CheckFamilyInvitationRequest) (*pswds_api.CheckFamilyInvitationResponse, error) {
	startTS := time.Now()
	var resp pswds_api.CheckFamilyInvitationResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "CheckFamilyInvitation", req)
	defer func() { s.deferLogResponseData(rpcCtx, "CheckFamilyInvitation", &resp) }()
	data, appError := s.FamilyService.CheckFamilyInvitation(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	resp.Data = data
	return &resp, nil
}

func (s *PswdsServer) ProcessFamilyInvitation(ctx context.Context, req *pswds_api.ProcessFamilyInvitationRequest) (*pswds_api.ProcessFamilyInvitationResponse, error) {
	startTS := time.Now()
	var resp pswds_api.ProcessFamilyInvitationResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ProcessFamilyInvitation", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ProcessFamilyInvitation", &resp) }()
	appError := s.FamilyService.ProcessFamilyInvitation(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *PswdsServer) RemoveFamilyMember(ctx context.Context, req *pswds_api.RemoveFamilyMemberRequest) (*pswds_api.RemoveFamilyMemberResponse, error) {
	startTS := time.Now()
	var resp pswds_api.RemoveFamilyMemberResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "RemoveFamilyMember", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RemoveFamilyMember", &resp) }()
	appError := s.FamilyService.RemoveFamilyMember(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *PswdsServer) LeaveFamily(ctx context.Context, req *pswds_api.LeaveFamilyRequest) (*pswds_api.LeaveFamilyResponse, error) {
	startTS := time.Now()
	var resp pswds_api.LeaveFamilyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "LeaveFamily", req)
	defer func() { s.deferLogResponseData(rpcCtx, "LeaveFamily", &resp) }()
	appError := s.FamilyService.LeaveFamily(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *PswdsServer) ShareDataToFamily(ctx context.Context, req *pswds_api.ShareDataToFamilyRequest) (*pswds_api.ShareDataToFamilyResponse, error) {
	startTS := time.Now()
	var resp pswds_api.ShareDataToFamilyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ShareDataToFamily", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ShareDataToFamily", &resp) }()
	appError := s.FamilyService.ShareDataToFamily(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *PswdsServer) ManageSharingData(ctx context.Context, req *pswds_api.ManageSharingDataRequest) (*pswds_api.ManageSharingDataResponse, error) {
	startTS := time.Now()
	var resp pswds_api.ManageSharingDataResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ManageSharingData", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ManageSharingData", &resp) }()
	appError := s.FamilyService.ManageSharingData(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *PswdsServer) HandleAdminAuthority(ctx context.Context, req *pswds_api.HandleAdminAuthorityRequest) (*pswds_api.HandleAdminAuthorityResponse, error) {
	startTS := time.Now()
	var resp pswds_api.HandleAdminAuthorityResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "HandleAdminAuthority", req)
	defer func() { s.deferLogResponseData(rpcCtx, "HandleAdminAuthority", &resp) }()
	appError := s.FamilyService.HandleAdminAuthority(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *PswdsServer) RemoveFamily(ctx context.Context, req *pswds_api.RemoveFamilyRequest) (*pswds_api.RemoveFamilyResponse, error) {
	startTS := time.Now()
	var resp pswds_api.RemoveFamilyResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "RemoveFamily", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RemoveFamily", &resp) }()
	appError := s.FamilyService.RemoveFamily(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *PswdsServer) GetFamilyMessages(ctx context.Context, req *pswds_api.GetFamilyMessagesRequest) (*pswds_api.GetFamilyMessagesResponse, error) {
	startTS := time.Now()
	var resp pswds_api.GetFamilyMessagesResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetFamilyMessages", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetFamilyMessages", &resp) }()
	data, appError := s.FamilyService.GetFamilyMessages(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	resp.Data = data
	return &resp, nil
}

func (s *PswdsServer) GetFamilyBackups(ctx context.Context, req *pswds_api.GetFamilyBackupsRequest) (*pswds_api.GetFamilyBackupsResponse, error) {
	startTS := time.Now()
	var resp pswds_api.GetFamilyBackupsResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetFamilyBackups", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetFamilyBackups", &resp) }()
	data, appError := s.FamilyService.GetFamilyBackups(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	resp.Data = data
	return &resp, nil
}

func (s *PswdsServer) SetFamilyBackup(ctx context.Context, req *pswds_api.SetFamilyBackupRequest) (*pswds_api.SetFamilyBackupResponse, error) {
	startTS := time.Now()
	var resp pswds_api.SetFamilyBackupResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "SetFamilyBackup", req)
	defer func() { s.deferLogResponseData(rpcCtx, "SetFamilyBackup", &resp) }()
	appError := s.FamilyService.SetFamilyBackup(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *PswdsServer) RequestFamilyRecover(ctx context.Context, req *pswds_api.RequestFamilyRecoverRequest) (*pswds_api.RequestFamilyRecoverResponse, error) {
	startTS := time.Now()
	var resp pswds_api.RequestFamilyRecoverResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "RequestFamilyRecover", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RequestFamilyRecover", &resp) }()
	appError := s.FamilyService.RequestFamilyRecover(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *PswdsServer) HelpFamilyRecover(ctx context.Context, req *pswds_api.HelpFamilyRecoverRequest) (*pswds_api.HelpFamilyRecoverResponse, error) {
	startTS := time.Now()
	var resp pswds_api.HelpFamilyRecoverResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "HelpFamilyRecover", req)
	defer func() { s.deferLogResponseData(rpcCtx, "HelpFamilyRecover", &resp) }()
	appError := s.FamilyService.HelpFamilyRecover(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *PswdsServer) RejectFamilyRecover(ctx context.Context, req *pswds_api.RejectFamilyRecoverRequest) (*pswds_api.RejectFamilyRecoverResponse, error) {
	startTS := time.Now()
	var resp pswds_api.RejectFamilyRecoverResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "RejectFamilyRecover", req)
	defer func() { s.deferLogResponseData(rpcCtx, "RejectFamilyRecover", &resp) }()
	appError := s.FamilyService.RejectFamilyRecover(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}

func (s *PswdsServer) GetFamilyBackupRecovers(ctx context.Context, req *pswds_api.GetFamilyBackupRecoversRequest) (*pswds_api.GetFamilyBackupRecoversResponse, error) {
	startTS := time.Now()
	var resp pswds_api.GetFamilyBackupRecoversResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "GetFamilyBackupRecovers", req)
	defer func() { s.deferLogResponseData(rpcCtx, "GetFamilyBackupRecovers", &resp) }()
	data, appError := s.FamilyService.GetFamilyBackupRecovers(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	resp.Data = data
	return &resp, nil
}

func (s *PswdsServer) ConfirmFamilyRecover(ctx context.Context, req *pswds_api.ConfirmFamilyRecoverRequest) (*pswds_api.ConfirmFamilyRecoverResponse, error) {
	startTS := time.Now()
	var resp pswds_api.ConfirmFamilyRecoverResponse
	rpcCtx := rpc.NewContext(ctx, s.localizeManager)
	s.logRequestData(rpcCtx, "ConfirmFamilyRecover", req)
	defer func() { s.deferLogResponseData(rpcCtx, "ConfirmFamilyRecover", &resp) }()
	appError := s.FamilyService.ConfirmFamilyRecover(ctx, rpcCtx, req)
	if appError != nil {
		resp.Code = appError.Code
		resp.Message = appError.Message
		resp.DebugMessage = appError.DebugMessage
		// oracle field
		resp.Oracle = s.oracleField(startTS)
		return &resp, nil
	}
	resp.Code = response.StatusCodeOK
	resp.Message = response.StatusMessageOk
	// oracle field
	resp.Oracle = s.oracleField(startTS)
	return &resp, nil
}
