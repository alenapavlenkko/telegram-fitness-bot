package database

import (
	"github.com/alenapavlenkko/telegramfitnes/internal/models"
	"gorm.io/gorm"
)

func SeedData(db *gorm.DB) error {
	var trainingCount int64

	db.Model(&models.TrainingProgram{}).Count(&trainingCount)
	if trainingCount > 0 {
		return nil
	}

	// Категории
	categories := []models.Category{
		{Name: "Кардио", Description: "Тренировки для выносливости и жиросжигания", Type: "training"},
		{Name: "Силовые", Description: "Упражнения для силы и мышц", Type: "training"},
		{Name: "Растяжка", Description: "Гибкость, восстановление и расслабление", Type: "training"},
		{Name: "Похудение", Description: "Питание и тренировки для снижения веса", Type: "general"},
		{Name: "Завтраки", Description: "Полезные завтраки", Type: "nutrition"},
		{Name: "Обеды", Description: "Сбалансированные обеды", Type: "nutrition"},
		{Name: "Ужины", Description: "Лёгкие и полезные ужины", Type: "nutrition"},
		{Name: "Перекусы", Description: "Полезные перекусы", Type: "nutrition"},
	}

	if err := db.Create(&categories).Error; err != nil {
		return err
	}

	cardioID := categories[0].ID
	strengthID := categories[1].ID
	stretchID := categories[2].ID

	breakfastID := categories[4].ID
	lunchID := categories[5].ID
	dinnerID := categories[6].ID
	snackID := categories[7].ID

	// Тренировки
	trainings := []models.TrainingProgram{
		{
			Title:       "Утренняя зарядка",
			Description: "Лёгкая тренировка на всё тело для бодрого начала дня.",
			Difficulty:  "Легкая",
			Duration:    15,
			CategoryID:  &stretchID,
			YouTubeLink: "https://www.youtube.com/watch?v=2L2lnxIcNmo",
		},
		{
			Title:       "Кардио для похудения",
			Description: "Интенсивная тренировка для сжигания калорий и улучшения выносливости.",
			Difficulty:  "Средняя",
			Duration:    30,
			CategoryID:  &cardioID,
			YouTubeLink: "https://www.youtube.com/watch?v=ml6cT4AZdqI",
		},
		{
			Title:       "Силовая тренировка дома",
			Description: "Комплекс упражнений без оборудования для укрепления мышц.",
			Difficulty:  "Средняя",
			Duration:    40,
			CategoryID:  &strengthID,
			YouTubeLink: "https://www.youtube.com/watch?v=UItWltVZZmE",
		},
		{
			Title:       "Растяжка после тренировки",
			Description: "Комплекс для расслабления мышц и восстановления.",
			Difficulty:  "Легкая",
			Duration:    20,
			CategoryID:  &stretchID,
			YouTubeLink: "https://www.youtube.com/watch?v=g_tea8ZNk5A",
		},
		{
			Title:       "HIIT тренировка",
			Description: "Интервальная тренировка высокой интенсивности для быстрого результата.",
			Difficulty:  "Сложная",
			Duration:    25,
			CategoryID:  &cardioID,
			YouTubeLink: "https://www.youtube.com/watch?v=ml6cT4AZdqI",
		},
	}

	if err := db.Create(&trainings).Error; err != nil {
		return err
	}

	// Питание
	nutrition := []models.NutritionPlan{
		{
			Title:       "Овсянка с бананом и орехами",
			Description: "Полезный завтрак с медленными углеводами.",
			Calories:    380,
			Protein:     12,
			Carbs:       58,
			Fats:        11,
			CategoryID:  breakfastID,
		},
		{
			Title:       "Омлет с овощами",
			Description: "Белковый завтрак для энергии и насыщения.",
			Calories:    320,
			Protein:     24,
			Carbs:       10,
			Fats:        20,
			CategoryID:  breakfastID,
		},
		{
			Title:       "Куриная грудка с гречкой",
			Description: "Сбалансированный обед с белком и сложными углеводами.",
			Calories:    520,
			Protein:     42,
			Carbs:       55,
			Fats:        12,
			CategoryID:  lunchID,
		},
		{
			Title:       "Рыба с овощами",
			Description: "Лёгкий и полезный ужин.",
			Calories:    410,
			Protein:     36,
			Carbs:       20,
			Fats:        18,
			CategoryID:  dinnerID,
		},
		{
			Title:       "Творог с ягодами",
			Description: "Полезный белковый перекус.",
			Calories:    240,
			Protein:     26,
			Carbs:       18,
			Fats:        6,
			CategoryID:  snackID,
		},
		{
			Title:       "Салат с индейкой",
			Description: "Лёгкий обед или ужин с высоким содержанием белка.",
			Calories:    360,
			Protein:     34,
			Carbs:       18,
			Fats:        14,
			CategoryID:  lunchID,
		},
	}

	if err := db.Create(&nutrition).Error; err != nil {
		return err
	}

	// Недельное меню
	menu := models.WeeklyMenu{
		Name:          "Сбалансированное меню на неделю",
		Description:   "Готовый рацион для поддержания формы и здорового питания.",
		TotalCalories: 11200,
		Active:        true,
	}

	if err := db.Create(&menu).Error; err != nil {
		return err
	}

	days := []models.MenuDay{
		{MenuID: menu.ID, DayNumber: 1, DayName: "Понедельник", TotalCalories: 1580},
		{MenuID: menu.ID, DayNumber: 2, DayName: "Вторник", TotalCalories: 1620},
		{MenuID: menu.ID, DayNumber: 3, DayName: "Среда", TotalCalories: 1590},
		{MenuID: menu.ID, DayNumber: 4, DayName: "Четверг", TotalCalories: 1610},
		{MenuID: menu.ID, DayNumber: 5, DayName: "Пятница", TotalCalories: 1600},
		{MenuID: menu.ID, DayNumber: 6, DayName: "Суббота", TotalCalories: 1650},
		{MenuID: menu.ID, DayNumber: 7, DayName: "Воскресенье", TotalCalories: 1550},
	}

	if err := db.Create(&days).Error; err != nil {
		return err
	}

	for _, day := range days {
		meals := []models.DayMeal{
			{
				DayID:       day.ID,
				MealType:    "Завтрак",
				MealTime:    "09:00",
				NutritionID: nutrition[0].ID,
				Notes:       "Можно добавить корицу или ягоды.",
			},
			{
				DayID:       day.ID,
				MealType:    "Обед",
				MealTime:    "13:30",
				NutritionID: nutrition[2].ID,
				Notes:       "Подходит после тренировки.",
			},
			{
				DayID:       day.ID,
				MealType:    "Перекус",
				MealTime:    "16:30",
				NutritionID: nutrition[4].ID,
				Notes:       "Лёгкий белковый перекус.",
			},
			{
				DayID:       day.ID,
				MealType:    "Ужин",
				MealTime:    "19:30",
				NutritionID: nutrition[3].ID,
				Notes:       "Лёгкий ужин без тяжести.",
			},
		}

		if err := db.Create(&meals).Error; err != nil {
			return err
		}
	}

	return nil
}
