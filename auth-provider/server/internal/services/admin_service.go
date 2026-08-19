package services

import (
	"net/http"
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/pkg/crypto"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/pkg/response"
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/repository"
)

type AdminService struct{
	UserRepo repository.UserRepository
	SessionRepo repository.SessionRepository
}

func (s *AdminService) AdminAuth (cookie string)(*models.User, error){
	session, err := s.SessionRepo.GetSSOSessionByHash(crypto.HashSHA256(cookie))
	if err!=nil || session.Status!="active" || session.ExpiresAt.Before(time.Now()){
		return nil,response.NewAppError(http.StatusUnauthorized,"UNAUTHORIZED","Session Not Found")
	}

	user , err := s.UserRepo.GetUserWithGroupsByID(session.UserID)
	if err!=nil || user.Status!= "active" {
		return nil, response.NewAppError(http.StatusBadRequest,"BAD REQUEST", "User Not Found")
	}

	isAdmin:= false
	for _,group := range user.Groups{
		if group.Name == "Admin" {
			isAdmin = true 
			break 
		}
	}
	
	if !isAdmin {
		return nil,response.NewAppError(http.StatusForbidden,"FORBIDDEN", "Admin Privileges Required")
	}

	return &user,nil
}
