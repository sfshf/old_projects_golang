package dto

import "github.com/nextsurfer/book-manage-api/internal/tools"

type OperateLogDTO struct {
	ID                 int64      `json:"id"`
	Operator           string     `json:"operator"`
	OperateTime        tools.Time `json:"operateTime"`
	OperateStatus      int32      `json:"operateStatus"`
	OperateType        string     `json:"operateType"`
	BookID             int64      `json:"bookID"`
	DefinitionID       int64      `json:"definitionID"`
	OtherOperateParams string     `json:"otherOperateParams"`
	Error              string     `json:"error"`
}
