package services

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nextsurfer/book-manage-api/api/code"
	"github.com/nextsurfer/book-manage-api/internal/app/dao"
	"github.com/nextsurfer/book-manage-api/internal/tools"
)

type SystemService struct {
	dao *dao.Manager
}

func NewSystemService() *SystemService {
	return &SystemService{
		dao: dao.NewManagerWithDB(tools.MysqlDB()),
	}
}

func (s *SystemService) CheckAdmin(c *gin.Context) (bool, int32, string) {
	password := c.Query("password")

	// password check
	if !tools.CheckAdminPassword(password) {
		return false, code.ErrPassword, "wrong password"
	}
	return true, code.Ok, ""
}

func (s *SystemService) ListStaff(c *gin.Context) ([]string, int32, string) {
	password := c.Query("password")

	// password check
	if !tools.CheckPassword(password) {
		return nil, code.ErrPassword, "wrong password"
	}

	var err error

	var nonAdmin bool
	nonAdminStr := c.Query("nonAdmin")
	if nonAdminStr != "" {
		nonAdmin, err = strconv.ParseBool(nonAdminStr)
		if err != nil {
			return nil, code.ErrWrongParam, "invalid nonAdmin parameter"
		}
	}

	var list []string
	for _, staff := range tools.StaffList() {
		if staff.User == "admin" && nonAdmin {
			continue
		}
		list = append(list, staff.User)
	}
	return list, code.Ok, ""
}
