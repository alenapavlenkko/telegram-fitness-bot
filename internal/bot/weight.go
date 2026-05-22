package bot

import (
	"strconv"
)

// handleWeightInput сохраняет вес
func (b *BotApp) handleWeightInput(
	chatID int64,
	text string,
) {

	weight, err := strconv.ParseFloat(text, 64)
	if err != nil {

		b.sendText(
			chatID,
			"❌ Введите корректный вес",
		)

		return
	}

	// Получаем пользователя
	user, err := b.userService.GetUserByTelegramID(chatID)
	if err != nil {

		b.sendText(
			chatID,
			"❌ Пользователь не найден",
		)

		return
	}

	// Обновляем текущий вес
	user.Weight = weight

	b.userService.UpdateUser(user)

	// Сохраняем в историю
	err = b.weightService.LogWeight(
		user.ID,
		weight,
	)

	if err != nil {

		b.sendText(
			chatID,
			"❌ Ошибка сохранения веса",
		)

		return
	}

	b.sendText(
		chatID,
		"✅ Вес успешно сохранен",
	)

	// Очищаем состояние
	delete(b.userStates, chatID)

	// Возвращаем меню
	b.showMainMenu(chatID)
}
