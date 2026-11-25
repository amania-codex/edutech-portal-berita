package models

import "gorm.io/gorm"

// Video — video YouTube untuk kolom Video/Podcast di beranda dan halaman /video.
type Video struct {
	gorm.Model
	Title       string `json:"title" gorm:"type:varchar(255);not null"`
	YoutubeID   string `json:"youtubeId" gorm:"type:varchar(20);not null"`
	Description string `json:"description" gorm:"type:text"`
	// Urutan tampil: angka kecil lebih dulu, lalu yang terbaru
	SortOrder   int  `json:"sortOrder" gorm:"not null;default:0"`
	IsPublished bool `json:"isPublished" gorm:"not null;index"`
}
