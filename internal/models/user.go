package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	TelegramID int64 `gorm:"uniqueIndex"`
	Username   string
	FirstName  string
	LastName   string
	Name       string
	Role       string `gorm:"default:'user'"`
	IsAdmin    bool
}
