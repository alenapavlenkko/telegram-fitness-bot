package bot

import (
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleRegularMessage обрабатывает обычные сообщения
func (b *BotApp) handleRegularMessage(
	update tgbotapi.Update,
) {

	userID := int64(update.Message.From.ID)
	chatID := update.Message.Chat.ID
	text := strings.TrimSpace(update.Message.Text)

	log.Printf(
		"user=%d text=%s",
		userID,
		text,
	)

	// Проверка admin actions
	state, isAdminAction := b.adminHandler.GetState(userID)

	if isAdminAction {

		b.adminHandler.HandleAdminActions(
			chatID,
			userID,
			state,
			text,
		)

		return
	}

	// FSM состояния
	if state, exists := b.userStates[userID]; exists {

		switch state {

		case StateAwaitingWeight:

			b.handleWeightInput(
				chatID,
				text,
			)

			return

		case StateCalcWeight,
			StateCalcHeight,
			StateCalcAge,
			StateCalcGender,
			StateCalcActivity,
			StateCalcGoal:

			b.handleCalculatorFlow(update)
			return
		}
	}

	// Кнопки меню
	switch text {

	case "🏋️ Тренировки":

		b.showTrainingsForUser(chatID)

	case "🍎 Питание":

		b.showNutritionForUser(chatID)

	case "📂 Категории":

		b.showCategoriesForUser(chatID)

	case "📊 Статистика":

		b.handleStatsCommand(update, chatID)

	case "👤 Профиль":

		b.showProfile(chatID)

	case "⚖️ Записать вес":

		b.userStates[userID] = StateAwaitingWeight

		b.sendText(
			chatID,
			"⚖️ Введите ваш вес в кг:",
		)

	case "🧮 Калькулятор":

		b.startCalculator(
			chatID,
			userID,
		)

	default:

		b.sendText(
			chatID,
			"❌ Неизвестная команда",
		)
	}
}
