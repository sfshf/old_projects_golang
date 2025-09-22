package service

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/nextsurfer/oracle/internal/common/random"
	"github.com/nextsurfer/oracle/internal/common/statistic"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type CronService struct {
	*GatewayService

	Cron        *cron.Cron
	CronEntries []*CronEntryStatus
}

func NewCronService(ctx context.Context, gatewayService *GatewayService) (*CronService, error) {
	s := &CronService{
		GatewayService: gatewayService,
	}
	// cron jobs
	beijing, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil, err
	}
	// statistic upstream access info goroutine
	crontab := cron.New(cron.WithLocation(beijing))
	s.Cron = crontab
	var cronEntries []*CronEntryStatus
	// cron jobs
	statisticUpstreamAccessInfoEntry, err := s.UploadProtoStatistics(ctx, crontab)
	if err != nil {
		return nil, err
	}
	cronEntries = append(cronEntries,
		statisticUpstreamAccessInfoEntry,
	)
	s.CronEntries = cronEntries
	// start cron jobs
	s.Cron.Start()
	return s, nil
}

type CronEntryStatus struct {
	Name              string
	EntryID           cron.EntryID
	Started           bool
	StartedOrStopedAt time.Time
	ScheduleSpec      string
	LastExecError     string
}

// cron jobs -----------------------------------------------------------------------------------------

// refresh tls certificate job
func (s *GatewayService) UploadProtoStatistics(ctx context.Context, crontab *cron.Cron) (*CronEntryStatus, error) {
	spec := "@hourly"
	if cronSpec := os.Getenv("UPLOAD_PROTO_STATISTIC_CRON_SPEC"); cronSpec != "" {
		spec = cronSpec
	}
	entryStatus := &CronEntryStatus{
		Name:              "UploadProtoStatistics",
		Started:           true,
		StartedOrStopedAt: time.Now(),
		ScheduleSpec:      spec,
	}
	entryID, err := crontab.AddFunc(entryStatus.ScheduleSpec, func() {
		s.Logger.Info("Cron Job -- upload proto statistics", zap.Time("timestamp", time.Now()))
		if err := s.handleUploadProtoStatistics(ctx); err != nil {
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

func (s *GatewayService) handleUploadProtoStatistics(ctx context.Context) error {
	s.Mu.Lock()
	protoStatistics := s.ProtoStatistics
	// reset
	s.ProtoStatistics = make(map[string]map[string]*statistic.ProtoStatisticHourly, 2)
	s.Mu.Unlock()
	var list []interface{}
	for _, dateProtoStatistics := range protoStatistics {
		for _, pathProtoStatistic := range dateProtoStatistics {
			val, err := json.Marshal(pathProtoStatistic)
			if err != nil {
				return err
			}
			list = append(list, string(val))
		}
	}
	// upload to redis
	if len(list) > 0 {
		key := statistic.RedisKeyPrefixStatisticInfo + s.Name + "::" + random.GenerateUUID()
		if err := s.RedisClient.RPush(ctx, key, list...).Err(); err != nil {
			s.Logger.Error("Upload proto statistics to redis", zap.NamedError("appError", err))
			return err
		}
		if err := s.RedisClient.Expire(ctx, key, 4*time.Hour).Err(); err != nil {
			s.Logger.Error("Upload proto statistics to redis", zap.NamedError("appError", err))
			return err
		}
	}
	return nil
}
