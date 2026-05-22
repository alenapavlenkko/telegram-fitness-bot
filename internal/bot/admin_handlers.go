package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// handleAdminActions обрабатывает действия админа
func (b *BotApp) handleAdminActions(
	update tgbotapi.Update,
) {

	chatID := update.Message.Chat.ID

	if !b.isAdmin(int64(update.Message.From.ID)) {
		b.sendText(chatID, "⛔ Нет доступа")
		return
	}

	b.adminHandler.ShowAdminPanel(chatID)
}
