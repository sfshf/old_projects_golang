package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-acme/lego/certificate"
	"github.com/nextsurfer/ground/pkg/rpc"
	"github.com/nextsurfer/oracle/internal/common/acme"
	"github.com/nextsurfer/oracle/internal/common/connector"
	"github.com/nextsurfer/oracle/internal/common/simpleemail"
	"github.com/nextsurfer/oracle/internal/common/simplehttp"
	"github.com/nextsurfer/oracle/internal/common/statistic"
	. "github.com/nextsurfer/oracle/internal/model"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type CronService struct {
	*ConsoleService

	Cron         *cron.Cron
	CronEntries  []*CronEntryStatus
	HealthChecks map[string]time.Time
}

type CronEntryStatus struct {
	Name              string
	EntryID           cron.EntryID
	Started           bool
	StartedOrStopedAt time.Time
	ScheduleSpec      string
	LastExecError     string
}

func NewCronService(ctx context.Context, consoleService *ConsoleService) (*CronService, error) {
	s := &CronService{
		ConsoleService: consoleService,
		HealthChecks:   make(map[string]time.Time),
	}
	// cron jobs
	beijing, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil, err
	}
	crontab := cron.New(cron.WithLocation(beijing))
	s.Cron = crontab
	var cronEntries []*CronEntryStatus
	refreshTlsCertificateEntry, err := s.RefreshTlsCertificateEntry(ctx, crontab)
	if err != nil {
		return nil, err
	}
	statisticGatewayEntry, err := s.StatisticGatewayEntry(ctx, crontab, beijing)
	if err != nil {
		return nil, err
	}
	serviceHealthCheckEntry, err := s.ServiceHealthCheckEntry(ctx, crontab)
	if err != nil {
		return nil, err
	}
	cronEntries = append(cronEntries,
		refreshTlsCertificateEntry,
		statisticGatewayEntry,
		serviceHealthCheckEntry,
	)
	s.CronEntries = cronEntries
	// start cron jobs
	s.Cron.Start()
	return s, nil
}

// cron jobs -----------------------------------------------------------------------------------------

// refresh tls certificate job
func (s *CronService) RefreshTlsCertificateEntry(ctx context.Context, crontab *cron.Cron) (*CronEntryStatus, error) {
	spec := "@daily"
	if cronSpec := os.Getenv("REFRESH_TLS_CERTIFICATE_CRON_SPEC"); cronSpec != "" {
		spec = cronSpec
	}
	entryStatus := &CronEntryStatus{
		Name:              "RefreshTlsCertificate",
		Started:           true,
		StartedOrStopedAt: time.Now(),
		ScheduleSpec:      spec,
	}
	entryID, err := crontab.AddFunc(entryStatus.ScheduleSpec, func() {
		s.Logger.Info("Cron Job -- refresh tls certificate", zap.Time("timestamp", time.Now()))
		if err := s.handleRefreshTlsCertificate(ctx); err != nil {
			entryStatus.LastExecError = err.Error()
			s.Logger.Error("cron error", zap.NamedError("appError", err))
		}
	})
	if err != nil {
		return nil, err
	}
	entryStatus.EntryID = entryID
	return entryStatus, nil
}

func (s *CronService) handleRefreshTlsCertificate(ctx context.Context) error {
	// first, fetch from db
	acmeResources, err := s.DaoManager.AcmeResourceDAO.GetAll(ctx)
	if err != nil {
		return err
	}
	alarmEmails, err := s.DaoManager.AlarmEmailDAO.GetAll(ctx)
	if err != nil {
		return err
	}
	s.iterateAcmeResources(ctx, acmeResources, alarmEmails)
	return nil
}

func (s *CronService) iterateAcmeResources(ctx context.Context, acmeResources []*AcmeResource, alarmEmails []*AlarmEmail) {
	var emailMsgs []string
	var emailMsg string
	http01Provider := acme.NewHttp01Provider(s.DaoManager)
	for _, acmeResource := range acmeResources {
		// send email message, if has
		if emailMsg != "" {
			emailMsgs = append(emailMsgs, emailMsg)
		}
		// reset email msg
		emailMsg = ""
		// not expired
		now := time.Now()
		if updatedAt := acmeResource.UpdatedAt.AddDate(0, 0, 90-15); updatedAt.After(now) {
			continue
		}
		var resource *certificate.Resource
		var err error
		if acmeResource.Certificate == "" {
			if acmeResource.Domain == "" {
				continue
			}
			resource, err = acme.NewAcmeResource(acmeResource.Domain, http01Provider)
			if err != nil {
				emailMsg = fmt.Sprintf("NewAcmeResource tls certificate of domain [%s] failed: %s", acmeResource.Domain, err)
				continue
			}
		} else {
			// second, renew certificates, if has expired
			resource, err = acme.RenewAcmeResource(acmeResource, http01Provider)
			if err != nil {
				emailMsg = fmt.Sprintf("RenewAcmeResource tls certificate of domain [%s] failed: %s", acmeResource.Domain, err)
				continue
			}
		}
		acmeResource.CertURL = resource.CertURL
		acmeResource.CertStableURL = resource.CertStableURL
		acmeResource.PrivateKey = string(resource.PrivateKey)
		acmeResource.Certificate = string(resource.Certificate)
		acmeResource.IssuerCertificate = string(resource.IssuerCertificate)
		acmeResource.Csr = string(resource.CSR)
		// third, update the record in db
		if err := s.DaoManager.AcmeResourceDAO.Update(ctx, acmeResource); err != nil {
			emailMsg = fmt.Sprintf("Update new tls certificate of domain [%s] to db failed: %s", acmeResource.Domain, err)
			continue
		}
		emailMsg = fmt.Sprintf("Refresh tls certificate of domain [%s] success", acmeResource.Domain)
		// forth, notification gateway server to use new tls certificate
		if err := s.notifyRefreshCertificate(ctx, acmeResource.Domain); err != nil {
			emailMsg += fmt.Sprintf(", but notify gateways to RefreshCertificate of domain [%s] failed: %s", acmeResource.Domain, err)
			continue
		}
	}
	// send email message, if has
	if len(emailMsgs) > 0 {
		emailMsg := strings.Join(emailMsgs, "!!! ")
		for _, email := range alarmEmails {
			if err := simpleemail.SendCronJobNotificationEmail(ctx, email.Address, emailMsg); err != nil {
				s.Logger.Error("send refresh certificates email failed", zap.NamedError("appError", err))
				continue
			}
		}
	}
}

func (s *CronService) notifyRefreshCertificate(ctx context.Context, domain string) error {
	nodes, err := s.DaoManager.GatewayNodeDAO.GetAll(ctx)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		if err := simplehttp.NotifyGatewayRefreshCertificate(node.Ipv4, node.RPCPort, domain); err != nil {
			s.Logger.Error("notify gateway refresh certificate error", zap.NamedError("appError", err))
		}
	}
	return nil
}

// statistic gateway job
func (s *CronService) StatisticGatewayEntry(ctx context.Context, crontab *cron.Cron, timezone *time.Location) (*CronEntryStatus, error) {
	spec := "@every 3h"
	if cronSpec := os.Getenv("PROTO_STATISTIC_CRON_SPEC"); cronSpec != "" {
		spec = cronSpec
	}
	entryStatus := &CronEntryStatus{
		Name:              "StatisticGateway",
		Started:           true,
		StartedOrStopedAt: time.Now(),
		ScheduleSpec:      spec,
	}
	entryID, err := crontab.AddFunc(entryStatus.ScheduleSpec, func() {
		s.Logger.Info("Cron Job -- statistic gateway", zap.Time("timestamp", time.Now()))
		s.handleStatisticGateway(ctx, timezone)
	})
	if err != nil {
		return nil, err
	}
	entryStatus.EntryID = entryID
	return entryStatus, nil
}

func (s *CronService) handleStatisticGateway(ctx context.Context, timezone *time.Location) error {
	// fetch statistic info from redis
	var cursor uint64
	var hasKeys bool
	for {
		var keys []string
		var err error
		// scan keys
		keys, cursor, err = s.RedisClient.Scan(ctx, cursor, statistic.RedisKeyPrefixStatisticInfo+"*", 10).Result()
		if err != nil {
			s.Logger.Error("scan statistic keys from redis", zap.NamedError("appError", err))
			return err
		}
		// handle keys
		for _, key := range keys {
			hourlyProtoStatistics, err := s.generateHourlyStatistics(ctx, key)
			if err != nil {
				continue
			}
			if err := s.createOrUpdateProtoStatistic(ctx, hourlyProtoStatistics, timezone); err != nil {
				continue
			}
		}
		if len(keys) > 0 {
			if err := s.RedisClient.Del(ctx, keys...).Err(); err != nil {
				s.Logger.Error("delete statistic keys from redis", zap.NamedError("appError", err))
			}
			hasKeys = true
		}
		if cursor == 0 {
			if !hasKeys {
				if err := s.createOrUpdateProtoStatistic(ctx, nil, timezone); err != nil {
					s.Logger.Error("createOrUpdateProtoStatistic fail", zap.NamedError("appError", err))
				}
			}
			break
		}
	}
	// delete data that over 120 hours
	now := time.Now().In(timezone).Add(-120 * time.Hour)
	if err := s.DaoManager.ProtoStatisticHourlyDAO.Delete(ctx, "timestamp<?", now.UnixNano()); err != nil {
		s.Logger.Error("delete data that over 120 hours", zap.NamedError("appError", err))
		return err
	}
	return nil
}

func (s *CronService) generateHourlyStatistics(ctx context.Context, key string) ([]*ProtoStatisticHourly, error) {
	var gatewayNode string
	if splits := strings.Split(key, "::"); len(splits) == 3 {
		gatewayNode = splits[1]
	}
	list, err := s.RedisClient.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		s.Logger.Error("get statistic info from redis", zap.NamedError("appError", err))
		return nil, err
	}
	var res []*ProtoStatisticHourly
	for _, val := range list {
		var protoStatistic statistic.ProtoStatisticHourly
		if err := json.Unmarshal([]byte(val), &protoStatistic); err != nil {
			s.Logger.Error("unmarshal proto statistic", zap.NamedError("appError", err))
			return nil, err
		}
		// generate ProtoStatisticHourly
		one := &ProtoStatisticHourly{
			Timestamp:              protoStatistic.Timestamp,
			GatewayNode:            gatewayNode,
			Path:                   protoStatistic.Path,
			Application:            protoStatistic.Application,
			Service:                protoStatistic.Service,
			Hit:                    protoStatistic.Hit,
			SuccessHit:             protoStatistic.SuccessHit,
			ProxySuccessHit:        protoStatistic.ProxySuccessHit,
			DurationAverage:        protoStatistic.DurationTotal / protoStatistic.Hit,
			DurationMin:            protoStatistic.DurationMin,
			DurationMax:            protoStatistic.DurationMax,
			ServiceDurationAverage: protoStatistic.ServiceDurationTotal / protoStatistic.Hit,
			ServiceDurationMin:     protoStatistic.ServiceDurationMin,
			ServiceDurationMax:     protoStatistic.ServiceDurationMax,
		}
		res = append(res, one)
	}
	// insert hourly records
	err = s.DaoManager.ProtoStatisticHourlyDAO.Create(ctx, res)
	return res, err
}

func (s *CronService) createOrUpdateProtoStatistic(ctx context.Context, hourlyProtoStatistics []*ProtoStatisticHourly, timezone *time.Location) error {
	// iterate all service cache to keep one record a day
	services, err := s.DaoManager.ServiceDAO.GetAllServices(ctx, false /*omitProtoFile*/, true /*omitFileDescriptor*/)
	if err != nil {
		return err
	}
	if err := s.iterateServices(ctx, services, timezone); err != nil {
		return err
	}
	// handle hourly proto statistics
	for _, hourlyProtoStatistic := range hourlyProtoStatistics {
		date := time.Unix(0, hourlyProtoStatistic.Timestamp).In(timezone)
		if err := s.handleHourlyProtoStatistic(ctx, hourlyProtoStatistic, date); err != nil {
			return err
		}
	}
	return nil
}

func (s *CronService) iterateServices(ctx context.Context, services []*Service, timezone *time.Location) error {
	re := regexp.MustCompile(`post\s*:\s*"(.+)"`)
	now := time.Now().In(timezone)
	nowFmt := now.Format("2006-01-02")
	for _, service := range services {
		matrix := re.FindAllStringSubmatch(service.ProtoFile, -1)
		for _, slice := range matrix {
			protoStatistic, err := s.DaoManager.ProtoStatisticDAO.GetByDateAndPath(ctx, nowFmt, slice[1])
			if err != nil {
				s.Logger.Error("get proto statistic record from db", zap.NamedError("appError", err))
				return err
			}
			if protoStatistic == nil {
				protoStatistic = &ProtoStatistic{
					Date:          nowFmt, // date
					ApplicationID: service.ApplicationID,
					ServiceID:     service.ID,
					Path:          slice[1],
				}
				if err := s.DaoManager.ProtoStatisticDAO.Create(ctx, protoStatistic); err != nil {
					s.Logger.Error("iterate all service cache to keep one record a day", zap.NamedError("appError", err))
					continue
				}
			}
		}
	}
	return nil
}

func (s *CronService) handleHourlyProtoStatistic(ctx context.Context, hourlyProtoStatistic *ProtoStatisticHourly, date time.Time) error {
	dateFmt := date.Format("2006-01-02")
	protoStatistic, err := s.DaoManager.ProtoStatisticDAO.GetByDateAndPath(ctx, dateFmt, hourlyProtoStatistic.Path)
	if err != nil {
		s.Logger.Error("get proto statistic record from db", zap.NamedError("appError", err))
		return err
	}
	if protoStatistic == nil {
		protoStatistic = &ProtoStatistic{
			Date: dateFmt, // date
		}
	}
	if protoStatistic.ApplicationID == 0 { // application id
		application, err := s.DaoManager.ApplicationDAO.GetByName(ctx, hourlyProtoStatistic.Application)
		if err != nil {
			s.Logger.Error("get application record from db", zap.NamedError("appError", err))
			return err
		}
		if application == nil {
			s.Logger.Error(fmt.Sprintf("get application record by name [%s]", hourlyProtoStatistic.Application))
			return err
		}
		protoStatistic.ApplicationID = application.ID
	}
	if protoStatistic.ServiceID == 0 { // service id
		service, err := s.DaoManager.ServiceDAO.GetByName(ctx, hourlyProtoStatistic.Service, true /*omitProtoFile*/, true /*omitDeleted*/)
		if err != nil {
			s.Logger.Error("get service record from db", zap.NamedError("appError", err))
			return err
		}
		if service == nil {
			s.Logger.Error(fmt.Sprintf("get service record by name [%s]", hourlyProtoStatistic.Service))
			return err
		}
		protoStatistic.ServiceID = service.ID
	}
	if protoStatistic.Path == "" { // path
		protoStatistic.Path = hourlyProtoStatistic.Path
	}
	if protoStatistic.Hit+hourlyProtoStatistic.Hit > 0 {
		protoStatistic.DurationAverage = (protoStatistic.DurationAverage*protoStatistic.Hit + hourlyProtoStatistic.DurationAverage*hourlyProtoStatistic.Hit) /
			(protoStatistic.Hit + hourlyProtoStatistic.Hit) // duration average -- 计算有零值误差，但不影响接口耗时判断
	}
	if protoStatistic.DurationMin == 0 || protoStatistic.DurationMin > hourlyProtoStatistic.DurationMin {
		protoStatistic.DurationMin = hourlyProtoStatistic.DurationMin
	} // duration  min
	if protoStatistic.DurationMax == 0 || protoStatistic.DurationMax < hourlyProtoStatistic.DurationMax {
		protoStatistic.DurationMax = hourlyProtoStatistic.DurationMax
	} // duration max
	if protoStatistic.Hit+hourlyProtoStatistic.Hit > 0 {
		protoStatistic.ServiceDurationAverage = (protoStatistic.ServiceDurationAverage*protoStatistic.Hit + hourlyProtoStatistic.ServiceDurationAverage*hourlyProtoStatistic.Hit) /
			(protoStatistic.Hit + hourlyProtoStatistic.Hit) // service duration average -- 计算有零值误差，但不影响接口耗时判断
	}
	if protoStatistic.ServiceDurationMin == 0 || protoStatistic.ServiceDurationMin > hourlyProtoStatistic.ServiceDurationMin {
		protoStatistic.ServiceDurationMin = hourlyProtoStatistic.ServiceDurationMin
	} // service duration  min
	if protoStatistic.ServiceDurationMax == 0 || protoStatistic.ServiceDurationMax < hourlyProtoStatistic.ServiceDurationMax {
		protoStatistic.ServiceDurationMax = hourlyProtoStatistic.ServiceDurationMax
	} // service duration max
	protoStatistic.Hit += hourlyProtoStatistic.Hit                         // hit
	protoStatistic.SuccessHit += hourlyProtoStatistic.SuccessHit           // success hit
	protoStatistic.ProxySuccessHit += hourlyProtoStatistic.ProxySuccessHit // proxy success hit
	// create or update proto_statistic record
	if protoStatistic.ID > 0 {
		if err := s.DaoManager.ProtoStatisticDAO.Update(ctx, protoStatistic); err != nil {
			s.Logger.Error("update proto statistic record", zap.NamedError("appError", err))
		}
	} else {
		if err := s.DaoManager.ProtoStatisticDAO.Create(ctx, protoStatistic); err != nil {
			s.Logger.Error("create proto statistic record", zap.NamedError("appError", err))
		}
	}
	return nil
}

// service health check
func (s *CronService) ServiceHealthCheckEntry(ctx context.Context, crontab *cron.Cron) (*CronEntryStatus, error) {
	spec := "@every 5m"
	if cronSpec := os.Getenv("SERVICE_HEALTH_CHECK_CRON_SPEC"); cronSpec != "" {
		spec = cronSpec
	}
	entryStatus := &CronEntryStatus{
		Name:              "ServiceHealthCheck",
		Started:           true,
		StartedOrStopedAt: time.Now(),
		ScheduleSpec:      spec,
	}
	entryID, err := crontab.AddFunc(entryStatus.ScheduleSpec, func() {
		s.Logger.Info("Cron Job -- service health check", zap.Time("timestamp", time.Now()))
		s.handleServiceHealthCheck(ctx)
	})
	if err != nil {
		return nil, err
	}
	entryStatus.EntryID = entryID
	return entryStatus, nil
}

func (s *CronService) handleServiceHealthCheck(ctx context.Context) {
	services, _, err := s.ConsulClient.Catalog().Services(nil)
	if err != nil {
		s.Logger.Error("consul client agent fetch services", zap.NamedError("appError", err))
		return
	}
	alarmEmails, err := s.DaoManager.AlarmEmailDAO.GetAll(ctx)
	if err != nil {
		s.Logger.Error("internal error", zap.NamedError("appError", err))
		return
	}
	// check service whether exists, send alarm email if not exists
	servicesInDB, err := s.DaoManager.ServiceDAO.GetAllServices(ctx, true /*omitProtoFile*/, true /*omitFileDescriptor*/)
	if err != nil {
		s.Logger.Error("get all services from db", zap.NamedError("appError", err))
		return
	}
	s.iterateServicesInDB(ctx, servicesInDB, services, alarmEmails)
}

func (s *CronService) iterateServicesInDB(ctx context.Context, servicesInDB []*Service, services map[string][]string, alarmEmails []*AlarmEmail) {
	for _, srvInDB := range servicesInDB {
		var exists bool
		var message string
		for srv := range services {
			if srvInDB.Name == srv {
				exists = true
				// health check
				entries, _, err := s.ConsulClient.Health().Service(srv, "", true, nil)
				if err != nil {
					s.Logger.Error("fetch service health info", zap.String("service_name", srv), zap.NamedError("appError", err))
					message = fmt.Sprintf("service [name=%s] fetch health info error", srv)
					break
				}
				if len(entries) == 0 || entries[0].Checks[0].Status != "passing" {
					s.Logger.Error("service is not healthy", zap.String("service_name", srv), zap.Any("health_info", entries[0].Checks[0].Status))
					message = fmt.Sprintf("service [name=%s] is not healthy", srv)
					break
				}
				break
			}
		}
		if !exists {
			message = fmt.Sprintf("service [name=%s] has no running node", srvInDB.Name)
		}
		if message != "" {
			// service alarm email begins
			if s.HealthChecks[srvInDB.Name].IsZero() {
				s.HealthChecks[srvInDB.Name] = time.Now()
				for _, alarmEmail := range alarmEmails {
					if err := simpleemail.SendCronJobNotificationEmail(ctx, alarmEmail.Address, message); err != nil {
						s.Logger.Error("internal error", zap.NamedError("appError", err))
						return
					}
				}
			}
		} else {
			// service alarm email ends
			if start := s.HealthChecks[srvInDB.Name]; !start.IsZero() {
				for _, alarmEmail := range alarmEmails {
					message = fmt.Sprintf("service [name=%s] recovers, alarm duration is %s", srvInDB.Name, time.Since(start).String())
					if err := simpleemail.SendCronJobNotificationEmail(ctx, alarmEmail.Address, message); err != nil {
						s.Logger.Error("internal error", zap.NamedError("appError", err))
						return
					}
				}
				delete(s.HealthChecks, srvInDB.Name)
			}
		}
	}
}

type ListCronJobsData struct {
	List []*CronEntryStatus `json:"list"`
}

type ListCronJobsRequest struct {
	ApiKey string `json:"apiKey" validate:"required"`
}

func (s *CronService) ListCronJobs(ctx context.Context, request any) (any, error) {
	req := request.(*ListCronJobsRequest)
	// validate api key
	if err := connector.ValidateApiKey(ctx, rpc.NewContext(metadata.NewIncomingContext(ctx, metadata.MD{}), s.LocalizeManager), s.AppID, req.ApiKey, connector.RoleRead); err != nil {
		s.Logger.Error("bad request", zap.NamedError("appError", err))
		return nil, err
	}
	return &ListCronJobsData{List: s.CronEntries}, nil
}
