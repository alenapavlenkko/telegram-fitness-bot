package models

import "gorm.io/gorm"

// Category — категория тренировок или питания
type Category struct {
	gorm.Model

	// Название категории
	Name string

	// Описание категории
	Description string

	// Тип:
	// training / nutrition / general
	Type string
}
