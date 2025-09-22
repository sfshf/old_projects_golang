package service

import (
	"context"
	"os"

	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/oracle/internal/common/connector"
	"github.com/nextsurfer/oracle/internal/common/simplehttp"
	. "github.com/nextsurfer/oracle/internal/model"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type HostnameService struct {
	*ConsoleService
}

func NewHostnameService(ctx context.Context, consoleService *ConsoleService) *HostnameService {
	return &HostnameService{
		ConsoleService: consoleService,
	}
}

func (s *HostnameService) notifyRefreshProxy(ctx context.Context, domain string) error {
	nodes, err := s.DaoManager.GatewayNodeDAO.GetAll(ctx)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		if err := simplehttp.NotifyGatewayRefreshProxy(node.Ipv4, node.RPCPort, domain); err != nil {
			s.Logger.Error("notify gateway refresh proxy error", zap.NamedError("appError", err))
		}
	}
	return nil
}

type CreateHostnameRequest struct {
	ApiKey string `json:"apiKey" validate:"required"`
	Domain string `json:"domain" validate:"required"`
	RawURL string `json:"rawURL" validate:""`
}

func (s *HostnameService) CreateHostname(ctx context.Context, request any) (any, error) {
	req := request.(*CreateHostnameRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleWrite); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	if err := s.DaoManager.HostManageDAO.Create(ctx, &HostManage{
		Domain: req.Domain,
		RawURL: req.RawURL,
	}); err != nil {
		s.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, err
	}
	// notify gateways to RefreshProxy
	return nil, s.notifyRefreshProxy(ctx, req.Domain)
}

type UpdateHostnameRequest struct {
	ApiKey string `json:"apiKey" validate:"required"`
	ID     int64  `json:"id" validate:"required"`
	RawURL string `json:"rawURL" validate:""`
}

func (s *HostnameService) UpdateHostname(ctx context.Context, request any) (any, error) {
	req := request.(*UpdateHostnameRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleWrite); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	hostname, err := s.DaoManager.HostManageDAO.GetByID(ctx, req.ID)
	if err != nil {
		s.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, err
	}
	if hostname.Domain == os.Getenv("GATEWAY_API_HOSTNAME") {
		return nil, nil
	}
	if err := s.DaoManager.HostManageDAO.Update(ctx, &HostManage{
		ID:     req.ID,
		RawURL: req.RawURL,
	}); err != nil {
		s.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, err
	}
	// notify gateways to RefreshProxy
	return nil, s.notifyRefreshProxy(ctx, hostname.Domain)
}

type ListHostnamesItem struct {
	ID        int64  `json:"id"`
	Domain    string `json:"domain"`
	RawURL    string `json:"rawURL"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type ListHostnamesResponse struct {
	List []*ListHostnamesItem `json:"list"`
}

type ListHostnamesRequest struct {
	ApiKey string `json:"apiKey" validate:"required"`
}

func (s *HostnameService) ListHostnames(ctx context.Context, request any) (any, error) {
	req := request.(*ListHostnamesRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleRead); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	hostnames, err := s.DaoManager.HostManageDAO.GetAll(ctx)
	if err != nil {
		s.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, err
	}
	for idx, item := range hostnames {
		if item.Domain == os.Getenv("GATEWAY_API_HOSTNAME") {
			hostnames[0], hostnames[idx] = hostnames[idx], hostnames[0]
			break
		}
	}
	var list []*ListHostnamesItem
	for _, item := range hostnames {
		list = append(list, &ListHostnamesItem{
			ID:        item.ID,
			Domain:    item.Domain,
			RawURL:    item.RawURL,
			CreatedAt: item.CreatedAt.UnixMilli(),
			UpdatedAt: item.UpdatedAt.UnixMilli(),
		})
	}
	return &ListHostnamesResponse{List: list}, nil
}

type DeleteHostnameRequest struct {
	ApiKey string `json:"apiKey" validate:"required"`
	ID     int64  `json:"id" validate:"required"`
}

func (s *HostnameService) DeleteHostname(ctx context.Context, request any) (any, error) {
	req := request.(*DeleteHostnameRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleWrite); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	hostname, err := s.DaoManager.HostManageDAO.GetByID(ctx, req.ID)
	if err != nil {
		s.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, err
	}
	if hostname.Domain == os.Getenv("GATEWAY_API_HOSTNAME") {
		return nil, nil
	}
	if err := s.DaoManager.HostManageDAO.DeleteByID(ctx, req.ID); err != nil {
		s.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, err
	}
	// notify gateways to RefreshProxy
	return nil, s.notifyRefreshProxy(ctx, hostname.Domain)
}
