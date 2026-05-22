package service

import (
	"github.com/alenapavlenkko/telegramfitnes/internal/models"
	"github.com/alenapavlenkko/telegramfitnes/internal/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// CreateUser - создать пользователя
func (s *UserService) CreateUser(dto CreateUserDTO) (*models.User, error) {
	user := &models.User{
		TelegramID: dto.TelegramID,
		Name:       dto.Name,
		Role:       dto.Role,
	}
	return s.repo.Create(user)
}

// GetUserByTelegramID - получить пользователя по Telegram ID
func (s *UserService) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	return s.repo.FindByTelegramID(telegramID)
}

// GetUsersCount - количество пользователей
func (s *UserService) GetUsersCount() (int64, error) {
	return s.repo.Count()
}

// GetAllUsers - все пользователи
func (s *UserService) GetAllUsers() ([]*models.User, error) {
	return s.repo.FindAll()
}

type UpdateProfileDTO struct {
	TelegramID   int64   `json:"telegramId"`
	Name         string  `json:"name"`
	Age          int     `json:"age"`
	Gender       string  `json:"gender"`
	Height       float64 `json:"height"`
	Weight       float64 `json:"weight"`
	Goal         string  `json:"goal"`
	Activity     string  `json:"activity"`
	FitnessLevel string  `json:"fitnessLevel"`
	TargetWeight float64 `json:"targetWeight"`
}

func (s *UserService) UpdateProfile(dto UpdateProfileDTO) (*models.User, error) {
	user, err := s.repo.FindByTelegramID(dto.TelegramID)
	if err != nil {
		user = &models.User{
			TelegramID: dto.TelegramID,
			Role:       "user",
		}
	}

	user.Name = dto.Name
	user.Age = dto.Age
	user.Gender = dto.Gender
	user.Height = dto.Height
	user.Weight = dto.Weight
	user.Goal = dto.Goal
	user.Activity = dto.Activity
	user.FitnessLevel = dto.FitnessLevel
	user.TargetWeight = dto.TargetWeight

	if user.ID == 0 {
		return s.repo.Create(user)
	}

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser обновляет данные пользователя
func (s *UserService) UpdateUser(
	user *models.User,
) error {

	return s.repo.Update(user)
}
