package models

import "gorm.io/gorm"

// User — пользователь Telegram бота
type User struct {
	gorm.Model

	// Telegram
	TelegramID int64 `gorm:"uniqueIndex"`

	Username  string
	FirstName string
	LastName  string

	// Отображаемое имя
	Name string

	// Роль пользователя
	Role string `gorm:"default:'user'"`

	// Администратор
	IsAdmin bool

	// ========================================
	// Профиль
	// ========================================

	Age int

	Gender string

	Height float64

	Weight float64

	Goal string

	// Низкая / Средняя / Высокая
	Activity string

	// Начальный / Средний / Продвинутый
	FitnessLevel string

	TargetWeight float64
}
