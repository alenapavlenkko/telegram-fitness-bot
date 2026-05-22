package bot

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleStatsCommand показывает статистику пользователя
func (b *BotApp) handleStatsCommand(
	update tgbotapi.Update,
	chatID int64,
) {

	user, err := b.userService.GetUserByTelegramID(
		int64(update.Message.From.ID),
	)

	if err != nil {

		b.sendText(
			chatID,
			"❌ Пользователь не найден",
		)

		return
	}

	// Получаем историю веса
	logs, err := b.weightService.GetUserHistory(user.ID)

	if err != nil || len(logs) == 0 {

		b.sendText(
			chatID,
			"📊 Статистики пока нет",
		)

		return
	}

	// ========================================
	// Аналитика
	// ========================================

	startWeight := logs[0].Weight
	currentWeight := logs[len(logs)-1].Weight

	minWeight := logs[0].Weight
	maxWeight := logs[0].Weight

	var sum float64

	for _, item := range logs {

		sum += item.Weight

		if item.Weight < minWeight {
			minWeight = item.Weight
		}

		if item.Weight > maxWeight {
			maxWeight = item.Weight
		}
	}

	avgWeight := sum / float64(len(logs))

	change := currentWeight - startWeight

	changeText := fmt.Sprintf(
		"%+.1f",
		change,
	)

	// ========================================
	// Целевой вес
	// ========================================

	targetWeight := user.TargetWeight

	if targetWeight == 0 {
		targetWeight = 50
	}

	// ========================================
	// RESULT
	// ========================================

	stats := fmt.Sprintf(
		`📊 *Ваша статистика*

⚖️ Текущий вес: %.1f кг
📍 Начальный вес: %.1f кг
🎯 Желаемый вес: %.1f кг
📉 Изменение: %s кг
📊 Минимум: %.1f кг
📊 Максимум: %.1f кг
📈 Средний вес: %.1f кг
📝 Записей: %d`,
		currentWeight,
		startWeight,
		targetWeight,
		changeText,
		minWeight,
		maxWeight,
		avgWeight,
		len(logs),
	)

	b.sendText(chatID, stats)
}
