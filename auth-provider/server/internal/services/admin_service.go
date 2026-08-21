package services

import (
	"fmt"
	"net/http"
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/dto"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/pkg/crypto"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/pkg/response"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	UserRepo    repository.UserRepository
	SessionRepo repository.SessionRepository
	GroupRepo   repository.GroupRepository
	AppRepo     repository.AppRepository
	PolicyRepo  repository.PolicyRepository
	AuditRepo   repository.AuditRepository
}

func (s *AdminService) AdminAuth(cookie string) (*models.User, error) {
	session, err := s.SessionRepo.GetSSOSessionByHash(crypto.HashSHA256(cookie))
	if err != nil || session.Status != "active" || session.ExpiresAt.Before(time.Now()) {
		return nil, response.NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", "Session Not Found")
	}

	user, err := s.UserRepo.GetUserWithGroupsByID(session.UserID)
	if err != nil || user.Status != "active" {
		return nil, response.NewAppError(http.StatusForbidden, "FORBIDDEN", "User Inactive or Not Found")
	}

	isAdmin := false
	for _, group := range user.Groups {
		if group.Name == "Admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		return nil, response.NewAppError(http.StatusForbidden, "FORBIDDEN", "Admin Privileges Required")
	}

	return &user, nil
}

func (s *AdminService) ListUsers() ([]dto.UserDetailResponse, error) {
	users, err := s.UserRepo.ListUsers()
	if err != nil {
		return nil, err
	}

	var res []dto.UserDetailResponse
	for _, u := range users {
		var groups []dto.GroupResponse
		for _, g := range u.Groups {
			groups = append(groups, dto.GroupResponse{
				ID:          g.ID,
				Name:        g.Name,
				Description: g.Description,
				CreatedAt:   g.CreatedAt,
			})
		}
		res = append(res, dto.UserDetailResponse{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			Status:    u.Status,
			Groups:    groups,
			CreatedAt: u.CreatedAt,
		})
	}
	return res, nil
}

func (s *AdminService) GetUserByID(id uuid.UUID) (*dto.UserDetailResponse, error) {
	u, err := s.UserRepo.GetUserWithGroupsByID(id)
	if err != nil {
		return nil, response.NewAppError(http.StatusNotFound, "NOT_FOUND", "User not found")
	}

	var groups []dto.GroupResponse
	for _, g := range u.Groups {
		groups = append(groups, dto.GroupResponse{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			CreatedAt:   g.CreatedAt,
		})
	}

	return &dto.UserDetailResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Status:    u.Status,
		Groups:    groups,
		CreatedAt: u.CreatedAt,
	}, nil
}

func (s *AdminService) CreateUser(req dto.CreateUserRequest) (*dto.UserDetailResponse, error) {
	if _, err := s.UserRepo.GetUserByEmail(req.Email); err == nil {
		return nil, response.NewAppError(http.StatusConflict, "CONFLICT", "Email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
		Status:       "active",
	}

	if err := s.UserRepo.CreateUser(&user, req.GroupIDs); err != nil {
		return nil, err
	}

	return s.GetUserByID(user.ID)
}

func (s *AdminService) UpdateUser(id uuid.UUID, req dto.UpdateUserRequest) (*dto.UserDetailResponse, error) {
	user, err := s.UserRepo.GetUserByID(id)
	if err != nil {
		return nil, response.NewAppError(http.StatusNotFound, "NOT_FOUND", "User not found")
	}

	user.Name = req.Name
	user.Email = req.Email
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = string(hash)
	}

	if err := s.UserRepo.UpdateUser(&user, req.GroupIDs); err != nil {
		return nil, err
	}

	return s.GetUserByID(user.ID)
}

func (s *AdminService) DeleteUser(id uuid.UUID) error {
	if _, err := s.UserRepo.GetUserByID(id); err != nil {
		return response.NewAppError(http.StatusNotFound, "NOT_FOUND", "User not found")
	}
	return s.UserRepo.DeleteUser(id)
}

func (s *AdminService) UpdateUserStatus(userID uuid.UUID, status string) error {
	if _, err := s.UserRepo.GetUserByID(userID); err != nil {
		return response.NewAppError(http.StatusNotFound, "NOT_FOUND", "User Not Found")
	}

	var event *models.Event
	if status == "inactive" {
		event = &models.Event{
			EventType: "SessionRevoked",
			UserID:    userID,
			Payload:   `{"reason":"USER_DEACTIVATED"}`,
			Status:    "pending",
		}
	}

	if err := s.UserRepo.UpdateUserStatusTx(userID, status, event); err != nil {
		return err
	}

	return nil
}

func (s *AdminService) ListGroups() ([]dto.GroupResponse, error) {
	groups, err := s.GroupRepo.ListGroups()
	if err != nil {
		return nil, err
	}

	var res []dto.GroupResponse
	for _, g := range groups {
		res = append(res, dto.GroupResponse{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			CreatedAt:   g.CreatedAt,
		})
	}
	return res, nil
}

func (s *AdminService) GetGroupByID(id uuid.UUID) (*dto.GroupDetailResponse, error) {
	g, err := s.GroupRepo.GetGroupByID(id)
	if err != nil {
		return nil, response.NewAppError(http.StatusNotFound, "NOT_FOUND", "Group not found")
	}

	var users []dto.UserHeader
	for _, u := range g.Users {
		users = append(users, dto.UserHeader{
			ID:     u.ID,
			Name:   u.Name,
			Email:  u.Email,
			Status: u.Status,
		})
	}

	return &dto.GroupDetailResponse{
		ID:          g.ID,
		Name:        g.Name,
		Description: g.Description,
		Users:       users,
		CreatedAt:   g.CreatedAt,
	}, nil
}

func (s *AdminService) CreateGroup(req dto.CreateGroupRequest) (*dto.GroupResponse, error) {
	group := models.Group{
		Name:        req.Name,
		Description: req.Description,
	}
	if err := s.GroupRepo.CreateGroup(&group); err != nil {
		return nil, err
	}
	return &dto.GroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt,
	}, nil
}

func (s *AdminService) UpdateGroup(id uuid.UUID, req dto.UpdateGroupRequest) (*dto.GroupResponse, error) {
	var group models.Group
	if err := s.GroupRepo.DB.Where("id = ?", id).First(&group).Error; err != nil {
		return nil, response.NewAppError(http.StatusNotFound, "NOT_FOUND", "Group not found")
	}

	group.Name = req.Name
	group.Description = req.Description
	if err := s.GroupRepo.UpdateGroup(&group); err != nil {
		return nil, err
	}

	return &dto.GroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt,
	}, nil
}

func (s *AdminService) DeleteGroup(id uuid.UUID) error {
	var group models.Group
	if err := s.GroupRepo.DB.Where("id = ?", id).First(&group).Error; err != nil {
		return response.NewAppError(http.StatusNotFound, "NOT_FOUND", "Group not found")
	}
	return s.GroupRepo.DeleteGroup(id)
}

func (s *AdminService) AssignUserToGroup(groupID, userID uuid.UUID) error {
	if _, err := s.UserRepo.GetUserByID(userID); err != nil {
		return response.NewAppError(http.StatusNotFound, "NOT_FOUND", "User not found")
	}
	var group models.Group
	if err := s.GroupRepo.DB.Where("id = ?", groupID).First(&group).Error; err != nil {
		return response.NewAppError(http.StatusNotFound, "NOT_FOUND", "Group not found")
	}
	return s.GroupRepo.AssignUser(groupID, userID)
}

func (s *AdminService) UnassignUserFromGroup(groupID, userID uuid.UUID) error {
	return s.GroupRepo.UnassignUser(groupID, userID)
}

func (s *AdminService) ListApps() ([]dto.AppResponse, error) {
	apps, err := s.AppRepo.ListApps()
	if err != nil {
		return nil, err
	}

	var res []dto.AppResponse
	for _, a := range apps {
		var uris []string
		var uriItems []dto.RedirectURIResponse
		for _, u := range a.RedirectURIs {
			uris = append(uris, u.RedirectURI)
			uriItems = append(uriItems, dto.RedirectURIResponse{ID: u.ID, RedirectURI: u.RedirectURI})
		}
		res = append(res, dto.AppResponse{
			ID:                    a.ID,
			Name:                  a.Name,
			ClientID:              a.ClientID,
			ClientSecretPrefix:    a.ClientSecretPrefix,
			Status:                a.Status,
			LaunchURL:             a.LaunchURL,
			LogoutNotificationURL: a.LogoutNotificationURL,
			RedirectURIs:          uris,
			RedirectURIItems:      uriItems,
			CreatedAt:             a.CreatedAt,
		})
	}
	return res, nil
}

func (s *AdminService) GetAppByID(id uuid.UUID) (*dto.AppResponse, error) {
	a, err := s.AppRepo.GetAppByID(id)
	if err != nil {
		return nil, response.NewAppError(http.StatusNotFound, "NOT_FOUND", "App not found")
	}

	var uris []string
	var uriItems []dto.RedirectURIResponse
	for _, u := range a.RedirectURIs {
		uris = append(uris, u.RedirectURI)
		uriItems = append(uriItems, dto.RedirectURIResponse{ID: u.ID, RedirectURI: u.RedirectURI})
	}

	return &dto.AppResponse{
		ID:                    a.ID,
		Name:                  a.Name,
		ClientID:              a.ClientID,
		ClientSecretPrefix:    a.ClientSecretPrefix,
		Status:                a.Status,
		LaunchURL:             a.LaunchURL,
		LogoutNotificationURL: a.LogoutNotificationURL,
		RedirectURIs:          uris,
		RedirectURIItems:      uriItems,
		CreatedAt:             a.CreatedAt,
	}, nil
}

func (s *AdminService) CreateApp(req dto.CreateAppRequest) (*dto.CreateAppResponse, error) {
	randomID, err := crypto.GenerateRandomString()
	if err != nil {
		return nil, err
	}
	clientID := fmt.Sprintf("app_%s", randomID[:12])

	secretValue, err := crypto.GenerateRandomString()
	if err != nil {
		return nil, err
	}
	rawClientSecret := "sg_" + secretValue
	hashedSecret := crypto.HashSHA256(rawClientSecret)
	secretPrefix := rawClientSecret
	if len(secretPrefix) > 11 {
		secretPrefix = secretPrefix[:11]
	}

	app := models.Application{
		Name:                  req.Name,
		ClientID:              clientID,
		ClientSecretHash:      hashedSecret,
		ClientSecretPrefix:    secretPrefix,
		Status:                "active",
		LaunchURL:             req.LaunchURL,
		LogoutNotificationURL: req.LogoutNotificationURL,
	}

	if err := s.AppRepo.CreateApp(&app, req.RedirectURIs); err != nil {
		return nil, err
	}

	return &dto.CreateAppResponse{
		ID:                    app.ID,
		Name:                  app.Name,
		ClientID:              app.ClientID,
		ClientSecret:          rawClientSecret,
		ClientSecretPrefix:    app.ClientSecretPrefix,
		Status:                app.Status,
		LaunchURL:             app.LaunchURL,
		LogoutNotificationURL: app.LogoutNotificationURL,
		RedirectURIs:          req.RedirectURIs,
		CreatedAt:             app.CreatedAt,
	}, nil
}

func (s *AdminService) UpdateApp(id uuid.UUID, req dto.UpdateAppRequest) (*dto.AppResponse, error) {
	app, err := s.AppRepo.GetAppByID(id)
	if err != nil {
		return nil, response.NewAppError(http.StatusNotFound, "NOT_FOUND", "App not found")
	}

	app.Name = req.Name
	app.LaunchURL = req.LaunchURL
	app.LogoutNotificationURL = req.LogoutNotificationURL
	app.Status = req.Status

	if err := s.AppRepo.UpdateApp(&app); err != nil {
		return nil, err
	}

	return s.GetAppByID(app.ID)
}

func (s *AdminService) DeleteApp(id uuid.UUID) error {
	if _, err := s.AppRepo.GetAppByID(id); err != nil {
		return response.NewAppError(http.StatusNotFound, "NOT_FOUND", "App not found")
	}
	return s.AppRepo.DeleteApp(id)
}

func (s *AdminService) AddRedirectURI(appID uuid.UUID, req dto.AddRedirectURIRequest) (*models.ApplicationRedirectURI, error) {
	if _, err := s.AppRepo.GetAppByID(appID); err != nil {
		return nil, response.NewAppError(http.StatusNotFound, "NOT_FOUND", "App not found")
	}
	uri, err := s.AppRepo.AddRedirectURI(appID, req.URI)
	if err != nil {
		return nil, err
	}
	return &uri, nil
}

func (s *AdminService) DeleteRedirectURI(uriID uuid.UUID) error {
	return s.AppRepo.DeleteRedirectURI(uriID)
}

func (s *AdminService) ListPolicies() ([]dto.PolicyResponse, error) {
	policies, err := s.PolicyRepo.ListPolicies()
	if err != nil {
		return nil, err
	}

	var res []dto.PolicyResponse
	for _, p := range policies {
		res = append(res, dto.PolicyResponse{
			ID:              p.ID,
			ApplicationID:   p.ApplicationID,
			ApplicationName: p.Application.Name,
			GroupID:         p.GroupID,
			GroupName:       p.Group.Name,
			Effect:          p.Effect,
			CreatedAt:       p.CreatedAt,
		})
	}
	return res, nil
}

func (s *AdminService) CreatePolicy(req dto.CreatePolicyRequest) (*dto.PolicyResponse, error) {
	if _, err := s.AppRepo.GetAppByID(req.ApplicationID); err != nil {
		return nil, response.NewAppError(http.StatusNotFound, "NOT_FOUND", "App not found")
	}
	var group models.Group
	if err := s.GroupRepo.DB.Where("id = ?", req.GroupID).First(&group).Error; err != nil {
		return nil, response.NewAppError(http.StatusNotFound, "NOT_FOUND", "Group not found")
	}

	policy := models.ApplicationGroupPolicy{
		ApplicationID: req.ApplicationID,
		GroupID:       req.GroupID,
		Effect:        req.Effect,
	}

	if err := s.PolicyRepo.CreatePolicy(&policy); err != nil {
		return nil, err
	}

	created, err := s.PolicyRepo.GetPolicyByID(policy.ID)
	if err != nil {
		return nil, err
	}

	return &dto.PolicyResponse{
		ID:              created.ID,
		ApplicationID:   created.ApplicationID,
		ApplicationName: created.Application.Name,
		GroupID:         created.GroupID,
		GroupName:       created.Group.Name,
		Effect:          created.Effect,
		CreatedAt:       created.CreatedAt,
	}, nil
}

func (s *AdminService) DeletePolicy(id uuid.UUID) error {
	policy, err := s.PolicyRepo.GetPolicyByID(id)
	if err != nil {
		return response.NewAppError(http.StatusNotFound, "NOT_FOUND", "Policy not found")
	}

	return s.PolicyRepo.DeletePolicyWithEvents(id, policy.ApplicationID, policy.GroupID)
}

func (s *AdminService) ListAuditLogs() ([]dto.AuditLogResponse, error) {
	logs, err := s.AuditRepo.ListAuditLogs()
	if err != nil {
		return nil, err
	}

	var res []dto.AuditLogResponse
	for _, l := range logs {
		res = append(res, dto.AuditLogResponse{
			ID:            l.ID,
			EventType:     l.EventType,
			ActorID:       l.ActorID,
			UserID:        l.UserID,
			ApplicationID: l.ApplicationID,
			SessionID:     l.SessionID,
			Result:        l.Result,
			Metadata:      l.Metadata,
			IPAddress:     l.IPAddress,
			CreatedAt:     l.CreatedAt,
		})
	}
	return res, nil
}

func (s *AdminService) ListEvents() ([]dto.EventResponse, error) {
	events, err := s.AuditRepo.ListEvents()
	if err != nil {
		return nil, err
	}

	var res []dto.EventResponse
	for _, e := range events {
		deliveries := make([]dto.EventDeliveryResponse, 0, len(e.Deliveries))
		for _, delivery := range e.Deliveries {
			deliveries = append(deliveries, dto.EventDeliveryResponse{
				ID:              delivery.ID,
				ApplicationID:   delivery.ApplicationID,
				ApplicationName: delivery.Application.Name,
				Status:          delivery.Status,
				AttemptCount:    delivery.AttemptCount,
				NextRetryAt:     delivery.NextRetryAt,
				ProcessedAt:     delivery.ProcessedAt,
				LastError:       delivery.LastError,
			})
		}
		res = append(res, dto.EventResponse{
			ID:            e.ID,
			EventType:     e.EventType,
			UserID:        e.UserID,
			ApplicationID: e.ApplicationID,
			Status:        e.Status,
			CreatedAt:     e.CreatedAt,
			Deliveries:    deliveries,
		})
	}
	return res, nil
}
