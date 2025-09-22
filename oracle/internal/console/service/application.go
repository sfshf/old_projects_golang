package service

import (
	"context"
	"errors"

	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/oracle/internal/common/connector"
	"github.com/nextsurfer/oracle/internal/dao"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
)

type ApplicationService struct {
	*ConsoleService
}

func NewApplicationService(ctx context.Context, consoleService *ConsoleService) *ApplicationService {
	return &ApplicationService{
		ConsoleService: consoleService,
	}
}

type ListApplicationsItem struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"createdAt"`
}

type ListApplicationsData struct {
	List []ListApplicationsItem `json:"list"`
}

type ListApplicationsRequest struct {
	ApiKey string `json:"apiKey" validate:"required"`
}

func (s *ApplicationService) ListApplications(ctx context.Context, request any) (any, error) {
	req := request.(*ListApplicationsRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleRead); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	apps, err := s.DaoManager.ApplicationDAO.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	var list []ListApplicationsItem
	for _, app := range apps {
		list = append(list, ListApplicationsItem{
			ID:        app.ID,
			Name:      app.Name,
			CreatedAt: app.CreatedAt.UnixMilli(),
		})
	}
	return &ListApplicationsData{List: list}, nil
}

type DeleteApplicationRequest struct {
	ApiKey string `json:"apiKey" validate:"required"`
	Name   string `json:"name" validate:"required"`
}

func (s *ApplicationService) DeleteApplication(ctx context.Context, request any) (any, error) {
	req := request.(*DeleteApplicationRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleWrite); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	app, err := s.DaoManager.ApplicationDAO.GetByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, nil
	}
	serviceCount, err := s.DaoManager.ServiceDAO.CountByApplicationID(ctx, app.ID)
	if err != nil {
		return nil, err
	}
	if serviceCount > 0 {
		return nil, errors.New("remove all service of the application first")
	}
	if err := s.DaoManager.DB.Transaction(func(tx *gorm.DB) error {
		DaoManager := dao.NewManagerWithDB(tx)
		if err := DaoManager.ApplicationDAO.DeleteByName(ctx, req.Name); err != nil {
			return err
		}
		if err := DaoManager.ServiceDAO.DeleteAllByApplicationID(ctx, app.ID); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return nil, nil
}
