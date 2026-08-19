package handler

import (
	"net/http"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/pkg/response"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/services"
	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	AdminSvc services.AdminService
}

func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	logs, err := h.AdminSvc.ListAuditLogs()
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, logs)
}

func (h *AuditHandler) ListEvents(c *gin.Context) {
	events, err := h.AdminSvc.ListEvents()
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, events)
}
