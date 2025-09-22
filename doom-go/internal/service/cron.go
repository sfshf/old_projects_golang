package service

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
)

type CronService struct {
	*DoomService

	Cron        *cron.Cron
	CronEntries []*CronEntryStatus
}

func NewCronService(ctx context.Context, DoomService *DoomService) (*CronService, error) {
	s := &CronService{
		DoomService: DoomService,
	}
	// cron jobs
	crontab := cron.New(cron.WithSeconds())
	s.Cron = crontab
	var cronEntries []*CronEntryStatus
	// cron jobs
	// cronEntries = append(cronEntries)
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
