package bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleCallback обрабатывает callback кнопки
func (b *BotApp) handleCallback(
	callback *tgbotapi.CallbackQuery,
) {

	userID := int64(callback.From.ID)

	// Проверяем права администратора
	if !b.isAdmin(userID) {

		b.answerCallback(
			callback.ID,
			"⛔ Доступ запрещен",
		)

		return
	}

	// Передаем обработку admin handler
	b.adminHandler.HandleAdminCallback(callback)
}

// answerCallback отвечает на callback query
func (b *BotApp) answerCallback(
	callbackID string,
	text string,
) {

	callback := tgbotapi.NewCallback(
		callbackID,
		text,
	)

	if _, err := b.API.Request(callback); err != nil {
		log.Printf("callback error: %v", err)
	}
}

// editMessage редактирует сообщение
func (b *BotApp) editMessage(
	chatID int64,
	messageID int,
	text string,
) {

	edit := tgbotapi.NewEditMessageText(
		chatID,
		messageID,
		text,
	)

	if _, err := b.API.Send(edit); err != nil {
		log.Printf("editMessage error: %v", err)
	}
}
