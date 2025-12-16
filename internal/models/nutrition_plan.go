package models

import "gorm.io/gorm"

type NutritionPlan struct {
	gorm.Model
	Title       string
	Description string
	Calories    int
	Protein     float64
	Carbs       float64
	Fats        float64
	CategoryID  uint
	Category    Category `gorm:"foreignKey:CategoryID"` // Без models.
}

// ==================== НОВЫЕ МОДЕЛИ ДЛЯ НЕДЕЛЬНОГО МЕНЮ ====================

// WeeklyMenu - недельное меню
type WeeklyMenu struct {
	gorm.Model
	Name          string    `gorm:"size:255;not null"` // Название меню
	Description   string    `gorm:"type:text"`         // Описание
	TotalCalories int       // Общее количество калорий за неделю
	Active        bool      `gorm:"default:false"` // Активно ли меню (только одно может быть активным)
	Days          []MenuDay `gorm:"foreignKey:MenuID"`
}

// MenuDay - день в недельном меню
type MenuDay struct {
	gorm.Model
	MenuID        uint      `gorm:"not null"`
	DayNumber     int       `gorm:"not null"`         // 1-7 (понедельник-воскресенье)
	DayName       string    `gorm:"size:20;not null"` // Понедельник, Вторник...
	TotalCalories int       // Калории за день
	Meals         []DayMeal `gorm:"foreignKey:DayID"`
}

// DayMeal - прием пищи в конкретный день
type DayMeal struct {
	gorm.Model
	DayID       uint          `gorm:"not null"`
	MealType    string        `gorm:"size:50;not null"` // Завтрак, Обед, Ужин, Перекус
	MealTime    string        `gorm:"size:50"`          // Время приема пищи (09:00, 13:30 и т.д.)
	NutritionID uint          // Ссылка на конкретное блюдо/продукт
	Nutrition   NutritionPlan `gorm:"foreignKey:NutritionID"`
	Notes       string        `gorm:"type:text"` // Дополнительные заметки
}
