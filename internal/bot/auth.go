package bot

import (
	"github.com/alenapavlenkko/telegramfitnes/internal/models"
	"github.com/alenapavlenkko/telegramfitnes/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// isAdmin проверяет администратора
func (b *BotApp) isAdmin(userID int64) bool {

	for _, id := range b.Admins {
		if id == userID {
			return true
		}
	}

	return false
}

// authenticateUser авторизует пользователя
func (b *BotApp) authenticateUser(update tgbotapi.Update) (*models.User, error) {

	tgUser := update.Message.From

	user, err := b.userService.GetUserByTelegramID(int64(tgUser.ID))
	if err == nil {
		return user, nil
	}

	return b.userService.CreateUser(service.CreateUserDTO{
		TelegramID: int64(tgUser.ID),
		Name:       tgUser.UserName,
		Role:       "user",
	})
}

// isAuthorized проверяет доступ
func (b *BotApp) isAuthorized(userID int64, requiredRole string) bool {

	if requiredRole == "admin" && b.isAdmin(userID) {
		return true
	}

	user, err := b.userService.GetUserByTelegramID(userID)
	if err != nil {
		return false
	}

	if requiredRole == "user" {
		return true
	}

	return user.Role == requiredRole
}
