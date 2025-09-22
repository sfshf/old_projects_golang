package service

import (
	"context"

	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/oracle/api/response"
	"github.com/nextsurfer/oracle/internal/common/connector"
	. "github.com/nextsurfer/oracle/internal/model"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type GatewayNodeService struct {
	*ConsoleService
}

func NewGatewayNodeService(ctx context.Context, consoleService *ConsoleService) *GatewayNodeService {
	return &GatewayNodeService{
		ConsoleService: consoleService,
	}
}

func (s *GatewayNodeService) RegisterGatewayNode(ctx context.Context, rpcCtx *rpc.Context, name, ipv4 string, rpcPort int32) *gerror.AppError {
	node := &GatewayNode{
		Name:    name,
		Ipv4:    ipv4,
		RPCPort: rpcPort,
	}
	if err := s.DaoManager.GatewayNodeDAO.Create(ctx, node); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	return nil
}

type ListGatewayNodesItem struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Ipv4      string `json:"domain"`
	RpcPort   int32  `json:"rpcPort"`
	CreatedAt int64  `json:"createdAt"`
}

type ListGatewayNodesData struct {
	List []ListGatewayNodesItem `json:"list"`
}

type ListGatewayNodesRequest struct {
	ApiKey string `json:"apiKey" validate:"required"`
}

func (s *GatewayNodeService) ListGatewayNodes(ctx context.Context, request any) (any, error) {
	req := request.(*ListGatewayNodesRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleRead); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	nodes, err := s.DaoManager.GatewayNodeDAO.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	var list []ListGatewayNodesItem
	for _, node := range nodes {
		list = append(list, ListGatewayNodesItem{
			ID:        node.ID,
			Name:      node.Name,
			Ipv4:      node.Ipv4,
			RpcPort:   node.RPCPort,
			CreatedAt: node.CreatedAt.UnixMilli(),
		})
	}
	return &ListGatewayNodesData{List: list}, nil
}
