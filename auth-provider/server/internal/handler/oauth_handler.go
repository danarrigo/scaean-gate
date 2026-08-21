package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/danarrigo/scaean-gate/auth-provider/server/config"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/dto"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/pkg/response"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/services"
	"github.com/gin-gonic/gin"
)

type OauthHandler struct {
	OAuthSvc services.OauthService
	Cfg      *config.Config
}

func (h *OauthHandler) Authorize(c *gin.Context) {
	var req dto.AuthorizeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid Query Params")
		return
	}

	rawToken, err := c.Cookie("sso_session")
	if err != nil || rawToken == "" {
		returnTo := url.QueryEscape(c.Request.RequestURI)
		frontendURL := "http://localhost:4200"
		if h.Cfg != nil && h.Cfg.FrontendURL != "" {
			frontendURL = h.Cfg.FrontendURL
		}
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/login?return_to=%s", frontendURL, returnTo))
		return
	}

	redirectURL, err := h.OAuthSvc.AuthorizeService(rawToken, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.Redirect(http.StatusFound, redirectURL)
}

func (h *OauthHandler) Token(c *gin.Context) {
	var req dto.TokenRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid token request")
		return
	}

	tokenResp, err := h.OAuthSvc.ExchangeToken(req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, tokenResp)
}

func (h *OauthHandler) UserInfo(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid Authorization header")
		return
	}

	tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if tokenStr == "" {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing access token")
		return
	}

	userInfo, err := h.OAuthSvc.GetUserInfo(tokenStr)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, userInfo)
}
