// Package handler consists of handlers
package handler

import (
	"net/http"
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/dto"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/pkg/response"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/services"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	AuthSvc services.AuthService
}

func (u *UserHandler) LoginHandler(c *gin.Context) {
	var body dto.LoginRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	result, err := u.AuthSvc.Login(body.Email, body.Password, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("sso_session", result.RawToken, int(24*time.Hour.Seconds()), "/", "", false, true)

	response.JSON(c, http.StatusOK, dto.LoginResponse{
		Message: "Login Successful",
		User: dto.UserData{
			ID:    result.User.ID,
			Name:  result.User.Name,
			Email: result.User.Email,
		},
	})
}

func (u *UserHandler)Logout(c *gin.Context){
	cookie,err:= c.Cookie("sso_session");if err!= nil{
		response.Error(c,http.StatusUnauthorized,"UNAUTHORIZED","No active session")
		return 
	}

}

func ChangePassword(c *gin.Context) {
}

func ShowProfile(c *gin.Context) {
}
