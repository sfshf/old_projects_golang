package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	gerror "github.com/nextsurfer/ground/pkg/err"
	"github.com/nextsurfer/ground/pkg/rpc"
	console_api "github.com/nextsurfer/oracle/api/console"
	"github.com/nextsurfer/oracle/api/response"
	"github.com/nextsurfer/oracle/internal/common/connector"
	"github.com/nextsurfer/oracle/internal/common/simplehash"
	"github.com/nextsurfer/oracle/internal/common/simplehttp"
	"github.com/nextsurfer/oracle/internal/common/simpleproto"
	"github.com/nextsurfer/oracle/internal/dao"
	. "github.com/nextsurfer/oracle/internal/model"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
)

type ServiceService struct {
	*ConsoleService
}

func NewServiceService(ctx context.Context, consoleService *ConsoleService) *ServiceService {
	return &ServiceService{
		ConsoleService: consoleService,
	}
}

type ListServicesItem struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Application  string `json:"application"`
	URL          string `json:"url"`
	PathPrefix   string `json:"pathPrefix"`
	ProtoFileMd5 string `json:"protoFileMd5"`
	CreatedAt    int64  `json:"createdAt"`
}

type ListServicesData struct {
	List []ListServicesItem `json:"list"`
}

type ListServicesRequest struct {
	ApiKey        string `json:"apiKey" validate:"required"`
	ApplicationID int64  `json:"applicationID" validate:""`
}

func (s *ServiceService) ListServices(ctx context.Context, request any) (any, error) {
	req := request.(*ListServicesRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleRead); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	apps, err := s.DaoManager.ApplicationDAO.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	appNames := make(map[int64]string, len(apps))
	for _, app := range apps {
		appNames[app.ID] = app.Name
	}
	services, err := s.DaoManager.ServiceDAO.GetAllServices(ctx, true /*omitProtoFile*/, true /*omitFileDescriptor*/)
	if err != nil {
		return nil, err
	}
	var list []ListServicesItem
	for _, service := range services {
		if req.ApplicationID > 0 {
			if service.ApplicationID != req.ApplicationID {
				continue
			}
		}
		list = append(list, ListServicesItem{
			ID:           service.ID,
			Name:         service.Name,
			Application:  appNames[service.ApplicationID],
			URL:          service.URL,
			PathPrefix:   service.PathPrefix,
			ProtoFileMd5: service.ProtoFileMd5,
			CreatedAt:    service.CreatedAt.UnixMilli(),
		})
	}
	return &ListServicesData{List: list}, nil
}

type ListServicePathsData struct {
	List []string `json:"list"`
}

func (s *ServiceService) getServicePaths(protoFile string) []string {
	re := regexp.MustCompile(`post\s*:\s*"(.+)"`)
	matrix := re.FindAllStringSubmatch(protoFile, -1)
	var res []string
	for _, slice := range matrix {
		res = append(res, slice[1])
	}
	return res
}

type ListServicePathsRequest struct {
	ApiKey    string `json:"apiKey" validate:"required"`
	ServiceID int64  `json:"serviceID" validate:"required"`
}

func (s *ServiceService) ListServicePaths(ctx context.Context, request any) (any, error) {
	req := request.(*ListServicePathsRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleRead); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	service, err := s.DaoManager.ServiceDAO.GetByID(ctx, req.ServiceID, false /*omitProtoFile*/, true /*omitFileDescriptor*/)
	if err != nil {
		s.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, err
	}
	if service == nil {
		err = fmt.Errorf("service [id=%d] not found", req.ServiceID)
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	return &ListServicePathsData{List: s.getServicePaths(service.ProtoFile)}, nil
}

func (s *ServiceService) createService(ctx context.Context, DaoManager *dao.Manager, name, application, url, pathPrefix string, protoFile []byte, md5sum string) (*Service, error) {
	// check application
	app, err := DaoManager.ApplicationDAO.GetByName(ctx, application)
	if err != nil {
		return nil, err
	}
	// create one application, if not exists
	if app == nil {
		app = &Application{
			Name: application,
		}
		if err := DaoManager.ApplicationDAO.Create(ctx, app); err != nil {
			return nil, err
		}
	}
	// fetch file descriptor proto data
	b64fdp, err := simpleproto.Base64FileDescriptorProto("service/"+name+"/http.proto", protoFile)
	if err != nil {
		return nil, err
	}
	// create service
	service := &Service{
		Name:               name,
		ApplicationID:      app.ID,
		URL:                url,
		PathPrefix:         pathPrefix,
		ProtoFile:          string(protoFile),
		ProtoFileMd5:       md5sum,
		FileDescriptorData: b64fdp,
	}
	if err := DaoManager.ServiceDAO.Create(ctx, service); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *ServiceService) updateService(ctx context.Context, DaoManager *dao.Manager, service *Service, name, application, url, pathPrefix string, protoFile []byte, md5sum string) error {
	// check application
	app, err := DaoManager.ApplicationDAO.GetByID(ctx, service.ApplicationID)
	if err != nil {
		return err
	}
	// create one application, if not exists
	if app == nil {
		return fmt.Errorf("internal error: app [id=%d] of service [name=%s] not exists", service.ApplicationID, service.Name)
	}
	if app.Name != application {
		app = &Application{
			Name: application,
		}
		if err := DaoManager.ApplicationDAO.Create(ctx, app); err != nil {
			return err
		}
		service.ApplicationID = app.ID
	}
	// url
	if url != "" {
		service.URL = url
	}
	// path_prefix
	if pathPrefix != "" {
		service.PathPrefix = pathPrefix
	}
	// check proto file md5
	if service.ProtoFileMd5 != md5sum {
		// fetch file descriptor proto data
		b64fdp, err := simpleproto.Base64FileDescriptorProto("service/"+name+"/http.proto", protoFile)
		if err != nil {
			return err
		}
		service.ProtoFile = string(protoFile)
		service.ProtoFileMd5 = md5sum
		service.FileDescriptorData = b64fdp
	}
	// update the service
	if err = DaoManager.ServiceDAO.Update(ctx, service); err != nil {
		return err
	}
	return nil
}

func (s *ServiceService) UpsertService(ctx context.Context, rpcCtx *rpc.Context, req *console_api.UpsertServiceRequest) *gerror.AppError {
	b64ProtoFile := strings.TrimSpace(req.ProtoFile)
	protoFile, err := base64.StdEncoding.DecodeString(b64ProtoFile)
	if err != nil {
		rpcCtx.Logger.Error("bad request", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("ClientErrMsg_WrongRequestParameters")).WithCode(response.StatusCodeWrongParameters)
	}
	var service *Service
	if err := s.DaoManager.DB.Transaction(func(tx *gorm.DB) error {
		DaoManager := dao.NewManagerWithDB(tx)
		// proto file md5
		md5sum := simplehash.HexMd5ToString(protoFile)
		// check path_prefix
		if req.PathPrefix != "" {
			service, err = DaoManager.ServiceDAO.GetByPathPrefix(ctx, req.PathPrefix)
			if err != nil {
				return err
			}
			if service != nil {
				if service.Name != req.Name {
					return fmt.Errorf("service of path prefix [%s] has existed, and its name [%s] is not %s", req.PathPrefix, service.Name, req.Name)
				}
			}
		}
		if service == nil {
			// get by name
			service, err = DaoManager.ServiceDAO.GetByName(ctx, req.Name, true /*omitProtoFile*/, true /*omitDeleted*/)
			if err != nil {
				return err
			}
		}
		// create one service, if not exists
		if service == nil {
			service, err = s.createService(ctx, DaoManager, req.Name, req.Application, req.Url, req.PathPrefix, protoFile, md5sum)
			if err != nil {
				return err
			}
			return nil
		}
		// update the service
		return s.updateService(ctx, DaoManager, service, req.Name, req.Application, req.Url, req.PathPrefix, protoFile, md5sum)
	}); err != nil {
		rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
		return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
	}
	// notify gateway server to refresh service cache
	if service != nil {
		if err := s.notifyRefreshService(ctx, service.Name); err != nil {
			rpcCtx.Logger.Error("internal error", zap.NamedError("appError", err))
			return gerror.NewError(err).WithMessage(rpcCtx.Localizer.Localize("FatalErrMsg")).WithCode(response.StatusCodeInternalServerError)
		}
	}
	return nil
}

func (s *ServiceService) notifyRefreshService(ctx context.Context, name string) error {
	nodes, err := s.DaoManager.GatewayNodeDAO.GetAll(ctx)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		if err := simplehttp.NotifyGatewayRefreshService(node.Ipv4, node.RPCPort, name); err != nil {
			s.Logger.Error("notify gateway refresh service error", zap.NamedError("appError", err))
		}
	}
	return nil
}

type DeleteServiceRequest struct {
	ApiKey string `json:"apiKey" validate:"required"`
	Name   string `json:"name" validate:"required"`
}

func (s *ServiceService) DeleteService(ctx context.Context, request any) (any, error) {
	req := request.(*DeleteServiceRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleWrite); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	service, err := s.DaoManager.ServiceDAO.GetByName(ctx, req.Name, true /*omitProtoFile*/, true /*omitDeleted*/)
	if err != nil {
		return nil, err
	}
	if service == nil {
		return nil, nil
	}
	if err := s.DaoManager.DB.Transaction(func(tx *gorm.DB) error {
		DaoManager := dao.NewManagerWithDB(tx)
		if err := DaoManager.ServiceDAO.DeleteByName(ctx, req.Name); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	// notify gateway server to refresh service cache
	if err := s.notifyRefreshService(ctx, service.Name); err != nil {
		return nil, err
	}
	return nil, nil
}

type ListTimeoutStatisticsRequest struct {
	ApiKey        string `json:"apiKey"`
	PageSize      int    `json:"pageSize"`
	PageNumber    int    `json:"pageNumber"`
	ApplicationID int64  `json:"applicationID"`
	ServiceID     int64  `json:"serviceID"`
	Path          string `json:"path"`
	Date          string `json:"date"`
}

type ListTimeoutStatisticsItem struct {
	ID          int64     `json:"id"`
	Application string    `json:"application"`
	Service     string    `json:"service"`
	Path        string    `json:"path"`
	Count       int64     `json:"count"`
	Date        time.Time `json:"date"`
}

type ListTimeoutStatisticsData struct {
	Total int64                       `json:"total"`
	List  []ListTimeoutStatisticsItem `json:"list"`
}

func (s *ServiceService) generateConditions(applicationID, serviceID int64, path, date string) (map[string]interface{}, error) {
	conditions := make(map[string]interface{})
	if applicationID > 0 {
		conditions["application_id = ?"] = applicationID
	}
	if serviceID > 0 {
		conditions["service_id = ?"] = serviceID
	}
	if path != "" {
		conditions["path = ?"] = path
	}
	if date != "" {
		dt, err := time.Parse("2006-01-02", date)
		if err != nil {
			s.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, err
		}
		conditions["date = ?"] = dt
	}
	return conditions, nil
}

func (s *ServiceService) ListTimeoutStatistics(ctx context.Context, request any) (any, error) {
	req := request.(*ListTimeoutStatisticsRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleRead); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	conditions, err := s.generateConditions(req.ApplicationID, req.ServiceID, req.Path, req.Date)
	if err != nil {
		return nil, err
	}
	records, total, err := s.DaoManager.TimeoutStatisticDAO.GetPaginationByConditions(ctx, conditions, req.PageSize, req.PageNumber)
	if err != nil {
		s.Logger.Error("internal error", zap.NamedError("appError", err))
		return nil, err
	}
	apps, err := s.DaoManager.ApplicationDAO.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	appNames := make(map[int64]string, len(apps))
	for _, app := range apps {
		appNames[app.ID] = app.Name
	}
	services, err := s.DaoManager.ServiceDAO.GetAllServices(ctx, true /*omitProtoFile*/, true /*omitFileDescriptor*/)
	if err != nil {
		return nil, err
	}
	serviceNames := make(map[int64]string, len(services))
	for _, service := range services {
		serviceNames[service.ID] = service.Name
	}
	var list []ListTimeoutStatisticsItem
	for _, record := range records {
		list = append(list, ListTimeoutStatisticsItem{
			ID:          record.ID,
			Application: appNames[record.ApplicationID],
			Service:     serviceNames[record.ServiceID],
			Path:        record.Path,
			Count:       record.Count,
			Date:        record.Date,
		})
	}
	return &ListTimeoutStatisticsData{List: list, Total: total}, nil
}
