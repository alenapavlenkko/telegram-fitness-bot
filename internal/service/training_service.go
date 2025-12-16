package service

import (
	"fmt"
	"log"

	"github.com/alenapavlenkko/telegramfitnes/internal/models"
	"github.com/alenapavlenkko/telegramfitnes/internal/repository"
)

type TrainingService struct {
	repo repository.TrainingRepository
}

func NewTrainingService(repo repository.TrainingRepository) *TrainingService {
	return &TrainingService{repo: repo}
}

func (s *TrainingService) CreateTraining(dto CreateTrainingDTO) (*models.TrainingProgram, error) {
	// Валидация
	if dto.Title == "" {
		return nil, fmt.Errorf("название тренировки не может быть пустым")
	}
	if dto.Duration <= 0 {
		return nil, fmt.Errorf("длительность должна быть положительным числом")
	}

	training := &models.TrainingProgram{
		Title:       dto.Title,
		Description: dto.Description,
		Duration:    dto.Duration,
		Difficulty:  dto.Difficulty,
		CategoryID:  dto.CategoryID,
		YouTubeLink: dto.YouTubeLink,
	}

	return s.repo.Create(training)
}

func (s *TrainingService) ListTrainings() ([]*models.TrainingProgram, error) {
	trainings, err := s.repo.FindAll()
	if err != nil {
		log.Printf("Error in ListTrainings: %v", err)
		return nil, err
	}
	log.Printf("ListTrainings: found %d trainings", len(trainings))
	return trainings, nil
}

func (s *TrainingService) GetTrainingByID(id uint) (*models.TrainingProgram, error) {
	if id == 0 {
		return nil, fmt.Errorf("неверный ID")
	}
	return s.repo.FindByID(id)
}

func (s *TrainingService) DeleteTraining(id uint) error {
	if id == 0 {
		return fmt.Errorf("неверный ID")
	}
	return s.repo.Delete(id)
}

func (s *TrainingService) UpdateTraining(id uint, dto UpdateTrainingDTO) error {
	if id == 0 {
		return fmt.Errorf("неверный ID")
	}

	// Получаем существующую тренировку
	training, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("тренировка не найдена: %w", err)
	}

	// Валидация
	if dto.Title != "" {
		training.Title = dto.Title
	}
	if dto.Description != "" {
		training.Description = dto.Description
	}
	if dto.Difficulty != "" {
		training.Difficulty = dto.Difficulty
	}
	if dto.Duration > 0 {
		training.Duration = dto.Duration
	}
	if dto.CategoryID != nil {
		training.CategoryID = dto.CategoryID
	}

	return s.repo.Update(training)
}
