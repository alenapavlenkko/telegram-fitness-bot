package bot

import (
	"strings"
)

// showCategoriesForUser показывает категории
func (b *BotApp) showCategoriesForUser(chatID int64) {

	categories, err := b.categoryService.ListCategories()
	if err != nil {

		b.sendText(
			chatID,
			"❌ Ошибка загрузки категорий",
		)

		return
	}

	if len(categories) == 0 {

		b.sendText(
			chatID,
			"📂 Категории пока отсутствуют",
		)

		return
	}

	// ========================================
	// Группы категорий
	// ========================================

	var trainingCategories []string
	var nutritionCategories []string
	var otherCategories []string

	for _, category := range categories {

		name := category.Name

		// Распределяем категории
		switch strings.ToLower(name) {

		case "силовая",
			"кардио",
			"растяжка",
			"hiit",
			"домашние тренировки":

			trainingCategories = append(
				trainingCategories,
				name,
			)

		case "завтраки",
			"обеды",
			"ужины",
			"перекусы",
			"пп рецепты":

			nutritionCategories = append(
				nutritionCategories,
				name,
			)

		default:

			otherCategories = append(
				otherCategories,
				name,
			)
		}
	}

	// ========================================
	// Красивый вывод
	// ========================================

	msg := "📂 *Категории:*\n\n"

	// Тренировки
	if len(trainingCategories) > 0 {

		msg += "🏋️ *Тренировки:*\n"

		for _, item := range trainingCategories {

			msg += "• " + item + "\n"
		}

		msg += "\n"
	}

	// Питание
	if len(nutritionCategories) > 0 {

		msg += "🍎 *Питание:*\n"

		for _, item := range nutritionCategories {

			msg += "• " + item + "\n"
		}

		msg += "\n"
	}

	// Остальные
	if len(otherCategories) > 0 {

		msg += "📋 *Общие:*\n"

		for _, item := range otherCategories {

			msg += "• " + item + "\n"
		}
	}

	b.sendText(chatID, msg)
}
