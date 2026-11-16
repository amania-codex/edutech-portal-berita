package models

import "gorm.io/gorm"

// Registration model for event participants
type Registration struct {
	gorm.Model
	TicketCode   string  `json:"ticketCode" gorm:"type:varchar(100);uniqueIndex"`
	UserID       *uint   `json:"userId" gorm:"index"`
	User         *User   `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
	EventID      uint    `json:"eventId" gorm:"not null;index"`
	Event        *Event  `json:"event,omitempty" gorm:"foreignKey:EventID;constraint:OnDelete:CASCADE"`
	Name         string  `json:"name" gorm:"type:varchar(255);not null"`
	Email        string  `json:"email" gorm:"type:varchar(150);not null;index"`
	Phone        string  `json:"phone" gorm:"type:varchar(50)"`
	PaymentProof string  `json:"paymentProof" gorm:"type:varchar(500)"`
	Status       string  `json:"status" gorm:"type:varchar(50);default:'pending';index"` // pending, waiting, verified, rejected
	Tier         string  `json:"tier" gorm:"type:varchar(50);default:'free'"`            // free, premium
	TotalAmount  float64 `json:"totalAmount" gorm:"type:decimal(15,2);default:0.00"`
}
