package bot

import (
	"fmt"

	"github.com/alenapavlenkko/telegramfitnes/internal/models"
)

// showNutritionForUser показывает питание
func (b *BotApp) showNutritionForUser(chatID int64) {

	nutritionList, err := b.nutritionService.ListNutrition()
	if err != nil {

		b.sendText(
			chatID,
			"❌ Ошибка загрузки питания",
		)

		return
	}

	if len(nutritionList) == 0 {

		b.sendText(
			chatID,
			"🍎 Планы питания пока отсутствуют",
		)

		return
	}

	msg := "🍎 *Планы питания:*\n\n"

	for i, item := range nutritionList {

		msg += formatNutrition(i+1, item)
	}

	b.sendText(chatID, msg)
}

// formatNutrition красиво форматирует питание
func formatNutrition(
	index int,
	n *models.NutritionPlan,
) string {

	return fmt.Sprintf(
		"%d. *%s* - %d ккал\n"+
			"%s\n"+
			"Б:%.1fг, У:%.1fг, Ж:%.1fг\n\n",
		index,
		n.Title,
		n.Calories,
		n.Description,
		n.Protein,
		n.Carbs,
		n.Fats,
	)
}
