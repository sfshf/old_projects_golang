package service

import (
	"context"

	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/oracle/internal/common/connector"
	. "github.com/nextsurfer/oracle/internal/model"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type RateLimitService struct {
	*ConsoleService
}

func NewRateLimitService(ctx context.Context, consoleService *ConsoleService) *RateLimitService {
	return &RateLimitService{
		ConsoleService: consoleService,
	}
}

type ListRateLimitRulesItem struct {
	ID        int64  `json:"id"`
	Type      int32  `json:"type"`
	Target    string `json:"target"`
	Capacity  int64  `json:"capacity"`
	Enabled   bool   `json:"enabled"`
	CreatedAt int64  `json:"createdAt"`
}

type ListRateLimitRulesData struct {
	List []*ListRateLimitRulesItem `json:"list"`
}

type ListRateLimitRulesRequest struct {
	ApiKey string `json:"apiKey" validate:"required"`
}

func (s *RateLimitService) ListRateLimitRules(ctx context.Context, request any) (any, error) {
	req := request.(*ListRateLimitRulesRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleRead); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	rateLimitRules, err := s.DaoManager.RateLimitRuleDAO.GetAll(ctx)
	if err != nil {
		s.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, err
	}
	var res []*ListRateLimitRulesItem
	for _, elem := range rateLimitRules {
		res = append(res, &ListRateLimitRulesItem{
			ID:        elem.ID,
			Type:      elem.Type,
			Target:    elem.Target,
			Capacity:  elem.Capacity,
			Enabled:   elem.Enabled,
			CreatedAt: elem.CreatedAt.UnixMilli(),
		})
	}
	return &ListRateLimitRulesData{List: res}, nil
}

type AddRateLimitRuleRequest struct {
	ApiKey   string `json:"apiKey" validate:"required"`
	Type     int32  `json:"type" validate:"required"`
	Target   string `json:"target" validate:"required"`
	Capacity int64  `json:"capacity" validate:"required"`
	Enabled  bool   `json:"enabled" validate:""`
}

func (s *RateLimitService) AddRateLimitRule(ctx context.Context, request any) (any, error) {
	req := request.(*AddRateLimitRuleRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleWrite); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	if err := s.DaoManager.RateLimitRuleDAO.Create(ctx, &RateLimitRule{
		Type:     req.Type,
		Target:   req.Target,
		Capacity: req.Capacity,
		Enabled:  req.Enabled,
	}); err != nil {
		s.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, err
	}
	return nil, nil
}

type DeleteRateLimitRuleRequest struct {
	ApiKey string `json:"apiKey" validate:"required"`
	ID     int64  `json:"id" validate:"required"`
}

func (s *RateLimitService) DeleteRateLimitRule(ctx context.Context, request any) (any, error) {
	req := request.(*DeleteRateLimitRuleRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleWrite); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	if err := s.DaoManager.RateLimitRuleDAO.DeleteByID(ctx, req.ID); err != nil {
		s.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, err
	}
	return nil, nil
}

type UpdateRateLimitRuleRequest struct {
	ApiKey   string `json:"apiKey" validate:"required"`
	ID       int64  `json:"id" validate:"required"`
	Type     int32  `json:"type" validate:"required"`
	Target   string `json:"target" validate:"required"`
	Capacity int64  `json:"capacity" validate:"required"`
	Enabled  bool   `json:"enabled" validate:""`
}

func (s *RateLimitService) UpdateRateLimitRule(ctx context.Context, request any) (any, error) {
	req := request.(*UpdateRateLimitRuleRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleWrite); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	if err := s.DaoManager.RateLimitRuleDAO.Update(ctx, &RateLimitRule{
		ID:       req.ID,
		Type:     req.Type,
		Target:   req.Target,
		Capacity: req.Capacity,
		Enabled:  req.Enabled,
	}); err != nil {
		s.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, err
	}
	return nil, nil
}
