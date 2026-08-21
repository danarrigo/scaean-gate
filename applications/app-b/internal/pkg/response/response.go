package response

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AppError struct {
	StatusCode int
	Code       string
	Message    string
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

type ErrorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	RequestID string `json:"requestId"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

func JSON(c *gin.Context, statusCode int, data any) {
	c.JSON(statusCode, data)
}

func Error(c *gin.Context, statusCode int, code, message string) {
	reqID := c.GetString("RequestID")
	if reqID == "" {
		reqID = uuid.New().String()
	}

	c.JSON(statusCode, ErrorResponse{
		Error: ErrorDetail{
			Code:      code,
			Message:   message,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			RequestID: reqID,
		},
	})
}

func HandleError(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		Error(c, appErr.StatusCode, appErr.Code, appErr.Message)
		return
	}

	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
}
