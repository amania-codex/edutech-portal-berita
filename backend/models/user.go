package models

import "gorm.io/gorm"

// User model — MVC Model layer
type User struct {
	gorm.Model
	Name     string `json:"name" gorm:"type:varchar(100);not null"`
	Email    string `json:"email" gorm:"type:varchar(150);unique;not null"`
	Password string `json:"-" gorm:"type:varchar(255);not null"`
	Role     string `json:"role" gorm:"type:varchar(20);default:'user'"`
	Avatar   string `json:"avatar" gorm:"type:varchar(500)"`
	Bio      string `json:"bio" gorm:"type:text"`
}
