package service

import (
	"context"
	"errors"
	"time"

	expirableCache "github.com/go-pkgz/expirable-cache/v3"
	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	pswds_api "github.com/nextsurfer/pswds_backend/api"
	"github.com/nextsurfer/pswds_backend/api/response"
	"github.com/nextsurfer/pswds_backend/internal/common/random"
	"go.uber.org/zap"
)

type ShareService struct {
	*PswdsService

	PswdsUUIDCache expirableCache.Cache[string, *PswdsValue]
}

type PswdsValue struct {
	CipherText string
	Options    string
}

func NewShareService(ctx context.Context, pswdsService *PswdsService) *ShareService {
	return &ShareService{
		PswdsService:   pswdsService,
		PswdsUUIDCache: expirableCache.NewCache[string, *PswdsValue](),
	}
}

func (s *ShareService) GetAirdropID(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.GetAirdropIDRequest) (*pswds_api.GetAirdropIDResponse_Data, *gerror.AppError) {
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	one := random.NewUUIDHexEncoding()
	s.PswdsUUIDCache.Set(one, &PswdsValue{}, time.Minute*5)
	return &pswds_api.GetAirdropIDResponse_Data{Uuid: one}, nil
}

func (s *ShareService) RequestAirdropData(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.RequestAirdropDataRequest) (*pswds_api.RequestAirdropDataResponse_Data, *gerror.AppError) {
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return nil, appError
	}
	result, has := s.PswdsUUIDCache.Get(req.Uuid)
	if !has {
		err := errors.New("no pswds uuid record")
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	return &pswds_api.RequestAirdropDataResponse_Data{CipherText: result.CipherText, Options: result.Options}, nil
}

func (s *ShareService) UploadAirdropData(ctx context.Context, rpcCtx *rpc.Context, req *pswds_api.UploadAirdropDataRequest) *gerror.AppError {
	// validate request basically
	if appError := s.ValidateRequest(ctx, rpcCtx, req); appError != nil {
		return appError
	}
	result, has := s.PswdsUUIDCache.Get(req.Uuid)
	if !has {
		err := errors.New("no pswds uuid record")
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_ResourceNotFound")).WithCode(response.StatusCodeNotFound)
	}
	result.CipherText = req.CipherText
	result.Options = req.Options
	s.PswdsUUIDCache.Set(req.Uuid, result, time.Minute*10)
	return nil
}
