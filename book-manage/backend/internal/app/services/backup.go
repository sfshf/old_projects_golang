package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/nextsurfer/book-manage-api/api"
	"github.com/nextsurfer/book-manage-api/api/code"
	"github.com/nextsurfer/book-manage-api/internal/app/batch"
	"github.com/nextsurfer/book-manage-api/internal/app/dao"
	"github.com/nextsurfer/book-manage-api/internal/app/model"
	"github.com/nextsurfer/book-manage-api/internal/tools"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type BackupService struct {
	dao *dao.Manager

	cron       *cron.Cron
	entryID    cron.EntryID
	cronStatus struct {
		Started           bool
		StartedOrStopedAt time.Time
		ScheduleSpec      string
		LastExecError     string
	}

	updating    bool
	pendingLogs []string
	wordCount   int
	// percentage of progress:
	// 0-5% is for read csv
	// 5-100% is for word count
	progress    int
	uploadError error
}

func NewBackupService() *BackupService {
	s := &BackupService{
		dao: dao.NewManagerWithDB(tools.MysqlDB()),
	}
	cron := cron.New()
	entryID, err := cron.AddFunc("@weekly", func() {
		if err := makeBackup(s.dao, "all"); err != nil {
			s.cronStatus.LastExecError = err.Error()
		}
	})
	if err != nil {
		panic(err)
	}
	s.entryID = entryID
	s.cron = cron
	s.cronStatus.StartedOrStopedAt = time.Now()
	s.cronStatus.ScheduleSpec = "@weekly"
	return s
}

func (s *BackupService) ListBackups(c *gin.Context) ([]model.Backup, int32, string) {
	password := c.Query("password")

	// password check
	if !tools.CheckAdminPassword(password) {
		return nil, code.ErrPassword, "wrong password"
	}

	backupLogs, err := s.dao.BackupDAO.GetAll(context.TODO())
	if err != nil {
		return nil, code.ErrInternal, err.Error()
	}

	return backupLogs, code.Ok, ""
}

func (s *BackupService) MakeBackup(c *gin.Context) (int32, string) {
	var err error
	var req api.MakeBackupRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		return code.ErrWrongParam, err.Error()
	}

	// password check
	if !tools.CheckAdminPassword(req.Password) {
		return code.ErrPassword, "wrong password"
	}

	if err = makeBackup(s.dao, req.Book); err != nil {
		return code.ErrInternal, err.Error()
	}

	return code.Ok, ""
}

func makeBackup(daoManager *dao.Manager, book string) error {
	var err error

	var books []model.Book
	if book != "all" {
		bookID, err := strconv.ParseInt(book, 10, 64)
		if err != nil {
			return errors.New("bookID wrong")
		}
		// check book id
		book, err := daoManager.BookDAO.GetFromID(context.TODO(), bookID)
		if err != nil {
			return errors.New("bookID wrong")
		}
		books = append(books, *book)
	} else {
		books, err = daoManager.BookDAO.GetAll(context.TODO())
		if err != nil {
			return err
		}
	}

	backupPath := tools.Config().BackupPath
	// check backup directory exist
	if _, err = os.Stat(backupPath); os.IsNotExist(err) {
		// if not exist, create folder
		err = os.Mkdir(backupPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	var backups []model.Backup
outer:
	for _, book := range books {
		filedir := backupPath + "/" + strconv.FormatInt(book.ID, 10) + "/"
		filename := strconv.FormatInt(book.UpdatedAt.UnixMilli(), 10) + ".csv"
		// check file directory exist
		if _, err = os.Stat(filedir); os.IsNotExist(err) {
			// if not exist, create folder
			err = os.Mkdir(filedir, os.ModePerm)
			if err != nil {
				return err
			}
		}
		// check file exist
		files, err := os.ReadDir(filedir)
		if err != nil {
			return err
		}
		for _, file := range files {
			if file.Name() == filename {
				continue outer
			}
		}
		if len(files) == 3 {
			sort.Slice(files, func(i, j int) bool {
				fileInfo1, _ := files[i].Info()
				fileInfo2, _ := files[j].Info()
				return fileInfo1.ModTime().Before(fileInfo2.ModTime())
			})
			// remove the oldest file
			if err = os.Remove(filedir + files[0].Name()); err != nil {
				return err
			}
		}
		err = batch.ExportBook(book.ID, filedir+filename, daoManager)
		if err != nil {
			return err
		}
		backups = append(backups, model.Backup{
			BookID:   book.ID,
			Filepath: filedir + filename,
		})
	}

	if err = daoManager.DB.Transaction(func(tx *gorm.DB) error {
		manager := dao.NewManagerWithDB(tx)

		for _, backup := range backups {
			exists, err := manager.BackupDAO.GetByBookID(context.TODO(), backup.BookID)
			if err != nil {
				return err
			}
			if len(exists) == 3 {
				sort.Slice(exists, func(i int, j int) bool {
					return exists[i].CreatedAt.Before(exists[j].CreatedAt)
				})
				if err = manager.BackupDAO.DeleteByID(context.TODO(), exists[0].ID); err != nil {
					return err
				}
			}
			if err = manager.BackupDAO.Create(context.TODO(), &backup); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *BackupService) RegainBackup(c *gin.Context) (int32, string) {
	if s.updating {
		return code.ErrWrongParam, "A file is uploading, please wait"
	}

	var err error
	var req api.RegainBackupRequest
	if err = c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		return code.ErrWrongParam, err.Error()
	}

	// check bookID and backupID
	book, err := s.dao.BookDAO.GetFromID(context.TODO(), req.BookID)
	if err != nil {
		return code.ErrInternal, err.Error()
	}
	backup, err := s.dao.BackupDAO.GetFromID(context.TODO(), req.BackupID)
	if err != nil {
		return code.ErrInternal, err.Error()
	}
	if book.ID != backup.BookID {
		return code.ErrWrongParam, "invalid book id or backup id"
	}

	s.start()
	go batch.RegainBackup(backup.Filepath, book, s.dao, s)

	return code.Ok, ""
}

func (s *BackupService) CheckRegainingLog(c *gin.Context) (*api.UploadingLogResponseData, int32, string) {
	password := c.Query("password")

	// password check
	if !tools.CheckAdminPassword(password) {
		return nil, code.ErrPassword, "wrong password"
	}

	data := &api.UploadingLogResponseData{
		Logs:      s.pendingLogs,
		WordCount: s.wordCount,
		Progress:  s.progress,
		Error:     "",
	}
	if s.uploadError != nil {
		data.Error = s.uploadError.Error()
	}
	s.pendingLogs = []string{}

	return data, code.Ok, ""
}

func (s *BackupService) GetCronStatus(c *gin.Context) (*api.GetCronStatusResponseData, int32, string) {
	password := c.Query("password")

	// password check
	if !tools.CheckAdminPassword(password) {
		return nil, code.ErrPassword, "wrong password"
	}

	return &api.GetCronStatusResponseData{
		Started:           s.cronStatus.Started,
		StartedOrStopedAt: s.cronStatus.StartedOrStopedAt,
		NextTime:          s.cron.Entry(s.entryID).Next,
		ScheduleSpec:      s.cronStatus.ScheduleSpec,
		LastExecError:     s.cronStatus.LastExecError,
	}, code.Ok, ""
}

func (s *BackupService) UpdateCronSetting(c *gin.Context) (*api.GetCronStatusResponseData, int32, string) {
	var req api.SetCronJobRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		return nil, code.ErrWrongParam, err.Error()
	}

	// password check
	if !tools.CheckAdminPassword(req.Password) {
		return nil, code.ErrPassword, "wrong password"
	}

	if req.StartCron != "" {
		s.cronStatus.Started = req.StartCron == "start"
		s.cronStatus.StartedOrStopedAt = time.Now()
		if s.cronStatus.Started {
			s.cron.Start()
		} else {
			s.cron.Stop()
		}
	} else if req.ScheduleSpec != "" {
		s.cronStatus.ScheduleSpec = req.ScheduleSpec
		cur := s.cron.Entry(s.entryID)
		s.cron.Remove(s.entryID)
		newEntryID, err := s.cron.AddJob(s.cronStatus.ScheduleSpec, cur.Job)
		if err != nil {
			return nil, code.ErrInternal, ""
		}
		s.entryID = newEntryID
	}

	return &api.GetCronStatusResponseData{
		Started:           s.cronStatus.Started,
		StartedOrStopedAt: s.cronStatus.StartedOrStopedAt,
		NextTime:          s.cron.Entry(s.entryID).Next,
		ScheduleSpec:      s.cronStatus.ScheduleSpec,
		LastExecError:     s.cronStatus.LastExecError,
	}, code.Ok, ""
}

func (s *BackupService) Complete(err error, params ...any) {
	operateLog := &model.OperateLog{
		CreatedBy:          "", // operator or admin
		OperateStatus:      1,  // default is failure
		OperateType:        "",
		OtherOperateParams: "{}",
		BookID:             0,
	}
	if len(params) >= 1 {
		operateLog.CreatedBy = params[0].(string)
	}
	if len(params) >= 2 {
		operateLog.OperateType = params[1].(string)
	}
	if len(params) >= 3 {
		operateLog.BookID = params[2].(int64)
	}
	if len(params) >= 4 {
		operateLog.DefinitionID = params[3].(int64)
	}
	if err != nil {
		s.updating = false
		s.uploadError = err
		operateLog.Error = err.Error()
	} else {
		s.progress = 100
		s.updating = false
		operateLog.OperateStatus = 2 // success
	}
	if err := s.dao.OperateLogDAO.Create(context.TODO(), operateLog); err != nil {
		fmt.Println(err)
	}
}

func (s *BackupService) VerbosePrint(v ...any) {
	if s.updating {
		s.pendingLogs = append(s.pendingLogs, fmt.Sprintln(v...))
	}
}

func (s *BackupService) InfoPrint(v ...any) {
	if s.updating {
		fmt.Println(v...)
		s.pendingLogs = append(s.pendingLogs, fmt.Sprintln(v...))
	}
}

func (s *BackupService) Progress(line int, total int) {
	if s.updating {
		s.progress = 5 + line*95/total
	}
}

func (s *BackupService) CountDefinition(count int) {
	if s.updating {
		s.wordCount = count
	}
}

func (s *BackupService) start() {
	s.updating = true
	s.pendingLogs = []string{}
	s.wordCount = 0
	s.progress = 0
	s.uploadError = nil
}
