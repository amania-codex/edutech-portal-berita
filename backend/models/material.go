package models

import "gorm.io/gorm"

// Material model for event slides, recordings, and files
type Material struct {
	gorm.Model
	EventID  uint   `json:"eventId" gorm:"not null;index"`
	Title    string `json:"title" gorm:"type:varchar(255);not null"`
	Type     string `json:"type" gorm:"type:varchar(50);not null"` // 'file' or 'video'
	FilePath string `json:"filePath" gorm:"type:varchar(500)"`
	Link     string `json:"link" gorm:"type:varchar(500)"`
}
