package service

import (
	"fmt"
	"time"

	"github.com/alenapavlenkko/telegramfitnes/internal/models"
	"github.com/alenapavlenkko/telegramfitnes/internal/repository"
)

type WeightService struct {
	repo repository.WeightRepository
}

func NewWeightService(repo repository.WeightRepository) *WeightService {
	return &WeightService{repo: repo}
}

func (s *WeightService) LogWeight(userID uint, weight float64) error {
	if weight <= 20 || weight >= 300 {
		return fmt.Errorf("вес должен быть между 20 и 300 кг")
	}

	log := &models.WeightLog{
		UserID: userID,
		Weight: weight,
		Date:   time.Now(),
	}

	return s.repo.Create(log)
}

func (s *WeightService) GetUserHistory(userID uint) ([]*models.WeightLog, error) {
	return s.repo.GetByUser(userID)
}
