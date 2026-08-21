package repository

import (
	"errors"

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
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var userIDs []uuid.UUID
		if err := tx.Table("user_groups").
			Distinct("user_groups.user_id").
			Joins("JOIN application_group_policies ON application_group_policies.group_id = user_groups.group_id").
			Where("application_group_policies.application_id = ? AND application_group_policies.deleted_at IS NULL AND user_groups.deleted_at IS NULL", id).
			Pluck("user_groups.user_id", &userIDs).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Application{}).Where("id = ?", id).Update("status", "inactive").Error; err != nil {
			return err
		}
		if err := tx.Where("application_id = ?", id).Delete(&models.ApplicationRedirectURI{}).Error; err != nil {
			return err
		}
		if err := tx.Where("application_id = ?", id).Delete(&models.ApplicationGroupPolicy{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.Application{}, "id = ?", id).Error; err != nil {
			return err
		}
		return enqueueLostAccessEvents(tx, userIDs, []uuid.UUID{id})
	})
}

func (r *AppRepository) AddRedirectURI(appID uuid.UUID, uri string) (models.ApplicationRedirectURI, error) {
	var redirect models.ApplicationRedirectURI
	err := r.DB.Unscoped().Where("application_id = ? AND redirect_uri = ?", appID, uri).First(&redirect).Error
	if err == nil {
		if !redirect.DeletedAt.Valid {
			return redirect, nil
		}
		if err := r.DB.Unscoped().Model(&redirect).Update("deleted_at", nil).Error; err != nil {
			return models.ApplicationRedirectURI{}, err
		}
		return redirect, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.ApplicationRedirectURI{}, err
	}
	redirect = models.ApplicationRedirectURI{ApplicationID: appID, RedirectURI: uri}
	if err := r.DB.Create(&redirect).Error; err != nil {
		return models.ApplicationRedirectURI{}, err
	}
	return redirect, nil
}

func (r *AppRepository) DeleteRedirectURI(uriID uuid.UUID) error {
	return r.DB.Delete(&models.ApplicationRedirectURI{}, "id = ?", uriID).Error
}
