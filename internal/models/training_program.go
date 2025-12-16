package models

import "gorm.io/gorm"

type TrainingProgram struct {
	gorm.Model           // добавляет ID, CreatedAt, UpdatedAt, DeletedAt
	Title       string   `gorm:"type:varchar(100);not null"`
	Description string   `gorm:"type:text"`
	Difficulty  string   `gorm:"type:varchar(50)"`
	Duration    int      `gorm:"not null"` // длительность в минутах
	CategoryID  *uint    // связь с Category
	Category    Category `gorm:"foreignKey:CategoryID"`
	YouTubeLink string   `gorm:"type:text"`
}
