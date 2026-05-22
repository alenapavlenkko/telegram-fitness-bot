package bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleCommand обрабатывает команды
func (b *BotApp) handleCommand(update tgbotapi.Update) {

	cmd := update.Message.Command()
	chatID := update.Message.Chat.ID

	switch cmd {

	case "start":

		_, err := b.authenticateUser(update)
		if err != nil {
			b.sendText(chatID, "❌ Ошибка авторизации")
			return
		}

		b.sendText(
			chatID,
			"👋 Рад вас видеть! Сейчас открою главное меню...",
		)

		b.showMainMenu(chatID)

	case "help":

		helpMsg := `📚 *Помощь по использованию Fitness Bot*

Основные команды:
/start - Главное меню
/help - Помощь
/admin - Панель администратора

Используйте кнопки меню для навигации.`

		b.sendText(chatID, helpMsg)

	case "admin":

		log.Printf("Admin command from %d", update.Message.From.ID)

		if !b.isAdmin(int64(update.Message.From.ID)) {

			b.sendText(chatID, "⛔ Недостаточно прав")
			return
		}

		b.adminHandler.ShowAdminPanel(chatID)

	default:

		b.sendText(chatID, "❌ Неизвестная команда")
	}
}
