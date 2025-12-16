package models

import (
	"time"

	"gorm.io/gorm"
)

type UserProgress struct {
	gorm.Model
	UserID      uint
	User        User `gorm:"foreignKey:UserID"` // Без models.
	TrainingID  uint
	Training    TrainingProgram `gorm:"foreignKey:TrainingID"` // Без models.
	CompletedAt time.Time
	Notes       string
}
