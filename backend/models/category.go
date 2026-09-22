package models

import "gorm.io/gorm"

// Category model — MVC Model layer
type Category struct {
	gorm.Model
	Name string `json:"name" gorm:"type:varchar(100);unique;not null"`
	Slug string `json:"slug" gorm:"type:varchar(150);unique;not null"`
}
