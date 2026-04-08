package models

import "gorm.io/gorm"
import "time"

type WeightLog struct {
	gorm.Model
	UserID uint      `gorm:"not null"`
	Weight float64   `gorm:"not null"` // кг
	Date   time.Time `gorm:"default:current_timestamp"`
}
