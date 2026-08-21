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

func (r *PolicyRepository) DeletePolicyWithEvents(id, applicationID, groupID uuid.UUID) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.ApplicationGroupPolicy{}, "id = ?", id).Error; err != nil {
			return err
		}

		var userIDs []uuid.UUID
		err := tx.Table("user_groups AS affected").
			Distinct("affected.user_id").
			Where("affected.group_id = ?", groupID).
			Where(`NOT EXISTS (
				SELECT 1
				FROM user_groups remaining_groups
				JOIN application_group_policies remaining_policies
				  ON remaining_policies.group_id = remaining_groups.group_id
				WHERE remaining_groups.user_id = affected.user_id
				  AND remaining_policies.application_id = ?
				  AND remaining_policies.effect = 'allow'
			)`, applicationID).
			Pluck("affected.user_id", &userIDs).Error
		if err != nil {
			return err
		}

		for _, userID := range userIDs {
			event := models.Event{
				EventType:     "AccessPolicyChanged",
				UserID:        userID,
				ApplicationID: &applicationID,
				Payload:       `{"reason":"ACCESS_POLICY_CHANGED"}`,
				Status:        "pending",
			}
			if err := tx.Create(&event).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
