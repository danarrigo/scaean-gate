package repository

import (
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PolicyRepository struct {
	DB *gorm.DB
}

func (r *PolicyRepository) HasUserAccess(userID uuid.UUID, appID uuid.UUID) (bool, error) {
	var count int64
	err := r.DB.Table("application_group_policies").
		Joins("JOIN user_groups ON user_groups.group_id = application_group_policies.group_id").
		Where("user_groups.user_id = ? AND application_group_policies.application_id = ? AND application_group_policies.effect = 'allow'", userID, appID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PolicyRepository) ListPolicies() ([]models.ApplicationGroupPolicy, error) {
	var policies []models.ApplicationGroupPolicy
	if err := r.DB.Preload("Application").Preload("Group").Find(&policies).Error; err != nil {
		return nil, err
	}
	return policies, nil
}

func (r *PolicyRepository) GetPolicyByID(id uuid.UUID) (models.ApplicationGroupPolicy, error) {
	var policy models.ApplicationGroupPolicy
	if err := r.DB.Preload("Application").Preload("Group").Where("id = ?", id).First(&policy).Error; err != nil {
		return models.ApplicationGroupPolicy{}, err
	}
	return policy, nil
}

func (r *PolicyRepository) CreatePolicy(policy *models.ApplicationGroupPolicy) error {
	return r.DB.Create(policy).Error
}

func (r *PolicyRepository) DeletePolicyTx(id uuid.UUID, event *models.Event) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.ApplicationGroupPolicy{}, "id = ?", id).Error; err != nil {
			return err
		}
		if event != nil {
			if err := tx.Create(event).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
