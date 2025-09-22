package api

import (
	"net/http"

	"github.com/nextsurfer/book-manage-api/api/code"

	"github.com/gin-gonic/gin"
)

// Response data
type Response struct {
	Code         int32       `json:"code"`
	Data         interface{} `json:"data"`
	Message      string      `json:"message"`
	DebugMessage string      `json:"debugMessage"`
}

// NewResponse ...
func NewResponse(errCode int32, errMsg string, data interface{}) *Response {
	res := new(Response)
	if data != nil {
		res.Data = data
	}
	res.Code = errCode
	res.Message = errMsg
	return res
}

// ErrorResponse error response
func ErrorResponse(c *gin.Context, errCode int32, errMsg string) {
	c.JSON(http.StatusOK, NewResponse(errCode, errMsg, nil))
	c.Abort()
}

// SuccessResponse success response
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, NewResponse(code.Ok, "", data))
	c.Abort()
}
