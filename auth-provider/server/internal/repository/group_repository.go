package repository

import (
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
	return r.DB.Create(group).Error
}

func (r *GroupRepository) UpdateGroup(group *models.Group) error {
	return r.DB.Save(group).Error
}

func (r *GroupRepository) DeleteGroup(id uuid.UUID) error {
	return r.DB.Delete(&models.Group{}, "id = ?", id).Error
}

func (r *GroupRepository) AssignUser(groupID, userID uuid.UUID) error {
	var group models.Group
	if err := r.DB.Where("id = ?", groupID).First(&group).Error; err != nil {
		return err
	}
	var user models.User
	if err := r.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return err
	}
	return r.DB.Model(&group).Association("Users").Append(&user)
}

func (r *GroupRepository) UnassignUser(groupID, userID uuid.UUID) error {
	var group models.Group
	if err := r.DB.Where("id = ?", groupID).First(&group).Error; err != nil {
		return err
	}
	var user models.User
	if err := r.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return err
	}
	return r.DB.Model(&group).Association("Users").Delete(&user)
}
