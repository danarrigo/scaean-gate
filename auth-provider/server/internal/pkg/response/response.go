package response

import (
	"errors"
	"net/http"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/dto"
	"github.com/gin-gonic/gin"
)

type AppError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(statusCode int, code, message string) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

func Error(c *gin.Context, statusCode int, code, message string) {
	resp := dto.NewErrorResponse(code, message)
	if reqID := c.GetString("requestId"); reqID != "" {
		resp.Error.RequestID = reqID
	}
	c.JSON(statusCode, resp)
}

func HandleError(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		Error(c, appErr.StatusCode, appErr.Code, appErr.Message)
		return
	}
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
}

func JSON(c *gin.Context, statusCode int, data any) {
	c.JSON(statusCode, data)
}
