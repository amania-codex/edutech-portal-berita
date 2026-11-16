package models

import "gorm.io/gorm"

// Speaker model for event keynotes and trainers
type Speaker struct {
	gorm.Model
	EventID uint   `json:"eventId" gorm:"not null;index"`
	Name    string `json:"name" gorm:"type:varchar(255);not null"`
	Role    string `json:"role" gorm:"type:varchar(255)"`
	Photo   string `json:"photo" gorm:"type:varchar(500)"`
}
