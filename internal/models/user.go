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

	Age          int
	Gender       string
	Height       float64
	Weight       float64
	Goal         string
	Activity     string
	FitnessLevel string
	TargetWeight float64
}
