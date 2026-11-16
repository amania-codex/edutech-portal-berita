package models

import (
	"time"

	"gorm.io/gorm"
)

// Event model for seminars, webinars, workshops, and educational events
type Event struct {
	gorm.Model
	Title           string         `json:"title" gorm:"type:varchar(255);not null"`
	Slug            string         `json:"slug" gorm:"type:varchar(255);unique;not null;index"`
	Description     string         `json:"description" gorm:"type:longtext;not null"`
	Venue           string         `json:"venue" gorm:"type:varchar(255);not null"`
	StartTime       time.Time      `json:"startTime" gorm:"not null"`
	EndTime         time.Time      `json:"endTime" gorm:"not null"`
	Quota           int            `json:"quota" gorm:"default:0"`
	Price           int            `json:"price" gorm:"default:0"`
	PremiumPrice    *int           `json:"premiumPrice"`
	CertificateLink string         `json:"certificateLink" gorm:"type:varchar(500)"`
	ImageURL        string         `json:"imageUrl" gorm:"type:varchar(500)"`
	Status          string         `json:"status" gorm:"type:varchar(50);default:'published';index"` // published, draft, completed, cancelled
	Speakers        []Speaker      `json:"speakers,omitempty" gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE"`
	Materials       []Material     `json:"materials,omitempty" gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE"`
	BankAccounts    []BankAccount  `json:"bankAccounts,omitempty" gorm:"many2many:bank_account_event;constraint:OnDelete:CASCADE"`
	Registrations   []Registration `json:"registrations,omitempty" gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE"`
}
