package service

import (
	"github.com/alenapavlenkko/telegramfitnes/internal/models"
	"github.com/alenapavlenkko/telegramfitnes/internal/repository"
)

type CategoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

// CreateCategory - создать категорию
func (s *CategoryService) CreateCategory(dto CreateCategoryDTO) (*models.Category, error) {
	category := &models.Category{Name: dto.Name}
	return s.repo.Create(category) // repo.Create уже возвращает (*Category, error)
}

// ListCategories - список категорий
func (s *CategoryService) ListCategories() ([]*models.Category, error) {
	return s.repo.FindAll() // Используем FindAll вместо List
}

// GetCategoryByID - получить категорию по ID
func (s *CategoryService) GetCategoryByID(id uint) (*models.Category, error) {
	return s.repo.FindByID(id)
}

// DeleteCategory - удалить категорию
func (s *CategoryService) DeleteCategory(id uint) error {
	return s.repo.Delete(id)
}

// UpdateCategory - обновить категорию
func (s *CategoryService) UpdateCategory(id uint, dto UpdateCategoryDTO) error {
	category, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if dto.Name != "" {
		category.Name = dto.Name
	}
	if dto.Description != "" {
		category.Description = dto.Description
	}
	if dto.Type != "" {
		category.Type = dto.Type
	}

	return s.repo.Update(category)
}
