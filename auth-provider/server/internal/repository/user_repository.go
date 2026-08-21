package repository

import (
	"errors"
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func (r *UserRepository) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	if err := r.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetUserByID(id uuid.UUID) (models.User, error) {
	var user models.User
	if err := r.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetUserWithGroupsByID(id uuid.UUID) (models.User, error) {
	var user models.User
	if err := r.DB.Preload("Groups").Where("id = ?", id).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *UserRepository) ListUsers() ([]models.User, error) {
	var users []models.User
	if err := r.DB.Preload("Groups").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) CreateUser(user *models.User, groupIDs []uuid.UUID) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		if len(groupIDs) > 0 {
			if err := replaceUserGroups(tx, user.ID, groupIDs); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *UserRepository) UpdateUser(user *models.User, groupIDs []uuid.UUID, passwordChanged bool) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var removedGroupIDs []uuid.UUID
		if groupIDs != nil {
			var currentGroupIDs []uuid.UUID
			if err := tx.Model(&models.UserGroup{}).Where("user_id = ?", user.ID).Pluck("group_id", &currentGroupIDs).Error; err != nil {
				return err
			}
			selected := make(map[uuid.UUID]struct{}, len(groupIDs))
			for _, groupID := range groupIDs {
				selected[groupID] = struct{}{}
			}
			for _, groupID := range currentGroupIDs {
				if _, retained := selected[groupID]; !retained {
					removedGroupIDs = append(removedGroupIDs, groupID)
				}
			}
		}

		candidateApplicationIDs, err := allowedApplicationIDsForGroups(tx, removedGroupIDs)
		if err != nil {
			return err
		}
		if err := tx.Save(user).Error; err != nil {
			return err
		}
		if groupIDs != nil {
			if err := replaceUserGroups(tx, user.ID, groupIDs); err != nil {
				return err
			}
			if err := enqueueLostAccessEvents(tx, []uuid.UUID{user.ID}, candidateApplicationIDs); err != nil {
				return err
			}
		}

		if passwordChanged {
			now := time.Now()
			if err := tx.Model(&models.SSOSession{}).Where("user_id = ? AND status = 'active'", user.ID).Updates(map[string]any{
				"status":        "revoked",
				"revoked_at":    now,
				"revoke_reason": "PASSWORD_CHANGED",
			}).Error; err != nil {
				return err
			}
			event := models.Event{
				EventType: "PasswordChanged",
				UserID:    user.ID,
				Payload:   `{"reason":"PASSWORD_CHANGED"}`,
				Status:    "pending",
			}
			if err := tx.Create(&event).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *UserRepository) DeleteUser(id uuid.UUID) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Model(&models.User{}).Where("id = ?", id).Update("status", "inactive").Error; err != nil {
			return err
		}
		if err := tx.Model(&models.SSOSession{}).Where("user_id = ? AND status = 'active'", id).Updates(map[string]any{
			"status":        "revoked",
			"revoked_at":    now,
			"revoke_reason": "USER_DELETED",
		}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", id).Delete(&models.UserGroup{}).Error; err != nil {
			return err
		}
		event := models.Event{EventType: "SessionRevoked", UserID: id, Payload: `{"reason":"USER_DELETED"}`, Status: "pending"}
		if err := tx.Create(&event).Error; err != nil {
			return err
		}
		return tx.Delete(&models.User{}, "id = ?", id).Error
	})
}

func (r *UserRepository) ChangePasswordTx(userID uuid.UUID, newPasswordHash string, event *models.Event) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).Where("id = ?", userID).Update("password_hash", newPasswordHash).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.SSOSession{}).Where("user_id = ? AND status = 'active'", userID).Updates(map[string]any{
			"status":        "revoked",
			"revoked_at":    time.Now(),
			"revoke_reason": "PASSWORD_CHANGED",
		}).Error; err != nil {
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

func replaceUserGroups(tx *gorm.DB, userID uuid.UUID, groupIDs []uuid.UUID) error {
	groupIDs = uniqueUUIDs(groupIDs)
	if len(groupIDs) > 0 {
		var validGroupCount int64
		if err := tx.Model(&models.Group{}).Where("id IN ?", groupIDs).Count(&validGroupCount).Error; err != nil {
			return err
		}
		if validGroupCount != int64(len(groupIDs)) {
			return gorm.ErrRecordNotFound
		}
	}

	selected := make(map[uuid.UUID]struct{}, len(groupIDs))
	for _, groupID := range groupIDs {
		selected[groupID] = struct{}{}
	}
	var current []models.UserGroup
	if err := tx.Where("user_id = ?", userID).Find(&current).Error; err != nil {
		return err
	}
	for _, membership := range current {
		if _, retained := selected[membership.GroupID]; retained {
			delete(selected, membership.GroupID)
			continue
		}
		if err := tx.Delete(&membership).Error; err != nil {
			return err
		}
	}
	for groupID := range selected {
		var membership models.UserGroup
		err := tx.Unscoped().Where("user_id = ? AND group_id = ?", userID, groupID).First(&membership).Error
		if err == nil {
			if err := tx.Unscoped().Model(&membership).Update("deleted_at", nil).Error; err != nil {
				return err
			}
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := tx.Create(&models.UserGroup{UserID: userID, GroupID: groupID, CreatedAt: time.Now()}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *UserRepository) UpdateUserStatusTx(userID uuid.UUID, status string, event *models.Event) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).Where("id = ?", userID).Update("status", status).Error; err != nil {
			return err
		}

		if status == "inactive" {
			if err := tx.Model(&models.SSOSession{}).Where("user_id = ? AND status = 'active'", userID).Updates(map[string]any{
				"status":        "revoked",
				"revoke_reason": "USER_DEACTIVATED",
				"revoked_at":    time.Now(),
			}).Error; err != nil {
				return err
			}

			if event != nil {
				if err := tx.Create(event).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}
