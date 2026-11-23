package models

import "gorm.io/gorm"

// Setting — pengaturan situs (deskripsi, kontak, media sosial)
type Setting struct {
	Key   string `json:"key" gorm:"primaryKey;type:varchar(64)"`
	Value string `json:"value" gorm:"type:text"`
}

// Page — halaman statis yang diisi admin (Tentang Kami, Kontak, Pedoman Media Siber, ...)
type Page struct {
	gorm.Model
	Title       string `json:"title" gorm:"type:varchar(150);not null"`
	Slug        string `json:"slug" gorm:"type:varchar(150);uniqueIndex;not null"`
	Content     string `json:"content" gorm:"type:longtext"`
	MenuOrder   int    `json:"menuOrder" gorm:"not null;default:0"`
	IsPublished bool   `json:"isPublished" gorm:"not null;index"`
}
