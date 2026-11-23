package models

import "time"

// NewsletterSubscriber — pendaftar newsletter dari beranda.
type NewsletterSubscriber struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"type:varchar(150);uniqueIndex;not null"`
	CreatedAt time.Time `json:"created_at"`
}
