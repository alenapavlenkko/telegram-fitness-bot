package repository

import (
	"github.com/alenapavlenkko/telegramfitnes/internal/models"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(category *models.Category) (*models.Category, error)
	FindAll() ([]*models.Category, error)
	FindByID(id uint) (*models.Category, error)
	Update(category *models.Category) error
	Delete(id uint) error
}

type categoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepo(db *gorm.DB) CategoryRepository {
	return &categoryRepo{db: db}
}

func (r *categoryRepo) Create(category *models.Category) (*models.Category, error) {
	err := r.db.Create(category).Error
	return category, err
}

func (r *categoryRepo) FindAll() ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.Find(&categories).Error
	return categories, err
}

func (r *categoryRepo) FindByID(id uint) (*models.Category, error) {
	var category models.Category
	err := r.db.First(&category, id).Error
	return &category, err
}

func (r *categoryRepo) Update(category *models.Category) error {
	return r.db.Save(category).Error
}

func (r *categoryRepo) Delete(id uint) error {
	return r.db.Delete(&models.Category{}, id).Error
}
