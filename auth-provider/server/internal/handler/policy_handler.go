package handler

import (
	"net/http"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/dto"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/pkg/response"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PolicyHandler struct {
	AdminSvc services.AdminService
}

func (h *PolicyHandler) ListPolicies(c *gin.Context) {
	policies, err := h.AdminSvc.ListPolicies()
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, policies)
}

func (h *PolicyHandler) GetPolicy(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid policy ID format")
		return
	}
	policy, err := h.AdminSvc.GetPolicyByID(id)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, policy)
}

func (h *PolicyHandler) CreatePolicy(c *gin.Context) {
	var req dto.CreatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	policy, err := h.AdminSvc.CreatePolicy(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, policy)
}

func (h *PolicyHandler) UpdatePolicy(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid policy ID format")
		return
	}
	var req dto.UpdatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}
	policy, err := h.AdminSvc.UpdatePolicy(id, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, policy)
}

func (h *PolicyHandler) DeletePolicy(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid policy ID format")
		return
	}

	if err := h.AdminSvc.DeletePolicy(id); err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"message": "Policy deleted successfully"})
}
