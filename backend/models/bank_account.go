package models

import "gorm.io/gorm"

// BankAccount model for payment destinations
type BankAccount struct {
	gorm.Model
	BankName      string  `json:"bankName" gorm:"type:varchar(100);not null"`
	AccountNumber string  `json:"accountNumber" gorm:"type:varchar(100);not null"`
	AccountHolder string  `json:"accountHolder" gorm:"type:varchar(150);not null"`
	Logo          string  `json:"logo" gorm:"type:varchar(500)"`
	IsActive      bool    `json:"isActive" gorm:"not null"` // tanpa default:true, agar false tidak diabaikan GORM
	Events        []Event `json:"events,omitempty" gorm:"many2many:bank_account_event"`
}
