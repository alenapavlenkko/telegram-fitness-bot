package bot

import "fmt"

// showProfile показывает профиль пользователя
func (b *BotApp) showProfile(chatID int64) {

	user, err := b.userService.GetUserByTelegramID(chatID)
	if err != nil {

		b.sendText(
			chatID,
			"❌ Профиль не найден",
		)

		return
	}

	age := user.Age
	height := int(user.Height)
	weight := user.Weight

	targetWeight := user.TargetWeight

	if age == 0 {
		age = 20
	}

	if height == 0 {
		height = 160
	}

	if weight == 0 {
		weight = 53
	}

	if targetWeight == 0 {
		targetWeight = 50
	}

	goal := user.Goal
	if goal == "" {
		goal = "Похудение"
	}

	activity := user.Activity
	if activity == "" {
		activity = "Средняя"
	}

	level := user.FitnessLevel
	if level == "" {
		level = "Продвинутый"
	}

	profile := fmt.Sprintf(
		`👤 *Ваш профиль*

Имя: %s
Возраст: %d
Рост: %d см
Вес: %.1f кг
Цель: %s
Активность: %s
Уровень: %s
Желаемый вес: %.1f кг`,
		user.Name,
		age,
		height,
		weight,
		goal,
		activity,
		level,
		targetWeight,
	)

	b.sendText(chatID, profile)
}
