package bot

import (
	"fmt"

	"github.com/alenapavlenkko/telegramfitnes/internal/models"
)

// showTrainingsForUser показывает тренировки
func (b *BotApp) showTrainingsForUser(chatID int64) {

	trainings, err := b.trainingService.ListTrainings()
	if err != nil {

		b.sendText(
			chatID,
			"❌ Ошибка загрузки тренировок",
		)

		return
	}

	if len(trainings) == 0 {

		b.sendText(
			chatID,
			"🏋️ Тренировок пока нет",
		)

		return
	}

	msg := "🏋️ *Доступные тренировки:*\n\n"

	for i, t := range trainings {

		msg += formatTraining(i+1, t)
	}

	b.sendText(chatID, msg)
}

// formatTraining красиво форматирует тренировку
func formatTraining(
	index int,
	t *models.TrainingProgram,
) string {

	video := ""

	if t.YouTubeLink != "" {

		video = fmt.Sprintf(
			"\n🎥 [Смотреть на YouTube](%s)",
			t.YouTubeLink,
		)
	}

	return fmt.Sprintf(
		"%d. *%s* - %d мин\n%s%s\n\n",
		index,
		t.Title,
		t.Duration,
		t.Description,
		video,
	)
}
