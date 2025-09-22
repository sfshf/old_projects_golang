package service

import (
	"context"
	"fmt"
	"time"

	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/oracle/internal/common/connector"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type ProtocolService struct {
	*ConsoleService
}

func NewProtocolService(ctx context.Context, consoleService *ConsoleService) *ProtocolService {
	return &ProtocolService{
		ConsoleService: consoleService,
	}
}

type ListProtoStatisticsItem struct {
	Application        string `json:"application"`
	Service            string `json:"service"`
	Path               string `json:"path"`
	Hit                int64  `json:"hit"`
	SuccessRate        int    `json:"successRate"`
	ProxySuccessRate   int    `json:"proxySuccessRate"`
	DurationAvg        string `json:"durationAvg"`
	DurationMin        string `json:"durationMin"`
	DurationMax        string `json:"durationMax"`
	ServiceDurationAvg string `json:"serviceDurationAvg"`
	ServiceDurationMin string `json:"serviceDurationMin"`
	ServiceDurationMax string `json:"serviceDurationMax"`
}

type ListProtoStatisticsData struct {
	Total int64                     `json:"total"`
	List  []ListProtoStatisticsItem `json:"list"`
}

func (s *ProtocolService) generateConditions(applicationID, serviceID int64, path, date string, latestWeek, latestMonth bool) (map[string]interface{}, error) {
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
	beijing, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil, err
	}
	if latestWeek {
		today := time.Now().In(beijing)
		startDay := today.AddDate(0, 0, -7)
		conditions["date > ?"] = startDay
		conditions["date <= ?"] = today
		return conditions, nil
	}
	if latestMonth {
		today := time.Now().In(beijing)
		startDay := today.AddDate(0, -1, 0)
		conditions["date > ?"] = startDay
		conditions["date <= ?"] = today
		return conditions, nil
	}
	if date != "" {
		dt, err := time.Parse("2006-01-02", date)
		if err != nil {
			s.Logger.Error("internal error", zap.NamedError("appError", err))
			return nil, err
		}
		conditions["date = ?"] = dt
		return conditions, nil
	}
	return conditions, nil
}

type ListProtoStatisticsRequest struct {
	ApiKey        string `json:"apiKey"`
	PageSize      int    `json:"pageSize"`
	PageNumber    int    `json:"pageNumber"`
	ApplicationID int64  `json:"applicationID"`
	ServiceID     int64  `json:"serviceID"`
	Path          string `json:"path"`
	Date          string `json:"date"`
	LatestWeek    bool   `json:"latestWeek"`
	LatestMonth   bool   `json:"latestMonth"`
}

func (s *ProtocolService) ListProtoStatistics(ctx context.Context, request any) (any, error) {
	req := request.(*ListProtoStatisticsRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleRead); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	conditions, err := s.generateConditions(req.ApplicationID, req.ServiceID, req.Path, req.Date, req.LatestWeek, req.LatestMonth)
	if err != nil {
		return nil, err
	}
	records, total, err := s.DaoManager.ProtoStatisticDAO.GetPaginationByConditions(ctx, conditions, req.PageSize, req.PageNumber, req.LatestWeek || req.LatestMonth)
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
	var list []ListProtoStatisticsItem
	for _, record := range records {
		one := ListProtoStatisticsItem{
			Path:               record.Path,
			Application:        appNames[record.ApplicationID],
			Service:            serviceNames[record.ServiceID],
			Hit:                record.Hit,
			DurationMin:        fmt.Sprintf("%dms", record.DurationMin),
			DurationMax:        fmt.Sprintf("%dms", record.DurationMax),
			ServiceDurationMin: fmt.Sprintf("%dms", record.ServiceDurationMin),
			ServiceDurationMax: fmt.Sprintf("%dms", record.ServiceDurationMax),
		}
		// 零值找补
		if record.DurationAverage == 0 {
			one.DurationAvg = "1ms"
		} else {
			one.DurationAvg = fmt.Sprintf("%dms", record.DurationAverage)
		}
		if record.ServiceDurationAverage == 0 {
			one.ServiceDurationAvg = "1ms"
		} else {
			one.ServiceDurationAvg = fmt.Sprintf("%dms", record.ServiceDurationAverage)
		}
		if record.Hit != 0 {
			one.SuccessRate = int(float64(record.SuccessHit) / float64(record.Hit) * 100)
			one.ProxySuccessRate = int(float64(record.ProxySuccessHit) / float64(record.Hit) * 100)
		}
		list = append(list, one)
	}
	return &ListProtoStatisticsData{List: list, Total: total}, nil
}
