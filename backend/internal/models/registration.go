package models

import "gorm.io/gorm"

type Registration struct {
	gorm.Model
	UserID  uint
	EventID uint
	Status  string
}