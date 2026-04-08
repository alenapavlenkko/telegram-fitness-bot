package repository

import (
	"github.com/alenapavlenkko/telegramfitnes/internal/models"
	"gorm.io/gorm"
)

type WeightRepository interface {
	Create(log *models.WeightLog) error
	GetByUser(userID uint) ([]*models.WeightLog, error)
	DeleteByID(id uint) error
}

type weightRepo struct {
	db *gorm.DB
}

func NewWeightRepository(db *gorm.DB) WeightRepository {
	return &weightRepo{db: db}
}

func (r *weightRepo) Create(log *models.WeightLog) error {
	return r.db.Create(log).Error
}

func (r *weightRepo) GetByUser(userID uint) ([]*models.WeightLog, error) {
	var logs []*models.WeightLog
	err := r.db.Where("user_id = ?", userID).Order("date asc").Find(&logs).Error
	return logs, err
}

func (r *weightRepo) DeleteByID(id uint) error {
	return r.db.Delete(&models.WeightLog{}, id).Error
}
