package service

import (
	"context"
	"fmt"
	"regexp"

	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/oracle/internal/common/connector"
	. "github.com/nextsurfer/oracle/internal/model"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type AlarmEmailService struct {
	*ConsoleService
}

func NewAlarmEmailService(ctx context.Context, consoleService *ConsoleService) *AlarmEmailService {
	return &AlarmEmailService{
		ConsoleService: consoleService,
	}
}

type ListAdminEmailsItem struct {
	ID        int64  `json:"id"`
	Address   string `json:"address"`
	CreatedAt int64  `json:"createdAt"`
}

type ListAdminEmailsData struct {
	List []*ListAdminEmailsItem `json:"list"`
}

type ListAlarmEmailsRequest struct {
	ApiKey string `json:"apiKey" validate:"required"`
}

func (s *AlarmEmailService) ListAlarmEmails(ctx context.Context, request any) (any, error) {
	req := request.(*ListAlarmEmailsRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleRead); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	alermEmails, err := s.DaoManager.AlarmEmailDAO.GetAll(ctx)
	if err != nil {
		s.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, err
	}
	var res []*ListAdminEmailsItem
	for _, elem := range alermEmails {
		res = append(res, &ListAdminEmailsItem{
			ID:        elem.ID,
			Address:   elem.Address,
			CreatedAt: elem.CreatedAt.UnixMilli(),
		})
	}
	return &ListAdminEmailsData{List: res}, nil
}

type AddAlarmEmailRequest struct {
	ApiKey  string `json:"apiKey" validate:"required"`
	Address string `json:"address" validate:"required"`
}

func (s *AlarmEmailService) AddAlarmEmail(ctx context.Context, request any) (any, error) {
	req := request.(*AddAlarmEmailRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleWrite); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	// validate email parameter
	matched, err := regexp.MatchString(`[\w]+(\.[\w]+)*@[\w]+(\.[\w])+`, req.Address)
	if err != nil {
		s.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, err
	} else if !matched {
		err = fmt.Errorf("deformed email [%s]", req.Address)
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	if err := s.DaoManager.AlarmEmailDAO.Create(ctx, &AlarmEmail{
		Address: req.Address,
	}); err != nil {
		s.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, err
	}
	return nil, nil
}

type DeleteAlarmEmailRequest struct {
	ApiKey string `json:"apiKey" validate:"required"`
	ID     int64  `json:"id" validate:"required"`
}

func (s *AlarmEmailService) DeleteAlarmEmail(ctx context.Context, request any) (any, error) {
	req := request.(*DeleteAlarmEmailRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleWrite); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	if err := s.DaoManager.AlarmEmailDAO.DeleteByID(ctx, req.ID); err != nil {
		s.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, err
	}
	return nil, nil
}

type UpdateAlarmEmailRequest struct {
	ApiKey  string `json:"apiKey" validate:"required"`
	ID      int64  `json:"id" validate:"required"`
	Address string `json:"address" validate:"required"`
}

func (s *AlarmEmailService) UpdateAlarmEmail(ctx context.Context, request any) (any, error) {
	req := request.(*UpdateAlarmEmailRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleWrite); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	// validate email parameter
	matched, err := regexp.MatchString(`[\w]+(\.[\w]+)*@[\w]+(\.[\w])+`, req.Address)
	if err != nil {
		s.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, err
	} else if !matched {
		err = fmt.Errorf("deformed email [%s]", req.Address)
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	if err := s.DaoManager.AlarmEmailDAO.Update(ctx, &AlarmEmail{ID: req.ID, Address: req.Address}); err != nil {
		s.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, err
	}
	return nil, nil
}
