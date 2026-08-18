package repository

import (

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	"gorm.io/gorm"
)

type AppRepository struct{
	DB *gorm.DB
}

func (r *AppRepository) GetAppByID (id string)(models.Application,error){
	app := models.Application{}
	if err := r.DB.Preload("RedirectURIs").Where("client_id = ?",id).First(&app).Error;err!=nil{
		return models.Application{},err
	}
	return app, nil
}

