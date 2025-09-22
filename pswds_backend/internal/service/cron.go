package service

import (
	"context"
	"os"
	"time"

	"github.com/nextsurfer/pswds_backend/internal/dao"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CronService struct {
	*PswdsService

	Cron        *cron.Cron
	CronEntries []*CronEntryStatus
}

func NewCronService(ctx context.Context, pswdsService *PswdsService) (*CronService, error) {
	s := &CronService{
		PswdsService: pswdsService,
	}
	// cron jobs
	crontab := cron.New(cron.WithSeconds())
	s.Cron = crontab
	var cronEntries []*CronEntryStatus
	// cron jobs
	clearExpiredPrivacyEmailsEntry, err := s.ClearExpiredPrivacyEmails(ctx, crontab)
	if err != nil {
		return nil, err
	}
	cronEntries = append(cronEntries, clearExpiredPrivacyEmailsEntry)
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

func (s *CronService) ClearExpiredPrivacyEmails(ctx context.Context, crontab *cron.Cron) (*CronEntryStatus, error) {
	spec := "@weekly"
	if s := os.Getenv("CLEAR_EXPIRED_PRIVACY_EMAIL_CRON"); s != "" {
		spec = s
	}
	entryStatus := &CronEntryStatus{
		Name:              "ClearExpiredPrivacyEmailsEntry",
		Started:           true,
		StartedOrStopedAt: time.Now(),
		ScheduleSpec:      spec,
	}
	entryID, err := crontab.AddFunc(entryStatus.ScheduleSpec, func() {
		s.Logger.Info("Cron Job -- ClearExpiredPrivacyEmailsEntry")
		if err := s.handleClearExpiredPrivacyEmailsEntry(ctx); err != nil {
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

func (s *CronService) handleClearExpiredPrivacyEmailsEntry(ctx context.Context) error {
	expiredAt := time.Now().Add(-15 * 24 * time.Hour).UnixMilli()
	if err := s.DaoManager.TransFunc(func(tx *gorm.DB) error {
		daoManager := dao.ManagerWithDB(tx)
		if err := daoManager.PrivacyEmailDAO.DeleteExpiredEmails(ctx, expiredAt); err != nil {
			return err
		}
		if err := daoManager.PrivacyEmailContentDAO.DeleteExpiredEmails(ctx, expiredAt); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
