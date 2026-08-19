package repository

import (
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
			var groups []models.Group
			if err := tx.Where("id IN ?", groupIDs).Find(&groups).Error; err != nil {
				return err
			}
			if err := tx.Model(user).Association("Groups").Replace(groups); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *UserRepository) UpdateUser(user *models.User, groupIDs []uuid.UUID) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(user).Error; err != nil {
			return err
		}
		if groupIDs != nil {
			var groups []models.Group
			if len(groupIDs) > 0 {
				if err := tx.Where("id IN ?", groupIDs).Find(&groups).Error; err != nil {
					return err
				}
			}
			if err := tx.Model(user).Association("Groups").Replace(groups); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *UserRepository) DeleteUser(id uuid.UUID) error {
	return r.DB.Delete(&models.User{}, "id = ?", id).Error
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
