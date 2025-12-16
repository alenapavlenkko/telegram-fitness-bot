package repository

import (
	"github.com/alenapavlenkko/telegramfitnes/internal/models"
	"gorm.io/gorm"
)

type ProgressRepository interface {
	Create(progress *models.UserProgress) (*models.UserProgress, error)
	FindByUserID(userID uint) ([]*models.UserProgress, error)
	CountCompleted() (int64, error)
}

type progressRepo struct {
	db *gorm.DB
}

func NewProgressRepo(db *gorm.DB) ProgressRepository {
	return &progressRepo{db: db}
}

func (r *progressRepo) Create(progress *models.UserProgress) (*models.UserProgress, error) {
	err := r.db.Create(progress).Error
	return progress, err
}

func (r *progressRepo) FindByUserID(userID uint) ([]*models.UserProgress, error) {
	var progresses []*models.UserProgress
	err := r.db.Where("user_id = ?", userID).Find(&progresses).Error
	return progresses, err
}

func (r *progressRepo) CountCompleted() (int64, error) {
	var count int64
	err := r.db.Model(&models.UserProgress{}).Count(&count).Error
	return count, err
}
