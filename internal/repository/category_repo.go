package repository

import (
	"github.com/alenapavlenkko/telegramfitnes/internal/models"
	"gorm.io/gorm"
)

// CategoryRepository отвечает за работу с категориями в БД
type CategoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository создает новый repository для категорий
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Create создает новую категорию
func (r *CategoryRepository) Create(category *models.Category) (*models.Category, error) {
	err := r.db.Create(category).Error
	return category, err
}

// FindAll возвращает список всех категорий
func (r *CategoryRepository) FindAll() ([]*models.Category, error) {
	var categories []*models.Category

	err := r.db.Find(&categories).Error

	return categories, err
}

// FindByID возвращает категорию по ID
func (r *CategoryRepository) FindByID(id uint) (*models.Category, error) {
	var category models.Category

	err := r.db.First(&category, id).Error

	return &category, err
}

// Update обновляет категорию
func (r *CategoryRepository) Update(category *models.Category) error {
	return r.db.Save(category).Error
}

// Delete удаляет категорию
func (r *CategoryRepository) Delete(id uint) error {
	return r.db.Delete(&models.Category{}, id).Error
}
