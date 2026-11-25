package models

import (
	"time"
	
	"gorm.io/gorm"
)

// User model — MVC Model layer
type User struct {
	gorm.Model
	Name     string `json:"name" gorm:"type:varchar(100);not null"`
	Email    string `json:"email" gorm:"type:varchar(150);unique;not null"`
	Password string `json:"-" gorm:"type:varchar(255);not null"`
	Role          string     `json:"role" gorm:"type:varchar(20);default:'user'"`
	Phone         string     `json:"phone" gorm:"type:varchar(50)"`
	Avatar        string     `json:"avatar" gorm:"type:varchar(500)"`
	Bio           string     `json:"bio" gorm:"type:text"`
	ResetCode     string     `json:"-" gorm:"type:varchar(64)"`
	ResetCodeExp  *time.Time `json:"-"`
	ResetAttempts int        `json:"-" gorm:"default:0"`
	// Akun yang dinonaktifkan admin tidak bisa masuk; tulisannya tetap ada.
	Suspended   bool       `json:"suspended" gorm:"not null;default:false"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
}

// Peran akun. Staf redaksi (admin, editor, penulis) bisa masuk dashboard.
const (
	RoleAdmin   = "admin"
	RoleEditor  = "editor"
	RolePenulis = "penulis"
	RoleUser    = "user"
)

// IsStaffRole — peran yang boleh menulis berita.
func IsStaffRole(role string) bool {
	return role == RoleAdmin || role == RoleEditor || role == RolePenulis
}
