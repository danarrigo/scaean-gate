package handler

import (
	"net/http"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/dto"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/pkg/response"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GroupHandler struct {
	AdminSvc services.AdminService
}

func (h *GroupHandler) ListGroups(c *gin.Context) {
	groups, err := h.AdminSvc.ListGroups()
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, groups)
}

func (h *GroupHandler) GetGroup(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid group ID format")
		return
	}

	group, err := h.AdminSvc.GetGroupByID(id)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, group)
}

func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var req dto.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	group, err := h.AdminSvc.CreateGroup(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, group)
}

func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid group ID format")
		return
	}

	var req dto.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	group, err := h.AdminSvc.UpdateGroup(id, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, group)
}

func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid group ID format")
		return
	}

	if err := h.AdminSvc.DeleteGroup(id); err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"message": "Group deleted successfully"})
}

func (h *GroupHandler) AssignUser(c *gin.Context) {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid group ID format")
		return
	}

	var req dto.AssignUserGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.AdminSvc.AssignUserToGroup(groupID, req.UserID); err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"message": "User assigned to group successfully"})
}

func (h *GroupHandler) UnassignUser(c *gin.Context) {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid group ID format")
		return
	}

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid user ID format")
		return
	}

	if err := h.AdminSvc.UnassignUserFromGroup(groupID, userID); err != nil {
		response.HandleError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"message": "User removed from group successfully"})
}
