package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Name        string
	Description string
	Type        string // "training", "nutrition", "general"
}
