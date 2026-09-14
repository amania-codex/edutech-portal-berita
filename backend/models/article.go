package models

import "gorm.io/gorm"

// Article model — MVC Model layer
type Article struct {
	gorm.Model
	Title       string `json:"title" gorm:"type:varchar(255);not null"`
	Slug        string `json:"slug" gorm:"type:varchar(255);unique;not null"`
	Excerpt     string `json:"excerpt" gorm:"type:text"`
	Content     string `json:"content" gorm:"type:longtext;not null"`
	ImageURL    string `json:"imageUrl" gorm:"type:varchar(500)"`
	Category    string `json:"category" gorm:"type:varchar(100)"`
	Author      string `json:"author" gorm:"type:varchar(100)"`
	AuthorID    uint   `json:"author_id" gorm:"default:0"`
	Tags        string `json:"tags" gorm:"type:varchar(500)"`
	ViewCount   int    `json:"viewCount" gorm:"default:0"`
	IsFeatured  bool   `json:"isFeatured" gorm:"default:false"`
	IsTrending  bool   `json:"isTrending" gorm:"default:false"`
	IsPublished bool   `json:"isPublished" gorm:"default:true"`
	Status      string `json:"status" gorm:"type:varchar(20);default:'published'"`
}