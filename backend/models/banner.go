package models

import (
	"time"

	"gorm.io/gorm"
)

type Banner struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Title     string         `json:"title"`
	ImageUrl  string         `json:"image_url"`
	TargetUrl string         `json:"target_url"`
	Placement string         `json:"placement" gorm:"type:varchar(20);not null;default:'kanan'"` // header, tengah, kiri, kanan
	IsActive  bool           `gorm:"not null" json:"is_active"` // tanpa default:true, agar false tidak diabaikan GORM
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
