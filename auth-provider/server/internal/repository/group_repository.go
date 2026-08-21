package repository

import (
	"errors"
	"time"

	"github.com/danarrigo/scaean-gate/auth-provider/server/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GroupRepository struct {
	DB *gorm.DB
}

func (r *GroupRepository) ListGroups() ([]models.Group, error) {
	var groups []models.Group
	if err := r.DB.Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *GroupRepository) GetGroupByID(id uuid.UUID) (models.Group, error) {
	var group models.Group
	if err := r.DB.Preload("Users").Where("id = ?", id).First(&group).Error; err != nil {
		return models.Group{}, err
	}
	return group, nil
}

func (r *GroupRepository) CreateGroup(group *models.Group) error {
	var existing models.Group
	err := r.DB.Unscoped().Where("name = ?", group.Name).First(&existing).Error
	if err == nil {
		if !existing.DeletedAt.Valid {
			return r.DB.Create(group).Error
		}
		description := group.Description
		if err := r.DB.Unscoped().Model(&existing).Updates(map[string]any{
			"description": description,
			"deleted_at":  nil,
		}).Error; err != nil {
			return err
		}
		*group = existing
		group.Description = description
		group.DeletedAt = gorm.DeletedAt{}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return r.DB.Create(group).Error
}

func (r *GroupRepository) UpdateGroup(group *models.Group) error {
	return r.DB.Save(group).Error
}

func (r *GroupRepository) DeleteGroupWithEvents(id uuid.UUID) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var userIDs []uuid.UUID
		if err := tx.Model(&models.UserGroup{}).Where("group_id = ?", id).Pluck("user_id", &userIDs).Error; err != nil {
			return err
		}
		applicationIDs, err := allowedApplicationIDsForGroups(tx, []uuid.UUID{id})
		if err != nil {
			return err
		}
		if err := tx.Where("group_id = ?", id).Delete(&models.UserGroup{}).Error; err != nil {
			return err
		}
		if err := tx.Where("group_id = ?", id).Delete(&models.ApplicationGroupPolicy{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.Group{}, "id = ?", id).Error; err != nil {
			return err
		}
		return enqueueLostAccessEvents(tx, userIDs, applicationIDs)
	})
}

func (r *GroupRepository) AssignUser(groupID, userID uuid.UUID) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&models.Group{}, "id = ?", groupID).Error; err != nil {
			return err
		}
		if err := tx.First(&models.User{}, "id = ?", userID).Error; err != nil {
			return err
		}
		var membership models.UserGroup
		err := tx.Unscoped().Where("group_id = ? AND user_id = ?", groupID, userID).First(&membership).Error
		if err == nil {
			if !membership.DeletedAt.Valid {
				return nil
			}
			return tx.Unscoped().Model(&membership).Update("deleted_at", nil).Error
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return tx.Create(&models.UserGroup{GroupID: groupID, UserID: userID, CreatedAt: time.Now()}).Error
	})
}

func (r *GroupRepository) UnassignUserWithEvents(groupID, userID uuid.UUID) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		applicationIDs, err := allowedApplicationIDsForGroups(tx, []uuid.UUID{groupID})
		if err != nil {
			return err
		}
		result := tx.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&models.UserGroup{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return enqueueLostAccessEvents(tx, []uuid.UUID{userID}, applicationIDs)
	})
}
