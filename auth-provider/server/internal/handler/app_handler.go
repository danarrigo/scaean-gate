package handler

import (
	"net/http"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/dto"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/pkg/response"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AppHandler struct {
	AdminSvc services.AdminService
}

func (h *AppHandler) ListApps(c *gin.Context) {
	apps, err := h.AdminSvc.ListApps()
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, apps)
}

func (h *AppHandler) GetApp(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid app ID format")
		return
	}

	app, err := h.AdminSvc.GetAppByID(id)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, app)
}

func (h *AppHandler) CreateApp(c *gin.Context) {
	var req dto.CreateAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	app, err := h.AdminSvc.CreateApp(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, app)
}

func (h *AppHandler) UpdateApp(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid app ID format")
		return
	}

	var req dto.UpdateAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	app, err := h.AdminSvc.UpdateApp(id, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, app)
}

func (h *AppHandler) DeleteApp(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid app ID format")
		return
	}

	if err := h.AdminSvc.DeleteApp(id); err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"message": "App deleted successfully"})
}

func (h *AppHandler) AddRedirectURI(c *gin.Context) {
	appID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid app ID format")
		return
	}

	var req dto.AddRedirectURIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	uri, err := h.AdminSvc.AddRedirectURI(appID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, uri)
}

func (h *AppHandler) DeleteRedirectURI(c *gin.Context) {
	uriID, err := uuid.Parse(c.Param("uri_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid redirect URI ID format")
		return
	}

	if err := h.AdminSvc.DeleteRedirectURI(uriID); err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"message": "Redirect URI deleted successfully"})
}
