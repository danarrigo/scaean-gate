package repository

import (
	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func allowedApplicationIDsForGroups(tx *gorm.DB, groupIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(groupIDs) == 0 {
		return nil, nil
	}
	var applicationIDs []uuid.UUID
	if err := tx.Model(&models.ApplicationGroupPolicy{}).
		Distinct("application_id").
		Where("group_id IN ? AND effect = 'allow' AND deleted_at IS NULL", groupIDs).
		Pluck("application_id", &applicationIDs).Error; err != nil {
		return nil, err
	}
	return applicationIDs, nil
}

func enqueueLostAccessEvents(tx *gorm.DB, userIDs, applicationIDs []uuid.UUID) error {
	for _, userID := range uniqueUUIDs(userIDs) {
		for _, applicationID := range uniqueUUIDs(applicationIDs) {
			var remaining int64
			if err := tx.Table("application_group_policies AS policies").
				Joins("JOIN user_groups ON user_groups.group_id = policies.group_id").
				Where("user_groups.user_id = ? AND user_groups.deleted_at IS NULL AND policies.application_id = ? AND policies.effect = 'allow' AND policies.deleted_at IS NULL", userID, applicationID).
				Count(&remaining).Error; err != nil {
				return err
			}
			if remaining > 0 {
				continue
			}
			if err := enqueueAccessPolicyChangedEvent(tx, userID, applicationID); err != nil {
				return err
			}
		}
	}
	return nil
}

func enqueueAccessPolicyChangedEvent(tx *gorm.DB, userID, applicationID uuid.UUID) error {
	event := models.Event{
		EventType:     "AccessPolicyChanged",
		UserID:        userID,
		ApplicationID: &applicationID,
		Payload:       `{"reason":"ACCESS_POLICY_CHANGED"}`,
		Status:        "pending",
	}
	return tx.Create(&event).Error
}

func uniqueUUIDs(values []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(values))
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
