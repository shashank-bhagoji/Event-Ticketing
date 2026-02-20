package models

import (
	"time"
	"gorm.io/gorm"
)

type Event struct {
	gorm.Model
	Title             string
	Description       string
	EventDate         time.Time
	TotalCapacity     int
	AvailableCapacity int
	CreatedBy         uint
}