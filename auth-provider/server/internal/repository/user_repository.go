// Package repository contains CRUD ops
package repository

import (
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"

	"gorm.io/gorm"
)

type UserRepository struct{
	DB *gorm.DB
}


func (r* UserRepository) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	if err := r.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}
