package repository

import (
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AppRepository struct {
	DB *gorm.DB
}

func (r *AppRepository) GetAppByID(id uuid.UUID) (models.Application, error) {
	var app models.Application
	if err := r.DB.Preload("RedirectURIs").Where("id = ?", id).First(&app).Error; err != nil {
		return models.Application{}, err
	}
	return app, nil
}

func (r *AppRepository) GetAppByClientID(clientID string) (models.Application, error) {
	var app models.Application
	if err := r.DB.Preload("RedirectURIs").Where("client_id = ?", clientID).First(&app).Error; err != nil {
		return models.Application{}, err
	}
	return app, nil
}

func (r *AppRepository) ListApps() ([]models.Application, error) {
	var apps []models.Application
	if err := r.DB.Preload("RedirectURIs").Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (r *AppRepository) CreateApp(app *models.Application, redirectURIs []string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(app).Error; err != nil {
			return err
		}
		for _, uri := range redirectURIs {
			redirect := models.ApplicationRedirectURI{
				ApplicationID: app.ID,
				RedirectURI:   uri,
			}
			if err := tx.Create(&redirect).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *AppRepository) UpdateApp(app *models.Application) error {
	return r.DB.Save(app).Error
}

func (r *AppRepository) DeleteApp(id uuid.UUID) error {
	return r.DB.Delete(&models.Application{}, "id = ?", id).Error
}

func (r *AppRepository) AddRedirectURI(appID uuid.UUID, uri string) (models.ApplicationRedirectURI, error) {
	redirect := models.ApplicationRedirectURI{
		ApplicationID: appID,
		RedirectURI:   uri,
	}
	if err := r.DB.Create(&redirect).Error; err != nil {
		return models.ApplicationRedirectURI{}, err
	}
	return redirect, nil
}

func (r *AppRepository) DeleteRedirectURI(uriID uuid.UUID) error {
	return r.DB.Delete(&models.ApplicationRedirectURI{}, "id = ?", uriID).Error
}
