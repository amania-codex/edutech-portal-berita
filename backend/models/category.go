package models

import "gorm.io/gorm"

// Category model — MVC Model layer
type Category struct {
	gorm.Model
	Name     string `json:"name" gorm:"type:varchar(100);unique;not null"`
	Slug     string `json:"slug" gorm:"type:varchar(150);unique;not null"`
	ParentID *uint  `json:"parentId" gorm:"index"`
	WPID     uint   `json:"wpId" gorm:"column:wp_id;default:0;index"`

	// Diisi saat request ?with_count=1 (bukan kolom tabel)
	Count int64 `json:"count" gorm:"-"`
}
