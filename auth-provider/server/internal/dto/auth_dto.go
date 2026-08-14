// Package dto consists of Data Transfer Objects
// agy --conversation=92ac2d4d-af5a-4063-8677-df3f6760f5d5
package dto

import (
	"github.com/google/uuid"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserData struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

type LoginResponse struct {
	Message string   `json:"message"`
	User    UserData `json:"user"`
}
