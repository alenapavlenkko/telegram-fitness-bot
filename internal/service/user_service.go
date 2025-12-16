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
