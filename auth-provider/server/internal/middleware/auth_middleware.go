// Package middleware consists of middleware that intercepts http requests
package middleware

import (
	"net/http"
	
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/pkg/response"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthMiddlewareHandler struct {
	AuthSvc services.AdminService
}

func (h *AuthMiddlewareHandler) AuthMiddleWare (c *gin.Context){
	cookie,err := c.Cookie("sso_session")
	if err!=nil{
		response.Error(c,http.StatusUnauthorized, "UNAUTHORIZED", "Session Not Found")
		c.Abort()
		return
	}

	user,err:= h.AuthSvc.AdminAuth(cookie)
	if err!=nil {
		response.HandleError(c,err)
		c.Abort()
		return
	}

	c.Set("currentUser", user)
	c.Next()

}
