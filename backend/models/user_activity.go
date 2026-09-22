package models

import (
	"time"

	"gorm.io/gorm"
)

type ReadingHistory struct {
	gorm.Model
	UserID uint      `json:"user_id" gorm:"not null;index"`
	NewsID uint      `json:"news_id" gorm:"not null;index"`
	ReadAt time.Time `json:"read_at" gorm:"not null"`

	// Relasi virtual untuk API response
	NewsTitle string `json:"news_title" gorm:"-"`
	NewsSlug  string `json:"news_slug" gorm:"-"`
	NewsImage string `json:"news_image" gorm:"-"`
}

type Bookmark struct {
	gorm.Model
	UserID uint `json:"user_id" gorm:"not null;index"`
	NewsID uint `json:"news_id" gorm:"not null;index"`

	// Relasi virtual untuk API response
	NewsTitle string `json:"news_title" gorm:"-"`
	NewsSlug  string `json:"news_slug" gorm:"-"`
	NewsImage string `json:"news_image" gorm:"-"`
}
