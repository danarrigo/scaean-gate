package repository

import (
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
