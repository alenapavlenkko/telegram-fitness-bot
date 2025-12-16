package service

import (
	"github.com/alenapavlenkko/telegramfitnes/internal/models"
	"github.com/alenapavlenkko/telegramfitnes/internal/repository"
)

type ProgressService struct {
	repo repository.ProgressRepository
}

func NewProgressService(repo repository.ProgressRepository) *ProgressService {
	return &ProgressService{repo: repo}
}

func (s *ProgressService) AddProgress(userID uint, trainingID uint) (*models.UserProgress, error) {
	progress := &models.UserProgress{
		UserID:     userID,
		TrainingID: trainingID,
	}
	return s.repo.Create(progress)
}

// GetUserProgress - получить прогресс пользователя
func (s *ProgressService) GetUserProgress(userID uint) ([]*models.UserProgress, error) {
	return s.repo.FindByUserID(userID)
}

// GetTotalCompleted - общее количество выполненных тренировок
func (s *ProgressService) GetTotalCompleted() (int64, error) {
	return s.repo.CountCompleted()
}
